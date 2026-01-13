package pongo2

import (
	"strings"
	"testing"
)

// FuzzLexer directly fuzzes the lexer to find tokenization edge cases.
// This tests the low-level tokenization process without template execution.
func FuzzLexer(f *testing.F) {
	// Basic template structures
	f.Add("{{ variable }}")
	f.Add("{% tag %}")
	f.Add("{# comment #}")
	f.Add("plain text")
	f.Add("")

	// Whitespace variations
	f.Add("{{  }}")
	f.Add("{%  %}")
	f.Add("{#  #}")
	f.Add("{{ variable    }}")
	f.Add("{%    tag    %}")
	f.Add("{#    comment    #}")
	f.Add("{{ \t variable \t }}")
	f.Add("{% \t tag \t %}")

	// Whitespace trim control
	f.Add("{{- variable -}}")
	f.Add("{%- tag -%}")
	f.Add("  {{- x -}}  ")
	f.Add("  {%- x -%}  ")
	f.Add("{{-x-}}")
	f.Add("{%-x-%}")

	// String literals with escapes
	f.Add(`{{ "hello" }}`)
	f.Add(`{{ 'hello' }}`)
	f.Add(`{{ "hello\"world" }}`)
	f.Add(`{{ 'hello\'world' }}`)
	f.Add(`{{ "line1\nline2" }}`)
	f.Add(`{{ "tab\there" }}`)
	f.Add(`{{ "return\rhere" }}`)
	f.Add(`{{ "\\" }}`)
	f.Add(`{{ "nested\"\\\"quotes" }}`)
	f.Add(`{{ "" }}`)
	f.Add(`{{ '' }}`)

	// Numbers
	f.Add("{{ 0 }}")
	f.Add("{{ 1 }}")
	f.Add("{{ -1 }}")
	f.Add("{{ 123 }}")
	f.Add("{{ -123 }}")
	f.Add("{{ 999999999 }}")
	f.Add("{{ 9999999999999999999 }}")
	f.Add("{{ 0.0 }}")
	f.Add("{{ 0.123 }}")
	f.Add("{{ 123.456 }}")
	f.Add("{{ -123.456 }}")
	f.Add("{{ .5 }}")

	// Identifiers
	f.Add("{{ a }}")
	f.Add("{{ abc }}")
	f.Add("{{ abc123 }}")
	f.Add("{{ _underscore }}")
	f.Add("{{ __dunder__ }}")
	f.Add("{{ CamelCase }}")
	f.Add("{{ UPPERCASE }}")
	f.Add("{{ lowercase }}")

	// Keywords
	f.Add("{{ true }}")
	f.Add("{{ false }}")
	f.Add("{{ in }}")
	f.Add("{{ and }}")
	f.Add("{{ or }}")
	f.Add("{{ not }}")
	f.Add("{{ as }}")
	f.Add("{{ export }}")

	// Symbols - 3-char
	f.Add("{{-}}")
	f.Add("-}}")
	f.Add("{%-")
	f.Add("-%}")

	// Symbols - 2-char
	f.Add("{{ == }}")
	f.Add("{{ >= }}")
	f.Add("{{ <= }}")
	f.Add("{{ != }}")
	f.Add("{{ <> }}")
	f.Add("{{ && }}")
	f.Add("{{ || }}")
	f.Add("{{ x }}")
	f.Add("{% x %}")

	// Symbols - 1-char
	f.Add("{{ ( }}")
	f.Add("{{ ) }}")
	f.Add("{{ + }}")
	f.Add("{{ - }}")
	f.Add("{{ * }}")
	f.Add("{{ / }}")
	f.Add("{{ ^ }}")
	f.Add("{{ < }}")
	f.Add("{{ > }}")
	f.Add("{{ , }}")
	f.Add("{{ . }}")
	f.Add("{{ ! }}")
	f.Add("{{ | }}")
	f.Add("{{ : }}")
	f.Add("{{ = }}")
	f.Add("{{ % }}")
	f.Add("{{ [ }}")
	f.Add("{{ ] }}")

	// Complex expressions
	f.Add("{{ (a + b) * c }}")
	f.Add("{{ a.b.c.d }}")
	f.Add("{{ a[0][1][2] }}")
	f.Add("{{ a['key'] }}")
	f.Add(`{{ a["key"] }}`)
	f.Add("{{ func() }}")
	f.Add("{{ func(a, b, c) }}")
	f.Add("{{ a|filter }}")
	f.Add("{{ a|filter:arg }}")
	f.Add("{{ a|f1|f2|f3 }}")

	// Verbatim handling
	f.Add("{% verbatim %}{{ not parsed }}{% endverbatim %}")
	f.Add("{% verbatim %}{% also not parsed %}{% endverbatim %}")
	f.Add("{% verbatim %}{# also not #}{% endverbatim %}")
	f.Add("{% verbatim %}{% endverbatim %}")

	// Comment edge cases
	f.Add("{# simple #}")
	f.Add("{##}")
	f.Add("{#   #}")
	f.Add("{# {{ variable }} #}")
	f.Add("{# {% tag %} #}")
	f.Add("{# {# nested #} #}")

	// Mixed content
	f.Add("before{{ x }}after")
	f.Add("{{ x }}{{ y }}")
	f.Add("{% a %}{% b %}")
	f.Add("{# c #}{# d #}")
	f.Add("{{ x }}{% y %}{# z #}")
	f.Add("text {{ var }} more text {% tag %} end {# comment #}")

	// Unicode content
	f.Add("{{ 你好 }}")
	f.Add("{{ こんにちは }}")
	f.Add("{{ Привіт }}")
	f.Add("{{ Ελληνικά }}")
	f.Add(`{{ "你好世界" }}`)
	f.Add(`{{ "日本語" }}`)
	f.Add(`{{ "🎉🎊🎁" }}`)
	f.Add("你好 {{ var }} 世界")

	// Malformed/incomplete
	f.Add("{{")
	f.Add("}}")
	f.Add("{%")
	f.Add("%}")
	f.Add("{#")
	f.Add("#}")
	f.Add("{{ unclosed")
	f.Add("{% unclosed")
	f.Add("{# unclosed")
	f.Add("{{ }}")
	f.Add("{%%}")
	f.Add("{# #}")
	f.Add("{{ x")
	f.Add("x }}")
	f.Add("{% x")
	f.Add("x %}")

	// Deep nesting
	f.Add("{{ ((((((x)))))) }}")
	f.Add("{{ a.b.c.d.e.f.g.h.i.j }}")
	f.Add("{{ a[0][1][2][3][4][5] }}")
	f.Add("{{ a|b|c|d|e|f|g|h|i|j }}")

	// Special characters in strings
	f.Add(`{{ "\x00" }}`)
	f.Add(`{{ "\x01\x02\x03" }}`)
	f.Add(`{{ "line1\nline2\nline3" }}`)
	f.Add(`{{ "\t\t\t" }}`)

	// Long inputs
	f.Add(strings.Repeat("a", 1000))
	f.Add(strings.Repeat("{{ x }}", 100))
	f.Add("{{ " + strings.Repeat("a", 500) + " }}")
	f.Add(`{{ "` + strings.Repeat("x", 500) + `" }}`)

	// Array literals
	f.Add("{{ [] }}")
	f.Add("{{ [1] }}")
	f.Add("{{ [1, 2, 3] }}")
	f.Add("{{ [1, 2, 3,] }}")
	f.Add(`{{ ["a", "b", "c"] }}`)
	f.Add("{{ [1, [2, [3]]] }}")

	// Function calls
	f.Add("{{ func() }}")
	f.Add("{{ func(1) }}")
	f.Add("{{ func(1, 2) }}")
	f.Add("{{ func(1, 2, 3) }}")
	f.Add("{{ obj.method() }}")
	f.Add("{{ obj.method(arg) }}")
	f.Add("{{ a.b.c.d() }}")

	f.Fuzz(func(t *testing.T, input string) {
		// Lexer should never panic, only return errors
		tokens, err := lex("fuzz-test", input)
		if err != nil {
			// Errors are expected for malformed input
			return
		}
		// Verify tokens are valid
		for _, tok := range tokens {
			if tok == nil {
				t.Error("lexer returned nil token")
			}
		}
	})
}
