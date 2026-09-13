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
let pollSeconds = 30, nextPoll = 0, failures = 0, operations = 0, cancelling = false;
let here = false, arrivalUntil = 0, nonceSeconds = 120;
let currentGrid: Grid | undefined;
let pendingNonce: { token: string; expires: number; issued: boolean } | null = null;
function begin() { operations++; busy = true; render(); }
function finish() { operations--; busy = operations > 0; render(); }
let map: maplibregl.Map | undefined;
let initialized = false;
interface Invitation { id: string; crossing: { id: string; label: string; point: [number, number] }; ends_at: number; state: string }
let invitation: Invitation | null = null, going = false, destinationMap: maplibregl.Map | undefined;
let renderedDestination: string | null = null;
function invitationView() {
  const card = el('invitation'); card.hidden = !invitation;
  if (!invitation) { destinationMap?.remove(); destinationMap = undefined; renderedDestination = null; return; }
  el('collective-title').textContent = invitation.state === 'jemi_ketu' ? 'JEMI KËTU.' : 'JEMI GATI.';
  el<HTMLButtonElement>('arrive').hidden = !going || here;
  el<HTMLButtonElement>('arrive').disabled = busy;
  el<HTMLButtonElement>('retract').hidden = !here;
  el<HTMLButtonElement>('retract').disabled = busy;
  el('arrival-status').textContent = here ? 'Mbërritja jote është konfirmuar përkohësisht.' : going ? 'Kur të mbërrish, konfirmo me vendndodhjen një herë.' : '';
  el('destination').textContent = invitation.crossing.label;
  el('gathering-time').textContent = `Takimi përfundon pas rreth ${Math.max(1, Math.ceil((invitation.ends_at - Date.now()) / 60_000))} minutash.`;
  el('going-status').textContent = going ? 'Ke zgjedhur të shkosh.' : 'Mund të zgjedhësh nëse do të shkosh.';
  el<HTMLButtonElement>('going').hidden = going;
  el<HTMLButtonElement>('going').disabled = busy;
  el<HTMLButtonElement>('decline').disabled = busy;
  el('decline').textContent = going ? 'Nuk po shkoj më' : 'JO TANI';
  if (renderedDestination === invitation.id) return;
  destinationMap?.remove(); renderedDestination = invitation.id; el('destination-map').removeAttribute('data-ready');
  try {
    destinationMap = new maplibregl.Map({ container: 'destination-map', center: invitation.crossing.point, zoom: 15,
      attributionControl: false, locale: { 'Map.Title': 'Harta e pikës së takimit' },
      style: { version: 8, sources: { roads: { type: 'geojson', data: '/api/map/roads' } }, layers: [
        { id: 'background', type: 'background', paint: { 'background-color': '#f0eee6' } },
        { id: 'roads', type: 'line', source: 'roads', paint: { 'line-color': '#bec3b8', 'line-width': 3 } }] } });
    destinationMap.on('idle', () => { el('destination-map').setAttribute('data-ready', 'true'); });
    // This marker is the shared mapped destination, never a person's position.
    const marker = document.createElement('span'); marker.className = 'destination-marker'; marker.textContent = '🦩'; marker.setAttribute('role', 'img'); marker.setAttribute('aria-label', 'Pika e takimit');
    new maplibregl.Marker({ element: marker }).setLngLat(invitation.crossing.point).addTo(destinationMap);
    destinationMap.addControl(new maplibregl.AttributionControl({ compact: false, customAttribution: '© OpenStreetMap · ODbL' }));
    destinationMap.getCanvas().setAttribute('aria-label', 'Pika e takimit pranë vendkalimit për këmbësorë');
  } catch { el('destination-map').textContent = 'Harta nuk mund të hapet në këtë pajisje.'; }
}
const message = (text: string) => { status.textContent = text; };
function render() {
  form.hidden = session !== null; active.hidden = session === null;
  cancel.disabled = cancelling; retry.disabled = busy;
  ready.disabled = busy || !selected;
  retry.hidden = !session || session.confirmed;
  el('active-title').textContent = session?.confirmed ? (here ? 'JAM KËTU.' : 'JAM GATI.') : 'Po kontrollojmë gatishmërinë…';
  invitationView();
  if (session) el('remaining').textContent = `Përfundon pas rreth ${Math.max(1, Math.ceil((session.expires - Date.now()) / 60_000))} minutash.`;
}
function end(text: string) {
  session = null; sessions.clear(); selected = null; invitation = null; going = false; here = false; arrivalUntil = 0; pendingNonce = null;
  map?.getSource<maplibregl.GeoJSONSource>('selection')?.setData({ type: 'FeatureCollection', features: [] });
  el('area-status').textContent = 'Ende nuk ke zgjedhur zonë.';
  render(); map?.resize(); message(text);
}
async function request(method: string, path: string, body?: unknown, extraHeaders: Record<string, string> = {}) {
  if (!session || session.expires <= Date.now()) { end('Gatishmëria përfundoi.'); throw new Error('expired'); }
  return fetch(path, { method, credentials: 'omit', cache: 'no-store', signal: AbortSignal.timeout(10_000),
    headers: { Authorization: `Bearer ${session.token}`, ...(body ? { 'Content-Type': 'application/json' } : {}), ...extraHeaders },
    ...(body ? { body: JSON.stringify(body) } : {}) });
}
async function sync(create = false) {
  if (!session || busy) return;
  begin();
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
      applySignal(signal);
      failures = 0; message('Gatishmëria jote është aktive.');
    }
  } catch { if (!session) return; failures = Math.min(failures + 1, 4); message('Lidhja u ndërpre. Gatishmëria mund të jetë aktive; provo përsëri.'); }
  finally { finish(); nextPoll = Date.now() + pollSeconds * 1000 * 2 ** failures * (1 + Math.random() * .2); render(); }
}
form.addEventListener('submit', event => {
  event.preventDefault();
  if (busy || !selected || session) return;
  session = sessions.create({ cell: selected, radius_km: Number(radius.value), availability_minutes: Number(duration.value) });
  sessions.save(session); void sync(true);
});
retry.onclick = () => { void sync(true); };
cancel.onclick = async () => {
  if (cancelling || !session) return;
  cancelling = true; begin();
  try {
    const response = await request('DELETE', '/api/signal');
    if (!response.ok && response.status !== 410) throw new Error('unavailable');
    end('Gatishmëria u mbyll.');
  } catch { if (session) message('Mbyllja nuk u konfirmua. Provo përsëri; gatishmëria përfundon vetë në afatin e saj.'); }
  finally { cancelling = false; finish(); }
};

async function intent(action: 'going' | 'decline') {
  if (!session || !invitation || busy) return;
  begin();
  try {
    const response = await request('POST', `/api/${action}`, { gathering_id: invitation.id });
    if (response.status === 410) { invitation = null; going = false; message('Ky takim nuk është më i hapur për t’u bashkuar. Gatishmëria jote mund të vazhdojë.'); return; }
    if (!response.ok) throw new Error('unavailable');
    const signal = await response.json();
    applySignal(signal);
    if (session) message(action === 'going' ? 'Zgjedhja u ruajt.' : 'Në rregull. Gatishmëria jote vazhdon.');
  } catch { if (session) message('Zgjedhja nuk u konfirmua. Provo përsëri.'); }
  finally { finish(); }
}
el<HTMLButtonElement>('going').onclick = () => { void intent('going'); };
el<HTMLButtonElement>('decline').onclick = () => { void intent('decline'); };

function applySignal(signal: { invitation?: Invitation; state: string; arrival_until?: number }) {
  if (!session) return;
  if (invitation?.id !== signal.invitation?.id) pendingNonce = null;
  invitation = signal.invitation ?? null; going = signal.state === 'going' || signal.state === 'here';
  here = signal.state === 'here'; arrivalUntil = signal.arrival_until ?? 0;
}
function locateOnce(): Promise<GeolocationPosition> {
  return new Promise((resolve, reject) => {
    // Bound even time spent waiting for the browser permission prompt.
    const timer = setTimeout(() => reject(new Error('location unavailable')), 12_000);
    if (!navigator.geolocation) { clearTimeout(timer); reject(new Error('location unavailable')); return; }
    navigator.geolocation.getCurrentPosition(position => { clearTimeout(timer); resolve(position); }, () => { clearTimeout(timer); reject(new Error('location unavailable')); }, { enableHighAccuracy: false, maximumAge: 0, timeout: 10_000 });
  });
}
el<HTMLButtonElement>('arrive').onclick = async () => {
  if (busy || !session || !invitation || !currentGrid) return;
  const owner = session.token; begin();
  try {
    if (!pendingNonce || pendingNonce.expires <= Date.now()) pendingNonce = { token: sessions.randomToken(), expires: Date.now() + nonceSeconds * 1000, issued: false };
    const nonce = pendingNonce;
    const headers = { 'X-Gati-Arrival-Nonce': nonce.token };
    if (!nonce.issued) {
      const issued = await request('POST', '/api/arrival-nonce', undefined, headers);
      if (!issued.ok) throw new Error('unavailable');
      const challenge = await issued.json();
      if (!Number.isFinite(challenge.expires_at) || challenge.expires_at <= Date.now()) throw new Error('expired challenge');
      nonce.expires = Math.min(nonce.expires, challenge.expires_at); nonce.issued = true;
    }
    if (session?.token !== owner) return;
    const position = await locateOnce();
    if (session?.token !== owner) return;
    const cell = cellAt(currentGrid, position.coords.longitude, position.coords.latitude);
    if (!cell) { message('Aktualisht GATI mbulon vetëm Tiranën.'); return; }
    const response = await request('POST', '/api/arrival', { cell }, headers);
    if (!response.ok) throw new Error('unavailable');
    if (session?.token !== owner) return;
    applySignal(await response.json()); pendingNonce = null;
    message('Mbërritja u konfirmua përkohësisht.');
  } catch { if (session?.token === owner) message('Mbërritja nuk u konfirmua. Kontrollo vendndodhjen dhe provo përsëri pranë pikës së takimit.'); }
  finally { finish(); }
};
el<HTMLButtonElement>('retract').onclick = async () => {
  if (busy || !session) return; begin();
  try {
    const response = await request('DELETE', '/api/arrival'); if (!response.ok) throw new Error('unavailable');
    applySignal(await response.json()); pendingNonce = null; if (session) message('Konfirmimi i mbërritjes u hoq.');
  } catch { if (session) message('Heqja nuk u konfirmua. Provo përsëri.'); }
  finally { finish(); }
};

async function start() {
  const [configResponse, gridResponse] = await Promise.all(['/api/config', '/api/geography'].map(url => fetch(url, { credentials: 'omit', cache: 'no-store' })));
  if (!configResponse.ok || !gridResponse.ok) throw new Error('unavailable');
  const configuration = await configResponse.json();
  if (configuration.schema_version !== 4) throw new Error('unsupported schema');
  const config = configuration.config, grid: Grid = await gridResponse.json();
  pollSeconds = config.notifications.foreground_poll_seconds;
  nonceSeconds = config.arrivals.nonce_seconds; currentGrid = grid;
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
    if (pendingNonce && pendingNonce.expires <= Date.now()) pendingNonce = null;
    if (here && arrivalUntil <= Date.now()) { here = false; arrivalUntil = 0; }
    if (invitation && invitation.ends_at <= Date.now()) { invitation = null; going = false; }
    if (session && session.expires <= Date.now()) end('Gatishmëria përfundoi.');
    if (session && session.confirmed && !document.hidden && Date.now() >= nextPoll) void sync();
    if (session) render();
  }, 1000);
  document.addEventListener('visibilitychange', () => { if (!document.hidden && session) void sync(); });
}
setInterval(() => {
  if (session && session.expires <= Date.now()) {
    if (initialized) end('Gatishmëria përfundoi.');
    else { session = null; sessions.clear(); active.hidden = true; message('Gatishmëria përfundoi.'); }
  }
}, 1000);
start().catch(() => { message('GATI nuk u hap. Kontrollo lidhjen dhe rifresko faqen.'); });
