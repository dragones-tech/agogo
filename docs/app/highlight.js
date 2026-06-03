// @ts-check
// A tiny, dependency-free syntax highlighter (no library, no build — on brand).
// One pass over the code: comments and strings are matched as whole tokens FIRST,
// so keywords inside them aren't recolored. Good enough for the curated samples
// here (Go, bash, trees). Colors live in index.html and adapt to light/dark.

const KEYWORDS = [
  'func', 'package', 'import', 'type', 'struct', 'interface', 'return', 'var',
  'const', 'range', 'for', 'if', 'else', 'go', 'defer', 'map', 'chan', 'select',
  'switch', 'case', 'nil', 'true', 'false', 'new', 'make',
];

const TOKEN = new RegExp(
  '(\\/\\/[^\\n]*|#[^\\n]*)' +                 // 1: line comment (// or #)
    '|("(?:[^"\\\\]|\\\\.)*"|`[^`]*`)' +       // 2: string ("…" or `…`)
    '|\\b(' + KEYWORDS.join('|') + ')\\b' +    // 3: keyword
    '|\\b(\\d[\\w.]*)\\b',                      // 4: number
  'g',
);

const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

function highlight(code) {
  return esc(code).replace(TOKEN, (m, comment, str, kw, num) =>
    comment ? `<span class="hl-c">${comment}</span>`
      : str ? `<span class="hl-s">${str}</span>`
      : kw ? `<span class="hl-k">${kw}</span>`
      : num ? `<span class="hl-n">${num}</span>`
      : m,
  );
}

/** Highlight every <pre><code> under `root`. */
export function highlightAll(root) {
  for (const el of root.querySelectorAll('pre code')) {
    el.innerHTML = highlight(el.textContent || '');
  }
}
