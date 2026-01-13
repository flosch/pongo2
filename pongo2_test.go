package pongo2_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/flosch/pongo2/v6"
)

var testSuite2 = pongo2.NewSet("test suite 2", pongo2.MustNewLocalFileSystemLoader(""))

func mustEqual(t *testing.T, s, pattern string) {
	if !regexp.MustCompile(pattern).MatchString(s) {
		t.Fatalf("mustEqual failed: '%v' does not match pattern '%v'", s, pattern)
	}
}

func mustPanicMatch(t *testing.T, fn func(), pattern string) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Expected panic with pattern '%v', nothing happened", pattern)
		}

		if !regexp.MustCompile(pattern).MatchString(fmt.Sprintf("%v", err)) {
			t.Fatalf("Expected panic with pattern '%v', but got '%v'", pattern, err)
		}
	}()

	// We expect fn to panic
	fn()
}

func parseTemplate(s string, c pongo2.Context) string {
	t, err := testSuite2.FromString(s)
	if err != nil {
		panic(err)
	}
	out, err := t.Execute(c)
	if err != nil {
		panic(err)
	}
	return out
}

func parseTemplateFn(s string, c pongo2.Context) func() {
	return func() {
		parseTemplate(s, c)
	}
}

func TestMisc(t *testing.T) {
	// Must
	// TODO: Add better error message (see issue #18)
	mustPanicMatch(
		t,
		func() { pongo2.Must(testSuite2.FromFile("template_tests/inheritance/base2.tpl")) },
		`\[Error \(where: fromfile\) in .*template_tests[/\\]inheritance[/\\]doesnotexist.tpl | Line 1 Col 12 near 'doesnotexist.tpl'\] open .*template_tests[/\\]inheritance[/\\]doesnotexist.tpl: no such file or directory`,
	)

	// Context
	mustPanicMatch(t, parseTemplateFn("", pongo2.Context{"'illegal": nil}), ".*not a valid identifier.*")

	// Registers
	mustEqual(t, pongo2.RegisterFilter("escape", nil).Error(), ".*is already registered")
	mustEqual(t, pongo2.RegisterTag("for", nil).Error(), ".*is already registered")

	// ApplyFilter
	v, err := pongo2.ApplyFilter("title", pongo2.AsValue("this is a title"), nil)
	if err != nil {
		t.Fatal(err)
	}
	mustEqual(t, v.String(), "This Is A Title")
	mustPanicMatch(t, func() {
		_, err := pongo2.ApplyFilter("doesnotexist", nil, nil)
		if err != nil {
			panic(err)
		}
	}, `\[Error \(where: applyfilter\)\] filter with name 'doesnotexist' not found`)
}

func TestImplicitExecCtx(t *testing.T) {
	tpl, err := pongo2.FromString("{{ ImplicitExec }}")
	if err != nil {
		t.Fatalf("Error in FromString: %v", err)
	}

	val := "a stringy thing"

	res, err := tpl.Execute(pongo2.Context{
		"Value": val,
		"ImplicitExec": func(ctx *pongo2.ExecutionContext) string {
			return ctx.Public["Value"].(string)
		},
	})
	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}

	mustEqual(t, res, val)

	// The implicit ctx should not be persisted from call-to-call
	res, err = tpl.Execute(pongo2.Context{
		"ImplicitExec": func() string {
			return val
		},
	})

	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}

	mustEqual(t, res, val)
}

// TestForloopRevcounter verifies that forloop.Revcounter and forloop.Revcounter0
// match Django's documented behavior:
// - Revcounter: iterations remaining (1-indexed, counts down to 1)
// - Revcounter0: iterations remaining (0-indexed, counts down to 0)
func TestForloopRevcounter(t *testing.T) {
	items := []int{10, 20, 30, 40, 50}

	tpl, err := pongo2.FromString(`{% for item in items %}{{ forloop.Counter }}:{{ forloop.Revcounter }},{{ forloop.Revcounter0 }} {% endfor %}`)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}

	result, err := tpl.Execute(pongo2.Context{"items": items})
	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}

	// For 5 items:
	// Counter=1: Revcounter=5, Revcounter0=4
	// Counter=2: Revcounter=4, Revcounter0=3
	// Counter=3: Revcounter=3, Revcounter0=2
	// Counter=4: Revcounter=2, Revcounter0=1
	// Counter=5: Revcounter=1, Revcounter0=0
	expected := "1:5,4 2:4,3 3:3,2 4:2,1 5:1,0 "
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestUrlizeFilter tests the urlize filter with various URL formats and TLDs
func TestUrlizeFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Existing TLDs that should work
		{
			name:     "http URL with .com",
			input:    "Visit http://example.com today",
			expected: `Visit <a href="http://example.com" rel="nofollow">http://example.com</a> today`,
		},
		{
			name:     "www URL with .org",
			input:    "Check www.example.org please",
			expected: `Check <a href="http://www.example.org" rel="nofollow">www.example.org</a> please`,
		},
		// New TLDs that should be supported
		{
			name:     "URL with .io TLD",
			input:    "Visit github.io for pages",
			expected: `Visit <a href="http://github.io" rel="nofollow">github.io</a> for pages`,
		},
		{
			name:     "URL with .dev TLD",
			input:    "Check web.dev for tips",
			expected: `Check <a href="http://web.dev" rel="nofollow">web.dev</a> for tips`,
		},
		{
			name:     "URL with .co TLD",
			input:    "See example.co now",
			expected: `See <a href="http://example.co" rel="nofollow">example.co</a> now`,
		},
		{
			name:     "URL with .ai TLD",
			input:    "Try claude.ai today",
			expected: `Try <a href="http://claude.ai" rel="nofollow">claude.ai</a> today`,
		},
		{
			name:     "URL with .app TLD",
			input:    "Download from myapp.app",
			expected: `Download from <a href="http://myapp.app" rel="nofollow">myapp.app</a>`,
		},
		// Country-code TLDs
		{
			name:     "URL with .uk TLD",
			input:    "Visit bbc.co.uk for news",
			expected: `Visit <a href="http://bbc.co.uk" rel="nofollow">bbc.co.uk</a> for news`,
		},
		{
			name:     "URL with .fr TLD",
			input:    "See example.fr please",
			expected: `See <a href="http://example.fr" rel="nofollow">example.fr</a> please`,
		},
		// Email addresses
		{
			name:     "email with .info TLD",
			input:    "Contact user@example.info",
			expected: `Contact <a href="mailto:user@example.info">user@example.info</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(`{{ input|urlize|safe }}`)
			if err != nil {
				t.Fatalf("Error parsing template: %v", err)
			}

			result, err := tpl.Execute(pongo2.Context{"input": tt.input})
			if err != nil {
				t.Fatalf("Error executing template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

// TestValueEqualTo tests the Value.EqualTo method for various types
func TestValueEqualTo(t *testing.T) {
	tests := []struct {
		name     string
		v1       interface{}
		v2       interface{}
		expected bool
	}{
		// Integers
		{"equal integers", 42, 42, true},
		{"different integers", 42, 43, false},
		{"int and int64 equal", 42, int64(42), true},

		// Floats
		{"equal floats", 3.14, 3.14, true},
		{"different floats", 3.14, 2.71, false},

		// Strings
		{"equal strings", "hello", "hello", true},
		{"different strings", "hello", "world", false},

		// Booleans
		{"equal bools true", true, true, true},
		{"equal bools false", false, false, true},
		{"different bools", true, false, false},

		// Nil values
		// Note: Current implementation returns false for two nil values
		// Using reflect.Value.Equal would return true for two invalid values
		{"both nil", nil, nil, false},
		{"one nil", nil, "hello", false},

		// Slices (not comparable, should return false)
		{"slices", []int{1, 2}, []int{1, 2}, false},

		// Maps (not comparable, should return false)
		{"maps", map[string]int{"a": 1}, map[string]int{"a": 1}, false},

		// Structs (comparable)
		{"equal structs", struct{ A int }{1}, struct{ A int }{1}, true},
		{"different structs", struct{ A int }{1}, struct{ A int }{2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1 := pongo2.AsValue(tt.v1)
			v2 := pongo2.AsValue(tt.v2)

			result := v1.EqualValueTo(v2)
			if result != tt.expected {
				t.Errorf("EqualTo(%v, %v) = %v, want %v", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

// BenchmarkValueEqualTo benchmarks the EqualValueTo method
func BenchmarkValueEqualTo(b *testing.B) {
	// Pre-create values to avoid allocation overhead in benchmark
	intVal1 := pongo2.AsValue(42)
	intVal2 := pongo2.AsValue(42)
	strVal1 := pongo2.AsValue("hello world")
	strVal2 := pongo2.AsValue("hello world")
	structVal1 := pongo2.AsValue(struct{ A, B, C int }{1, 2, 3})
	structVal2 := pongo2.AsValue(struct{ A, B, C int }{1, 2, 3})
	sliceVal1 := pongo2.AsValue([]int{1, 2, 3})
	sliceVal2 := pongo2.AsValue([]int{1, 2, 3})

	b.Run("integers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			intVal1.EqualValueTo(intVal2)
		}
	})

	b.Run("strings", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			strVal1.EqualValueTo(strVal2)
		}
	})

	b.Run("structs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			structVal1.EqualValueTo(structVal2)
		}
	})

	b.Run("slices_not_comparable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sliceVal1.EqualValueTo(sliceVal2)
		}
	})
}

type DummyLoader struct{}

func (l *DummyLoader) Abs(base, name string) string {
	return filepath.Join(filepath.Dir(base), name)
}

func (l *DummyLoader) Get(path string) (io.Reader, error) {
	return nil, errors.New("dummy not found")
}

func FuzzSimpleExecution(f *testing.F) {
	tpls, err := filepath.Glob("template_tests/*.tpl")
	if err != nil {
		f.Fatalf("glob: %v", err)
	}
	files := []string{"README.md"}
	files = append(files, tpls...)

	for _, tplPath := range files {
		buf, err := os.ReadFile(tplPath)
		if err != nil {
			f.Fatalf("could not read file '%v': %v", tplPath, err)
		}
		f.Add(string(buf), "test-value")
	}

	f.Add("{{ foobar }}", "00000000")

	// Edge case: empty and whitespace templates
	f.Add("", "")
	f.Add("   ", "")
	f.Add("\n\n\n", "")
	f.Add("\t\t", "")
	f.Add(" \n \t ", "test")

	// Basic variable access patterns
	f.Add("{{ foobar }}", "test")
	f.Add("{{ foobar.attr }}", "test")
	f.Add("{{ foobar.0 }}", "test")
	f.Add("{{ foobar[0] }}", "test")
	f.Add("{{ foobar['key'] }}", "test")
	f.Add(`{{ foobar["key"] }}`, "test")
	f.Add("{{ foobar.nested.deep.value }}", "test")
	f.Add("{{ foobar[foobar] }}", "test")

	// Expression edge cases
	f.Add("{{ 0 }}", "")
	f.Add("{{ -0 }}", "")
	f.Add("{{ -1 }}", "")
	f.Add("{{ 999999999999999999 }}", "")
	f.Add("{{ -999999999999999999 }}", "")
	f.Add("{{ 0.0 }}", "")
	f.Add("{{ 0.000001 }}", "")
	f.Add("{{ 1e10 }}", "")
	f.Add("{{ 1.5e-10 }}", "")
	f.Add("{{ true }}", "")
	f.Add("{{ false }}", "")
	f.Add("{{ nil }}", "")
	f.Add("{{ nothing }}", "")

	// String literals
	f.Add(`{{ "" }}`, "")
	f.Add(`{{ '' }}`, "")
	f.Add(`{{ "hello" }}`, "")
	f.Add(`{{ 'hello' }}`, "")
	f.Add(`{{ "hello \"world\"" }}`, "")
	f.Add(`{{ 'hello \'world\'' }}`, "")
	f.Add(`{{ "line1\nline2" }}`, "")
	f.Add(`{{ "\t\r\n" }}`, "")

	// Arithmetic expressions
	f.Add("{{ 1 + 1 }}", "")
	f.Add("{{ 10 - 5 }}", "")
	f.Add("{{ 3 * 4 }}", "")
	f.Add("{{ 10 / 2 }}", "")
	f.Add("{{ 10 / 0 }}", "")
	f.Add("{{ 10 % 3 }}", "")
	f.Add("{{ 10 % 0 }}", "")
	f.Add("{{ 2 ^ 10 }}", "")
	f.Add("{{ 2 ^ 0 }}", "")
	f.Add("{{ -(10-100) }}", "")
	f.Add("{{ -1 * (-(-(10-100))) }}", "")
	f.Add("{{ -1 * (-(-(10-100)) ^ 2) ^ 3 + 3 * (5 - 17) + 1 + 2 }}", "")

	// Comparison operators
	f.Add("{{ 1 == 1 }}", "")
	f.Add("{{ 1 != 1 }}", "")
	f.Add("{{ 1 < 2 }}", "")
	f.Add("{{ 1 > 2 }}", "")
	f.Add("{{ 1 <= 1 }}", "")
	f.Add("{{ 1 >= 1 }}", "")
	f.Add(`{{ "a" == "a" }}`, "")
	f.Add(`{{ "a" < "b" }}`, "")

	// Logical operators
	f.Add("{{ true && true }}", "")
	f.Add("{{ true || false }}", "")
	f.Add("{{ !true }}", "")
	f.Add("{{ not true }}", "")
	f.Add("{{ true and false }}", "")
	f.Add("{{ true or false }}", "")
	f.Add("{{ !(true || false) }}", "")
	f.Add("{{ true && (true && (true && (true && (1 == 1 || false)))) }}", "")

	// In operator
	f.Add("{{ 1 in foobar }}", "test")
	f.Add("{{ 'x' in foobar }}", "test")
	f.Add("{{ !(5 in foobar) }}", "test")
	f.Add("{{ not(7 in foobar) }}", "test")

	// String concatenation
	f.Add(`{{ "a" + "b" }}`, "")
	f.Add(`{{ 1 + "a" }}`, "")
	f.Add(`{{ "a" + 1 }}`, "")

	// Filter chains
	f.Add("{{ foobar|safe }}", "<script>alert(1)</script>")
	f.Add("{{ foobar|escape }}", "<script>alert(1)</script>")
	f.Add("{{ foobar|upper }}", "test")
	f.Add("{{ foobar|lower }}", "TEST")
	f.Add("{{ foobar|title }}", "test")
	f.Add("{{ foobar|length }}", "test")
	f.Add("{{ foobar|default:'fallback' }}", "")
	f.Add("{{ foobar|default:foobar }}", "test")
	f.Add("{{ foobar|upper|lower|title }}", "test")
	f.Add("{{ foobar|add:1 }}", "5")
	f.Add("{{ foobar|add:'suffix' }}", "prefix")
	f.Add("{{ foobar|truncatechars:5 }}", "hello world")
	f.Add("{{ foobar|slice:':5' }}", "hello world")
	f.Add("{{ foobar|slice:'5:' }}", "hello world")
	f.Add("{{ foobar|slice:'-1:' }}", "hello world")
	f.Add("{{ foobar|slice:':-1' }}", "hello world")
	f.Add("{{ foobar|join:', ' }}", "test")
	f.Add("{{ foobar|split:' ' }}", "hello world")
	f.Add("{{ foobar|first }}", "test")
	f.Add("{{ foobar|last }}", "test")
	f.Add("{{ foobar|random }}", "test")
	f.Add("{{ foobar|striptags }}", "<b>test</b>")
	f.Add("{{ foobar|removetags:'script' }}", "<script>test</script>")
	f.Add("{{ foobar|urlencode }}", "http://example.com?a=1&b=2")
	f.Add("{{ foobar|floatformat }}", "3.14159")
	f.Add("{{ foobar|floatformat:2 }}", "3.14159")
	f.Add("{{ foobar|pluralize }}", "1")
	f.Add("{{ foobar|pluralize:'y,ies' }}", "2")
	f.Add("{{ foobar|yesno }}", "true")
	f.Add("{{ foobar|yesno:'ja,nein,vielleicht' }}", "")
	f.Add("{{ foobar|wordcount }}", "hello world test")
	f.Add("{{ foobar|wordwrap:5 }}", "hello world test")
	f.Add("{{ foobar|linebreaks }}", "hello\nworld")
	f.Add("{{ foobar|linebreaksbr }}", "hello\nworld")
	f.Add("{{ foobar|slugify }}", "Hello World!")
	f.Add("{{ foobar|filesizeformat }}", "1048576")
	f.Add("{{ foobar|center:20 }}", "test")
	f.Add("{{ foobar|ljust:20 }}", "test")
	f.Add("{{ foobar|rjust:20 }}", "test")
	f.Add("{{ foobar|cut:' ' }}", "hello world")
	f.Add("{{ foobar|capfirst }}", "hello")
	f.Add("{{ foobar|make_list }}", "hello")
	f.Add("{{ foobar|divisibleby:3 }}", "9")
	f.Add("{{ foobar|get_digit:2 }}", "12345")
	f.Add("{{ foobar|integer }}", "3.14")
	f.Add("{{ foobar|float }}", "3")
	f.Add("{{ foobar|phone2numeric }}", "1-800-PONGO")
	f.Add("{{ foobar|addslashes }}", "it's a test")
	f.Add("{{ foobar|escapejs }}", "line1\nline2")
	f.Add("{{ foobar|linenumbers }}", "line1\nline2\nline3")
	f.Add("{{ foobar|iriencode }}", "?foo=1&bar=2")
	f.Add("{{ foobar|json_script:'data-id' }}", "test")
	f.Add("{{ foobar|stringformat:'%05d' }}", "42")

	// If/else tags
	f.Add("{% if true %}yes{% endif %}", "")
	f.Add("{% if false %}no{% else %}yes{% endif %}", "")
	f.Add("{% if false %}no{% elif true %}yes{% else %}no{% endif %}", "")
	f.Add("{% if foobar %}yes{% endif %}", "test")
	f.Add("{% if !foobar %}yes{% endif %}", "")
	f.Add("{% if not foobar %}yes{% endif %}", "")
	f.Add("{% if foobar == 'test' %}yes{% endif %}", "test")
	f.Add("{% if foobar != 'test' %}no{% else %}yes{% endif %}", "test")
	f.Add("{% if foobar and true %}yes{% endif %}", "test")
	f.Add("{% if foobar or false %}yes{% endif %}", "test")
	f.Add("{% if 1 < 2 and 2 < 3 %}yes{% endif %}", "")
	f.Add("{% if nothing %}no{% else %}yes{% endif %}", "")

	// For loops
	f.Add("{% for i in foobar %}{{ i }}{% endfor %}", "test")
	f.Add("{% for i in foobar %}{{ forloop.Counter }}{% endfor %}", "test")
	f.Add("{% for i in foobar %}{{ forloop.Counter0 }}{% endfor %}", "test")
	f.Add("{% for i in foobar %}{{ forloop.First }}{% endfor %}", "test")
	f.Add("{% for i in foobar %}{{ forloop.Last }}{% endfor %}", "test")
	f.Add("{% for i in foobar %}{{ forloop.Revcounter }}{% endfor %}", "test")
	f.Add("{% for i in foobar %}{{ forloop.Revcounter0 }}{% endfor %}", "test")
	f.Add("{% for i in foobar reversed %}{{ i }}{% endfor %}", "test")
	f.Add("{% for i in foobar sorted %}{{ i }}{% endfor %}", "test")
	f.Add("{% for i in foobar reversed sorted %}{{ i }}{% endfor %}", "test")
	f.Add("{% for i in foobar %}{% for j in i %}{{ j }}{% endfor %}{% endfor %}", "test")
	f.Add("{% for i in foobar %}{% empty %}empty{% endfor %}", "")

	// Set tag
	f.Add("{% set x = 1 %}{{ x }}", "")
	f.Add("{% set x = foobar %}{{ x }}", "test")
	f.Add("{% set x = foobar|upper %}{{ x }}", "test")
	f.Add("{% set x, y = 1, 2 %}{{ x }}-{{ y }}", "")

	// With tag
	f.Add("{% with x=1 %}{{ x }}{% endwith %}", "")
	f.Add("{% with x=foobar %}{{ x }}{% endwith %}", "test")
	f.Add("{% with x=1 y=2 %}{{ x }}-{{ y }}{% endwith %}", "")

	// Block/extends (will fail but shouldn't crash)
	f.Add("{% block content %}test{% endblock %}", "")
	f.Add("{% block content %}{{ foobar }}{% endblock %}", "test")

	// Comment tags
	f.Add("{# this is a comment #}", "")
	f.Add("{# {{ foobar }} #}", "test")
	f.Add("{# {% if true %}yes{% endif %} #}", "")
	f.Add("before{# comment #}after", "")

	// Verbatim tag
	f.Add("{% verbatim %}{{ foobar }}{% endverbatim %}", "test")
	f.Add("{% verbatim %}{% if true %}test{% endif %}{% endverbatim %}", "")

	// Spaceless tag
	f.Add("{% spaceless %}<p>  text  </p>{% endspaceless %}", "")
	f.Add("{% spaceless %}\n  <p>\n    text\n  </p>\n{% endspaceless %}", "")

	// Autoescape tag
	f.Add("{% autoescape on %}{{ foobar }}{% endautoescape %}", "<script>")
	f.Add("{% autoescape off %}{{ foobar }}{% endautoescape %}", "<script>")

	// Firstof tag
	f.Add("{% firstof nothing foobar 'default' %}", "test")
	f.Add("{% firstof nothing nothing2 'default' %}", "")
	f.Add("{% firstof foobar 'unused' %}", "test")

	// Templatetag
	f.Add("{% templatetag openblock %}", "")
	f.Add("{% templatetag closeblock %}", "")
	f.Add("{% templatetag openvariable %}", "")
	f.Add("{% templatetag closevariable %}", "")
	f.Add("{% templatetag openbrace %}", "")
	f.Add("{% templatetag closebrace %}", "")
	f.Add("{% templatetag opencomment %}", "")
	f.Add("{% templatetag closecomment %}", "")

	// Widthratio tag
	f.Add("{% widthratio 5 10 100 %}", "")
	f.Add("{% widthratio foobar 100 200 %}", "50")
	f.Add("{% widthratio 0 10 100 %}", "")
	f.Add("{% widthratio 10 0 100 %}", "")

	// Cycle tag
	f.Add("{% cycle 'a' 'b' 'c' %}", "")
	f.Add("{% for i in foobar %}{% cycle 'odd' 'even' %}{% endfor %}", "test")

	// Ifchanged tag
	f.Add("{% ifchanged %}{{ foobar }}{% endifchanged %}", "test")
	f.Add("{% for i in foobar %}{% ifchanged %}{{ i }}{% endifchanged %}{% endfor %}", "aabbcc")

	// Filter tag
	f.Add("{% filter upper %}hello{% endfilter %}", "")
	f.Add("{% filter upper|title %}hello world{% endfilter %}", "")
	f.Add("{% filter escape %}{{ foobar }}{% endfilter %}", "<script>")

	// Lorem tag
	f.Add("{% lorem %}", "")
	f.Add("{% lorem 5 %}", "")
	f.Add("{% lorem 5 w %}", "")
	f.Add("{% lorem 5 p %}", "")
	f.Add("{% lorem 5 b %}", "")
	f.Add("{% lorem 5 w random %}", "")

	// Now tag
	f.Add("{% now 'Y-m-d' %}", "")
	f.Add("{% now 'H:i:s' %}", "")

	// Macro
	f.Add("{% macro test() %}hello{% endmacro %}{{ test() }}", "")
	f.Add("{% macro test(name) %}hello {{ name }}{% endmacro %}{{ test('world') }}", "")
	f.Add("{% macro test(name='guest') %}hello {{ name }}{% endmacro %}{{ test() }}", "")
	f.Add("{% macro test(a, b, c) %}{{ a }}-{{ b }}-{{ c }}{% endmacro %}{{ test(1, 2, 3) }}", "")

	// Unicode templates
	f.Add("{{ foobar }}", "你好世界")
	f.Add("你好 {{ foobar }} 世界", "test")
	f.Add("{{ foobar|length }}", "你好世界")
	f.Add("{{ foobar|first }}", "你好世界")
	f.Add("{{ foobar|last }}", "你好世界")
	f.Add("{{ foobar|upper }}", "Ελληνικά")
	f.Add("{{ foobar|lower }}", "Ελληνικά")
	f.Add("{{ foobar|title }}", "привет мир")
	f.Add("{% for c in foobar %}{{ c }}{% endfor %}", "日本語")
	f.Add("{{ foobar|slice:'0:2' }}", "日本語テスト")

	// Malformed/edge case templates
	f.Add("{{", "")
	f.Add("}}", "")
	f.Add("{%", "")
	f.Add("%}", "")
	f.Add("{#", "")
	f.Add("#}", "")
	f.Add("{{ }", "")
	f.Add("{% %}", "")
	f.Add("{# #}", "")
	f.Add("{{  }}", "")
	f.Add("{%%}", "")
	f.Add("{{ foobar", "test")
	f.Add("foobar }}", "test")
	f.Add("{% if true %}", "")
	f.Add("{% endif %}", "")
	f.Add("{% for i in x %}", "")
	f.Add("{% endfor %}", "")
	f.Add("{% if true %}{% if true %}{% endif %}", "")
	f.Add("{{ foobar|unknown_filter }}", "test")
	f.Add("{% unknown_tag %}", "")
	f.Add("{{ . }}", "")
	f.Add("{{ .. }}", "")
	f.Add("{{ ... }}", "")
	f.Add("{{ foobar.....attr }}", "test")
	f.Add("{{ foobar[[[0]]] }}", "test")
	f.Add("{{ foobar|add: }}", "test")
	f.Add("{{ foobar|:arg }}", "test")
	f.Add("{{ | }}", "")
	f.Add("{{ foobar|| }}", "test")
	f.Add("{{ foobar|filter| }}", "test")

	// Deeply nested structures
	f.Add("{% if true %}{% if true %}{% if true %}{% if true %}nested{% endif %}{% endif %}{% endif %}{% endif %}", "")
	f.Add("{% for i in foobar %}{% for j in i %}{% for k in j %}{{ k }}{% endfor %}{% endfor %}{% endfor %}", "test")
	f.Add("{{ foobar|upper|lower|upper|lower|upper|lower|title }}", "test")
	f.Add("{{ ((((((1 + 2) * 3) - 4) / 5) % 6) ^ 7) }}", "")

	// Whitespace control
	f.Add("{%- if true -%}yes{%- endif -%}", "")
	f.Add("{{- foobar -}}", "test")
	f.Add("{%- for i in foobar -%}{{ i }}{%- endfor -%}", "test")
	f.Add("  {%- if true -%}  yes  {%- endif -%}  ", "")

	// Special characters in strings
	f.Add(`{{ "hello\x00world" }}`, "")
	f.Add(`{{ "line1\nline2\ttab" }}`, "")
	f.Add(`{{ "\\" }}`, "")
	f.Add(`{{ "\"" }}`, "")
	f.Add(`{{ "\'" }}`, "")

	// XSS-like patterns
	f.Add("{{ foobar }}", "<script>alert('xss')</script>")
	f.Add("{{ foobar }}", "<img src=x onerror=alert(1)>")
	f.Add("{{ foobar }}", "javascript:alert(1)")
	f.Add("{{ foobar }}", "<svg/onload=alert(1)>")
	f.Add("{{ foobar }}", "{{constructor.constructor('alert(1)')()}}")
	f.Add("{{ foobar|safe }}", "<script>alert('xss')</script>")

	// Large inputs
	f.Add("{{ foobar }}", string(make([]byte, 10000)))
	f.Add(string(make([]byte, 1000)), "test")

	f.Fuzz(func(t *testing.T, tpl, contextValue string) {
		ts := pongo2.NewSet("fuzz-test", &DummyLoader{})
		out, err := ts.FromString(tpl)
		if err != nil && out != nil {
			t.Errorf("%v", err)
		}
		if err == nil {
			mycontext := pongo2.Context{
				"foobar": contextValue,
			}
			mycontext.Update(tplContext)
			out.Execute(mycontext) //nolint:errcheck
		}
	})
}
