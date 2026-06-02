// ProductItem: una tarjeta del catálogo. Una vista por modelo; solo toca los
// nodos que le importan (this.ui), sin re-render del subárbol.
import { View } from '/static/lumen/src/view.js';

export class ProductItem extends View {
  static template = '#product-item';

  onMount() {
    const product = this.props.model;
    this.ui.link.textContent = product.get('titulo');
    this.ui.link.href = `/productos/${product.get('slug')}`;
    this.ui.price.textContent = `$${product.get('precio')} MXN`;
  }
}
