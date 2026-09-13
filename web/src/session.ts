export interface Willingness { cell: string; radius_km: number; availability_minutes: number }
export interface Session { token: string; expires: number; request: Willingness; confirmed: boolean }
const key = 'gati-session-v1';
// Session storage survives reload, never creates a permanent identity. Browsers
// may restore tabs: recheck the absolute deadline before any authenticated request.
export function restore(): Session | null {
  try {
    const s = JSON.parse(sessionStorage.getItem(key) ?? 'null');
    if (s && /^[A-Za-z0-9_-]{43}$/.test(s.token) && Number.isFinite(s.expires) && s.expires > Date.now() && s.expires <= Date.now() + 120 * 60_000 && typeof s.request?.cell === 'string' && Number.isInteger(s.request.radius_km) && Number.isInteger(s.request.availability_minutes)) return s;
  } catch { /* Storage can be unavailable; the active tab still works. */ }
  clear();
  return null;
}
export function save(s: Session) { try { sessionStorage.setItem(key, JSON.stringify(s)); } catch { /* Memory-only fallback. */ } }
export function clear() { try { sessionStorage.removeItem(key); } catch { /* No accessible storage. */ } }
export function create(request: Willingness): Session {
  const token = randomToken();
  return { token, request, expires: Date.now() + request.availability_minutes * 60_000, confirmed: false };
}

export function randomToken(): string {
 const bytes = crypto.getRandomValues(new Uint8Array(32));
 return btoa(String.fromCharCode(...bytes)).replaceAll('+', '-').replaceAll('/', '_').replace(/=+$/, '');
}
