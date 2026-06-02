// Validación del formulario de contacto: instantánea en cliente con lumen,
// ESPEJO de las reglas del servidor (que sigue siendo la autoridad: CSRF +
// revalidación en el handler). Sin JS, el servidor valida y re-renderiza igual.
import { $, refs } from '/static/lumen/src/dom.js';
import { runRules, required, email, minLength } from '/static/lumen/src/validate.js';

const form = $('[data-ref="form"]');

if (form) {
  const ui = refs(form);
  // Nombres de campo del servidor (contrato del POST); no se traducen.
  const fields = ['nombre', 'email', 'mensaje'];

  // Las mismas reglas que internal/contacto/handler.go.
  const rules = {
    nombre: [required('El nombre es obligatorio.')],
    email: [required('El email es obligatorio.'), email('Email inválido.')],
    mensaje: [minLength(10, 'El mensaje es muy corto (mínimo 10 caracteres).')],
  };

  const values = () => ({
    nombre: ui.nombre.value.trim(),
    email: ui.email.value.trim(),
    mensaje: ui.mensaje.value.trim(),
  });

  const paint = (errors) => {
    for (const field of fields) ui['error-' + field].textContent = errors[field]?.[0] ?? '';
  };

  // Al enviar: si hay errores, los mostramos y no enviamos.
  form.addEventListener('submit', (event) => {
    const errors = runRules(values(), rules);
    if (Object.keys(errors).length) {
      event.preventDefault();
      paint(errors);
    }
  });

  // Al salir de un campo: validamos solo ese.
  for (const field of fields) {
    ui[field].addEventListener('blur', () => {
      const errors = runRules(values(), { [field]: rules[field] });
      ui['error-' + field].textContent = errors[field]?.[0] ?? '';
    });
  }
}
