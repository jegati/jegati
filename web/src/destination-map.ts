import * as maplibregl from 'maplibre-gl';

// Only private invitations/previews use destination geometry. Public activity
// maps have different cell-only publication rules and their own renderer.
export function createDestinationMap(container: string, point: [number, number], title: string) {
  return new maplibregl.Map({
    container, center: point, zoom: 15,
    attributionControl: false, locale: { 'Map.Title': title },
    style: {
      version: 8,
      sources: { roads: { type: 'geojson', data: '/api/map/roads' } },
      layers: [
        { id: 'background', type: 'background', paint: { 'background-color': '#f0eee6' } },
        { id: 'roads', type: 'line', source: 'roads', paint: { 'line-color': '#bec3b8', 'line-width': 3 } },
      ],
    },
  });
}

// The caller retains the map before decoration, so it can still remove the map
// if marker/control initialization fails. Preserve each view's accessible labels.
export function decorateDestinationMap(map: maplibregl.Map, point: [number, number], accessibleImage = false) {
  // This marker is the shared mapped destination, never a person's position.
  const marker = document.createElement('span');
  marker.className = 'destination-marker';
  marker.textContent = '🦩';
  if (accessibleImage) marker.setAttribute('role', 'img');
  marker.setAttribute('aria-label', 'Pika e takimit');
  new maplibregl.Marker({ element: marker }).setLngLat(point).addTo(map);
  map.addControl(new maplibregl.AttributionControl({ compact: false, customAttribution: '© OpenStreetMap · ODbL' }));
}
