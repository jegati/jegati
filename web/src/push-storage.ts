// Only an explicit notification opt-in may create this database/record. No location
// or participation history is stored here. The worker reads this same v1 schema.
export interface PushResume { token: string; binding: string; revision: number; expires: number; status: 'pending' | 'enabled' }
const database = 'gati-push-v1', collection = 'resume';
export function validResume(value: unknown): value is PushResume {
  const v = value as PushResume | null;
  return !!v && /^[A-Za-z0-9_-]{43}$/.test(v.token) && /^[a-f0-9]{64}$/.test(v.binding) && Number.isSafeInteger(v.revision) && v.revision >= 0 && Number.isFinite(v.expires) && v.expires > Date.now() && v.expires <= Date.now() + 120 * 60000 && (v.status === 'pending' || v.status === 'enabled');
}
async function open(create: boolean): Promise<IDBDatabase | null> {
  return new Promise((resolve, reject) => {
    let absent = false, blocked = false;
    const request = indexedDB.open(database, 1);
    request.onupgradeneeded = () => {
      if (!create) { absent = true; request.transaction!.abort(); }
      else request.result.createObjectStore(collection);
    };
    request.onsuccess = () => { if (blocked) { request.result.close(); return; } request.result.onversionchange = () => request.result.close(); resolve(request.result); };
    request.onerror = () => absent ? resolve(null) : reject(new Error('Ruajtja për njoftimet nuk është e mundur.'));
    request.onblocked = () => { blocked = true; reject(new Error('Mbyll skedat e tjera të GATI dhe provo përsëri.')); };
  });
}
export async function readResume(): Promise<PushResume | null> {
  const db = await open(false); if (!db) return null;
  try {
    return await new Promise((resolve, reject) => {
      const transaction = db.transaction(collection, 'readwrite'); const request = transaction.objectStore(collection).get('current');
      let value: PushResume | null = null;
      request.onsuccess = () => { if (validResume(request.result)) value = request.result; else transaction.objectStore(collection).delete('current'); };
      transaction.oncomplete = () => resolve(value); transaction.onerror = () => reject(new Error('Ruajtja për njoftimet nuk u lexua.'));
    });
  } finally { db.close(); }
}
export async function claimResume(value: PushResume): Promise<PushResume> {
  const db = await open(true); if (!db) throw new Error('Ruajtja për njoftimet nuk është e mundur.');
  try {
    return await new Promise((resolve, reject) => {
      const tx = db.transaction(collection, 'readwrite'), store = tx.objectStore(collection), request = store.get('current');
      let chosen = value;
      request.onsuccess = () => {
        const old = request.result;
        if (validResume(old)) { if (old.token !== value.token) { tx.abort(); return; } chosen = old; }
        if (!validResume(chosen)) { tx.abort(); return; } store.put(chosen, 'current');
      };
      tx.oncomplete = () => resolve(chosen); tx.onabort = tx.onerror = () => reject(new Error('Njoftimet janë të lidhura me një gatishmëri tjetër. Çaktivizoji së pari.'));
    });
  } finally { db.close(); }
}
export async function updateResume(value: PushResume): Promise<boolean> {
  const db = await open(false); if (!db) return false;
  try {
    return await new Promise((resolve, reject) => {
      const tx = db.transaction(collection, 'readwrite'), store = tx.objectStore(collection), request = store.get('current'); let updated = false;
      request.onsuccess = () => { const old = request.result; if (validResume(old) && validResume(value) && old.token === value.token && old.binding === value.binding) { store.put({ ...value, expires: Math.min(value.expires, old.expires) }, 'current'); updated = true; } };
      tx.oncomplete = () => resolve(updated); tx.onerror = () => reject(new Error('Ruajtja e njoftimeve dështoi.'));
    });
  } finally { db.close(); }
}
export async function clearResume(token: string, binding?: string): Promise<boolean> {
  const db = await open(false); if (!db) return false;
  try {
    return await new Promise((resolve, reject) => {
      const tx = db.transaction(collection, 'readwrite'), store = tx.objectStore(collection), request = store.get('current'); let removed = false;
      request.onsuccess = () => { if (request.result?.token === token && (!binding || request.result.binding === binding)) { store.delete('current'); removed = true; } };
      tx.oncomplete = () => resolve(removed); tx.onerror = () => reject(new Error('Njoftimet nuk u hoqën nga pajisja.'));
    });
  } finally { db.close(); }
}
