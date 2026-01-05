package pongo2_test

import (
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestIssue151(t *testing.T) {
	tpl, err := pongo2.FromString("{{ mydict.51232_3 }}{{ 12345_123}}{{ 995189baz }}")
	if err != nil {
		t.Fatal(err)
	}

	str, err := tpl.Execute(pongo2.Context{
		"mydict": map[string]string{
			"51232_3": "foo",
		},
		"12345_123": "bar",
		"995189baz": "baz",
	})
	if err != nil {
		t.Fatal(err)
	}

	if str != "foobarbaz" {
		t.Fatalf("Expected output 'foobarbaz', but got '%s'.", str)
	}
}

func TestIssue297(t *testing.T) {
	tpl, err := pongo2.FromString("Testing: {{ input|wordwrap:4 }}!")
	if err != nil {
		t.Fatal(err)
	}

	str, err := tpl.Execute(pongo2.Context{"input": "one two three four five six"})
	if err != nil {
		t.Fatal(err)
	}

	if str != "Testing: one two three four\nfive six!" {
		t.Fatalf("Expected `Testing: one two three four\nfive six!`, but got `%v`.", str)
	}
}

func TestIssue343(t *testing.T) {
	// Test \n escape sequence (newline) - the main issue
	tpl, err := pongo2.FromString(`{% for val in my_string|split:"\n" %}[{{ val }}]{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err := tpl.Execute(pongo2.Context{"my_string": "Line 1\nLine 2\nLine 3"})
	if err != nil {
		t.Fatal(err)
	}
	if str != "[Line 1][Line 2][Line 3]" {
		t.Fatalf(`Expected "[Line 1][Line 2][Line 3]", but got %q`, str)
	}

	// Test \t escape sequence (tab)
	tpl, err = pongo2.FromString(`{% for val in my_string|split:"\t" %}[{{ val }}]{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(pongo2.Context{"my_string": "A\tB\tC"})
	if err != nil {
		t.Fatal(err)
	}
	if str != "[A][B][C]" {
		t.Fatalf(`Expected "[A][B][C]", but got %q`, str)
	}

	// Test \r escape sequence (carriage return)
	tpl, err = pongo2.FromString(`{% for val in my_string|split:"\r" %}[{{ val }}]{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(pongo2.Context{"my_string": "X\rY\rZ"})
	if err != nil {
		t.Fatal(err)
	}
	if str != "[X][Y][Z]" {
		t.Fatalf(`Expected "[X][Y][Z]", but got %q`, str)
	}

	// Test \' escape sequence in single-quoted strings
	// Use |safe to prevent autoescape from converting ' to &#39;
	tpl, err = pongo2.FromString(`{{ 'it\'s working'|safe }}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if str != "it's working" {
		t.Fatalf(`Expected "it's working", but got %q`, str)
	}

	// Test \\n produces literal backslash+n, not newline
	tpl, err = pongo2.FromString(`{{ "hello\\nworld" }}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if str != "hello\\nworld" {
		t.Fatalf(`Expected "hello\\nworld" (literal backslash-n), but got %q`, str)
	}

	// Test existing escape sequences still work: \"
	// Use |safe to prevent autoescape from converting " to &quot;
	tpl, err = pongo2.FromString(`{{ "say \"hello\""|safe }}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if str != `say "hello"` {
		t.Fatalf(`Expected 'say "hello"', but got %q`, str)
	}

	// Test existing escape sequences still work: \\
	tpl, err = pongo2.FromString(`{{ "back\\slash" }}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if str != "back\\slash" {
		t.Fatalf(`Expected "back\\slash", but got %q`, str)
	}
}
