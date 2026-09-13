import './style.css';
import { ActivityView } from './activity';
import 'maplibre-gl/dist/maplibre-gl.css';
import * as maplibregl from 'maplibre-gl';
import workerURL from 'maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url';
maplibregl.setWorkerUrl(workerURL);
import { polygon, type Grid } from './area';
import * as sessions from './session';
import { locateCell } from './location';

const el = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
const form = el<HTMLFormElement>('willingness'), active = el('active'), status = el('status');
const duration = el<HTMLSelectElement>('duration'), radius = el<HTMLSelectElement>('radius');
const ready = el<HTMLButtonElement>('ready'), cancel = el<HTMLButtonElement>('cancel'), retry = el<HTMLButtonElement>('retry');
let session = sessions.restore(), busy = false;
let pollSeconds = 30, nextPoll = 0, failures = 0, operations = 0, cancelling = false;
let here = false, arrivalUntil = 0, nonceSeconds = 120;
let currentGrid: Grid | undefined;
let locationMaxAccuracy = 100, locationMaxAge = 60;
let locationRequest: AbortController | null = null;
let drawArea: (cell: string) => void = () => {};
let pendingNonce: { token: string; expires: number; issued: boolean } | null = null;
function begin() { operations++; busy = true; render(); }
function finish() { operations--; busy = operations > 0; render(); }
let map: maplibregl.Map | undefined;
let initialized = false;
let activityView: ActivityView | undefined;
let preview: { candidate: sessions.Session; invitation: Invitation; expires: number; existing: boolean } | null = null;
let previewMap: maplibregl.Map | undefined;
let joinTarget: { id: string; ends_at: number } | null = null;
interface Invitation { id: string; intersection: { id: string; label: string; point: [number, number] }; ends_at: number; state: string }
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
  el('arrival-status').textContent = here ? `Mbërritja jote është konfirmuar për rreth ${Math.max(1, Math.ceil((arrivalUntil - Date.now()) / 60000))} minuta të tjera.` : going ? 'Kur të mbërrish, konfirmo me vendndodhjen një herë.' : '';
  el('destination').textContent = invitation.intersection.label;
  el('gathering-time').textContent = `Takimi përfundon pas rreth ${Math.max(1, Math.ceil((invitation.ends_at - Date.now()) / 60_000))} minutash.`;
  el('going-status').textContent = going ? 'Ke zgjedhur të shkosh.' : 'Mund të zgjedhësh nëse do të shkosh.';
  el<HTMLButtonElement>('going').hidden = going;
  el<HTMLButtonElement>('going').disabled = busy;
  el<HTMLButtonElement>('decline').disabled = busy;
  el('decline').textContent = going ? 'Nuk po shkoj më' : 'JO TANI';
  if (renderedDestination === invitation.id) return;
  destinationMap?.remove(); renderedDestination = invitation.id; el('destination-map').removeAttribute('data-ready');
  try {
    destinationMap = new maplibregl.Map({ container: 'destination-map', center: invitation.intersection.point, zoom: 15,
      attributionControl: false, locale: { 'Map.Title': 'Harta e pikës së takimit' },
      style: { version: 8, sources: { roads: { type: 'geojson', data: '/api/map/roads' } }, layers: [
        { id: 'background', type: 'background', paint: { 'background-color': '#f0eee6' } },
        { id: 'roads', type: 'line', source: 'roads', paint: { 'line-color': '#bec3b8', 'line-width': 3 } }] } });
    destinationMap.on('idle', () => { el('destination-map').setAttribute('data-ready', 'true'); });
    // This marker is the shared mapped destination, never a person's position.
    const marker = document.createElement('span'); marker.className = 'destination-marker'; marker.textContent = '🦩'; marker.setAttribute('role', 'img'); marker.setAttribute('aria-label', 'Pika e takimit');
    new maplibregl.Marker({ element: marker }).setLngLat(invitation.intersection.point).addTo(destinationMap);
    destinationMap.addControl(new maplibregl.AttributionControl({ compact: false, customAttribution: '© OpenStreetMap · ODbL' }));
    destinationMap.getCanvas().setAttribute('aria-label', 'Pika e takimit pranë kryqëzimit');
  } catch { el('destination-map').textContent = 'Harta nuk mund të hapet në këtë pajisje.'; }
}
const message = (text: string) => { status.textContent = text; };
function render() {
  form.hidden = session !== null; active.hidden = session === null;
  cancel.disabled = cancelling; retry.disabled = busy;
  ready.disabled = busy || !initialized || !navigator.onLine;
  duration.disabled = busy; radius.disabled = busy;
  el<HTMLButtonElement>('join-back').disabled = busy;
  el('location-cancel').hidden = !locationRequest;
  el<HTMLButtonElement>('refresh-status').disabled = busy || !navigator.onLine;
  el('connection-status').textContent = !navigator.onLine ? 'Nuk ka lidhje. Veprimet nuk dërgohen; gatishmëria përfundon në afatin e saj.' : failures ? 'Lidhja me GATI nuk u konfirmua. Mund të provosh përsëri.' : '';
  el('waiting-guidance').textContent = !session?.confirmed ? 'Po kontrollojmë nëse veprimi u pranua.' : here ? 'Mbërritja është e përkohshme. Mund ta heqësh konfirmimin kur të duash.' : going ? 'Ke zgjedhur të shkosh. Kur të mbërrish, shtyp JAM KËTU.' : invitation ? 'Zgjidh nëse dëshiron të shkosh. JO TANI e mban gatishmërinë aktive.' : 'Gatishmëria jote është aktive. Po presim një takim të përshtatshëm. Mungesa e shifrave publike nuk do të thotë që je vetëm.';
  retry.hidden = !session || session.confirmed;
  el('active-title').textContent = session?.confirmed ? (here ? 'JAM KËTU.' : 'JAM GATI.') : 'Po kontrollojmë gatishmërinë…';
  invitationView();
  activityView?.render();
  if (preview) {
    const expired = preview.expires <= Date.now();
    el<HTMLButtonElement>('public-join-confirm').disabled = busy || expired || !navigator.onLine;
    el('preview-status').textContent = expired ? 'Parapamja përfundoi. Mbylle dhe kontrollo sërish pikën e takimit.' : !navigator.onLine ? 'Nuk ka lidhje. Provo përsëri kur të rikthehet.' : '';
  }
  ready.textContent = locationRequest ? 'Po merret vendndodhja…' : joinTarget ? 'Shiko pikën e takimit' : 'JAM GATI';
  el('join-choice').hidden = !joinTarget; el('join-back').hidden = !joinTarget;
  if (joinTarget) el('join-choice').textContent = 'Shiko pikën e takimit me vendndodhjen nga pajisja, pastaj zgjidh nëse do të shkosh.';
  if (session) el('remaining').textContent = `Përfundon pas rreth ${Math.max(1, Math.ceil((session.expires - Date.now()) / 60_000))} minutash.`;
}
function end(text: string) {
  closePreview(); locationRequest?.abort(); joinTarget = null; session = null; sessions.clear(); invitation = null; going = false; here = false; arrivalUntil = 0; pendingNonce = null;
  map?.getSource<maplibregl.GeoJSONSource>('selection')?.setData({ type: 'FeatureCollection', features: [] });
  el('area-status').textContent = 'Vendndodhja ende nuk është marrë.';
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
    const response = await request(create ? 'POST' : 'GET', create ? (session.joinGathering ? '/api/join' : '/api/signals') : '/api/signal', create ? { ...session.request, ...(session.joinGathering ? { gathering_id: session.joinGathering } : {}) } : undefined);
    if (session?.joinGathering && [409, 410].includes(response.status)) { end('Nuk mund të bashkohesh në këtë takim. Mund të shprehësh sërish gatishmërinë.'); return; }
    if (response.status === 410) { end('Gatishmëria përfundoi ose shërbimi u rinis. Mund të shprehësh sërish gatishmërinë.'); return; }
    if (!response.ok) throw new Error('unavailable');
    const signal = await response.json();
    if (!Number.isFinite(signal.expires_at) || !Number.isFinite(signal.created_at) || signal.expires_at <= signal.created_at) throw new Error('invalid deadline');
    if (session) {
      // Never extend the local maximum on retries or reload. Server enforces its
      // own clock/deadline even if this browser's wall clock has been changed.
      session.expires = Math.min(session.expires, signal.expires_at);
      session.confirmed = true; delete session.joinGathering; joinTarget = null; sessions.save(session);
      applySignal(signal);
      failures = 0; message(here ? 'Mbërritja jote është konfirmuar përkohësisht.' : going ? 'Ke zgjedhur të shkosh.' : 'Gatishmëria jote është aktive.');
    }
  } catch { if (!session) return; failures = Math.min(failures + 1, 4); message('Lidhja u ndërpre. Gatishmëria mund të jetë aktive; provo përsëri.'); }
  finally { finish(); nextPoll = Date.now() + pollSeconds * 1000 * 2 ** failures * (1 + Math.random() * .2); render(); }
}
form.addEventListener('submit', async event => {
  event.preventDefault();
  if (busy || session || !currentGrid || !navigator.onLine) return;
  const target = joinTarget;
  const choices = { radius_km: Number(radius.value), availability_minutes: Number(duration.value) };
  const controller = new AbortController(); locationRequest = controller; begin();
  el('area-status').textContent = 'Po merret vendndodhja nga pajisja…';
  let fix: { cell: string; expires: number } | undefined;
  try {
    fix = await locateCell(currentGrid, locationMaxAccuracy, locationMaxAge, controller.signal);
    drawArea(fix.cell);
  } catch (error) {
    const text = error instanceof DOMException && error.name === 'AbortError' ? 'Veprimi u anulua. Nuk u dërgua gatishmëri.' : error instanceof Error ? error.message : 'Vendndodhja nuk u mor. Provo përsëri.';
    el('area-status').textContent = text; message(text);
  } finally { locationRequest = null; finish(); }
  if (!fix || controller.signal.aborted) return;
  const candidate = sessions.create({ cell: fix.cell, ...choices });
  if (target) await loadPreview(target.id, candidate, Math.min(fix.expires, candidate.expires), false);
  else { session = candidate; sessions.save(session); void sync(true); }
});
el<HTMLButtonElement>('location-cancel').onclick = () => locationRequest?.abort();
el<HTMLButtonElement>('public-join-cancel').onclick = closePreview;
el<HTMLDialogElement>('public-join-dialog').addEventListener('cancel', closePreview);
el<HTMLButtonElement>('public-join-confirm').onclick = () => {
  const chosen = preview;
  if (!chosen || chosen.expires <= Date.now() || !navigator.onLine || busy) return;
  closePreview();
  if (chosen.existing) {
    if (session?.token === chosen.candidate.token) void intent('going', chosen.invitation.id);
  } else if (!session) {
    session = { ...chosen.candidate, expires: Date.now() + chosen.candidate.request.availability_minutes * 60000, joinGathering: chosen.invitation.id };
    sessions.save(session); void sync(true);
  }
};
el<HTMLButtonElement>('join-back').onclick = () => { joinTarget = null; closePreview(); render(); };
el<HTMLButtonElement>('refresh-status').onclick = () => { void sync(!session?.confirmed); };
function closePreview() {
  preview = null; previewMap?.remove(); previewMap = undefined;
  el<HTMLDialogElement>('public-join-dialog').close();
  el('preview-destination').textContent = ''; el('preview-time').textContent = ''; el('preview-status').textContent = '';
}
async function loadPreview(id: string, candidate: sessions.Session, expires: number, existing: boolean) {
  if (busy) return; begin();
  try {
    const response = await fetch('/api/gathering-preview', { method: 'POST', credentials: 'omit', cache: 'no-store', signal: AbortSignal.timeout(10000), headers: { Authorization: `Bearer ${candidate.token}`, 'Content-Type': 'application/json' }, body: JSON.stringify({ ...candidate.request, gathering_id: id }) });
    if (!response.ok) throw new Error('Ky takim nuk është i arritshëm ose nuk është më i hapur. Mund të zgjedhësh një tjetër.');
    const result = await response.json();
    if (!Number.isFinite(result.preview_expires_at) || result.preview_expires_at <= Date.now() || expires <= Date.now() || (existing && session?.token !== candidate.token)) throw new Error('Parapamja përfundoi. Provo përsëri.');
    closePreview(); preview = { candidate, invitation: result.invitation, expires: Math.min(expires, result.preview_expires_at), existing };
    el('preview-destination').textContent = preview.invitation.intersection.label;
    el('preview-time').textContent = `Takimi përfundon pas rreth ${Math.max(1, Math.ceil((preview.invitation.ends_at - Date.now()) / 60000))} minutash.`;
    el<HTMLButtonElement>('public-join-confirm').disabled = false;
    el<HTMLDialogElement>('public-join-dialog').showModal();
    try {
      previewMap = new maplibregl.Map({ container: 'preview-map', center: preview.invitation.intersection.point, zoom: 15, attributionControl: false, locale: { 'Map.Title': 'Parapamja e takimit' }, style: { version: 8, sources: { roads: { type: 'geojson', data: '/api/map/roads' } }, layers: [{ id: 'background', type: 'background', paint: { 'background-color': '#f0eee6' } }, { id: 'roads', type: 'line', source: 'roads', paint: { 'line-color': '#bec3b8', 'line-width': 3 } }] } });
      const marker = document.createElement('span'); marker.className = 'destination-marker'; marker.textContent = '🦩'; marker.setAttribute('aria-label', 'Pika e takimit');
      new maplibregl.Marker({ element: marker }).setLngLat(preview.invitation.intersection.point).addTo(previewMap);
      previewMap.addControl(new maplibregl.AttributionControl({ compact: false, customAttribution: '© OpenStreetMap · ODbL' }));
    } catch { el('preview-map').textContent = 'Harta nuk mund të hapet në këtë pajisje.'; }
  } catch (error) { message(error instanceof Error ? error.message : 'Parapamja nuk u hap. Provo përsëri.'); }
  finally { finish(); }
}
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

async function intent(action: 'going' | 'decline', target = invitation?.id) {
  if (!session || !target || busy) return;
  begin();
  try {
    const response = await request('POST', `/api/${action}`, { gathering_id: target });
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
    const { cell } = await locateCell(currentGrid, locationMaxAccuracy, locationMaxAge);
    if (session?.token !== owner) return;
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
  if (configuration.schema_version !== 7) throw new Error('unsupported schema');
  const config = configuration.config, grid: Grid = await gridResponse.json();
  pollSeconds = config.notifications.foreground_poll_seconds;
  nonceSeconds = config.arrivals.nonce_seconds; currentGrid = grid;
  locationMaxAccuracy = config.geography.location_max_accuracy_meters;
  locationMaxAge = config.geography.location_fix_max_age_seconds;
  for (const minutes of config.availability.choices_minutes) duration.add(new Option(`${minutes} minuta`, String(minutes)));
  for (const km of config.geography.travel_radius_choices_km) radius.add(new Option(`${km} km`, String(km)));
  radius.value = String(config.geography.travel_radius_choices_km.includes(3) ? 3 : config.geography.travel_radius_choices_km[0]);
  el('nearby-area-description').textContent = `Në zonën publike ${config.public_activity.area_size_meters / 1000} km që përmban vendndodhjen e dhënë. Nuk është rrezja jote e udhëtimit.`;
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
    map.on('error', () => { el('area-status').textContent = 'Harta nuk u hap plotësisht. Mund të përdorësh vendndodhjen një herë.'; });
  } catch {
    el('map').textContent = 'Harta nuk mund të hapet në këtë pajisje.';
  }
  activityView = new ActivityView(map, configuration.sha256,
    () => ({ cell: session?.confirmed ? session.request.cell : undefined, gathering: invitation?.id }),
    event => {
      if (busy) return;
      if (session) { void loadPreview(event.id, session, session.expires, true); return; }
      joinTarget = event; render();
      form.scrollIntoView({ block: 'start' }); ready.focus();
    }, config.matching.late_join_min_remaining_minutes);
  drawArea = (cell: string) => {
    const ratio = config.public_activity.area_size_meters / grid.size_meters;
    const parts = cell.split(':');
    const publicCell = `${grid.version}:${config.public_activity.area_size_meters}:${Math.floor(Number(parts[2]) / ratio)}:${Math.floor(Number(parts[3]) / ratio)}`;
    const publicGrid: Grid = { ...grid, size_meters: config.public_activity.area_size_meters, lon_step: grid.lon_step * ratio, lat_step: grid.lat_step * ratio };
    const coordinates = polygon(publicGrid, publicCell);
    const draw = () => map?.getSource<maplibregl.GeoJSONSource>('selection')?.setData({ type: 'Feature', properties: {}, geometry: { type: 'Polygon', coordinates: [coordinates] } });
    if (map?.isStyleLoaded()) draw(); else map?.once('load', draw);
    // Camera uses the coarse center too; exact device coordinates are discarded.
    map?.jumpTo({ center: [(coordinates[0][0] + coordinates[2][0]) / 2, (coordinates[0][1] + coordinates[2][1]) / 2] });
    el('area-status').textContent = 'Zona u mor nga pajisja. Pozicioni yt i saktë nuk dërgohet.'; render();
  }
  message('');
  if (session) { drawArea(session.request.cell); await sync(); }
  setInterval(() => {
    if (preview) render();
    if (joinTarget && joinTarget.ends_at <= Date.now()) { joinTarget = null; render(); }
    if (pendingNonce && pendingNonce.expires <= Date.now()) pendingNonce = null;
    if (here && arrivalUntil <= Date.now()) { here = false; arrivalUntil = 0; message('Konfirmimi i mbërritjes përfundoi. Nëse je ende aty, mund të shtypësh sërish JAM KËTU.'); }
    if (invitation && invitation.ends_at <= Date.now()) { invitation = null; going = false; }
    if (session && session.expires <= Date.now()) end('Gatishmëria përfundoi.');
    if (session && session.confirmed && !document.hidden && Date.now() >= nextPoll) void sync();
    if (session) render();
  }, 1000);
  window.addEventListener('offline', render);
  window.addEventListener('online', () => { render(); if (session) void sync(); });
  document.addEventListener('visibilitychange', () => { if (!document.hidden && session) void sync(); });
}
setInterval(() => {
  if (session && session.expires <= Date.now()) {
    if (initialized) end('Gatishmëria përfundoi.');
    else { session = null; sessions.clear(); active.hidden = true; message('Gatishmëria përfundoi.'); }
  }
}, 1000);
start().catch(() => { message('GATI nuk u hap. Kontrollo lidhjen dhe rifresko faqen.'); });
