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

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", false},
		{"valid identifier", "foo", true},
		{"valid with underscore", "foo_bar", true},
		{"valid with numbers", "foo123", true},
		{"valid single char", "x", true},
		{"invalid with hyphen", "foo-bar", false},
		{"invalid with dot", "foo.bar", false},
		{"invalid with space", "foo bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("isValidIdentifier(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckForValidIdentifiersWithEmptyKey(t *testing.T) {
	ctx := Context{
		"":    "some value",
		"foo": "bar",
	}

	err := ctx.checkForValidIdentifiers()
	if err == nil {
		t.Error("expected error for empty key, got nil")
	}
}
