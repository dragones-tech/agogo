// @ts-check
import { Router } from 'lumenjs';
import { Layout } from './layout.js';
import {
  Home, Empezar, Estructura, Modulo, Activar, Config, NotFound,
  ModulePage, SymbolPage,
} from './sections.js';

// Mount the shell once; its `main` region is the router outlet.
const layout = await new Layout().mount(/** @type {HTMLElement} */ (document.querySelector('#app')));
const outlet = layout.regions.main;
outlet.transition = true; // native View Transitions where supported

new Router()
  // Home + guide (bilingual)
  .add('/',           () => outlet.show(new Home()))
  .add('/empezar',    () => outlet.show(new Empezar()))
  .add('/modulo',     () => outlet.show(new Modulo()))
  .add('/estructura', () => outlet.show(new Estructura()))
  .add('/activar',    () => outlet.show(new Activar()))
  .add('/config',     () => outlet.show(new Config()))
  // Module reference (English): /m/<pkg> and /m/<pkg>/<symbol>
  .add('/m/:pkg/:sym', (p) => outlet.show(new SymbolPage({ pkg: p.pkg, name: p.sym })))
  .add('/m/:pkg',      (p) => outlet.show(new ModulePage({ pkg: p.pkg })))
  .notFound(()        => outlet.show(new NotFound()))
  .start();
