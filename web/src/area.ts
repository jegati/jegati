export interface Grid {
  version: string; size_meters: number; west: number; south: number; east: number; north: number;
  lon_step: number; lat_step: number; columns: number; rows: number;
}
export function cellAt(g: Grid, lon: number, lat: number): string | null {
  if (!Number.isFinite(lon) || !Number.isFinite(lat) || lon < g.west || lon >= g.east || lat < g.south || lat >= g.north) return null;
  return `${g.version}:${g.size_meters}:${Math.floor((lon - g.west) / g.lon_step)}:${Math.floor((lat - g.south) / g.lat_step)}`;
}
export function polygon(g: Grid, id: string): [number, number][] {
  const [, , x, y] = id.split(':');
  const lon = g.west + Number(x) * g.lon_step, lat = g.south + Number(y) * g.lat_step;
  return [[lon, lat], [lon + g.lon_step, lat], [lon + g.lon_step, lat + g.lat_step], [lon, lat + g.lat_step], [lon, lat]];
}
