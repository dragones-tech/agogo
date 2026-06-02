// Mejora progresiva del formulario de contacto: validación instantánea en el
// cliente con lumen, ESPEJO de las reglas del servidor (que siguen siendo la
// autoridad: el form lleva CSRF y el handler revalida). Sin JS, el servidor
// valida y re-renderiza igual. Importamos solo dom + validate de lumen.
import { $, refs } from '/static/lumen/src/dom.js';
import { runRules, required, email, minLength } from '/static/lumen/src/validate.js';

const form = $('[data-ref="form"]');

if (form) {
  const r = refs(form);
  const campos = ['nombre', 'email', 'mensaje'];

  // Las mismas reglas que internal/contacto/handler.go.
  const reglas = {
    nombre: [required('El nombre es obligatorio.')],
    email: [required('El email es obligatorio.'), email('Email inválido.')],
    mensaje: [minLength(10, 'El mensaje es muy corto (mínimo 10 caracteres).')],
  };

  const datos = () => ({
    nombre: r.nombre.value.trim(),
    email: r.email.value.trim(),
    mensaje: r.mensaje.value.trim(),
  });

  const pintar = (errs) => {
    for (const c of campos) r['error-' + c].textContent = errs[c]?.[0] ?? '';
  };

  // Al enviar: si hay errores, los mostramos y no enviamos.
  form.addEventListener('submit', (e) => {
    const errs = runRules(datos(), reglas);
    if (Object.keys(errs).length) {
      e.preventDefault();
      pintar(errs);
    }
  });

  // Al salir de un campo: validamos solo ese, sin gritar por los demás.
  for (const c of campos) {
    r[c].addEventListener('blur', () => {
      const errs = runRules(datos(), { [c]: reglas[c] });
      r['error-' + c].textContent = errs[c]?.[0] ?? '';
    });
  }
}
