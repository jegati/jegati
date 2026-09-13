import './style.css';
import 'maplibre-gl/dist/maplibre-gl.css';
import * as maplibregl from 'maplibre-gl';
import workerURL from 'maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url';
maplibregl.setWorkerUrl(workerURL);
import { cellAt, polygon, type Grid } from './area';
import * as sessions from './session';

const el = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
const form = el<HTMLFormElement>('willingness'), active = el('active'), status = el('status');
const duration = el<HTMLSelectElement>('duration'), radius = el<HTMLSelectElement>('radius');
const ready = el<HTMLButtonElement>('ready'), cancel = el<HTMLButtonElement>('cancel'), retry = el<HTMLButtonElement>('retry');
let session = sessions.restore(), selected: string | null = null, busy = false;
let pollSeconds = 30, nextPoll = 0, failures = 0;
let map: maplibregl.Map | undefined;
let initialized = false;
const message = (text: string) => { status.textContent = text; };
function render() {
  form.hidden = session !== null; active.hidden = session === null;
  cancel.disabled = busy; retry.disabled = busy;
  ready.disabled = busy || !selected;
  retry.hidden = !session || session.confirmed;
  el('active-title').textContent = session?.confirmed ? 'JAM GATI.' : 'Po kontrollojmë gatishmërinë…';
  if (session) el('remaining').textContent = `Përfundon pas rreth ${Math.max(1, Math.ceil((session.expires - Date.now()) / 60_000))} minutash.`;
}
function end(text: string) {
  session = null; sessions.clear(); selected = null;
  map?.getSource<maplibregl.GeoJSONSource>('selection')?.setData({ type: 'FeatureCollection', features: [] });
  el('area-status').textContent = 'Ende nuk ke zgjedhur zonë.';
  render(); map?.resize(); message(text);
}
async function request(method: string, path: string, body?: sessions.Willingness) {
  if (!session || session.expires <= Date.now()) { end('Gatishmëria përfundoi.'); throw new Error('expired'); }
  return fetch(path, { method, credentials: 'omit', cache: 'no-store', signal: AbortSignal.timeout(10_000),
    headers: { Authorization: `Bearer ${session.token}`, ...(body ? { 'Content-Type': 'application/json' } : {}) },
    ...(body ? { body: JSON.stringify(body) } : {}) });
}
async function sync(create = false) {
  if (!session || busy) return;
  busy = true; render();
  try {
    const response = await request(create ? 'POST' : 'GET', create ? '/api/signals' : '/api/signal', create ? session.request : undefined);
    if (response.status === 410) { end('Gatishmëria përfundoi.'); return; }
    if (!response.ok) throw new Error('unavailable');
    const signal = await response.json();
    if (!Number.isFinite(signal.expires_at) || !Number.isFinite(signal.created_at) || signal.expires_at <= signal.created_at) throw new Error('invalid deadline');
    if (session) {
      // Never extend the local maximum on retries or reload. Server enforces its
      // own clock/deadline even if this browser's wall clock has been changed.
      session.expires = Math.min(session.expires, signal.expires_at);
      session.confirmed = true; sessions.save(session);
      failures = 0; message('Gatishmëria jote është aktive.');
    }
  } catch { failures = Math.min(failures + 1, 4); message('Lidhja u ndërpre. Gatishmëria mund të jetë aktive; provo përsëri.'); }
  finally { busy = false; nextPoll = Date.now() + pollSeconds * 1000 * 2 ** failures * (1 + Math.random() * .2); render(); }
}
form.addEventListener('submit', event => {
  event.preventDefault();
  if (busy || !selected || session) return;
  session = sessions.create({ cell: selected, radius_km: Number(radius.value), availability_minutes: Number(duration.value) });
  sessions.save(session); void sync(true);
});
retry.onclick = () => { void sync(true); };
cancel.onclick = async () => {
  if (busy || !session) return;
  busy = true; render();
  try {
    const response = await request('DELETE', '/api/signal');
    if (!response.ok && response.status !== 410) throw new Error('unavailable');
    end('Gatishmëria u mbyll.');
  } catch { message('Mbyllja nuk u konfirmua. Provo përsëri; gatishmëria përfundon vetë në afatin e saj.'); }
  finally { busy = false; render(); }
};

async function start() {
  const [configResponse, gridResponse] = await Promise.all(['/api/config', '/api/geography'].map(url => fetch(url, { credentials: 'omit', cache: 'no-store' })));
  if (!configResponse.ok || !gridResponse.ok) throw new Error('unavailable');
  const configuration = await configResponse.json();
  if (configuration.schema_version !== 2) throw new Error('unsupported schema');
  const config = configuration.config, grid: Grid = await gridResponse.json();
  pollSeconds = config.notifications.foreground_poll_seconds;
  for (const minutes of config.availability.choices_minutes) duration.add(new Option(`${minutes} minuta`, String(minutes)));
  for (const km of config.geography.travel_radius_choices_km) radius.add(new Option(`${km} km`, String(km)));
  radius.value = String(config.geography.travel_radius_choices_km.includes(3) ? 3 : config.geography.travel_radius_choices_km[0]);
  initialized = true; render();
  try {
    map = new maplibregl.Map({ container: 'map', center: [19.818, 41.327], zoom: 12,
      maxBounds: [[grid.west, grid.south], [grid.east, grid.north]], minZoom: 10, maxZoom: 16,
      attributionControl: false, locale: { 'Map.Title': 'Harta e Tiranës' },
      style: { version: 8, sources: { roads: { type: 'geojson', data: '/api/map/roads', attribution: '© OpenStreetMap · ODbL' }, selection: { type: 'geojson', data: { type: 'FeatureCollection', features: [] } } },
        layers: [{ id: 'background', type: 'background', paint: { 'background-color': '#f0eee6' } },
          { id: 'roads', type: 'line', source: 'roads', paint: { 'line-color': '#c4c5ba', 'line-width': 2 } },
          { id: 'selection', type: 'fill', source: 'selection', paint: { 'fill-color': '#ce5c76', 'fill-opacity': .28 } },
          { id: 'outline', type: 'line', source: 'selection', paint: { 'line-color': '#a03451', 'line-width': 2 } }] } });
    map.addControl(new maplibregl.AttributionControl({ compact: false, customAttribution: '<a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener noreferrer">Të dhënat e hartës · OpenStreetMap</a>' }));
    map.getCanvas().setAttribute('aria-label', 'Harta e Tiranës. Përdor shigjetat për të lëvizur, plus dhe minus për zmadhimin.');
    map.on('idle', () => { el('map').setAttribute('data-ready', 'true'); });
    map.on('click', event => choose(event.lngLat.lng, event.lngLat.lat));
    map.on('error', () => { el('area-status').textContent = 'Harta nuk u hap plotësisht. Mund të përdorësh vendndodhjen një herë.'; });
    el<HTMLButtonElement>('map-center').onclick = () => { const center = map!.getCenter(); choose(center.lng, center.lat); };
  } catch {
    el('map').textContent = 'Harta nuk mund të hapet në këtë pajisje.';
    el<HTMLButtonElement>('map-center').disabled = true;
  }
  function choose(lon: number, lat: number) {
    if (session || busy) return;
    const cell = cellAt(grid, lon, lat);
    if (!cell) { selected = null; render(); map?.getSource<maplibregl.GeoJSONSource>('selection')?.setData({ type: 'FeatureCollection', features: [] }); el('area-status').textContent = 'Aktualisht GATI mbulon vetëm Tiranën. Zgjidh një zonë brenda hartës.'; return; }
    selected = cell;
    const coordinates = polygon(grid, cell);
    const draw = () => map?.getSource<maplibregl.GeoJSONSource>('selection')?.setData({ type: 'Feature', properties: {}, geometry: { type: 'Polygon', coordinates: [coordinates] } });
    if (map?.isStyleLoaded()) draw(); else map?.once('load', draw);
    // Camera uses the coarse center too; exact device coordinates are discarded.
    map?.jumpTo({ center: [(coordinates[0][0] + coordinates[2][0]) / 2, (coordinates[0][1] + coordinates[2][1]) / 2] });
    el('area-status').textContent = 'Zona u zgjodh. Pozicioni yt i saktë nuk dërgohet.'; render();
  }
  el<HTMLButtonElement>('location').onclick = () => {
    if (!navigator.geolocation) { el('area-status').textContent = 'Vendndodhja nuk mbështetet. Zgjidh zonën në hartë.'; return; }
    const button = el<HTMLButtonElement>('location'); button.disabled = true;
    navigator.geolocation.getCurrentPosition(position => { button.disabled = false; choose(position.coords.longitude, position.coords.latitude); }, () => {
      button.disabled = false; el('area-status').textContent = 'Vendndodhja nuk u mor. Mund të zgjedhësh zonën në hartë.';
    }, { enableHighAccuracy: false, maximumAge: 0, timeout: 10_000 });
  };
  message('');
  if (session) await sync();
  setInterval(() => {
    if (session && session.expires <= Date.now() && !busy) end('Gatishmëria përfundoi.');
    if (session && session.confirmed && !document.hidden && Date.now() >= nextPoll) void sync();
    if (session) render();
  }, 1000);
  document.addEventListener('visibilitychange', () => { if (!document.hidden && session) void sync(); });
}
setInterval(() => {
  if (session && session.expires <= Date.now() && !busy) {
    if (initialized) end('Gatishmëria përfundoi.');
    else { session = null; sessions.clear(); active.hidden = true; message('Gatishmëria përfundoi.'); }
  }
}, 1000);
start().catch(() => { message('GATI nuk u hap. Kontrollo lidhjen dhe rifresko faqen.'); });
