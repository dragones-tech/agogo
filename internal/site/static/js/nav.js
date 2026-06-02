// nav.js — navegación PARCIAL del sitio. Es una pieza propia de agogo (no de
// lumen): mejora la experiencia entre secciones sin recargar el documento.
//
// Intercepta los clics en enlaces internos, pide al servidor SOLO el fragmento
// (cabecera X-Fragment) —título + contenido de <main> + los scripts de la
// página— y reemplaza <main>. Header, footer, CSS y este mismo script quedan
// fijos; no se re-renderiza todo (aprovecha que el bloque "content" ya es un
// partial). Sin JS, o ante cualquier fallo, la navegación es la normal del
// navegador (full reload), así SEO, refresh y robustez quedan intactos.

let version = 0; // URL única por swap → los módulos de página se reejecutan

const runnable = (s) => !s.type || s.type === 'module' || /javascript/.test(s.type);

async function navigate(url, { push = true } = {}) {
  let res;
  try {
    res = await fetch(url, { headers: { 'X-Fragment': '1' }, credentials: 'same-origin' });
  } catch {
    location.href = url; // sin red → que navegue el navegador
    return;
  }
  // Si la respuesta no es un fragmento HTML (p. ej. un /api/... JSON), full nav.
  if (!res.ok || !(res.headers.get('content-type') || '').includes('text/html')) {
    location.href = url;
    return;
  }

  const doc = new DOMParser().parseFromString(await res.text(), 'text/html');
  const newTitle = doc.querySelector('title')?.textContent;

  // Los scripts EJECUTABLES se reinyectan aparte (DOMParser no los corre); el
  // resto del cuerpo (contenido, <template>, JSON island) va dentro de <main>.
  const scripts = [...doc.body.querySelectorAll('script')].filter(runnable);
  scripts.forEach((s) => s.remove());

  const main = document.querySelector('main');
  const apply = () => {
    if (newTitle) document.title = newTitle;
    main.replaceChildren(...doc.body.childNodes);
  };

  // Crossfade nativo dentro de la misma página, si el navegador lo soporta.
  if (document.startViewTransition) await document.startViewTransition(apply).updateCallbackDone;
  else apply();

  // Reejecuta los módulos de la página: URL única (?v=) → se vuelven a ejecutar
  // y recablean el <main> nuevo; sus imports (lumen, etc.) siguen cacheados.
  // Se inyectan DENTRO de <main>: así el replaceChildren del próximo swap los
  // limpia solo (si no, se acumularían en <body> en cada navegación).
  version++;
  for (const old of scripts) {
    const fresh = document.createElement('script');
    for (const attr of old.attributes) fresh.setAttribute(attr.name, attr.value);
    if (old.src) fresh.src = old.src + (old.src.includes('?') ? '&' : '?') + 'v=' + version;
    else fresh.textContent = old.textContent;
    main.appendChild(fresh);
  }

  if (push) history.pushState(null, '', url);
  window.scrollTo(0, 0);
}

// Intercepta clics en enlaces internos (delegado en document → sobrevive swaps).
document.addEventListener('click', (e) => {
  if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  const a = e.target.closest('a');
  if (!a || a.target || a.hasAttribute('download') || /\bexternal\b/.test(a.rel)) return;
  const url = new URL(a.href, location.href);
  if (url.origin !== location.origin) return;                 // enlace externo
  if (url.pathname === location.pathname && url.hash) return; // ancla en la misma página
  e.preventDefault();
  navigate(url.href);
});

// Atrás/adelante: re-pide el fragmento sin empujar otra entrada al historial.
addEventListener('popstate', () => navigate(location.href, { push: false }));
