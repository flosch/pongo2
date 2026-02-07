package pongo2

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type DummyLoader struct{}

func (l *DummyLoader) Abs(base, name string) string {
	return filepath.Join(filepath.Dir(base), name)
}

func (l *DummyLoader) Get(path string) (io.Reader, error) {
	return nil, errors.New("dummy not found")
}

func TestTimeDiff(t *testing.T) {
	tests := []struct {
		name     string
		from     time.Time
		to       time.Time
		expected string
	}{
		{
			name:     "zero difference",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "0 minutes",
		},
		{
			name:     "less than a minute",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 12, 0, 30, 0, time.UTC),
			expected: "0 minutes",
		},
		{
			name:     "exactly 1 minute",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC),
			expected: "1 minute",
		},
		{
			name:     "multiple minutes",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 12, 45, 0, 0, time.UTC),
			expected: "45 minutes",
		},
		{
			name:     "exactly 1 hour",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			expected: "1 hour",
		},
		{
			name:     "1 hour and minutes",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 13, 30, 0, 0, time.UTC),
			expected: "1 hour, 30 minutes",
		},
		{
			name:     "multiple hours",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC),
			expected: "5 hours",
		},
		{
			name:     "exactly 1 day",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			expected: "1 day",
		},
		{
			name:     "1 day and hours",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 2, 18, 0, 0, 0, time.UTC),
			expected: "1 day, 6 hours",
		},
		{
			name:     "multiple days",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC),
			expected: "4 days",
		},
		{
			name:     "exactly 1 week",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC),
			expected: "1 week",
		},
		{
			name:     "1 week and days",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			expected: "1 week, 2 days",
		},
		{
			name:     "multiple weeks",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 22, 12, 0, 0, 0, time.UTC),
			expected: "3 weeks",
		},
		{
			name:     "exactly 1 month (30 days)",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC),
			expected: "1 month",
		},
		{
			name:     "1 month and weeks",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC),
			expected: "1 month, 2 weeks",
		},
		{
			name:     "multiple months",
			from:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2023, 4, 1, 12, 0, 0, 0, time.UTC),
			expected: "3 months",
		},
		{
			name:     "exactly 1 year",
			from:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "1 year",
		},
		{
			name:     "1 year and months",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2025, 4, 1, 12, 0, 0, 0, time.UTC),
			expected: "1 year, 3 months",
		},
		{
			name:     "multiple years",
			from:     time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "4 years, 1 day",
		},
		{
			name:     "negative difference (from > to) should be absolute",
			from:     time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "4 days",
		},
		{
			name:     "only two units shown - years and months only",
			from:     time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 6, 15, 18, 30, 0, 0, time.UTC),
			expected: "4 years, 5 months",
		},
		{
			name:     "complex duration - shows top two units",
			from:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 2, 10, 5, 30, 0, 0, time.UTC),
			expected: "1 month, 1 week",
		},
		{
			name:     "leap year - Feb 29 to Mar 1 in leap year",
			from:     time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
			expected: "1 day",
		},
		{
			name:     "leap year - across leap day",
			from:     time.Date(2024, 2, 28, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
			expected: "2 days",
		},
		{
			name:     "leap year - Feb 28 non-leap to Mar 1",
			from:     time.Date(2023, 2, 28, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2023, 3, 1, 12, 0, 0, 0, time.UTC),
			expected: "1 day",
		},
		{
			name:     "leap year - spanning leap year Feb 29",
			from:     time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC),
			expected: "1 year",
		},
		{
			name:     "sub-second difference returns 0 minutes",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 12, 0, 0, 999999999, time.UTC),
			expected: "0 minutes",
		},
		{
			name:     "nanosecond precision - just under 1 minute",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 12, 0, 59, 999999999, time.UTC),
			expected: "0 minutes",
		},
		{
			name:     "exactly at minute boundary with nanoseconds",
			from:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			to:       time.Date(2024, 1, 1, 12, 1, 0, 1, time.UTC),
			expected: "1 minute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeDiff(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("timeDiff() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFilterTimesince(t *testing.T) {
	baseTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		in       *Value
		param    *Value
		expected string
	}{
		{
			name:     "non-time input returns empty string",
			in:       AsValue("not a time"),
			param:    AsValue(nil),
			expected: "",
		},
		{
			name:     "integer input returns empty string",
			in:       AsValue(12345),
			param:    AsValue(nil),
			expected: "",
		},
		{
			name:     "time with comparison time - 1 hour ago",
			in:       AsValue(baseTime.Add(-1 * time.Hour)),
			param:    AsValue(baseTime),
			expected: "1 hour",
		},
		{
			name:     "time with comparison time - 2 days ago",
			in:       AsValue(baseTime.Add(-48 * time.Hour)),
			param:    AsValue(baseTime),
			expected: "2 days",
		},
		{
			name:     "time with comparison time - 1 week 3 days ago",
			in:       AsValue(baseTime.Add(-10 * 24 * time.Hour)),
			param:    AsValue(baseTime),
			expected: "1 week, 3 days",
		},
		{
			name:     "valid time value with valid time argument - years apart",
			in:       AsValue(time.Date(2020, 3, 15, 10, 30, 0, 0, time.UTC)),
			param:    AsValue(time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)),
			expected: "4 years, 1 day",
		},
		{
			name:     "valid time value with valid time argument - months apart",
			in:       AsValue(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			param:    AsValue(time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)),
			expected: "3 months, 1 day",
		},
		{
			name:     "valid time value with valid time argument - complex duration",
			in:       AsValue(time.Date(2023, 6, 1, 8, 0, 0, 0, time.UTC)),
			param:    AsValue(time.Date(2024, 8, 15, 14, 30, 0, 0, time.UTC)),
			expected: "1 year, 2 months",
		},
		{
			name:     "time with non-time param falls back to now",
			in:       AsValue(baseTime),
			param:    AsValue("invalid"),
			expected: "", // Will vary based on current time, just check no error
		},
		{
			name:     "time with nil param",
			in:       AsValue(baseTime),
			param:    AsValue(nil),
			expected: "", // Will vary based on current time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterTimesince(tt.in, tt.param)
			if err != nil {
				t.Errorf("filterTimesince() error = %v", err)
				return
			}
			// For tests with fixed comparison times, check exact match
			if tt.param != nil && !tt.param.IsNil() {
				if _, ok := tt.param.Interface().(time.Time); ok {
					if result.String() != tt.expected {
						t.Errorf("filterTimesince() = %q, want %q", result.String(), tt.expected)
					}
				}
			}
			// For non-time input, check empty string
			if _, ok := tt.in.Interface().(time.Time); !ok {
				if result.String() != "" {
					t.Errorf("filterTimesince() for non-time input = %q, want empty string", result.String())
				}
			}
		})
	}
}

func TestFilterTimeuntil(t *testing.T) {
	baseTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		in       *Value
		param    *Value
		expected string
	}{
		{
			name:     "non-time input returns empty string",
			in:       AsValue("not a time"),
			param:    AsValue(nil),
			expected: "",
		},
		{
			name:     "integer input returns empty string",
			in:       AsValue(12345),
			param:    AsValue(nil),
			expected: "",
		},
		{
			name:     "time with comparison time - 1 hour from now",
			in:       AsValue(baseTime.Add(1 * time.Hour)),
			param:    AsValue(baseTime),
			expected: "1 hour",
		},
		{
			name:     "time with comparison time - 2 days from now",
			in:       AsValue(baseTime.Add(48 * time.Hour)),
			param:    AsValue(baseTime),
			expected: "2 days",
		},
		{
			name:     "time with comparison time - 1 week 3 days from now",
			in:       AsValue(baseTime.Add(10 * 24 * time.Hour)),
			param:    AsValue(baseTime),
			expected: "1 week, 3 days",
		},
		{
			name:     "valid time value with valid time argument - years apart",
			in:       AsValue(time.Date(2028, 3, 15, 10, 30, 0, 0, time.UTC)),
			param:    AsValue(time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)),
			expected: "4 years, 1 day",
		},
		{
			name:     "valid time value with valid time argument - months apart",
			in:       AsValue(time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)),
			param:    AsValue(time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)),
			expected: "3 months, 1 day",
		},
		{
			name:     "valid time value with valid time argument - complex duration",
			in:       AsValue(time.Date(2025, 10, 20, 16, 45, 0, 0, time.UTC)),
			param:    AsValue(time.Date(2024, 8, 15, 14, 30, 0, 0, time.UTC)),
			expected: "1 year, 2 months",
		},
		{
			name:     "time with non-time param falls back to now",
			in:       AsValue(baseTime),
			param:    AsValue("invalid"),
			expected: "", // Will vary based on current time
		},
		{
			name:     "time with nil param",
			in:       AsValue(baseTime),
			param:    AsValue(nil),
			expected: "", // Will vary based on current time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterTimeuntil(tt.in, tt.param)
			if err != nil {
				t.Errorf("filterTimeuntil() error = %v", err)
				return
			}
			// For tests with fixed comparison times, check exact match
			if tt.param != nil && !tt.param.IsNil() {
				if _, ok := tt.param.Interface().(time.Time); ok {
					if result.String() != tt.expected {
						t.Errorf("filterTimeuntil() = %q, want %q", result.String(), tt.expected)
					}
				}
			}
			// For non-time input, check empty string
			if _, ok := tt.in.Interface().(time.Time); !ok {
				if result.String() != "" {
					t.Errorf("filterTimeuntil() for non-time input = %q, want empty string", result.String())
				}
			}
		})
	}
}

func TestFilterDictsort(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		key      string
		expected []string // expected order of the key values
		wantErr  bool
	}{
		{
			name: "sort maps by string key",
			input: []map[string]any{
				{"name": "Charlie", "age": 25},
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 20},
			},
			key:      "name",
			expected: []string{"Alice", "Bob", "Charlie"},
		},
		{
			name: "sort maps by numeric key (as string)",
			input: []map[string]any{
				{"name": "Charlie", "age": 25},
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 20},
			},
			key:      "age",
			expected: []string{"20", "25", "30"},
		},
		{
			name:     "empty slice",
			input:    []map[string]any{},
			key:      "name",
			expected: []string{},
		},
		{
			name: "single element",
			input: []map[string]any{
				{"name": "Alice"},
			},
			key:      "name",
			expected: []string{"Alice"},
		},
		{
			name: "missing key in some items",
			input: []map[string]any{
				{"name": "Charlie"},
				{"other": "value"},
				{"name": "Alice"},
			},
			key:      "name",
			expected: []string{"<nil>", "Alice", "Charlie"},
		},
		{
			name:    "nil key parameter",
			input:   []map[string]any{{"name": "Alice"}},
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var param *Value
			if tt.key == "" && tt.wantErr {
				param = AsValue(nil)
			} else {
				param = AsValue(tt.key)
			}

			result, err := filterDictsort(AsValue(tt.input), param)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the order
			var resultKeys []string
			result.Iterate(func(idx, count int, key, value *Value) bool {
				item := key.Interface()
				if m, ok := item.(map[string]any); ok {
					resultKeys = append(resultKeys, fmt.Sprintf("%v", m[tt.key]))
				}
				return true
			}, func() {})

			if len(resultKeys) != len(tt.expected) {
				t.Errorf("got %d items, want %d", len(resultKeys), len(tt.expected))
				return
			}

			for i, exp := range tt.expected {
				if resultKeys[i] != exp {
					t.Errorf("at index %d: got %q, want %q", i, resultKeys[i], exp)
				}
			}
		})
	}
}

func TestFilterDictsortNonMapInput(t *testing.T) {
	// Test with integer - not sliceable
	intInput := 42
	result, err := filterDictsort(AsValue(intInput), AsValue("name"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Integer is returned unchanged
	if result.Integer() != 42 {
		t.Errorf("expected integer to be returned unchanged, got %v", result.Integer())
	}

	// Test with nil
	result, err = filterDictsort(AsValue(nil), AsValue("name"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNil() {
		t.Errorf("expected nil to be returned unchanged")
	}
}

func TestFilterDictsortReversed(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		key      string
		expected []string
	}{
		{
			name: "reverse sort maps by string key",
			input: []map[string]any{
				{"name": "Alice", "age": 30},
				{"name": "Charlie", "age": 25},
				{"name": "Bob", "age": 20},
			},
			key:      "name",
			expected: []string{"Charlie", "Bob", "Alice"},
		},
		{
			name: "reverse sort maps by numeric key",
			input: []map[string]any{
				{"name": "Charlie", "age": 25},
				{"name": "Alice", "age": 30},
				{"name": "Bob", "age": 20},
			},
			key:      "age",
			expected: []string{"30", "25", "20"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterDictsortReversed(AsValue(tt.input), AsValue(tt.key))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var resultKeys []string
			result.Iterate(func(idx, count int, key, value *Value) bool {
				item := key.Interface()
				if m, ok := item.(map[string]any); ok {
					resultKeys = append(resultKeys, fmt.Sprintf("%v", m[tt.key]))
				}
				return true
			}, func() {})

			if len(resultKeys) != len(tt.expected) {
				t.Errorf("got %d items, want %d", len(resultKeys), len(tt.expected))
				return
			}

			for i, exp := range tt.expected {
				if resultKeys[i] != exp {
					t.Errorf("at index %d: got %q, want %q", i, resultKeys[i], exp)
				}
			}
		})
	}
}

type testPerson struct {
	Name string
	Age  int
}

func TestFilterDictsortWithStructs(t *testing.T) {
	input := []testPerson{
		{Name: "Charlie", Age: 25},
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 20},
	}

	t.Run("sort structs by Name field", func(t *testing.T) {
		result, err := filterDictsort(AsValue(input), AsValue("Name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{"Alice", "Bob", "Charlie"}
		var resultNames []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			item := key.Interface()
			if p, ok := item.(testPerson); ok {
				resultNames = append(resultNames, p.Name)
			}
			return true
		}, func() {})

		for i, exp := range expected {
			if resultNames[i] != exp {
				t.Errorf("at index %d: got %q, want %q", i, resultNames[i], exp)
			}
		}
	})

	t.Run("sort structs by Age field", func(t *testing.T) {
		result, err := filterDictsort(AsValue(input), AsValue("Age"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []int{20, 25, 30}
		var resultAges []int
		result.Iterate(func(idx, count int, key, value *Value) bool {
			item := key.Interface()
			if p, ok := item.(testPerson); ok {
				resultAges = append(resultAges, p.Age)
			}
			return true
		}, func() {})

		for i, exp := range expected {
			if resultAges[i] != exp {
				t.Errorf("at index %d: got %d, want %d", i, resultAges[i], exp)
			}
		}
	})

	t.Run("sort structs by non-existent field", func(t *testing.T) {
		result, err := filterDictsort(AsValue(input), AsValue("NonExistent"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// All sort keys are empty, so order is stable (original order)
		var count int
		result.Iterate(func(idx, cnt int, key, value *Value) bool {
			count++
			return true
		}, func() {})

		if count != 3 {
			t.Errorf("expected 3 items, got %d", count)
		}
	})
}

func TestFilterDictsortWithPointers(t *testing.T) {
	input := []*testPerson{
		{Name: "Charlie", Age: 25},
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 20},
	}

	result, err := filterDictsort(AsValue(input), AsValue("Name"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"Alice", "Bob", "Charlie"}
	var resultNames []string
	result.Iterate(func(idx, count int, key, value *Value) bool {
		item := key.Interface()
		if p, ok := item.(*testPerson); ok {
			resultNames = append(resultNames, p.Name)
		}
		return true
	}, func() {})

	for i, exp := range expected {
		if resultNames[i] != exp {
			t.Errorf("at index %d: got %q, want %q", i, resultNames[i], exp)
		}
	}
}

func TestFilterUnorderedList(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "simple flat list",
			input:    []string{"Item 1", "Item 2", "Item 3"},
			expected: "<li>Item 1</li><li>Item 2</li><li>Item 3</li>",
		},
		{
			name:     "empty list",
			input:    []string{},
			expected: "",
		},
		{
			name:     "single item",
			input:    []string{"Only Item"},
			expected: "<li>Only Item</li>",
		},
		{
			name:     "list with HTML escaping",
			input:    []string{"<script>alert('xss')</script>", "Normal Item"},
			expected: "<li>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</li><li>Normal Item</li>",
		},
		{
			name:     "list with numbers",
			input:    []int{1, 2, 3},
			expected: "<li>1</li><li>2</li><li>3</li>",
		},
		{
			name:     "non-slice input",
			input:    "not a slice",
			expected: "",
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterUnorderedList(AsValue(tt.input), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterUnorderedListNested tests nested lists using template execution
// since direct iteration over []any doesn't work well with pongo2's Value wrapper
func TestFilterUnorderedListNested(t *testing.T) {
	ts := NewSet("test", &DummyLoader{})

	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "simple list via template",
			template: "{{ items|unordered_list }}",
			context:  Context{"items": []string{"A", "B", "C"}},
			expected: "<li>A</li><li>B</li><li>C</li>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := ts.FromString(tt.template)
			if err != nil {
				t.Fatalf("template parse error: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("template execute error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFilterUnorderedListRecursionLimit(t *testing.T) {
	// Create a deeply nested structure that exceeds the recursion limit
	var createDeep func(depth int) any
	createDeep = func(depth int) any {
		if depth == 0 {
			return "leaf"
		}
		return []any{createDeep(depth - 1)}
	}

	// Create a structure deeper than maxUnorderedListDepth (100)
	deepInput := createDeep(150)

	result, err := filterUnorderedList(AsValue(deepInput), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The result should not panic and should have some output
	// The exact output depends on how far we got before hitting the limit
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestFilterSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple words",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "with special characters",
			input:    "Hello, World!",
			expected: "hello-world",
		},
		{
			name:     "multiple spaces",
			input:    "Hello   World",
			expected: "hello-world",
		},
		{
			name:     "leading/trailing spaces",
			input:    "  Hello World  ",
			expected: "hello-world",
		},
		{
			name:     "already lowercase",
			input:    "hello world",
			expected: "hello-world",
		},
		{
			name:     "with numbers",
			input:    "Hello World 123",
			expected: "hello-world-123",
		},
		{
			name:     "unicode characters normalized via NFKD",
			input:    "Hello Wörld",
			expected: "hello-world",
		},
		{
			name:     "all special characters",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "existing hyphens",
			input:    "hello-world",
			expected: "hello-world",
		},
		{
			name:     "multiple hyphens",
			input:    "hello---world",
			expected: "hello-world",
		},
		{
			name:     "mixed case with punctuation",
			input:    "The Quick Brown Fox!",
			expected: "the-quick-brown-fox",
		},
		{
			name:     "numbers only",
			input:    "123 456",
			expected: "123-456",
		},
		{
			name:     "leading special characters",
			input:    "!!!Hello",
			expected: "hello",
		},
		{
			name:     "trailing special characters",
			input:    "Hello!!!",
			expected: "hello",
		},
		{
			name:     "real-world blog title",
			input:    "10 Tips for Better Go Code!",
			expected: "10-tips-for-better-go-code",
		},
		{
			name:     "question as title",
			input:    "What is Go?",
			expected: "what-is-go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterSlugify(AsValue(tt.input), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.String() != tt.expected {
				t.Errorf("filterSlugify(%q) = %q, want %q", tt.input, result.String(), tt.expected)
			}
		})
	}
}

func TestFilterFilesizeformat(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "zero bytes",
			input:    0,
			expected: "0 bytes",
		},
		{
			name:     "one byte",
			input:    1,
			expected: "1 bytes",
		},
		{
			name:     "bytes range",
			input:    100,
			expected: "100 bytes",
		},
		{
			name:     "exactly 1KB",
			input:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "KB range",
			input:    1536, // 1.5 KB
			expected: "1.5 KB",
		},
		{
			name:     "exactly 1MB",
			input:    1024 * 1024,
			expected: "1.0 MB",
		},
		{
			name:     "MB range",
			input:    1024 * 1024 * 5,
			expected: "5.0 MB",
		},
		{
			name:     "Django example: 123456789",
			input:    123456789,
			expected: "117.7 MB",
		},
		{
			name:     "exactly 1GB",
			input:    1024 * 1024 * 1024,
			expected: "1.0 GB",
		},
		{
			name:     "GB range",
			input:    1024 * 1024 * 1024 * 2,
			expected: "2.0 GB",
		},
		{
			name:     "exactly 1TB",
			input:    1024 * 1024 * 1024 * 1024,
			expected: "1.0 TB",
		},
		{
			name:     "exactly 1PB",
			input:    1024 * 1024 * 1024 * 1024 * 1024,
			expected: "1.0 PB",
		},
		{
			name:     "large PB value",
			input:    1024 * 1024 * 1024 * 1024 * 1024 * 5,
			expected: "5.0 PB",
		},
		{
			name:     "1023 bytes (just under 1KB)",
			input:    1023,
			expected: "1023 bytes",
		},
		{
			name:     "fractional KB",
			input:    2560, // 2.5 KB
			expected: "2.5 KB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterFilesizeformat(AsValue(tt.input), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.String() != tt.expected {
				t.Errorf("filterFilesizeformat(%d) = %q, want %q", tt.input, result.String(), tt.expected)
			}
		})
	}
}

func TestFilterFilesizeformatNegative(t *testing.T) {
	// Negative values should be treated as 0
	result, err := filterFilesizeformat(AsValue(-100), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.String() != "0 bytes" {
		t.Errorf("filterFilesizeformat(-100) = %q, want %q", result.String(), "0 bytes")
	}
}

func TestFilterFilesizeformatNonInteger(t *testing.T) {
	// String inputs should be converted (or treated as 0)
	result, err := filterFilesizeformat(AsValue("not a number"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// AsValue("not a number").Integer() returns 0
	if result.String() != "0 bytes" {
		t.Errorf("filterFilesizeformat(\"not a number\") = %q, want %q", result.String(), "0 bytes")
	}
}

// TestDictsortItemsSortInterface tests the sort.Interface implementation for dictsortItems
func TestDictsortItemsSortInterface(t *testing.T) {
	t.Run("Len", func(t *testing.T) {
		items := dictsortItems{
			entries: []struct {
				item   *Value
				sortBy *Value
			}{
				{item: AsValue("a"), sortBy: AsValue("1")},
				{item: AsValue("b"), sortBy: AsValue("2")},
				{item: AsValue("c"), sortBy: AsValue("3")},
			},
		}
		if items.Len() != 3 {
			t.Errorf("Len() = %d, want 3", items.Len())
		}

		emptyItems := dictsortItems{}
		if emptyItems.Len() != 0 {
			t.Errorf("Len() of empty = %d, want 0", emptyItems.Len())
		}
	})

	t.Run("Swap", func(t *testing.T) {
		items := dictsortItems{
			entries: []struct {
				item   *Value
				sortBy *Value
			}{
				{item: AsValue("a"), sortBy: AsValue("1")},
				{item: AsValue("b"), sortBy: AsValue("2")},
			},
		}
		items.Swap(0, 1)
		if items.entries[0].sortBy.String() != "2" || items.entries[1].sortBy.String() != "1" {
			t.Errorf("Swap failed: got [%s, %s], want [2, 1]", items.entries[0].sortBy.String(), items.entries[1].sortBy.String())
		}
	})

	t.Run("Less with string keys", func(t *testing.T) {
		items := dictsortItems{
			entries: []struct {
				item   *Value
				sortBy *Value
			}{
				{item: AsValue("a"), sortBy: AsValue("apple")},
				{item: AsValue("b"), sortBy: AsValue("banana")},
				{item: AsValue("c"), sortBy: AsValue("apple")},
			},
		}

		if !items.Less(0, 1) {
			t.Error("Less(0, 1) should be true: apple < banana")
		}
		if items.Less(1, 0) {
			t.Error("Less(1, 0) should be false: banana > apple")
		}
		if items.Less(0, 2) {
			t.Error("Less(0, 2) should be false: apple == apple")
		}
	})

	t.Run("numeric keys sort numerically", func(t *testing.T) {
		items := dictsortItems{
			allNumeric: true,
			entries: []struct {
				item   *Value
				sortBy *Value
			}{
				{item: AsValue("a"), sortBy: AsValue(10)},
				{item: AsValue("b"), sortBy: AsValue(2)},
				{item: AsValue("c"), sortBy: AsValue(1)},
			},
		}

		// Numeric comparison: 1 < 2 < 10
		if !items.Less(2, 1) { // 1 < 2
			t.Error("Less should compare numerically: 1 < 2")
		}
		if !items.Less(1, 0) { // 2 < 10
			t.Error("Less should compare numerically: 2 < 10")
		}
		if items.Less(0, 1) { // 10 > 2
			t.Error("Less should compare numerically: 10 is not < 2")
		}
	})
}

// TestDictsortHelperEdgeCases tests edge cases in dictsortHelper
func TestDictsortHelperEdgeCases(t *testing.T) {
	t.Run("string input (sliceable but not map/struct items)", func(t *testing.T) {
		// Strings are sliceable, but individual characters are not maps/structs
		result, err := filterDictsort(AsValue("hello"), AsValue("key"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should return a slice of characters with empty sort keys
		if result.Len() != 5 {
			t.Errorf("expected 5 items (characters), got %d", result.Len())
		}
	})

	t.Run("slice of integers (not map/struct)", func(t *testing.T) {
		input := []int{3, 1, 2}
		result, err := filterDictsort(AsValue(input), AsValue("key"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Integers are not maps/structs, so sortKey is empty for all
		// Order should remain stable since all keys are equal
		if result.Len() != 3 {
			t.Errorf("expected 3 items, got %d", result.Len())
		}
	})

	t.Run("slice of strings (not map/struct)", func(t *testing.T) {
		input := []string{"charlie", "alice", "bob"}
		result, err := filterDictsort(AsValue(input), AsValue("key"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Strings are not maps/structs, so sortKey is empty for all
		if result.Len() != 3 {
			t.Errorf("expected 3 items, got %d", result.Len())
		}
	})

	t.Run("mixed types in slice", func(t *testing.T) {
		input := []any{
			map[string]any{"name": "Alice"},
			"not a map",
			map[string]any{"name": "Bob"},
		}
		result, err := filterDictsort(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The non-map item will have empty sortKey
		if result.Len() != 3 {
			t.Errorf("expected 3 items, got %d", result.Len())
		}
	})

	t.Run("map with interface values", func(t *testing.T) {
		input := []map[string]any{
			{"name": "Charlie", "value": 100},
			{"name": "Alice", "value": "string"},
			{"name": "Bob", "value": true},
		}
		result, err := filterDictsort(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var names []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if m, ok := key.Interface().(map[string]any); ok {
				names = append(names, fmt.Sprintf("%v", m["name"]))
			}
			return true
		}, func() {})

		expected := []string{"Alice", "Bob", "Charlie"}
		for i, exp := range expected {
			if names[i] != exp {
				t.Errorf("at index %d: got %q, want %q", i, names[i], exp)
			}
		}
	})

	t.Run("slice of pointers to maps", func(t *testing.T) {
		m1 := map[string]any{"name": "Charlie"}
		m2 := map[string]any{"name": "Alice"}
		m3 := map[string]any{"name": "Bob"}
		input := []*map[string]any{&m1, &m2, &m3}

		result, err := filterDictsort(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var names []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if m, ok := key.Interface().(*map[string]any); ok {
				names = append(names, fmt.Sprintf("%v", (*m)["name"]))
			}
			return true
		}, func() {})

		expected := []string{"Alice", "Bob", "Charlie"}
		for i, exp := range expected {
			if names[i] != exp {
				t.Errorf("at index %d: got %q, want %q", i, names[i], exp)
			}
		}
	})

	t.Run("equal sort keys maintain order", func(t *testing.T) {
		input := []map[string]any{
			{"name": "Same", "id": 1},
			{"name": "Same", "id": 2},
			{"name": "Same", "id": 3},
		}
		result, err := filterDictsort(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// All have same sort key, stable sort should preserve original order
		var ids []int
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if m, ok := key.Interface().(map[string]any); ok {
				ids = append(ids, m["id"].(int))
			}
			return true
		}, func() {})

		if len(ids) != 3 {
			t.Errorf("expected 3 items, got %d", len(ids))
		}
	})

	t.Run("bool value", func(t *testing.T) {
		// Bool is not sliceable, should return unchanged
		result, err := filterDictsort(AsValue(true), AsValue("key"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Bool() {
			t.Error("expected true to be returned unchanged")
		}
	})

	t.Run("struct with unexported fields", func(t *testing.T) {
		type privateStruct struct {
			Name    string
			private int //nolint:unused
		}
		input := []privateStruct{
			{Name: "Charlie"},
			{Name: "Alice"},
			{Name: "Bob"},
		}

		result, err := filterDictsort(AsValue(input), AsValue("Name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var names []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if p, ok := key.Interface().(privateStruct); ok {
				names = append(names, p.Name)
			}
			return true
		}, func() {})

		expected := []string{"Alice", "Bob", "Charlie"}
		for i, exp := range expected {
			if names[i] != exp {
				t.Errorf("at index %d: got %q, want %q", i, names[i], exp)
			}
		}
	})
}

// TestDictsortReversedEdgeCases tests edge cases specifically for dictsortreversed
func TestDictsortReversedEdgeCases(t *testing.T) {
	t.Run("nil key parameter", func(t *testing.T) {
		input := []map[string]any{{"name": "Alice"}}
		_, err := filterDictsortReversed(AsValue(input), AsValue(nil))
		if err == nil {
			t.Error("expected error for nil key parameter")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []map[string]any{}
		result, err := filterDictsortReversed(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 0 {
			t.Errorf("expected empty result, got %d items", result.Len())
		}
	})

	t.Run("single element", func(t *testing.T) {
		input := []map[string]any{{"name": "Only"}}
		result, err := filterDictsortReversed(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var names []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if m, ok := key.Interface().(map[string]any); ok {
				names = append(names, fmt.Sprintf("%v", m["name"]))
			}
			return true
		}, func() {})

		if len(names) != 1 || names[0] != "Only" {
			t.Errorf("expected [Only], got %v", names)
		}
	})

	t.Run("already sorted input reversed", func(t *testing.T) {
		input := []map[string]any{
			{"name": "Alice"},
			{"name": "Bob"},
			{"name": "Charlie"},
		}
		result, err := filterDictsortReversed(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var names []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if m, ok := key.Interface().(map[string]any); ok {
				names = append(names, fmt.Sprintf("%v", m["name"]))
			}
			return true
		}, func() {})

		expected := []string{"Charlie", "Bob", "Alice"}
		for i, exp := range expected {
			if names[i] != exp {
				t.Errorf("at index %d: got %q, want %q", i, names[i], exp)
			}
		}
	})

	t.Run("non-sliceable input", func(t *testing.T) {
		result, err := filterDictsortReversed(AsValue(42), AsValue("key"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Integer() != 42 {
			t.Errorf("expected 42, got %d", result.Integer())
		}
	})
}

// TestDictsortWithSpecialCharKeys tests sorting with special character keys
func TestDictsortWithSpecialCharKeys(t *testing.T) {
	t.Run("unicode sort keys", func(t *testing.T) {
		input := []map[string]any{
			{"name": "日本"},
			{"name": "中国"},
			{"name": "한국"},
		}
		result, err := filterDictsort(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should sort by unicode code points
		if result.Len() != 3 {
			t.Errorf("expected 3 items, got %d", result.Len())
		}
	})

	t.Run("empty string keys", func(t *testing.T) {
		input := []map[string]any{
			{"name": ""},
			{"name": "Bob"},
			{"name": ""},
		}
		result, err := filterDictsort(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var names []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if m, ok := key.Interface().(map[string]any); ok {
				names = append(names, fmt.Sprintf("%q", m["name"]))
			}
			return true
		}, func() {})

		// Empty strings should come first
		if names[0] != `""` || names[1] != `""` {
			t.Errorf("empty strings should sort first, got %v", names)
		}
	})

	t.Run("whitespace keys", func(t *testing.T) {
		input := []map[string]any{
			{"name": "  z"},
			{"name": "a"},
			{"name": " b"},
		}
		result, err := filterDictsort(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var names []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if m, ok := key.Interface().(map[string]any); ok {
				names = append(names, fmt.Sprintf("%v", m["name"]))
			}
			return true
		}, func() {})

		// Space comes before letters in ASCII: " " < "  " < "a"
		// "  z" starts with two spaces, " b" starts with one space
		expected := []string{"  z", " b", "a"}
		for i, exp := range expected {
			if names[i] != exp {
				t.Errorf("at index %d: got %q, want %q", i, names[i], exp)
			}
		}
	})
}

// TestDictsortWithNestedValues tests sorting maps with nested values
func TestDictsortWithNestedValues(t *testing.T) {
	t.Run("sort by nested map key", func(t *testing.T) {
		// The sortKey is extracted from top-level only
		input := []map[string]any{
			{"name": "Charlie", "info": map[string]any{"age": 25}},
			{"name": "Alice", "info": map[string]any{"age": 30}},
			{"name": "Bob", "info": map[string]any{"age": 20}},
		}
		result, err := filterDictsort(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var names []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if m, ok := key.Interface().(map[string]any); ok {
				names = append(names, fmt.Sprintf("%v", m["name"]))
			}
			return true
		}, func() {})

		expected := []string{"Alice", "Bob", "Charlie"}
		for i, exp := range expected {
			if names[i] != exp {
				t.Errorf("at index %d: got %q, want %q", i, names[i], exp)
			}
		}
	})
}

// TestDictsortInterfaceSlice tests dictsort with []any containing maps
func TestDictsortInterfaceSlice(t *testing.T) {
	t.Run("interface slice with maps", func(t *testing.T) {
		input := []any{
			map[string]any{"name": "Charlie"},
			map[string]any{"name": "Alice"},
			map[string]any{"name": "Bob"},
		}
		result, err := filterDictsort(AsValue(input), AsValue("name"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var names []string
		result.Iterate(func(idx, count int, key, value *Value) bool {
			if m, ok := key.Interface().(map[string]any); ok {
				names = append(names, fmt.Sprintf("%v", m["name"]))
			}
			return true
		}, func() {})

		expected := []string{"Alice", "Bob", "Charlie"}
		for i, exp := range expected {
			if names[i] != exp {
				t.Errorf("at index %d: got %q, want %q", i, names[i], exp)
			}
		}
	})
}

// TestMustRegisterFilterPanic tests that mustRegisterFilter panics on duplicate registration
func TestMustRegisterFilterPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("mustRegisterFilter should panic when registering a duplicate filter")
		}
	}()
	// "escape" is already registered in init()
	mustRegisterFilter("escape", func(in *Value, param *Value) (*Value, error) {
		return in, nil
	})
}

// TestFilterTruncateHTMLHelperRuneError tests handling of invalid UTF-8 sequences
func TestFilterTruncateHTMLHelperRuneError(t *testing.T) {
	// Create a string with invalid UTF-8 in various positions
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid utf8 at start",
			input: "\xff<p>hello</p>",
		},
		{
			name:  "invalid utf8 in tag",
			input: "<p\xff>hello</p>",
		},
		{
			name:  "invalid utf8 in close tag",
			input: "<p>hello</p\xff>",
		},
		{
			name:  "invalid utf8 in content",
			input: "<p>hel\xfflo</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// truncatechars_html should not panic on invalid UTF-8
			result, err := filterTruncatecharsHTML(AsValue(tt.input), AsValue(100))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			// Just verify it doesn't panic and returns something
			_ = result.String()
		})
	}
}

// TestFilterTruncatewordsHTMLRuneError tests truncatewords_html with invalid UTF-8
func TestFilterTruncatewordsHTMLRuneError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid utf8 in word",
			input: "<p>word1 wor\xffd2 word3</p>",
		},
		{
			name:  "invalid utf8 at word boundary",
			input: "<p>word1\xff word2</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterTruncatewordsHTML(AsValue(tt.input), AsValue(2))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			_ = result.String()
		})
	}
}

// TestFilterCenterMaxPadding tests the error case when padding exceeds maxCharPadding
func TestFilterCenterMaxPadding(t *testing.T) {
	result, err := filterCenter(AsValue("test"), AsValue(20000))
	if err == nil {
		t.Error("expected error for excessive padding")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestFilterLjustMaxPadding tests the error case when padding exceeds maxCharPadding
func TestFilterLjustMaxPadding(t *testing.T) {
	result, err := filterLjust(AsValue("test"), AsValue(20000))
	if err == nil {
		t.Error("expected error for excessive padding")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestFilterRjustMaxPadding tests the error case when padding exceeds maxCharPadding
func TestFilterRjustMaxPadding(t *testing.T) {
	result, err := filterRjust(AsValue("test"), AsValue(20000))
	if err == nil {
		t.Error("expected error for excessive padding")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestFilterDateNonTimeInput tests that filterDate returns error for non-time input
func TestFilterDateNonTimeInput(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"string", "not a time"},
		{"integer", 12345},
		{"float", 3.14},
		{"nil", nil},
		{"bool", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterDate(AsValue(tt.input), AsValue("2006-01-02"))
			if err == nil {
				t.Error("expected error for non-time input")
			}
			if result != nil {
				t.Error("expected nil result on error")
			}
		})
	}
}

// TestFilterDateValidTime tests filterDate with valid time input
func TestFilterDateValidTime(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
	result, err := filterDate(AsValue(testTime), AsValue("2006-01-02"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "2024-03-15" {
		t.Errorf("got %q, want %q", result.String(), "2024-03-15")
	}
}

// TestFilterFloatformatMaxDecimals tests the error case when decimals exceed maximum
func TestFilterFloatformatMaxDecimals(t *testing.T) {
	result, err := filterFloatformat(AsValue(3.14), AsValue(2000))
	if err == nil {
		t.Error("expected error for excessive decimals")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestFilterFloatformatEdgeCases tests various edge cases
func TestFilterFloatformatEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		param    any
		expected string
	}{
		{"whole number with negative param", 34.0, -3, "34"},
		{"decimal with zero param trims", 34.5, 0, "34"}, // zero param means trim, whole part only
		{"nil param on whole number", 42.0, nil, "42"},
		{"nil param on decimal", 42.5, nil, "42.5"},
		{"non-number param trims whole", 42.0, "abc", "42"}, // non-number param with whole number
		{"negative value", -3.14159, 2, "-3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterFloatformat(AsValue(tt.input), AsValue(tt.param))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterPluralizeNonNumber tests that pluralize returns error for non-number input
func TestFilterPluralizeNonNumber(t *testing.T) {
	result, err := filterPluralize(AsValue("not a number"), AsValue("s"))
	if err == nil {
		t.Error("expected error for non-number input")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestFilterPluralizeTooManyArgs tests that pluralize returns error for >2 arguments
func TestFilterPluralizeTooManyArgs(t *testing.T) {
	result, err := filterPluralize(AsValue(5), AsValue("a,b,c,d"))
	if err == nil {
		t.Error("expected error for too many arguments")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

// TestFilterRemovetagsEdgeCases tests various edge cases
func TestFilterRemovetagsEdgeCases(t *testing.T) {
	t.Run("valid single letter tag", func(t *testing.T) {
		result, err := filterRemovetags(AsValue("<b>bold</b> and <i>italic</i>"), AsValue("b"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "bold and <i>italic</i>" {
			t.Errorf("got %q, want %q", result.String(), "bold and <i>italic</i>")
		}
	})

	t.Run("valid tag name - multiple letters", func(t *testing.T) {
		result, err := filterRemovetags(AsValue("<div>content</div>"), AsValue("div"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "content" {
			t.Errorf("got %q, want %q", result.String(), "content")
		}
	})

	t.Run("invalid tag name - starts with number", func(t *testing.T) {
		result, err := filterRemovetags(AsValue("<1>content</1>"), AsValue("1"))
		if err == nil {
			t.Error("expected error for invalid tag name")
		}
		if result != nil {
			t.Error("expected nil result on error")
		}
	})

	t.Run("invalid tag name - special char", func(t *testing.T) {
		result, err := filterRemovetags(AsValue("text"), AsValue("@"))
		if err == nil {
			t.Error("expected error for invalid tag name")
		}
		if result != nil {
			t.Error("expected nil result on error")
		}
	})

	t.Run("multiple valid tags", func(t *testing.T) {
		result, err := filterRemovetags(AsValue("<a><b>text</b></a>"), AsValue("a,b"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "text" {
			t.Errorf("got %q, want %q", result.String(), "text")
		}
	})
}

// TestFilterRemovetagsNested tests removetags with nested/obfuscated tags
func TestFilterRemovetagsNested(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tags     string
		expected string
	}{
		{
			name:     "simple tag removal",
			input:    "<b>bold</b> text",
			tags:     "b",
			expected: "bold text",
		},
		{
			name:     "nested obfuscated tag - security test",
			input:    "<sc<script>ript>alert('XSS')</sc</script>ript>",
			tags:     "script",
			expected: "alert('XSS')",
		},
		{
			name:     "double nested",
			input:    "<scr<scr<script>ipt>ipt>alert(1)</scr</scr</script>ipt>ipt>",
			tags:     "script",
			expected: "alert(1)",
		},
		{
			name:     "tag with attributes",
			input:    `<a href="http://example.com" class="link">click</a>`,
			tags:     "a",
			expected: "click",
		},
		{
			name:     "multiple tags",
			input:    "<b><i>bold italic</i></b>",
			tags:     "b,i",
			expected: "bold italic",
		},
		{
			name:     "case insensitive",
			input:    "<SCRIPT>alert(1)</SCRIPT>",
			tags:     "script",
			expected: "alert(1)",
		},
		{
			name:     "self-closing tag",
			input:    "line1<br>line2<br/>line3",
			tags:     "br",
			expected: "line1line2line3",
		},
		{
			name:     "tag not in list preserved",
			input:    "<b>bold</b> and <i>italic</i>",
			tags:     "b",
			expected: "bold and <i>italic</i>",
		},
		{
			name:     "empty input",
			input:    "",
			tags:     "script",
			expected: "",
		},
		{
			name:     "no tags present",
			input:    "plain text",
			tags:     "script",
			expected: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterRemovetags(AsValue(tt.input), AsValue(tt.tags))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterRemovetagsInvalidTag tests error handling for invalid tag names
func TestFilterRemovetagsInvalidTag(t *testing.T) {
	invalidTags := []string{
		"<script>",
		"123",
		"tag with space",
		"tag<>",
	}

	for _, tag := range invalidTags {
		t.Run(tag, func(t *testing.T) {
			_, err := filterRemovetags(AsValue("test"), AsValue(tag))
			if err == nil {
				t.Errorf("expected error for invalid tag %q", tag)
			}
		})
	}
}

// TestFilterRemovetagsMaxIterations tests that removetags returns an error when max iterations is reached
func TestFilterRemovetagsMaxIterations(t *testing.T) {
	// Create a pathological input that would require more than 100 iterations
	// by deeply nesting tag fragments
	input := strings.Repeat("<scr", 150) + "ipt>" + strings.Repeat("ipt>", 149)
	_, err := filterRemovetags(AsValue(input), AsValue("script"))
	if err == nil {
		t.Error("expected error when max iterations reached")
	}
	if err != nil && !strings.Contains(err.Error(), "max iterations") {
		t.Errorf("expected max iterations error, got: %v", err)
	}
}

// TestFilterSliceInvalidFormat tests the error case for invalid slice format
func TestFilterSliceInvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		param string
	}{
		{"no colon", "5"},
		{"too many colons", "1:2:3"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterSlice(AsValue([]int{1, 2, 3}), AsValue(tt.param))
			if err == nil {
				t.Error("expected error for invalid slice format")
			}
			if result != nil {
				t.Error("expected nil result on error")
			}
		})
	}
}

// TestFilterSliceNonSliceable tests slice with non-sliceable input
func TestFilterSliceNonSliceable(t *testing.T) {
	result, err := filterSlice(AsValue(42), AsValue("1:2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Non-sliceable input is returned unchanged
	if result.Integer() != 42 {
		t.Errorf("got %d, want 42", result.Integer())
	}
}

// TestFilterYesnoErrors tests error cases for yesno filter
func TestFilterYesnoErrors(t *testing.T) {
	t.Run("too many arguments", func(t *testing.T) {
		result, err := filterYesno(AsValue(true), AsValue("a,b,c,d"))
		if err == nil {
			t.Error("expected error for too many arguments")
		}
		if result != nil {
			t.Error("expected nil result on error")
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		result, err := filterYesno(AsValue(true), AsValue("only_one"))
		if err == nil {
			t.Error("expected error for too few arguments")
		}
		if result != nil {
			t.Error("expected nil result on error")
		}
	})
}

// TestFilterYesnoCustomOptions tests yesno with custom 2-argument options
func TestFilterYesnoCustomOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		param    string
		expected string
	}{
		{"true with 2 args", true, "on,off", "on"},
		{"false with 2 args", false, "on,off", "off"},
		{"nil with 2 args maps to no value", nil, "on,off", "off"},
		{"true with 3 args", true, "yes,no,unknown", "yes"},
		{"false with 3 args", false, "yes,no,unknown", "no"},
		{"nil with 3 args", nil, "yes,no,unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterYesno(AsValue(tt.input), AsValue(tt.param))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterUrlizeWithAutoescape tests urlize with different autoescape settings
func TestFilterUrlizeWithAutoescape(t *testing.T) {
	t.Run("autoescape true (default)", func(t *testing.T) {
		input := `Check www.example.com/test="value"`
		result, err := filterUrlize(AsValue(input), AsValue(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should contain escaped quotes
		if !strings.Contains(result.String(), "&quot;") {
			t.Errorf("expected escaped quotes, got %q", result.String())
		}
	})

	t.Run("autoescape false", func(t *testing.T) {
		input := `Check www.example.com/test="value"`
		result, err := filterUrlize(AsValue(input), AsValue(false))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should not contain escaped quotes
		if strings.Contains(result.String(), "&quot;") {
			t.Errorf("expected unescaped quotes, got %q", result.String())
		}
	})

	t.Run("non-bool param defaults to autoescape", func(t *testing.T) {
		input := "Check www.example.com"
		result, err := filterUrlize(AsValue(input), AsValue("not a bool"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should work normally
		if !strings.Contains(result.String(), "<a href=") {
			t.Errorf("expected link, got %q", result.String())
		}
	})
}

// TestFilterUrlizetruncEdgeCases tests urlizetrunc edge cases.
// Django uses Unicode ellipsis (…, U+2026) for truncation, not three dots.
func TestFilterUrlizetruncEdgeCases(t *testing.T) {
	t.Run("truncate long URL uses Unicode ellipsis", func(t *testing.T) {
		input := "Visit www.verylongdomainname.com/with/long/path/here"
		result, err := filterUrlizetrunc(AsValue(input), AsValue(15))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should contain truncated title with Unicode ellipsis (…), not three dots
		if !strings.Contains(result.String(), "…") {
			t.Errorf("expected truncated URL with Unicode ellipsis (…), got %q", result.String())
		}
		// Should NOT contain three dots
		if strings.Contains(result.String(), "...") {
			t.Errorf("should use Unicode ellipsis (…) not three dots (...), got %q", result.String())
		}
	})

	t.Run("short URL not truncated", func(t *testing.T) {
		input := "Visit www.ex.com"
		result, err := filterUrlizetrunc(AsValue(input), AsValue(100))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should not contain ellipsis
		if strings.Contains(result.String(), "…") {
			t.Errorf("short URL should not be truncated, got %q", result.String())
		}
	})

	t.Run("truncate email uses Unicode ellipsis", func(t *testing.T) {
		input := "Email verylongusername@verylongdomain.com"
		result, err := filterUrlizetrunc(AsValue(input), AsValue(10))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should contain truncated email with Unicode ellipsis (…), not three dots
		if !strings.Contains(result.String(), "…") {
			t.Errorf("expected truncated email with Unicode ellipsis (…), got %q", result.String())
		}
		// Should NOT contain three dots
		if strings.Contains(result.String(), "...") {
			t.Errorf("should use Unicode ellipsis (…) not three dots (...), got %q", result.String())
		}
	})
}

// TestUnorderedListHelperDeepNesting tests the recursion depth guard
func TestUnorderedListHelperDeepNesting(t *testing.T) {
	// Create deeply nested structure via template
	ts := NewSet("test", &DummyLoader{})

	// Test with a structure that exceeds depth limit
	var createDeep func(depth int) any
	createDeep = func(depth int) any {
		if depth == 0 {
			return []any{"leaf"}
		}
		return []any{createDeep(depth - 1)}
	}

	// Create 150 levels of nesting (exceeds maxUnorderedListDepth of 100)
	deepInput := createDeep(150)

	result, err := filterUnorderedList(AsValue(deepInput), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The function should stop at depth 100 and not panic
	_ = result.String()

	// Also test via template to ensure integration works
	tpl, err := ts.FromString("{{ items|unordered_list }}")
	if err != nil {
		t.Fatalf("template parse error: %v", err)
	}

	_, err = tpl.Execute(Context{"items": deepInput})
	if err != nil {
		t.Fatalf("template execute error: %v", err)
	}
}

// TestUnorderedListWithNestedLists tests unordered_list with various nested structures
func TestUnorderedListWithNestedLists(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "two levels",
			input:    []any{"Item 1", []any{"Sub 1", "Sub 2"}},
			expected: "<li>Item 1<ul><li>Sub 1</li><li>Sub 2</li></ul></li>",
		},
		{
			name:     "list only contains nested list",
			input:    []any{[]any{"A", "B"}},
			expected: "<ul><li>A</li><li>B</li></ul>",
		},
		{
			name:     "mixed nested",
			input:    []any{"Top", []any{"Middle"}, "Bottom"},
			expected: "<li>Top<ul><li>Middle</li></ul></li><li>Bottom</li>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterUnorderedList(AsValue(tt.input), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestUnorderedListWithMap tests unordered_list with a map (non-slice input)
func TestUnorderedListWithMap(t *testing.T) {
	// Maps are not slice/array so they return empty
	input := map[string]string{"a": "value1", "b": "value2"}
	result, err := filterUnorderedList(AsValue(input), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Map is not a slice/array, so output should be empty
	if result.String() != "" {
		t.Errorf("expected empty output for map input, got %q", result.String())
	}
}

// TestFilterWordwrapZeroWidth tests wordwrap with zero or negative width
func TestFilterWordwrapZeroWidth(t *testing.T) {
	input := "one two three"

	t.Run("zero width", func(t *testing.T) {
		result, err := filterWordwrap(AsValue(input), AsValue(0))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should return input unchanged
		if result.String() != input {
			t.Errorf("got %q, want %q", result.String(), input)
		}
	})

	t.Run("negative width", func(t *testing.T) {
		result, err := filterWordwrap(AsValue(input), AsValue(-5))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should return input unchanged
		if result.String() != input {
			t.Errorf("got %q, want %q", result.String(), input)
		}
	})
}

// TestFilterWordwrapCharacterBased verifies that wordwrap wraps at the given
// character width, not word count, matching Django's behavior.
func TestFilterWordwrapCharacterBased(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "basic wrap",
			input:    "Joel is a slug",
			width:    5,
			expected: "Joel\nis a\nslug",
		},
		{
			name:     "fits on one line",
			input:    "Joel is a slug",
			width:    14,
			expected: "Joel is a slug",
		},
		{
			name:     "long word not broken",
			input:    "superlongword short",
			width:    5,
			expected: "superlongword\nshort",
		},
		{
			name:     "preserve newlines",
			input:    "hello\nworld",
			width:    5,
			expected: "hello\nworld",
		},
		{
			name:     "wrap at 10",
			input:    "one two three four",
			width:    10,
			expected: "one two\nthree four",
		},
		{
			name:     "empty string",
			input:    "",
			width:    5,
			expected: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filterWordwrap(AsValue(tc.input), AsValue(tc.width))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tc.expected {
				t.Errorf("wordwrap(%q, %d) = %q, want %q",
					tc.input, tc.width, result.String(), tc.expected)
			}
		})
	}
}

// TestFilterJoinNonSliceable tests join with non-sliceable input
func TestFilterJoinNonSliceable(t *testing.T) {
	result, err := filterJoin(AsValue(42), AsValue(","))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Non-sliceable input is returned unchanged
	if result.Integer() != 42 {
		t.Errorf("got %d, want 42", result.Integer())
	}
}

// TestFilterJoinEmptySeparator tests join with empty separator
func TestFilterJoinEmptySeparator(t *testing.T) {
	result, err := filterJoin(AsValue([]string{"a", "b", "c"}), AsValue(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty separator returns the string representation of the slice
	// The filter returns AsValue(in.String()) which wraps the Value's String()
	// Just verify we get a non-empty result and the code path is exercised
	if result.IsNil() {
		t.Error("expected non-nil result")
	}
}

// TestFilterRandomEmptyInput tests random with empty input
func TestFilterRandomEmptyInput(t *testing.T) {
	result, err := filterRandom(AsValue([]int{}), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Len() != 0 {
		t.Error("expected empty slice to be returned unchanged")
	}
}

// TestFilterRandomNonSliceable tests random with non-sliceable input
func TestFilterRandomNonSliceable(t *testing.T) {
	result, err := filterRandom(AsValue(42), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Integer() != 42 {
		t.Errorf("got %d, want 42", result.Integer())
	}
}

// TestFilterFirstEmpty tests first with empty input
func TestFilterFirstEmpty(t *testing.T) {
	result, err := filterFirst(AsValue([]int{}), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("got %q, want empty string", result.String())
	}
}

// TestFilterLastEmpty tests last with empty input
func TestFilterLastEmpty(t *testing.T) {
	result, err := filterLast(AsValue([]int{}), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("got %q, want empty string", result.String())
	}
}

// TestFilterLinebreaksEmpty tests linebreaks with empty input.
// Django wraps even empty input in paragraph tags: <p></p>
func TestFilterLinebreaksEmpty(t *testing.T) {
	result, err := filterLinebreaks(AsValue(""), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "<p></p>" {
		t.Errorf("got %q, want %q", result.String(), "<p></p>")
	}
}

// TestFilterDivisibleByZero tests divisibleby with zero divisor
func TestFilterDivisibleByZero(t *testing.T) {
	result, err := filterDivisibleby(AsValue(10), AsValue(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("expected false for division by zero")
	}
}

// TestFilterCapfirstEmpty tests capfirst with empty string
func TestFilterCapfirstEmpty(t *testing.T) {
	result, err := filterCapfirst(AsValue(""), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("got %q, want empty string", result.String())
	}
}

// TestFilterTitleNonString tests title with non-string input
func TestFilterTitleNonString(t *testing.T) {
	result, err := filterTitle(AsValue(12345), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("got %q, want empty string", result.String())
	}
}

// TestFilterGetdigitEdgeCases tests get_digit edge cases
func TestFilterGetdigitEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		param    int
		expected string
	}{
		{"position 0", "12345", 0, "12345"},
		{"negative position", "12345", -1, "12345"},
		{"position > length", "12345", 10, "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterGetdigit(AsValue(tt.input), AsValue(tt.param))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterDefaultIfNone tests default_if_none with various inputs
func TestFilterDefaultIfNone(t *testing.T) {
	t.Run("nil returns default", func(t *testing.T) {
		result, err := filterDefaultIfNone(AsValue(nil), AsValue("default"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "default" {
			t.Errorf("got %q, want %q", result.String(), "default")
		}
	})

	t.Run("empty string returns empty string", func(t *testing.T) {
		result, err := filterDefaultIfNone(AsValue(""), AsValue("default"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "" {
			t.Errorf("got %q, want empty string", result.String())
		}
	})

	t.Run("zero returns zero", func(t *testing.T) {
		result, err := filterDefaultIfNone(AsValue(0), AsValue(42))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Integer() != 0 {
			t.Errorf("got %d, want 0", result.Integer())
		}
	})
}

// TestFilterEscapejs tests escapejs with Django's test vectors.
// Test cases from: https://github.com/django/django/blob/main/tests/utils_tests/test_html.py
func TestFilterEscapejs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double and single quotes",
			input:    `"double quotes" and 'single quotes'`,
			expected: `\u0022double quotes\u0022 and \u0027single quotes\u0027`,
		},
		{
			name:     "backslashes",
			input:    `\ : backslashes, too`,
			expected: `\u005C : backslashes, too`,
		},
		{
			name:     "whitespace characters",
			input:    "and lots of whitespace: \r\n\t\v\f\b",
			expected: `and lots of whitespace: \u000D\u000A\u0009\u000B\u000C\u0008`,
		},
		{
			name:     "script tags",
			input:    `<script>and this</script>`,
			expected: `\u003Cscript\u003Eand this\u003C/script\u003E`,
		},
		{
			name:     "line and paragraph separators",
			input:    "paragraph separator:\u2029and line separator:\u2028",
			expected: `paragraph separator:\u2029and line separator:\u2028`,
		},
		{
			name:     "backtick",
			input:    "`",
			expected: `\u0060`,
		},
		{
			name:     "DEL character (0x7F)",
			input:    "\u007f",
			expected: `\u007F`,
		},
		{
			name:     "C1 control start (0x80)",
			input:    "\u0080",
			expected: `\u0080`,
		},
		{
			name:     "C1 control end (0x9F)",
			input:    "\u009f",
			expected: `\u009F`,
		},
		{
			name:     "ampersand",
			input:    "Tom & Jerry",
			expected: `Tom \u0026 Jerry`,
		},
		{
			name:     "equals sign",
			input:    "a=b",
			expected: `a\u003Db`,
		},
		{
			name:     "hyphen/minus",
			input:    "a-b",
			expected: `a\u002Db`,
		},
		{
			name:     "semicolon",
			input:    "a;b",
			expected: `a\u003Bb`,
		},
		{
			name:     "NUL character",
			input:    "a\x00b",
			expected: `a\u0000b`,
		},
		{
			name:     "plain text unchanged",
			input:    "Hello World 123",
			expected: "Hello World 123",
		},
		// Unicode character tests
		{
			name:     "emoji basic",
			input:    "Hello 😀 World",
			expected: "Hello 😀 World",
		},
		{
			name:     "emoji sequence",
			input:    "🎉🎊🎁",
			expected: "🎉🎊🎁",
		},
		{
			name:     "emoji with skin tone modifier",
			input:    "👋🏽",
			expected: "👋🏽",
		},
		{
			name:     "emoji ZWJ sequence (family)",
			input:    "👨‍👩‍👧‍👦",
			expected: "👨‍👩‍👧‍👦",
		},
		{
			name:     "Chinese characters",
			input:    "你好世界",
			expected: "你好世界",
		},
		{
			name:     "Japanese hiragana and katakana",
			input:    "こんにちはカタカナ",
			expected: "こんにちはカタカナ",
		},
		{
			name:     "Korean hangul",
			input:    "안녕하세요",
			expected: "안녕하세요",
		},
		{
			name:     "Arabic text",
			input:    "مرحبا بالعالم",
			expected: "مرحبا بالعالم",
		},
		{
			name:     "Hebrew text",
			input:    "שלום עולם",
			expected: "שלום עולם",
		},
		{
			name:     "Thai text",
			input:    "สวัสดีโลก",
			expected: "สวัสดีโลก",
		},
		{
			name:     "Greek text",
			input:    "Γειά σου κόσμε",
			expected: "Γειά σου κόσμε",
		},
		{
			name:     "Cyrillic text",
			input:    "Привет мир",
			expected: "Привет мир",
		},
		{
			name:     "combining characters (e with acute)",
			input:    "café",
			expected: "café",
		},
		{
			name:     "combining diacritical marks",
			input:    "a\u0301\u0327", // a + combining acute + combining cedilla
			expected: "a\u0301\u0327",
		},
		{
			name:     "mixed scripts with special chars",
			input:    `日本語 & "English" <test>`,
			expected: `日本語 \u0026 \u0022English\u0022 \u003Ctest\u003E`,
		},
		{
			name:     "mathematical symbols",
			input:    "∑∏∫∂∆",
			expected: "∑∏∫∂∆",
		},
		{
			name:     "currency symbols",
			input:    "€£¥₹₽",
			expected: "€£¥₹₽",
		},
		{
			name:     "box drawing characters",
			input:    "┌─┐│└─┘",
			expected: "┌─┐│└─┘",
		},
		{
			name:     "musical symbols",
			input:    "♩♪♫♬",
			expected: "♩♪♫♬",
		},
		{
			name:     "zero width joiner",
			input:    "a\u200Db",
			expected: "a\u200Db",
		},
		{
			name:     "zero width non-joiner",
			input:    "a\u200Cb",
			expected: "a\u200Cb",
		},
		{
			name:     "byte order mark",
			input:    "\uFEFFtext",
			expected: "\uFEFFtext",
		},
		{
			name:     "right-to-left override",
			input:    "\u202Etext",
			expected: "\u202Etext",
		},
		{
			name:     "private use area",
			input:    "\uE000\uE001",
			expected: "\uE000\uE001",
		},
		{
			name:     "supplementary plane character (Gothic letter)",
			input:    "𐌰𐌱𐌲",
			expected: "𐌰𐌱𐌲",
		},
		{
			name:     "emoji in supplementary plane",
			input:    "𝄞", // Musical G clef
			expected: "𝄞",
		},
		{
			name:     "all C0 control characters",
			input:    "\x01\x02\x03\x04\x05\x06\x07\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f",
			expected: `\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u000E\u000F\u0010\u0011\u0012\u0013\u0014\u0015\u0016\u0017\u0018\u0019\u001A\u001B\u001C\u001D\u001E\u001F`,
		},
		{
			name:     "Unicode emoji with special chars",
			input:    `<script>alert("🔥")</script>`,
			expected: `\u003Cscript\u003Ealert(\u0022🔥\u0022)\u003C/script\u003E`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterEscapejs(AsValue(tt.input), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterEscapejsEscapeSequences tests pongo2-specific \r and \n escape sequence handling.
// pongo2 interprets literal \r and \n in input strings as escape sequences.
func TestFilterEscapejsEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "literal backslash-r",
			input:    `line1\rline2`,
			expected: `line1\u000Dline2`,
		},
		{
			name:     "literal backslash-n",
			input:    `line1\nline2`,
			expected: `line1\u000Aline2`,
		},
		{
			name:     "both escape sequences",
			input:    `\r\n`,
			expected: `\u000D\u000A`,
		},
		{
			name:     "backslash followed by other char",
			input:    `\t`,
			expected: `\u005Ct`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterEscapejs(AsValue(tt.input), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterTruncatewordsZero tests truncatewords with zero words
func TestFilterTruncatewordsZero(t *testing.T) {
	result, err := filterTruncatewords(AsValue("hello world"), AsValue(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("got %q, want empty string", result.String())
	}
}

// TestFilterTruncatewordsNegative tests truncatewords with negative count
func TestFilterTruncatewordsNegative(t *testing.T) {
	result, err := filterTruncatewords(AsValue("hello world"), AsValue(-5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("got %q, want empty string", result.String())
	}
}

// TestFilterTruncatecharsZero tests truncatechars with zero length.
// Django returns just the ellipsis for length <= 0 when the string is longer.
func TestFilterTruncatecharsZero(t *testing.T) {
	result, err := filterTruncatechars(AsValue("hello"), AsValue(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "…" {
		t.Errorf("got %q, want %q", result.String(), "…")
	}
}

// TestFilterTruncatecharsLengthOne tests truncatechars with length = 1
func TestFilterTruncatecharsLengthOne(t *testing.T) {
	result, err := filterTruncatechars(AsValue("hello"), AsValue(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Length 1 means just ellipsis (using proper ellipsis character "…")
	if result.String() != "…" {
		t.Errorf("got %q, want %q", result.String(), "…")
	}
}

// TestFilterTruncatecharsEllipsis tests that truncatechars uses proper ellipsis character
func TestFilterTruncatecharsEllipsis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		contains string
	}{
		{
			name:     "truncated string has ellipsis",
			input:    "Hello World",
			length:   8,
			contains: "…", // Unicode ellipsis U+2026
		},
		{
			name:     "short string not truncated",
			input:    "Hi",
			length:   10,
			contains: "Hi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterTruncatechars(AsValue(tt.input), AsValue(tt.length))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result.String(), tt.contains) {
				t.Errorf("got %q, want to contain %q", result.String(), tt.contains)
			}
		})
	}

	// Verify we're NOT using three dots
	t.Run("not using three dots", func(t *testing.T) {
		result, err := filterTruncatechars(AsValue("Hello World"), AsValue(8))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(result.String(), "...") {
			t.Errorf("should use ellipsis (…) not three dots (...), got %q", result.String())
		}
	})
}

// TestFilterRjustSmallerThanInput tests rjust when width is smaller than input
func TestFilterRjustSmallerThanInput(t *testing.T) {
	result, err := filterRjust(AsValue("hello"), AsValue(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Width is 3, but input is 5 chars, so it should still be formatted
	// The fmt.Sprintf with %3s will not truncate, but won't pad either
	if result.String() != "hello" {
		t.Errorf("got %q, want %q", result.String(), "hello")
	}
}

// TestFilterLjustSmallerThanInput tests ljust when width is smaller than input
func TestFilterLjustSmallerThanInput(t *testing.T) {
	result, err := filterLjust(AsValue("hello"), AsValue(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Width is 3, input is 5, no padding needed
	if result.String() != "hello" {
		t.Errorf("got %q, want %q", result.String(), "hello")
	}
}

// TestFilterCenterSmallerThanInput tests center when width is smaller than input
func TestFilterCenterSmallerThanInput(t *testing.T) {
	result, err := filterCenter(AsValue("hello"), AsValue(3))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Width is 3, input is 5, return unchanged
	if result.String() != "hello" {
		t.Errorf("got %q, want %q", result.String(), "hello")
	}
}

// TestFilterCenterPaddingDirection verifies that center filter puts extra space
// on the right when padding is odd, matching Django's str.center() behavior.
func TestFilterCenterPaddingDirection(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"test", 19, "       test        "},   // 7 left, 8 right (extra on right)
		{"test", 20, "        test        "},  // 8 left, 8 right (even)
		{"test2", 19, "       test2       "},  // 7 left, 7 right (even)
		{"test2", 20, "       test2        "}, // 7 left, 8 right (extra on right)
		{"x", 4, " x  "},                      // 1 left, 2 right (extra on right)
		{"ab", 5, " ab  "},                    // 1 left, 2 right (extra on right)
	}
	for _, tc := range tests {
		result, err := filterCenter(AsValue(tc.input), AsValue(tc.width))
		if err != nil {
			t.Fatalf("unexpected error for center(%q, %d): %v", tc.input, tc.width, err)
		}
		if result.String() != tc.expected {
			t.Errorf("center(%q, %d) = %q (left=%d, right=%d), want %q (left=%d, right=%d)",
				tc.input, tc.width,
				result.String(),
				len(result.String())-len(strings.TrimLeft(result.String(), " ")),
				len(result.String())-len(strings.TrimRight(result.String(), " ")),
				tc.expected,
				len(tc.expected)-len(strings.TrimLeft(tc.expected, " ")),
				len(tc.expected)-len(strings.TrimRight(tc.expected, " ")),
			)
		}
	}
}

// TestFilterJSONScript tests the json_script filter comprehensively
func TestFilterJSONScript(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		elementID string
		expected  string
	}{
		{
			name:      "simple string",
			input:     "hello",
			elementID: "my-data",
			expected:  `<script id="my-data" type="application/json">"hello"</script>`,
		},
		{
			name:      "integer",
			input:     42,
			elementID: "num-data",
			expected:  `<script id="num-data" type="application/json">42</script>`,
		},
		{
			name:      "float",
			input:     3.14159,
			elementID: "float-data",
			expected:  `<script id="float-data" type="application/json">3.14159</script>`,
		},
		{
			name:      "boolean true",
			input:     true,
			elementID: "bool-data",
			expected:  `<script id="bool-data" type="application/json">true</script>`,
		},
		{
			name:      "boolean false",
			input:     false,
			elementID: "bool-data",
			expected:  `<script id="bool-data" type="application/json">false</script>`,
		},
		{
			name:      "null/nil",
			input:     nil,
			elementID: "nil-data",
			expected:  `<script id="nil-data" type="application/json">null</script>`,
		},
		{
			name:      "simple map",
			input:     map[string]any{"key": "value"},
			elementID: "map-data",
			expected:  `<script id="map-data" type="application/json">{"key":"value"}</script>`,
		},
		{
			name:      "simple slice",
			input:     []string{"a", "b", "c"},
			elementID: "slice-data",
			expected:  `<script id="slice-data" type="application/json">["a","b","c"]</script>`,
		},
		{
			name:      "element ID with spaces",
			input:     "test",
			elementID: "my data",
			expected:  `<script id="my data" type="application/json">"test"</script>`,
		},
		{
			name:      "element ID with hyphens and underscores",
			input:     "test",
			elementID: "my-data_123",
			expected:  `<script id="my-data_123" type="application/json">"test"</script>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterJSONScript(AsValue(tt.input), AsValue(tt.elementID))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterJSONScriptXSSPrevention tests XSS prevention in json_script
func TestFilterJSONScriptXSSPrevention(t *testing.T) {
	tests := []struct {
		name          string
		input         any
		elementID     string
		shouldHave    []string
		shouldNotHave []string
	}{
		{
			name:          "element ID with quotes is escaped",
			input:         "test",
			elementID:     `my"data`,
			shouldHave:    []string{`id="my&quot;data"`},
			shouldNotHave: []string{`id="my"data"`},
		},
		{
			name:          "JSON with script tag is escaped",
			input:         "</script><script>alert('xss')</script>",
			elementID:     "safe-data",
			shouldHave:    []string{`\u003c/script\u003e`, `\u003cscript\u003e`},
			shouldNotHave: []string{`</script><script>`},
		},
		{
			name:          "JSON with less than is escaped",
			input:         "a < b",
			elementID:     "data",
			shouldHave:    []string{`\u003c`},
			shouldNotHave: []string{`"a < b"`},
		},
		{
			name:          "JSON with greater than is escaped",
			input:         "a > b",
			elementID:     "data",
			shouldHave:    []string{`\u003e`},
			shouldNotHave: []string{`"a > b"`},
		},
		{
			name:          "JSON with ampersand is escaped",
			input:         "a & b",
			elementID:     "data",
			shouldHave:    []string{`\u0026`},
			shouldNotHave: []string{`"a & b"`},
		},
		{
			name:          "multiple XSS vectors in one value",
			input:         "<script>alert('xss')</script>&<div>",
			elementID:     "xss-test",
			shouldHave:    []string{`\u003cscript\u003e`, `\u0026`, `\u003cdiv\u003e`},
			shouldNotHave: []string{`<script>`, `&`, `<div>`},
		},
		{
			name:          "XSS in map values",
			input:         map[string]string{"html": "<script>evil()</script>"},
			elementID:     "map-xss",
			shouldHave:    []string{`\u003cscript\u003e`},
			shouldNotHave: []string{`<script>`},
		},
		{
			name:          "XSS in slice elements",
			input:         []string{"<img src=x onerror=alert(1)>"},
			elementID:     "arr-xss",
			shouldHave:    []string{`\u003cimg`},
			shouldNotHave: []string{`<img`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterJSONScript(AsValue(tt.input), AsValue(tt.elementID))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			resultStr := result.String()

			for _, shouldHave := range tt.shouldHave {
				if !strings.Contains(resultStr, shouldHave) {
					t.Errorf("result should contain %q, got %q", shouldHave, resultStr)
				}
			}

			for _, shouldNotHave := range tt.shouldNotHave {
				if strings.Contains(resultStr, shouldNotHave) {
					t.Errorf("result should not contain %q, got %q", shouldNotHave, resultStr)
				}
			}
		})
	}
}

// TestFilterJSONScriptOptionalID tests json_script without element_id (Django 4.1+)
func TestFilterJSONScriptOptionalID(t *testing.T) {
	t.Run("nil element_id outputs script without id", func(t *testing.T) {
		result, err := filterJSONScript(AsValue("test"), AsValue(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `<script type="application/json">"test"</script>`
		if result.String() != expected {
			t.Errorf("got %q, want %q", result.String(), expected)
		}
		// Verify no id attribute
		if strings.Contains(result.String(), "id=") {
			t.Error("output should not contain id attribute when element_id is nil")
		}
	})

	t.Run("empty string element_id outputs script without id", func(t *testing.T) {
		result, err := filterJSONScript(AsValue("test"), AsValue(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `<script type="application/json">"test"</script>`
		if result.String() != expected {
			t.Errorf("got %q, want %q", result.String(), expected)
		}
	})

	t.Run("complex value without id", func(t *testing.T) {
		input := map[string]any{"key": "value", "num": 42}
		result, err := filterJSONScript(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(result.String(), `<script type="application/json">`) {
			t.Errorf("result should start with script tag without id, got %q", result.String())
		}
		if !strings.HasSuffix(result.String(), "</script>") {
			t.Errorf("result should end with </script>, got %q", result.String())
		}
		if strings.Contains(result.String(), "id=") {
			t.Error("output should not contain id attribute")
		}
	})
}

// TestFilterJSONScriptErrors tests error cases for json_script
func TestFilterJSONScriptErrors(t *testing.T) {
	t.Run("unmarshalable value - channel", func(t *testing.T) {
		ch := make(chan int)
		result, err := filterJSONScript(AsValue(ch), AsValue("data"))
		if err == nil {
			t.Error("expected error for unmarshalable value")
		}
		if result != nil {
			t.Error("expected nil result on error")
		}
		if !strings.Contains(err.Error(), "json marshalling error") {
			t.Errorf("error should mention json marshalling, got %v", err)
		}
	})

	t.Run("unmarshalable value - function", func(t *testing.T) {
		fn := func() {}
		result, err := filterJSONScript(AsValue(fn), AsValue("data"))
		if err == nil {
			t.Error("expected error for unmarshalable function")
		}
		if result != nil {
			t.Error("expected nil result on error")
		}
	})
}

// TestFilterJSONScriptComplexTypes tests json_script with complex Go types
func TestFilterJSONScriptComplexTypes(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type Config struct {
		Debug   bool           `json:"debug"`
		Timeout int            `json:"timeout"`
		Tags    []string       `json:"tags"`
		Meta    map[string]any `json:"meta"`
	}

	tests := []struct {
		name      string
		input     any
		elementID string
		contains  []string
	}{
		{
			name:      "struct with json tags",
			input:     Person{Name: "Alice", Age: 30},
			elementID: "person",
			contains:  []string{`"name":"Alice"`, `"age":30`},
		},
		{
			name:      "pointer to struct",
			input:     &Person{Name: "Bob", Age: 25},
			elementID: "person-ptr",
			contains:  []string{`"name":"Bob"`, `"age":25`},
		},
		{
			name: "complex nested struct",
			input: Config{
				Debug:   true,
				Timeout: 30,
				Tags:    []string{"api", "v2"},
				Meta:    map[string]any{"version": "1.0"},
			},
			elementID: "config",
			contains:  []string{`"debug":true`, `"timeout":30`, `"tags":["api","v2"]`, `"version":"1.0"`},
		},
		{
			name:      "slice of structs",
			input:     []Person{{Name: "A", Age: 1}, {Name: "B", Age: 2}},
			elementID: "people",
			contains:  []string{`[{`, `"name":"A"`, `"name":"B"`},
		},
		{
			name:      "map with struct values",
			input:     map[string]Person{"first": {Name: "First", Age: 1}},
			elementID: "map-struct",
			contains:  []string{`"first":{`, `"name":"First"`},
		},
		{
			name:      "empty slice",
			input:     []string{},
			elementID: "empty-slice",
			contains:  []string{`[]`},
		},
		{
			name:      "empty map",
			input:     map[string]string{},
			elementID: "empty-map",
			contains:  []string{`{}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterJSONScript(AsValue(tt.input), AsValue(tt.elementID))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			resultStr := result.String()

			for _, c := range tt.contains {
				if !strings.Contains(resultStr, c) {
					t.Errorf("result should contain %q, got %q", c, resultStr)
				}
			}

			// Verify proper script tag structure
			if !strings.HasPrefix(resultStr, `<script id="`) {
				t.Errorf("result should start with <script id=\", got %q", resultStr)
			}
			if !strings.HasSuffix(resultStr, "</script>") {
				t.Errorf("result should end with </script>, got %q", resultStr)
			}
			if !strings.Contains(resultStr, `type="application/json"`) {
				t.Errorf("result should contain type=\"application/json\", got %q", resultStr)
			}
		})
	}
}

// TestFilterJSONScriptViaTemplate tests json_script through template execution
func TestFilterJSONScriptViaTemplate(t *testing.T) {
	ts := NewSet("test", &DummyLoader{})

	tests := []struct {
		name     string
		template string
		context  Context
		contains []string
	}{
		{
			name:     "simple variable",
			template: `{{ data|json_script:"my-data" }}`,
			context:  Context{"data": map[string]string{"key": "value"}},
			contains: []string{`<script id="my-data"`, `{"key":"value"}`},
		},
		{
			name:     "string literal",
			template: `{{ "hello"|json_script:"greeting" }}`,
			context:  nil,
			contains: []string{`<script id="greeting"`, `"hello"`},
		},
		{
			name:     "number",
			template: `{{ 42|json_script:"number" }}`,
			context:  nil,
			contains: []string{`<script id="number"`, `42`},
		},
		{
			name:     "xss vector through template",
			template: `{{ html|json_script:"safe" }}`,
			context:  Context{"html": "<script>alert(1)</script>"},
			contains: []string{`\u003cscript\u003e`},
		},
		{
			name:     "without element_id (Django 4.1+)",
			template: `{{ data|json_script }}`,
			context:  Context{"data": map[string]string{"hello": "world"}},
			contains: []string{`<script type="application/json">`, `{"hello":"world"}`, `</script>`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := ts.FromString(tt.template)
			if err != nil {
				t.Fatalf("template parse error: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("template execute error: %v", err)
			}

			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("result should contain %q, got %q", c, result)
				}
			}
		})
	}
}

// TestFilterJSONScriptOutputIsSafe tests that json_script output is marked as safe
func TestFilterJSONScriptOutputIsSafe(t *testing.T) {
	// The output should not be double-escaped when rendered in a template
	ts := NewSet("test", &DummyLoader{})

	template := `{{ data|json_script:"test" }}`
	context := Context{"data": "hello"}

	tpl, err := ts.FromString(template)
	if err != nil {
		t.Fatalf("template parse error: %v", err)
	}

	result, err := tpl.Execute(context)
	if err != nil {
		t.Fatalf("template execute error: %v", err)
	}

	// The < and > in script tags should NOT be escaped
	if strings.Contains(result, "&lt;script") {
		t.Error("script tag should not be HTML-escaped in output")
	}
	if !strings.Contains(result, "<script") {
		t.Error("output should contain literal <script tag")
	}
}

// TestFilterJSONScriptUnicodeHandling tests proper Unicode handling
func TestFilterJSONScriptUnicodeHandling(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		elementID string
	}{
		{
			name:      "chinese characters",
			input:     "你好世界",
			elementID: "chinese",
		},
		{
			name:      "emoji",
			input:     "Hello 👋 World 🌍",
			elementID: "emoji",
		},
		{
			name:      "mixed unicode",
			input:     "Ñoño日本語한국어",
			elementID: "mixed",
		},
		{
			name:      "unicode in element ID",
			input:     "test",
			elementID: "データ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterJSONScript(AsValue(tt.input), AsValue(tt.elementID))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify the output is valid
			if result == nil {
				t.Error("expected non-nil result")
			}

			// The result should contain our element ID
			if !strings.Contains(result.String(), tt.elementID) {
				t.Errorf("result should contain element ID %q, got %q", tt.elementID, result.String())
			}
		})
	}
}

// TestFilterJSONScriptSpecialJSONValues tests special JSON values
func TestFilterJSONScriptSpecialJSONValues(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: `<script id="data" type="application/json">""</script>`,
		},
		{
			name:     "zero",
			input:    0,
			expected: `<script id="data" type="application/json">0</script>`,
		},
		{
			name:     "negative number",
			input:    -42,
			expected: `<script id="data" type="application/json">-42</script>`,
		},
		{
			name:     "very large number",
			input:    9999999999999999,
			expected: `<script id="data" type="application/json">9999999999999999</script>`,
		},
		{
			name:     "floating point",
			input:    0.123456789,
			expected: `<script id="data" type="application/json">0.123456789</script>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterJSONScript(AsValue(tt.input), AsValue("data"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterSafeseq tests the safeseq filter
func TestFilterSafeseq(t *testing.T) {
	t.Run("string slice", func(t *testing.T) {
		input := []string{"<b>bold</b>", "<i>italic</i>"}
		result, err := filterSafeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Result should be a slice of Values
		if !result.CanSlice() {
			t.Error("result should be sliceable")
		}
		if result.Len() != 2 {
			t.Errorf("result length = %d, want 2", result.Len())
		}
	})

	t.Run("values are marked safe", func(t *testing.T) {
		input := []string{"<script>alert('xss')</script>"}

		// When used in template, safe values should not be escaped
		ts := NewSet("test", &DummyLoader{})
		tpl, err := ts.FromString(`{% for item in items %}{{ item }}{% endfor %}`)
		if err != nil {
			t.Fatalf("template parse error: %v", err)
		}

		// Apply safeseq and use result
		safeItems, err := filterSafeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("filterSafeseq error: %v", err)
		}
		out, err := tpl.Execute(Context{"items": safeItems.Interface()})
		if err != nil {
			t.Fatalf("template execute error: %v", err)
		}

		// Should NOT be escaped because items are marked safe
		if strings.Contains(out, "&lt;") {
			t.Errorf("safe values should not be escaped, got %q", out)
		}
		if !strings.Contains(out, "<script>") {
			t.Errorf("output should contain literal <script>, got %q", out)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []string{}
		result, err := filterSafeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 0 {
			t.Errorf("result length = %d, want 0", result.Len())
		}
	})

	t.Run("int slice", func(t *testing.T) {
		input := []int{1, 2, 3}
		result, err := filterSafeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 3 {
			t.Errorf("result length = %d, want 3", result.Len())
		}
	})

	t.Run("non-sliceable input returns unchanged", func(t *testing.T) {
		input := 42 // integers are not sliceable
		result, err := filterSafeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Integer() != 42 {
			t.Errorf("non-sliceable input should be returned unchanged, got %v", result.Interface())
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result, err := filterSafeseq(AsValue(nil), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsNil() {
			t.Error("nil input should return nil-ish result")
		}
	})

	t.Run("mixed type slice", func(t *testing.T) {
		input := []any{"<b>text</b>", 42, true}
		result, err := filterSafeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 3 {
			t.Errorf("result length = %d, want 3", result.Len())
		}
	})
}

// TestFilterEscapeseq tests the escapeseq filter
func TestFilterEscapeseq(t *testing.T) {
	t.Run("string slice with HTML", func(t *testing.T) {
		input := []string{"<b>bold</b>", "<i>italic</i>"}
		result, err := filterEscapeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.CanSlice() {
			t.Error("result should be sliceable")
		}
		if result.Len() != 2 {
			t.Errorf("result length = %d, want 2", result.Len())
		}

		// Check that HTML is escaped
		first := result.Index(0).String()
		if !strings.Contains(first, "&lt;b&gt;") {
			t.Errorf("first element should be escaped, got %q", first)
		}
	})

	t.Run("XSS prevention", func(t *testing.T) {
		input := []string{"<script>alert('xss')</script>"}
		result, err := filterEscapeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		escaped := result.Index(0).String()
		if strings.Contains(escaped, "<script>") {
			t.Errorf("script tag should be escaped, got %q", escaped)
		}
		if !strings.Contains(escaped, "&lt;script&gt;") {
			t.Errorf("script tag should be escaped to &lt;script&gt;, got %q", escaped)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []string{}
		result, err := filterEscapeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 0 {
			t.Errorf("result length = %d, want 0", result.Len())
		}
	})

	t.Run("non-sliceable input returns unchanged", func(t *testing.T) {
		input := 42 // integers are not sliceable
		result, err := filterEscapeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Integer() != 42 {
			t.Errorf("non-sliceable input should be returned unchanged, got %v", result.Interface())
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result, err := filterEscapeseq(AsValue(nil), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsNil() {
			t.Error("nil input should return nil-ish result")
		}
	})

	t.Run("int slice converts to escaped strings", func(t *testing.T) {
		input := []int{1, 2, 3}
		result, err := filterEscapeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 3 {
			t.Errorf("result length = %d, want 3", result.Len())
		}
		// Integers don't contain HTML so should be unchanged
		if result.Index(0).String() != "1" {
			t.Errorf("first element = %q, want \"1\"", result.Index(0).String())
		}
	})

	t.Run("special HTML characters", func(t *testing.T) {
		input := []string{
			"<",
			">",
			"&",
			"\"",
			"'",
		}
		result, err := filterEscapeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []string{
			"&lt;",
			"&gt;",
			"&amp;",
			"&quot;",
			"&#39;",
		}

		for i, exp := range expected {
			got := result.Index(i).String()
			if got != exp {
				t.Errorf("element %d = %q, want %q", i, got, exp)
			}
		}
	})

	t.Run("already safe content", func(t *testing.T) {
		input := []string{"plain text", "hello world"}
		result, err := filterEscapeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Plain text should remain unchanged
		if result.Index(0).String() != "plain text" {
			t.Errorf("plain text should remain unchanged, got %q", result.Index(0).String())
		}
	})

	t.Run("mixed content", func(t *testing.T) {
		input := []any{"<b>bold</b>", 42, "<script>"}
		result, err := filterEscapeseq(AsValue(input), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 3 {
			t.Errorf("result length = %d, want 3", result.Len())
		}

		// First should be escaped HTML
		if !strings.Contains(result.Index(0).String(), "&lt;b&gt;") {
			t.Errorf("HTML should be escaped, got %q", result.Index(0).String())
		}
		// Second should be number as string
		if result.Index(1).String() != "42" {
			t.Errorf("number should be \"42\", got %q", result.Index(1).String())
		}
		// Third should be escaped
		if !strings.Contains(result.Index(2).String(), "&lt;script&gt;") {
			t.Errorf("script tag should be escaped, got %q", result.Index(2).String())
		}
	})
}

// TestFilterSafeseqViaTemplate tests safeseq through template execution
func TestFilterSafeseqViaTemplate(t *testing.T) {
	ts := NewSet("test", &DummyLoader{})

	tests := []struct {
		name     string
		template string
		context  Context
		contains string
		excludes string
	}{
		{
			name:     "HTML not escaped with safeseq",
			template: `{% for item in items|safeseq %}{{ item }}{% endfor %}`,
			context:  Context{"items": []string{"<b>bold</b>"}},
			contains: "<b>bold</b>",
			excludes: "&lt;",
		},
		{
			name:     "multiple items",
			template: `{% for item in items|safeseq %}[{{ item }}]{% endfor %}`,
			context:  Context{"items": []string{"<a>", "<b>"}},
			contains: "[<a>][<b>]",
			excludes: "&lt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := ts.FromString(tt.template)
			if err != nil {
				t.Fatalf("template parse error: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("template execute error: %v", err)
			}

			if !strings.Contains(result, tt.contains) {
				t.Errorf("result should contain %q, got %q", tt.contains, result)
			}
			if tt.excludes != "" && strings.Contains(result, tt.excludes) {
				t.Errorf("result should not contain %q, got %q", tt.excludes, result)
			}
		})
	}
}

// TestFilterEscapeseqViaTemplate tests escapeseq through template execution
func TestFilterEscapeseqViaTemplate(t *testing.T) {
	ts := NewSet("test", &DummyLoader{})

	tests := []struct {
		name     string
		template string
		context  Context
		contains string
		excludes string
	}{
		{
			name:     "HTML escaped with escapeseq and safe",
			template: `{% for item in items|escapeseq %}{{ item|safe }}{% endfor %}`,
			context:  Context{"items": []string{"<b>bold</b>"}},
			contains: "&lt;b&gt;bold&lt;/b&gt;",
			excludes: "<b>",
		},
		{
			name:     "script tag escaped",
			template: `{% for item in items|escapeseq %}{{ item|safe }}{% endfor %}`,
			context:  Context{"items": []string{"<script>alert('xss')</script>"}},
			contains: "&lt;script&gt;",
			excludes: "<script>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := ts.FromString(tt.template)
			if err != nil {
				t.Fatalf("template parse error: %v", err)
			}

			result, err := tpl.Execute(tt.context)
			if err != nil {
				t.Fatalf("template execute error: %v", err)
			}

			if !strings.Contains(result, tt.contains) {
				t.Errorf("result should contain %q, got %q", tt.contains, result)
			}
			if tt.excludes != "" && strings.Contains(result, tt.excludes) {
				t.Errorf("result should not contain %q, got %q", tt.excludes, result)
			}
		})
	}
}

// TestFilterStriptags tests striptags filter with Django test vectors and edge cases.
// The filter uses a regex that handles quoted attributes containing >, and applies
// stripping recursively to handle obfuscated tags.
func TestFilterStriptags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Django test vectors
		{
			name:     "basic paragraph with entities",
			input:    "<p>See: &#39;&eacute; is an apostrophe followed by e acute</p>",
			expected: "See: &#39;&eacute; is an apostrophe followed by e acute",
		},
		{
			name:     "unclosed tag at start",
			input:    "<adf>a",
			expected: "a",
		},
		{
			name:     "closing tag without opener",
			input:    "</adf>a",
			expected: "a",
		},
		{
			name:     "nested unclosed tags",
			input:    "<asdf><asdf>e",
			expected: "e",
		},
		{
			name:     "incomplete tag not stripped",
			input:    "hi, <f x",
			expected: "hi, <f x",
		},
		{
			name:     "less-than comparison not stripped",
			input:    "234<235, right?",
			expected: "234<235, right?",
		},
		{
			name:     "single char between tags",
			input:    "<x>b<y>",
			expected: "b",
		},
		{
			name:     "onclick with embedded quotes",
			input:    "a<p onclick=\"alert('<test>')\">b</p>c",
			expected: "abc",
		},
		{
			name:     "adjacent tags no space",
			input:    "<strong>foo</strong><a href=\"...\">bar</a>",
			expected: "foobar",
		},
		{
			name:     "ampersand and entities preserved",
			input:    "&gotcha&#;<>",
			expected: "&gotcha&#;<>",
		},
		// Additional edge cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no tags",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "self-closing tags",
			input:    "Hello<br/>World<hr />!",
			expected: "HelloWorld!",
		},
		{
			name:     "multiple spaces in tag",
			input:    "<div   class=\"foo\"  >content</div>",
			expected: "content",
		},
		{
			name:     "newlines in tag",
			input:    "<div\nclass=\"foo\"\n>content</div>",
			expected: "content",
		},
		{
			name:     "tabs in tag",
			input:    "<div\tclass=\"foo\">content</div>",
			expected: "content",
		},
		{
			name:     "deeply nested tags",
			input:    "<div><span><b><i>text</i></b></span></div>",
			expected: "text",
		},
		{
			name:     "tag with many attributes",
			input:    `<a href="url" class="cls" id="id" data-x="y">link</a>`,
			expected: "link",
		},
		{
			name:     "uppercase tags",
			input:    "<DIV>content</DIV>",
			expected: "content",
		},
		{
			name:     "mixed case tags",
			input:    "<DiV>content</dIv>",
			expected: "content",
		},
		{
			name:     "script tag",
			input:    "<script>alert('xss')</script>",
			expected: "alert('xss')",
		},
		{
			name:     "style tag",
			input:    "<style>.foo { color: red; }</style>",
			expected: ".foo { color: red; }",
		},
		{
			name:     "html comment",
			input:    "before<!-- comment -->after",
			expected: "beforeafter",
		},
		{
			name:     "multiline comment",
			input:    "a<!--\nmultiline\ncomment\n-->b",
			expected: "ab",
		},
		{
			name:     "DOCTYPE",
			input:    "<!DOCTYPE html><html>content</html>",
			expected: "content",
		},
		{
			name:     "CDATA section",
			input:    "<![CDATA[some data]]>text",
			expected: "text",
		},
		{
			name:     "XML processing instruction",
			input:    "<?xml version=\"1.0\"?>content",
			expected: "content",
		},
		{
			name:     "unicode content preserved",
			input:    "<p>你好世界</p>",
			expected: "你好世界",
		},
		{
			name:     "emoji preserved",
			input:    "<span>Hello 😀 World</span>",
			expected: "Hello 😀 World",
		},
		{
			name:     "RTL text preserved",
			input:    "<div>مرحبا</div>",
			expected: "مرحبا",
		},
		{
			name:     "tag inside attribute value",
			input:    `<a title="<b>bold</b>">link</a>`,
			expected: "link",
		},
		{
			name:     "greater-than in attribute",
			input:    `<div data-value="a>b">content</div>`,
			expected: "content",
		},
		{
			name:     "single quotes in attribute",
			input:    "<div class='foo'>content</div>",
			expected: "content",
		},
		{
			name:     "unquoted attribute",
			input:    "<div class=foo>content</div>",
			expected: "content",
		},
		{
			name:     "multiple tags on line",
			input:    "<b>a</b> <i>b</i> <u>c</u>",
			expected: "a b c",
		},
		{
			name:     "whitespace between tags trimmed",
			input:    "  <p>text</p>  ",
			expected: "text",
		},
		{
			name:     "void elements",
			input:    "a<br>b<hr>c<img src='x'>d",
			expected: "abcd",
		},
		{
			name:     "SVG tag",
			input:    "<svg><circle cx=\"50\"/></svg>text",
			expected: "text",
		},
		{
			name:     "math tag",
			input:    "<math><mi>x</mi></math>",
			expected: "x",
		},
		// Security-sensitive edge cases
		{
			name:     "null byte in tag name",
			input:    "<scr\x00ipt>alert(1)</script>",
			expected: "alert(1)",
		},
		{
			name:     "null byte before tag",
			input:    "\x00<script>x</script>",
			expected: "x",
		},
		{
			name:     "encoded angle brackets",
			input:    "&lt;script&gt;alert(1)&lt;/script&gt;",
			expected: "&lt;script&gt;alert(1)&lt;/script&gt;",
		},
		{
			name:     "double encoded not decoded",
			input:    "&amp;lt;script&amp;gt;",
			expected: "&amp;lt;script&amp;gt;",
		},
		{
			name:     "incomplete close tag",
			input:    "text</",
			expected: "text</",
		},
		{
			name:     "incomplete open tag at end",
			input:    "text<div",
			expected: "text<div",
		},
		{
			name:     "just angle brackets",
			input:    "<>",
			expected: "<>",
		},
		{
			name:     "angle bracket space",
			input:    "< >",
			expected: "< >",
		},
		{
			name:     "multiple less-than",
			input:    "a<<b",
			expected: "a<<b",
		},
		{
			name:     "multiple greater-than",
			input:    "a>>b",
			expected: "a>>b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterStriptags(AsValue(tt.input), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

// TestFilterStriptagsDjangoCompatibility verifies behavior matches Django's strip_tags.
// Test vectors verified against Django's HTMLParser-based implementation.
// Some obfuscation patterns leave partial content - this is expected Django behavior.
func TestFilterStriptagsDjangoCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		comment  string
	}{
		// Django test vectors from tests/utils_tests/test_html.py
		{
			name:     "paragraph with entities",
			input:    "<p>See: &#39;&eacute; is an apostrophe followed by e acute</p>",
			expected: "See: &#39;&eacute; is an apostrophe followed by e acute",
			comment:  "entities preserved",
		},
		{
			name:     "unclosed tag at start",
			input:    "<adf>a",
			expected: "a",
			comment:  "unclosed tag stripped",
		},
		{
			name:     "closing tag without opener",
			input:    "</adf>a",
			expected: "a",
			comment:  "orphan closing tag stripped",
		},
		{
			name:     "nested unclosed tags",
			input:    "<asdf><asdf>e",
			expected: "e",
			comment:  "multiple unclosed tags stripped",
		},
		{
			name:     "incomplete tag not stripped",
			input:    "hi, <f x",
			expected: "hi, <f x",
			comment:  "incomplete tag preserved (no closing >)",
		},
		{
			name:     "less-than comparison",
			input:    "234<235, right?",
			expected: "234<235, right?",
			comment:  "math comparison preserved",
		},
		{
			name:     "single char between tags",
			input:    "<x>b<y>",
			expected: "b",
			comment:  "content extracted",
		},
		{
			name:     "adjacent tags",
			input:    "<strong>foo</strong><a href=\"...\">bar</a>",
			expected: "foobar",
			comment:  "both tags stripped",
		},
		{
			name:     "ampersand and entities",
			input:    "&gotcha&#;<>",
			expected: "&gotcha&#;<>",
			comment:  "non-tags preserved",
		},
		// Obfuscation patterns - Django also leaves partial content
		{
			name:     "nested script tags",
			input:    "<sc<script>ript>alert(1)</script>",
			expected: "ript>alert(1)",
			comment:  "Django: inner tag breaks outer structure",
		},
		{
			name:     "comment obfuscation",
			input:    "<sc<!-- -->ript>test</script>",
			expected: "ript>test",
			comment:  "Django: comment stripped first leaves broken tag",
		},
		{
			name:     "double angle bracket",
			input:    "<<script>script>alert(1)</script>",
			expected: "alert(1)",
			comment:  "multi-pass handles this case",
		},
		{
			name:     "triple nested",
			input:    "<<<script>>script>alert(1)</script>",
			expected: "<<>script>alert(1)",
			comment:  "Django: leaves <<> as non-tag",
		},
		// More obfuscation attempts
		{
			name:     "tag inside tag name",
			input:    "<scr<b>ipt>alert(1)</script>",
			expected: "ipt>alert(1)",
			comment:  "inner tag breaks tag name",
		},
		{
			name:     "closing tag inside opener",
			input:    "<scr</b>ipt>alert(1)</script>",
			expected: "ipt>alert(1)",
			comment:  "closing tag in opener breaks structure",
		},
		{
			name:     "multiple nested tags",
			input:    "<s<s<script>cript>cript>x</script>",
			expected: "cript>cript>x",
			comment:  "multiple nested breaks",
		},
		{
			name:     "attribute with nested tag",
			input:    "<div title=\"<script>\">content</div>",
			expected: "content",
			comment:  "quoted attributes handled correctly",
		},
		{
			name:     "newline in tag",
			input:    "<script\n>alert(1)</script\n>",
			expected: "alert(1)",
			comment:  "newlines in tags handled",
		},
		{
			name:     "tab in tag",
			input:    "<script\t>alert(1)</script\t>",
			expected: "alert(1)",
			comment:  "tabs in tags handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterStriptags(AsValue(tt.input), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("got %q, want %q (%s)", result.String(), tt.expected, tt.comment)
			}
		})
	}
}

// TestFilterStriptagsMaxIterations tests that striptags returns an error when max iterations is reached.
// This protects against denial-of-service attacks with maliciously crafted input.
func TestFilterStriptagsMaxIterations(t *testing.T) {
	// Create input that requires more than 50 iterations to fully strip.
	// Pattern: N opening brackets followed by N letter-closing pairs
	// Example: <<<a>b>c> requires 3 iterations because:
	//   - Iter 0: removes <a>, leaves <<b>c>
	//   - Iter 1: removes <b>, leaves <c>
	//   - Iter 2: removes <c>, leaves empty
	// With 60 levels, we need 60 iterations to fully strip.
	var builder strings.Builder
	for range 60 {
		builder.WriteByte('<')
	}
	for i := range 60 {
		builder.WriteByte('a' + byte(i%26))
		builder.WriteByte('>')
	}

	input := builder.String()

	_, err := filterStriptags(AsValue(input), nil)
	if err == nil {
		t.Error("expected error when max iterations reached")
	}
	if err != nil && !strings.Contains(err.Error(), "did not converge") {
		t.Errorf("expected convergence error, got: %v", err)
	}
}

// TestFilterStriptagsConverges tests that normal input converges within the iteration limit.
func TestFilterStriptagsConverges(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple tag", "<script>alert(1)</script>"},
		{"nested tags", "<div><span><b>text</b></span></div>"},
		{"double nested", "<<script>script>alert(1)</script>"},
		{"triple nested angle", "<<<div>>>content</div>"},
		{"many tags", strings.Repeat("<b>x</b>", 100)},
		{"deep nesting", strings.Repeat("<div>", 20) + "content" + strings.Repeat("</div>", 20)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := filterStriptags(AsValue(tt.input), nil)
			if err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
			}
		})
	}
}

// FuzzFilterStriptags fuzzes the striptags filter for security testing
func FuzzFilterStriptags(f *testing.F) {
	// Django test vectors from tests/utils_tests/test_html.py
	f.Add("<p>See: &#39;&eacute; is an apostrophe followed by e acute</p>")
	f.Add("<adf>a")
	f.Add("</adf>a")
	f.Add("<asdf><asdf>e")
	f.Add("hi, <f x")
	f.Add("234<235, right?")
	f.Add("<x>b<y>")
	f.Add("<strong>foo</strong><a href=\"...\">bar</a>")
	f.Add("&gotcha&#;<>")

	// Basic HTML tags
	f.Add("<p>Hello World</p>")
	f.Add("<div>content</div>")
	f.Add("<span>text</span>")
	f.Add("<b>bold</b>")
	f.Add("<i>italic</i>")
	f.Add("<strong>strong</strong>")
	f.Add("<em>emphasis</em>")

	// Script and style tags (dangerous)
	f.Add("<script>alert(1)</script>")
	f.Add("<SCRIPT>alert(1)</SCRIPT>")
	f.Add("<ScRiPt>alert(1)</ScRiPt>")
	f.Add("<style>.x{color:red}</style>")
	f.Add("<script src='evil.js'></script>")
	f.Add("<script type='text/javascript'>x</script>")

	// Tags with attributes containing >
	f.Add("<p onclick=\"alert('>')\">text</p>")
	f.Add("<div data-x=\"a>b\">content</div>")
	f.Add("<a title=\"<b>bold</b>\">link</a>")
	f.Add("<input value=\"test>value\">")
	f.Add("<div style=\"content: '>'\">x</div>")

	// Single-quoted attributes with >
	f.Add("<p onclick='alert(\">\")'>text</p>")
	f.Add("<div data-x='a>b'>content</div>")
	f.Add("<a title='<b>bold</b>'>link</a>")

	// Mixed quotes
	f.Add("<div data-x=\"'>\" data-y='\"<'>x</div>")
	f.Add("<a href=\"test\" title='<script>'>link</a>")

	// Self-closing tags
	f.Add("<br/>")
	f.Add("<br />")
	f.Add("<hr/>")
	f.Add("<img src='x'/>")
	f.Add("<input type='text'/>")

	// Comments
	f.Add("<!-- comment -->")
	f.Add("<!--<script>x</script>-->")
	f.Add("before<!-- -->after")
	f.Add("<!--\nmultiline\n-->")

	// DOCTYPE and XML
	f.Add("<!DOCTYPE html>")
	f.Add("<?xml version=\"1.0\"?>")
	f.Add("<![CDATA[data]]>")

	// Nested tags
	f.Add("<div><span><b>text</b></span></div>")
	f.Add("<p><script>x</script></p>")
	f.Add("<a><b><c><d>deep</d></c></b></a>")

	// Obfuscated/nested injection attempts (Django-verified test vectors)
	f.Add("<sc<script>ript>alert(1)</script>")
	f.Add("<scr<scr<script>ipt>ipt>x</scr</scr</script>ipt>ipt>")
	f.Add("<<script>script>alert(1)</script>")
	f.Add("<script<script>>x</script</script>>")
	f.Add("<<<script>>script>alert(1)</script>")
	f.Add("<sc<!-- -->ript>test</script>")
	f.Add("<scr<b>ipt>alert(1)</script>")
	f.Add("<scr</b>ipt>alert(1)</script>")
	f.Add("<s<s<script>cript>cript>x</script>")
	f.Add("<div title=\"<script>\">content</div>")
	f.Add("<script\n>alert(1)</script\n>")
	f.Add("<script\t>alert(1)</script\t>")

	// Incomplete/malformed tags
	f.Add("<div")
	f.Add("div>")
	f.Add("</div")
	f.Add("<>")
	f.Add("< >")
	f.Add("</>")
	f.Add("<<")
	f.Add(">>")
	f.Add("<div<")
	f.Add(">div>")
	f.Add("hi, <f x")
	f.Add("234<235, right?")

	// Null bytes and control characters
	f.Add("<scr\x00ipt>x</script>")
	f.Add("\x00<script>x</script>")
	f.Add("<script\x00>x</script>")
	f.Add("<script>\x00</script>")
	f.Add("<div\x01class='x'>y</div>")

	// Event handlers
	f.Add("<div onclick=alert(1)>x</div>")
	f.Add("<img src=x onerror=alert(1)>")
	f.Add("<svg onload=alert(1)>")
	f.Add("<body onload=alert(1)>")
	f.Add("<input onfocus=alert(1) autofocus>")

	// URL-based XSS vectors
	f.Add("<a href='javascript:alert(1)'>x</a>")
	f.Add("<iframe src='javascript:alert(1)'></iframe>")
	f.Add("<object data='javascript:alert(1)'>")
	f.Add("<embed src='javascript:alert(1)'>")

	// Unicode content
	f.Add("<p>你好世界</p>")
	f.Add("<div>مرحبا</div>")
	f.Add("<span>שלום</span>")
	f.Add("<b>Привет</b>")
	f.Add("<p>Hello 😀 World</p>")
	f.Add("<div>🎉🎊🎁</div>")

	// Entities
	f.Add("&lt;script&gt;")
	f.Add("&#60;script&#62;")
	f.Add("&#x3c;script&#x3e;")
	f.Add("&amp;lt;script&amp;gt;")
	f.Add("&gotcha&#;<>")

	// Whitespace variations
	f.Add("<div   class='x'  >y</div>")
	f.Add("<div\nclass='x'\n>y</div>")
	f.Add("<div\tclass='x'\t>y</div>")
	f.Add("<div\rclass='x'\r>y</div>")
	f.Add("  <p>text</p>  ")

	// Empty and special cases
	f.Add("")
	f.Add("   ")
	f.Add("no tags here")
	f.Add("Hello World 123")
	f.Add("<")
	f.Add(">")
	f.Add("a<b")
	f.Add("a>b")
	f.Add("a<<b")
	f.Add("a>>b")
	f.Add("a<>b")

	// Long inputs
	f.Add("<div>" + string(make([]byte, 1000)) + "</div>")
	f.Add(string(make([]byte, 100)) + "<script>x</script>" + string(make([]byte, 100)))

	// SVG and MathML
	f.Add("<svg><circle cx='50'/></svg>")
	f.Add("<math><mi>x</mi></math>")
	f.Add("<svg><script>x</script></svg>")

	// Multiple adjacent tags
	f.Add("<b>a</b><i>b</i><u>c</u>")
	f.Add("<strong>x</strong><a href='#'>y</a>")

	f.Fuzz(func(t *testing.T, input string) {
		result, err := filterStriptags(AsValue(input), nil)
		if err != nil {
			// Errors are expected for maliciously crafted input that doesn't converge
			return
		}

		// Basic sanity checks
		output := result.String()

		// Output should not contain valid HTML tags (tags that actually match our pattern)
		// We use the same regex as the filter to check
		if reStriptags.MatchString(output) {
			t.Errorf("output still contains tags: input=%q, output=%q", input, output)
		}

		// Output should not contain null bytes (we strip them)
		if strings.Contains(output, "\x00") {
			t.Errorf("output contains null bytes: input=%q, output=%q", input, output)
		}

		// Output should be trimmed
		if output != strings.TrimSpace(output) {
			t.Errorf("output not trimmed: input=%q, output=%q", input, output)
		}
	})
}
