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
