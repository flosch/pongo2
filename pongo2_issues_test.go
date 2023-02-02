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

type StructWithoutTag struct {
	Name string
	Age  int
}

type StructWithTag struct {
	Name string `pongo2:"my_name"`
	Age  int    `pongo2:"my_age"`
}

func TestIssue319StrutWithTag(t *testing.T) {
	s := StructWithTag{
		Name: "Andy Dufresne",
		Age:  23,
	}
	ctx := pongo2.Context{
		"mystruct": s,
	}
	template := `My name is {{ mystruct.my_name }}, age is {{ mystruct.my_age }}`
	out, err := pongo2.FromString(template)
	if err != nil {
		t.Error(err)
	}
	execute, err := out.Execute(ctx)
	if err != nil {
		t.Error(err)
	}
	mustEqual(t, execute, "My name is Andy Dufresne, age is 23")
}

func TestIssue319StrutWithoutTag(t *testing.T) {
	s := StructWithoutTag{
		Name: "Andy Dufresne",
		Age:  23,
	}
	ctx := pongo2.Context{
		"mystruct": s,
	}
	template := `My name is {{ mystruct.Name }}, age is {{ mystruct.Age }}`
	out, err := pongo2.FromString(template)
	if err != nil {
		t.Error(err)
	}
	execute, err := out.Execute(ctx)
	if err != nil {
		t.Error(err)
	}
	mustEqual(t, execute, "My name is Andy Dufresne, age is 23")
}
