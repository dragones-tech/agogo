// Package jsonld construye datos estructurados schema.org como template.JS,
// listos para incrustar en <script type="application/ld+json"> (SEO).
//
// Dos niveles:
//   - Script(typ, props): el primitivo. Sirve para CUALQUIER tipo schema.org,
//     incluso los que aún no tienen builder tipado. Añade @context y @type.
//   - Builders tipados (Product, BlogPosting, ...): azúcar con campos
//     verificados por el compilador para los tipos comunes. Agregar uno nuevo
//     es declarar un struct y un método Script() que delega en el primitivo.
package jsonld

import (
	"encoding/json"
	"html/template"
)

// Script serializa un nodo schema.org del @type dado con sus propiedades.
// Es el primitivo extensible del framework.
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

// --- builders tipados para tipos comunes (batteries included) ---

// Product → schema.org/Product (con Offer si hay precio).
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
