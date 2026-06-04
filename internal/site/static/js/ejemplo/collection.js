// Collection de usuarios respaldada por un JSON público.
//
// La Collection de lumen es IN-MEMORY: no trae `fetch`. El traer datos es
// responsabilidad tuya — acá con el cliente `Http` de lumen (en el NAVEGADOR,
// no en el servidor) — y luego `reset()` reemplaza el contenido y emite el
// evento `reset`, que dispara el re-render del CollectionView.
//
// El fetch al navegador a un dominio externo necesita ese dominio en la
// directiva `connect-src` del CSP (ver internal/middleware).
import { Collection, Http } from 'lumen';
import { User } from './model.js';

const api = new Http({ baseURL: 'https://jsonplaceholder.typicode.com' });

export class Users extends Collection {
  constructor(items = []) {
    super(items, User); // cada objeto plano del JSON se envuelve en un User
  }

  /**
   * Trae los usuarios del JSON público y reemplaza el contenido (emite `reset`).
   * @param {AbortSignal} [signal] - para cancelar (p. ej. la `signal` de una vista).
   */
  async load(signal) {
    const data = await api.get('/users', { signal });
    this.reset(data);
    return this;
  }
}
