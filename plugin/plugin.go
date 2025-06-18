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

var mediaTypesMap = map[string][]string{
	".md":  {"text/markdown"},
	".mkd": {"text/markdown"},
}

func GetMediaType(ext string) ([]string, bool) {
	mt, ok := mediaTypesMap[ext]
	return mt, ok
}

func AddMediaType(mediaType string, exts ...string) {
	for _, ext := range exts {
		mediaTypes, ok := mediaTypesMap[ext]
		if !ok {
			mediaTypes = make([]string, 0, 1)
		}
		mediaTypes = append(mediaTypes, mediaType)
		// TODO: clean up duplications in mediaTypes
		mediaTypesMap[ext] = mediaTypes
	}
}

var InferMediaType func(file http.File, ext string, mediaTypes []string) (string, error) = firstMediaType

func firstMediaType(file http.File, ext string, mediaTypes []string) (string, error) {
	return mediaTypes[0], nil
}

type HTMLRenderer interface {
	Render(w io.Writer, upath string, f http.File) error
}

var MediaTypeToRenderer = map[string]HTMLRenderer{}

var funcMap = template.FuncMap{}

func GetTemplateFuncMap() template.FuncMap {
	return funcMap
}

func AddTemplateFunc(name string, fn any) {
	// TODO: detect override of function.
	funcMap[name] = fn
}
