// Package markdown provides markdown plugin for iview.
package markdown

import (
	"html/template"
	"net/url"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/koron/iview/plugin"
)

func init() {
	plugin.AddMediaType("text/markdown", ".md", ".mkd", ".markdown")
	plugin.AddTemplateFunc("markdown", ToHTML)
}

func ToHTML(src string) template.HTML {
	doc := markdown.Parse([]byte(src), parser.NewWithExtensions(parser.CommonExtensions|parser.AutoHeadingIDs))

	// For images hosted locally, add the "raw" parameter to the URL to display
	// the image as is.
	ast.WalkFunc(doc, func(rawNode ast.Node, entering bool) ast.WalkStatus {
		switch node := rawNode.(type) {
		case *ast.Image:
			u, err := url.Parse(string(node.Destination))
			if err == nil && u.Scheme == "" && u.Host == "" {
				u.RawQuery = "raw"
				node.Destination = []byte(u.String())
			}
		}
		return ast.GoToNext
	})

	dst := markdown.Render(doc, html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags |
			html.NofollowLinks |
			html.NoreferrerLinks |
			html.NoopenerLinks |
			//html.HrefTargetBlank |
			html.FootnoteReturnLinks,
	}))
	return template.HTML(dst)
}
