package pongo2_test

import (
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/flosch/pongo2/v6"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var (
	_          = Suite(&TestSuite{})
	testSuite2 = pongo2.NewSet("test suite 2", pongo2.MustNewLocalFileSystemLoader(""))
)

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

func (s *TestSuite) TestMisc(c *C) {
	// Must
	// TODO: Add better error message (see issue #18)
	c.Check(
		func() { pongo2.Must(testSuite2.FromFile("template_tests/inheritance/base2.tpl")) },
		PanicMatches,
		`\[Error \(where: fromfile\) in .*template_tests[/\\]inheritance[/\\]doesnotexist.tpl | Line 1 Col 12 near 'doesnotexist.tpl'\] open .*template_tests[/\\]inheritance[/\\]doesnotexist.tpl: no such file or directory`,
	)

	// Context
	c.Check(parseTemplateFn("", pongo2.Context{"'illegal": nil}), PanicMatches, ".*not a valid identifier.*")

	// Registers
	c.Check(pongo2.RegisterFilter("escape", nil).Error(), Matches, ".*is already registered")
	c.Check(pongo2.RegisterTag("for", nil).Error(), Matches, ".*is already registered")

	// ApplyFilter
	v, err := pongo2.ApplyFilter("title", pongo2.AsValue("this is a title"), nil)
	if err != nil {
		c.Fatal(err)
	}
	c.Check(v.String(), Equals, "This Is A Title")
	c.Check(func() {
		_, err := pongo2.ApplyFilter("doesnotexist", nil, nil)
		if err != nil {
			panic(err)
		}
	}, PanicMatches, `\[Error \(where: applyfilter\)\] filter with name 'doesnotexist' not found`)
}

func (s *TestSuite) TestImplicitExecCtx(c *C) {
	tpl, err := pongo2.FromString("{{ ImplicitExec }}")
	if err != nil {
		c.Fatalf("Error in FromString: %v", err)
	}

	val := "a stringy thing"

	res, err := tpl.Execute(pongo2.Context{
		"Value": val,
		"ImplicitExec": func(ctx *pongo2.ExecutionContext) string {
			return ctx.Public["Value"].(string)
		},
	})
	if err != nil {
		c.Fatalf("Error executing template: %v", err)
	}

	c.Check(res, Equals, val)

	// The implicit ctx should not be persisted from call-to-call
	res, err = tpl.Execute(pongo2.Context{
		"ImplicitExec": func() string {
			return val
		},
	})

	if err != nil {
		c.Fatalf("Error executing template: %v", err)
	}

	c.Check(res, Equals, val)
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
		buf, err := ioutil.ReadFile(tplPath)
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
			out.Execute(mycontext)
		}
	})
}
