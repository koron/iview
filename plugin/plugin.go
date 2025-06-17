// Package plugin provides extensible ponits of iview.
package plugin

import (
	"html/template"
	"io"
	"net/http"
)

const (
	MediaTypeDefault   = "text/plain"
	MediaTypeDirectory = "application/vnd.iview.directory"
)

var ExtToMediaType = map[string][]string{
	".md":  {"text/markdown"},
	".mkd": {"text/markdown"},
}

var InferMediaType func(file http.File, ext string, mediaTypes []string) (string, error) = firstMediaType

func firstMediaType(file http.File, ext string, mediaTypes []string) (string, error) {
	return mediaTypes[0], nil
}

type HTMLRenderer interface {
	Render(w io.Writer, upath string, f http.File) error
}

var MediaTypeToRenderer = map[string]HTMLRenderer{}

var TemplateFuncMap = template.FuncMap{
	// TODO: add "markdown" function for template
}
