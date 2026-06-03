// @ts-check
// API reference data — generated from `go doc`/go/doc over internal/<pkg>
// (exported symbols + signatures + docs). The per-module `example` is hand-written.
// Regenerate the structured part with the docs generator; English only.
export const API = [
 {
  "pkg": "app",
  "doc": "Package app is the framework CORE: the host that composes modules. A server\nis assembled by creating an App and plugging modules into it with Use().\nWhatever you don't plug in isn't imported, so it doesn't exist in the binary.\n\nThe host offers cheap shared services (Session, Auth) that any module can\nuse; the visible features (login, content) are modules.\n",
  "symbols": [
   {
    "name": "App",
    "kind": "type",
    "sig": "type App struct {\n\tConfig\t\tconfig.Config\n\tDB\t\t*sql.DB\n\tRouter\t\t*router.Router\n\tSession\t\t*session.Manager\n\tIdentity\t*identity.Service\n\n\tmiddleware\t[]Middleware\n\tmigrations\t[]func(context.Context, *sql.DB) error\n\tsitemaps\t[]sitemap.Source\n}",
    "doc": "App is the host: shared services + hook points for modules.\n",
    "methods": [
     {
      "name": "New",
      "kind": "func",
      "sig": "func New(cfg config.Config, db *sql.DB) *App",
      "doc": ""
     },
     {
      "name": "AddMigration",
      "kind": "func",
      "sig": "func (a *App) AddMigration(fn func(context.Context, *sql.DB) error)",
      "doc": ""
     },
     {
      "name": "AddSitemap",
      "kind": "func",
      "sig": "func (a *App) AddSitemap(src sitemap.Source)",
      "doc": ""
     },
     {
      "name": "Handler",
      "kind": "func",
      "sig": "func (a *App) Handler() http.Handler",
      "doc": "Handler assembles the final http.Handler: routes + security baseline (core,\nalways) + the modules' global middleware (e.g. logs) as the outer layer.\n"
     },
     {
      "name": "Migrate",
      "kind": "func",
      "sig": "func (a *App) Migrate(ctx context.Context) error",
      "doc": "Migrate runs the migrations of all plugged-in modules.\n"
     },
     {
      "name": "SitemapSources",
      "kind": "func",
      "sig": "func (a *App) SitemapSources() []sitemap.Source",
      "doc": ""
     },
     {
      "name": "Use",
      "kind": "func",
      "sig": "func (a *App) Use(mods ...Module) error",
      "doc": "Use plugs in modules in order. If one fails to register, it stops with its name.\n"
     },
     {
      "name": "UseMiddleware",
      "kind": "func",
      "sig": "func (a *App) UseMiddleware(mw Middleware)",
      "doc": ""
     }
    ]
   },
   {
    "name": "Middleware",
    "kind": "type",
    "sig": "type Middleware = func(http.Handler) http.Handler",
    "doc": "Middleware is global: it wraps the whole server (unlike the per-route kind).\n"
   },
   {
    "name": "Module",
    "kind": "type",
    "sig": "type Module interface {\n\tName() string\n\tRegister(*App) error\n}",
    "doc": "Module is what a domain or plugin implements to plug into the server.\n"
   }
  ],
  "example": "cfg, _ := config.Load()\ndb, _ := sql.Open(\"sqlite\", cfg.DSN())\na := app.New(cfg, db)\na.Use(home.Module(), logs.Module(), site.Module())\nhttp.ListenAndServe(cfg.Addr, a.Handler())"
 },
 {
  "pkg": "router",
  "doc": "Package router wraps http.ServeMux with an expressive Express-style API:\n\n\tr.Get(\"/productos/{slug}\", fn)\n\tr.Post(\"/contacto\", fn)\n\tr.Get(\"/cuenta\", fn, authsvc.Require)   // with per-route middleware\n\nIt's our own code on top of the stdlib: each method prepends the HTTP verb to\nthe ServeMux pattern and applies the given middleware (the first being the\noutermost). The handler follows Go's idiomatic signature: func(w, r).\n",
  "symbols": [
   {
    "name": "Middleware",
    "kind": "type",
    "sig": "type Middleware = func(http.HandlerFunc) http.HandlerFunc",
    "doc": "Middleware wraps a handler to add behavior (auth, etc.).\n"
   },
   {
    "name": "Router",
    "kind": "type",
    "sig": "type Router struct {\n\tmux *http.ServeMux\n}",
    "doc": "",
    "methods": [
     {
      "name": "New",
      "kind": "func",
      "sig": "func New() *Router",
      "doc": ""
     },
     {
      "name": "Delete",
      "kind": "func",
      "sig": "func (r *Router) Delete(path string, fn http.HandlerFunc, mw ...Middleware)",
      "doc": ""
     },
     {
      "name": "Get",
      "kind": "func",
      "sig": "func (r *Router) Get(path string, fn http.HandlerFunc, mw ...Middleware)",
      "doc": ""
     },
     {
      "name": "Handle",
      "kind": "func",
      "sig": "func (r *Router) Handle(pattern string, h http.Handler)",
      "doc": "Handle mounts an http.Handler on a full pattern (verb included),\ne.g. a file server at \"GET /static/\".\n"
     },
     {
      "name": "Handler",
      "kind": "func",
      "sig": "func (r *Router) Handler() http.Handler",
      "doc": "Handler returns the http.Handler to pass to the http.Server.\n"
     },
     {
      "name": "Post",
      "kind": "func",
      "sig": "func (r *Router) Post(path string, fn http.HandlerFunc, mw ...Middleware)",
      "doc": ""
     },
     {
      "name": "Put",
      "kind": "func",
      "sig": "func (r *Router) Put(path string, fn http.HandlerFunc, mw ...Middleware)",
      "doc": ""
     }
    ]
   }
  ],
  "example": "r := a.Router\nr.Get(\"/productos/{slug}\", h.Detail)\nr.Post(\"/login\", h.Login)\nr.Get(\"/cuenta\", h.Account, a.Identity.Require) // per-route middleware"
 },
 {
  "pkg": "config",
  "doc": "Package config reads the configuration from the environment once, typed and\nvalidated. Both the server and the commands (cmd/*) use it.\n",
  "symbols": [
   {
    "name": "Config",
    "kind": "type",
    "sig": "type Config struct {\n\tDB\t\tstring\t// path to the SQLite file\n\tBaseURL\t\tstring\t// base URL (canonical, sitemap)\n\tAddr\t\tstring\t// server listen address\n\tSecretKey\t[]byte\t// key for signing sessions (HMAC)\n\tDevSecret\tbool\t// true if it fell back to the (insecure) development key\n\tSecure\t\tbool\t// marks cookies as Secure (only sent over HTTPS)\n}",
    "doc": "",
    "methods": [
     {
      "name": "Load",
      "kind": "func",
      "sig": "func Load() (Config, error)",
      "doc": "Load builds the configuration from the environment. It returns an error on\nhard validations (e.g. a key that's too short).\n"
     },
     {
      "name": "DSN",
      "kind": "func",
      "sig": "func (c Config) DSN() string",
      "doc": "DSN is the SQLite connection string: the path + PRAGMAs that the modernc\ndriver applies when opening each connection. WAL allows concurrent reads with\none write; busy_timeout makes a write wait (up to 5s) instead of failing\ninstantly with \"database is locked\"; foreign_keys enables FKs.\n"
     }
    ]
   }
  ],
  "example": "cfg, err := config.Load()        // reads AGOGO_* once\nif err != nil { log.Fatal(err) }\ndb, _ := sql.Open(\"sqlite\", cfg.DSN()) // WAL + busy_timeout"
 },
 {
  "pkg": "view",
  "doc": "Package view provides the HTML layout shared by all domains. Here lives, once,\nthe <head> with SEO, the common navigation and footer; each domain contributes\nonly its \"content\" block.\n",
  "symbols": [
   {
    "name": "Layout",
    "kind": "func",
    "sig": "func Layout(contentFS fs.FS, files ...string) *template.Template",
    "doc": "Layout composes the base layout with the domain's content templates.\ncontentFS is the domain's embed.FS; files are paths within it\n(each must define the \"content\" block).\n"
   },
   {
    "name": "NotFound",
    "kind": "func",
    "sig": "func NotFound(w http.ResponseWriter, _ *http.Request)",
    "doc": "NotFound responds with a 404 using the site layout. For HTML handlers.\n"
   },
   {
    "name": "Render",
    "kind": "func",
    "sig": "func Render(w http.ResponseWriter, r *http.Request, t *template.Template, data any)",
    "doc": "Render executes the template with the page data. If the request carries\nX-Fragment (partial site navigation), it renders only the \"fragment\" block\n(title + content + scripts); otherwise the full page (\"base\"). This way a soft\nnavigation reuses header/footer/CSS and only swaps <main>, but a direct hit, a\ncrawler or no-JS receives the whole document (SEO intact).\n\nIt renders to a buffer first: if the template fails midway, we respond a clean\n500 instead of a 200 with truncated HTML already sent to the client.\n"
   },
   {
    "name": "ServerError",
    "kind": "func",
    "sig": "func ServerError(w http.ResponseWriter, r *http.Request, err error)",
    "doc": "ServerError logs the REAL error (with method and path, for debugging) and\nresponds with a 500 using the site layout, without leaking internal details\nto the client.\n"
   },
   {
    "name": "Meta",
    "kind": "type",
    "sig": "type Meta struct {\n\tTitle\t\tstring\n\tDescription\tstring\n\tCanonical\tstring\n\tOGType\t\tstring\t// \"website\", \"article\", ...\n\tJSONLD\t\ttemplate.JS\n}",
    "doc": "Meta is the header (SEO) data that each page fills in.\n"
   }
  ],
  "example": "//go:embed templates/*.html\nvar tplFS embed.FS\nvar tpl = view.Layout(tplFS, \"templates/page.html\") // base + content\n\nfunc handler(w http.ResponseWriter, r *http.Request) {\n    view.Render(w, r, tpl, struct{ Meta view.Meta }{view.Meta{Title: \"Page\"}})\n}"
 },
 {
  "pkg": "middleware",
  "doc": "Package middleware gathers the CORE security baseline (always active): panic\nrecovery and security headers. (Access logging lives in the opt-in module\ninternal/logs.)\n",
  "symbols": [
   {
    "name": "Gzip",
    "kind": "func",
    "sig": "func Gzip(next http.Handler) http.Handler",
    "doc": "Gzip compresses the response with gzip when the client accepts it and the\ncontent is text (HTML/CSS/JS/JSON/XML). It decides when headers are set, based\non the Content-Type the handler put, so it doesn't recompress images. Our own\ncode on top of the stdlib (compress/gzip). In production this is usually\nterminated at the edge (Caddy: encode gzip zstd); this middleware serves the\nbare binary.\n"
   },
   {
    "name": "LimitBody",
    "kind": "func",
    "sig": "func LimitBody(next http.Handler) http.Handler",
    "doc": "LimitBody caps the body of each request before any handler reads it\n(ParseForm, json.Decode...), avoiding exhausting memory with a huge POST.\n"
   },
   {
    "name": "Recover",
    "kind": "func",
    "sig": "func Recover(next http.Handler) http.Handler",
    "doc": "Recover catches any panic in a handler and responds with a 500 without\nbringing down the whole server.\n"
   },
   {
    "name": "SecurityHeaders",
    "kind": "func",
    "sig": "func SecurityHeaders(next http.Handler) http.Handler",
    "doc": "SecurityHeaders adds hardening headers to all responses.\n"
   }
  ],
  "example": "// Applied automatically by app.Handler(): Recover, LimitBody,\n// SecurityHeaders, Gzip. Add your own globally from a module:\nfunc (mod) Register(a *app.App) error {\n    a.UseMiddleware(middleware.SecurityHeaders) // example\n    return nil\n}"
 },
 {
  "pkg": "respond",
  "doc": "Package respond writes common HTTP responses, shared by the domains. They're\nflat functions: what you see is what happens, with no state or reflection.\n",
  "symbols": [
   {
    "name": "Error",
    "kind": "func",
    "sig": "func Error(w http.ResponseWriter, status int, msg string)",
    "doc": "Error writes a {\"error\": msg} body with the given status.\n"
   },
   {
    "name": "JSON",
    "kind": "func",
    "sig": "func JSON(w http.ResponseWriter, status int, v any)",
    "doc": "JSON writes v as JSON with the given status.\n"
   },
   {
    "name": "ServerError",
    "kind": "func",
    "sig": "func ServerError(w http.ResponseWriter, r *http.Request, err error)",
    "doc": "ServerError logs the REAL error (with method and path, for debugging) and\nresponds with a generic 500 to the client, without leaking internal details.\nUse it in JSON handlers when an operation fails due to the server.\n"
   }
  ],
  "example": "func listJSON(w http.ResponseWriter, r *http.Request) {\n    items, err := q.List(r.Context())\n    if err != nil { respond.ServerError(w, r, err); return } // logs + generic 500\n    respond.JSON(w, http.StatusOK, items)\n}"
 },
 {
  "pkg": "validate",
  "doc": "Package validate is a tiny, explicit input validator for handlers/forms: you\naccumulate per-field error messages as you check, with NO reflection and NO\nstruct tags — nothing hidden, you read the checks. Each domain composes only\nthe rules it needs. It works on strings (form values); parse numbers/dates in\nthe handler and validate the result.\n\nIt's a leaf utility (like respond or password), not a Module: handlers import\nand call it; it registers nothing on the App.\n",
  "symbols": [
   {
    "name": "Errors",
    "kind": "type",
    "sig": "type Errors map[string]string",
    "doc": "Errors maps a field name to its error message. Empty means valid. Its\nunderlying type is map[string]string, so it drops straight into a template's\npage data (e.g. {{.Errors.email}}).\n",
    "methods": [
     {
      "name": "New",
      "kind": "func",
      "sig": "func New() Errors",
      "doc": "New returns an empty error set ready to accumulate.\n"
     },
     {
      "name": "Add",
      "kind": "func",
      "sig": "func (e Errors) Add(field, msg string)",
      "doc": "Add records msg for field, keeping the FIRST message set for that field, so\nthe earliest (most important) rule wins and later checks don't overwrite it.\n"
     },
     {
      "name": "Email",
      "kind": "func",
      "sig": "func (e Errors) Email(field, value, msg string)",
      "doc": "Email fails if value is not a parseable email address (net/mail, stdlib).\n"
     },
     {
      "name": "MaxLen",
      "kind": "func",
      "sig": "func (e Errors) MaxLen(field, value string, n int, msg string)",
      "doc": "MaxLen fails if value has more than n characters (runes).\n"
     },
     {
      "name": "MinLen",
      "kind": "func",
      "sig": "func (e Errors) MinLen(field, value string, n int, msg string)",
      "doc": "MinLen fails if the trimmed value has fewer than n characters (runes, so\naccents and emoji count as one).\n"
     },
     {
      "name": "OK",
      "kind": "func",
      "sig": "func (e Errors) OK() bool",
      "doc": "OK reports whether every field passed.\n"
     },
     {
      "name": "Required",
      "kind": "func",
      "sig": "func (e Errors) Required(field, value, msg string)",
      "doc": "Required fails if value is empty after trimming whitespace.\n"
     }
    ]
   }
  ],
  "example": "e := validate.New()\ne.Required(\"name\", name, \"El nombre es obligatorio.\")\ne.Email(\"email\", email, \"Email inválido.\")\ne.MinLen(\"msg\", msg, 10, \"Muy corto.\")\nif !e.OK() { /* render e (map field→message) */ }"
 },
 {
  "pkg": "session",
  "doc": "Package session implements sessions via a SIGNED COOKIE (HMAC-SHA256), with\nno server-side store and no dependencies. Our own code on top of the stdlib,\nin the same vein as internal/csrf.\n\nThe cookie is SIGNED (tamper-proof) but NOT encrypted: the data is readable by\nthe client. Don't store secrets in the session; do store things like the user\nid or a flash message. Practical limit: ~4 KB per cookie.\n",
  "symbols": [
   {
    "name": "Manager",
    "kind": "type",
    "sig": "type Manager struct {\n\tkey\t[]byte\n\tSecure\tbool\t// set to true when serving behind TLS\n}",
    "doc": "Manager signs and verifies sessions with a secret key.\n",
    "methods": [
     {
      "name": "NewManager",
      "kind": "func",
      "sig": "func NewManager(key []byte) *Manager",
      "doc": ""
     },
     {
      "name": "Get",
      "kind": "func",
      "sig": "func (m *Manager) Get(r *http.Request) *Session",
      "doc": "Get reads the session from the cookie (empty if absent, the signature fails to\nvalidate, or it expired).\n"
     },
     {
      "name": "Save",
      "kind": "func",
      "sig": "func (m *Manager) Save(w http.ResponseWriter, s *Session)",
      "doc": "Save writes the signed session into a cookie, stamping it with its expiration.\n"
     }
    ]
   },
   {
    "name": "Session",
    "kind": "type",
    "sig": "type Session struct {\n\tValues map[string]string\n}",
    "doc": "Session holds the values of the current session.\n",
    "methods": [
     {
      "name": "Delete",
      "kind": "func",
      "sig": "func (s *Session) Delete(key string)",
      "doc": ""
     },
     {
      "name": "Get",
      "kind": "func",
      "sig": "func (s *Session) Get(key string) string",
      "doc": ""
     },
     {
      "name": "Set",
      "kind": "func",
      "sig": "func (s *Session) Set(key, val string)",
      "doc": ""
     }
    ]
   }
  ],
  "example": "s := mgr.Get(r)        // read (empty if none/expired)\ns.Set(\"flash\", \"¡Guardado!\")\nmgr.Save(w, s)        // signed cookie"
 },
 {
  "pkg": "identity",
  "doc": "Package identity provides identity on top of the session: log in/out, knowing\nwho the current user is, and a middleware to protect routes. It doesn't know\nthe DB; it only stores the user id in the signed session.\n",
  "symbols": [
   {
    "name": "Guard",
    "kind": "type",
    "sig": "type Guard struct {\n\tr\t*router.Router\n\ts\t*Service\n}",
    "doc": "Guard registers routes that ALWAYS require a session: it applies Require for\nyou, so you don't have to remember to wrap each handler (and you don't\naccidentally leave a route public). Obtain one with Service.Group(r).\n",
    "methods": [
     {
      "name": "Delete",
      "kind": "func",
      "sig": "func (g *Guard) Delete(path string, fn http.HandlerFunc)",
      "doc": ""
     },
     {
      "name": "Get",
      "kind": "func",
      "sig": "func (g *Guard) Get(path string, fn http.HandlerFunc)",
      "doc": ""
     },
     {
      "name": "Post",
      "kind": "func",
      "sig": "func (g *Guard) Post(path string, fn http.HandlerFunc)",
      "doc": ""
     },
     {
      "name": "Put",
      "kind": "func",
      "sig": "func (g *Guard) Put(path string, fn http.HandlerFunc)",
      "doc": ""
     }
    ]
   },
   {
    "name": "Service",
    "kind": "type",
    "sig": "type Service struct {\n\tsess *session.Manager\n}",
    "doc": "",
    "methods": [
     {
      "name": "New",
      "kind": "func",
      "sig": "func New(sess *session.Manager) *Service",
      "doc": ""
     },
     {
      "name": "Group",
      "kind": "func",
      "sig": "func (s *Service) Group(r *router.Router) *Guard",
      "doc": "Group returns a Guard over the router: everything you register with it ends up\nprotected. Example:\n\n\tprotected := authsvc.Group(r)\n\tprotected.Get(\"/cuenta\", h.Cuenta)\n"
     },
     {
      "name": "Login",
      "kind": "func",
      "sig": "func (s *Service) Login(w http.ResponseWriter, r *http.Request, userID string)",
      "doc": "Login marks the user as authenticated by storing their id in the session.\n"
     },
     {
      "name": "LoginInto",
      "kind": "func",
      "sig": "func (s *Service) LoginInto(w http.ResponseWriter, sess *session.Session, userID string)",
      "doc": "LoginInto marks the user on an ALREADY loaded session and saves it in a single\nSet-Cookie. Use it when the caller made other mutations that must persist in\nthe same save (e.g. oauth, which consumes its anti-CSRF state before logging\nin); two separate Saves would emit two cookies with the same name and the\nclient would keep only one.\n"
     },
     {
      "name": "Logout",
      "kind": "func",
      "sig": "func (s *Service) Logout(w http.ResponseWriter, r *http.Request)",
      "doc": "Logout clears the identity from the session.\n"
     },
     {
      "name": "Require",
      "kind": "func",
      "sig": "func (s *Service) Require(next http.HandlerFunc) http.HandlerFunc",
      "doc": "Require protects a route: if there's no session, it redirects to /login.\n"
     },
     {
      "name": "UserID",
      "kind": "func",
      "sig": "func (s *Service) UserID(r *http.Request) string",
      "doc": "UserID returns the current user's id (\"\" if no session is started).\n"
     }
    ]
   }
  ],
  "example": "// log in after verifying credentials:\na.Identity.Login(w, r, strconv.FormatInt(userID, 10))\n// protect a route:\nr.Get(\"/cuenta\", h.Account, a.Identity.Require)"
 },
 {
  "pkg": "password",
  "doc": "Package password hashes and verifies passwords with PBKDF2-HMAC-SHA256, all\nwith the stdlib (crypto/pbkdf2, available since Go 1.24). No dependencies.\n\nHonest note: bcrypt/argon2 (golang.org/x/crypto) are preferred for their\nmemory hardness; PBKDF2 with many iterations is approved by NIST/OWASP and\nkeeps us in the stdlib. The stored format carries the parameters, so the\nalgorithm can be migrated without breaking old hashes.\n",
  "symbols": [
   {
    "name": "Hash",
    "kind": "func",
    "sig": "func Hash(plain string) (string, error)",
    "doc": "Hash returns \"pbkdf2-sha256$iter$salt$hash\" (salt and hash in base64).\n"
   },
   {
    "name": "Match",
    "kind": "func",
    "sig": "func Match(plain, encoded string) bool",
    "doc": "Match verifies a password against an encoded hash, in constant time.\n"
   }
  ],
  "example": "hash, _ := password.Hash(\"correct horse\")   // store this\nok := password.Match(\"correct horse\", hash)  // true"
 },
 {
  "pkg": "csrf",
  "doc": "Package csrf implements CSRF protection via the \"double submit cookie\"\npattern: the server issues a token, stores it in a cookie and embeds it in a\nhidden form field; on receiving the POST it compares the two. Our own code,\nstdlib, no dependencies.\n",
  "symbols": [
   {
    "name": "Issue",
    "kind": "func",
    "sig": "func Issue(w http.ResponseWriter, secure bool) string",
    "doc": "Issue generates a token, stores it in a cookie and returns it for the <input\nhidden>. secure marks the cookie as Secure (only sent over HTTPS); pass it\nfrom the configuration (Config.Secure).\n"
   },
   {
    "name": "Valid",
    "kind": "func",
    "sig": "func Valid(r *http.Request) bool",
    "doc": "Valid compares the cookie's token with the form's (constant-time comparison).\nIt fails if either is missing or they don't match.\n"
   }
  ],
  "example": "// GET: issue a token for the form's hidden field\ntoken := csrf.Issue(w, cfg.Secure)\n// POST: reject if it doesn't match the cookie\nif !csrf.Valid(r) { http.Error(w, \"CSRF\", 403); return }"
 },
 {
  "pkg": "web",
  "doc": "Package web contains Resource[T], the generic DB resource (list/detail ×\nHTML/JSON), written once. It does NOT know its URL: its handlers are wired up\nin the domain's module.go (r.Get(path, recurso.ListHTML), ...). The canonical\nand the JSON-LD are computed from the request URL, so the same handler serves\nat any route where you mount it.\n",
  "symbols": [
   {
    "name": "ItemPage",
    "kind": "type",
    "sig": "type ItemPage[T any] struct {\n\tMeta\tview.Meta\n\tItem\tT\n}",
    "doc": ""
   },
   {
    "name": "Page",
    "kind": "type",
    "sig": "type Page[T any] struct {\n\tMeta\tview.Meta\n\tItems\t[]T\n}",
    "doc": "Page and ItemPage are the data that travel to the templates.\n"
   },
   {
    "name": "Resource",
    "kind": "type",
    "sig": "type Resource[T any] struct {\n\tBaseURL\tstring\n\n\tList\tfunc(context.Context) ([]T, error)\n\tGet\tfunc(context.Context, string) (T, error)\n\tSlug\tfunc(T) string\n\n\tTplList, TplItem\t*template.Template\n\tListMeta\t\tfunc(url string) view.Meta\n\tItemMeta\t\tfunc(it T, url string) view.Meta\n\n\tSitemapFreq\tstring\n\tSitemapPrio\tstring\n}",
    "doc": "Resource describes the BEHAVIOR of a resource (queries, templates, SEO).\n",
    "methods": [
     {
      "name": "DetailHTML",
      "kind": "func",
      "sig": "func (r Resource[T]) DetailHTML(w http.ResponseWriter, req *http.Request)",
      "doc": ""
     },
     {
      "name": "DetailJSON",
      "kind": "func",
      "sig": "func (r Resource[T]) DetailJSON(w http.ResponseWriter, req *http.Request)",
      "doc": ""
     },
     {
      "name": "ListHTML",
      "kind": "func",
      "sig": "func (r Resource[T]) ListHTML(w http.ResponseWriter, req *http.Request)",
      "doc": ""
     },
     {
      "name": "ListJSON",
      "kind": "func",
      "sig": "func (r Resource[T]) ListJSON(w http.ResponseWriter, req *http.Request)",
      "doc": ""
     },
     {
      "name": "SitemapSource",
      "kind": "func",
      "sig": "func (r Resource[T]) SitemapSource(listPath, itemBase string) sitemap.Source",
      "doc": "SitemapSource creates the resource's sitemap source: the list entry\n(listPath) plus one per item under itemBase. You pass the paths in module.go.\n"
     }
    ]
   }
  ],
  "example": "res := web.Resource[db.Product]{\n    BaseURL: baseURL,\n    List:    q.ListProducts,\n    Get:     q.GetProduct,\n    Slug:    func(p db.Product) string { return p.Slug },\n    TplList: tplList, TplItem: tplItem,\n    ListMeta: listMeta, ItemMeta: itemMeta,\n}\nr.Get(\"/productos/{slug}\", res.DetailHTML)\nr.Get(\"/api/productos\", res.ListJSON)"
 },
 {
  "pkg": "jsonld",
  "doc": "Package jsonld builds schema.org structured data as template.JS, ready to\nembed in <script type=\"application/ld+json\"> (SEO).\n\nTwo levels:\n  - Script(typ, props): the primitive. Works for ANY schema.org type, even\n    those that don't yet have a typed builder. It adds @context and @type.\n  - Typed builders (Product, BlogPosting, ...): sugar with compiler-checked\n    fields for the common types. Adding a new one is declaring a struct and a\n    Script() method that delegates to the primitive.\n",
  "symbols": [
   {
    "name": "Script",
    "kind": "func",
    "sig": "func Script(typ string, props map[string]any) template.JS",
    "doc": "Script serializes a schema.org node of the given @type with its properties.\nIt's the framework's extensible primitive.\n"
   },
   {
    "name": "BlogPosting",
    "kind": "type",
    "sig": "type BlogPosting struct {\n\tHeadline\tstring\n\tDescription\tstring\n\tURL\t\tstring\n\tDatePublished\tstring\n}",
    "doc": "BlogPosting → schema.org/BlogPosting.\n",
    "methods": [
     {
      "name": "Script",
      "kind": "func",
      "sig": "func (b BlogPosting) Script() template.JS",
      "doc": ""
     }
    ]
   },
   {
    "name": "Product",
    "kind": "type",
    "sig": "type Product struct {\n\tName\t\tstring\n\tDescription\tstring\n\tURL\t\tstring\n\tPrice\t\tint64\n\tCurrency\tstring\n}",
    "doc": "Product → schema.org/Product (with an Offer if there's a price).\n",
    "methods": [
     {
      "name": "Script",
      "kind": "func",
      "sig": "func (p Product) Script() template.JS",
      "doc": ""
     }
    ]
   }
  ],
  "example": "meta.JSONLD = jsonld.Product{\n    Name: p.Title, Description: p.Desc,\n    URL: url, Price: p.Price, Currency: \"MXN\",\n}.Script() // → <script type=\"application/ld+json\">"
 },
 {
  "pkg": "sitemap",
  "doc": "Package sitemap is the CONTRACT (a leaf, with no internal dependencies) for\nany module to contribute URLs to the sitemap. The site module serves them; the\napp host collects them. It lives apart so app and site don't form an import\ncycle.\n",
  "symbols": [
   {
    "name": "Source",
    "kind": "type",
    "sig": "type Source interface {\n\tSitemapURLs(ctx context.Context) ([]URL, error)\n}",
    "doc": "Source is implemented by anything that wants to appear in the sitemap.\n",
    "methods": [
     {
      "name": "StaticURLs",
      "kind": "func",
      "sig": "func StaticURLs(paths ...string) Source",
      "doc": "StaticURLs creates a source for fixed routes (pages, forms).\n"
     }
    ]
   },
   {
    "name": "URL",
    "kind": "type",
    "sig": "type URL struct {\n\tPath\t\tstring\n\tChangeFreq\tstring\n\tPriority\tstring\n}",
    "doc": "URL is a sitemap entry (relative Path; site prepends the baseURL).\n",
    "methods": [
     {
      "name": "Entries",
      "kind": "func",
      "sig": "func Entries[T any](items []T, prefix, changeFreq, priority string, slug func(T) string) []URL",
      "doc": "Entries builds per-item entries under a prefix (generic, written once).\n"
     }
    ]
   }
  ],
  "example": "// a static page contributes one URL:\na.AddSitemap(sitemap.StaticURLs(\"/quienes-somos\"))\n// a resource contributes one per item (see web.Resource.SitemapSource)."
 },
 {
  "pkg": "logs",
  "doc": "Package logs is an opt-in MODULE: it registers a global middleware that logs\neach request (method, path, status, duration). If you don't plug it in with\napp.Use(logs.Module()), there are no access logs and the code doesn't even\nenter the binary.\n",
  "symbols": [
   {
    "name": "Module",
    "kind": "func",
    "sig": "func Module() app.Module",
    "doc": ""
   }
  ],
  "example": "// one line in main.go's app.Use(...):\na.Use(logs.Module()) // logs method, path, status, duration"
 },
 {
  "pkg": "site",
  "doc": "Package site is a MODULE: it serves robots.txt, sitemap.xml and the embedded\nstatic files. The sitemap is built from ALL the sources the other modules\nregistered in the App; they're read on EVERY request, so the order they were\nwired up in doesn't matter.\n",
  "symbols": [
   {
    "name": "Module",
    "kind": "func",
    "sig": "func Module() app.Module",
    "doc": ""
   },
   {
    "name": "Static",
    "kind": "func",
    "sig": "func Static() http.Handler",
    "doc": "Static returns the handler for the embedded static files (with caching).\n"
   }
  ],
  "example": "// serves robots.txt, sitemap.xml, /static, favicon, styled 404:\na.Use(site.Module())"
 },
 {
  "pkg": "home",
  "doc": "Package home is the starter's only active section: a \"hello world\" landing\nwith a link to the docs. It's the smallest possible module — one route, one\ntemplate. Copy it to start a section of your own; see main.go for the wiring.\n",
  "symbols": [
   {
    "name": "Module",
    "kind": "func",
    "sig": "func Module() app.Module",
    "doc": "Module wires the home into the host. This is the whole pattern: implement\nModule{ Name() string; Register(*App) error } and register routes in Register.\n"
   }
  ],
  "example": "// the smallest module — copy it to start a section:\nfunc (mod) Register(a *app.App) error {\n    a.Router.Get(\"/{$}\", func(w http.ResponseWriter, r *http.Request) {\n        view.Render(w, r, tpl, struct{ Meta view.Meta }{view.Meta{Title: \"agogo\"}})\n    })\n    return nil\n}"
 }
];
