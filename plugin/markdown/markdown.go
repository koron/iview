// Package markdown provides markdown plugin for iview.
package markdown

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/url"
	"sync"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	layoutdto "github.com/koron/iview/layout/dto"
	"github.com/koron/iview/plugin"
)

const MediaType = "text/markdown"

func init() {
	plugin.AddMediaType(MediaType, ".md", ".mkd", ".markdown")

	plugin.AddLayoutDocumentFilter(MediaType, layoutdto.DocumentFilterFunc(markdownDocWrap))

	plugin.AddTemplateGlobalFunc("markdown", func(src string) template.HTML {
		body, _ := ToHTML(src)
		return body
	})
}

type markdownDoc struct {
	layoutdto.Document

	renderOnce    sync.Once
	renderHTML    template.HTML
	renderHeading template.HTML
	renderErr     error
}

func markdownDocWrap(base layoutdto.Document) layoutdto.Document {
	return &markdownDoc{
		Document: base,
	}
}

func (doc *markdownDoc) renderMarkdown() {
	doc.renderOnce.Do(func() {
		// Load contents and render it as markdown.
		var src string
		src, doc.renderErr = doc.ReadAllString()
		if doc.renderErr != nil {
			return
		}
		doc.renderHTML, doc.renderHeading = ToHTML(src)
		return
	})
}

func (doc *markdownDoc) MarkdownBody() (template.HTML, error) {
	doc.renderMarkdown()
	return doc.renderHTML, doc.renderErr
}

func (doc *markdownDoc) MarkdownHeading() (template.HTML, error) {
	doc.renderMarkdown()
	return doc.renderHeading, doc.renderErr
}

func writeInnerText(w io.Writer, node ast.Node) {
	if c := node.AsContainer(); c != nil {
		if len(c.Literal) > 0 {
			w.Write(c.Literal)
		} else if len(c.Content) > 0 {
			w.Write(c.Content)
		}
	} else if l := node.AsLeaf(); l != nil {
		if len(l.Literal) > 0 {
			w.Write(l.Literal)
		} else if len(l.Content) > 0 {
			w.Write(l.Content)
		}
	}
	for _, child := range node.GetChildren() {
		writeInnerText(w, child)
	}
}

func innerText(node ast.Node) string {
	bb := &bytes.Buffer{}
	writeInnerText(bb, node)
	return bb.String()
}

func ToHTML(src string) (body template.HTML, heading template.HTML) {
	doc := markdown.Parse([]byte(src), parser.NewWithExtensions(parser.CommonExtensions|parser.AutoHeadingIDs))

	// For images hosted locally, add the "raw" parameter to the URL to display
	// the image as is.
	ast.WalkFunc(doc, func(rawNode ast.Node, entering bool) ast.WalkStatus {
		switch node := rawNode.(type) {
		case *ast.Image:
			if entering {
				break
			}
			u, err := url.Parse(string(node.Destination))
			if err == nil && u.Scheme == "" && u.Host == "" {
				u.RawQuery = "raw"
				node.Destination = []byte(u.String())
			}

		case *ast.Heading:
			if entering {
				break
			}
			log.Printf("entering=%t level=%d id=%+v %q", entering, node.Level, node.HeadingID, innerText(node))
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
	return template.HTML(dst), "" // TODO:
}
