import './style.css';

// Scaffold only: no location request, storage, participant action or subscription.
const status = document.querySelector<HTMLElement>('#status');
fetch('/api/config', { credentials: 'omit', cache: 'no-store' })
  .then(async (response) => {
    if (!response.ok) throw new Error('configuration unavailable');
    const body: unknown = await response.json();
    if (typeof body !== 'object' || body === null || !('schema_version' in body) || body.schema_version !== 1) {
      throw new Error('unsupported configuration');
    }
    if (status) status.textContent = 'GATI po përgatitet. Së shpejti, këtu.';
  })
  .catch(() => {
    if (status) status.textContent = 'Lidhja nuk është e mundur tani. Provo përsëri më vonë.';
  });
