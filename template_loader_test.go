package pongo2_test

import (
	"embed"
	"testing"

	"github.com/flosch/pongo2/v4"
)

var (
	//go:embed template_tests
	testTemplateFS embed.FS
)

func TestEmbedFSLoader(t *testing.T) {
	set := pongo2.NewSet("test embed", pongo2.MustNewEmbededLoader(testTemplateFS))
	_ = pongo2.Must(set.FromFile("template_tests/complex.tpl"))
	_ = pongo2.Must(set.FromFile("template_tests/inheritance/base.tpl"))
}
