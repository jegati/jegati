/* First-party push worker. No fetch handler or response cache. Database schema is
   shared with src/push-storage.ts: one temporary capability, deadline and binding. */
self.addEventListener('install', event => event.waitUntil(self.skipWaiting()));
self.addEventListener('activate', event => event.waitUntil(self.clients.claim()));
const valid = value => value && /^[A-Za-z0-9_-]{43}$/.test(value.token) && /^[a-f0-9]{64}$/.test(value.binding) && Number.isSafeInteger(value.revision) && value.revision >= 0 && Number.isFinite(value.expires) && value.expires > Date.now() && value.expires <= Date.now() + 120 * 60000 && (value.status === 'enabled' || value.status === 'pending');
function record(removeToken) {
  return new Promise((resolve, reject) => {
    let absent = false, blocked = false;
    const request = indexedDB.open('gati-push-v1', 1);
    request.onupgradeneeded = () => { absent = true; request.transaction.abort(); };
    request.onerror = () => absent ? resolve(null) : reject(new Error('storage unavailable'));
    request.onblocked = () => { blocked = true; resolve(null); };
    request.onsuccess = () => {
      const db = request.result; if (blocked) { db.close(); return; } db.onversionchange = () => db.close();
      const tx = db.transaction('resume', 'readwrite'), store = tx.objectStore('resume'), get = store.get('current'); let value = null;
      get.onsuccess = () => {
        const candidate = get.result;
        if (candidate?.token === removeToken || candidate?.expires <= Date.now()) store.delete('current');
        else if (valid(candidate)) value = candidate;
      };
      tx.oncomplete = () => { db.close(); resolve(value); };
      tx.onerror = () => { db.close(); reject(new Error('storage unavailable')); };
    };
  });
}
const withLock = action => self.navigator.locks.request('gati-push', { signal: AbortSignal.timeout(15000) }, action);
async function drop(token) {
  await record(token);
  if (await record()) return;
  for (const notification of await self.registration.getNotifications({ tag: 'gati-current' })) notification.close();
  await (await self.registration.pushManager.getSubscription())?.unsubscribe();
}
self.addEventListener('push', event => event.waitUntil((async () => {
  try {
    const data = event.data?.json(), state = await record();
    if (!state) { await withLock(() => drop()); return; }
    if (Notification.permission !== 'granted') { await withLock(() => drop(state.token)); return; }
    if (state.status !== 'enabled') return;
    if (!data || data.binding !== state.binding || !Number.isFinite(data.expires_at) || data.expires_at <= Date.now() || data.expires_at > state.expires) return;
    // Verify current server lifetime/state; receipt cannot create willingness or arrival.
    const response = await fetch('/api/signal', { credentials: 'omit', cache: 'no-store', signal: AbortSignal.timeout(10000), headers: { Authorization: `Bearer ${state.token}` } });
    if ([401, 410].includes(response.status)) { await withLock(() => drop(state.token)); return; }
    if (!response.ok) return;
    const current = await response.json();
    if (!current.invitation || current.expires_at <= Date.now() || current.invitation.ends_at <= Date.now()) return;
    const latest = await record();
    if (!latest || latest.status !== 'enabled' || latest.binding !== state.binding || latest.expires <= Date.now() || data.expires_at <= Date.now()) return;
    await self.registration.showNotification('GATI 🦩', { body: 'Ka një përditësim. Hape GATI për të parë gjendjen.', tag: 'gati-current', renotify: false, data: { binding: state.binding, expires_at: Math.min(data.expires_at, current.expires_at) } });
  } catch { /* Fail closed on unreadable, stale or unavailable state; no secret logs. */ }
})()));
self.addEventListener('notificationclick', event => {
  event.notification.close();
  event.waitUntil((async () => {
    const windows = await self.clients.matchAll({ type: 'window', includeUncontrolled: true });
    const target = windows.find(client => new URL(client.url).origin === self.location.origin && new URL(client.url).pathname === '/');
    if (target) { await target.focus(); target.postMessage({ type: 'gati-refresh' }); }
    else await self.clients.openWindow('/'); // Never place a capability or gathering ID in a URL.
  })());
});
