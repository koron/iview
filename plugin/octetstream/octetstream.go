package octetstream

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"iter"

	"github.com/koron/iview/plugin"
)

func init() {
	plugin.AddTemplateMediaTypeFuncMap("application/octet-stream", template.FuncMap{
		"ascii":     ascii,
		"readbytes": readbytes,
		"step":      step,
	})
}

func ascii(data []byte) string {
	bb := &bytes.Buffer{}
	for _, b := range data {
		if b >= 0x20 && b <= 0x7e {
			bb.WriteByte(b)
		} else {
			bb.WriteByte('.')
		}
	}
	return bb.String()
}

func readbytes(r io.Reader, n int) ([]byte, error) {
	b := make([]byte, n)
	m, err := r.Read(b)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return b[:m], nil
}

func step(start, end, step int) iter.Seq[int] {
	if step == 0 {
		step = 1
	}
	return func(yield func(int) bool) {
		for i := start; i < end; i += step {
			if !yield(i) {
				break
			}
		}
	}
}
