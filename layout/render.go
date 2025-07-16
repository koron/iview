package layout

import (
	"html/template"
	"io"
	"net/http"
)

type Renderer struct {
	*template.Template

	Extensions *LayoutExtensions
}

func (r *Renderer) Render(w io.Writer, rawPath string, f http.File) error {
	doc := &BaseDoc{
		File:    f,
		rawPath: rawPath,
	}
	if r.Extensions != nil && r.Extensions.Head != nil {
		doc.extHead = r.Extensions.Head.(template.HTML)
	}
	// TODO:
	return r.Execute(w, doc)
}
