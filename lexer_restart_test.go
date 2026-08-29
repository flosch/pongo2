package pongo2_test

import (
	"strings"
	"testing"

	"github.com/flosch/pongo2/v7"
)

func TestLexerRestartsAfterCommentAndVerbatim(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source string
		want   string
	}{
		"adjacent comments": {
			source: "{# a #}{# b #}X",
			want:   "X",
		},
		"verbatim after comment": {
			source: "{# a #}{% verbatim %}{{ raw }}{% endverbatim %}",
			want:   "{{ raw }}",
		},
		"adjacent verbatim regions": {
			source: "{% verbatim %}a{% endverbatim %}{% verbatim %}b{% endverbatim %}",
			want:   "ab",
		},
		"comment after verbatim": {
			source: "{% verbatim %}{# literal #}{% endverbatim %}{# drop #}X",
			want:   "{# literal #}X",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tpl, err := pongo2.FromString(test.source)
			if err != nil {
				t.Fatalf("FromString(%q): %v", test.source, err)
			}
			got, err := tpl.Execute(nil)
			if err != nil {
				t.Fatalf("Execute(%q): %v", test.source, err)
			}
			if got != test.want {
				t.Fatalf("Execute(%q) = %q, want %q", test.source, got, test.want)
			}
		})
	}
}

// This differential oracle checks the Django comment contract: a short
// comment produces nothing and every other fragment produces its own output.
func TestTemplateCommentSemanticsMatchDjangoOracle(t *testing.T) {
	t.Parallel()
	fragments := []struct{ source, output string }{
		{"A", "A"},
		{"{# c #}", ""},
		{"{#x#}", ""},
		{" ", " "},
		{`{{ "v" }}`, "v"},
		{"{% if true %}T{% endif %}", "T"},
		{"{% verbatim %}{{ raw }}{% endverbatim %}", "{{ raw }}"},
	}

	sequence := make([]int, 0, 3)
	cases := 0
	var walk func(int)
	walk = func(depth int) {
		if depth != 0 {
			var source, want strings.Builder
			for _, index := range sequence {
				source.WriteString(fragments[index].source)
				want.WriteString(fragments[index].output)
			}
			cases++
			tpl, err := pongo2.FromString(source.String())
			if err != nil {
				t.Fatalf("FromString(%q): %v", source.String(), err)
			}
			got, err := tpl.Execute(nil)
			if err != nil {
				t.Fatalf("Execute(%q): %v", source.String(), err)
			}
			if got != want.String() {
				t.Fatalf("%q rendered %q, Django renders %q",
					source.String(), got, want.String())
			}
		}
		if depth == cap(sequence) {
			return
		}
		for index := range fragments {
			sequence = append(sequence, index)
			walk(depth + 1)
			sequence = sequence[:len(sequence)-1]
		}
	}
	walk(0)
	wantCases := len(fragments) + len(fragments)*len(fragments) +
		len(fragments)*len(fragments)*len(fragments)
	if cases != wantCases {
		t.Fatalf("generated %d sources, want %d", cases, wantCases)
	}
}
