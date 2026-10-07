package service

import "github.com/microcosm-cc/bluemonday"

// global sanitizer built once in init, reused
var htmlPolicy *bluemonday.Policy

func init() {
	p := bluemonday.UGCPolicy()

	// structure/layout tags: UGCPolicy allows most; add required ones here
	p.AllowElements(
		"h1", "h2", "h3", "h4", "h5", "h6",
		"p", "br",
		"strong", "b", "em", "i", "u", "s",
		"ul", "ol", "li", "blockquote",
		"table", "thead", "tbody", "tr", "th", "td", "colgroup", "col",
		"div", "span", "hr", "header", "footer",
	)

	// keep links href only; UGCPolicy strips javascript:/data: schemes
	p.AllowAttrs("href").OnElements("a")
	// keep img src/alt
	p.AllowAttrs("src", "alt").OnElements("img")

	// allow style attr on these layout tags
	p.AllowAttrs("style").OnElements(
		"p", "div", "span", "h1", "h2", "h3", "h4", "h5", "h6",
		"td", "th", "li", "blockquote", "img",
	)
	// style: keep only safe layout attrs, strip the rest
	p.AllowStyles(
		"text-align",
		"font-weight", "font-style",
		"text-decoration",
		"color", "background-color",
		"width", "height",
		"float",
		"margin", "padding",
		"page-break-after",
		"column-count",
	).Globally()

	htmlPolicy = p
}

// SanitizeHTML before DB write; strips script/iframe/on*/javascript:
func SanitizeHTML(s string) string {
	return htmlPolicy.Sanitize(s)
}
