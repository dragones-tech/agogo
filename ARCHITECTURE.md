# Agogo Architecture

This document explains **how** Agogo is conceived and **why**. For usage
instructions (running, routes, env vars) see the [README](README.md).

> In one sentence: **minimalist Go, domain-oriented, and you own the code.**
> Iterative design driven by honest trade-offs, stdlib-first, tuned until the
> tool stops getting in your way.

---

## Principles

1. **Stdlib first — "only the essentials".**
   The HTTP server, the templates, the JSON, the CSRF and the middlewares are
   **our own code on top of Go's standard library**. There is no
   web framework. The only unavoidable dependency is the SQLite driver
   (`modernc.org/sqlite`, pure Go, to keep the binary static and the `scratch`
   image).

2. **"No magic" = no external black boxes.**
   The rejection is not of abstraction, but of opaque third-party dependencies
   that hide behavior at runtime (e.g. an ORM). **Our own** abstractions
   —generics, helpers, `Resource[T]`— are worth it: they live in the repo, you
   read them and you fix them yourself. `sqlc` generates *our* code.

3. **Package-by-feature.**
   Each domain is self-sufficient: its `handler.go`, its `sql/`, its
   `templates/`, and (if it has a DB) `migrate.go` and `seed.go`. To understand
   `auth` you open one folder, you don't jump between technical layers.

4. **Hand-written SQL with sqlc, no ORM (Repository pattern).**
   One typed function per query, parameterized (injection impossible by design).
   The data "API" is the domain's vocabulary, not SQL primitives.

5. **YAGNI first, DRY later.**
   Nothing is abstracted in advance. Duplication is extracted when it **pays off**
   (e.g. `web.Resource[T]` showed up when thinking about scale, not with 2
   domains).

6. **A toolbox, not a mega-framework.**
   Small, composable pieces. Each domain uses the one that fits; not everything is
   a `Resource[T]`.

7. **Minimal core + pluggable modules.**
   The server is a host (`app`) onto which opt-in modules are plugged
   (`app.Use(...)`). What you don't plug in doesn't enter the binary. Explicit
   composition, no magic auto-registration; behavior is written, not generated.
   URLs stay decoupled from the domain.

8. **The server only serves.**
   The schema and the sample data are managed **outside the site**, as
   development commands (`cmd/migrate`, `cmd/seed`).

9. **First-class security.**
   `scratch` container, non-root user, security headers, CSRF on forms,
   parameterized SQL, everything typed.

10. **SEO-first + JSON API from a single source of truth.**
    Server-rendered HTML (`html/template`, JSON-LD, sitemap) and JSON come out of
    the same model.

---

## Structure

```
.
├── main.go              # the "Gemfile": creates the App and plugs modules with Use()
├── cmd/
│   ├── migrate/         # applies the schema (bootstrap; used on deploy)
│   └── seed/            # sample data (development only)
└── internal/
    │  ── CORE + utilities (base, not modules) ──
    ├── app/             # HOST: App, Module interface, hook points
    ├── router/          # Express-style router on top of net/http (+ middleware)
    ├── config/          # typed configuration validated from the environment
    ├── session/         # signed-cookie sessions (HMAC), no store
    ├── identity/           # identity service on top of session + Require
    ├── password/        # PBKDF2-HMAC-SHA256 hashing (stdlib)
    ├── web/             # Resource[T]: generic DB resource
    ├── view/            # shared HTML layout + Render
    ├── respond/         # respond.JSON / respond.Error / respond.ServerError
    ├── validate/        # minimal, explicit input validator (no reflection)
    ├── jsonld/          # schema.org builders
    ├── csrf/            # double-cookie CSRF protection
    ├── sitemap/         # sitemap contract (URL, Source, Entries) — leaf
    ├── middleware/      # security baseline: Recover, SecurityHeaders
    │  ── MODULES (opt-in; each one with module.go → Module()) ──
    ├── logs/            # global access-logging middleware (active)
    ├── home/            # "hola mundo" landing (active)
    ├── otw/             # BFF: renders a token-gated API as an HTML fragment (active)
    ├── paginas/         # static page example + the otw demo (no DB; opt-in breadcrumb)
    ├── auth/            # username/password login, reuses identity (needs the DB; opt-in)
    ├── oauth/           # OAuth 2.0 login, reuses identity, stdlib (no DB; opt-in)
    └── site/            # robots.txt, sitemap.xml, /static, favicon, styled 404
```

**Direction of dependencies:** the **modules** import `app` (the host) and the
base utilities; **`app` imports no module**. The leaves (`router`, `respond`,
`view`, `sitemap`, `csrf`, `config`, `session`, `password`) do not depend on
other internal packages. No cycles.

---

## Core + modules

The server is a **minimal core (`app`) onto which modules are plugged**. A module
implements `Module{ Name() string; Register(*App) error }` and in its `Register`
it hooks into the host: it adds routes (`a.Router.Get`), global middleware
(`a.UseMiddleware`), migrations (`a.AddMigration`), sitemap sources
(`a.AddSitemap`) and uses shared services (`a.DB`, `a.Session`, `a.Identity`).

`main.go` is the "Gemfile": `app.Use(logs.Module(), home.Module(), otw.Module(),
site.Module())`, with `auth`/`oauth`/`paginas` sitting one commented line away.
**What you don't plug in is not imported, so it doesn't enter the binary** (clean
if unused). Each domain's routes live in its module; the global picture is the
list of `app.Use(...)`. Third-party plugin distribution = Go modules (`go get`) +
implementing `Module`; no registry of our own.

---

## The three domain shapes

Not everything fits the same mold. There are three shapes; each module registers
its handlers in its `Register`:

1. **DB resource — `web.Resource[T]`** (the toolbox for list/detail resources).
   Generic, written once: list/detail × HTML/JSON. A domain only contributes
   behavior (sqlc queries, templates, SEO metadata, JSON-LD) and gets
   `ListHTML`, `DetailHTML`, `ListJSON`, `DetailJSON`. The starter ships the
   generic in `internal/web`; a `productos`/`blog`-style module is the canonical
   thing you'd build on top of it.

2. **Static page** (home, paginas).
   No DB: template + SEO. It exposes an `http.HandlerFunc` (e.g.
   `paginas.Example(baseURL)`).

3. **Form / custom handler** (auth, otw).
   Unique logic (validation, CSRF, PRG, or a BFF call). It exposes its own
   handlers (`auth.Login`, `otw.panel`, …).

---

## Routing

Each **module registers its routes** in its `Register(a *app.App)`, using the
router (Express-style, with per-route middleware):

```go
// a productos-style DB resource you'd add (uses internal/web.Resource[T]):
func (mod) Register(a *app.App) error {
    res := New(db.New(a.DB), a.Config.BaseURL)
    r := a.Router
    r.Get("/productos", res.ListHTML)
    r.Get("/productos/{slug}", res.DetailHTML)
    r.Get("/api/productos", res.ListJSON)
    a.AddSitemap(res.SitemapSource("/productos", "/productos"))
    a.AddMigration(Migrate)
    return nil
}
// internal/auth/module.go (shipped) — identity as per-route middleware:
r.Get("/cuenta", h.Cuenta, a.Identity.Require)
```

The **global picture** of what the server does is the `app.Use(...)` list in
`main.go`; the detail of each domain lives in its `module.go`. (There used to be
a central, flat `routes.go`; it was swapped for the modular model when the
project became a framework — the trade was losing that single table.)

URLs stay **decoupled**: each handler computes its `canonical` from the request
(`baseURL + r.URL.Path`), so the same handler serves on any route where the
module mounts it.

Two levels of middleware: **global** (modules via `a.UseMiddleware`, e.g. `logs`;
plus the core's security baseline) and **per-route** (`r.Get(path, fn, mw...)`,
e.g. `a.Identity.Require`).

---

## Data layer

- **SQLite** via `modernc.org/sqlite` (pure Go, no cgo → static binary). It opens
  in **WAL** mode with **`busy_timeout`** (`Config.DSN()`): concurrent reads with
  one writer, and writes wait instead of failing with `database is locked`.
- **sqlc** generates, per domain, a typed function per query from
  `internal/<dominio>/sql/queries.sql` and `schema.sql`. The generated code lives
  in `internal/<dominio>/db/` (not edited).
- Queries are explicit and parameterized. A write (`:exec`) and a read
  (`:one`/`:many`) get the same treatment.

Flow to add a query: edit `queries.sql` → `sqlc generate` → use the generated
function.

---

## Development vs. site

The server binary does **not** create the schema or seed data. That lives apart:

- `go run ./cmd/migrate` — applies the schema (bootstrap; also on deploy).
- `go run ./cmd/seed` — sample data (development only; idempotent).

Both commands **assemble the same `App` as the server** (`app.Use(...)`) and run
`app.Migrate(ctx)`, which runs the migrations each module registered with
`a.AddMigration`. Adding a domain with a DB = plugging it into the list; its
schema is applied on its own, without touching the `cmd/*`.

This way the "site" contains no development logic, and the linker drops the seed
code from the server binary.

---

## Security

- `scratch` image (no OS, no shell), non-root user.
- Core middlewares: `Recover` (a panic doesn't take down the server), `LimitBody`
  (cap of 1 MiB per request, before any `ParseForm`/decode), `SecurityHeaders`
  (`Content-Security-Policy: default-src 'self'` plus `cdn.jsdelivr.net` allowed
  on `style-src`/`script-src` for scc/lumen, and a `sha256` that whitelists the
  inline lumen import map without `'unsafe-inline'`; `X-Content-Type-Options`,
  `X-Frame-Options`, `Referrer-Policy`) and `Gzip` (compresses text responses
  —HTML/CSS/JS/JSON— if the client accepts it; stdlib's `compress/gzip`, decides
  by `Content-Type`, respects `Range`).
- `Secure` cookies: inferred from the scheme of `AGOGO_BASE_URL` (`https://` →
  `Secure`) and can be forced with `AGOGO_SECURE_COOKIES`. Applies to the session
  cookie and the CSRF one.
- Double-cookie CSRF on forms (`internal/csrf`).
- Authenticated pages (`/cuenta`, `/oauth/me`) respond with `Cache-Control: no-store`.
- Parameterized SQL (sqlc) → injection impossible by design.

---

## Errors and operations

- **Errors that you can see.** A server failure is **logged with method and
  route** (`respond.ServerError` for JSON, `view.ServerError` for HTML) but the
  client only gets a generic 500 — internal details are never leaked. Previously
  the real error was discarded; now it stays in the log.
- **Error pages with the site.** 404 and 500 are rendered with the layout
  (`view.NotFound` / `view.ServerError`, template `error.html`), not as bare
  text. A catch-all (`/`) in the `site` module gives the styled 404 to any
  unregistered route.
- **Graceful shutdown.** `main` listens for `SIGINT`/`SIGTERM`, stops accepting,
  gives in-flight requests up to 10s (`srv.Shutdown`) and only then returns, so
  the `defer`s run (including closing the DB). Stdlib (`os/signal`).

---

## Frontend: same philosophy, another layer

The frontend is free (agogo only serves HTML), but out of the box it leans on two
sibling projects with **the same principle** as agogo —no build, no magic, native
primitives, everything readable and auditable—, one for each layer of the web:

- **[scc](https://github.com/dragones-tech/scc)** (CSS): styling semantic HTML
  with native primitives (`@layer`, `@scope`, custom properties). The golden rule:
  a class says *what it is* (`card`, `field`, `alert`), never *how it looks*.
  Presentation lives in the CSS, the HTML stays readable.
- **[lumen](https://github.com/dragones-tech/lumen)** (JS): vanilla-JS UI with no
  build, with native `<template>` and ES modules (the browser only downloads what
  you import).

They are used as **progressive enhancement**: the server renders and validates
the same without JS, and lumen only adds convenience on top.

### How they're loaded: jsDelivr CDN (with a vendor escape hatch)

scc and lumen **are worked on in their repos** (`dragones-tech/scc`,
`dragones-tech/lumen`), not here, and they **change often** — so agogo loads
them from the **jsDelivr CDN** instead of vendoring a copy that would drift:

- **scc** via a `<link>` to `…/scc@main/scc.css` in `view/base.html`.
- **lumen** via an inline **import map** in `base.html` that maps the bare
  `lumen` / `lumen/` specifiers to `…/lumen@main/src/`, so any page script can
  `import { … } from 'lumen'` out of the box (the browser only fetches what it
  imports). The import map is inline, so the CSP whitelists it with a `sha256`
  hash (no `'unsafe-inline'`); recompute the hash in `internal/middleware` if you
  edit the block.

The trade is that the CSP must allow that CDN for `style-src`/`script-src`. To go
back to a strict `'self'`, **vendor** the assets: `scripts/vendor-frontend.sh`
downloads scc + lumen into `internal/site/static/{scc,lumen}` (git-ignored,
re-fetchable like `npm install`); then point the `base.html` URLs at
`/static/scc/scc.css` and `/static/lumen/src/`, and drop the CDN (and the inline
import map's hash) from the CSP.

What is specific to agogo lives in `internal/site/static/style.css` (not in the
CDN'd vendor):

- **The brand** (green): overrides scc's "knobs" (`--accent`, `--neutral-hue`,
  `--radius`) from `:root`. Being outside any `@layer`, it beats scc's defaults
  without touching them.
- **The page frame** (centered, footer at the bottom, card grid).
- **`[hidden] { display: none }`**: scc does not re-assert `[hidden]` and a
  component with its own `display` (e.g. `.card`) would override it. *(Candidate
  to contribute to scc's reset.)*
- **`@view-transition { navigation: auto }`**: native crossfade between pages
  (removes the flicker of the MPA reload), with an opaque background on `<html>`.

To clone and run: `git clone` then `go run .` — nothing else to install.

### Partial navigation (Turbo-style, our own)

A piece **of agogo's own** (not lumen's), server-side + a small JS: when changing
section the document is not reloaded, only `<main>` is replaced. Header, footer,
CSS and the script itself stay fixed → no flicker, no re-parsing assets.

It takes advantage of the fact that the **`content` block is already a partial**:

- **Server** (`view.Render`): if the request carries the `X-Fragment` header, it
  renders only the `fragment` block (title + `content` + `scripts`), not the
  whole page. It responds with `Vary: X-Fragment`.
- **Client** (`static/js/nav.js`): it intercepts clicks on internal links,
  requests the fragment, and replaces `<main>` (with `document.startViewTransition`
  if supported, for a crossfade on the same page). It re-runs the page's
  `<script>`s with a unique URL (`?v=`) to rewire the new `<main>`; their imports
  stay cached. It updates the title and history (`pushState`/`popstate`).

**Degrades cleanly:** without JS, on a network error, a non-HTML response (a
JSON `/api/...`), a direct access, a refresh or a crawler → normal browser
navigation with the full page. SEO and robustness do not depend on the JS; this
is just an experience improvement on top. It does not replace lumen nor does it
stream (it is not full Turbo): it only avoids re-rendering what does not change.

### BFF: HTML over the wire from a token-gated API (`otw` module)

The usual SPA flow is: the browser fetches JSON and a view builds the DOM. When
the data lives behind **a secret you can't expose** (an external API gated by a
token), that flow forces the token into the browser. The `otw` module shows the
alternative — a **backend-for-frontend**. The concrete example fetches a GitHub
repo's stats (the GitHub API, which you call with a PAT for a higher rate limit
or private repos):

1. The browser asks the server for a section (`GET /otw/panel`), triggered by a
   button on the example page (`/ejemplo`, the `paginas` module).
2. The **server** calls the GitHub API with the bearer token (the token lives
   only on the server, never reaches the client), gets JSON, and renders it as
   an **HTML fragment** (`html/template`, so external data is auto-escaped).
3. The client just does `innerHTML` on the section (`static/js/otw.js`, ~15
   lines, no framework, no JSON on the page).

This is "HTML over the wire": rendering and the secret stay on the server; only
HTML crosses. It's the right shape for token-gated/external data — NOT for your
own DB (there you render the fragment directly from the query, no internal HTTP
call). Configured by its own env (`OTW_API_URL`, `OTW_API_TOKEN`); unconfigured,
the BFF points at a built-in `/otw/demo-api` that returns the **same shape**
GitHub does (also token-gated), so the demo runs without credentials — the
decode code is identical for the live and simulated APIs. The BFF absorbs
upstream failures (logs the real error, returns a friendly fragment).

## Configuration and sessions

- **`config`**: reads the environment **once**, typed and validated
  (`config.Load() (Config, error)`). The server and the commands use it. It
  includes `SecretKey` (session signing): if missing, it falls back to an
  insecure development key and flags `DevSecret` (the server warns); if defined,
  it requires ≥32 characters.
- **`session`**: sessions via a **cookie signed with HMAC-SHA256**
  (`crypto/hmac`), with no server-side store and no dependencies — same line as
  `csrf`. The cookie is signed (tamper-proof) but **not encrypted**: do not store
  secrets in it. It carries a **signed expiration** (`_exp` stamp, 7 days): the
  server enforces it even if the client keeps the file, so a stolen cookie is not
  good forever. Current use: the contact form's *flash* message
  (Post-Redirect-Get) and the user id. Limit ~4 KB per cookie.

## Authentication

- **`password`**: **PBKDF2-HMAC-SHA256** hashing with `crypto/pbkdf2` (stdlib
  since Go 1.24), random salt and constant-time comparison. The stored hash
  carries its parameters (`pbkdf2-sha256$iter$salt$hash`), so the algorithm can
  be migrated without breaking existing hashes. (bcrypt/argon2 would be
  preferable but are a dependency; stdlib was chosen.)
- **`identity`** (base): the identity **mechanism** on top of the session. It
  stores `user_id` in the signed cookie; it offers `Login`/`Logout`/`UserID` and
  `Require`, which is a normal **middleware** applied per route (`r.Get("/cuenta",
  fn, a.Identity.Require)`) or per block (`a.Identity.Group(r)`). Always available
  in the core, so any module can protect routes.
- **`auth`** (module): the **functionality** — `usuarios` table, login/logout and
  the protected `/cuenta` page. It verifies credentials (with `password`) and
  delegates to `a.Identity.Login`. It does not reveal whether an email exists;
  login and logout carry CSRF. Another module (`oauth`, `apikeys`) could reuse
  `identity` without rewriting anything.

**Two levels of middleware:** *global* (the `logs` module via `a.UseMiddleware`,
plus the core's security baseline in `app.Handler`) and *per-route* (the router
accepts `mw ...Middleware` after the handler). `identity.Require` is just one more
middleware; tomorrow rate-limit, roles, etc. fit through the same path.

## Working methodology

The cycle we follow: **Discuss → decide → build → verify → document.**

- Options are proposed with **honest trade-offs**; a decision is made; it is
  implemented.
- **Verify by running**: `go build ./...`, `go vet ./...` and `curl` on every
  change. Nothing is taken for granted without seeing it respond.
- **Refactor without fear**: the design evolves (the routing was redone several
  times until landing on the flat, decoupled shape).
- Decisions are captured; the non-urgent ones are documented as pending instead
  of being forced.

---

## References

Sibling projects (same DNA, one layer each):

- **[lumen](https://github.com/dragones-tech/lumen)** — vanilla-JS UI, no build,
  no magic. The client layer.
- **[scc](https://github.com/dragones-tech/scc)** — Scoped Cascade CSS, semantic,
  no build. The style layer.

The design coincides, without having sought it, with the canonical Go practices:

- **Mat Ryer — "How I write HTTP services in Go after 13 years"** (Grafana): the
  `routes.go` pattern as the single place for routes.
- **Ben Johnson — [`benbjohnson/wtf`](https://github.com/benbjohnson/wtf)**:
  organization by domain with interfaces as contracts.
- **[`pocketbase/pocketbase`](https://github.com/pocketbase/pocketbase)**: same
  single-binary + embedded SQLite philosophy.
- **[`walterwanderley/sqlc-http`](https://github.com/walterwanderley/sqlc-http)**:
  same stack (net/http + sqlc), but *generating* the HTTP layer from the SQL —
  the opposite approach to ours (we write it and decouple it by hand).
