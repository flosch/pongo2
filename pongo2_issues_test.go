package pongo2_test

import (
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

func TestIssue342(t *testing.T) {
	// Test that adding a numeric string and a number results in arithmetic addition.
	// Bug: In v6, "10" + 5 returns "105" (string concatenation).
	// Expected: "10" + 5 should return 15 (arithmetic addition) as in v4.
	// See: https://github.com/flosch/pongo2/issues/342

	tests := []struct {
		name     string
		template string
		context  pongo2.Context
		expected string
	}{
		{
			name:     "numeric string + integer",
			template: "{{ a + b }}",
			context:  pongo2.Context{"a": "10", "b": 5},
			expected: "15",
		},
		{
			name:     "integer + numeric string",
			template: "{{ a + b }}",
			context:  pongo2.Context{"a": 5, "b": "10"},
			expected: "15",
		},
		{
			name:     "numeric string + float",
			template: "{{ a + b }}",
			context:  pongo2.Context{"a": "10.5", "b": 2.5},
			expected: "13.000000",
		},
		{
			name:     "float + numeric string",
			template: "{{ a + b }}",
			context:  pongo2.Context{"a": 2.5, "b": "10.5"},
			expected: "13.000000",
		},
		{
			name:     "non-numeric string + integer stays string concatenation",
			template: "{{ a + b }}",
			context:  pongo2.Context{"a": "hello", "b": 5},
			expected: "hello5",
		},
		{
			name:     "integer + non-numeric string stays string concatenation",
			template: "{{ a + b }}",
			context:  pongo2.Context{"a": 5, "b": "hello"},
			expected: "5hello",
		},
		{
			name:     "two strings remain concatenation",
			template: "{{ a + b }}",
			context:  pongo2.Context{"a": "hello", "b": "world"},
			expected: "helloworld",
		},
		{
			name:     "two numeric strings remain concatenation",
			template: "{{ a + b }}",
			context:  pongo2.Context{"a": "10", "b": "5"},
			expected: "105",
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
