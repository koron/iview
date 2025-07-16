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
	data := &Data{
		File:       f,
		RawPath:    rawPath,
		LayoutExtensions: r.Extensions,
	}
	return r.Execute(w, data)
}
