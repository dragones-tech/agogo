// Vistas lumen del demo: una tarjeta por usuario (UserView), la lista que las
// reconcilia (UsersView), y un pequeño controlador que ata buscador, paginador,
// contador y "quitar de la vista". Mismo patrón que usarías para tu dominio.
import { View, CollectionView, Collection, $, fadeIn, slideOut } from 'lumen';
import { Users } from './collection.js';

const PAGE_SIZE = 6;

// Una tarjeta por usuario. Clona el <template id="ejemplo-user"> y rellena sus
// [data-ref] desde el model — actualización quirúrgica, sin re-render del DOM.
class UserView extends View {
  static template = '#ejemplo-user';

  onMount() {
    const user = this.props.model;
    this.ui.avatar.textContent = user.initials();
    this.ui.avatar.style.setProperty('--avatar-hue', user.hue());
    this.ui.name.textContent = user.get('name');
    this.ui.email.textContent = user.get('email');
    this.ui.email.href = `mailto:${user.get('email')}`;
    this.ui.company.textContent = user.companyName();
    this.listen(this.ui.remove, 'click', () => this.remove());
  }

  // Entrada: la tarjeta aparece con un fundido (lumen lo salta solo si el usuario
  // pidió menos movimiento). Como el render monta las tarjetas en paralelo, los
  // fundidos corren a la vez — rápido, no en cascada.
  animateIn() {
    return fadeIn(this.el, { duration: 180 });
  }

  // Salida (solo en el borrado deliberado): la tarjeta se desliza antes de irse,
  // y RECIÉN entonces avisamos al controlador para quitarla de la collection.
  async remove() {
    this.ui.remove.disabled = true;
    await slideOut(this.el, { duration: 160 });
    this.props.onRemove(this.props.model);
  }
}

// La lista: una UserView por model, reconciliada al vuelo cuando la collection
// emite add/remove/reset. El root es la grilla de tarjetas (.cards).
class UsersView extends CollectionView {
  static childView = UserView;

  render() {
    const el = document.createElement('div');
    el.className = 'cards';
    return el;
  }

  // Pasa el callback de borrado a cada tarjeta hija.
  childProps() {
    return { onRemove: this.props.onRemove };
  }
}

// Arranque: solo corre donde existe [data-ejemplo] (la sección de /ejemplo).
async function mount() {
  const host = $('[data-ejemplo]');
  if (!host) return;

  const status = $('[data-ejemplo-status]');
  const searchInput = $('[data-ejemplo-search]');
  const countLabel = $('[data-ejemplo-count]');
  const listHost = $('[data-ejemplo-list]');
  const pager = $('[data-ejemplo-pager]');
  const prevButton = $('[data-ejemplo-prev]');
  const nextButton = $('[data-ejemplo-next]');
  const pageLabel = $('[data-ejemplo-page]');

  const source = new Users(); // todos los usuarios (fuente de verdad)
  const shown = new Collection(); // el subconjunto visible (página filtrada)

  let term = '';
  let page = 1;

  // Consulta NO mutante: los usuarios que matchean la búsqueda.
  const matches = () => {
    const q = term.trim().toLowerCase();
    if (!q) return source.models;
    return source.filter(
      (u) =>
        u.get('name').toLowerCase().includes(q) ||
        u.get('email').toLowerCase().includes(q) ||
        u.companyName().toLowerCase().includes(q),
    );
  };

  // Actualiza contador y paginador a partir del total filtrado; clampa la página.
  const updateMeta = (all) => {
    const pages = Math.max(1, Math.ceil(all.length / PAGE_SIZE));
    page = Math.min(page, pages);
    countLabel.textContent = `${all.length} usuario${all.length === 1 ? '' : 's'}`;
    pageLabel.textContent = `Página ${page} de ${pages}`;
    prevButton.disabled = page <= 1;
    nextButton.disabled = page >= pages;
    pager.hidden = pages <= 1;
  };

  // Búsqueda / paginación: re-renderiza la página completa (las cards entran con
  // fundido). El reset es instantáneo al salir (no hay animateOut en el View).
  const render = () => {
    const all = matches();
    updateMeta(all);
    const start = (page - 1) * PAGE_SIZE;
    shown.reset(all.slice(start, start + PAGE_SIZE));
  };

  // Reconcilia `shown` con la página actual SIN reset: quita las que sobran y
  // AGREGA la que falta (el backfill). Así, al borrar, el hueco se rellena con el
  // siguiente usuario (entra con fundido) en vez de quedar la página corta.
  const syncPage = () => {
    const before = page;
    const all = matches();
    updateMeta(all); // puede clampar la página
    if (page !== before) {
      render(); // la página cambió (p. ej. se vació la última) → re-render limpio
      return;
    }
    const pageItems = all.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);
    for (const model of [...shown.models]) {
      if (!pageItems.includes(model)) shown.remove(model); // sale (unmount instantáneo)
    }
    for (const model of pageItems) {
      if (!shown.models.includes(model)) shown.add(model); // entra el backfill (animateIn)
    }
  };

  // Borrado: la card ya animó su salida en UserView.remove(). La quitamos de la
  // fuente y reconciliamos la página (rellena el hueco con el siguiente).
  const onRemove = (user) => {
    source.remove(user);
    syncPage();
  };

  await new UsersView({ collection: shown, onRemove }).mount(listHost);

  searchInput.addEventListener('input', () => {
    term = searchInput.value;
    page = 1;
    render();
  });
  prevButton.addEventListener('click', () => {
    page--;
    render();
  });
  nextButton.addEventListener('click', () => {
    page++;
    render();
  });

  // Carga el JSON público (lumen Http, en el navegador) y pinta la primera página.
  try {
    await source.load();
    status?.remove();
    render();
  } catch {
    if (status) {
      status.textContent = 'No se pudieron cargar los usuarios.';
      status.className = 'alert';
      status.dataset.variant = 'danger';
      status.removeAttribute('aria-busy');
    }
  }
}

mount();
