// Inicializa Swagger UI apuntando a nuestra spec. Archivo propio y local, así
// la CSP de /docs no necesita 'unsafe-inline' para scripts.
window.onload = function () {
  window.ui = SwaggerUIBundle({
    url: "/openapi.json",
    dom_id: "#swagger-ui",
  });
};
