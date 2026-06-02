# Agogo

A compact web server in **pure Go**: server-rendered HTML (for SEO) and a JSON
API, from a single source of truth. No frameworks or "all-in-one" libraries —
just the standard library and niche tools.

This repo is a **clean starter**: clone it, run it, and you get a "hello world"
home with a link to the docs. You grow your app by adding sections (modules);
the framework toolbox already lives in `internal/`.

## Quickstart

Requires **Go 1.26+**. No build step, nothing to install (scc/lumen load from a
CDN):

```bash
git clone https://github.com/dragones-tech/agogo
cd agogo
go run .          # → http://localhost:46060
```

You'll see a centered "¡Hola, mundo!" and a link to the documentation.

## Philosophy

- **Only the essentials.** HTTP server, templates and JSON come from Go's stdlib
  (`net/http`, `html/template`, `encoding/json`). Zero web framework.
- **No magic.** No ORM, no reflection-driven config, no runtime black boxes. The
  code that runs is the code you read and edit — it all lives in `internal/`.
- **Minimal core + pluggable modules.** A host (`app`) you plug opt-in modules
  into with `app.Use(...)`. What you don't wire in never enters the binary.
- **YAGNI first, DRY later.** Nothing is abstracted in advance; duplication is
  extracted only when it pays off.

The only dependency is the SQLite driver (`modernc.org/sqlite`, pure Go) so the
binary stays static and runs on a `scratch` image.

## Configuration (environment)

Read once at startup, typed and validated (`internal/config`).

| Variable | Default | Description |
|----------|---------|-------------|
| `AGOGO_ADDR` | `:46060` | Listen address |
| `AGOGO_DB` | `agogo.db` | SQLite file path (opened in WAL + `busy_timeout`) |
| `AGOGO_BASE_URL` | `http://localhost:46060` | Base URL (canonical, sitemap) |
| `AGOGO_SECRET_KEY` | (insecure dev key) | Session signing key (HMAC). In production, ≥32 chars. |
| `AGOGO_SECURE_COOKIES` | (derived from `https://` in base URL) | Force `Secure` cookies on/off |

## Structure

```
.
├── main.go                  # the "Gemfile": creates the App and wires modules with app.Use(...)
├── Dockerfile               # multi-stage → scratch
└── internal/
    │  ── CORE + toolbox (base packages, not modules) ──
    ├── app/                 # HOST: App, Module interface, hook points
    ├── router/              # Express-style router on net/http (+ middleware)
    ├── config/              # typed env config (incl. SQLite DSN: WAL + busy_timeout)
    ├── view/                # shared HTML layout + Render
    ├── middleware/          # baseline: Recover, LimitBody, SecurityHeaders, Gzip
    ├── respond/ validate/   # JSON responses + a tiny input validator
    ├── session/ identity/ password/ csrf/   # auth/session toolbox
    ├── web/ jsonld/ sitemap/                 # generic DB resource, schema.org, sitemap
    ├── logs/                # opt-in access-logging middleware
    │  ── SECTIONS (modules) ──
    ├── home/                # the active "hello world" section
    ├── paginas/             # EXAMPLE static section (wired but commented in main.go)
    └── site/                # robots.txt, sitemap.xml, /static, favicon, styled 404
```

Each module implements `Module{ Name() string; Register(*App) error }` and in
its `Register` hooks into the host: routes (`a.Router.Get`), global middleware
(`a.UseMiddleware`), migrations (`a.AddMigration`), sitemap sources
(`a.AddSitemap`), and shared services (`a.DB`, `a.Session`, `a.Identity`).
`main.go`'s `app.Use(...)` is the overview of what the server does.

## Add a section

A section is a module. The smallest one is `internal/home`:

1. Copy `internal/home` (or `internal/paginas`) to `internal/<your-section>`.
2. In its `Register`, register the routes and adjust the template.
3. Wire it in `main.go`'s `app.Use(...)` with one line.

`main.go` keeps a commented `paginas.Module()` line as a live example (it serves
`/ejemplo` once you uncomment it).

## Data layer (when you add a DB section)

The starter ships no database section, but the toolbox is ready:

- **SQLite** via `modernc.org/sqlite` (pure Go, no cgo → static binary), opened
  in **WAL** with **`busy_timeout`** (`config.DSN()`).
- **sqlc** generates one typed function per query from real SQL (no ORM,
  injection impossible by design): write the query → `sqlc generate` → use it.
- **`web.Resource[T]`** writes the 4 handlers (list/detail × HTML/JSON), route
  registration and sitemap contribution **once**; a domain only configures its
  types, queries, templates and SEO metadata.
- The schema/seed live **outside the server** as small commands (the server only
  serves). Add a `cmd/migrate` that builds the App and runs `app.Migrate(ctx)`.

## Frontend

The frontend is your choice. Out of the box agogo styles its HTML with **scc**
(semantic CSS — classes say *what it is*, not *how it looks*) and can enhance it
with **lumen** (vanilla-JS UI, no build), two sibling projects with the same
no-build, no-magic DNA. They load from the **jsDelivr CDN**, so there's nothing
to vendor or build. agogo's branding lives in `internal/site/static/style.css`
(it overrides scc's "knobs"). The CSP allows that CDN on `style-src`/`script-src`
— pin a tag/commit or self-host to go back to a strict `'self'`.

## Security

- `scratch` image (no OS, no shell), non-root user.
- Core middlewares: `Recover` (a panic doesn't take down the server), `LimitBody`
  (1 MiB cap per request), `SecurityHeaders` (CSP, `X-Content-Type-Options`,
  `X-Frame-Options`, `Referrer-Policy`) and `Gzip`.
- Signed-cookie sessions (HMAC), `Secure` cookies derived from the base URL,
  CSRF (double cookie) and PBKDF2 hashing — all stdlib, in the toolbox.
- Parameterized SQL via sqlc; styled `404`/`500` that never leak internals.
- Graceful shutdown on `SIGINT`/`SIGTERM` (`srv.Shutdown`).

## Deploy & HTTP protocols

Build a static binary and run it on a minimal image:

```bash
docker build -t agogo .
docker run -p 46060:46060 agogo      # → http://localhost:46060
```

The binary speaks **HTTP/1.1** (and **HTTP/2** automatically over TLS, via
`ListenAndServeTLS`). For **HTTP/3** + automatic TLS, terminate at the edge and
keep the binary simple:

```caddyfile
agogo.com {
	encode gzip zstd
	reverse_proxy localhost:46060
}
```

## Documentation

Full design notes, patterns and examples live in the documentation site (linked
from the home page) and in [ARCHITECTURE.md](ARCHITECTURE.md).
