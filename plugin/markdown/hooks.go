package markdown

import (
	"bytes"
	"io"
	"log"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/gomarkdown/markdown/ast"
	"github.com/koron/iview/internal/highlight"
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

// findDetailsEnd finds the end of element, </details>, with considering nested elements.
func findDetailsEnd(data []byte) int {
	curr := 0
	for curr < len(data) {
		end := bytes.Index(data[curr:], []byte(detailsEnd))
		if end < 0 {
			// No </details> found.
			break
		}
		begin := bytes.Index(data[curr:], []byte(detailsBegin))
		if begin < 0 || begin > end {
			// No nested <details> found.
			return curr + end
		}
		// Find end of the nested details.
		nestedBegin := curr + begin + len(detailsBegin)
		nestedEnd := findDetailsEnd(data[nestedBegin:])
		if nestedEnd < 0 {
			// Imcomplete nested <details> element.
			break
		}
		curr = nestedBegin + nestedEnd + len(detailsEnd)
	}
	return -1
}

func parseDetails(data []byte) (ast.Node, []byte, int) {
	if !bytes.HasPrefix(data, []byte(detailsBegin)) {
		return nil, nil, 0
	}
	start := len(detailsBegin)
	end := findDetailsEnd(data[start:])
	if end < 0 {
		return nil, nil, 0
	}
	end += start
	return &Details{}, data[start:end], end + len(detailsEnd)
}

func RenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch n := node.(type) {
	case *Details:
		renderDetails(w, n, entering)
		return ast.GoToNext, true
	case *ast.CodeBlock:
		return renderCode(w, n, entering)
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

func renderCode(w io.Writer, codeBlock *ast.CodeBlock, entering bool) (ast.WalkStatus, bool) {
	lang := string(codeBlock.Info)
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	iter, err := lexer.Tokenise(nil, string(codeBlock.Literal))
	if err != nil {
		log.Printf("renderCode: lexer.Tokenise failed: %s", err)
		return ast.GoToNext, false
	}

	err = highlight.FormatHTML(w, iter)
	if err != nil {
		log.Printf("renderCode: formatter.Format failed: %s", err)
		return ast.GoToNext, false
	}

	return ast.GoToNext, true
}
