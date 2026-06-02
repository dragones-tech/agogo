// Mejora progresiva del catálogo: filtra la lista en el cliente con lumen.
// Sin JS, el servidor ya entrega el catálogo completo, así que esto solo añade
// comodidad (degrada limpio). Importamos SOLO el submódulo que usamos: con ES
// modules el navegador no descarga el resto de lumen.
import { $, $$ } from '/static/lumen/src/dom.js';

const filtro = $('[data-ref="filtro"]');
const lista = $('[data-ref="lista"]');

if (filtro && lista) {
  const items = $$('li[data-nombre]', lista);
  const vacio = $('[data-ref="vacio"]');

  filtro.addEventListener('input', () => {
    const q = filtro.value.trim().toLowerCase();
    let visibles = 0;
    for (const li of items) {
      const coincide = li.dataset.nombre.toLowerCase().includes(q);
      li.hidden = !coincide;
      if (coincide) visibles++;
    }
    if (vacio) vacio.hidden = visibles > 0;
  });
}
