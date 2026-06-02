// Catálogo con lumen: Model + Collection + CollectionView.
//
// El servidor renderiza la lista (SEO, funciona sin JS) y emite los datos como
// un JSON island. Al cargar, lumen toma el control: arma una Collection de
// productos y un CollectionView que renderiza una vista por modelo. El filtro no
// oculta nodos a mano: hace collection.reset(subconjunto) y el CollectionView
// reconcilia (monta/desmonta solo lo que cambió). Degrada limpio: sin JS, queda
// la lista del servidor. Importamos solo los submódulos que usamos.
import { $ } from '/static/lumen/src/dom.js';
import { Model } from '/static/lumen/src/model.js';
import { Collection } from '/static/lumen/src/collection.js';
import { CollectionView } from '/static/lumen/src/collection-view.js';
import { View } from '/static/lumen/src/view.js';

// Un producto: dato observable. La validación/normalización viviría aquí.
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
  const todos = JSON.parse(datos.textContent); // [{slug,titulo,descripcion,precio}]
  const coleccion = new Collection(todos, Producto);
  const catalogo = new Catalogo({ collection: coleccion });

  // lumen reemplaza la lista del servidor por la suya, en el mismo lugar.
  const padre = listaServidor.parentNode;
  const ancla = listaServidor.nextSibling;
  listaServidor.remove();
  await catalogo.mount(padre);
  if (ancla) padre.insertBefore(catalogo.el, ancla);

  const norm = (s) => s.toLowerCase();
  filtro.addEventListener('input', () => {
    const q = norm(filtro.value.trim());
    const subconjunto = todos.filter((p) => norm(p.titulo).includes(q));
    coleccion.reset(subconjunto); // → el CollectionView re-renderiza
    if (vacio) vacio.hidden = subconjunto.length > 0;
  });
}
