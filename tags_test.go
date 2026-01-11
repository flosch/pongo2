package pongo2

import "testing"

func TestReplaceTag(t *testing.T) {
	t.Run("non-existent tag", func(t *testing.T) {
		err := ReplaceTag("nonexistent_tag_xyz", func(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
			return nil, nil
		})
		if err == nil {
			t.Error("ReplaceTag should return error for non-existent tag")
		}
	})

	t.Run("existing tag", func(t *testing.T) {
		originalTag := tags["comment"]
		defer func() { tags["comment"] = originalTag }()

		newParser := func(doc *Parser, start *Token, arguments *Parser) (INodeTag, error) {
			return originalTag.parser(doc, start, arguments)
		}

		err := ReplaceTag("comment", newParser)
		if err != nil {
			t.Errorf("ReplaceTag failed for existing tag: %v", err)
		}
	})
}

func TestTagIfEqual(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "equal strings",
			template: "{% ifequal a b %}equal{% endifequal %}",
			context:  Context{"a": "test", "b": "test"},
			expected: "equal",
		},
		{
			name:     "not equal strings",
			template: "{% ifequal a b %}equal{% endifequal %}",
			context:  Context{"a": "test", "b": "other"},
			expected: "",
		},
		{
			name:     "equal integers",
			template: "{% ifequal a b %}equal{% endifequal %}",
			context:  Context{"a": 42, "b": 42},
			expected: "equal",
		},
		{
			name:     "not equal integers",
			template: "{% ifequal a b %}equal{% endifequal %}",
			context:  Context{"a": 42, "b": 43},
			expected: "",
		},
		{
			name:     "with else - equal",
			template: "{% ifequal a b %}equal{% else %}not equal{% endifequal %}",
			context:  Context{"a": "same", "b": "same"},
			expected: "equal",
		},
		{
			name:     "with else - not equal",
			template: "{% ifequal a b %}equal{% else %}not equal{% endifequal %}",
			context:  Context{"a": "same", "b": "different"},
			expected: "not equal",
		},
		{
			name:     "literal strings equal",
			template: `{% ifequal "hello" "hello" %}equal{% endifequal %}`,
			context:  Context{},
			expected: "equal",
		},
		{
			name:     "literal strings not equal",
			template: `{% ifequal "hello" "world" %}equal{% endifequal %}`,
			context:  Context{},
			expected: "",
		},
		{
			name:     "literal integers",
			template: "{% ifequal 5 5 %}equal{% endifequal %}",
			context:  Context{},
			expected: "equal",
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

func TestTagIfNotEqual(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "not equal strings",
			template: "{% ifnotequal a b %}not equal{% endifnotequal %}",
			context:  Context{"a": "test", "b": "other"},
			expected: "not equal",
		},
		{
			name:     "equal strings",
			template: "{% ifnotequal a b %}not equal{% endifnotequal %}",
			context:  Context{"a": "test", "b": "test"},
			expected: "",
		},
		{
			name:     "not equal integers",
			template: "{% ifnotequal a b %}not equal{% endifnotequal %}",
			context:  Context{"a": 42, "b": 43},
			expected: "not equal",
		},
		{
			name:     "equal integers",
			template: "{% ifnotequal a b %}not equal{% endifnotequal %}",
			context:  Context{"a": 42, "b": 42},
			expected: "",
		},
		{
			name:     "with else - not equal",
			template: "{% ifnotequal a b %}not equal{% else %}equal{% endifnotequal %}",
			context:  Context{"a": "same", "b": "different"},
			expected: "not equal",
		},
		{
			name:     "with else - equal",
			template: "{% ifnotequal a b %}not equal{% else %}equal{% endifnotequal %}",
			context:  Context{"a": "same", "b": "same"},
			expected: "equal",
		},
		{
			name:     "literal strings not equal",
			template: `{% ifnotequal "hello" "world" %}not equal{% endifnotequal %}`,
			context:  Context{},
			expected: "not equal",
		},
		{
			name:     "literal strings equal",
			template: `{% ifnotequal "hello" "hello" %}not equal{% endifnotequal %}`,
			context:  Context{},
			expected: "",
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

func TestTagLorem(t *testing.T) {
	tests := []struct {
		name     string
		template string
	}{
		{"default lorem", "{% lorem %}"},
		{"lorem with count", "{% lorem 3 %}"},
		{"lorem words", "{% lorem 5 w %}"},
		{"lorem paragraphs", "{% lorem 2 p %}"},
		{"lorem random", "{% lorem 3 w random %}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := FromString(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tpl.Execute(Context{})
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if len(result) == 0 {
				t.Error("Lorem should produce output")
			}
		})
	}
}

func TestTagWidthratio(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "simple ratio",
			template: "{% widthratio 50 100 200 %}",
			context:  Context{},
			expected: "101",
		},
		{
			name:     "with variables",
			template: "{% widthratio value max_value width %}",
			context:  Context{"value": 50, "max_value": 100, "width": 200},
			expected: "101",
		},
		{
			name:     "exact ratio",
			template: "{% widthratio 25 100 200 %}",
			context:  Context{},
			expected: "51",
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

func TestTagTemplatetag(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"openblock", "{% templatetag openblock %}", "{%"},
		{"closeblock", "{% templatetag closeblock %}", "%}"},
		{"openvariable", "{% templatetag openvariable %}", "{{"},
		{"closevariable", "{% templatetag closevariable %}", "}}"},
		{"openbrace", "{% templatetag openbrace %}", "{"},
		{"closebrace", "{% templatetag closebrace %}", "}"},
		{"opencomment", "{% templatetag opencomment %}", "{#"},
		{"closecomment", "{% templatetag closecomment %}", "#}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := FromString(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tpl.Execute(Context{})
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTagSet(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "set simple value",
			template: "{% set myvar = 42 %}{{ myvar }}",
			context:  Context{},
			expected: "42",
		},
		{
			name:     "set string value",
			template: `{% set myvar = "hello" %}{{ myvar }}`,
			context:  Context{},
			expected: "hello",
		},
		{
			name:     "set from context",
			template: "{% set myvar = other %}{{ myvar }}",
			context:  Context{"other": "value"},
			expected: "value",
		},
		{
			name:     "set expression",
			template: "{% set result = 10 + 5 %}{{ result }}",
			context:  Context{},
			expected: "15",
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

func TestTagWith(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "with simple",
			template: "{% with a=1 %}{{ a }}{% endwith %}",
			context:  Context{},
			expected: "1",
		},
		{
			name:     "with multiple",
			template: "{% with a=1 b=2 %}{{ a }}-{{ b }}{% endwith %}",
			context:  Context{},
			expected: "1-2",
		},
		{
			name:     "with from context",
			template: "{% with local=global %}{{ local }}{% endwith %}",
			context:  Context{"global": "value"},
			expected: "value",
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

func TestTagSpaceless(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "remove spaces between tags",
			template: "{% spaceless %}<p>  </p>{% endspaceless %}",
			expected: "<p></p>",
		},
		{
			name:     "preserve content spaces",
			template: "{% spaceless %}<p>hello world</p>{% endspaceless %}",
			expected: "<p>hello world</p>",
		},
		{
			name:     "multiple tags",
			template: "{% spaceless %}<div>   <p>  </p>   </div>{% endspaceless %}",
			expected: "<div><p></p></div>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := FromString(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tpl.Execute(Context{})
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTagCycle(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "simple cycle",
			template: "{% for i in items %}{% cycle 'a' 'b' %}{% endfor %}",
			expected: "abab",
		},
		{
			name:     "cycle with as silent",
			template: "{% for i in items %}{% cycle 'x' 'y' as myvar silent %}{{ myvar }}{% endfor %}",
			expected: "xyxy",
		},
	}

	ctx := Context{"items": []int{1, 2, 3, 4}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := FromString(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			result, err := tpl.Execute(ctx)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTagAutoescape(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "autoescape on",
			template: "{% autoescape on %}{{ content }}{% endautoescape %}",
			context:  Context{"content": "<script>alert('xss')</script>"},
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "autoescape off",
			template: "{% autoescape off %}{{ content }}{% endautoescape %}",
			context:  Context{"content": "<b>bold</b>"},
			expected: "<b>bold</b>",
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

func TestTagFirstof(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "first truthy",
			template: "{% firstof a b c %}",
			context:  Context{"a": "", "b": "second", "c": "third"},
			expected: "second",
		},
		{
			name:     "all truthy",
			template: "{% firstof a b c %}",
			context:  Context{"a": "first", "b": "second", "c": "third"},
			expected: "first",
		},
		{
			name:     "with fallback",
			template: `{% firstof a b "fallback" %}`,
			context:  Context{"a": "", "b": ""},
			expected: "fallback",
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

func TestTagIfchanged(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "simple ifchanged",
			template: "{% for item in items %}{% ifchanged %}{{ item }}{% endifchanged %}{% endfor %}",
			context:  Context{"items": []string{"a", "a", "b", "b", "c"}},
			expected: "abc",
		},
		{
			name:     "ifchanged with variable",
			template: "{% for item in items %}{% ifchanged item %}{{ item }}{% endifchanged %}{% endfor %}",
			context:  Context{"items": []string{"a", "a", "b", "b", "c"}},
			expected: "abc",
		},
		{
			name:     "ifchanged with else",
			template: "{% for item in items %}{% ifchanged item %}{{ item }}{% else %}.{% endifchanged %}{% endfor %}",
			context:  Context{"items": []string{"a", "a", "b"}},
			expected: "a.b",
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

func TestForLoopVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "forloop.counter",
			template: "{% for i in items %}{{ forloop.Counter }}{% endfor %}",
			context:  Context{"items": []int{1, 2, 3}},
			expected: "123",
		},
		{
			name:     "forloop.counter0",
			template: "{% for i in items %}{{ forloop.Counter0 }}{% endfor %}",
			context:  Context{"items": []int{1, 2, 3}},
			expected: "012",
		},
		{
			name:     "forloop.first",
			template: "{% for i in items %}{% if forloop.First %}F{% endif %}{% endfor %}",
			context:  Context{"items": []int{1, 2, 3}},
			expected: "F",
		},
		{
			name:     "forloop.last",
			template: "{% for i in items %}{% if forloop.Last %}L{% endif %}{% endfor %}",
			context:  Context{"items": []int{1, 2, 3}},
			expected: "L",
		},
		{
			name:     "for with empty",
			template: "{% for i in items %}{{ i }}{% empty %}empty{% endfor %}",
			context:  Context{"items": []int{}},
			expected: "empty",
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

func TestComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		template string
		context  Context
		expected string
	}{
		{
			name:     "arithmetic",
			template: "{{ 10 + 5 * 2 }}",
			context:  Context{},
			expected: "20",
		},
		{
			name:     "comparison",
			template: "{% if 5 > 3 %}yes{% endif %}",
			context:  Context{},
			expected: "yes",
		},
		{
			name:     "boolean and",
			template: "{% if a and b %}yes{% endif %}",
			context:  Context{"a": true, "b": true},
			expected: "yes",
		},
		{
			name:     "boolean or",
			template: "{% if a or b %}yes{% endif %}",
			context:  Context{"a": false, "b": true},
			expected: "yes",
		},
		{
			name:     "not operator",
			template: "{% if not a %}yes{% endif %}",
			context:  Context{"a": false},
			expected: "yes",
		},
		{
			name:     "in operator",
			template: "{% if 2 in items %}yes{% endif %}",
			context:  Context{"items": []int{1, 2, 3}},
			expected: "yes",
		},
		{
			name:     "parentheses",
			template: "{{ (2 + 3) * 4 }}",
			context:  Context{},
			expected: "20",
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
