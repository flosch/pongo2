package pongo2

import (
	"testing"
)

// BenchmarkLexer measures lexer tokenization performance
func BenchmarkLexer(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"keyword_in", "{% for x in items %}{{ x }}{% endfor %}"},
		{"keyword_and_or", "{% if a and b or c %}yes{% endif %}"},
		{"no_keywords", "{{ variable.field.subfield }}"},
		{"mixed", "{% if item in items and active %}{{ item|escape }}{% endif %}"},
		{"many_identifiers", "{{ a.b.c.d.e.f.g.h.i.j }}"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := lex("benchmark", tc.input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkLexerStrings measures string escape handling performance
func BenchmarkLexerStrings(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"simple_string", `{{ "hello world" }}`},
		{"escaped_string", `{{ "hello \"world\" with \\backslash" }}`},
		{"newline_string", `{{ "line1\nline2\ttab" }}`},
		{"multiple_strings", `{{ "one" }}{{ "two" }}{{ "three" }}`},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := lex("benchmark", tc.input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
