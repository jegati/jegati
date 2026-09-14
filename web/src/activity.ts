import type { Map as CityMap, GeoJSONSource } from 'maplibre-gl';
import { polygon, type Grid } from './area';

interface Event { id: string; cell: string; ends_at: number; state: string; going?: number; here?: number }
interface Release { version: number; id: string; config_sha256: string; observed_from: number; observed_until: number; release_at: number; expires_at: number; grid: Grid; areas: { cell: string; willing: number }[]; gatherings: Event[] }
interface Own { cell?: string; gathering?: string }
const el = (id: string) => document.getElementById(id)!;
const bucket = (value?: number) => value ? `${value}+` : 'Nuk ka shifër të publikuar';

// One canonical release feeds all cards and both map layers. No per-user query,
// smoothing, exact destination geometry or private response enters this module.
export class ActivityView {
  private release: Release | null = null;
  private fetching = false;
  private nextFetch = 0;
  private drawn = '';
  private selectedCell = '';
  private layer = 'willingness';
  private rendered = '';
  constructor(private map: CityMap | undefined, private hash: string, private own: () => Own,
    private join: (event: Event) => void, private cutoffMinutes: number, private refreshSeconds = 30) {
    const selector = el('activity-layer') as HTMLSelectElement;
    selector.onchange = () => { this.layer = selector.value; this.drawn = ''; this.render(); };
    map?.on('load', () => { this.drawn = ''; this.render(); });
    map?.on('click', e => {
      if (!map.getLayer('activity-fill')) return;
      const cell = map.queryRenderedFeatures(e.point, { layers: ['activity-fill'] })[0]?.properties?.cell;
      if (typeof cell === 'string') { this.selectedCell = cell; this.rendered = ''; this.render(); }
    });
    (el('activity-all') as HTMLButtonElement).onclick = () => { this.selectedCell = ''; this.rendered = ''; this.render(); };
    document.addEventListener('visibilitychange', () => { if (!document.hidden) void this.refresh(); });
    window.addEventListener('offline', () => this.render());
    window.addEventListener('online', () => { this.nextFetch = 0; void this.refresh(); });
    setInterval(() => { this.render(); if (!document.hidden && Date.now() >= this.nextFetch) void this.refresh(); }, 1000);
    void this.refresh();
  }
  private async refresh() {
    if (this.fetching) return; this.fetching = true;
    try {
      const response = await fetch('/api/activity/latest', { credentials: 'omit', signal: AbortSignal.timeout(10000) });
      if (response.status === 204) this.release = null;
      else {
        if (!response.ok) throw new Error('unavailable');
        const r: Release = await response.json();
        if (r.version !== 1 || r.config_sha256 !== this.hash || !Number.isFinite(r.expires_at) || r.expires_at <= Date.now() || r.release_at > Date.now() || !Array.isArray(r.areas) || !Array.isArray(r.gatherings) || r.grid.size_meters < 1000) throw new Error('invalid release');
        this.release = r;
      }
    } catch { /* Keep only a previously fetched, still-unexpired public release. */ }
    finally { this.fetching = false; this.nextFetch = Date.now() + Math.max(5, Math.min(30, this.refreshSeconds)) * 1000 * (1 + Math.random() * .2); this.render(); }
  }
  render() {
    const r = this.release && this.release.expires_at > Date.now() ? this.release : null;
    const own = this.own();
    const now = Date.now();
    const liveGatherings = r?.gatherings.filter(g => g.ends_at > now) ?? [];
    el('nearby-statistics').hidden = !own.cell;
    el('activity-time').textContent = r ? `Vëzhguar më ${new Date(r.observed_from).toLocaleTimeString('sq-AL', { hour: '2-digit', minute: '2-digit', hour12: false })}–${new Date(r.observed_until).toLocaleTimeString('sq-AL', { hour: '2-digit', minute: '2-digit', hour12: false })}. Shifrat janë një pamje e publikuar, jo numërim i çastit.` : 'Ende nuk ka të dhëna të publikuara. Kjo nuk do të thotë që nuk ka aktivitet.';
    let nearby = '';
    if (r && own.cell) {
      const [version, size, x, y] = own.cell.split(':');
      const ratio = r.grid.size_meters / Number(size);
      if (Number.isInteger(ratio) && version === r.grid.version) nearby = `${version}:${r.grid.size_meters}:${Math.floor(Number(x) / ratio)}:${Math.floor(Number(y) / ratio)}`;
    }
    const willingness = r?.areas.find(a => a.cell === nearby)?.willing;
    el('nearby-count').textContent = willingness ? `${willingness}+ GATI` : 'Nuk ka shifër të publikuar';
    const event = r?.gatherings.find(g => g.id === own.gathering);
    el('going-count').textContent = `Kanë zgjedhur të shkojnë: ${bucket(event?.going)}`;
    el('here-count').textContent = `Kanë konfirmuar mbërritjen: ${bucket(event?.here)}`;
    const renderKey = `${r?.id ?? ''}:${this.selectedCell}:${navigator.onLine}:${r?.gatherings.map(g => `${now + this.cutoffMinutes * 60000 < g.ends_at}:${g.ends_at <= now}`).join(',')}`;
    if (this.rendered !== renderKey) {
      this.rendered = renderKey;
      const list = el('activity-gatherings'); list.replaceChildren();
      const selectedWilling = r?.areas.find(a => a.cell === this.selectedCell)?.willing;
      el('activity-selected-count').textContent = this.selectedCell ? `GATI në zonën e zgjedhur: ${bucket(selectedWilling)}` : '';
      el('activity-selection').textContent = this.selectedCell ? 'Takimet në zonën e zgjedhur' : 'Takimet në zonat e publikuara';
      for (const g of r?.gatherings ?? []) {
        if (this.selectedCell && g.cell !== this.selectedCell) continue;
        const card = document.createElement('article'); card.className = 'public-gathering'; card.dataset.gathering = g.id;
        const ended = g.ends_at <= now;
        card.dataset.state = ended ? 'ended' : 'open';
        const title = document.createElement('h3'); title.textContent = ended ? 'Takimi përfundoi.' : g.state === 'jemi_ketu' ? 'JEMI KËTU.' : 'JEMI GATI.';
        const text = document.createElement('p'); text.textContent = `Kanë zgjedhur të shkojnë: ${bucket(g.going)} · Kanë konfirmuar mbërritjen: ${bucket(g.here)}`;
        const area = document.createElement('button'); area.type = 'button'; area.textContent = 'Shiko zonën në hartë';
        area.onclick = () => { const corners = polygon(r!.grid, g.cell); this.map?.fitBounds([corners[0], corners[2]], { padding: 40, maxZoom: 14, duration: 0 }); this.selectedCell = g.cell; this.rendered = ''; this.render(); };
        const button = document.createElement('button'); button.type = 'button'; button.textContent = ended ? 'Takimi ka përfunduar' : 'Dua të bashkohem';
        button.disabled = !navigator.onLine || Date.now() + this.cutoffMinutes * 60000 >= g.ends_at;
        button.onclick = () => { if (this.release && this.release.expires_at > Date.now() && navigator.onLine && Date.now() + this.cutoffMinutes * 60000 < g.ends_at) this.join(g); };
        card.append(title, text, area, button); list.append(card);
      }
      if (!list.children.length) { const p = document.createElement('p'); p.textContent = 'Nuk ka takime të publikuara për këtë pamje.'; list.append(p); }
    }
    if (!this.map?.isStyleLoaded()) return;
    const drawKey = `${r?.id ?? 'empty'}:${this.layer}:${liveGatherings.map(g => g.id).join(',')}`;
    if (this.drawn === drawKey) return; this.drawn = drawKey;
    const cells = new Map<string, number>();
    if (r) {
      if (this.layer === 'willingness') for (const a of r.areas) cells.set(a.cell, a.willing);
      else for (const g of liveGatherings) cells.set(g.cell, Math.max(cells.get(g.cell) ?? 0, g.going ?? g.here ?? 1));
    }
    const features = [...cells].map(([cell, value]) => ({ type: 'Feature' as const, properties: { cell, bucket: value }, geometry: { type: 'Polygon' as const, coordinates: [polygon(r!.grid, cell)] } }));
    const data = { type: 'FeatureCollection' as const, features };
    const source = this.map.getSource<GeoJSONSource>('activity');
    if (source) source.setData(data);
    else {
      this.map.addSource('activity', { type: 'geojson', data });
      this.map.addLayer({ id: 'activity-fill', type: 'fill', source: 'activity', paint: { 'fill-color': '#963754', 'fill-opacity': ['step', ['get', 'bucket'], .18, 50, .3, 100, .42, 250, .55] } }, 'selection');
      this.map.addLayer({ id: 'activity-outline', type: 'line', source: 'activity', paint: { 'line-color': '#963754', 'line-width': 1 } }, 'selection');
    }
    el('map').setAttribute('data-activity', r?.id ?? 'unavailable');
  }
}
