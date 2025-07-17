// Package plugin provides extensible ponits of iview.
package plugin

import (
	"html/template"
	"io"
	"net/http"

	layoutdto "github.com/koron/iview/layout/dto"
)

const (
	MediaTypeBinary    = "application/octet-stream"
	MediaTypeDirectory = "application/vnd.iview.directory"
	MediaTypePlainText = "text/plain"

	MediaTypeDefault = MediaTypeBinary
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

var globalFuncMap = template.FuncMap{}

func GetTemplateGlobalFuncMap() template.FuncMap {
	return globalFuncMap
}

func AddTemplateGlobalFunc(name string, fn any) {
	// TODO: detect override of function.
	globalFuncMap[name] = fn
}

var mediaTypeFuncMap = map[string]template.FuncMap{}

func GetTemplateMediaTypeFuncMap(mediaType string) template.FuncMap {
	return mediaTypeFuncMap[mediaType]
}

func AddTemplateMediaTypeFuncMap(mediaType string, funcMap template.FuncMap) {
	mediaTypeFuncMap[mediaType] = funcMap
}

var layoutDocumentFilters = map[string][]layoutdto.DocumentFilter{}

func AddLayoutDocumentFilter(mediaType string, filters...layoutdto.DocumentFilter) {
	if len(filters) == 0 {
		return
	}
	curr := layoutDocumentFilters[mediaType]
	layoutDocumentFilters[mediaType] = append(curr, filters...)
}

func GetLayoutDocumentFilters(mediaType string) []layoutdto.DocumentFilter {
	return layoutDocumentFilters[mediaType]
}
