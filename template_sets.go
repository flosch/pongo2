package pongo2

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// A template set allows you to create your own group of templates with their own global context (which is shared
// among all members of the set), their own configuration (like a specific base directory) and their own sandbox.
// It's useful for a separation of different kind of templates (e. g. web templates vs. mail templates).
type TemplateSet struct {
	name string

	// Globals will be provided to all templates created within this template set
	Globals Context

	// If debug is true (default false), ExecutionContext.Logf() will work and output to STDOUT. Furthermore,
	// FromCache() won't cache the templates. Make sure to synchronize the access to it in case you're changing this
	// variable during program execution (and template compilation/execution).
	Debug bool

	// The TemplateLoader specifies the loader which fulfills all template loading requests (From*-functions as well
	// as include/extends-tags). You can either use a built-in template loader or use your own
	// (for example for enforcing some restrictions on paths or simulating a file system).
	loader TemplateLoader

	// Sandbox feature:
	// Disallow access to specific tags and/or filters (using BanTag() and BanFilter()).
	//
	// For efficiency reasons you can ban tags/filters only *before* you have added your first
	// template to the set (restrictions are statically checked). After you added one, it's not possible anymore
	// (for your personal security).
	firstTemplateCreated bool
	bannedTags           map[string]bool
	bannedFilters        map[string]bool

	// Template cache (for FromCache())
	templateCache      map[string]*Template
	templateCacheMutex sync.Mutex
}

// Create your own template sets to separate different kind of templates (e. g. web from mail templates) with
// different globals or other configurations (like base directories). Set loader to nil, if you want to
// use a default filesystem loader without setting any base directory and path restrictions.
func NewSet(name string, loader TemplateLoader) *TemplateSet {
	if loader == nil {
		loader = NewFilesystemLoader("", nil)
	}
	return &TemplateSet{
		name:          name,
		loader:        loader,
		Globals:       make(Context),
		bannedTags:    make(map[string]bool),
		bannedFilters: make(map[string]bool),
		templateCache: make(map[string]*Template),
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

// FromCache() is a convenient method to cache templates. It is thread-safe
// and will only compile the template associated with a filename once.
// If TemplateSet.Debug is true (for example during development phase),
// FromCache() will not cache the template and instead recompile it on any
// call (to make changes to a template live instantaneously).
// Like FromFile(), FromCache() takes a relative path to a set base directory.
// Sandbox restrictions apply (if given).
func (set *TemplateSet) FromCache(filename string) (*Template, error) {
	if set.Debug {
		// Recompile on any request
		return set.FromFile(filename)
	} else {
		// Cache the template
		cleaned_filename := set.loader.AbsPath(nil, filename)

		set.templateCacheMutex.Lock()
		defer set.templateCacheMutex.Unlock()

		tpl, has := set.templateCache[cleaned_filename]

		// Cache miss
		if !has {
			tpl, err := set.FromFile(cleaned_filename)
			if err != nil {
				return nil, err
			}
			set.templateCache[cleaned_filename] = tpl
			return tpl, nil
		}

		// Cache hit
		return tpl, nil
	}
}

// Loads  a template from string and returns a Template instance.
func (set *TemplateSet) FromString(tpl string) (*Template, error) {
	set.firstTemplateCreated = true

	return newTemplateString(set, tpl)
}

// Receives content from the template loader according to the filename,
// parses the file and returns a compiled Template instance.
func (set *TemplateSet) FromFile(filename string) (*Template, error) {
	set.firstTemplateCreated = true

	rd, err := set.loader.OpenFile(filename)
	if err != nil {
		return nil, &Error{
			Filename: filename,
			Sender:   "fromfile",
			ErrorMsg: err.Error(),
		}
	}
	return set.FromReader(filename, rd)
}

// Receives content from the template loader according to the filename,
// parses the file and returns a compiled Template instance.
func (set *TemplateSet) FromReader(filename string, reader io.ReadCloser) (*Template, error) {
	set.firstTemplateCreated = true

	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, &Error{
			Filename: filename,
			Sender:   "fromreader",
			ErrorMsg: err.Error(),
		}
	}

	err = reader.Close()
	if err != nil {
		panic(err)
	}

	return newTemplate(set, set.loader.AbsPath(nil, filename), false, string(buf))
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
	DefaultSet = NewSet("default", NewFilesystemLoader("", nil))

	// Methods on the default set
	FromString           = DefaultSet.FromString
	FromFile             = DefaultSet.FromFile
	FromCache            = DefaultSet.FromCache
	RenderTemplateString = DefaultSet.RenderTemplateString
	RenderTemplateFile   = DefaultSet.RenderTemplateFile

	// Globals for the default set
	Globals = DefaultSet.Globals
)
