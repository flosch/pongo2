package pongo2_test

import (
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/flosch/pongo2/v7"
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

func TestIssue338(t *testing.T) {
	// Test that include with 'with' clause works inside a for loop
	// that iterates over a map with a single variable.
	// Bug: {% for key in map %} was setting Private[""] = value for maps,
	// which caused errors when include copied the context.
	// See: https://github.com/flosch/pongo2/issues/338

	memFS := fstest.MapFS{
		"main.tpl": &fstest.MapFile{
			Data: []byte(`{% for key in mymap %}{% include "field.html" with name=key id="test" %}{% endfor %}`),
		},
		"field.html": &fstest.MapFile{
			Data: []byte(`<label for="{{ id }}">{{ name }}</label>`),
		},
	}

	loader := pongo2.NewFSLoader(memFS)
	set := pongo2.NewSet("test", loader)

	tpl, err := set.FromFile("main.tpl")
	if err != nil {
		t.Fatalf("failed to load template: %v", err)
	}

	ctx := pongo2.Context{
		"mymap": map[string]string{
			"a": "1",
			"b": "2",
		},
	}

	result, err := tpl.Execute(ctx)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	// Should render the label twice with keys from the map
	// Map iteration order is not guaranteed, so check for both possible outputs
	expected1 := `<label for="test">a</label><label for="test">b</label>`
	expected2 := `<label for="test">b</label><label for="test">a</label>`
	if result != expected1 && result != expected2 {
		t.Errorf("expected %q or %q, got %q", expected1, expected2, result)
	}
}

func TestIssue343(t *testing.T) {
	// Test \n escape sequence (newline) - the main issue
	tpl, err := pongo2.FromString(`{% for val in my_string|split:"\n" %}[{{ val }}]{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err := tpl.Execute(pongo2.Context{"my_string": "Line 1\nLine 2\nLine 3"})
	if err != nil {
		t.Fatal(err)
	}
	if str != "[Line 1][Line 2][Line 3]" {
		t.Fatalf(`Expected "[Line 1][Line 2][Line 3]", but got %q`, str)
	}

	// Test \t escape sequence (tab)
	tpl, err = pongo2.FromString(`{% for val in my_string|split:"\t" %}[{{ val }}]{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(pongo2.Context{"my_string": "A\tB\tC"})
	if err != nil {
		t.Fatal(err)
	}
	if str != "[A][B][C]" {
		t.Fatalf(`Expected "[A][B][C]", but got %q`, str)
	}

	// Test \r escape sequence (carriage return)
	tpl, err = pongo2.FromString(`{% for val in my_string|split:"\r" %}[{{ val }}]{% endfor %}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(pongo2.Context{"my_string": "X\rY\rZ"})
	if err != nil {
		t.Fatal(err)
	}
	if str != "[X][Y][Z]" {
		t.Fatalf(`Expected "[X][Y][Z]", but got %q`, str)
	}

	// Test \' escape sequence in single-quoted strings
	// Use |safe to prevent autoescape from converting ' to &#39;
	tpl, err = pongo2.FromString(`{{ 'it\'s working'|safe }}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if str != "it's working" {
		t.Fatalf(`Expected "it's working", but got %q`, str)
	}

	// Test \\n produces literal backslash+n, not newline
	tpl, err = pongo2.FromString(`{{ "hello\\nworld" }}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if str != "hello\\nworld" {
		t.Fatalf(`Expected "hello\\nworld" (literal backslash-n), but got %q`, str)
	}

	// Test existing escape sequences still work: \"
	// Use |safe to prevent autoescape from converting " to &quot;
	tpl, err = pongo2.FromString(`{{ "say \"hello\""|safe }}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if str != `say "hello"` {
		t.Fatalf(`Expected 'say "hello"', but got %q`, str)
	}

	// Test existing escape sequences still work: \\
	tpl, err = pongo2.FromString(`{{ "back\\slash" }}`)
	if err != nil {
		t.Fatal(err)
	}
	str, err = tpl.Execute(nil)
	if err != nil {
		t.Fatal(err)
	}
	if str != "back\\slash" {
		t.Fatalf(`Expected "back\\slash", but got %q`, str)
	}
}

func TestIssue341(t *testing.T) {
	// Test that comparing two undefined/missing variables returns True
	// Bug: {{ _missing == null }} returned False on master but True in v6.0.0
	// Both _missing and null are undefined, so they should be equal.
	// See: https://github.com/flosch/pongo2/issues/341

	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "two missing variables are equal",
			template: "{{ _missing == _also_missing }}",
			context:  pongo2.Context{},
			expected: "True",
		},
		{
			name:     "missing variable equals undefined literal",
			template: "{{ _missing == null }}",
			context:  pongo2.Context{},
			expected: "True",
		},
		{
			name:     "missing variable not equal to defined value",
			template: "{{ _missing == defined }}",
			context:  pongo2.Context{"defined": "value"},
			expected: "False",
		},
		{
			name:     "missing variable not equal to empty string",
			template: "{{ _missing == '' }}",
			context:  pongo2.Context{},
			expected: "False",
		},
		{
			name:     "missing variable not equal to zero",
			template: "{{ _missing == 0 }}",
			context:  pongo2.Context{},
			expected: "False",
		},
		{
			name:     "nil context value equals missing variable",
			template: "{{ nil_value == _missing }}",
			context:  pongo2.Context{"nil_value": nil},
			expected: "True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
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

func TestIssue322(t *testing.T) {
	// Test for data race in FromFile when called concurrently.
	// The race was on firstTemplateCreated field being written without synchronization.
	// See: https://github.com/flosch/pongo2/issues/322

	memFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte(`<h1>{{ status }}</h1>`),
		},
	}

	loader := pongo2.NewFSLoader(memFS)
	set := pongo2.NewSet("test", loader)

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()
			tpl, err := set.FromFile("index.html")
			if err != nil {
				t.Errorf("FromFile failed: %v", err)
				return
			}
			_, err = tpl.Execute(pongo2.Context{"status": "Hello World"})
			if err != nil {
				t.Errorf("Execute failed: %v", err)
			}
		}()
	}

	wg.Wait()
}

func TestIssue209(t *testing.T) {
	// Test that "not X in Y" is correctly parsed as "not (X in Y)"
	// Bug: "not X in Y" was being parsed as "(not X) in Y" due to
	// incorrect operator precedence. The "not" operator should have
	// lower precedence than "in".
	// See: https://github.com/flosch/pongo2/issues/209

	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "not in - element exists",
			template: `{% if not "Hello" in text %}is not{% else %}is{% endif %}`,
			context:  pongo2.Context{"text": "<h2>Hello!</h2><p>Welcome"},
			expected: "is",
		},
		{
			name:     "not in - element does not exist",
			template: `{% if not "World" in text %}is not{% else %}is{% endif %}`,
			context:  pongo2.Context{"text": "<h2>Hello!</h2><p>Welcome"},
			expected: "is not",
		},
		{
			name:     "not in with list - element exists",
			template: `{% if not 2 in numbers %}not found{% else %}found{% endif %}`,
			context:  pongo2.Context{"numbers": []int{1, 2, 3}},
			expected: "found",
		},
		{
			name:     "not in with list - element does not exist",
			template: `{% if not 5 in numbers %}not found{% else %}found{% endif %}`,
			context:  pongo2.Context{"numbers": []int{1, 2, 3}},
			expected: "not found",
		},
		{
			name:     "not in with map - key exists",
			template: `{% if not "foo" in mymap %}not found{% else %}found{% endif %}`,
			context:  pongo2.Context{"mymap": map[string]int{"foo": 1, "bar": 2}},
			expected: "found",
		},
		{
			name:     "not in with map - key does not exist",
			template: `{% if not "baz" in mymap %}not found{% else %}found{% endif %}`,
			context:  pongo2.Context{"mymap": map[string]int{"foo": 1, "bar": 2}},
			expected: "not found",
		},
		{
			name:     "double negation - not not in",                               //nolint:dupword
			template: `{% if not not "Hello" in text %}yes{% else %}no{% endif %}`, //nolint:dupword
			context:  pongo2.Context{"text": "Hello World"},
			expected: "yes",
		},
		{
			name:     "not in combined with and",
			template: `{% if flag and not "x" in text %}yes{% else %}no{% endif %}`,
			context:  pongo2.Context{"flag": true, "text": "abc"},
			expected: "yes",
		},
		{
			name:     "not in combined with or",
			template: `{% if not "x" in text or flag %}yes{% else %}no{% endif %}`,
			context:  pongo2.Context{"flag": false, "text": "abc"},
			expected: "yes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
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

func TestIssue237(t *testing.T) {
	// Test that ifchanged with else clause works correctly when content doesn't change.
	// Bug: ifchanged without watched expressions would panic or not render else block
	// when the content doesn't change.
	// See: https://github.com/flosch/pongo2/issues/237

	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "content-based ifchanged with else - content changes",
			template: `{% for item in items %}{% ifchanged %}{{ item }}{% else %}*{% endifchanged %}{% endfor %}`,
			context:  pongo2.Context{"items": []string{"a", "b", "b", "c"}},
			expected: "ab*c",
		},
		{
			name:     "content-based ifchanged with else - all same",
			template: `{% for item in items %}{% ifchanged %}{{ item }}{% else %}*{% endifchanged %}{% endfor %}`,
			context:  pongo2.Context{"items": []string{"a", "a", "a"}},
			expected: "a**",
		},
		{
			name:     "content-based ifchanged without else",
			template: `{% for item in items %}{% ifchanged %}[{{ item }}]{% endifchanged %}{% endfor %}`,
			context:  pongo2.Context{"items": []string{"a", "a", "b", "b", "c"}},
			expected: "[a][b][c]",
		},
		{
			name:     "expression-based ifchanged with else",
			template: `{% for item in items %}{% ifchanged item %}{{ item }}{% else %}*{% endifchanged %}{% endfor %}`,
			context:  pongo2.Context{"items": []string{"a", "a", "b", "c", "c"}},
			expected: "a*bc*",
		},
		{
			name:     "expression-based ifchanged without else",
			template: `{% for item in items %}{% ifchanged item %}[{{ item }}]{% endifchanged %}{% endfor %}`,
			context:  pongo2.Context{"items": []string{"a", "a", "b", "b"}},
			expected: "[a][b]",
		},
		{
			name:     "nested loop with content-based ifchanged and else",
			template: `{% for outer in outers %}{% for inner in inners %}{% ifchanged %}{{ outer }}-{{ inner }}{% else %}*{% endifchanged %}{% endfor %}{% endfor %}`,
			context: pongo2.Context{
				"outers": []string{"A", "B"},
				"inners": []string{"1", "1", "2"},
			},
			expected: "A-1*A-2B-1*B-2",
		},
		{
			name:     "nested loop with expression-based ifchanged without else",
			template: `{% for outer in outers %}{% for inner in inners %}{% ifchanged inner %}[{{ outer }}-{{ inner }}]{% endifchanged %}{% endfor %}{% endfor %}`,
			context: pongo2.Context{
				"outers": []string{"A", "B"},
				"inners": []string{"1", "1", "2"},
			},
			expected: "[A-1][A-2][B-1][B-2]",
		},
		{
			name:     "nested loop with content-based ifchanged without else - original issue scenario",
			template: `{% for outer in outers %}{% for inner in inners %}{% ifchanged %}{{ inner }}{% endifchanged %}{% endfor %}{% endfor %}`,
			context: pongo2.Context{
				"outers": []string{"A", "B"},
				"inners": []string{"x", "x", "y"},
			},
			expected: "xyxy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
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

func TestBugStringByteIndexing(t *testing.T) {
	// Bug: String indexing used byte offset instead of rune offset.
	// For multi-byte UTF-8 characters (e.g., Chinese, emoji), accessing
	// string[1] would return a garbled byte instead of the second character.
	// This affects both integer index access (e.g., "var.0") and subscript
	// access (e.g., "var[idx]").

	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "ASCII string index 0",
			template: `{{ s.0 }}`,
			context:  pongo2.Context{"s": "hello"},
			expected: "h",
		},
		{
			name:     "ASCII string index 1",
			template: `{{ s.1 }}`,
			context:  pongo2.Context{"s": "hello"},
			expected: "e",
		},
		{
			name:     "multi-byte string index 0",
			template: `{{ s.0 }}`,
			context:  pongo2.Context{"s": "日本語"},
			expected: "日",
		},
		{
			name:     "multi-byte string index 1",
			template: `{{ s.1 }}`,
			context:  pongo2.Context{"s": "日本語"},
			expected: "本",
		},
		{
			name:     "multi-byte string index 2",
			template: `{{ s.2 }}`,
			context:  pongo2.Context{"s": "日本語"},
			expected: "語",
		},
		{
			name:     "mixed ASCII and multi-byte index",
			template: `{{ s.1 }}`,
			context:  pongo2.Context{"s": "aüc"},
			expected: "ü",
		},
		{
			name:     "subscript access multi-byte string",
			template: `{{ s[idx] }}`,
			context:  pongo2.Context{"s": "日本語", "idx": 2},
			expected: "語",
		},
		{
			name:     "out of bounds returns empty for multi-byte",
			template: `{{ s.3 }}`,
			context:  pongo2.Context{"s": "日本語"},
			expected: "",
		},
		{
			name:     "emoji string indexing",
			template: `{{ s.1 }}`,
			context:  pongo2.Context{"s": "A😀B"},
			expected: "😀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
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

func TestBugLjustRjustErrorMessage(t *testing.T) {
	// Bug: ljust and rjust used %c format verb instead of %d in error messages.
	// With maxCharPadding=10000, %c would render U+2710 (✐) instead of "10000".

	tests := []struct {
		name            string
		template        string
		expectedContain string
	}{
		{
			name:            "ljust error contains number not unicode",
			template:        `{{ "x"|ljust:20000 }}`,
			expectedContain: "10000",
		},
		{
			name:            "rjust error contains number not unicode",
			template:        `{{ "x"|rjust:20000 }}`,
			expectedContain: "10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			_, err = tpl.Execute(nil)
			if err == nil {
				t.Fatal("expected an error but got none")
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.expectedContain) {
				t.Errorf("error message should contain %q, got %q", tt.expectedContain, errMsg)
			}
		})
	}
}

func TestBugAutoescapeNotRestoredOnError(t *testing.T) {
	// Bug: tagAutoescapeNode.Execute did not restore ctx.Autoescape when
	// wrapper.Execute returned an error. The fix uses defer to ensure
	// the autoescape state is always restored.

	tpl, err := pongo2.FromString(`{% autoescape off %}{{ safe }}{% endautoescape %}{{ unsafe }}`)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	result, err := tpl.Execute(pongo2.Context{
		"safe":   "<b>bold</b>",
		"unsafe": "<script>xss</script>",
	})
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	expected := "<b>bold</b>&lt;script&gt;xss&lt;/script&gt;"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	// Nested autoescape blocks restore correctly.
	tpl2, err := pongo2.FromString(`{% autoescape off %}{% autoescape on %}{{ inner }}{% endautoescape %}{{ outer }}{% endautoescape %}`)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	result2, err := tpl2.Execute(pongo2.Context{
		"inner": "<i>inner</i>",
		"outer": "<b>outer</b>",
	})
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	expected2 := "&lt;i&gt;inner&lt;/i&gt;<b>outer</b>"
	if result2 != expected2 {
		t.Errorf("expected %q, got %q", expected2, result2)
	}
}

func TestBugLexerColumnMultiByteUTF8(t *testing.T) {
	// Bug: The lexer incremented col by byte width instead of 1 for each rune.
	// For multi-byte UTF-8 characters, error column numbers would be wrong.

	tests := []struct {
		name        string
		template    string
		expectedCol string
	}{
		{
			name:        "error column with only ASCII (baseline)",
			template:    "abc{{ invalid_syntax( }}",
			expectedCol: "Col 23",
		},
		{
			name:        "error column after multi-byte chars",
			template:    "日本語{{ invalid_syntax( }}",
			expectedCol: "Col 23",
		},
		{
			name:        "error column after mixed ASCII and multi-byte",
			template:    "aüc{{ invalid_syntax( }}",
			expectedCol: "Col 23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pongo2.FromString(tt.template)
			if err == nil {
				t.Fatal("expected a parse error but got none")
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.expectedCol) {
				t.Errorf("error message should contain %q, got %q", tt.expectedCol, errMsg)
			}
		})
	}
}

func TestBugNegateFloatReturns1Point1(t *testing.T) {
	// Bug: Negate() returned float64(1.1) instead of float64(1.0) when
	// negating a zero float value. This was a typo in the source code.

	tpl, err := pongo2.FromString(`{% if not zero_float %}yes{% else %}no{% endif %}`)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	result, err := tpl.Execute(pongo2.Context{"zero_float": 0.0})
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	if result != "yes" {
		t.Errorf("expected %q, got %q", "yes", result)
	}

	tpl2, err := pongo2.FromString(`{{ zero_float|default:"fallback" }}`)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	result2, err := tpl2.Execute(pongo2.Context{"zero_float": 0.0})
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	if result2 != "fallback" {
		t.Errorf("expected %q, got %q", "fallback", result2)
	}

	// Test that negated non-zero float gives 0.0
	v := pongo2.AsValue(3.14)
	negated := v.Negate()
	if negated.Float() != 0.0 {
		t.Errorf("Negate() of non-zero float should be 0.0, got %v", negated.Float())
	}

	// Test that negated zero float gives 1.0 (not 1.1)
	v2 := pongo2.AsValue(0.0)
	negated2 := v2.Negate()
	if negated2.Float() != 1.0 {
		t.Errorf("Negate() of zero float should be 1.0, got %v", negated2.Float())
	}
}

func TestBugWidthratioDoubleRounding(t *testing.T) {
	// Bug: widthratio used Ceil(x + 0.5) which double-rounds.
	// For exact integer results like 50/100*200 = 100.0,
	// Ceil(100.0 + 0.5) = 101 instead of the correct 100.

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "exact integer result not inflated",
			template: `{% widthratio 50 100 200 %}`,
			expected: "100",
		},
		{
			name:     "exact integer result 75/100*400",
			template: `{% widthratio 75 100 400 %}`,
			expected: "300",
		},
		{
			name:     "fractional result rounds correctly",
			template: `{% widthratio 175 200 100 %}`,
			expected: "88",
		},
		{
			name:     "small fraction rounds down",
			template: `{% widthratio 1 3 100 %}`,
			expected: "33",
		},
		{
			name:     "zero value",
			template: `{% widthratio 0 100 200 %}`,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			result, err := tpl.Execute(nil)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBugIfnotequalErrorMessage(t *testing.T) {
	// Bug: ifnotequal's error message incorrectly says "ifequal" instead of "ifnotequal".

	_, err := pongo2.FromString(`{% ifnotequal a b c %}{% endifnotequal %}`)
	if err == nil {
		t.Fatal("expected an error but got none")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "ifnotequal") {
		t.Errorf("error message should mention 'ifnotequal', got %q", errMsg)
	}
}

func TestBugCycleSharedState(t *testing.T) {
	// Bug: tagCycleNode.idx was stored on the AST node, which is shared
	// across all concurrent executions of the same parsed template.
	// This caused two problems:
	// 1. Data race: concurrent idx++ without synchronization
	// 2. Semantic bug: one execution's cycle position affected another,
	//    producing wrong output (e.g., "bcabca" instead of "abcabc")
	//
	// Each template execution must have independent cycle state.

	tpl, err := pongo2.FromString(`{% for i in items %}{% cycle "a" "b" "c" %}{% endfor %}`)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	ctx := pongo2.Context{"items": []int{1, 2, 3, 4, 5, 6}}
	const expected = "abcabc"

	// First: verify sequential executions produce consistent results.
	// With shared state on the node, the second execution would start
	// where the first left off (idx=6), producing wrong output.
	for i := range 5 {
		result, err := tpl.Execute(ctx)
		if err != nil {
			t.Fatalf("execution %d: unexpected error: %v", i, err)
		}
		if result != expected {
			t.Errorf("execution %d: got %q, want %q", i, result, expected)
		}
	}

	// Second: verify concurrent executions are also correct.
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := tpl.Execute(ctx)
			if err != nil {
				t.Errorf("concurrent: unexpected error: %v", err)
				return
			}
			if result != expected {
				t.Errorf("concurrent: got %q, want %q", result, expected)
			}
		}()
	}
	wg.Wait()
}

func TestBugIfchangedSharedState(t *testing.T) {
	// Bug: tagIfchangedNode.lastValues and lastContent were stored on the
	// AST node, which is shared across all concurrent executions.
	// This caused two problems:
	// 1. Data race: concurrent writes to lastValues/lastContent
	// 2. Semantic bug: ifchanged compares against state from a DIFFERENT
	//    execution, so the first item might be suppressed if a previous
	//    execution ended with the same value.
	//
	// Each template execution must have independent ifchanged state.

	tpl, err := pongo2.FromString(`{% for item in items %}{% ifchanged %}{{ item }}{% endifchanged %}{% endfor %}`)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	ctx := pongo2.Context{"items": []string{"a", "a", "b", "b", "c"}}
	const expected = "abc"

	// First: verify sequential executions produce consistent results.
	// With shared state, the second execution starts with lastContent="c"
	// from the first execution, so "a" would be correctly shown (different
	// from "c"), but the pattern breaks in more complex scenarios.
	// Use a case where it definitely breaks: items starting with the same
	// value the previous execution ended with.
	tplSameEnd, err := pongo2.FromString(`{% for item in items %}{% ifchanged %}{{ item }}{% endifchanged %}{% endfor %}`)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	// Items end with "x", so next execution starting with "x" would skip it
	ctxEndsX := pongo2.Context{"items": []string{"a", "b", "x"}}
	ctxStartsX := pongo2.Context{"items": []string{"x", "y", "z"}}

	// First execution ends with lastContent="x"
	result1, err := tplSameEnd.Execute(ctxEndsX)
	if err != nil {
		t.Fatalf("execution 1: unexpected error: %v", err)
	}
	if result1 != "abx" {
		t.Errorf("execution 1: got %q, want %q", result1, "abx")
	}

	// Second execution should output "x" even though previous ended with "x"
	result2, err := tplSameEnd.Execute(ctxStartsX)
	if err != nil {
		t.Fatalf("execution 2: unexpected error: %v", err)
	}
	if result2 != "xyz" {
		t.Errorf("execution 2: got %q, want %q (ifchanged state leaked from previous execution)", result2, "xyz")
	}

	// Also verify concurrent executions produce correct results.
	var wg2 sync.WaitGroup
	for range 20 {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			result, err := tpl.Execute(ctx)
			if err != nil {
				t.Errorf("concurrent: unexpected error: %v", err)
				return
			}
			if result != expected {
				t.Errorf("concurrent: got %q, want %q", result, expected)
			}
		}()
	}
	wg2.Wait()
}

func TestBugWidthratioDivisionByZero(t *testing.T) {
	// Bug: widthratio does not check for division by zero when max=0.
	// Django returns "0" in this case.

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "max is zero literal",
			template: `{% widthratio 50 0 200 %}`,
			expected: "0",
		},
		{
			name:     "max is zero variable",
			template: `{% widthratio value max_value width %}`,
			expected: "0",
		},
		{
			name:     "all zeros",
			template: `{% widthratio 0 0 0 %}`,
			expected: "0",
		},
		{
			name:     "max zero with as variable",
			template: `{% widthratio 50 0 200 as result %}{{ result }}`,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			ctx := pongo2.Context{"value": 50, "max_value": 0, "width": 200}
			result, err := tpl.Execute(ctx)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestWidthratioBankersRounding(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "half rounds to even (12.5 -> 12)",
			template: "{% widthratio 1 4 50 %}",
			expected: "12",
		},
		{
			name:     "half rounds to even (87.5 -> 88)",
			template: "{% widthratio 175 200 100 %}",
			expected: "88",
		},
		{
			name:     "half rounds to even (0.5 -> 0)",
			template: "{% widthratio 1 200 100 %}",
			expected: "0",
		},
		{
			name:     "half rounds to even (2.5 -> 2)",
			template: "{% widthratio 1 20 50 %}",
			expected: "2",
		},
		{
			name:     "non-half rounds normally (33.33 -> 33)",
			template: "{% widthratio 1 3 100 %}",
			expected: "33",
		},
		{
			name:     "exact value (50.0 -> 50)",
			template: "{% widthratio 50 100 100 %}",
			expected: "50",
		},
		{
			name:     "zero max_value returns 0",
			template: "{% widthratio 50 0 100 %}",
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}
			ctx := tt.context
			if ctx == nil {
				ctx = pongo2.Context{}
			}
			result, err := tpl.Execute(ctx)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("%s\ngot:  %q\nwant: %q", tt.template, result, tt.expected)
			}
		})
	}
}

func TestBugTruncatecharsHTMLByteVsRuneComparison(t *testing.T) {
	// Bug: truncatechars_html's finalize function checks `textcounter < len(value)`
	// where textcounter counts runes but len(value) counts bytes. For multi-byte
	// UTF-8 strings, the byte length is always greater than the rune count, causing
	// the ellipsis to be added incorrectly when the full text fits within the limit.

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "multi-byte chars exactly at limit should not add ellipsis",
			template: `{{ "你好世界"|truncatechars_html:4 }}`,
			expected: "你好世界", // 4 chars, limit 4, no truncation needed
		},
		{
			name:     "multi-byte chars over limit adds ellipsis",
			template: `{{ "你好世界天"|truncatechars_html:4 }}`,
			expected: "你好世…", // 5 chars, limit 4 -> 3 chars + ellipsis
		},
		{
			name:     "ASCII exactly at limit no ellipsis",
			template: `{{ "abcd"|truncatechars_html:4 }}`,
			expected: "abcd",
		},
		{
			name:     "multi-byte in HTML at limit",
			template: `{{ "<p>你好世界</p>"|truncatechars_html:4 }}`,
			expected: "<p>你好世界</p>", // 4 text chars, limit 4, HTML tags don't count
		},
		// Unusual / edge cases with bare angle brackets and malformed HTML
		{
			name:     "bare less-than treated as tag opener",
			template: `{{ "a < b"|truncatechars_html:5 }}`,
			expected: "a < b", // '<' starts a "tag", ' b' becomes tag content; only 2 text runes ("a" and " "), fits in limit
		},
		{
			name:     "bare greater-than counted as text",
			template: `{{ "a > b > c"|truncatechars_html:3 }}`,
			expected: "a …", // countHTMLTextRunes counts 7 text runes (> doesn't start/end tags outside a tag), truncation at 2 chars + ellipsis
		},
		{
			name:     "only less-than sign",
			template: `{{ "<"|truncatechars_html:5 }}`,
			expected: "<", // '<' starts tag, no text runes -> no truncation
		},
		{
			name:     "only greater-than sign",
			template: `{{ ">"|truncatechars_html:5 }}`,
			expected: ">", // '>' is counted by countHTMLTextRunes as ending tag state (but we're not in tag) -> 0 text runes? Let's check
		},
		{
			name:     "unclosed tag at end no truncation",
			template: `{{ "hello<br"|truncatechars_html:5 }}`,
			expected: "hello<br", // 5 text chars "hello" fits limit; early return preserves raw string including unclosed tag
		},
		{
			name:     "self-closing tag not counted",
			template: `{{ "<br/>hello"|truncatechars_html:5 }}`,
			expected: "<br/>hello", // 5 text runes, limit 5 -> no truncation
		},
		{
			name:     "nested tags with truncation",
			template: `{{ "<div><p>hello world</p></div>"|truncatechars_html:7 }}`,
			expected: "<div><p>hello …</p></div>", // 11 text chars, limit 7 -> 6 chars + ellipsis
		},
		{
			name:     "empty string",
			template: `{{ ""|truncatechars_html:5 }}`,
			expected: "", // no text at all
		},
		{
			name:     "only tags no text",
			template: `{{ "<p><br/></p>"|truncatechars_html:5 }}`,
			expected: "<p><br/></p>", // 0 text runes, fits in limit
		},
		{
			name:     "limit zero",
			template: `{{ "hello"|truncatechars_html:0 }}`,
			expected: "…", // over limit, 0-1 = max(0) -> immediate ellipsis
		},
		{
			name:     "limit one with text",
			template: `{{ "hi"|truncatechars_html:1 }}`,
			expected: "…", // 2 text runes > 1, reserve 1 for ellipsis -> 0 chars + ellipsis
		},
		{
			name:     "consecutive angle brackets",
			template: `{{ "<<>>abc"|truncatechars_html:10 }}`,
			expected: "<<>>abc", // fits in limit
		},
		{
			name:     "greater-than before less-than no truncation",
			template: `{{ "a>b<c"|truncatechars_html:10 }}`,
			expected: "a>b<c", // 2 text runes ("a", "b") fits limit; early return preserves raw string
		},
		{
			name:     "unclosed tag with truncation",
			template: `{{ "hello world<br"|truncatechars_html:5 }}`,
			expected: "hell…", // 11 text runes > 5; truncate at 4 + ellipsis; trailing <br without > is consumed but not pushed to stack
		},
		{
			name:     "bare less-than with truncation",
			template: `{{ "abcde < fgh"|truncatechars_html:4 }}`,
			expected: "abc…", // '<' starts tag so only "abcde " is 6 text runes by countHTMLTextRunes; truncate at 3 + ellipsis
		},
		{
			name:     "bare greater-than with truncation",
			template: `{{ "a>bcdefgh"|truncatechars_html:4 }}`,
			expected: "a>b…", // helper treats '>' as text when not in tag; 9 text runes > 4; truncate at 3 + ellipsis
		},
		{
			name:     "multiple unclosed tags with truncation",
			template: `{{ "<b><i>hello world</i></b>"|truncatechars_html:8 }}`,
			expected: "<b><i>hello w…</i></b>", // 11 text chars, limit 8 -> 7 + ellipsis
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}
			result, err := tpl.Execute(nil)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBugSlugifyAccentedCharacters(t *testing.T) {
	// Bug: slugify strips all non-ASCII characters without first normalizing
	// via NFKD. Django's slugify normalizes accented characters to their base
	// form (e.g., é → e) before stripping remaining non-ASCII. This means
	// pongo2 produces "hllo" for "Héllo" instead of Django's "hello".

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "accented e is normalized to e",
			template: `{{ "Héllo Wörld"|slugify }}`,
			expected: "hello-world",
		},
		{
			name:     "German umlauts",
			template: `{{ "Über uns"|slugify }}`,
			expected: "uber-uns",
		},
		{
			name:     "French accents",
			template: `{{ "Café résumé"|slugify }}`,
			expected: "cafe-resume",
		},
		{
			name:     "pure ASCII unchanged",
			template: `{{ "Hello World"|slugify }}`,
			expected: "hello-world",
		},
		{
			name:     "CJK characters still stripped",
			template: `{{ "hello 世界"|slugify }}`,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}
			result, err := tpl.Execute(nil)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBugDictsortNumericFieldStringSorted(t *testing.T) {
	// dictsort on a numeric field should sort numerically, not lexicographically.
	// With string sorting, "5" > "30" because '5' > '3', giving wrong order.
	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "single-digit vs multi-digit numeric sorting",
			template: `{% for item in items|dictsort:"age" %}{{ item.age }}{% if not forloop.Last %},{% endif %}{% endfor %}`,
			context: pongo2.Context{
				"items": []map[string]any{
					{"name": "Charlie", "age": 30},
					{"name": "Alice", "age": 5},
					{"name": "Bob", "age": 25},
				},
			},
			expected: "5,25,30",
		},
		{
			name:     "dictsortreversed with numeric field",
			template: `{% for item in items|dictsortreversed:"age" %}{{ item.age }}{% if not forloop.Last %},{% endif %}{% endfor %}`,
			context: pongo2.Context{
				"items": []map[string]any{
					{"name": "Alice", "age": 5},
					{"name": "Bob", "age": 25},
					{"name": "Charlie", "age": 30},
				},
			},
			expected: "30,25,5",
		},
		{
			name:     "float values sorted numerically",
			template: `{% for item in items|dictsort:"score" %}{{ item.score }}{% if not forloop.Last %},{% endif %}{% endfor %}`,
			context: pongo2.Context{
				"items": []map[string]any{
					{"name": "A", "score": 9.5},
					{"name": "B", "score": 10.1},
					{"name": "C", "score": 2.3},
				},
			},
			expected: "2.300000,9.500000,10.100000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := pongo2.FromString(tt.template)
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
