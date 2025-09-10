package markdown

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

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

const indent = "  "

type indexWriter struct {
	bytes.Buffer
	level int
}

func (iw *indexWriter) indentString() string {
	return strings.Repeat(indent, iw.level)
}

func (iw *indexWriter) addHeading(node *ast.Heading) {
	var (
		level = node.Level
		id    = node.HeadingID
		text  = innerText(node)
	)
	if level > iw.level {
		first := iw.level
		for level > iw.level {
			if iw.level > first {
				fmt.Fprintf(iw, "%s<li>", iw.indentString())
			}
			fmt.Fprint(iw, "<ul>\n")
			iw.level++
		}
	} else if level < iw.level {
		for level < iw.level {
			fmt.Fprint(iw, "</li>\n")
			iw.level--
			fmt.Fprintf(iw, "%s</ul></li>\n", iw.indentString())
		}
	} else {
		fmt.Fprint(iw, "</li>\n")
	}
	fmt.Fprintf(iw, "%s<li><a href=\"#%s\">%s</a>", iw.indentString(), html.EscapeString(id), html.EscapeString(text))
}

func (iw *indexWriter) closeHeading() {
	for iw.level > 0 {
		fmt.Fprint(iw, "</li>\n")
		iw.level--
		fmt.Fprintf(iw, "%s</ul>", iw.indentString())
		if iw.level == 0 {
			fmt.Fprint(iw, "\n")
		}
	}
}

func (iw *indexWriter) html() template.HTML {
	iw.closeHeading()
	return template.HTML(iw.String())
}
