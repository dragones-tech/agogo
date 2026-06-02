# Agogo

Servidor web compacto en **Go puro**: HTML renderizado en servidor (para SEO) y
API JSON, desde una sola fuente de verdad. Sin frameworks ni librerías "todo en
uno" — solo la biblioteca estándar y herramientas de nicho.

> Para el **porqué** del diseño (principios, formas de dominio, metodología),
> ver [ARCHITECTURE.md](ARCHITECTURE.md).

## Filosofía

- **Solo lo elemental.** Servidor HTTP, plantillas y JSON salen de la stdlib de
  Go (`net/http`, `html/template`, `encoding/json`). Cero framework web.
- **Herramientas de nicho, no genéricas.** La capa de datos usa el patrón
  *Repository*: cada método es exactamente una consulta que el proyecto necesita.
  Las consultas las genera [sqlc](https://sqlc.dev) a partir de SQL real — **una
  función tipada por query**, sin `.where()/.order()` genéricos ni ORM con magia.
- **Contenedor mínimo.** Compila a un binario estático (`CGO_ENABLED=0`) que corre
  en una imagen `scratch` (~17 MB), como usuario no-root, sin SO dentro.

> Nota honesta sobre dependencias: el servidor/SEO/API es stdlib pura. El único
> componente con dependencias es el driver `modernc.org/sqlite` (SQLite en Go
> puro, necesario para mantener el binario estático y la imagen `scratch`); este
> arrastra unas pocas dependencias indirectas. Es el precio mínimo de tener
> SQLite sin `cgo`.

## Por qué Agogo

- **Un solo binario, cero sorpresas.** Compila a un ejecutable estático que corre
  en una imagen `scratch` (~20 MB), como usuario no-root, sin SO ni runtime.
  Arranca en milisegundos.
- **Sin framework, solo stdlib.** Menos superficie de ataque, nada de "magia" en
  runtime, fácil de auditar. La única dependencia es el driver SQLite (Go puro).
- **Modular: pagas solo lo que usas.** Acoplas `auth`, `oauth`, `logs`, `blog`…
  con una línea; lo que no acoplas no entra al binario. Tu `main.go` es la lista
  de lo que tu app *es*.
- **SEO + API desde una sola fuente.** HTML renderizado en servidor (con JSON-LD
  y sitemap) y API JSON con OpenAPI, del mismo modelo de datos.
- **Seguro por defecto.** CSRF en formularios, cabeceras de seguridad, sesiones
  firmadas (HMAC), SQL parametrizado (sqlc) y hashing PBKDF2 — todo stdlib.
- **Type-safe y tuyo.** Sin ORM ni cajas negras: el código que corre es el que
  lees. Extender el servidor = escribir un `module.go`.
- **Despliegue trivial.** Un binario o una imagen mínima; nada que instalar en el
  servidor.

## Frontend: la misma filosofía en las tres capas

El frontend es **tu elección** — agogo sirve HTML y no te ata a nada. Pero de
cajón usa una trinca hermana que comparte exactamente el mismo ADN (sin build,
sin magia, primitivas nativas, "lo que escribes es lo que corre"), una capa por
cada parte de la web:

| Capa | Proyecto | Qué hace |
|------|----------|----------|
| Servidor (Go) | **[agogo](https://github.com/dragones-tech/agogo)** | HTML server-rendered + API JSON desde una fuente |
| Estilo (CSS) | **[scc](https://github.com/dragones-tech/scc)** | CSS semántico: las clases dicen *qué es*, no *cómo se ve* |
| UI (JS) | **[lumen](https://github.com/dragones-tech/lumen)** | UI vanilla-JS, sin build, `<template>` nativo |

scc y lumen **se trabajan en sus repos**, no aquí: en agogo son una copia
**vendorizada** en `internal/site/static/` (servida desde `'self'`, sin CDN, CSP
estricta — igual que Swagger UI) que **no se versiona** y se re-baja con
`./scripts/vendor-frontend.sh`. La marca de agogo y los ajustes de página viven
en `static/style.css` (no se edita el vendor). Ver [ARCHITECTURE.md](ARCHITECTURE.md#frontend-misma-filosofía-otra-capa).

Los ejemplos ya los usan como **mejora progresiva** (el servidor renderiza y
valida igual sin JS):

- **scc** estila el HTML semántico de todas las páginas (`class="card"`,
  `class="field"`, `class="alert"`, `<button data-variant="primary">`…).
- **lumen** añade comodidad en cliente: filtro instantáneo del catálogo
  (`static/js/catalogo.js`) y validación del formulario de contacto espejo de
  las reglas del servidor (`static/js/contacto.js`).

Cámbialos por lo que quieras: la filosofía no te obliga a usarlos, solo a que,
si los usas, los tres encajan sin fricción.

## Estructura (package-by-feature)

```
.
├── main.go                      # el "Gemfile": crea el App y acopla módulos con app.Use(...)
├── sqlc.yaml                    # configuración de sqlc
├── Dockerfile                   # multi-stage → scratch
├── cmd/
│   ├── migrate/                 # comando: aplica el esquema (bootstrap)
│   └── seed/                    # comando: datos de ejemplo (solo desarrollo)
└── internal/
    │  ── NÚCLEO + utilidades (base, no son módulos) ──
    ├── app/                      # HOST: App, interfaz Module, puntos de enganche
    ├── router/                   # Router estilo Express (r.Get/r.Post + middleware)
    ├── config/                   # carga de entorno tipada y validada
    ├── session/                  # sesiones por cookie firmada (HMAC)
    ├── identity/                    # servicio de identidad sobre sesión + Require
    ├── password/                 # hashing PBKDF2 (stdlib)
    ├── web/                      # Resource[T]: recurso de BD genérico
    ├── view/                     # layout HTML compartido + Render
    ├── respond/                  # respond.JSON / respond.Error
    ├── jsonld/                   # builders schema.org
    ├── csrf/                     # protección CSRF (doble cookie)
    ├── sitemap/                  # contrato del sitemap (URL, Source, Entries)
    ├── middleware/               # baseline de seguridad: Recover, SecurityHeaders
    │  ── MÓDULOS (opt-in; cada uno con module.go → Module()) ──
    ├── logs/                     # middleware global de logging
    ├── productos/  blog/         # recursos de BD (handler.go, module.go, sql/, db/, templates/)
    ├── paginas/                  # páginas estáticas (sin BD)
    ├── contacto/                 # formulario (GET/POST + CSRF + flash)
    ├── auth/                    # autenticación usuario/contraseña (login/cuenta, reusa identity)
    ├── oauth/                    # autenticación OAuth 2.0 (reusa identity), stdlib
    ├── openapi/                  # openapi.json + Swagger UI vendorizado
    └── site/                     # robots.txt, sitemap.xml, /static
        └── static/
            ├── scc/              # vendor (gitignored): CSS semántico — re-baja con scripts/vendor-frontend.sh
            ├── lumen/            # vendor (gitignored): UI vanilla-JS — idem
            ├── js/               # nuestro JS, package-by-feature (como el Go)
            │   ├── catalog/      #   product.js (Model) · product-item.js, catalog-view.js (View) · index.js (entry)
            │   └── contact/      #   index.js (validación del form)
            └── style.css         # marco de página (layout) + marca de agogo; el diseño lo pone scc
```

El servidor es un **núcleo (`app`) + módulos acoplables**. Cada módulo tiene un
`module.go` con `Module()` y registra en su `Register(*App)` sus rutas,
migraciones y fuentes de sitemap. `main.go` los acopla con `app.Use(...)` — esa
lista es el panorama de qué hace el servidor. Ver [ARCHITECTURE.md](ARCHITECTURE.md)
para el detalle del modelo.

### Recurso genérico (`web.Resource[T]`)

Los 4 handlers (lista/detalle × HTML/JSON), el registro de rutas y el aporte al
sitemap están escritos **una sola vez** en `internal/web/resource.go`, como un
genérico `Resource[T]`. Un dominio no reimplementa nada de eso: solo **configura**
un `Resource` con sus tipos, queries (sqlc), plantillas y metadatos SEO. Añadir
una ruta/comportamiento común = editar `resource.go` una vez y los N dominios lo
heredan. Es código nuestro y type-safe (genéricos en compilación, sin reflexión).

## Rutas

| Método y ruta             | Respuesta                                            |
|---------------------------|------------------------------------------------------|
| `GET /`                   | HTML — catálogo (title, description, canonical, OG)  |
| `GET /productos/{slug}`   | HTML — detalle + JSON-LD `schema.org/Product`        |
| `GET /api/productos`      | JSON — lista                                         |
| `GET /api/productos/{slug}` | JSON — detalle (404 si no existe)                  |
| `GET /blog`               | HTML — lista de artículos                            |
| `GET /blog/{slug}`        | HTML — artículo + JSON-LD `schema.org/BlogPosting`   |
| `GET /api/posts`          | JSON — lista                                         |
| `GET /api/posts/{slug}`   | JSON — detalle (404 si no existe)                    |
| `GET /quienes-somos`      | HTML — página estática (sin BD)                      |
| `GET /contacto`           | HTML — formulario (con token CSRF)                   |
| `POST /contacto`          | procesa el formulario: valida, guarda y redirige (PRG)|
| `GET /login`              | HTML — formulario de acceso                          |
| `POST /login`             | autentica (verifica hash, inicia sesión, PRG)        |
| `POST /logout`            | cierra sesión                                        |
| `GET /cuenta`             | HTML — **protegida** (redirige a /login sin sesión)  |
| `GET /oauth/login`        | inicia OAuth 2.0 (redirige al proveedor)             |
| `GET /oauth/callback`     | callback OAuth: token → userinfo → `identity.Login`  |
| `GET /oauth/me`           | HTML — **protegida** (muestra la identidad OAuth)    |
| `GET /openapi.json`       | spec OpenAPI 3.0 de la API (escrita a mano, sin deps)|
| `GET /docs`               | Swagger UI (vendorizado, self-contained)             |
| `GET /docs-assets/...`    | assets de Swagger UI embebidos (css/js)              |
| `GET /sitemap.xml`        | XML — sitemap agregado de todos los dominios         |
| `GET /robots.txt`         | texto — apunta al sitemap                            |
| `GET /salud`              | `ok` — healthcheck                                   |

## Instalación

Requisitos: **Go 1.26+**. (Docker es opcional, para la imagen de contenedor;
`sqlc` es opcional, **solo** para regenerar código tras tocar el SQL.)

Agogo no es una librería que se instale con `go get`: es una **base que clonas y
haces tuya** (los paquetes del framework viven en `internal/`, a propósito — el
código que corre es el que lees y editas).

```bash
git clone https://github.com/dragones-tech/agogo
cd agogo
./scripts/vendor-frontend.sh   # baja scc + lumen (no se versionan aquí; viven en sus repos)
```

Y lo corres como se explica abajo en **Cómo correrlo**.

## Cómo correrlo

### Local

El servidor **solo sirve**: no toca el esquema ni siembra datos (eso vive fuera
del sitio, como comandos de desarrollo). En una BD nueva:

```bash
go run ./cmd/migrate   # aplica el esquema (bootstrap; sirve también en deploy)
go run ./cmd/seed      # datos de ejemplo (SOLO desarrollo; idempotente)
go run .               # arranca el servidor → http://localhost:8888
```

En desarrollo basta `go run ./cmd/seed` (asegura el esquema y luego siembra) y
después `go run .`.

### En contenedor

```bash
docker build -t agogo .
docker run -p 8888:8888 agogo
# → http://localhost:8888
```

La BD vive en `/data` dentro del contenedor. Para persistirla, monta un volumen:

```bash
docker run -p 8888:8888 -v agogo-data:/data agogo
```

## Variables de entorno

| Variable             | Por defecto                | Descripción                  |
|----------------------|----------------------------|------------------------------|
| `AGOGO_ADDR`      | `:8888`                    | Dirección de escucha         |
| `AGOGO_DB`        | `agogo.db`              | Ruta del archivo SQLite      |
| `AGOGO_BASE_URL`  | `http://localhost:8888`    | URL base (canonical, sitemap)|
| `AGOGO_SECRET_KEY`| (clave de dev insegura)    | Firma de sesiones (HMAC). En producción, ≥32 chars. |
| `OAUTH_CLIENT_ID` … | (vacío)                    | Config del módulo `oauth` (CLIENT_ID/SECRET, AUTH/TOKEN/USERINFO_URL, SCOPE). Sin esto, `/oauth/login` da 503. |

## Protocolos HTTP (HTTP/1.1, HTTP/2, HTTP/3)

> Estado en etapa de pruebas. El binario arranca hoy en texto plano
> (`ListenAndServe`), por lo que habla **HTTP/1.1**.

| Protocolo | Estado | Cómo se obtiene |
|-----------|--------|-----------------|
| **HTTP/1.1** | ✅ Activo | Por defecto en `net/http` |
| **HTTP/2** | ⚙️ A un paso | Servir sobre TLS → `net/http` negocia h2 vía ALPN, **sin dependencias** |
| **HTTP/3** | 🚫 No en el binario | QUIC no está en la stdlib de Go; se termina en el borde (proxy/CDN) |

### HTTP/2 dentro del binario (stdlib pura)

`net/http` activa HTTP/2 **automáticamente sobre TLS** (incrustado desde Go 1.6,
cero dependencias). Basta con servir con certificados:

```go
// h2 se negocia solo vía ALPN cuando el cliente lo soporta
srv.ListenAndServeTLS(certFile, keyFile)
```

HTTP/2 en texto plano (h2c) **no** está activo por defecto y requeriría
`golang.org/x/net/http2/h2c` (una dependencia), así que no lo usamos: HTTP/2 = TLS.

### Despliegue recomendado: terminar h2/h3 en el borde

La práctica estándar es **no** meter QUIC ni TLS en el binario, sino terminar los
protocolos modernos en un reverse proxy / CDN y dejar la app compacta:

```
Cliente ──h3/h2/TLS──► Caddy / nginx / CDN ──h1.1 o h2──► binario Go
        (UDP + TCP)        (el borde)        (red privada / socket)
```

El borde habla **HTTP/3 + HTTP/2 + TLS** hacia fuera; el binario sigue simple,
sin dependencias y sin gestionar certificados. (El binario ya **comprime con
gzip** las respuestas de texto por su cuenta —`compress/gzip`, stdlib—, así que
funciona bien incluso sin proxy; el borde añade zstd/brotli si quieres.) Ejemplo mínimo con Caddy
(HTTP/2 y HTTP/3 + TLS automático, listo de fábrica):

```caddyfile
# Caddyfile (ejemplo de despliegue, no se incluye en la imagen)
agogo.com {
	encode gzip zstd
	reverse_proxy localhost:8888   # o unix//run/agogo.sock
}
```

Caddy sirve HTTP/3 por defecto; no hace falta tocar el código Go.

## Estándares web

"Estándares web" tiene dos sentidos y conviene no confundirlos:

- **Protocolos y formatos de la web (HTTP, HTML5, JSON, TLS, sitemap,
  JSON-LD):** sí, plenamente. `net/http` implementa los RFC del IETF
  directamente y el HTML/JSON/XML que emitimos sigue los estándares
  correspondientes.
- **La "Web Platform API" (objetos Fetch `Request`/`Response`), que es lo que
  frameworks JS como Hono llaman "web standards":** no aplica a Go. Esa
  abstracción da portabilidad entre runtimes de JavaScript (Workers, Deno, Bun).
  Aquí la portabilidad se logra de otra forma: **un binario estático que corre
  en cualquier Linux sin runtime**. Mismo objetivo (no atarse a una API
  propietaria), distinto camino.

## Decisiones pendientes (para decidir después)

> **Nota:** Agogo se construye como **framework**, no como un sitio. Las
> decisiones se evalúan para el caso abierto ("qué se construirá encima"), no
> para esta app de ejemplo. Ver [ARCHITECTURE.md](ARCHITECTURE.md).

### Otras pendientes

- **Scaffold de dominios**: generador `cmd/scaffold` para crear un dominio
  nuevo. Como framework, sube de prioridad (es cómo se crean cosas encima); se
  hace cuando la forma del dominio esté estable.
- **Spec OpenAPI a mano vs generada**: hoy se mantiene a mano (coherente con
  "código nuestro"); como framework, reconsiderar un helper que la genere.

### Resueltas

- **`jsonLD`**: ✅ resuelto con `internal/jsonld` — primitivo extensible
  `Script(typ, props)` (cualquier tipo schema.org) + builders tipados
  (`Product`, `BlogPosting`). Reemplazó los mapas inline por dominio.
- **Swagger UI**: ✅ **vendorizado**. Los assets viven en
  `internal/openapi/static/` (embebidos), servidos en `/docs-assets/`. `/docs`
  ya no usa CDN y su CSP es estricta (`'self'`). Costo: +~2 MB en el binario del
  servidor (solo ahí; `cmd/*` no embeben Swagger).

## Añadir una consulta (flujo sqlc)

1. Escribe la consulta en `internal/productos/sql/queries.sql`:

   ```sql
   -- name: ListDestacados :many
   SELECT slug, titulo, descripcion, precio
   FROM productos WHERE precio >= ? ORDER BY precio DESC;
   ```

2. Regenera el código tipado:

   ```bash
   # sqlc se instala con: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   # y queda en el GOBIN de Go (puede no estar en el PATH por defecto):
   export PATH="$PATH:$(go env GOBIN)"
   sqlc generate
   ```

3. Usa la función generada (`q.ListDestacados(ctx, 100)`) en tu handler.

Lo que no escribes, no existe: no hay constructor de queries universal.

## Añadir un dominio nuevo

1. Crea `internal/<dominio>/` con su `handler.go`, `sql/`, `templates/`.
2. Expón un `NewHandler(...)` y un `Routes(mux)`.
3. Si tiene URLs públicas, implementa `site.SitemapSource` y pásalo a
   `site.NewHandler(baseURL, productosH, <dominioH>)` en `main.go`.
