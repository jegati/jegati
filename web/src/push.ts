import type { Session } from './session';
import { claimResume, clearResume, readResume, updateResume, validResume, type PushResume } from './push-storage';
const element = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
const supported = () => isSecureContext && 'Notification' in window && 'serviceWorker' in navigator && 'PushManager' in window && 'locks' in navigator;
class PushNotice extends Error {}
async function browserOperation<T>(operation: Promise<T>): Promise<T> {
  let timer: ReturnType<typeof setTimeout> | undefined;
  try { return await Promise.race([operation, new Promise<T>((_, reject) => { timer = setTimeout(() => reject(new PushNotice('Shfletuesi nuk u përgjigj. Provo përsëri.')), 10000); })]); }
  finally { if (timer) clearTimeout(timer); }
}
const randomBinding = () => Array.from(crypto.getRandomValues(new Uint8Array(32)), b => b.toString(16).padStart(2, '0')).join('');
function decodeKey(text: string): Uint8Array<ArrayBuffer> { return Uint8Array.from(atob(text.replaceAll('-', '+').replaceAll('_', '/')), character => character.charCodeAt(0)); }
async function api(record: PushResume, method: string, path: string, body?: unknown) {
  if (!validResume(record)) throw new PushNotice('Gatishmëria përfundoi.');
  return fetch(path, { method, credentials: 'omit', cache: 'no-store', signal: AbortSignal.timeout(10000), headers: { Authorization: `Bearer ${record.token}`, ...(body ? { 'Content-Type': 'application/json' } : {}) }, ...(body ? { body: JSON.stringify(body) } : {}) });
}
async function withLock<T>(action: () => Promise<T>): Promise<T> {
  return navigator.locks.request('gati-push', { signal: AbortSignal.timeout(15000) }, action);
}
async function unsubscribeIfUnused() {
  if (await readResume()) return;
  const registration = await browserOperation(navigator.serviceWorker.getRegistration('/'));
  if (!registration) return;
  const subscription = await browserOperation(registration.pushManager.getSubscription());
  if (subscription) await browserOperation(subscription.unsubscribe());
  for (const notification of await browserOperation(registration.getNotifications({ tag: 'gati-current' }))) notification.close();
}

export class PushController {
  private record: PushResume | null = null;
  private configured = false;
  private publicKey = '';
  private busy = false;
  private generation = 0;
  private recovering = false;
  private resumeOnly = false;
  private info = '';
  constructor(private current: () => Session | null, private resumed: (value: Session) => void, private changed: () => void) {
    element<HTMLButtonElement>('push-enable').onclick = () => { void this.enable(); };
    element<HTMLButtonElement>('push-disable').onclick = () => { void this.disable(); };
    element<HTMLButtonElement>('push-recover').onclick = () => { void this.recover(); };
    element<HTMLButtonElement>('push-cancel-session').onclick = () => { void this.cancelRecovery(); };
  }
  get pendingResume() { return this.recovering; }
  async initialize() {
    try {
      if (!this.resumeOnly) this.record = await readResume();
      this.recovering = !!this.record && !this.current();
      this.render(); this.changed();
      void this.configure(); // Optional transport must not delay the first JAM GATI action.
      if (this.record && typeof Notification !== 'undefined' && Notification.permission !== 'granted') await this.disable();
      if (this.recovering) await this.recover();
      else if (this.record) await this.checkRegistration();
      else if (supported()) void withLock(unsubscribeIfUnused).catch(() => {});
    } catch { this.info = 'Njoftimet nuk u kontrolluan. Gatishmëria mund të përdoret pa to.'; }
    this.render(); this.changed();
  }
  private async configure() {
    try {
      const response = await fetch('/api/push-config', { credentials: 'omit', cache: 'no-store', signal: AbortSignal.timeout(10000) });
      if (response.ok) { const data = await response.json(); this.configured = data.enabled === true && typeof data.public_key === 'string'; this.publicKey = data.public_key ?? ''; }
    } catch { this.configured = false; }
    this.render();
  }
  render() {
    const session = this.current(), own = this.record?.token === session?.token;
    element('push-controls').hidden = !session?.confirmed && !this.record && !this.recovering;
    const enable = element<HTMLButtonElement>('push-enable');
    enable.hidden = own && this.record?.status === 'enabled';
    enable.disabled = this.busy || !supported() || !this.configured || !session?.confirmed || !navigator.onLine || (!!this.record && !own);
    enable.textContent = own && this.record?.status === 'pending' ? 'Provo përsëri njoftimet' : 'Aktivizo njoftimet';
    element('push-disable').hidden = !this.record || this.resumeOnly;
    element('push-recover').hidden = !this.recovering;
    element<HTMLButtonElement>('push-recover').disabled = this.busy || !navigator.onLine;
    element('push-cancel-session').hidden = !this.recovering;
    element<HTMLButtonElement>('push-cancel-session').disabled = this.busy || !navigator.onLine;
    element('push-status').textContent = this.info || (this.recovering ? 'Po rikthehet gatishmëria e përkohshme. Nëse lidhja mungon, provo përsëri; nuk krijohet gatishmëri e re.' : !supported() ? 'Njoftimet në sfond nuk mbështeten këtu. Në disa pajisje nevojitet shtimi i GATI në ekranin kryesor.' : !this.configured ? 'Njoftimet në sfond nuk janë aktivizuar në këtë shërbim. Faqja e hapur vazhdon të përditësohet.' : this.record && !own ? 'Njoftimet janë të lidhura me një gatishmëri tjetër në këtë shfletues. Çaktivizoji për t’i lidhur me këtë.' : this.record?.status === 'enabled' ? 'Njoftimet janë aktive vetëm deri në përfundimin e gatishmërisë.' : 'Njoftimet në sfond janë të fikura.');
  }
  private async checkRegistration() {
    const record = this.record; if (!record) return;
    const response = await api(record, 'GET', '/api/push');
    if (!response.ok) return;
    const state = await response.json();
    if (state.enabled === true && state.binding === record.binding && Number.isFinite(state.expires_at)) {
      const updated: PushResume = { ...record, status: 'enabled', expires: Math.min(record.expires, state.expires_at) };
      if (await updateResume(updated)) this.record = updated;
    } else { const updated: PushResume = { ...record, status: 'pending' }; if (await updateResume(updated)) this.record = updated; }
  }
  private async enable() {
    const session = this.current(); if (this.busy || !session?.confirmed || !supported() || !this.configured || !navigator.onLine) return;
    const generation = ++this.generation, owner = session.token;
    this.busy = true; this.info = ''; this.render();
    // Request permission in the button's user gesture, never during initialization.
    const active = () => generation === this.generation && this.current()?.token === owner && session.expires > Date.now();
    try {
      const permission = Notification.requestPermission();
      if (await browserOperation(permission) !== 'granted') throw new PushNotice('Leja për njoftime nuk u dha. Gatishmëria jote vazhdon pa to.');
      if (!active()) return;
      await withLock(async () => {
        if (!active()) return;
        const previous = await readResume();
        const record = await claimResume({ token: owner, binding: randomBinding(), revision: 0, expires: session.expires, status: 'pending' });
        this.record = record; this.render();
        const registration = await browserOperation(navigator.serviceWorker.register('/push-worker.js', { updateViaCache: 'none' }));
        await browserOperation(navigator.serviceWorker.ready);
        if (!active()) { await clearResume(owner, record.binding); await unsubscribeIfUnused(); return; }
        let subscription = await browserOperation(registration.pushManager.getSubscription());
        if (!previous) {
          if (subscription) await browserOperation(subscription.unsubscribe()); subscription = null;
          const removed = await api(record, 'DELETE', '/api/push'); if (!removed.ok && removed.status !== 410) throw new PushNotice('Njoftimet nuk u përgatitën. Provo përsëri.');
        }
        if (!subscription) {
          const pending = registration.pushManager.subscribe({ userVisibleOnly: true, applicationServerKey: decodeKey(this.publicKey) });
          void pending.then(async () => { if (!await readResume()) await withLock(unsubscribeIfUnused); }).catch(() => {});
          subscription = await browserOperation(pending);
        }
        if (!active()) { await clearResume(owner, record.binding); await unsubscribeIfUnused(); return; }
        const keys = subscription.toJSON().keys;
        if (!keys?.p256dh || !keys.auth) throw new PushNotice('Njoftimet nuk u përgatitën në këtë pajisje.');
        const expires = Math.min(record.expires, subscription.expirationTime ?? record.expires);
        const metadataResponse = await api(record, 'GET', '/api/push');
        if (!metadataResponse.ok) throw new PushNotice('Njoftimet nuk u përgatitën. Provo përsëri.');
        const metadata = await metadataResponse.json();
        if (!Number.isSafeInteger(metadata.revision) || metadata.revision < 0) throw new PushNotice('Njoftimet nuk u përgatitën. Provo përsëri.');
        const prepared: PushResume = { ...record, revision: metadata.revision };
        if (!active() || !await updateResume(prepared)) return;
        this.record = prepared;
        const response = await api(prepared, 'POST', '/api/push', { revision: prepared.revision, binding: record.binding, endpoint: subscription.endpoint, p256dh: keys.p256dh, auth: keys.auth, expires_at: expires });
        if (!response.ok) throw new PushNotice('Njoftimet nuk u konfirmuan. Mund të provosh përsëri ose t’i çaktivizosh.');
        const result = await response.json();
        if (!Number.isFinite(result.expires_at) || result.expires_at <= Date.now()) throw new PushNotice('Njoftimet përfunduan.');
        const updated: PushResume = { ...prepared, expires: Math.min(expires, result.expires_at), status: 'enabled' };
        if (!active() || !await updateResume(updated)) { await clearResume(owner, record.binding); await unsubscribeIfUnused(); return; }
        this.record = updated; this.info = 'Njoftimet janë aktive vetëm deri në përfundimin e gatishmërisë.';
      });
    } catch (error) { if (active()) this.info = error instanceof PushNotice ? error.message : 'Njoftimet nuk u aktivizuan. Provo përsëri.'; }
    finally { this.busy = false; this.render(); this.changed(); }
  }
  async forget(token: string) {
    this.generation++;
    const own = this.record?.token === token;
    if (own) { this.record = null; this.recovering = false; this.resumeOnly = false; }
    try { await clearResume(token); if (supported()) await withLock(unsubscribeIfUnused); }
    catch { if (own) this.info = 'Heqja nga pajisja nuk u konfirmua. Afati në server vazhdon të zbatohet.'; }
    this.render(); this.changed();
  }
  private async disable() {
    const record = this.record; if (!record) return;
    // Revoke local display/resume before waiting on network or browser providers.
    this.generation++; const resumeOnly = this.recovering && !this.current();
    this.record = resumeOnly ? record : null; this.recovering = resumeOnly; this.resumeOnly = resumeOnly;
    this.info = 'Po çaktivizohen njoftimet…'; this.render(); this.changed();
    let localRemoved = false;
    try {
      await clearResume(record.token, record.binding); localRemoved = true;
      this.info = 'Njoftimet u çaktivizuan në këtë pajisje.';
      if (validResume(record)) { const response = await api(record, 'DELETE', '/api/push'); if (!response.ok && response.status !== 410) throw new PushNotice(); }
    } catch {
      this.info = localRemoved ? 'Njoftimet u fikën në këtë pajisje. Çaktivizimi në server nuk u konfirmua; lidhja në server përfundon vetë.' : 'Çaktivizimi në pajisje nuk u konfirmua. Provo përsëri.';
      if (!localRemoved) { this.record = record; this.resumeOnly = false; }
    }
    try { if (supported()) await withLock(unsubscribeIfUnused); } catch { /* Best-effort provider removal; server expiry is independent. */ }
    this.render();
  }
  private async recover() {
    const record = this.record; if (!record || this.busy || this.current()) return;
    const token = record.token;
    if (!validResume(record)) { await this.forget(token); return; }
    this.busy = true; this.recovering = true; this.render(); this.changed();
    try {
      const response = await api(record, 'GET', '/api/signal');
      if ([401, 410].includes(response.status)) { await this.forget(record.token); this.info = 'Gatishmëria përfundoi ose shërbimi u rinis.'; return; }
      if (!response.ok) throw new PushNotice();
      const signal = await response.json(), expires = Math.min(record.expires, signal.expires_at);
      if (!Number.isFinite(expires) || expires <= Date.now() || typeof signal.cell !== 'string' || !Number.isFinite(signal.radius_km) || !Number.isInteger(signal.availability_minutes)) throw new PushNotice();
      if (this.record?.binding !== record.binding || (!this.resumeOnly && (await readResume())?.binding !== record.binding)) return;
      this.recovering = false;
      this.resumed({ token: record.token, expires, confirmed: true, request: { cell: signal.cell, radius_km: signal.radius_km, availability_minutes: signal.availability_minutes } });
      if (this.resumeOnly) { this.record = null; this.resumeOnly = false; } else await this.checkRegistration();
    } catch { this.info = 'Rikthimi nuk u konfirmua. Kontrollo lidhjen dhe provo përsëri.'; }
    finally { this.busy = false; this.render(); this.changed(); }
  }
  private async cancelRecovery() {
    const record = this.record; if (!record || this.busy) return;
    try {
      const response = await api(record, 'DELETE', '/api/signal');
      if (!response.ok && response.status !== 410) throw new PushNotice();
      await this.forget(record.token); this.info = 'Gatishmëria u mbyll.';
    } catch { this.info = 'Mbyllja nuk u konfirmua. Provo përsëri; gatishmëria përfundon vetë.'; }
    this.render(); this.changed();
  }
  tick() { if (this.record && this.record.expires <= Date.now()) void this.forget(this.record.token); this.render(); }
}
