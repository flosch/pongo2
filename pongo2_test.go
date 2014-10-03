package pongo2

import (
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	tpl *Template
}

var (
	_           = Suite(&TestSuite{})
	test_suite2 = NewSet("test suite 2")
)

func parseTemplate(s string, c Context) string {
	t, err := test_suite2.FromString(s)
	if err != nil {
		panic(err)
	}
	out, err := t.Execute(c)
	if err != nil {
		panic(err)
	}
	return out
}

func parseTemplateFn(s string, c Context) func() {
	return func() {
		parseTemplate(s, c)
	}
}

func (s *TestSuite) TestMisc(c *C) {
	// Must
	// TODO: Add better error message (see issue #18)
	c.Check(func() { Must(test_suite2.FromFile("template_tests/inheritance/base2.tpl")) },
		PanicMatches,
		`\[Error \(where: fromfile\) in template_tests/inheritance/doesnotexist.tpl | Line 1 Col 12 near 'doesnotexist.tpl'\] open template_tests/inheritance/doesnotexist.tpl: no such file or directory`)

	// Context
	c.Check(parseTemplateFn("", Context{"'illegal": nil}), PanicMatches, ".*not a valid identifier.*")

	// Registers
	c.Check(func() { RegisterFilter("escape", nil) }, PanicMatches, ".*is already registered.*")
	c.Check(func() { RegisterTag("for", nil) }, PanicMatches, ".*is already registered.*")

	// ApplyFilter
	v, err := ApplyFilter("title", AsValue("this is a title"), nil)
	if err != nil {
		c.Fatal(err)
	}
	c.Check(v.String(), Equals, "This Is A Title")
	c.Check(func() {
		_, err := ApplyFilter("doesnotexist", nil, nil)
		if err != nil {
			panic(err)
		}
	}, PanicMatches, `\[Error \(where: applyfilter\)\] Filter with name 'doesnotexist' not found.`)
}
