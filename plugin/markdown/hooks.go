package markdown

import (
	"bytes"
	"io"

	"github.com/gomarkdown/markdown/ast"
)

func ParserHook(data []byte) (ast.Node, []byte, int) {
	if node, d, n := parseDetails(data); node != nil {
		return node, d, n
	}
	return nil, nil, 0
}

type Details struct {
	ast.Container
}

const (
	detailsBegin = "<details>"
	detailsEnd   = "</details>"
)

func parseDetails(data []byte) (ast.Node, []byte, int) {
	if !bytes.HasPrefix(data, []byte(detailsBegin)) {
		return nil, nil, 0
	}
	start := len(detailsBegin)
	end := bytes.Index(data[start:], []byte(detailsEnd)) + start
	if end < 0 {
		return nil, nil, 0
	}
	return &Details{}, data[start:end], end + len(detailsEnd)
}

func RenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch n := node.(type) {
	case *Details:
		renderDetails(w, n, entering)
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

func renderDetails(w io.Writer, details *Details, entering bool) {
	if entering {
		io.WriteString(w, detailsBegin)
	} else {
		io.WriteString(w, detailsEnd)
	}
}
