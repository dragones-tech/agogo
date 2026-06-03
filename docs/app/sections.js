// @ts-check
import { View } from 'lumenjs';
import { applyI18n } from './i18n.js';
import { highlightAll } from './highlight.js';
import { API } from './api-data.js';

const pkgOf = (name) => API.find((p) => p.pkg === name);

/** Render go-doc prose (blank-line-separated paragraphs) into `el` as <p>s. */
function setProse(el, text) {
  el.innerHTML = '';
  for (const para of (text || '').split(/\n\n+/)) {
    const p = document.createElement('p');
    p.textContent = para.replace(/\n/g, ' ').trim();
    if (p.textContent) el.appendChild(p);
  }
}

/** A titled code window (.code-window) with literal code (highlighted later). */
function codeWindow(title, code) {
  const win = document.createElement('div');
  win.className = 'code-window';
  const bar = document.createElement('div');
  bar.className = 'code-window-bar';
  bar.textContent = title;
  const pre = document.createElement('pre');
  const c = document.createElement('code');
  c.textContent = code;
  pre.appendChild(c);
  win.append(bar, pre);
  return win;
}

// ── Guide sections (bilingual prose + literal code) ───────────────────────
class Section extends View {
  onMount() {
    applyI18n(this.el, this.signal);
    highlightAll(this.el); // only <pre><code>, never the prose
  }
}
export class Home extends Section { static template = '#home'; }
export class Empezar extends Section { static template = '#empezar'; }
export class Estructura extends Section { static template = '#estructura'; }
export class Modulo extends Section { static template = '#modulo'; }
export class Activar extends Section { static template = '#activar'; }
export class Otw extends Section { static template = '#otw'; }
export class Config extends Section { static template = '#config'; }
export class NotFound extends Section { static template = '#notfound'; }

// ── Module reference (English; data from go doc) ──────────────────────────

/** A package's "General" page: overview + usage example + its symbols. */
export class ModulePage extends View {
  static template = '#modulepage';
  onMount() {
    const p = pkgOf(this.props.pkg);
    if (!p) { this.ui.name.textContent = this.props.pkg + ' — ?'; return; }
    this.ui.name.textContent = p.pkg;
    setProse(this.ui.doc, p.doc);
    if (p.example) this.ui.example.appendChild(codeWindow('Usage example', p.example));
    for (const s of p.symbols) {
      const a = document.createElement('a');
      a.href = `#/m/${p.pkg}/${s.name}`;
      a.innerHTML = `<b>${s.kind}</b> ${s.name}`;
      this.ui.syms.appendChild(a);
    }
    highlightAll(this.el);
  }
}

/** A single symbol's page: signature + doc (+ methods for a type). */
export class SymbolPage extends View {
  static template = '#symbolpage';
  onMount() {
    const p = pkgOf(this.props.pkg);
    const s = p && p.symbols.find((x) => x.name === this.props.name);
    if (!s) { this.ui.name.textContent = '?'; return; }
    this.ui.back.href = `#/m/${p.pkg}`;
    this.ui.back.textContent = '← ' + p.pkg;
    this.ui.name.textContent = s.name;
    this.ui.kind.textContent = s.kind;
    this.ui.sig.querySelector('code').textContent = s.sig;
    setProse(this.ui.doc, s.doc);

    this.ui.methods.innerHTML = '';
    if (s.methods && s.methods.length) {
      const h = document.createElement('h2');
      h.textContent = 'Methods';
      this.ui.methods.appendChild(h);
      for (const m of s.methods) {
        const h3 = document.createElement('h3');
        h3.className = 'apimethod';
        h3.textContent = m.name;
        this.ui.methods.appendChild(h3);
        this.ui.methods.appendChild(codeWindow(m.kind, m.sig));
        const d = document.createElement('div');
        setProse(d, m.doc);
        this.ui.methods.appendChild(d);
      }
    }
    highlightAll(this.el);
  }
}
