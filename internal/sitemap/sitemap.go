// Package sitemap is the CONTRACT (a leaf, with no internal dependencies) for
// any module to contribute URLs to the sitemap. The site module serves them; the
// app host collects them. It lives apart so app and site don't form an import
// cycle.
package sitemap

import "context"

// URL is a sitemap entry (relative Path; site prepends the baseURL).
type URL struct {
	Path       string
	ChangeFreq string
	Priority   string
}

// Source is implemented by anything that wants to appear in the sitemap.
type Source interface {
	SitemapURLs(ctx context.Context) ([]URL, error)
}

// Entries builds per-item entries under a prefix (generic, written once).
func Entries[T any](items []T, prefix, changeFreq, priority string, slug func(T) string) []URL {
	urls := make([]URL, 0, len(items))
	for _, it := range items {
		urls = append(urls, URL{Path: prefix + slug(it), ChangeFreq: changeFreq, Priority: priority})
	}
	return urls
}

// StaticURLs creates a source for fixed routes (pages, forms).
func StaticURLs(paths ...string) Source { return staticURLs(paths) }

type staticURLs []string

func (s staticURLs) SitemapURLs(ctx context.Context) ([]URL, error) {
	urls := make([]URL, 0, len(s))
	for _, p := range s {
		urls = append(urls, URL{Path: p, ChangeFreq: "yearly", Priority: "0.5"})
	}
	return urls, nil
}
