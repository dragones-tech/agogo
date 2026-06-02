// CatalogView: la rejilla. Monta un ProductItem por modelo de la colección y
// reconcilia al recibir add/remove/reset (no reconstruye todo).
import { CollectionView } from '/static/lumen/src/collection-view.js';
import { ProductItem } from './product-item.js';

export class CatalogView extends CollectionView {
  static template = '#catalog-grid';
  static childView = ProductItem;
  static container = 'items';
}
