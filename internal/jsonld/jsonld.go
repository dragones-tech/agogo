// Package jsonld builds schema.org structured data as template.JS, ready to
// embed in <script type="application/ld+json"> (SEO).
//
// Two levels:
//   - Script(typ, props): the primitive. Works for ANY schema.org type, even
//     those that don't yet have a typed builder. It adds @context and @type.
//   - Typed builders (Product, BlogPosting, ...): sugar with compiler-checked
//     fields for the common types. Adding a new one is declaring a struct and a
//     Script() method that delegates to the primitive.
package jsonld

import (
	"encoding/json"
	"html/template"
)

// Script serializes a schema.org node of the given @type with its properties.
// It's the framework's extensible primitive.
func Script(typ string, props map[string]any) template.JS {
	node := map[string]any{
		"@context": "https://schema.org",
		"@type":    typ,
	}
	for k, v := range props {
		node[k] = v
	}
	b, _ := json.Marshal(node)
	return template.JS(b)
}

// --- typed builders for common types (batteries included) ---

// Product → schema.org/Product (with an Offer if there's a price).
type Product struct {
	Name        string
	Description string
	URL         string
	Price       int64
	Currency    string
}

func (p Product) Script() template.JS {
	props := map[string]any{
		"name":        p.Name,
		"description": p.Description,
		"url":         p.URL,
	}
	if p.Price > 0 {
		props["offers"] = map[string]any{
			"@type":         "Offer",
			"price":         p.Price,
			"priceCurrency": p.Currency,
		}
	}
	return Script("Product", props)
}

// BlogPosting → schema.org/BlogPosting.
type BlogPosting struct {
	Headline      string
	Description   string
	URL           string
	DatePublished string
}

func (b BlogPosting) Script() template.JS {
	return Script("BlogPosting", map[string]any{
		"headline":      b.Headline,
		"description":   b.Description,
		"url":           b.URL,
		"datePublished": b.DatePublished,
	})
}
