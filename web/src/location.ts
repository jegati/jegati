import { cellAt, type Grid } from './area';

// Only the coarse cell and a short local deadline escape this function. Reported
// device accuracy is a quality hint, not proof against a modified/spoofed client.
export function locateCell(grid: Grid, maxAccuracy: number, maxAgeSeconds: number): Promise<{ cell: string; expires: number }> {
  return new Promise((resolve, reject) => {
    const unavailable = () => reject(new Error('Vendndodhja nuk u mor. Lejo vendndodhjen nga pajisja dhe provo përsëri.'));
    // The browser timeout may exclude permission-prompt time; bound that too.
    let settled = false;
    const timer = setTimeout(() => { settled = true; unavailable(); }, 12_000);
    if (!navigator.geolocation) { clearTimeout(timer); unavailable(); return; }
    navigator.geolocation.getCurrentPosition(position => {
      if (settled) return; settled = true; clearTimeout(timer);
      const now = Date.now(), { latitude, longitude, accuracy } = position.coords;
      if (!Number.isFinite(position.timestamp) || position.timestamp > now || now - position.timestamp >= maxAgeSeconds * 1000 ||
          !Number.isFinite(accuracy) || accuracy < 0 || accuracy > maxAccuracy) {
        reject(new Error('Vendndodhja është e vjetër ose jo mjaftueshëm e saktë. Provo përsëri.')); return;
      }
      const cell = cellAt(grid, longitude, latitude);
      if (!cell) { reject(new Error('Aktualisht GATI mbulon vetëm Tiranën.')); return; }
      resolve({ cell, expires: position.timestamp + maxAgeSeconds * 1000 });
    }, () => { if (settled) return; settled = true; clearTimeout(timer); unavailable(); },
    { enableHighAccuracy: true, maximumAge: 0, timeout: 10_000 });
  });
}
