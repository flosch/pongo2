package pongo2_test

import (
	"testing"
	"testing/fstest"

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

func TestIssue289(t *testing.T) {
	// Test negative integer in filter argument
	tpl, err := pongo2.FromString("{{ variable|add:-1 }}")
	if err != nil {
		t.Fatal(err)
	}
	str, err := tpl.Execute(pongo2.Context{"variable": 5})
	if err != nil {
		t.Fatal(err)
	}
	if str != "4" {
		t.Fatalf("Expected '4', but got '%s'.", str)
	}

	// Test negative float in filter argument
	tpl, err = pongo2.FromString("{{ variable|add:-1.5 }}")
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(pongo2.Context{"variable": 5.0})
	if err != nil {
		t.Fatal(err)
	}
	if str != "3.500000" {
		t.Fatalf("Expected '3.500000', but got '%s'.", str)
 	}
}

func TestIssue338(t *testing.T) {
	// Test that include with 'with' clause works inside a for loop
	// that iterates over a map with a single variable.
	// Bug: {% for key in map %} was setting Private[""] = value for maps,
	// which caused errors when include copied the context.
	// See: https://github.com/flosch/pongo2/issues/338

	memFS := fstest.MapFS{
		"main.tpl": &fstest.MapFile{
			Data: []byte(`{% for key in mymap %}{% include "field.html" with name=key id="test" %}{% endfor %}`),
		},
		"field.html": &fstest.MapFile{
			Data: []byte(`<label for="{{ id }}">{{ name }}</label>`),
		},
	}

	loader := pongo2.NewFSLoader(memFS)
	set := pongo2.NewSet("test", loader)

	tpl, err := set.FromFile("main.tpl")
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	ctx := pongo2.Context{
		"mymap": map[string]string{
			"a": "1",
			"b": "2",
		},
	}

	result, err := tpl.Execute(ctx)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	// Should render the label twice with keys from the map
	// Map iteration order is not guaranteed, so check for both possible outputs
	expected1 := `<label for="test">a</label><label for="test">b</label>`
	expected2 := `<label for="test">b</label><label for="test">a</label>`
	if result != expected1 && result != expected2 {
		t.Errorf("expected %q or %q, got %q", expected1, expected2, result)
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
