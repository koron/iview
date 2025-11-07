// Package highlight provides/wraps syntax highlighting feature.
package highlight

import (
	"io"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

var defaultStyle = styles.Get("github")

func Style() *chroma.Style {
	return defaultStyle
}

func htmlFormatter(options ...html.Option) *html.Formatter {
	return html.New(append(options, html.WithClasses(true))...)
}

func FormatHTML(w io.Writer, iter chroma.Iterator, options ...html.Option) error {
	return htmlFormatter(options...).Format(w, Style(), iter)
}

func WriteCSS(w io.Writer) error {
	return htmlFormatter().WriteCSS(w, Style())
}
