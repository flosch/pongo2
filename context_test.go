package pongo2

import "testing"

func TestSetAutoescape(t *testing.T) {
	original := DefaultSet.autoescape

	SetAutoescape(false)
	if DefaultSet.autoescape != false {
		t.Error("SetAutoescape(false) did not set autoescape to false")
	}

	SetAutoescape(true)
	if DefaultSet.autoescape != true {
		t.Error("SetAutoescape(true) did not set autoescape to true")
	}

	SetAutoescape(original)
}

func TestExecutionContextLogf(t *testing.T) {
	set := NewSet("test-logf", &DummyLoader{})
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

func TestSkipContextValidation(t *testing.T) {
	ctx := Context{
		"valid_key":    "rendered",
		"in-valid-key": "ignored", // not a valid identifier
	}

	set := NewSet("skip-validation", &DummyLoader{})
	tpl, err := set.FromString("{{ valid_key }}")
	if err != nil {
		t.Fatalf("FromString: %v", err)
	}

	// Default: validation enabled -> the invalid key triggers an error.
	if _, err := tpl.Execute(ctx); err == nil {
		t.Error("expected validation error for invalid context key, got nil")
	}

	// Opt-out: validation skipped -> renders without error.
	set.SkipContextValidation = true
	out, err := tpl.Execute(ctx)
	if err != nil {
		t.Fatalf("unexpected error with SkipContextValidation: %v", err)
	}
	if out != "rendered" {
		t.Errorf("got %q, want %q", out, "rendered")
	}
}
