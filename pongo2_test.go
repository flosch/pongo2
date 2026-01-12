package pongo2_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/flosch/pongo2/v6"
)

var testSuite2 = pongo2.NewSet("test suite 2", pongo2.MustNewLocalFileSystemLoader(""))

func mustEqual(t *testing.T, s, pattern string) {
	if !regexp.MustCompile(pattern).MatchString(s) {
		t.Fatalf("mustEqual failed: '%v' does not match pattern '%v'", s, pattern)
	}
}

func mustPanicMatch(t *testing.T, fn func(), pattern string) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Expected panic with pattern '%v', nothing happened", pattern)
		}

		if !regexp.MustCompile(pattern).MatchString(fmt.Sprintf("%v", err)) {
			t.Fatalf("Expected panic with pattern '%v', but got '%v'", pattern, err)
		}
	}()

	// We expect fn to panic
	fn()
}

func parseTemplate(s string, c pongo2.Context) string {
	t, err := testSuite2.FromString(s)
	if err != nil {
		panic(err)
	}
	out, err := t.Execute(c)
	if err != nil {
		panic(err)
	}
	return out
}

func parseTemplateFn(s string, c pongo2.Context) func() {
	return func() {
		parseTemplate(s, c)
	}
}

func TestMisc(t *testing.T) {
	// Must
	// TODO: Add better error message (see issue #18)
	mustPanicMatch(
		t,
		func() { pongo2.Must(testSuite2.FromFile("template_tests/inheritance/base2.tpl")) },
		`\[Error \(where: fromfile\) in .*template_tests[/\\]inheritance[/\\]doesnotexist.tpl | Line 1 Col 12 near 'doesnotexist.tpl'\] open .*template_tests[/\\]inheritance[/\\]doesnotexist.tpl: no such file or directory`,
	)

	// Context
	mustPanicMatch(t, parseTemplateFn("", pongo2.Context{"'illegal": nil}), ".*not a valid identifier.*")

	// Registers
	mustEqual(t, pongo2.RegisterFilter("escape", nil).Error(), ".*is already registered")
	mustEqual(t, pongo2.RegisterTag("for", nil).Error(), ".*is already registered")

	// ApplyFilter
	v, err := pongo2.ApplyFilter("title", pongo2.AsValue("this is a title"), nil)
	if err != nil {
		t.Fatal(err)
	}
	mustEqual(t, v.String(), "This Is A Title")
	mustPanicMatch(t, func() {
		_, err := pongo2.ApplyFilter("doesnotexist", nil, nil)
		if err != nil {
			panic(err)
		}
	}, `\[Error \(where: applyfilter\)\] filter with name 'doesnotexist' not found`)
}

func TestImplicitExecCtx(t *testing.T) {
	tpl, err := pongo2.FromString("{{ ImplicitExec }}")
	if err != nil {
		t.Fatalf("Error in FromString: %v", err)
	}

	val := "a stringy thing"

	res, err := tpl.Execute(pongo2.Context{
		"Value": val,
		"ImplicitExec": func(ctx *pongo2.ExecutionContext) string {
			return ctx.Public["Value"].(string)
		},
	})
	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}

	mustEqual(t, res, val)

	// The implicit ctx should not be persisted from call-to-call
	res, err = tpl.Execute(pongo2.Context{
		"ImplicitExec": func() string {
			return val
		},
	})

	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}

	mustEqual(t, res, val)
}

// TestForloopRevcounter verifies that forloop.Revcounter and forloop.Revcounter0
// match Django's documented behavior:
// - Revcounter: iterations remaining (1-indexed, counts down to 1)
// - Revcounter0: iterations remaining (0-indexed, counts down to 0)
func TestForloopRevcounter(t *testing.T) {
	items := []int{10, 20, 30, 40, 50}

	tpl, err := pongo2.FromString(`{% for item in items %}{{ forloop.Counter }}:{{ forloop.Revcounter }},{{ forloop.Revcounter0 }} {% endfor %}`)
	if err != nil {
		t.Fatalf("Error parsing template: %v", err)
	}

	result, err := tpl.Execute(pongo2.Context{"items": items})
	if err != nil {
		t.Fatalf("Error executing template: %v", err)
	}

	// For 5 items:
	// Counter=1: Revcounter=5, Revcounter0=4
	// Counter=2: Revcounter=4, Revcounter0=3
	// Counter=3: Revcounter=3, Revcounter0=2
	// Counter=4: Revcounter=2, Revcounter0=1
	// Counter=5: Revcounter=1, Revcounter0=0
	expected := "1:5,4 2:4,3 3:3,2 4:2,1 5:1,0 "
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestUrlizeFilter tests the urlize filter with various URL formats and TLDs
func TestUrlizeFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Existing TLDs that should work
		{
			name:     "http URL with .com",
			input:    "Visit http://example.com today",
			expected: `Visit <a href="http://example.com" rel="nofollow">http://example.com</a> today`,
		},
		{
			name:     "www URL with .org",
			input:    "Check www.example.org please",
			expected: `Check <a href="http://www.example.org" rel="nofollow">www.example.org</a> please`,
		},
		// New TLDs that should be supported
		{
			name:     "URL with .io TLD",
			input:    "Visit github.io for pages",
			expected: `Visit <a href="http://github.io" rel="nofollow">github.io</a> for pages`,
		},
		{
			name:     "URL with .dev TLD",
			input:    "Check web.dev for tips",
			expected: `Check <a href="http://web.dev" rel="nofollow">web.dev</a> for tips`,
		},
		{
			name:     "URL with .co TLD",
			input:    "See example.co now",
			expected: `See <a href="http://example.co" rel="nofollow">example.co</a> now`,
		},
		{
			name:     "URL with .ai TLD",
			input:    "Try claude.ai today",
			expected: `Try <a href="http://claude.ai" rel="nofollow">claude.ai</a> today`,
		},
		{
			name:     "URL with .app TLD",
			input:    "Download from myapp.app",
			expected: `Download from <a href="http://myapp.app" rel="nofollow">myapp.app</a>`,
		},
		// Country-code TLDs
		{
			name:     "URL with .uk TLD",
			input:    "Visit bbc.co.uk for news",
			expected: `Visit <a href="http://bbc.co.uk" rel="nofollow">bbc.co.uk</a> for news`,
		},
		{
			name:     "URL with .fr TLD",
			input:    "See example.fr please",
			expected: `See <a href="http://example.fr" rel="nofollow">example.fr</a> please`,
		},
		// Email addresses
		{
			name:     "email with .info TLD",
			input:    "Contact user@example.info",
			expected: `Contact <a href="mailto:user@example.info">user@example.info</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(`{{ input|urlize|safe }}`)
			if err != nil {
				t.Fatalf("Error parsing template: %v", err)
			}

			result, err := tpl.Execute(pongo2.Context{"input": tt.input})
			if err != nil {
				t.Fatalf("Error executing template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

// TestValueEqualTo tests the Value.EqualTo method for various types
func TestValueEqualTo(t *testing.T) {
	tests := []struct {
		name     string
		v1       interface{}
		v2       interface{}
		expected bool
	}{
		// Integers
		{"equal integers", 42, 42, true},
		{"different integers", 42, 43, false},
		{"int and int64 equal", 42, int64(42), true},

		// Floats
		{"equal floats", 3.14, 3.14, true},
		{"different floats", 3.14, 2.71, false},

		// Strings
		{"equal strings", "hello", "hello", true},
		{"different strings", "hello", "world", false},

		// Booleans
		{"equal bools true", true, true, true},
		{"equal bools false", false, false, true},
		{"different bools", true, false, false},

		// Nil values
		// Note: Current implementation returns false for two nil values
		// Using reflect.Value.Equal would return true for two invalid values
		{"both nil", nil, nil, false},
		{"one nil", nil, "hello", false},

		// Slices (not comparable, should return false)
		{"slices", []int{1, 2}, []int{1, 2}, false},

		// Maps (not comparable, should return false)
		{"maps", map[string]int{"a": 1}, map[string]int{"a": 1}, false},

		// Structs (comparable)
		{"equal structs", struct{ A int }{1}, struct{ A int }{1}, true},
		{"different structs", struct{ A int }{1}, struct{ A int }{2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1 := pongo2.AsValue(tt.v1)
			v2 := pongo2.AsValue(tt.v2)

			result := v1.EqualValueTo(v2)
			if result != tt.expected {
				t.Errorf("EqualTo(%v, %v) = %v, want %v", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

// BenchmarkValueEqualTo benchmarks the EqualValueTo method
func BenchmarkValueEqualTo(b *testing.B) {
	// Pre-create values to avoid allocation overhead in benchmark
	intVal1 := pongo2.AsValue(42)
	intVal2 := pongo2.AsValue(42)
	strVal1 := pongo2.AsValue("hello world")
	strVal2 := pongo2.AsValue("hello world")
	structVal1 := pongo2.AsValue(struct{ A, B, C int }{1, 2, 3})
	structVal2 := pongo2.AsValue(struct{ A, B, C int }{1, 2, 3})
	sliceVal1 := pongo2.AsValue([]int{1, 2, 3})
	sliceVal2 := pongo2.AsValue([]int{1, 2, 3})

	b.Run("integers", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			intVal1.EqualValueTo(intVal2)
		}
	})

	b.Run("strings", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			strVal1.EqualValueTo(strVal2)
		}
	})

	b.Run("structs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			structVal1.EqualValueTo(structVal2)
		}
	})

	b.Run("slices_not_comparable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sliceVal1.EqualValueTo(sliceVal2)
		}
	})
}

type DummyLoader struct{}

func (l *DummyLoader) Abs(base, name string) string {
	return filepath.Join(filepath.Dir(base), name)
}

func (l *DummyLoader) Get(path string) (io.Reader, error) {
	return nil, errors.New("dummy not found")
}

func FuzzSimpleExecution(f *testing.F) {
	tpls, err := filepath.Glob("template_tests/*.tpl")
	if err != nil {
		f.Fatalf("glob: %v", err)
	}
	files := []string{"README.md"}
	files = append(files, tpls...)

	for _, tplPath := range files {
		buf, err := os.ReadFile(tplPath)
		if err != nil {
			f.Fatalf("could not read file '%v': %v", tplPath, err)
		}
		f.Add(string(buf), "test-value")
	}

	f.Add("{{ foobar }}", "00000000")

	f.Fuzz(func(t *testing.T, tpl, contextValue string) {
		ts := pongo2.NewSet("fuzz-test", &DummyLoader{})
		out, err := ts.FromString(tpl)
		if err != nil && out != nil {
			t.Errorf("%v", err)
		}
		if err == nil {
			mycontext := pongo2.Context{
				"foobar": contextValue,
			}
			mycontext.Update(tplContext)
			out.Execute(mycontext) //nolint:errcheck
		}
	})
}
