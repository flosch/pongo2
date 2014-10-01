package pongo2

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type TemplateSet struct {
	name      string
	templates []*Template

	// Globals will be provided to all templates created within this template set
	Globals Context

	// If debug is true (default false), ExecutionContext.Logf() will work and output to STDOUT.
	Debug bool
}

func NewSet(name string) *TemplateSet {
	return &TemplateSet{
		name:    name,
		Globals: make(Context),
	}
}

// Loads  a template from string and returns a Template instance.
func (set *TemplateSet) FromString(tpl string) (*Template, error) {
	t, err := newTemplateString(set, tpl)
	return t, err
}

// Loads  a template from a filename and returns a Template instance.
// The filename must either be relative to the application's directory
// or be an absolute path.
func (set *TemplateSet) FromFile(filename string) (*Template, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	t, err := newTemplate(set, filename, string(buf))
	return t, err
}

// Shortcut; renders a template string directly. Panics when providing a
// malformed template or an error occurs during execution.
func (set *TemplateSet) RenderTemplateString(s string, ctx Context) string {
	tpl := Must(FromString(s))
	result, err := tpl.Execute(ctx)
	if err != nil {
		panic(err)
	}
	return result
}

// Shortcut; renders a template file directly. Panics when providing a
// malformed template or an error occurs during execution.
func (set *TemplateSet) RenderTemplateFile(fn string, ctx Context) string {
	tpl := Must(FromFile(fn))
	result, err := tpl.Execute(ctx)
	if err != nil {
		panic(err)
	}
	return result
}

var (
	logger = log.New(os.Stdout, "[pongo2] ", log.LstdFlags)
)

func (set *TemplateSet) logf(format string, args ...interface{}) {
	if set.Debug {
		logger.Printf(fmt.Sprintln("[%s] %s", set.name, format), args...)
	}
}

// Logging function (internally used)
func logf(format string, items ...interface{}) {
	if debug {
		logger.Printf(format, items...)
	}
}

var (
	debug bool // internal debugging

	// Creating a default set
	defaultSet = NewSet("default")

	// Methods on the default set
	FromString           = defaultSet.FromString
	FromFile             = defaultSet.FromFile
	RenderTemplateString = defaultSet.RenderTemplateString
	RenderTemplateFile   = defaultSet.RenderTemplateFile

	// Globals for the default set
	Globals = defaultSet.Globals
)
