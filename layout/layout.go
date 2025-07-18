package layout

import (
	"bytes"
	"html/template"
	"io"
	"io/fs"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/koron/iview/layout/dto"
)

//////////////////////////////////////////////////////////////////////////////
// DocBase

type DocBase struct {
	file    DocFile
	rawPath string
	extHead template.HTML
	lexer   chroma.Lexer
}

var _ dto.Document = (*DocBase)(nil)

type DocFile interface {
	Read([]byte) (int, error)
	Readdir(int) ([]fs.FileInfo, error)
	Stat() (fs.FileInfo, error)
}

type DocOption interface {
	apply(*DocBase)
}

type DocOptionFunc func(*DocBase)

func (f DocOptionFunc) apply(doc *DocBase) { f(doc) }

func DocWithPath(path string) DocOption {
	return DocOptionFunc(func(doc *DocBase) {
		doc.rawPath = path
	})
}

func DocWithExtHead(extHead template.HTML) DocOption {
	return DocOptionFunc(func(doc *DocBase) {
		doc.extHead = extHead
	})
}

func DocWithLexer(lexer chroma.Lexer) DocOption {
	return DocOptionFunc(func(doc *DocBase) {
		doc.lexer = lexer
	})
}

func NewDoc(file DocFile, options ...DocOption) dto.Document {
	doc := &DocBase{
		file: file,
	}
	for _, opt := range options {
		opt.apply(doc)
	}
	return doc
}

func (doc *DocBase) Name() (string, error) {
	fi, err := doc.file.Stat()
	if err != nil {
		return "", err
	}
	return fi.Name(), nil
}

func (doc *DocBase) Path() (string, error) {
	fi, err := doc.file.Stat()
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		return doc.rawPath + "/", nil
	}
	return doc.rawPath, nil
}

func (doc *DocBase) Breadcrumbs() ([]dto.Link, error) {
	dirs := strings.Split(doc.rawPath, "/")
	if len(dirs) < 2 {
		return nil, nil
	}
	if dirs[len(dirs)-1] == "" {
		dirs = dirs[:len(dirs)-1]
	}
	links := append(make([]dto.Link, 0, len(dirs)), dto.Link{Name: "(Root)", Path: "/"})
	for _, d := range dirs[1:] {
		links = append(links, dto.Link{
			Name: d,
			Path: links[len(links)-1].Path + d + "/",
		})
	}
	links[len(links)-1].Path = ""
	return links, nil
}

func (doc *DocBase) Read(b []byte) (int, error) {
	return doc.file.Read(b)
}

func (doc *DocBase) Readdir(count int) ([]fs.FileInfo, error) {
	return doc.file.Readdir(count)
}

func (doc *DocBase) ReadAllString() (string, error) {
	b, err := io.ReadAll(doc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (doc *DocBase) IsHighlighted() bool {
	return doc.lexer != nil
}

func (doc *DocBase) HighlightName() string {
	if doc.lexer == nil {
		return ""
	}
	if doc.lexer == lexers.Fallback {
		return "Plain"
	}
	return doc.lexer.Config().Name
}

func (doc *DocBase) HightlightCSS() (template.CSS, error) {
	formatter := html.New(html.WithClasses(true))
	style := styles.GitHub
	bb := &bytes.Buffer{}
	err := formatter.WriteCSS(bb, style)
	if err != nil {
		return "", err
	}
	return template.CSS(bb.String()), nil
}

func (doc *DocBase) HightlightedHTML() (template.HTML, error) {
	formatter := html.New(
		html.WithClasses(true),
		html.WithLineNumbers(true),
		html.WithLinkableLineNumbers(true, "L"),
		html.LineNumbersInTable(false),
	)
	style := styles.GitHub
	s, err := doc.ReadAllString()
	if err != nil {
		return "", err
	}
	iter, err := doc.lexer.Tokenise(nil, s)
	if err != nil {
		return "", err
	}
	bb := &bytes.Buffer{}
	err = formatter.Format(bb, style, iter)
	if err != nil {
		return "", err
	}
	return template.HTML(bb.String()), nil
}

func (doc *DocBase) ExtHead() (template.HTML, error) {
	return doc.extHead, nil
}
