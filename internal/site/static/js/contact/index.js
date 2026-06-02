// Contact form validation: instant on the client with lumen, MIRRORING the
// server's rules (which remains the authority: CSRF + revalidation in the
// handler). Without JS, the server validates and re-renders all the same.
import { $, refs } from '/static/lumen/src/dom.js';
import { runRules, required, email, minLength } from '/static/lumen/src/validate.js';

const form = $('[data-ref="form"]');

if (form) {
  const ui = refs(form);
  // Server field names (the POST contract); not translated.
  const fields = ['nombre', 'email', 'mensaje'];

  // The same rules as internal/contacto/handler.go.
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

  // On submit: if there are errors, show them and don't submit.
  form.addEventListener('submit', (event) => {
    const errors = runRules(values(), rules);
    if (Object.keys(errors).length) {
      event.preventDefault();
      paint(errors);
    }
  });

  // On leaving a field: validate only that one.
  for (const field of fields) {
    ui[field].addEventListener('blur', () => {
      const errors = runRules(values(), { [field]: rules[field] });
      ui['error-' + field].textContent = errors[field]?.[0] ?? '';
    });
  }
}
