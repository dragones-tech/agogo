// nav.js — the site's PARTIAL navigation. This is a piece of agogo's own (not
// lumen): it improves the experience between sections without reloading the document.
//
// It intercepts clicks on internal links, asks the server for ONLY the fragment
// (X-Fragment header) —title + <main> content + the page's scripts— and
// replaces <main>. Header, footer, CSS and this very script stay fixed; nothing
// is fully re-rendered (it takes advantage of the fact that the "content" block
// is already a partial). Without JS, or on any failure, navigation is the
// browser's normal one (full reload), so SEO, refresh and robustness stay intact.

let version = 0; // unique URL per swap → page modules get re-executed

const runnable = (s) => !s.type || s.type === 'module' || /javascript/.test(s.type);

async function navigate(url, { push = true } = {}) {
  let res;
  try {
    res = await fetch(url, { headers: { 'X-Fragment': '1' }, credentials: 'same-origin' });
  } catch {
    location.href = url; // no network → let the browser navigate
    return;
  }
  // If the response is not an HTML fragment (e.g. an /api/... JSON), full nav.
  if (!res.ok || !(res.headers.get('content-type') || '').includes('text/html')) {
    location.href = url;
    return;
  }

  const doc = new DOMParser().parseFromString(await res.text(), 'text/html');
  const newTitle = doc.querySelector('title')?.textContent;

  // EXECUTABLE scripts are re-injected separately (DOMParser doesn't run them);
  // the rest of the body (content, <template>, JSON island) goes inside <main>.
  const scripts = [...doc.body.querySelectorAll('script')].filter(runnable);
  scripts.forEach((s) => s.remove());

  const main = document.querySelector('main');
  const apply = () => {
    if (newTitle) document.title = newTitle;
    main.replaceChildren(...doc.body.childNodes);
  };

  // Native crossfade within the same page, if the browser supports it.
  if (document.startViewTransition) await document.startViewTransition(apply).updateCallbackDone;
  else apply();

  // Re-execute the page modules: unique URL (?v=) → they run again and rewire
  // the new <main>; their imports (lumen, etc.) stay cached. They are injected
  // INSIDE <main>: that way the next swap's replaceChildren cleans them up on
  // its own (otherwise they would pile up in <body> on every navigation).
  version++;
  for (const old of scripts) {
    const fresh = document.createElement('script');
    for (const attr of old.attributes) fresh.setAttribute(attr.name, attr.value);
    if (old.src) fresh.src = old.src + (old.src.includes('?') ? '&' : '?') + 'v=' + version;
    else fresh.textContent = old.textContent;
    main.appendChild(fresh);
  }

  if (push) history.pushState(null, '', url);
  window.scrollTo(0, 0);
}

// Intercept clicks on internal links (delegated on document → survives swaps).
document.addEventListener('click', (e) => {
  if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  const a = e.target.closest('a');
  if (!a || a.target || a.hasAttribute('download') || /\bexternal\b/.test(a.rel)) return;
  const url = new URL(a.href, location.href);
  if (url.origin !== location.origin) return;                 // external link
  if (url.pathname === location.pathname && url.hash) return; // anchor on the same page
  e.preventDefault();
  navigate(url.href);
});

// Back/forward: re-fetch the fragment without pushing another history entry.
addEventListener('popstate', () => navigate(location.href, { push: false }));
