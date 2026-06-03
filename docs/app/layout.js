// @ts-check
import { View } from 'lumenjs';
import { i18n, applyI18n } from './i18n.js';
import { API } from './api-data.js';

/**
 * The app shell: a sticky navbar (language + light/dark toggles) and a tree
 * sidebar (Inicio · Guía · Módulos→symbols), plus a `main` region the Router
 * swaps views into.
 */
export class Layout extends View {
  static template = '#layout';
  static regions = { main: 'outlet' };

  onMount() {
    const root = document.documentElement;
    const modes = ['', 'light', 'dark'];

    this.listen(this.ui.modeBtn, 'click', () => {
      const cur = root.dataset.mode || '';
      const next = modes[(modes.indexOf(cur) + 1) % modes.length];
      if (next) root.dataset.mode = next;
      else root.removeAttribute('data-mode');
      this.ui.modeBtn.textContent = '🌗 ' + (next || 'auto');
    });

    this.listen(this.ui.langBtn, 'click', () => {
      i18n.setLocale(i18n.locale === 'es' ? 'en' : 'es');
    });
    i18n.onChange(({ locale }) => {
      root.lang = locale;
      this.ui.langBtn.textContent = '🌐 ' + locale.toUpperCase();
    }, { signal: this.signal });

    // Build the Módulos tree: one <details> per package → General + symbols.
    for (const p of API) {
      const det = document.createElement('details');
      det.className = 'navpkg';
      const sum = document.createElement('summary');
      sum.textContent = p.pkg;
      det.appendChild(sum);
      const general = document.createElement('a');
      general.href = `#/m/${p.pkg}`;
      general.textContent = 'General';
      det.appendChild(general);
      for (const s of p.symbols) {
        const a = document.createElement('a');
        a.href = `#/m/${p.pkg}/${s.name}`;
        a.textContent = s.name;
        det.appendChild(a);
      }
      this.ui.modtree.appendChild(det);
    }

    applyI18n(this.el, this.signal); // sidebar labels (Inicio, Guía group…)
    this.ui.langBtn.textContent = '🌐 ' + i18n.locale.toUpperCase();

    this.listen(window, 'hashchange', () => this.syncActive());
    this.syncActive();
  }

  syncActive() {
    const cur = location.hash || '#/';
    for (const a of this.ui.menu.querySelectorAll('a')) {
      if (a.getAttribute('href') === cur) {
        a.setAttribute('aria-current', 'page');
        a.closest('details')?.setAttribute('open', ''); // expand its module
      } else {
        a.removeAttribute('aria-current');
      }
    }
  }
}
