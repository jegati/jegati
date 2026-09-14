import './style.css';
import { ActivityView } from './activity';
import 'maplibre-gl/dist/maplibre-gl.css';
import * as maplibregl from 'maplibre-gl';
import workerURL from 'maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url';
maplibregl.setWorkerUrl(workerURL);
import { polygon, type Grid } from './area';
import * as sessions from './session';
import { locateCell } from './location';
import { PushController } from './push';
import { createSessionController, type Invitation } from './session-controller';
import { createDestinationMap, decorateDestinationMap } from './destination-map';

const el = <T extends HTMLElement>(id: string) => document.getElementById(id) as T;
const form = el<HTMLFormElement>('willingness'), active = el('active'), status = el('status');
const duration = el<HTMLSelectElement>('duration'), radius = el<HTMLSelectElement>('radius');
const ready = el<HTMLButtonElement>('ready'), cancel = el<HTMLButtonElement>('cancel'), retry = el<HTMLButtonElement>('retry');
const controller = createSessionController({
  render: () => render(),
  message: text => message(text),
  beforeEnd: () => {
    if (state.session) void push?.forget(state.session.token);
    closePreview(); locationRequest?.abort(); joinTarget = null;
  },
  afterEnd: text => {
    map?.getSource<maplibregl.GeoJSONSource>('selection')?.setData({ type: 'FeatureCollection', features: [] });
    el('area-status').textContent = 'Vendndodhja ende nuk është marrë.';
    render(); map?.resize(); message(text);
  },
  accepted: () => { joinTarget = null; },
});
const { state, begin, finish, end, sync, intent } = controller;
let locationRequest: AbortController | null = null;
let drawArea: (cell: string) => void = () => {};
let map: maplibregl.Map | undefined;
let initialized = false;
let activityView: ActivityView | undefined;
let push: PushController | undefined;
let preview: { candidate: sessions.Session; invitation: Invitation; expires: number; existing: boolean } | null = null;
let previewMap: maplibregl.Map | undefined;
let joinTarget: { id: string; ends_at: number } | null = null;
let destinationMap: maplibregl.Map | undefined;
let renderedDestination: string | null = null;
function invitationView() {
  const card = el('invitation'); card.hidden = !state.invitation;
  if (!state.invitation) { destinationMap?.remove(); destinationMap = undefined; renderedDestination = null; return; }
  el('collective-title').textContent = state.invitation.state === 'jemi_ketu' ? 'JEMI KËTU.' : 'JEMI GATI.';
  el<HTMLButtonElement>('arrive').hidden = !state.going || state.here;
  el<HTMLButtonElement>('arrive').disabled = state.busy;
  el<HTMLButtonElement>('renew-arrival').hidden = !controller.canRenew();
  el<HTMLButtonElement>('renew-arrival').disabled = state.busy || !navigator.onLine;
  el<HTMLButtonElement>('retract').hidden = !state.here;
  el<HTMLButtonElement>('retract').disabled = state.busy;
  el('arrival-status').textContent = state.here ? `Mbërritja jote është konfirmuar për rreth ${Math.max(1, Math.ceil((state.arrivalUntil - Date.now()) / 60000))} minuta të tjera.` : state.going ? 'Kur të mbërrish, konfirmo me vendndodhjen një herë.' : '';
  el('destination').textContent = state.invitation.intersection.label;
  el('gathering-time').textContent = `Takimi përfundon pas rreth ${Math.max(1, Math.ceil((state.invitation.ends_at - Date.now()) / 60_000))} minutash.`;
  el('going-status').textContent = state.going ? 'Ke zgjedhur të shkosh.' : 'Mund të zgjedhësh nëse do të shkosh.';
  el<HTMLButtonElement>('going').hidden = state.going;
  el<HTMLButtonElement>('going').disabled = state.busy;
  el<HTMLButtonElement>('decline').disabled = state.busy;
  el('decline').textContent = state.going ? 'Nuk po shkoj më' : 'JO TANI';
  if (renderedDestination === state.invitation.id) return;
  destinationMap?.remove(); renderedDestination = state.invitation.id; el('destination-map').removeAttribute('data-ready');
  try {
    destinationMap = createDestinationMap('destination-map', state.invitation.intersection.point, 'Harta e pikës së takimit');
    destinationMap.on('idle', () => { el('destination-map').setAttribute('data-ready', 'true'); });
    decorateDestinationMap(destinationMap, state.invitation.intersection.point, true);
    destinationMap.getCanvas().setAttribute('aria-label', 'Pika e takimit pranë kryqëzimit');
  } catch { el('destination-map').textContent = 'Harta nuk mund të hapet në këtë pajisje.'; }
}
const message = (text: string) => { status.textContent = text; };
function render() {
  form.hidden = state.session !== null || !!push?.pendingResume; active.hidden = state.session === null;
  cancel.disabled = state.cancelling; retry.disabled = state.busy;
  ready.disabled = state.busy || !initialized || !navigator.onLine || !!push?.pendingResume;
  duration.disabled = state.busy; radius.disabled = state.busy;
  el<HTMLButtonElement>('join-back').disabled = state.busy;
  el('location-cancel').hidden = !locationRequest;
  el<HTMLButtonElement>('refresh-status').disabled = state.busy || !navigator.onLine;
  el('connection-status').textContent = !navigator.onLine ? 'Nuk ka lidhje. Veprimet nuk dërgohen; gatishmëria përfundon në afatin e saj.' : state.failures ? 'Lidhja me GATI nuk u konfirmua. Mund të provosh përsëri.' : '';
  el('waiting-guidance').textContent = !state.session?.confirmed ? 'Po kontrollojmë nëse veprimi u pranua.' : state.here ? 'Mbërritja është e përkohshme. Mund ta heqësh konfirmimin kur të duash.' : state.going ? 'Ke zgjedhur të shkosh. Kur të mbërrish, shtyp JAM KËTU.' : state.invitation ? 'Zgjidh nëse dëshiron të shkosh. JO TANI e mban gatishmërinë aktive.' : 'Gatishmëria jote është aktive. Po presim një takim të përshtatshëm. Mungesa e shifrave publike nuk do të thotë që je vetëm.';
  retry.hidden = !state.session || state.session.confirmed;
  el('active-title').textContent = state.session?.confirmed ? (state.here ? 'JAM KËTU.' : 'JAM GATI.') : 'Po kontrollojmë gatishmërinë…';
  invitationView();
  activityView?.render();
  push?.render();
  if (preview) {
    const expired = preview.expires <= Date.now();
    el<HTMLButtonElement>('public-join-confirm').disabled = state.busy || expired || !navigator.onLine;
    el('preview-status').textContent = expired ? 'Parapamja përfundoi. Mbylle dhe kontrollo sërish pikën e takimit.' : !navigator.onLine ? 'Nuk ka lidhje. Provo përsëri kur të rikthehet.' : '';
  }
  ready.textContent = locationRequest ? 'Po merret vendndodhja…' : joinTarget ? 'Shiko pikën e takimit' : 'JAM GATI';
  el('join-choice').hidden = !joinTarget; el('join-back').hidden = !joinTarget;
  if (joinTarget) el('join-choice').textContent = 'Shiko pikën e takimit me vendndodhjen nga pajisja, pastaj zgjidh nëse do të shkosh.';
  if (state.session) el('remaining').textContent = `Përfundon pas rreth ${Math.max(1, Math.ceil((state.session.expires - Date.now()) / 60_000))} minutash.`;
}
form.addEventListener('submit', async event => {
  event.preventDefault();
  if (state.busy || state.session || !state.currentGrid || !navigator.onLine) return;
  const target = joinTarget;
  const choices = { radius_km: Number(radius.value), availability_minutes: Number(duration.value) };
  const controller = new AbortController(); locationRequest = controller; begin();
  el('area-status').textContent = 'Po merret vendndodhja nga pajisja…';
  let fix: { cell: string; expires: number } | undefined;
  try {
    fix = await locateCell(state.currentGrid, state.locationMaxAccuracy, state.locationMaxAge, controller.signal);
    drawArea(fix.cell);
  } catch (error) {
    const text = error instanceof DOMException && error.name === 'AbortError' ? 'Veprimi u anulua. Nuk u dërgua gatishmëri.' : error instanceof Error ? error.message : 'Vendndodhja nuk u mor. Provo përsëri.';
    el('area-status').textContent = text; message(text);
  } finally { locationRequest = null; finish(); }
  if (!fix || controller.signal.aborted) return;
  const candidate = sessions.create({ cell: fix.cell, ...choices });
  if (target) await loadPreview(target.id, candidate, Math.min(fix.expires, candidate.expires), false);
  else { state.session = candidate; sessions.save(state.session); void sync(true); }
});
el<HTMLButtonElement>('location-cancel').onclick = () => locationRequest?.abort();
el<HTMLButtonElement>('public-join-cancel').onclick = closePreview;
el<HTMLDialogElement>('public-join-dialog').addEventListener('cancel', closePreview);
el<HTMLButtonElement>('public-join-confirm').onclick = () => {
  const chosen = preview;
  if (!chosen || chosen.expires <= Date.now() || !navigator.onLine || state.busy) return;
  closePreview();
  if (chosen.existing) {
    if (state.session?.token === chosen.candidate.token) void intent('going', chosen.invitation.id);
  } else if (!state.session) {
    state.session = { ...chosen.candidate, expires: Date.now() + chosen.candidate.request.availability_minutes * 60000, joinGathering: chosen.invitation.id };
    sessions.save(state.session); void sync(true);
  }
};
el<HTMLButtonElement>('join-back').onclick = () => { joinTarget = null; closePreview(); render(); };
el<HTMLButtonElement>('refresh-status').onclick = () => { void sync(!state.session?.confirmed); };
function closePreview() {
  preview = null; previewMap?.remove(); previewMap = undefined;
  el<HTMLDialogElement>('public-join-dialog').close();
  el('preview-destination').textContent = ''; el('preview-time').textContent = ''; el('preview-status').textContent = '';
}
async function loadPreview(id: string, candidate: sessions.Session, expires: number, existing: boolean) {
  if (state.busy) return; begin();
  try {
    const response = await fetch('/api/gathering-preview', { method: 'POST', credentials: 'omit', cache: 'no-store', signal: AbortSignal.timeout(10000), headers: { Authorization: `Bearer ${candidate.token}`, 'Content-Type': 'application/json' }, body: JSON.stringify({ ...candidate.request, gathering_id: id }) });
    if (!response.ok) throw new Error('Ky takim nuk është i arritshëm ose nuk është më i hapur. Mund të zgjedhësh një tjetër.');
    const result = await response.json();
    if (!Number.isFinite(result.preview_expires_at) || result.preview_expires_at <= Date.now() || expires <= Date.now() || (existing && state.session?.token !== candidate.token)) throw new Error('Parapamja përfundoi. Provo përsëri.');
    closePreview(); preview = { candidate, invitation: result.invitation, expires: Math.min(expires, result.preview_expires_at), existing };
    el('preview-destination').textContent = preview.invitation.intersection.label;
    el('preview-time').textContent = `Takimi përfundon pas rreth ${Math.max(1, Math.ceil((preview.invitation.ends_at - Date.now()) / 60000))} minutash.`;
    el<HTMLButtonElement>('public-join-confirm').disabled = false;
    el<HTMLDialogElement>('public-join-dialog').showModal();
    try {
      previewMap = createDestinationMap('preview-map', preview.invitation.intersection.point, 'Parapamja e takimit');
      decorateDestinationMap(previewMap, preview.invitation.intersection.point);
    } catch { el('preview-map').textContent = 'Harta nuk mund të hapet në këtë pajisje.'; }
  } catch (error) { message(error instanceof Error ? error.message : 'Parapamja nuk u hap. Provo përsëri.'); }
  finally { finish(); }
}
retry.onclick = () => { void sync(true); };
cancel.onclick = controller.cancelWillingness;
el<HTMLButtonElement>('arrive').onclick = controller.arrive;
el<HTMLButtonElement>('renew-arrival').onclick = controller.renewArrival;
el<HTMLButtonElement>('retract').onclick = controller.retract;
el<HTMLButtonElement>('going').onclick = () => { void intent('going'); };
el<HTMLButtonElement>('decline').onclick = () => { void intent('decline'); };

async function start() {
  const [configResponse, gridResponse] = await Promise.all(['/api/config', '/api/geography'].map(url => fetch(url, { credentials: 'omit', cache: 'no-store' })));
  if (!configResponse.ok || !gridResponse.ok) throw new Error('unavailable');
  const configuration = await configResponse.json();
  if (configuration.schema_version !== 9) throw new Error('unsupported schema');
  const config = configuration.config, grid: Grid = await gridResponse.json();
  state.pollSeconds = config.notifications.foreground_poll_seconds;
  state.nonceSeconds = config.arrivals.nonce_seconds;
  state.renewalWindowSeconds = config.arrivals.renewal_window_seconds; state.currentGrid = grid;
  state.locationMaxAccuracy = config.geography.location_max_accuracy_meters;
  state.locationMaxAge = config.geography.location_fix_max_age_seconds;
  for (const minutes of config.availability.choices_minutes) duration.add(new Option(`${minutes} minuta`, String(minutes)));
  for (const km of config.geography.travel_radius_choices_km) radius.add(new Option(`${km} km`, String(km)));
  radius.value = String(config.geography.travel_radius_choices_km.includes(3) ? 3 : config.geography.travel_radius_choices_km[0]);
  el('nearby-area-description').textContent = `Në zonën publike ${config.public_activity.area_size_meters / 1000} km që përmban vendndodhjen e dhënë. Nuk është rrezja jote e udhëtimit.`;
  push = new PushController(() => state.session, value => { if (!state.session) { state.session = value; sessions.save(value); if (initialized) { drawArea(value.request.cell); void sync(); } } }, render);
  await push.initialize();
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
    () => ({ cell: state.session?.confirmed ? state.session.request.cell : undefined, gathering: state.invitation?.id }),
    event => {
      if (state.busy) return;
      if (state.session) { void loadPreview(event.id, state.session, state.session.expires, true); return; }
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
  if (state.session) { drawArea(state.session.request.cell); await sync(); }
  setInterval(() => {
    push?.tick();
    if (preview) render();
    if (joinTarget && joinTarget.ends_at <= Date.now()) { joinTarget = null; render(); }
    controller.tick();
  }, 1000);
  window.addEventListener('offline', render);
  window.addEventListener('online', () => { render(); if (state.session) void sync(); else if (push?.pendingResume) void push.initialize(); });
  navigator.serviceWorker?.addEventListener('message', event => { if (event.data?.type === 'gati-refresh') { if (state.session) void sync(); else void push?.initialize(); } });
  document.addEventListener('visibilitychange', () => { if (!document.hidden && state.session) void sync(); });
}
setInterval(() => {
  if (state.session && state.session.expires <= Date.now()) {
    if (initialized) end('Gatishmëria përfundoi.');
    else { state.session = null; sessions.clear(); active.hidden = true; message('Gatishmëria përfundoi.'); }
  }
}, 1000);
start().catch(() => { message('GATI nuk u hap. Kontrollo lidhjen dhe rifresko faqen.'); });
