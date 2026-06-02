// CatalogView: the grid. It mounts one ProductItem per model in the collection
// and reconciles on add/remove/reset (it doesn't rebuild everything).
import { CollectionView } from '/static/lumen/src/collection-view.js';
import { ProductItem } from './product-item.js';

export class CatalogView extends CollectionView {
  static template = '#catalog-grid';
  static childView = ProductItem;
  static container = 'items';
}
