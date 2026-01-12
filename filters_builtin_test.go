package pongo2

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
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
