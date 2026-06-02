// otw.js — on a button click, loads a server-rendered HTML fragment into a
// section, with no page refresh: just innerHTML. The fragment comes from a
// server-side BFF that calls a token-gated external API (the token never reaches
// the browser) and returns ready HTML. No JSON and no framework on the page —
// HTML over the wire.
//
// Markup: a [data-otw] section (its value = the fragment URL) containing a
// [data-otw-load] button and a [data-otw-target] container.
const section = document.querySelector('[data-otw]');

if (section) {
  const button = section.querySelector('[data-otw-load]');
  const target = section.querySelector('[data-otw-target]');

  button?.addEventListener('click', async () => {
    button.disabled = true;
    target.innerHTML = '<p aria-busy="true">Cargando…</p>';
    try {
      const res = await fetch(section.dataset.otw, { credentials: 'same-origin' });
      target.innerHTML = res.ok
        ? await res.text()
        : '<p class="alert" data-variant="danger" role="alert">No se pudo cargar el panel.</p>';
    } catch {
      target.innerHTML = '<p class="alert" data-variant="danger" role="alert">No se pudo cargar el panel.</p>';
    } finally {
      button.disabled = false;
    }
  });
}
