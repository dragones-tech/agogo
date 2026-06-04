// @ts-check
import { I18n } from 'lumenjs';

/**
 * The docs' single I18n instance (Lumen's i18n). Only PROSE is translated; code
 * blocks stay as authored in the templates. Views fill their [data-i18n] nodes
 * with applyI18n() and re-render when the navbar toggle switches locale.
 */
export const i18n = new I18n({
  locale: 'es',
  fallback: 'en',
  messages: {
    en: {
      nav: {
        home: 'Home', guide: 'Guide', modules: 'Modules',
        start: 'Get started', structure: 'Structure', module: 'Create a module',
        activate: 'Activate modules', config: 'Configuration',
      },
      home: {
        lead: 'A compact web server in <strong>pure Go</strong>: server-rendered HTML (for SEO) and a JSON API, from a single source of truth. No frameworks, no magic.',
        cta1: 'Get started', cta2: 'GitHub',
        f1t: 'Only the essentials', f1d: 'HTTP, templates and JSON come from Go\'s stdlib. The only dependency is the SQLite driver (pure Go).',
        f2t: 'No magic', f2d: 'No ORM, no reflection-driven config, no runtime black boxes. The code that runs is the code you read.',
        f3t: 'Modular', f3d: 'A host (<code>app</code>) you plug opt-in modules into with <code>app.Use(...)</code>. What you don\'t wire in never enters the binary.',
        f4t: 'Secure by default', f4d: 'Security headers + CSP, gzip, body limit, graceful shutdown, signed sessions, PBKDF2 and CSRF — all stdlib.',
      },
      start: {
        title: 'Get started',
        lead: 'Requires <strong>Go 1.26+</strong>. No build step, nothing to install (scc and lumen load from a CDN).',
        s1t: 'Clone and run', s2t: 'Open the browser', s2d: 'You\'ll see a centered "¡Hola, mundo!" at <code>http://localhost:46060</code> and this documentation.',
        s3t: 'Create your first section', s3d: 'Copy <code>internal/home</code>, adjust its <code>Register</code> and wire it in <code>main.go</code>.',
      },
      structure: {
        title: 'Structure',
        lead: 'A core + toolbox (base packages), plus opt-in sections (modules).',
        body: 'Each module implements <code>Module{ Name() string; Register(*App) error }</code> and in its <code>Register</code> hooks into the host: routes (<code>a.Router.Get</code>), global middleware (<code>a.UseMiddleware</code>), migrations (<code>a.AddMigration</code>), sitemap sources (<code>a.AddSitemap</code>) and shared services (<code>a.DB</code>, <code>a.Session</code>, <code>a.Identity</code>). The <code>app.Use(...)</code> list in <code>main.go</code> is the overview of what the server does.',
      },
      module: {
        title: 'Create a module',
        lead: 'A section is a module: a type that implements the <code>Module</code> interface and registers its routes in <code>Register</code>. The smallest one is <code>internal/home</code> — copy it.',
        s1t: '1. The interface', s1d: 'A module is anything with a name and a <code>Register</code> that hooks into the host.',
        s2t: '2. The module', s2d: 'Embed templates, build them once, and register routes (and migrations/sitemap) in <code>Register</code>.',
        s3t: '3. Wire it', s3d: 'Add one line to <code>app.Use(...)</code> in <code>main.go</code>. Comment it out to remove the feature (and strip it from the binary).',
        hooks: 'Hooks available on <code>*App</code> inside <code>Register</code>:',
        hRoutes: 'register routes (with optional per-route middleware)',
        hMw: 'add global middleware (wraps the whole server)',
        hMig: 'register a schema migration (run by app.Migrate)',
        hMap: 'add a sitemap source',
        hSvc: 'shared services: the DB, the session manager, the identity service',
      },
      base: {
        title: 'Base modules & toolbox',
        lead: 'The core packages in <code>internal/</code>. The toolbox is always available; import what you need. Sections (modules) are built on top.',
        thPkg: 'Package', thWhat: 'What it does',
        app: 'The host: <code>App</code>, the <code>Module</code> interface, and the hook points (<code>Use</code>, <code>UseMiddleware</code>, <code>AddMigration</code>, <code>AddSitemap</code>, <code>Migrate</code>, <code>Handler</code>).',
        router: 'Express-style router over <code>net/http</code>: <code>Get/Post/Put/Delete(path, fn, mw...)</code> with per-route middleware.',
        config: 'Typed env config read once (<code>config.Load</code>); includes the SQLite <code>DSN()</code> (WAL + busy_timeout).',
        view: 'Shared HTML layout + <code>Render</code> (renders to a buffer, so a template error returns a clean 500); styled <code>NotFound</code>/<code>ServerError</code>.',
        middleware: 'Security baseline: <code>Recover</code>, <code>LimitBody</code> (1 MiB), <code>SecurityHeaders</code> (CSP) and <code>Gzip</code>.',
        respond: 'JSON helpers: <code>respond.JSON</code>, <code>respond.Error</code>, <code>respond.ServerError</code> (logs the real error, returns a generic 500).',
        validate: 'A tiny, explicit input validator (an <code>Errors</code> accumulator; no reflection, no struct tags).',
        session: 'Signed-cookie sessions (HMAC), no server store; carries a signed expiry.',
        identity: 'Identity over the session: <code>Login</code>/<code>Logout</code>/<code>UserID</code> and <code>Require</code> (a per-route middleware).',
        password: 'PBKDF2-HMAC-SHA256 hashing with constant-time compare; the stored hash carries its parameters.',
        csrf: 'Double-submit-cookie CSRF for forms: <code>Issue</code> + <code>Valid</code>.',
        web: '<code>Resource[T]</code>: the generic DB resource (list/detail × HTML/JSON) written once; a domain only configures queries, templates and SEO.',
        jsonld: 'schema.org JSON-LD builders for SEO.',
        sitemap: 'The sitemap contract (URL, Source, Entries) consumed by the <code>site</code> module.',
        logs: 'Opt-in global access-logging middleware (method, path, status, duration).',
        site: 'robots.txt, sitemap.xml, <code>/static</code>, favicon and the styled catch-all 404.',
      },
      activate: {
        title: 'Activate modules',
        lead: 'What the app does = the modules it wires in, one line each in <code>app.Use(...)</code>. Comment a line out to drop the feature (and strip it from the binary).',
        baseT: 'Base modules (active out of the box)',
        mLogs: '<code>logs</code> — global middleware that logs each request (method, path, status, duration).',
        mHome: '<code>home</code> — the "¡Hola, mundo!" landing, with links to the docs and the example.',
        mPaginas: '<code>paginas</code> — example static section at <code>/ejemplo</code>; hosts the lumen and over-the-wire demos.',
        mOtw: '<code>otw</code> — a BFF that renders a token-gated API as an HTML fragment (the token never reaches the browser).',
        mSite: '<code>site</code> — robots.txt, sitemap.xml, <code>/static</code>, favicon and the styled 404.',
        integrateT: 'Modules to integrate (opt-in)',
        integrateLead: 'Present in the repo but commented out in <code>main.go</code>: uncomment the line (and its import) to plug them in.',
        mAuth: '<code>auth</code> — username/password login (<code>/login</code>, <code>/cuenta</code>) with PBKDF2 and CSRF. Needs the DB: run <code>go run ./cmd/migrate</code> (and <code>./cmd/seed</code> for a demo user) once before enabling it.',
        mOauth: '<code>oauth</code> — OAuth 2.0 (authorization-code) login in pure stdlib that reuses <code>identity</code>. No DB; <code>/oauth/login</code> answers 503 until you set <code>OAUTH_*</code>.',
        addT: 'How to add and create them',
        addD: 'Wiring is one line in <code>app.Use(...)</code>; what you don\'t wire in isn\'t imported, so it stays out of the binary. To build your own, see <a href="#/modulo">Create a module</a>.',
        note: 'Both <code>auth</code> and <code>oauth</code> reuse the core <code>identity</code> service — a new login mechanism plugs in without rewriting it.',
      },
      api: {
        title: 'API reference',
        lead: 'The framework\'s Go packages — types, functions and methods with their signatures and docs. Generated from the source with <code>go doc</code>; English only (it\'s the code\'s own documentation).',
      },
      config: {
        title: 'Configuration',
        lead: 'Read once at startup, typed and validated (<code>internal/config</code>).',
        thVar: 'Variable', thDef: 'Default', thDesc: 'Description',
        addr: 'Listen address', db: 'SQLite file path (opened in WAL + busy_timeout)',
        baseurl: 'Base URL (canonical, sitemap)', secret: 'Session signing key (HMAC). In production, ≥32 chars.',
        secure: 'Force Secure cookies on/off', insecure: '(insecure dev key)', derived: '(derived from https:// in base URL)',
      },
      err: { title: 'Not found', desc: 'That section doesn\'t exist.', back: 'Back to home' },
    },
    es: {
      nav: {
        home: 'Inicio', guide: 'Guía', modules: 'Módulos',
        start: 'Empezar', structure: 'Estructura', module: 'Crear un módulo',
        activate: 'Activar módulos', config: 'Configuración',
      },
      home: {
        lead: 'Un servidor web compacto en <strong>Go puro</strong>: HTML renderizado en servidor (para SEO) y API JSON, desde una sola fuente de verdad. Sin frameworks, sin magia.',
        cta1: 'Empezar', cta2: 'GitHub',
        f1t: 'Solo lo elemental', f1d: 'HTTP, plantillas y JSON salen de la stdlib de Go. La única dependencia es el driver de SQLite (Go puro).',
        f2t: 'Sin magia', f2d: 'Nada de ORM, reflexión oculta ni cajas negras en runtime. El código que corre es el que lees.',
        f3t: 'Modular', f3d: 'Un host (<code>app</code>) al que acoplas módulos opt-in con <code>app.Use(...)</code>. Lo que no acoplas no entra al binario.',
        f4t: 'Seguro por defecto', f4d: 'Cabeceras + CSP, gzip, límite de cuerpo, apagado ordenado, sesiones firmadas, PBKDF2 y CSRF — todo stdlib.',
      },
      start: {
        title: 'Empezar',
        lead: 'Requiere <strong>Go 1.26+</strong>. Sin build, sin instalar nada (scc y lumen vienen del CDN).',
        s1t: 'Clona y corre', s2t: 'Abre el navegador', s2d: 'Verás un «¡Hola, mundo!» centrado en <code>http://localhost:46060</code> y esta documentación.',
        s3t: 'Crea tu primera sección', s3d: 'Copia <code>internal/home</code>, ajusta su <code>Register</code> y acóplala en <code>main.go</code>.',
      },
      structure: {
        title: 'Estructura',
        lead: 'Un núcleo + caja de herramientas (paquetes base), y secciones (módulos) opt-in.',
        body: 'Cada módulo implementa <code>Module{ Name() string; Register(*App) error }</code> y en su <code>Register</code> engancha al host: rutas (<code>a.Router.Get</code>), middleware global (<code>a.UseMiddleware</code>), migraciones (<code>a.AddMigration</code>), fuentes de sitemap (<code>a.AddSitemap</code>) y servicios compartidos (<code>a.DB</code>, <code>a.Session</code>, <code>a.Identity</code>). La lista <code>app.Use(...)</code> en <code>main.go</code> es el panorama de qué hace el servidor.',
      },
      module: {
        title: 'Crear un módulo',
        lead: 'Una sección es un módulo: un tipo que implementa la interfaz <code>Module</code> y registra sus rutas en <code>Register</code>. El más pequeño es <code>internal/home</code> — cópialo.',
        s1t: '1. La interfaz', s1d: 'Un módulo es cualquier cosa con un nombre y un <code>Register</code> que engancha al host.',
        s2t: '2. El módulo', s2d: 'Embebe plantillas, constrúyelas una vez, y registra rutas (y migraciones/sitemap) en <code>Register</code>.',
        s3t: '3. Acóplalo', s3d: 'Añade una línea a <code>app.Use(...)</code> en <code>main.go</code>. Coméntala para quitar la funcionalidad (y sacarla del binario).',
        hooks: 'Enganches disponibles en <code>*App</code> dentro de <code>Register</code>:',
        hRoutes: 'registra rutas (con middleware por ruta opcional)',
        hMw: 'añade middleware global (envuelve todo el servidor)',
        hMig: 'registra una migración de esquema (la corre app.Migrate)',
        hMap: 'añade una fuente de sitemap',
        hSvc: 'servicios compartidos: la BD, el manager de sesión, el servicio de identidad',
      },
      base: {
        title: 'Módulos base y toolbox',
        lead: 'Los paquetes del núcleo en <code>internal/</code>. El toolbox está siempre disponible; importa lo que necesites. Las secciones (módulos) se construyen encima.',
        thPkg: 'Paquete', thWhat: 'Qué hace',
        app: 'El host: <code>App</code>, la interfaz <code>Module</code> y los puntos de enganche (<code>Use</code>, <code>UseMiddleware</code>, <code>AddMigration</code>, <code>AddSitemap</code>, <code>Migrate</code>, <code>Handler</code>).',
        router: 'Router estilo Express sobre <code>net/http</code>: <code>Get/Post/Put/Delete(path, fn, mw...)</code> con middleware por ruta.',
        config: 'Config tipada del entorno, leída una vez (<code>config.Load</code>); incluye el <code>DSN()</code> de SQLite (WAL + busy_timeout).',
        view: 'Layout HTML compartido + <code>Render</code> (renderiza a buffer, así un error de plantilla da un 500 limpio); <code>NotFound</code>/<code>ServerError</code> con estilo.',
        middleware: 'Baseline de seguridad: <code>Recover</code>, <code>LimitBody</code> (1 MiB), <code>SecurityHeaders</code> (CSP) y <code>Gzip</code>.',
        respond: 'Helpers JSON: <code>respond.JSON</code>, <code>respond.Error</code>, <code>respond.ServerError</code> (loguea el error real, devuelve un 500 genérico).',
        validate: 'Validador de entrada mínimo y explícito (un acumulador <code>Errors</code>; sin reflexión ni struct tags).',
        session: 'Sesiones por cookie firmada (HMAC), sin store en servidor; lleva una expiración firmada.',
        identity: 'Identidad sobre la sesión: <code>Login</code>/<code>Logout</code>/<code>UserID</code> y <code>Require</code> (middleware por ruta).',
        password: 'Hashing PBKDF2-HMAC-SHA256 con comparación en tiempo constante; el hash guardado lleva sus parámetros.',
        csrf: 'CSRF por doble cookie para formularios: <code>Issue</code> + <code>Valid</code>.',
        web: '<code>Resource[T]</code>: el recurso de BD genérico (lista/detalle × HTML/JSON) escrito una sola vez; el dominio solo configura queries, plantillas y SEO.',
        jsonld: 'Builders de JSON-LD schema.org para SEO.',
        sitemap: 'El contrato del sitemap (URL, Source, Entries) que consume el módulo <code>site</code>.',
        logs: 'Middleware global de logging de acceso, opt-in (método, ruta, status, duración).',
        site: 'robots.txt, sitemap.xml, <code>/static</code>, favicon y el 404 con estilo (catch-all).',
      },
      activate: {
        title: 'Activar módulos',
        lead: 'Lo que hace la app = los módulos que acopla, una línea cada uno en <code>app.Use(...)</code>. Comenta una línea para quitar la funcionalidad (y sacarla del binario).',
        baseT: 'Módulos base (activos out of the box)',
        mLogs: '<code>logs</code> — middleware global que registra cada petición (método, ruta, status, duración).',
        mHome: '<code>home</code> — la landing «¡Hola, mundo!», con enlaces a la doc y al ejemplo.',
        mPaginas: '<code>paginas</code> — sección estática de ejemplo en <code>/ejemplo</code>; hospeda los demos de lumen y over-the-wire.',
        mOtw: '<code>otw</code> — un BFF que renderiza una API protegida por token como fragmento HTML (el token nunca llega al navegador).',
        mSite: '<code>site</code> — robots.txt, sitemap.xml, <code>/static</code>, favicon y el 404 con estilo.',
        integrateT: 'Módulos para integrar (opt-in)',
        integrateLead: 'Presentes en el repo pero comentados en <code>main.go</code>: descomenta la línea (y su import) para acoplarlos.',
        mAuth: '<code>auth</code> — login usuario/contraseña (<code>/login</code>, <code>/cuenta</code>) con PBKDF2 y CSRF. Necesita BD: corre <code>go run ./cmd/migrate</code> (y <code>./cmd/seed</code> para un usuario demo) una vez antes de activarlo.',
        mOauth: '<code>oauth</code> — login OAuth 2.0 (authorization-code) en stdlib pura que reusa <code>identity</code>. Sin BD; <code>/oauth/login</code> responde 503 hasta configurar <code>OAUTH_*</code>.',
        addT: 'Cómo agregarlos y crearlos',
        addD: 'Acoplar es una línea en <code>app.Use(...)</code>; lo que no acoplas no se importa, así que no entra al binario. Para crear los tuyos, mira <a href="#/modulo">Crear un módulo</a>.',
        note: 'Tanto <code>auth</code> como <code>oauth</code> reusan el servicio <code>identity</code> del núcleo — un mecanismo de login nuevo se acopla sin reescribirlo.',
      },
      api: {
        title: 'Referencia de la API',
        lead: 'Los paquetes Go del framework — tipos, funciones y métodos con sus firmas y documentación. Generado del código fuente con <code>go doc</code>; solo en inglés (es la documentación propia del código).',
      },
      config: {
        title: 'Configuración',
        lead: 'Leída una vez al arrancar, tipada y validada (<code>internal/config</code>).',
        thVar: 'Variable', thDef: 'Por defecto', thDesc: 'Descripción',
        addr: 'Dirección de escucha', db: 'Ruta del SQLite (abierto en WAL + busy_timeout)',
        baseurl: 'URL base (canonical, sitemap)', secret: 'Firma de sesiones (HMAC). En producción, ≥32 chars.',
        secure: 'Forzar cookies Secure', insecure: '(clave de dev insegura)', derived: '(según https:// en la URL base)',
      },
      err: { title: 'No encontrado', desc: 'Esa sección no existe.', back: 'Volver al inicio' },
    },
  },
});

/**
 * Fill every [data-i18n] node under `root` from the current locale, and re-fill
 * whenever the locale changes (cleaned up via signal).
 * @param {ParentNode} root @param {AbortSignal} [signal]
 */
export function applyI18n(root, signal) {
  const render = () => {
    for (const el of root.querySelectorAll('[data-i18n]')) {
      el.innerHTML = i18n.t(/** @type {HTMLElement} */ (el).dataset.i18n);
    }
  };
  render();
  if (signal) i18n.onChange(render, { signal });
}
