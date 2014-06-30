package pongo2

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	matches, err := filepath.Glob("./template_tests/*.tpl")
	if err != nil {
		t.Fatal(err)
	}
	for idx, match := range matches {
		t.Logf("[%3d] Testing '%s'", idx+1, match)
		tpl, err := FromFile(match)
		if err != nil {
			t.Fatal(err)
		}
		test_out, err := ioutil.ReadFile(fmt.Sprintf("%s.out", match))
		if err != nil {
			t.Fatal(err)
		}
		tpl_out, err := tpl.Execute(&Context{
			"number":        42,
			"name":          "flosch",
			"included_file": "INCLUDES.helper",
			"nil":           nil,
			"uint":          uint(8),
			"str":           "string",
			"bool_true":     true,
			"bool_false":    false,
		})
		if err != nil {
			t.Fatal(err)
		}
		if string(test_out) != tpl_out {
			t.Logf("rendered = '%s'\n", tpl_out)
			t.Fatalf("Failed: test_out != tpl_out for %s", match)
		}
	}
}
