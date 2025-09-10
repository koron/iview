package markdown

import (
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/google/go-cmp/cmp"
)

func testHeading(t *testing.T, want string, headings ...*ast.Heading) {
	t.Helper()
	iw := &indexWriter{}
	for _, h := range headings {
		iw.addHeading(h)
	}
	want = strings.TrimLeft(want, "\n")
	if d := cmp.Diff(want, string(iw.html())); d != "" {
		t.Errorf("unexpected heading index: -want +got\n%s", d)
	}
}

func heading(level int, id, text string) *ast.Heading {
	return &ast.Heading{
		Container: ast.Container{
			Content: []byte(text),
		},
		Level:     level,
		HeadingID: id,
	}
}

func TestIndexWriter(t *testing.T) {
	testHeading(t, `
<ul>
  <li><a href="#foo">Foo</a></li>
</ul>
`,
		heading(1, "foo", "Foo"),
	)

	testHeading(t, `
<ul>
  <li><a href="#foo">Foo</a></li>
  <li><a href="#bar">Bar</a></li>
  <li><a href="#baz">Baz</a></li>
</ul>
`,
		heading(1, "foo", "Foo"),
		heading(1, "bar", "Bar"),
		heading(1, "baz", "Baz"),
	)

	testHeading(t, `
<ul>
  <li><a href="#foo">Foo</a><ul>
    <li><a href="#bar">Bar</a></li>
    <li><a href="#baz">Baz</a></li>
  </ul></li>
</ul>
`,
		heading(1, "foo", "Foo"),
		heading(2, "bar", "Bar"),
		heading(2, "baz", "Baz"),
	)

	testHeading(t, `
<ul>
  <li><a href="#foo">Foo</a><ul>
    <li><a href="#bar">Bar</a></li>
    <li><a href="#baz">Baz</a></li>
  </ul></li>
  <li><a href="#quux">Quux</a></li>
</ul>
`,
		heading(1, "foo", "Foo"),
		heading(2, "bar", "Bar"),
		heading(2, "baz", "Baz"),
		heading(1, "quux", "Quux"),
	)
}

func TestIndexWriter_Irregular(t *testing.T) {
	testHeading(t, `
<ul>
  <li><ul>
    <li><a href="#second">Second</a></li>
  </ul></li>
</ul>
`,
		heading(2, "second", "Second"),
	)

	testHeading(t, `
<ul>
  <li><ul>
    <li><a href="#second">Second</a></li>
  </ul></li>
  <li><a href="#foo">Foo</a></li>
</ul>
`,
		heading(2, "second", "Second"),
		heading(1, "foo", "Foo"),
	)

	testHeading(t, `
<ul>
  <li><ul>
    <li><ul>
      <li><a href="#third">Third</a></li>
    </ul></li>
  </ul></li>
</ul>
`,
		heading(3, "third", "Third"),
	)

	testHeading(t, `
<ul>
  <li><ul>
    <li><ul>
      <li><a href="#third">Third</a><ul>
        <li><a href="#fourth">Fourth</a></li>
      </ul></li>
    </ul></li>
  </ul></li>
</ul>
`,
		heading(3, "third", "Third"),
		heading(4, "fourth", "Fourth"),
	)
}
