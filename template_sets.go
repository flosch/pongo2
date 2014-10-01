package pongo2

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// A template set allows you to create your own template sets with your own global context (which is shared
// among all members of the set) and your own configuration (like a specific base directory).
// It's useful for a separation of different kind of templates (e. g. web templates vs. mail templates).
type TemplateSet struct {
	name      string
	templates []*Template

	// Globals will be provided to all templates created within this template set
	Globals Context

	// If debug is true (default false), ExecutionContext.Logf() will work and output to STDOUT.
	Debug bool

	// Base directory: If you set the base directory (string is non-empty), all filename lookups in tags/filters are
	// relative to this directory. If it's empty, all lookups are relative to the current filename which is importing.
	BaseDirectory string
}

// Create your own template sets to separate different kind of templates (e. g. web from mail templates) with
// different globals or other configurations (like base directories).
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
	t, err := newTemplate(set, filename, false, string(buf))
	return t, err
}

// Shortcut; renders a template string directly. Panics when providing a
// malformed template or an error occurs during execution.
func (set *TemplateSet) RenderTemplateString(s string, ctx Context) string {
	tpl := Must(set.FromString(s))
	result, err := tpl.Execute(ctx)
	if err != nil {
		panic(err)
	}
	return result
}

// Shortcut; renders a template file directly. Panics when providing a
// malformed template or an error occurs during execution.
func (set *TemplateSet) RenderTemplateFile(fn string, ctx Context) string {
	tpl := Must(set.FromFile(fn))
	result, err := tpl.Execute(ctx)
	if err != nil {
		panic(err)
	}
	return result
}

func (set *TemplateSet) logf(format string, args ...interface{}) {
	if set.Debug {
		logger.Printf(fmt.Sprintln("[%s] %s", set.name, format), args...)
	}
}

func (set *TemplateSet) resolveFilename(tpl *Template, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}

	if set.BaseDirectory == "" {
		if tpl.is_tpl_string {
			return filename
		}
		base := filepath.Dir(tpl.name)
		return filepath.Join(base, filename)
	} else {
		return filepath.Join(set.BaseDirectory, filename)
	}
}

// Logging function (internally used)
func logf(format string, items ...interface{}) {
	if debug {
		logger.Printf(format, items...)
	}
}

var (
	debug  bool // internal debugging
	logger = log.New(os.Stdout, "[pongo2] ", log.LstdFlags)

	// Creating a default set
	DefaultSet = NewSet("default")

	// Methods on the default set
	FromString           = DefaultSet.FromString
	FromFile             = DefaultSet.FromFile
	RenderTemplateString = DefaultSet.RenderTemplateString
	RenderTemplateFile   = DefaultSet.RenderTemplateFile

	// Globals for the default set
	Globals = DefaultSet.Globals
)
