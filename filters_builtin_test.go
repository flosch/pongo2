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
			name:     "unicode characters removed",
			input:    "Hello Wörld",
			expected: "hello-wrld",
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
			{item: AsValue("a"), sortKey: "1"},
			{item: AsValue("b"), sortKey: "2"},
			{item: AsValue("c"), sortKey: "3"},
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
			{item: AsValue("a"), sortKey: "1"},
			{item: AsValue("b"), sortKey: "2"},
		}
		items.Swap(0, 1)
		if items[0].sortKey != "2" || items[1].sortKey != "1" {
			t.Errorf("Swap failed: got [%s, %s], want [2, 1]", items[0].sortKey, items[1].sortKey)
		}
	})

	t.Run("Less", func(t *testing.T) {
		items := dictsortItems{
			{item: AsValue("a"), sortKey: "apple"},
			{item: AsValue("b"), sortKey: "banana"},
			{item: AsValue("c"), sortKey: "apple"}, // Same as first
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

	t.Run("numeric string comparison", func(t *testing.T) {
		items := dictsortItems{
			{item: AsValue("a"), sortKey: "10"},
			{item: AsValue("b"), sortKey: "2"},
			{item: AsValue("c"), sortKey: "1"},
		}

		// String comparison: "1" < "10" < "2"
		if !items.Less(2, 0) { // "1" < "10"
			t.Error("Less should compare as strings: \"1\" < \"10\"")
		}
		if !items.Less(0, 1) { // "10" < "2"
			t.Error("Less should compare as strings: \"10\" < \"2\"")
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
		{"decimal with zero param trims", 34.5, 0, "34"},   // zero param means trim, whole part only
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

	t.Run("invalid tag name - multiple letters", func(t *testing.T) {
		result, err := filterRemovetags(AsValue("<div>content</div>"), AsValue("div"))
		if err == nil {
			t.Error("expected error for invalid tag name")
		}
		if result != nil {
			t.Error("expected nil result on error")
		}
	})

	t.Run("invalid tag name - number", func(t *testing.T) {
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
		{"nil with 2 args uses maybe default", nil, "on,off", "maybe"},
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

// TestFilterUrlizetruncEdgeCases tests urlizetrunc edge cases
func TestFilterUrlizetruncEdgeCases(t *testing.T) {
	t.Run("truncate long URL", func(t *testing.T) {
		input := "Visit www.verylongdomainname.com/with/long/path/here"
		result, err := filterUrlizetrunc(AsValue(input), AsValue(15))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should contain truncated title with ...
		if !strings.Contains(result.String(), "...") {
			t.Errorf("expected truncated URL with ellipsis, got %q", result.String())
		}
	})

	t.Run("short URL not truncated", func(t *testing.T) {
		input := "Visit www.ex.com"
		result, err := filterUrlizetrunc(AsValue(input), AsValue(100))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should not contain ellipsis
		if strings.Contains(result.String(), "...") {
			t.Errorf("short URL should not be truncated, got %q", result.String())
		}
	})

	t.Run("truncate email", func(t *testing.T) {
		input := "Email verylongusername@verylongdomain.com"
		result, err := filterUrlizetrunc(AsValue(input), AsValue(10))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should contain truncated email with ...
		if !strings.Contains(result.String(), "...") {
			t.Errorf("expected truncated email with ellipsis, got %q", result.String())
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

// TestFilterLinebreaksEmpty tests linebreaks with empty input
func TestFilterLinebreaksEmpty(t *testing.T) {
	result, err := filterLinebreaks(AsValue(""), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("got %q, want empty string", result.String())
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

// TestFilterEscapejsWithEscapeSequences tests escapejs with escape sequences
func TestFilterEscapejsWithEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"carriage return escape", "line1\\rline2", "\\u000D"},
		{"newline escape", "line1\\nline2", "\\u000A"},
		{"backslash", "path\\to\\file", "\\u005C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterEscapejs(AsValue(tt.input), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(result.String(), tt.contains) {
				t.Errorf("expected %q to contain %q", result.String(), tt.contains)
			}
		})
	}
}

// TestFilterEscapejsRuneError tests escapejs with invalid UTF-8
func TestFilterEscapejsRuneError(t *testing.T) {
	input := "hello\xffworld"
	result, err := filterEscapejs(AsValue(input), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should handle invalid UTF-8 gracefully
	_ = result.String()
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

// TestFilterTruncatecharsZero tests truncatechars with zero length
func TestFilterTruncatecharsZero(t *testing.T) {
	result, err := filterTruncatechars(AsValue("hello"), AsValue(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello" {
		t.Errorf("got %q, want %q", result.String(), "hello")
	}
}

// TestFilterTruncatecharsLessThanThree tests truncatechars with length < 3
func TestFilterTruncatecharsLessThanThree(t *testing.T) {
	result, err := filterTruncatechars(AsValue("hello"), AsValue(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Less than 3 chars: no room for ellipsis, just truncate
	if result.String() != "he" {
		t.Errorf("got %q, want %q", result.String(), "he")
	}
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

func FuzzBuiltinFilters(f *testing.F) {
	f.Add("foobar", "123")
	f.Add("foobar", `123,456`)
	f.Add("foobar", `123,456,"789"`)
	f.Add("foobar", `"test","test"`)
	f.Add("foobar", `123,"test"`)
	f.Add("foobar", "")
	f.Add("123", "foobar")

	f.Fuzz(func(t *testing.T, value, filterArg string) {
		ts := NewSet("fuzz-test", &DummyLoader{})
		for name := range filters {
			tpl, err := ts.FromString(fmt.Sprintf("{{ %v|%v:%v }}", value, name, filterArg))
			if tpl != nil && err != nil {
				t.Errorf("filter=%q value=%q, filterArg=%q, err=%v", name, value, filterArg, err)
			}
			if err == nil {
				tpl.Execute(nil) //nolint:errcheck
			}
		}
	})
}
