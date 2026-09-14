import type { Grid } from './area';
import * as sessions from './session';
import { locateCell } from './location';

export interface Invitation {
  id: string;
  intersection: { id: string; label: string; point: [number, number] };
  ends_at: number;
  state: string;
}

interface SessionState {
  session: sessions.Session | null;
  busy: boolean;
  operations: number;
  cancelling: boolean;
  pollSeconds: number;
  nextPoll: number;
  failures: number;
  here: boolean;
  arrivalUntil: number;
  nonceSeconds: number;
  currentGrid: Grid | undefined;
  locationMaxAccuracy: number;
  locationMaxAge: number;
  pendingNonce: { token: string; expires: number; issued: boolean } | null;
  invitation: Invitation | null;
  going: boolean;
}

interface Effects {
  render(): void;
  message(text: string): void;
  beforeEnd(): void;
  afterEnd(text: string): void;
  accepted(): void;
}

// Own-session requests, retries and deadlines live here. The view owns DOM/maps,
// startup, the one-second scheduler and optional push. Neither layer adds timers.
export function createSessionController(effects: Effects) {
  const state: SessionState = {
    session: sessions.restore(), busy: false, operations: 0, cancelling: false,
    pollSeconds: 30, nextPoll: 0, failures: 0,
    here: false, arrivalUntil: 0, nonceSeconds: 120,
    currentGrid: undefined, locationMaxAccuracy: 100, locationMaxAge: 60,
    pendingNonce: null, invitation: null, going: false,
  };
  const { render, message } = effects;
  function begin() { state.operations++; state.busy = true; render(); }
  function finish() { state.operations--; state.busy = state.operations > 0; render(); }
  function end(text: string) {
    effects.beforeEnd();
    state.session = null; sessions.clear(); state.invitation = null; state.going = false; state.here = false; state.arrivalUntil = 0; state.pendingNonce = null;
    effects.afterEnd(text);
  }

  async function request(method: string, path: string, body?: unknown, extraHeaders: Record<string, string> = {}) {
    if (!state.session || state.session.expires <= Date.now()) { end('Gatishmëria përfundoi.'); throw new Error('expired'); }
    return fetch(path, { method, credentials: 'omit', cache: 'no-store', signal: AbortSignal.timeout(10_000),
      headers: { Authorization: `Bearer ${state.session.token}`, ...(body ? { 'Content-Type': 'application/json' } : {}), ...extraHeaders },
      ...(body ? { body: JSON.stringify(body) } : {}) });
  }
  async function sync(create = false) {
    if (!state.session || state.busy) return;
    begin();
    try {
      const response = await request(create ? 'POST' : 'GET', create ? (state.session.joinGathering ? '/api/join' : '/api/signals') : '/api/signal', create ? { ...state.session.request, ...(state.session.joinGathering ? { gathering_id: state.session.joinGathering } : {}) } : undefined);
      if (state.session?.joinGathering && [409, 410].includes(response.status)) { end('Nuk mund të bashkohesh në këtë takim. Mund të shprehësh sërish gatishmërinë.'); return; }
      if (response.status === 410) { end('Gatishmëria përfundoi ose shërbimi u rinis. Mund të shprehësh sërish gatishmërinë.'); return; }
      if (!response.ok) throw new Error('unavailable');
      const signal = await response.json();
      if (!Number.isFinite(signal.expires_at) || !Number.isFinite(signal.created_at) || signal.expires_at <= signal.created_at) throw new Error('invalid deadline');
      if (state.session) {
        // Never extend the local maximum on retries or reload. Server enforces its
        // own clock/deadline even if this browser's wall clock has been changed.
        state.session.expires = Math.min(state.session.expires, signal.expires_at);
        state.session.confirmed = true; delete state.session.joinGathering; effects.accepted(); sessions.save(state.session);
        applySignal(signal);
        state.failures = 0; message(state.here ? 'Mbërritja jote është konfirmuar përkohësisht.' : state.going ? 'Ke zgjedhur të shkosh.' : 'Gatishmëria jote është aktive.');
      }
    } catch { if (!state.session) return; state.failures = Math.min(state.failures + 1, 4); message('Lidhja u ndërpre. Gatishmëria mund të jetë aktive; provo përsëri.'); }
    finally { finish(); state.nextPoll = Date.now() + state.pollSeconds * 1000 * 2 ** state.failures * (1 + Math.random() * .2); render(); }
  }

  async function cancelWillingness() {
    if (state.cancelling || !state.session) return;
    state.cancelling = true; begin();
    try {
      const response = await request('DELETE', '/api/signal');
      if (!response.ok && response.status !== 410) throw new Error('unavailable');
      end('Gatishmëria u mbyll.');
    } catch { if (state.session) message('Mbyllja nuk u konfirmua. Provo përsëri; gatishmëria përfundon vetë në afatin e saj.'); }
    finally { state.cancelling = false; finish(); }
  }

  async function intent(action: 'going' | 'decline', target = state.invitation?.id) {
    if (!state.session || !target || state.busy) return;
    begin();
    try {
      const response = await request('POST', `/api/${action}`, { gathering_id: target });
      if (response.status === 410) { state.invitation = null; state.going = false; message('Ky takim nuk është më i hapur për t’u bashkuar. Gatishmëria jote mund të vazhdojë.'); return; }
      if (!response.ok) throw new Error('unavailable');
      const signal = await response.json();
      applySignal(signal);
      if (state.session) message(action === 'going' ? 'Zgjedhja u ruajt.' : 'Në rregull. Gatishmëria jote vazhdon.');
    } catch { if (state.session) message('Zgjedhja nuk u konfirmua. Provo përsëri.'); }
    finally { finish(); }
  }

  function applySignal(signal: { invitation?: Invitation; state: string; arrival_until?: number }) {
    if (!state.session) return;
    if (state.invitation?.id !== signal.invitation?.id) state.pendingNonce = null;
    state.invitation = signal.invitation ?? null; state.going = signal.state === 'going' || signal.state === 'here';
    state.here = signal.state === 'here'; state.arrivalUntil = signal.arrival_until ?? 0;
  }
  async function arrive() {
    if (state.busy || !state.session || !state.invitation || !state.currentGrid) return;
    const owner = state.session.token; begin();
    try {
      if (!state.pendingNonce || state.pendingNonce.expires <= Date.now()) state.pendingNonce = { token: sessions.randomToken(), expires: Date.now() + state.nonceSeconds * 1000, issued: false };
      const nonce = state.pendingNonce;
      const headers = { 'X-Gati-Arrival-Nonce': nonce.token };
      if (!nonce.issued) {
        const issued = await request('POST', '/api/arrival-nonce', undefined, headers);
        if (!issued.ok) throw new Error('unavailable');
        const challenge = await issued.json();
        if (!Number.isFinite(challenge.expires_at) || challenge.expires_at <= Date.now()) throw new Error('expired challenge');
        nonce.expires = Math.min(nonce.expires, challenge.expires_at); nonce.issued = true;
      }
      if (state.session?.token !== owner) return;
      const { cell } = await locateCell(state.currentGrid, state.locationMaxAccuracy, state.locationMaxAge);
      if (state.session?.token !== owner) return;
      const response = await request('POST', '/api/arrival', { cell }, headers);
      if (!response.ok) throw new Error('unavailable');
      if (state.session?.token !== owner) return;
      applySignal(await response.json()); state.pendingNonce = null;
      message('Mbërritja u konfirmua përkohësisht.');
    } catch { if (state.session?.token === owner) message('Mbërritja nuk u konfirmua. Kontrollo vendndodhjen dhe provo përsëri pranë pikës së takimit.'); }
    finally { finish(); }
  }
  async function retract() {
    if (state.busy || !state.session) return; begin();
    try {
      const response = await request('DELETE', '/api/arrival'); if (!response.ok) throw new Error('unavailable');
      applySignal(await response.json()); state.pendingNonce = null; if (state.session) message('Konfirmimi i mbërritjes u hoq.');
    } catch { if (state.session) message('Heqja nuk u konfirmua. Provo përsëri.'); }
    finally { finish(); }
  }

  function tick() {
    if (state.pendingNonce && state.pendingNonce.expires <= Date.now()) state.pendingNonce = null;
    if (state.here && state.arrivalUntil <= Date.now()) { state.here = false; state.arrivalUntil = 0; message('Konfirmimi i mbërritjes përfundoi. Nëse je ende aty, mund të shtypësh sërish JAM KËTU.'); }
    if (state.invitation && state.invitation.ends_at <= Date.now()) { state.invitation = null; state.going = false; }
    if (state.session && state.session.expires <= Date.now()) end('Gatishmëria përfundoi.');
    if (state.session && state.session.confirmed && !document.hidden && Date.now() >= state.nextPoll) void sync();
    if (state.session) render();
  }
  return { state, begin, finish, end, sync, intent, cancelWillingness, arrive, retract, tick };
}
