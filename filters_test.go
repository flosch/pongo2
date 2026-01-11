package pongo2

import (
	"strings"
	"testing"
)

func TestReplaceFilter(t *testing.T) {
	t.Run("non-existent filter", func(t *testing.T) {
		err := ReplaceFilter("nonexistent_filter_xyz", func(in *Value, param *Value) (*Value, error) {
			return in, nil
		})
		if err == nil {
			t.Error("ReplaceFilter should return error for non-existent filter")
		}
	})

	t.Run("existing filter", func(t *testing.T) {
		originalFn := filters["upper"]
		defer func() { filters["upper"] = originalFn }()

		newFn := func(in *Value, param *Value) (*Value, error) {
			return AsValue("REPLACED"), nil
		}

		err := ReplaceFilter("upper", newFn)
		if err != nil {
			t.Errorf("ReplaceFilter failed for existing filter: %v", err)
		}

		result, err := ApplyFilter("upper", AsValue("test"), nil)
		if err != nil {
			t.Fatalf("ApplyFilter failed: %v", err)
		}
		if result.String() != "REPLACED" {
			t.Errorf("Filter was not replaced correctly, got %s", result.String())
		}
	})
}

func TestMustApplyFilter(t *testing.T) {
	t.Run("successful apply", func(t *testing.T) {
		result := MustApplyFilter("upper", AsValue("hello"), nil)
		if result.String() != "HELLO" {
			t.Errorf("MustApplyFilter returned %s, want HELLO", result.String())
		}
	})

	t.Run("panic on non-existent filter", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustApplyFilter should panic for non-existent filter")
			}
		}()
		MustApplyFilter("nonexistent_filter_xyz", AsValue("test"), nil)
	})
}

func TestFilterUrlize(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		contains string
	}{
		{
			name:     "url in text",
			template: "{{ text|urlize }}",
			context:  Context{"text": "Check out https://example.com for more info"},
			contains: "href=",
		},
		{
			name:     "email in text",
			template: "{{ text|urlize }}",
			context:  Context{"text": "Contact us at test@example.com"},
			contains: "href=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := FromString(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if !strings.Contains(result, tt.contains) {
				t.Errorf("Result should contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestFilterUrlizetrunc(t *testing.T) {
	tpl, err := FromString("{{ text|urlizetrunc:20 }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tpl.Execute(Context{"text": "Visit https://example.com/very/long/path/here"})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if !strings.Contains(result, "href=") {
		t.Error("Result should contain href")
	}
}

func TestFilterPluralize(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "singular",
			template: "{{ count }} item{{ count|pluralize }}",
			context:  Context{"count": 1},
			expected: "1 item",
		},
		{
			name:     "plural",
			template: "{{ count }} item{{ count|pluralize }}",
			context:  Context{"count": 5},
			expected: "5 items",
		},
		{
			name:     "custom suffix",
			template: "{{ count }} cherr{{ count|pluralize:\"y,ies\" }}",
			context:  Context{"count": 2},
			expected: "2 cherries",
		},
		{
			name:     "zero",
			template: "{{ count }} item{{ count|pluralize }}",
			context:  Context{"count": 0},
			expected: "0 items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := FromString(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFilterCenter(t *testing.T) {
	tpl, err := FromString("{{ text|center:10 }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tpl.Execute(Context{"text": "hi"})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if len(result) != 10 {
		t.Errorf("Result length = %d, want 10", len(result))
	}
}

func TestFilterLjust(t *testing.T) {
	tpl, err := FromString("{{ text|ljust:10 }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tpl.Execute(Context{"text": "hi"})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if len(result) != 10 {
		t.Errorf("Result length = %d, want 10", len(result))
	}
	if !strings.HasPrefix(result, "hi") {
		t.Errorf("Result should start with 'hi', got %q", result)
	}
}

func TestFilterRjust(t *testing.T) {
	tpl, err := FromString("{{ text|rjust:10 }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tpl.Execute(Context{"text": "hi"})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if len(result) != 10 {
		t.Errorf("Result length = %d, want 10", len(result))
	}
	if !strings.HasSuffix(result, "hi") {
		t.Errorf("Result should end with 'hi', got %q", result)
	}
}

func TestFilterYesno(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "true value",
			template: "{{ value|yesno }}",
			context:  Context{"value": true},
			expected: "yes",
		},
		{
			name:     "false value",
			template: "{{ value|yesno }}",
			context:  Context{"value": false},
			expected: "no",
		},
		{
			name:     "custom mapping",
			template: "{{ value|yesno:\"yeah,nope,maybe\" }}",
			context:  Context{"value": true},
			expected: "yeah",
		},
		{
			name:     "nil value with maybe",
			template: "{{ value|yesno:\"yeah,nope,maybe\" }}",
			context:  Context{"value": nil},
			expected: "maybe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := FromString(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFilterRemovetags(t *testing.T) {
	tpl, err := FromString("{{ html|removetags:\"b,i\" }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := tpl.Execute(Context{"html": "<b>bold</b> and <i>italic</i>"})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if strings.Contains(result, "<b>") || strings.Contains(result, "<i>") {
		t.Errorf("Tags should be removed, got %q", result)
	}
	expected := "bold and italic"
	if result != expected {
		t.Errorf("Got %q, want %q", result, expected)
	}
}

func TestFilterFloatformat(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "default precision",
			template: "{{ value|floatformat }}",
			context:  Context{"value": 3.14159},
			expected: "3.1",
		},
		{
			name:     "custom precision",
			template: "{{ value|floatformat:3 }}",
			context:  Context{"value": 3.14159},
			expected: "3.142",
		},
		{
			name:     "integer input",
			template: "{{ value|floatformat:2 }}",
			context:  Context{"value": 5},
			expected: "5.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := FromString(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFilterDate(t *testing.T) {
	tpl, err := FromString("{{ mydate|date:\"Y-m-d\" }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Test that it doesn't panic with non-time input
	_, _ = tpl.Execute(Context{"mydate": "2024-01-15"})
}
