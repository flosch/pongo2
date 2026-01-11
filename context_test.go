package pongo2

import "testing"

func TestSetAutoescape(t *testing.T) {
	original := autoescape

	SetAutoescape(false)
	if autoescape != false {
		t.Error("SetAutoescape(false) did not set autoescape to false")
	}

	SetAutoescape(true)
	if autoescape != true {
		t.Error("SetAutoescape(true) did not set autoescape to true")
	}

	SetAutoescape(original)
}

func TestExecutionContextLogf(t *testing.T) {
	loader := MustNewLocalFileSystemLoader("")
	set := NewSet("test-logf", loader)
	set.Debug = true

	tpl, err := set.FromString("test")
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	ctx := newExecutionContext(tpl, Context{})
	ctx.Logf("test message %s", "arg")
}
