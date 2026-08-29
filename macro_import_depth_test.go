package pongo2_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/flosch/pongo2/v7"
)

func TestMacroDepthLimitCoversDefinitionsAndImports(t *testing.T) {
	t.Parallel()
	tests := map[string]fstest.MapFS{
		"definition": {
			"main.tpl": {Data: []byte(
				`{% macro rec(n) %}{% if n > 0 %}{{ rec(n-1) }}{% endif %}{% endmacro %}{{ rec(1100) }}`)},
		},
		"import": {
			"main.tpl": {Data: []byte(`{% import "macros.tpl" rec %}{{ rec(1100) }}`)},
			"macros.tpl": {Data: []byte(
				`{% macro rec(n) export %}{% if n > 0 %}{{ rec(n-1) }}{% endif %}{% endmacro %}`)},
		},
	}
	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			set := pongo2.NewSet(name, pongo2.NewFSLoader(files))
			tpl, err := set.FromFile("main.tpl")
			if err != nil {
				t.Fatalf("FromFile: %v", err)
			}
			_, err = tpl.Execute(nil)
			if err == nil || !strings.Contains(err.Error(),
				"maximum recursive macro call depth reached (max is 1000)") {
				t.Fatalf("Execute error = %v, want macro depth limit", err)
			}
		})
	}
}
