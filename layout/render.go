package layout

import (
	"html/template"
	"io"
	"net/http"
)

type Renderer struct {
	*template.Template

	Extensions *Extensions
}

func (r *Renderer) Render(w io.Writer, rawPath string, f http.File) error {
	data := &Data{
		File:       f,
		RawPath:    rawPath,
		Extensions: r.Extensions,
	}
	return r.Execute(w, data)
}
