package layout

import (
	"errors"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"

	"github.com/alecthomas/chroma/v2"
	"github.com/koron/iview/internal/templatefs"
	"github.com/koron/iview/plugin"
)

type Renderer struct {
	*template.Template

	MediaType string
	Lexer     chroma.Lexer
	ExtHead   template.HTML
}

func OpenRenderer(fsys *templatefs.FS, mediaType string, lexer chroma.Lexer) (*Renderer, error) {
	opts := []templatefs.Option{
		templatefs.OptionFunc(func(tmpl *template.Template) (*template.Template, error) {
			tmpl.Funcs(plugin.GetTemplateGlobalFuncMap())
			funcMap := plugin.GetTemplateMediaTypeFuncMap(mediaType)
			if funcMap == nil {
				return tmpl, nil
			}
			return tmpl.Funcs(funcMap), nil
		}),
	}

	// Load layout and main templates.
	tmpl, err := fsys.Template2("layout.html", path.Join(mediaType, "main.html"), opts...)
	if err != nil {
		return nil, err
	}

	// Load layout extensions by media type
	head, err := loadLayoutExt(fsys, mediaType, "head")
	if err != nil {
		return nil, err
	}

	return &Renderer{
		Template:  tmpl,
		MediaType: mediaType,
		Lexer:     lexer,
		ExtHead:   head,
	}, nil
}

func loadLayoutExt(fsys fs.FS, mediaType, name string) (template.HTML, error) {
	f, err := fsys.Open(path.Join(mediaType, "layout_ext_"+name+".html"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return template.HTML(b), nil
}

func extractFilename(f http.File) string {
	if namer, ok := f.(interface{ Filename() string }); ok {
		return namer.Filename()
	}
	return ""
}

func (r *Renderer) Render(w io.Writer, rawPath string, f http.File) error {
	doc := NewDoc(f,
		DocWithPath(rawPath),
		DocWithFilename(extractFilename(f)),
		DocWithExtHead(r.ExtHead),
		DocWithLexer(r.Lexer),
	)
	// Apply layout document filters.
	for _, f := range plugin.GetLayoutDocumentFilters(r.MediaType) {
		doc = f.Apply(doc)
	}
	return r.Execute(w, doc)
}
