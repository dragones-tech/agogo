// Entrada del catálogo: cablea el DOM con la Collection y la búsqueda en servidor.
//
// El servidor renderiza la lista (SEO, sin-JS) y emite los datos como JSON island
// para el primer pintado sin round-trip. Al teclear, cada tecla pega a
// GET /api/productos?q=...; el JSON resetea la Collection y el CatalogView
// reconcilia. La petición anterior en vuelo se cancela para no llegar desordenada.
import { $ } from '/static/lumen/src/dom.js';
import { Collection } from '/static/lumen/src/collection.js';
import { Http } from '/static/lumen/src/http.js';
import { Product } from './product.js';
import { CatalogView } from './catalog-view.js';

const data = $('[data-ref="data"]');
const serverList = $('[data-ref="list"]');
const filter = $('[data-ref="filter"]');
const empty = $('[data-ref="empty"]');

if (data && serverList && filter) {
  // Primer pintado desde el JSON island (sin pedir nada al servidor).
  const collection = new Collection(JSON.parse(data.textContent), Product);
  const catalog = new CatalogView({ collection });

  // lumen reemplaza la lista del servidor por la suya, en el mismo lugar.
  const parent = serverList.parentNode;
  const anchor = serverList.nextSibling;
  serverList.remove();
  await catalog.mount(parent);
  if (anchor) parent.insertBefore(catalog.el, anchor);

  // Búsqueda en servidor: una petición JSON por tecla.
  const api = new Http();
  let inFlight; // AbortController de la petición anterior
  filter.addEventListener('input', async () => {
    inFlight?.abort();
    inFlight = new AbortController();
    try {
      const results = await api.get('/api/productos', {
        query: { q: filter.value.trim() },
        signal: inFlight.signal,
      });
      collection.reset(results); // → el CatalogView re-renderiza
      if (empty) empty.hidden = results.length > 0;
    } catch (err) {
      if (err.name !== 'AbortError') console.error('search failed', err);
    }
  });
}
