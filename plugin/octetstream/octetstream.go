package octetstream

import (
	"html/template"
	"io"

	"github.com/koron/iview/plugin"
)

func init() {
	plugin.AddTemplateMediaTypeFuncMap("application/octet-stream", template.FuncMap{
		"readbytes": ReadBytes,
	})
}

func ReadBytes(r io.Reader, n int) ([]byte, error) {
	b := make([]byte, n)
	m, err := r.Read(b)
	if err != nil {
		return nil, err
	}
	return b[:m], nil
}
