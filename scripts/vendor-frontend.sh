#!/usr/bin/env bash
# Vendoriza el frontend hermano (scc + lumen) dentro de internal/site/static.
#
# scc y lumen viven en SUS repos (dragones-tech/scc, dragones-tech/lumen) y se
# trabajan allí; agogo solo guarda una COPIA para servirla desde 'self' (CSP
# estricta, sin CDN). Esas copias NO se versionan en agogo (.gitignore): se
# re-bajan con este script, como un `npm install`.
#
# Uso:  ./scripts/vendor-frontend.sh
#
# La MARCA de agogo NO va aquí: vive en internal/site/static/style.css (sobrescribe
# las perillas de scc), así el vendor queda pristino y re-sincronizable.
set -euo pipefail

cd "$(dirname "$0")/.."
SCC_BASE="https://raw.githubusercontent.com/dragones-tech/scc/main"
LUMEN_BASE="https://raw.githubusercontent.com/dragones-tech/lumen/main"
STATIC="internal/site/static"

echo "→ scc"
mkdir -p "$STATIC/scc/components"
for f in reset theme tokens base animations utilities scc; do
  curl -fsSL "$SCC_BASE/$f.css" -o "$STATIC/scc/$f.css"
done
for c in accordion alert avatar badge breadcrumb button card chip choice dialog \
         divider field hero icon input kbd menu meter nav pagination progress \
         select skeleton spinner switch table tabs toast tooltip; do
  curl -fsSL "$SCC_BASE/components/$c.css" -o "$STATIC/scc/components/$c.css"
done

echo "→ lumen"
mkdir -p "$STATIC/lumen/src"
for f in animate collection-view collection dom element event-emitter http i18n \
         index model region router validate view; do
  curl -fsSL "$LUMEN_BASE/src/$f.js" -o "$STATIC/lumen/src/$f.js"
done

echo "✓ vendorizado: $(find "$STATIC/scc" -name '*.css' | wc -l) scc, $(find "$STATIC/lumen" -name '*.js' | wc -l) lumen"
