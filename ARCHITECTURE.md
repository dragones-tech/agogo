# Arquitectura de Agogo

Este documento explica **cómo está pensado** Agogo y **por qué**. Para
instrucciones de uso (correr, rutas, env vars) ver el [README](README.md).

> En una frase: **Go minimalista, orientado a dominios y dueños del código.**
> Diseño iterativo guiado por trade-offs honestos, stdlib-first, afinado hasta
> que la herramienta deja de estorbar.

---

## Principios

1. **Stdlib primero — "solo lo elemental".**
   El servidor HTTP, las plantillas, el JSON, el CSRF, los middlewares y la spec
   OpenAPI son **código nuestro sobre la biblioteca estándar de Go**. No hay
   framework web. La única dependencia inevitable es el driver de SQLite
   (`modernc.org/sqlite`, Go puro, para mantener el binario estático y la imagen
   `scratch`).

2. **"Sin magia" = sin cajas negras externas.**
   El rechazo no es a la abstracción, sino a las dependencias opacas de terceros
   que esconden comportamiento en runtime (p. ej. un ORM). Las abstracciones
   **propias** —genéricos, helpers, `Resource[T]`— sí valen: viven en el repo,
   las lees y las arreglas tú. `sqlc` genera *nuestro* código.

3. **Por dominio (package-by-feature).**
   Cada dominio se basta solo: su `handler.go`, su `sql/`, sus `templates/`, y
   (si tiene BD) `migrate.go` y `seed.go`. Para entender "productos" abres una
   carpeta, no saltas entre capas técnicas.

4. **SQL a mano con sqlc, sin ORM (patrón Repository).**
   Una función tipada por consulta, parametrizada (inyección imposible por
   diseño). La "API" de datos es el vocabulario del dominio, no primitivas SQL.

5. **YAGNI primero, DRY después.**
   No se abstrae por adelantado. La duplicación se extrae cuando se **gana**
   (p. ej. `web.Resource[T]` apareció al pensar en escala, no con 2 dominios).

6. **Caja de herramientas, no un mega-framework.**
   Piezas chicas y componibles. Cada dominio usa la que le toca; no todo es un
   `Resource[T]`.

7. **Núcleo mínimo + módulos acoplables.**
   El servidor es un host (`app`) al que se le acoplan módulos opt-in
   (`app.Use(...)`). Lo que no acoplas no entra al binario. Composición
   explícita, sin auto-registro mágico; el comportamiento se escribe, no se
   genera. Las URLs siguen desacopladas del dominio.

8. **El servidor solo sirve.**
   El esquema y los datos de ejemplo se gestionan **fuera del sitio**, como
   comandos de desarrollo (`cmd/migrate`, `cmd/seed`).

9. **Seguridad de primera clase.**
   Contenedor `scratch`, usuario no-root, cabeceras de seguridad, CSRF en
   formularios, SQL parametrizado, todo tipado.

10. **SEO-first + API JSON desde una sola fuente de verdad.**
    HTML renderizado en servidor (`html/template`, JSON-LD, sitemap) y JSON
    salen del mismo modelo.

---

## Estructura

```
.
├── main.go              # el "Gemfile": crea el App y acopla módulos con Use()
├── cmd/
│   ├── migrate/         # aplica el esquema (bootstrap; sirve en deploy)
│   └── seed/            # datos de ejemplo (solo desarrollo)
└── internal/
    │  ── NÚCLEO + utilidades (base, no son módulos) ──
    ├── app/             # HOST: App, interfaz Module, puntos de enganche
    ├── router/          # Router estilo Express sobre net/http (+ middleware)
    ├── config/          # configuración tipada y validada desde el entorno
    ├── session/         # sesiones por cookie firmada (HMAC), sin store
    ├── identity/           # servicio de identidad sobre sesión + Require
    ├── password/        # hashing PBKDF2-HMAC-SHA256 (stdlib)
    ├── web/             # Resource[T]: recurso de BD genérico
    ├── view/            # layout HTML compartido + Render
    ├── respond/         # respond.JSON / respond.Error
    ├── jsonld/          # builders schema.org
    ├── csrf/            # protección CSRF por doble cookie
    ├── sitemap/         # contrato del sitemap (URL, Source, Entries) — hoja
    ├── middleware/      # baseline de seguridad: Recover, SecurityHeaders
    │  ── MÓDULOS (opt-in; cada uno con module.go → Module()) ──
    ├── logs/            # middleware global de logging de acceso
    ├── auth/            # autenticación usuario/contraseña (reusa identity)
    ├── oauth/           # autenticación OAuth 2.0 (reusa identity), stdlib
    ├── productos/       # recurso de BD
    ├── blog/            # recurso de BD
    ├── paginas/         # páginas estáticas (sin BD)
    ├── contacto/        # formulario (GET muestra, POST procesa)
    ├── openapi/         # openapi.json + Swagger UI en /docs
    └── site/            # robots.txt, sitemap.xml, /static
```

**Dirección de dependencias:** los **módulos** importan `app` (el host) y las
utilidades base; **`app` no importa ningún módulo**. Las hojas (`router`,
`respond`, `view`, `sitemap`, `csrf`, `config`, `session`, `password`) no
dependen de otros paquetes internos. Sin ciclos.

---

## Núcleo + módulos

El servidor es un **núcleo mínimo (`app`) al que se le acoplan módulos**. Un
módulo implementa `Module{ Name() string; Register(*App) error }` y en su
`Register` se engancha al host: añade rutas (`a.Router.Get`), middleware global
(`a.UseMiddleware`), migraciones (`a.AddMigration`), fuentes de sitemap
(`a.AddSitemap`) y usa servicios compartidos (`a.DB`, `a.Session`, `a.Identity`).

`main.go` es el "Gemfile": `app.Use(productos.Module(), auth.Module(), ...)`.
**Lo que no acoplas no se importa, así que no entra al binario** (limpio si no
se usa). Las rutas de cada dominio viven en su módulo; el panorama global es la
lista de `app.Use(...)`. Distribución de plugins de terceros = Go modules
(`go get`) + implementar `Module`; sin registry propio.

---

## Las tres formas de dominio

No todo encaja en el mismo molde. Hay tres formas; cada módulo registra sus
handlers en su `Register`:

1. **Recurso de BD — `web.Resource[T]`** (productos, blog).
   Genérico, escrito una sola vez: lista/detalle × HTML/JSON. El dominio solo
   aporta comportamiento (queries sqlc, plantillas, metadatos SEO, JSON-LD).
   Expone `ListHTML`, `DetailHTML`, `ListJSON`, `DetailJSON`.

2. **Página estática** (paginas).
   Sin BD: plantilla + SEO. Expone un `http.HandlerFunc` (p. ej.
   `paginas.QuienesSomos(baseURL)`).

3. **Formulario / handler propio** (contacto).
   Lógica única (validación, CSRF, PRG). Expone sus handlers (`Mostrar`,
   `Recibir`).

---

## Routing

Cada **módulo registra sus rutas** en su `Register(a *app.App)`, usando el router
(estilo Express, con middleware por ruta):

```go
// internal/productos/module.go
func (mod) Register(a *app.App) error {
    res := New(db.New(a.DB), a.Config.BaseURL)
    r := a.Router
    r.Get("/{$}", res.ListHTML)                       // home (match exacto de "/")
    r.Get("/productos/{slug}", res.DetailHTML)
    r.Get("/api/productos", res.ListJSON)
    a.AddSitemap(res.SitemapSource("/", "/productos"))
    a.AddMigration(Migrate)
    return nil
}
// internal/auth/module.go — auth como middleware por ruta:
r.Get("/cuenta", h.Cuenta, a.Identity.Require)
```

El **panorama global** de qué hace el servidor es la lista `app.Use(...)` en
`main.go`; el detalle de cada dominio vive en su `module.go`. (Antes hubo un
`routes.go` central y plano; se cambió por el modelo modular al volver el
proyecto un framework — el trade fue perder esa tabla única.)

Las URLs siguen **desacopladas**: cada handler calcula su `canonical` desde la
petición (`baseURL + r.URL.Path`), así el mismo handler sirve en cualquier ruta
donde el módulo lo monte.

Dos niveles de middleware: **global** (módulos vía `a.UseMiddleware`, p. ej.
`logs`; más el baseline de seguridad del núcleo) y **por ruta** (`r.Get(path, fn,
mw...)`, p. ej. `a.Identity.Require`).

---

## Capa de datos

- **SQLite** vía `modernc.org/sqlite` (Go puro, sin cgo → binario estático). Se
  abre en modo **WAL** con **`busy_timeout`** (`Config.DSN()`): lecturas
  concurrentes con una escritura, y las escrituras esperan en vez de fallar con
  `database is locked`.
- **sqlc** genera, por dominio, una función tipada por consulta a partir de
  `internal/<dominio>/sql/queries.sql` y `schema.sql`. El código generado vive
  en `internal/<dominio>/db/` (no se edita).
- Las consultas son explícitas y parametrizadas. Una escritura (`:exec`) y una
  lectura (`:one`/`:many`) tienen el mismo trato.

Flujo para añadir una consulta: editar `queries.sql` → `sqlc generate` → usar la
función generada.

---

## Desarrollo vs. sitio

El binario del servidor **no** crea esquema ni siembra datos. Eso vive aparte:

- `go run ./cmd/migrate` — aplica el esquema (bootstrap; también en deploy).
- `go run ./cmd/seed` — datos de ejemplo (solo desarrollo; idempotente).

Ambos comandos **arman el mismo `App` que el servidor** (`app.Use(...)`) y corren
`app.Migrate(ctx)`, que ejecuta las migraciones que cada módulo registró con
`a.AddMigration`. Añadir un dominio con BD = acoplarlo a la lista; su esquema se
aplica solo, sin tocar los `cmd/*`.

Así el "sitio" no contiene lógica de desarrollo, y el linker descarta el código
de seed del binario del servidor.

---

## Seguridad

- Imagen `scratch` (sin SO, sin shell), usuario no-root.
- Middlewares del núcleo: `Recover` (un panic no tumba el server), `LimitBody`
  (tope de 1 MiB por petición, antes de cualquier `ParseForm`/decode),
  `SecurityHeaders` (`Content-Security-Policy: default-src 'self'`,
  `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`) y `Gzip`
  (comprime respuestas de texto —HTML/CSS/JS/JSON— si el cliente lo acepta;
  `compress/gzip` de la stdlib, decide por `Content-Type`, respeta `Range`).
- Cookies `Secure`: se deducen del esquema de `AGOGO_BASE_URL` (`https://` →
  `Secure`) y se pueden forzar con `AGOGO_SECURE_COOKIES`. Aplica a la cookie de
  sesión y a la de CSRF.
- CSRF por doble cookie en formularios (`internal/csrf`).
- Páginas autenticadas (`/cuenta`, `/oauth/me`) responden `Cache-Control: no-store`.
- SQL parametrizado (sqlc) → inyección imposible por diseño.
- Swagger UI (`/docs`) está **vendorizado** (assets embebidos en
  `internal/openapi/static`, servidos desde `/docs-assets/`): self-contained, sin
  CDN, CSP `'self'` (solo `style-src` permite `'unsafe-inline'` por los estilos
  que Swagger inyecta en runtime).

---

## Errores y operación

- **Errores que se ven.** Un fallo del servidor se **registra con método y ruta**
  (`respond.ServerError` para JSON, `view.ServerError` para HTML) pero al cliente
  solo le llega un 500 genérico — nunca se filtran detalles internos. Antes el
  error real se descartaba; ahora queda en el log.
- **Páginas de error con el sitio.** 404 y 500 se renderizan con el layout
  (`view.NotFound` / `view.ServerError`, plantilla `error.html`), no como texto
  pelón. Un catch-all (`/`) en el módulo `site` da el 404 con estilo a cualquier
  ruta no registrada.
- **Apagado ordenado.** `main` escucha `SIGINT`/`SIGTERM`, deja de aceptar, da
  hasta 10s a las peticiones en curso (`srv.Shutdown`) y recién entonces vuelve,
  así corren los `defer` (incl. cerrar la BD). Stdlib (`os/signal`).

---

## Frontend: misma filosofía, otra capa

El frontend es libre (agogo solo sirve HTML), pero de cajón se apoya en dos
proyectos hermanos con **el mismo principio** que agogo —sin build, sin magia,
primitivas nativas, todo legible y auditable—, uno por cada capa de la web:

- **[scc](https://github.com/dragones-tech/scc)** (CSS): estilo de HTML semántico
  con primitivas nativas (`@layer`, `@scope`, custom properties). La regla de oro:
  una clase dice *qué es* (`card`, `field`, `alert`), nunca *cómo se ve*. La
  presentación vive en el CSS, el HTML queda legible.
- **[lumen](https://github.com/dragones-tech/lumen)** (JS): UI vanilla-JS sin
  build, con `<template>` nativo y ES modules (el navegador solo baja lo que
  importas).

Ambos se **vendorizan** en `internal/site/static/` (servidos desde `'self'`, sin
CDN, CSP estricta intacta — mismo trato que Swagger UI). Se usan como **mejora
progresiva**: el servidor renderiza y valida igual sin JS, y lumen solo añade
comodidad (filtro de catálogo, validación de formulario espejo del servidor).

### Política de vendorizado

scc y lumen **se trabajan en sus repos** (`dragones-tech/scc`, `dragones-tech/lumen`),
no aquí. En agogo son una **copia pristina** y **no se versionan** (`.gitignore`):
se re-bajan con `./scripts/vendor-frontend.sh`, como un `npm install`. Por eso el
vendor se mantiene **sin editar** — cualquier mejora va a su repo, no a esta copia.

Lo específico de agogo NO vive en el vendor, sino en `internal/site/static/style.css`:

- **La marca** (verde): sobrescribe las "perillas" de scc (`--accent`, `--neutral-hue`,
  `--radius`) desde `:root`. Al estar sin `@layer`, gana a `scc/theme.css` sin tocarlo.
- **El marco de página** (centrado, footer abajo, rejilla de tarjetas).
- **`[hidden] { display: none }`**: scc no re-asegura `[hidden]` y un componente con
  `display` propio (p. ej. `.card`) lo anularía. *(Candidato a aportar al reset de scc.)*
- **`@view-transition { navigation: auto }`**: crossfade nativo entre páginas (quita
  el parpadeo del recargo MPA), con fondo opaco en `<html>`.

Para clonar y correr: tras `git clone`, ejecutar `./scripts/vendor-frontend.sh`
una vez (trae scc/lumen); luego `go run .` como siempre.

### Navegación parcial (estilo Turbo, propia)

Pieza **propia de agogo** (no de lumen), del lado servidor + un JS chico: al
cambiar de sección no se recarga el documento, solo se reemplaza `<main>`. Header,
footer, CSS y el propio script quedan fijos → sin parpadeo, sin re-parsear assets.

Aprovecha que el bloque **`content` ya es un partial**:

- **Servidor** (`view.Render`): si la petición trae la cabecera `X-Fragment`,
  renderiza solo el bloque `fragment` (título + `content` + `scripts`), no la
  página entera. Responde con `Vary: X-Fragment`.
- **Cliente** (`static/js/nav.js`): intercepta los clics en enlaces internos,
  pide el fragmento, y reemplaza `<main>` (con `document.startViewTransition` si
  hay soporte, para un crossfade en la misma página). Reejecuta los `<script>` de
  la página con una URL única (`?v=`) para recablear el `<main>` nuevo; sus
  imports siguen cacheados. Actualiza título e historial (`pushState`/`popstate`).

**Degrada limpio:** sin JS, ante un error de red, una respuesta no-HTML (un
`/api/...` JSON), un acceso directo, un refresh o un crawler → navegación normal
del navegador con la página completa. El SEO y la robustez no dependen del JS;
esto es solo una mejora de experiencia encima. No reemplaza a lumen ni hace
streaming (no es Turbo completo): solo evita re-renderizar lo que no cambia.

## Configuración y sesiones

- **`config`**: lee el entorno **una sola vez**, tipado y validado
  (`config.Load() (Config, error)`). Lo usan el servidor y los comandos. Incluye
  `SecretKey` (firma de sesiones): si falta, cae a una clave de desarrollo
  insegura y marca `DevSecret` (el servidor avisa); si se define, exige ≥32
  caracteres.
- **`session`**: sesiones por **cookie firmada con HMAC-SHA256** (`crypto/hmac`),
  sin store en servidor y sin dependencias — misma línea que `csrf`. La cookie
  está firmada (a prueba de manipulación) pero **no cifrada**: no guardar
  secretos en ella. Lleva una **expiración firmada** (sello `_exp`, 7 días): el
  servidor la impone aunque el cliente conserve el archivo, así una cookie robada
  no vale para siempre. Uso actual: mensaje *flash* del formulario de contacto
  (Post-Redirect-Get) e id de usuario. Límite ~4 KB por cookie.

## Autenticación

- **`password`**: hashing **PBKDF2-HMAC-SHA256** con `crypto/pbkdf2` (stdlib
  desde Go 1.24), salt aleatorio y comparación en tiempo constante. El hash
  guardado lleva sus parámetros (`pbkdf2-sha256$iter$salt$hash`), así se puede
  migrar el algoritmo sin romper hashes existentes. (bcrypt/argon2 serían
  preferibles pero son dependencia; se eligió stdlib.)
- **`identity`** (base): el **mecanismo** de identidad sobre la sesión. Guarda
  `user_id` en la cookie firmada; ofrece `Login`/`Logout`/`UserID` y `Require`,
  que es un **middleware** normal aplicado por ruta (`r.Get("/cuenta", fn,
  a.Identity.Require)`) o por bloque (`a.Identity.Group(r)`). Siempre disponible
  en el núcleo, para que cualquier módulo proteja rutas.
- **`auth`** (módulo): la **funcionalidad** — tabla `usuarios`, login/logout y la
  página protegida `/cuenta`. Verifica credenciales (con `password`) y delega en
  `a.Identity.Login`. No revela si un correo existe; login y logout llevan CSRF.
  Otro módulo (`oauth`, `apikeys`) podría reusar `identity` sin reescribir nada.

**Dos niveles de middleware:** *global* (el módulo `logs` vía `a.UseMiddleware`, más
el baseline de seguridad del núcleo en `app.Handler`) y *por ruta* (el router acepta
`mw ...Middleware` tras el handler). `identity.Require` es solo un middleware más;
mañana caben rate-limit, roles, etc. por la misma vía.

## Metodología de trabajo

El ciclo que seguimos: **Discutir → decidir → construir → verificar → documentar.**

- Se proponen opciones con **trade-offs honestos**; se decide; se implementa.
- **Verificar corriendo**: `go build ./...`, `go vet ./...` y `curl` en cada
  cambio. Nada se da por bueno sin verlo responder.
- **Refactor sin miedo**: el diseño evoluciona (el routing se rehízo varias
  veces hasta dar con la forma plana y desacoplada).
- Las decisiones se capturan; las no urgentes se documentan como pendientes en
  vez de forzarlas.

---

## Referencias

Proyectos hermanos (mismo ADN, una capa cada uno):

- **[lumen](https://github.com/dragones-tech/lumen)** — UI vanilla-JS, sin build,
  sin magia. La capa de cliente.
- **[scc](https://github.com/dragones-tech/scc)** — Scoped Cascade CSS, semántico,
  sin build. La capa de estilo.

El diseño coincide, sin haberlo buscado, con las prácticas canónicas de Go:

- **Mat Ryer — "How I write HTTP services in Go after 13 years"** (Grafana): el
  patrón `routes.go` como único lugar de rutas.
- **Ben Johnson — [`benbjohnson/wtf`](https://github.com/benbjohnson/wtf)**:
  organización por dominio con interfaces como contratos.
- **[`pocketbase/pocketbase`](https://github.com/pocketbase/pocketbase)**: misma
  filosofía de binario único + SQLite embebido.
- **[`walterwanderley/sqlc-http`](https://github.com/walterwanderley/sqlc-http)**:
  mismo stack (net/http + sqlc), pero *generando* la capa HTTP desde el SQL —
  el enfoque opuesto al nuestro (nosotros la escribimos y desacoplamos a mano).
