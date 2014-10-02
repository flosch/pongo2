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
	name string

	// Globals will be provided to all templates created within this template set
	Globals Context

	// If debug is true (default false), ExecutionContext.Logf() will work and output to STDOUT.
	Debug bool

	// Base directory: If you set the base directory (string is non-empty), all filename lookups in tags/filters are
	// relative to this directory. If it's empty, all lookups are relative to the current filename which is importing.
	BaseDirectory string

	// Sandbox features
	//  - Limit access to directories (using SandboxDirectories)
	//  - Disallow access to specific tags and/or filters (using BanTag() and BanFilter())
	//
	// You can limit file accesses (for all tags/filters which are using pongo2's file resolver technique)
	// to these sandbox directories. All default pongo2 filters/tags are respecting these restrictions.
	// For example, if you only have your base directory in the list, a {% ssi "/etc/passwd" %} will not work.
	// No items in SandboxDirectories means no restrictions at all.
	//
	// For efficiency reasons you can ban tags/filters only *before* you have added your first
	// template to the set (restrictions are statically checked). After you added one, it's not possible anymore
	// (for your personal security).
	//
	// SandboxDirectories can be changed at runtime. Please synchronize the access to it if you need to change it
	// after you've added your first template to the set. You *must* use this match pattern for your directories:
	//  http://golang.org/pkg/path/filepath/#Match
	SandboxDirectories   []string
	firstTemplateCreated bool
	bannedTags           map[string]bool
	bannedFilters        map[string]bool
}

// Create your own template sets to separate different kind of templates (e. g. web from mail templates) with
// different globals or other configurations (like base directories).
func NewSet(name string) *TemplateSet {
	return &TemplateSet{
		name:          name,
		Globals:       make(Context),
		bannedTags:    make(map[string]bool),
		bannedFilters: make(map[string]bool),
	}
}

// Ban a specific tag for this template set. See more in the documentation for TemplateSet.
func (set *TemplateSet) BanTag(name string) {
	_, has := tags[name]
	if !has {
		panic(fmt.Sprintf("Tag '%s' not found.", name))
	}
	if set.firstTemplateCreated {
		panic("You cannot ban any tags after you've added your first template to your template set.")
	}
	_, has = set.bannedTags[name]
	if has {
		panic(fmt.Sprintf("Tag '%s' is already banned.", name))
	}
	set.bannedTags[name] = true
}

// Ban a specific filter for this template set. See more in the documentation for TemplateSet.
func (set *TemplateSet) BanFilter(name string) {
	_, has := filters[name]
	if !has {
		panic(fmt.Sprintf("Filter '%s' not found.", name))
	}
	if set.firstTemplateCreated {
		panic("You cannot ban any filters after you've added your first template to your template set.")
	}
	_, has = set.bannedFilters[name]
	if has {
		panic(fmt.Sprintf("Filter '%s' is already banned.", name))
	}
	set.bannedFilters[name] = true
}

// Loads  a template from string and returns a Template instance.
func (set *TemplateSet) FromString(tpl string) (*Template, error) {
	set.firstTemplateCreated = true

	return newTemplateString(set, tpl)
}

// Loads  a template from a filename and returns a Template instance.
// The filename must either be relative to the application's directory
// or be an absolute path.
func (set *TemplateSet) FromFile(filename string) (*Template, error) {
	set.firstTemplateCreated = true

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return newTemplate(set, filename, false, string(buf))
}

// Shortcut; renders a template string directly. Panics when providing a
// malformed template or an error occurs during execution.
func (set *TemplateSet) RenderTemplateString(s string, ctx Context) string {
	set.firstTemplateCreated = true

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
	set.firstTemplateCreated = true

	tpl := Must(set.FromFile(fn))
	result, err := tpl.Execute(ctx)
	if err != nil {
		panic(err)
	}
	return result
}

func (set *TemplateSet) logf(format string, args ...interface{}) {
	if set.Debug {
		logger.Printf(fmt.Sprintf("[template set: %s] %s", set.name, format), args...)
	}
}

func (set *TemplateSet) resolveFilename(tpl *Template, filename string) (resolved_path string) {
	if len(set.SandboxDirectories) > 0 {
		defer func() {
			// Remove any ".." or other crap
			resolved_path = filepath.Clean(resolved_path)

			// Make the path absolute
			abs_path, err := filepath.Abs(resolved_path)
			if err != nil {
				panic(err)
			}
			resolved_path = abs_path

			// Check against the sandbox directories (once one pattern matches, we're done and can allow it)
			for _, pattern := range set.SandboxDirectories {
				matched, err := filepath.Match(pattern, resolved_path)
				if err != nil {
					panic("Wrong sandbox directory match pattern (see http://golang.org/pkg/path/filepath/#Match).")
				}
				if matched {
					// OK!
					return
				}
			}

			// No pattern matched, we have to log+deny the request
			set.logf("Access attempt outside of the sandbox directories (blocked): '%s'", resolved_path)
			resolved_path = ""
		}()
	}

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
