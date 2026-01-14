package pongo2

import (
	"reflect"
	"strings"
	"testing"
)

// TestResolveArrayDefinition tests the resolveArrayDefinition helper.
func TestResolveArrayDefinition(t *testing.T) {
	// Test array definitions work correctly with filters
	// Arrays in pongo2 return []*Value which is used internally
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "array length",
			template: "{{ [1, 2, 3]|length }}",
			context:  nil,
			expected: "3",
		},
		{
			name:     "array first",
			template: "{{ [1, 2, 3]|first }}",
			context:  nil,
			expected: "1",
		},
		{
			name:     "array last",
			template: "{{ [1, 2, 3]|last }}",
			context:  nil,
			expected: "3",
		},
		{
			name:     "array with variables length",
			template: "{{ [x, y, z]|length }}",
			context:  Context{"x": 1, "y": 2, "z": 3},
			expected: "3",
		},
		{
			name:     "empty array length",
			template: "{{ []|length }}",
			context:  nil,
			expected: "0",
		},
		{
			name:     "array in for loop",
			template: "{% for i in [1, 2, 3] %}{{ i }}{% endfor %}",
			context:  nil,
			expected: "123",
		},
	}

	set := NewSet("test", MustNewLocalFileSystemLoader(""))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := set.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestLookupInitialValue tests the lookupInitialValue helper.
func TestLookupInitialValue(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "lookup from public context",
			template: "{{ foo }}",
			context:  Context{"foo": "bar"},
			expected: "bar",
		},
		{
			name:     "lookup nonexistent returns empty",
			template: "{{ nonexistent }}",
			context:  Context{},
			expected: "",
		},
		{
			name:     "lookup nil value",
			template: "{{ nilval }}",
			context:  Context{"nilval": nil},
			expected: "",
		},
	}

	set := NewSet("test", MustNewLocalFileSystemLoader(""))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := set.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestResolveNextPart tests the resolveNextPart helper via templates.
func TestResolveNextPart(t *testing.T) {
	type Nested struct {
		Value string
	}
	type TestStruct struct {
		Name   string
		Nested Nested
	}

	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "field access on struct",
			template: "{{ s.Name }}",
			context:  Context{"s": TestStruct{Name: "hello"}},
			expected: "hello",
		},
		{
			name:     "nested field access",
			template: "{{ s.Nested.Value }}",
			context:  Context{"s": TestStruct{Nested: Nested{Value: "nested"}}},
			expected: "nested",
		},
		{
			name:     "map key access",
			template: "{{ m.key }}",
			context:  Context{"m": map[string]int{"key": 42}},
			expected: "42",
		},
		{
			name:     "pointer dereference",
			template: "{{ p.Name }}",
			context:  Context{"p": &TestStruct{Name: "pointer"}},
			expected: "pointer",
		},
		{
			name:     "nil pointer returns empty",
			template: "{{ p.Name }}",
			context:  Context{"p": (*TestStruct)(nil)},
			expected: "",
		},
	}

	set := NewSet("test", MustNewLocalFileSystemLoader(""))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := set.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestResolveIntIndex tests the resolveIntIndex helper via templates.
func TestResolveIntIndex(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		context     Context
		expected    string
		expectError bool
	}{
		{
			name:     "array index",
			template: "{{ items.0 }}",
			context:  Context{"items": []string{"first", "second"}},
			expected: "first",
		},
		{
			name:     "array last index",
			template: "{{ items.1 }}",
			context:  Context{"items": []string{"first", "second"}},
			expected: "second",
		},
		{
			name:     "string index",
			template: "{{ s.0 }}",
			context:  Context{"s": "hello"},
			expected: "104", // ASCII code for 'h'
		},
		{
			name:     "out of bounds returns empty",
			template: "{{ items.10 }}",
			context:  Context{"items": []string{"only"}},
			expected: "",
		},
	}

	set := NewSet("test", MustNewLocalFileSystemLoader(""))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := set.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("failed to execute template: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

// TestResolveIdentifier tests the resolveIdentifier helper via templates.
func TestResolveIdentifier(t *testing.T) {
	type TestStruct struct {
		Public  string
		private string //nolint:unused // testing unexported field behavior
	}

	tests := []struct {
		name        string
		template    string
		context     Context
		expected    string
		expectError bool
	}{
		{
			name:     "struct field",
			template: "{{ s.Public }}",
			context:  Context{"s": TestStruct{Public: "visible"}},
			expected: "visible",
		},
		{
			name:     "map string key",
			template: "{{ m.foo }}",
			context:  Context{"m": map[string]string{"foo": "bar"}},
			expected: "bar",
		},
		{
			name:     "nonexistent field returns empty",
			template: "{{ s.NonExistent }}",
			context:  Context{"s": TestStruct{Public: "visible"}},
			expected: "",
		},
		{
			name:     "nonexistent map key returns empty",
			template: "{{ m.missing }}",
			context:  Context{"m": map[string]string{}},
			expected: "",
		},
	}

	set := NewSet("test", MustNewLocalFileSystemLoader(""))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := set.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("failed to execute template: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

// TestResolveSubscript tests the resolveSubscript helper via templates.
func TestResolveSubscript(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		context     Context
		expected    string
		expectError bool
	}{
		{
			name:     "array subscript with literal",
			template: "{{ items[0] }}",
			context:  Context{"items": []string{"first", "second"}},
			expected: "first",
		},
		{
			name:     "array subscript with variable",
			template: "{{ items[idx] }}",
			context:  Context{"items": []string{"first", "second"}, "idx": 1},
			expected: "second",
		},
		{
			name:     "map subscript with string",
			template: `{{ m["key"] }}`,
			context:  Context{"m": map[string]int{"key": 42}},
			expected: "42",
		},
		{
			name:     "map subscript with variable",
			template: "{{ m[k] }}",
			context:  Context{"m": map[string]int{"foo": 1}, "k": "foo"},
			expected: "1",
		},
		{
			name:     "map subscript with int key",
			template: "{{ m[1] }}",
			context:  Context{"m": map[int]string{1: "one"}},
			expected: "one",
		},
		{
			name:     "struct subscript with string",
			template: `{{ s["Name"] }}`,
			context:  Context{"s": struct{ Name string }{Name: "test"}},
			expected: "test",
		},
		{
			name:     "out of bounds returns empty",
			template: "{{ items[100] }}",
			context:  Context{"items": []string{"only"}},
			expected: "",
		},
		{
			name:     "nil subscript on map returns empty",
			template: "{{ m[nil] }}",
			context:  Context{"m": map[string]int{"a": 1}},
			expected: "",
		},
	}

	set := NewSet("test", MustNewLocalFileSystemLoader(""))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := set.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("failed to execute template: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

// TestHandleFunctionCall tests the handleFunctionCall helper via templates.
func TestHandleFunctionCall(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		context     Context
		expected    string
		expectError bool
	}{
		{
			name:     "simple function call",
			template: "{{ fn() }}",
			context:  Context{"fn": func() string { return "called" }},
			expected: "called",
		},
		{
			name:     "function with argument",
			template: "{{ add(1, 2) }}",
			context:  Context{"add": func(a, b int) int { return a + b }},
			expected: "3",
		},
		{
			name:     "function with Value argument",
			template: "{{ fn(x) }}",
			context: Context{
				"fn": func(v *Value) string { return v.String() },
				"x":  "hello",
			},
			expected: "hello",
		},
		{
			name:     "function returning Value",
			template: "{{ fn() }}",
			context: Context{
				"fn": func() *Value { return AsSafeValue("safe") },
			},
			expected: "safe",
		},
		{
			name:     "function with error return",
			template: "{{ fn() }}",
			context: Context{
				"fn": func() (string, error) { return "result", nil },
			},
			expected: "result",
		},
		{
			name:     "variadic function",
			template: "{{ sum(1, 2, 3, 4) }}",
			context: Context{
				"sum": func(nums ...int) int {
					total := 0
					for _, n := range nums {
						total += n
					}
					return total
				},
			},
			expected: "10",
		},
		{
			name:     "method call with length filter",
			template: "{{ items|length }}",
			context:  Context{"items": []int{1, 2, 3, 4, 5}},
			expected: "5",
		},
		{
			name:        "call non-function returns error",
			template:    "{{ notfunc() }}",
			context:     Context{"notfunc": "string"},
			expectError: true,
		},
	}

	set := NewSet("test", MustNewLocalFileSystemLoader(""))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := set.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("failed to execute template: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

// TestUnpackValue tests the unpackValue helper directly.
func TestUnpackValue(t *testing.T) {
	vr := &variableResolver{}

	t.Run("non-Value passthrough", func(t *testing.T) {
		input := reflect.ValueOf(42)
		result, isSafe := vr.unpackValue(input, false)
		if !reflect.DeepEqual(input.Interface(), result.Interface()) {
			t.Errorf("expected %v, got %v", input, result)
		}
		if isSafe {
			t.Error("expected isSafe to be false")
		}
	})

	t.Run("preserves isSafe when not unpacking", func(t *testing.T) {
		input := reflect.ValueOf("string")
		result, isSafe := vr.unpackValue(input, true)
		if !reflect.DeepEqual(input.Interface(), result.Interface()) {
			t.Errorf("expected %v, got %v", input, result)
		}
		if !isSafe {
			t.Error("expected isSafe to be true")
		}
	})

	t.Run("unpacks Value", func(t *testing.T) {
		val := AsValue("inner")
		input := reflect.ValueOf(val)
		result, isSafe := vr.unpackValue(input, false)
		if result.Interface() != val.val.Interface() {
			t.Errorf("expected %v, got %v", val.val, result)
		}
		if isSafe {
			t.Error("expected isSafe to be false")
		}
	})

	t.Run("unpacks safe Value", func(t *testing.T) {
		val := AsSafeValue("safe inner")
		input := reflect.ValueOf(val)
		result, isSafe := vr.unpackValue(input, false)
		if result.Interface() != val.val.Interface() {
			t.Errorf("expected %v, got %v", val.val, result)
		}
		if !isSafe {
			t.Error("expected isSafe to be true")
		}
	})
}

// TestResolvePartByTypeErrors tests error conditions for resolvePartByType.
func TestResolvePartByTypeErrors(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		context     Context
		expectError bool
		errContains string
	}{
		{
			name:        "int index on non-indexable",
			template:    "{{ x.0 }}",
			context:     Context{"x": 42},
			expectError: true,
			errContains: "can't access an index on type",
		},
		{
			name:        "field on non-struct non-map",
			template:    "{{ x.field }}",
			context:     Context{"x": 42},
			expectError: true,
			errContains: "can't access a field by name on type",
		},
		{
			name:        "subscript on non-indexable",
			template:    "{{ x[0] }}",
			context:     Context{"x": 42},
			expectError: true,
			errContains: "can't access an index on type",
		},
	}

	set := NewSet("test", MustNewLocalFileSystemLoader(""))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := set.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			_, err = tpl.Execute(tt.context)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got none")
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestComplexVariableResolution tests complex variable resolution scenarios.
func TestComplexVariableResolution(t *testing.T) {
	type Inner struct {
		Values []int
	}
	type Outer struct {
		Inner Inner
		Map   map[string]*Inner
	}

	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "deeply nested access",
			template: "{{ o.Inner.Values.0 }}",
			context: Context{
				"o": Outer{Inner: Inner{Values: []int{10, 20, 30}}},
			},
			expected: "10",
		},
		{
			name:     "map to struct to array",
			template: "{{ o.Map.key.Values.1 }}",
			context: Context{
				"o": Outer{Map: map[string]*Inner{
					"key": {Values: []int{100, 200}},
				}},
			},
			expected: "200",
		},
		{
			name:     "chain of method calls",
			template: "{{ s.Upper.Len }}",
			context: Context{
				"s": stringWithMethods("hello"),
			},
			expected: "5",
		},
		{
			name:     "subscript with expression",
			template: "{{ items[idx + 1] }}",
			context:  Context{"items": []string{"a", "b", "c"}, "idx": 0},
			expected: "b",
		},
	}

	set := NewSet("test", MustNewLocalFileSystemLoader(""))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := set.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// stringWithMethods is a helper type for testing method chains.
type stringWithMethods string

func (s stringWithMethods) Upper() stringWithMethods {
	result := make([]byte, len(s))
	for i, c := range []byte(s) {
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return stringWithMethods(result)
}

func (s stringWithMethods) Len() int {
	return len(s)
}
