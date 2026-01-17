package pongo2_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flosch/pongo2/v7"
)

// FuzzSimpleExecution fuzzes template parsing and execution.
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
	f.Add("{{ foobar|title }}", "привіт світ")
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
