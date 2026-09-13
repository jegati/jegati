import { cellAt, type Grid } from './area';

// Only a coarse cell and deadline escape. Freshness is checked in the supported
// client; it is not authenticated GPS evidence at the server.
export function locateCell(grid: Grid, maxAccuracy: number, maxAgeSeconds: number, signal?: AbortSignal): Promise<{ cell: string; expires: number }> {
  return new Promise((resolve, reject) => {
    let settled = false;
    const stop = () => { if (settled) return false; settled = true; clearTimeout(timer); signal?.removeEventListener('abort', aborted); return true; };
    const unavailable = () => { if (stop()) reject(new Error('Vendndodhja nuk u mor. Lejo vendndodhjen nga pajisja dhe provo përsëri.')); };
    const aborted = () => { if (stop()) reject(new DOMException('Veprimi u anulua.', 'AbortError')); };
    // Include permission-prompt time in the app deadline. Late OS callbacks have
    // no effect after cancellation/timeout, even though the OS prompt may remain.
    const timer = setTimeout(unavailable, 12_000);
    if (signal?.aborted) { aborted(); return; }
    signal?.addEventListener('abort', aborted, { once: true });
    if (!navigator.geolocation) { unavailable(); return; }
    navigator.geolocation.getCurrentPosition(position => {
      if (!stop()) return;
      const now = Date.now(), { latitude, longitude, accuracy } = position.coords;
      if (!Number.isFinite(position.timestamp) || position.timestamp > now || now - position.timestamp >= maxAgeSeconds * 1000 || !Number.isFinite(accuracy) || accuracy < 0 || accuracy > maxAccuracy) {
        reject(new Error('Vendndodhja është e vjetër ose jo mjaftueshëm e saktë. Provo përsëri jashtë ose pranë një dritareje.')); return;
      }
      const cell = cellAt(grid, longitude, latitude);
      if (!cell) { reject(new Error('Aktualisht GATI mbulon vetëm Tiranën.')); return; }
      resolve({ cell, expires: position.timestamp + maxAgeSeconds * 1000 });
    }, unavailable, { enableHighAccuracy: true, maximumAge: 0, timeout: 10_000 });
  });
}
