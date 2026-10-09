package pongo2_test

import (
	"sync"
	"testing"

	"github.com/flosch/pongo2/v6"
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
	for i := 0; i < 5; i++ {
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
	for i := 0; i < 20; i++ {
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
	for i := 0; i < 20; i++ {
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

func TestIssue386(t *testing.T) {
	// TrimBlocks edits the shared tokens. A second Execute must not
	// drop another newline.
	tpl, err := pongo2.FromString("a\n{% if x %}\nb\n{% endif %}\n\nc\n")
	if err != nil {
		t.Fatal(err)
	}
	tpl.Options.TrimBlocks = true
	tpl.Options.LStripBlocks = true

	first, err := tpl.Execute(pongo2.Context{"x": true})
	if err != nil {
		t.Fatal(err)
	}
	const want = "a\nb\n\nc\n"
	if first != want {
		t.Fatalf("first render %q, want %q", first, want)
	}
	second, err := tpl.Execute(pongo2.Context{"x": true})
	if err != nil {
		t.Fatal(err)
	}
	if second != want {
		t.Fatalf("second render %q, want %q", second, want)
	}
}
