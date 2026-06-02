// ProductItem: a catalog card. One view per model; it only touches the nodes
// it cares about (this.ui), without re-rendering the subtree.
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
