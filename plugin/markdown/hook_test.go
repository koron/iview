package markdown

import "testing"

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
	} {
		got := findDetailsEnd([]byte(c.data))
		if got != c.want {
			t.Errorf("case #%d { data=%q, want=%d } failed: got=%d", i, c.data, c.want, got)
		}
	}
}
