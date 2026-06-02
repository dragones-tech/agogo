// Catalog entry point: wires the DOM up with the Collection and server-side search.
//
// The server renders the list (SEO, no-JS) and emits the data as a JSON island
// for the first paint without a round-trip. While typing, each keystroke hits
// GET /api/productos?q=...; the JSON resets the Collection and the CatalogView
// reconciles. The previous in-flight request is aborted so results can't arrive out of order.
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
  // First paint from the JSON island (without asking the server for anything).
  const collection = new Collection(JSON.parse(data.textContent), Product);
  const catalog = new CatalogView({ collection });

  // lumen replaces the server's list with its own, in the same place.
  const parent = serverList.parentNode;
  const anchor = serverList.nextSibling;
  serverList.remove();
  await catalog.mount(parent);
  if (anchor) parent.insertBefore(catalog.el, anchor);

  // Server-side search: one JSON request per keystroke.
  const api = new Http();
  let inFlight; // AbortController of the previous request
  filter.addEventListener('input', async () => {
    inFlight?.abort();
    inFlight = new AbortController();
    try {
      const results = await api.get('/api/productos', {
        query: { q: filter.value.trim() },
        signal: inFlight.signal,
      });
      collection.reset(results); // → the CatalogView re-renders
      if (empty) empty.hidden = results.length > 0;
    } catch (err) {
      if (err.name !== 'AbortError') console.error('search failed', err);
    }
  });
}
