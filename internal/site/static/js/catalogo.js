// Catálogo con lumen: Model + Collection + CollectionView + Http.
//
// El servidor renderiza la lista (SEO, funciona sin JS) y emite los datos como
// un JSON island para el primer pintado SIN round-trip. Al teclear en el filtro,
// cada tecla dispara GET /api/productos?q=... (búsqueda SQL en el servidor); el
// resultado JSON resetea la Collection y el CollectionView reconcilia. Las
// peticiones en vuelo se cancelan al teclear de nuevo, así no llegan desordenadas.
import { $ } from '/static/lumen/src/dom.js';
import { Model } from '/static/lumen/src/model.js';
import { Collection } from '/static/lumen/src/collection.js';
import { CollectionView } from '/static/lumen/src/collection-view.js';
import { View } from '/static/lumen/src/view.js';
import { Http } from '/static/lumen/src/http.js';

// Un producto: dato observable.
class Producto extends Model {}

// Una tarjeta del catálogo: una vista por modelo.
class ProductoItem extends View {
  static template = '#producto-tpl';
  onMount() {
    const p = this.props.model;
    this.ui.enlace.textContent = p.get('titulo');
    this.ui.enlace.href = `/productos/${p.get('slug')}`;
    this.ui.precio.textContent = `$${p.get('precio')} MXN`;
  }
}

// La rejilla: un CollectionView que monta un ProductoItem por modelo.
class Catalogo extends CollectionView {
  static template = '#catalogo-tpl';
  static childView = ProductoItem;
  static container = 'items';
}

const datos = $('[data-ref="datos"]');
const listaServidor = $('[data-ref="lista"]');
const filtro = $('[data-ref="filtro"]');
const vacio = $('[data-ref="vacio"]');

if (datos && listaServidor && filtro) {
  // Primer pintado desde el JSON island (sin pedir nada al servidor).
  const coleccion = new Collection(JSON.parse(datos.textContent), Producto);
  const catalogo = new Catalogo({ collection: coleccion });

  // lumen reemplaza la lista del servidor por la suya, en el mismo lugar.
  const padre = listaServidor.parentNode;
  const ancla = listaServidor.nextSibling;
  listaServidor.remove();
  await catalogo.mount(padre);
  if (ancla) padre.insertBefore(catalogo.el, ancla);

  // Búsqueda en servidor: una petición JSON por tecla.
  const api = new Http();
  let enVuelo; // AbortController de la petición anterior
  filtro.addEventListener('input', async () => {
    enVuelo?.abort();
    enVuelo = new AbortController();
    try {
      const resultados = await api.get('/api/productos', {
        query: { q: filtro.value.trim() },
        signal: enVuelo.signal,
      });
      coleccion.reset(resultados); // → el CollectionView re-renderiza
      if (vacio) vacio.hidden = resultados.length > 0;
    } catch (e) {
      if (e.name !== 'AbortError') console.error('búsqueda falló', e);
    }
  });
}
