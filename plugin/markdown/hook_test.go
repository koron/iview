package markdown

import (
	"testing"
)

func TestFindDetailsEnd(t *testing.T) {
	for i, c := range []struct {
		data string
		want int
	}{
		{"", -1},
		{"</details>", 0},
		{"<details></details>", -1},
		{"<details></details></details>", 19},
		{"<details></details><details></details>", -1},
		{"<details></details><details></details></details>", 38},
		{"<details><details></details></details>", -1},
		{"<details><details></details></details></details>", 38},

		{"aaa", -1},
		{"aaa</details>", 3},
		{"aaa<details>bbb</details>", -1},
		{"aaa<details>bbb</details>ccc</details>", 28},
		{"aaa<details>bbb<details>ccc</details>ddd</details>eee</details>", 53},
	} {
		got := findDetailsEnd([]byte(c.data))
		if got != c.want {
			t.Errorf("case #%d { data=%q, want=%d } failed: got=%d", i, c.data, c.want, got)
		}
	}
}
