package pongo2

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// TemplateLoader allows to implement a virtual file system.
type TemplateLoader interface {
	// Abs calculates the path to a given template. Whenever a path must be resolved
	// due to an import from another template, the base equals the parent template's path.
	Abs(base, name string) string

	// Get returns an io.Reader where the template's content can be read from.
	Get(path string) (io.Reader, error)
}

// TemplateSet allows you to create your own group of templates with their own
// global context (which is shared among all members of the set) and their own
// configuration.
// It's useful for a separation of different kind of templates
// (e. g. web templates vs. mail templates).
type TemplateSet struct {
	name    string
	loaders []TemplateLoader

	// Globals will be provided to all templates created within this template set
	Globals Context

	// If debug is true (default false), ExecutionContext.Logf() will work and output
	// to STDOUT. Furthermore, FromCache() won't cache the templates.
	// Make sure to synchronize the access to it in case you're changing this
	// variable during program execution (and template compilation/execution).
	Debug bool

	// autoescape controls whether template output is automatically HTML-escaped.
	// When true (default), string output will be escaped for safety.
	autoescape bool

	// Options allow you to change the behavior of template-engine.
	// You can change the options before calling the Execute method.
	Options *Options

	// Per-set tag and filter registries (lazily initialized via initOnce)
	tags     map[string]*tag
	filters  map[string]FilterFunction
	initOnce sync.Once

	// Sandbox features
	// - Disallow access to specific tags and/or filters (using BanTag() and BanFilter())
	//
	// For efficiency reasons you can ban tags/filters only *before* you have
	// added your first template to the set (restrictions are statically checked).
	// After you added one, it's not possible anymore (for your personal security).
	firstTemplateCreated bool
	bannedTags           map[string]bool
	bannedFilters        map[string]bool

	// Template cache (for FromCache())
	templateCache      map[string]*Template
	templateCacheMutex sync.Mutex
}

// NewSet can be used to create sets with different kind of templates
// (e. g. web from mail templates), with different globals or
// other configurations.
func NewSet(name string, loaders ...TemplateLoader) *TemplateSet {
	if len(loaders) == 0 {
		panic(fmt.Errorf("at least one template loader must be specified"))
	}

	return &TemplateSet{
		name:          name,
		loaders:       loaders,
		Globals:       make(Context),
		autoescape:    true,
		// tags and filters are lazily initialized via initOnce
		bannedTags:    make(map[string]bool),
		bannedFilters: make(map[string]bool),
		templateCache: make(map[string]*Template),
		Options:       newOptions(),
	}
}

func (set *TemplateSet) AddLoader(loaders ...TemplateLoader) {
	set.loaders = append(set.loaders, loaders...)
}

// initBuiltins copies the builtin tags and filters into this template set.
// This is called lazily via initOnce to ensure builtinTags and builtinFilters
// have been populated by init() functions before copying.
func (set *TemplateSet) initBuiltins() {
	set.tags = copyTags(builtinTags)
	set.filters = copyFilters(builtinFilters)
}

func (set *TemplateSet) resolveFilename(tpl *Template, path string) string {
	return set.resolveFilenameForLoader(set.loaders[0], tpl, path)
}

func (set *TemplateSet) resolveFilenameForLoader(loader TemplateLoader, tpl *Template, path string) string {
	name := ""
	if tpl != nil && tpl.isTplString {
		return path
	}
	if tpl != nil {
		name = tpl.name
	}

	return loader.Abs(name, path)
}

// BanTag bans a specific tag for this template set. See more in the documentation for TemplateSet.
func (set *TemplateSet) BanTag(name string) error {
	set.initOnce.Do(set.initBuiltins)
	_, has := set.tags[name]
	if !has {
		return fmt.Errorf("tag '%s' not found", name)
	}
	if set.firstTemplateCreated {
		return errors.New("you cannot ban any tags after you've added your first template to your template set")
	}
	_, has = set.bannedTags[name]
	if has {
		return fmt.Errorf("tag '%s' is already banned", name)
	}
	set.bannedTags[name] = true

	return nil
}

// BanFilter bans a specific filter for this template set. See more in the documentation for TemplateSet.
func (set *TemplateSet) BanFilter(name string) error {
	set.initOnce.Do(set.initBuiltins)
	_, has := set.filters[name]
	if !has {
		return fmt.Errorf("filter '%s' not found", name)
	}
	if set.firstTemplateCreated {
		return errors.New("you cannot ban any filters after you've added your first template to your template set")
	}
	_, has = set.bannedFilters[name]
	if has {
		return fmt.Errorf("filter '%s' is already banned", name)
	}
	set.bannedFilters[name] = true

	return nil
}

// RegisterFilter registers a new filter for this template set.
func (set *TemplateSet) RegisterFilter(name string, fn FilterFunction) error {
	set.initOnce.Do(set.initBuiltins)
	_, existing := set.filters[name]
	if existing {
		return fmt.Errorf("filter with name '%s' is already registered", name)
	}
	set.filters[name] = fn
	return nil
}

// RegisterFilter registers a new filter for this template set.
func (set *TemplateSet) SetAutoescape(v bool) {
	set.autoescape = v
}

// ReplaceFilter replaces an already registered filter in this template set.
// Use this function with caution since it allows you to change existing filter behaviour.
func (set *TemplateSet) ReplaceFilter(name string, fn FilterFunction) error {
	set.initOnce.Do(set.initBuiltins)
	_, existing := set.filters[name]
	if !existing {
		return fmt.Errorf("filter with name '%s' does not exist (therefore cannot be overridden)", name)
	}
	set.filters[name] = fn
	return nil
}

// RegisterTag registers a new tag for this template set.
func (set *TemplateSet) RegisterTag(name string, parserFn TagParser) error {
	set.initOnce.Do(set.initBuiltins)
	_, existing := set.tags[name]
	if existing {
		return fmt.Errorf("tag with name '%s' is already registered", name)
	}
	set.tags[name] = &tag{
		name:   name,
		parser: parserFn,
	}
	return nil
}

// ReplaceTag replaces an already registered tag in this template set.
// Use this function with caution since it allows you to change existing tag behaviour.
func (set *TemplateSet) ReplaceTag(name string, parserFn TagParser) error {
	set.initOnce.Do(set.initBuiltins)
	_, existing := set.tags[name]
	if !existing {
		return fmt.Errorf("tag with name '%s' does not exist (therefore cannot be overridden)", name)
	}
	set.tags[name] = &tag{
		name:   name,
		parser: parserFn,
	}
	return nil
}

// FilterExists returns true if the given filter is registered in this template set.
// This checks the set's filter registry, which initially contains copies of all builtin filters
// plus any filters registered via RegisterFilter.
func (set *TemplateSet) FilterExists(name string) bool {
	set.initOnce.Do(set.initBuiltins)
	_, existing := set.filters[name]
	return existing
}

// TagExists returns true if the given tag is registered in this template set.
// This checks the set's tag registry, which initially contains copies of all builtin tags
// plus any tags registered via RegisterTag.
func (set *TemplateSet) TagExists(name string) bool {
	set.initOnce.Do(set.initBuiltins)
	_, existing := set.tags[name]
	return existing
}

// ApplyFilter applies a filter registered in this template set to a given value
// using the given parameters. Returns a *pongo2.Value or an error.
// This is useful for applying set-specific filters, including any custom filters
// registered with RegisterFilter or replaced with ReplaceFilter.
func (set *TemplateSet) ApplyFilter(name string, value *Value, param *Value) (*Value, error) {
	set.initOnce.Do(set.initBuiltins)
	fn, existing := set.filters[name]
	if !existing {
		return nil, &Error{
			Sender:    "applyfilter",
			OrigError: fmt.Errorf("filter with name '%s' not found", name),
		}
	}

	// Make sure param is a *Value
	if param == nil {
		param = AsValue(nil)
	}

	return fn(value, param)
}

// MustApplyFilter behaves like ApplyFilter, but panics on an error.
// This uses the template set's filter registry.
func (set *TemplateSet) MustApplyFilter(name string, value *Value, param *Value) *Value {
	val, err := set.ApplyFilter(name, value, param)
	if err != nil {
		panic(err)
	}
	return val
}

func (set *TemplateSet) resolveTemplate(tpl *Template, path string) (name string, loader TemplateLoader, fd io.Reader, err error) {
	// iterate over loaders until we appear to have a valid template
	for _, loader = range set.loaders {
		name = set.resolveFilenameForLoader(loader, tpl, path)
		fd, err = loader.Get(name)
		if err == nil {
			return
		}
	}

	return path, nil, nil, fmt.Errorf("unable to resolve template")
}

// CleanCache cleans the template cache. If filenames is not empty,
// it will remove the template caches of those filenames.
// Or it will empty the whole template cache. It is thread-safe.
func (set *TemplateSet) CleanCache(filenames ...string) {
	set.templateCacheMutex.Lock()
	defer set.templateCacheMutex.Unlock()

	if len(filenames) == 0 {
		set.templateCache = make(map[string]*Template, len(set.templateCache))
	}

	for _, filename := range filenames {
		delete(set.templateCache, set.resolveFilename(nil, filename))
	}
}

// FromCache is a convenient method to cache templates. It is thread-safe
// and will only compile the template associated with a filename once.
// If TemplateSet.Debug is true (for example during development phase),
// FromCache() will not cache the template and instead recompile it on any
// call (to make changes to a template live instantaneously).
func (set *TemplateSet) FromCache(filename string) (*Template, error) {
	if set.Debug {
		// Recompile on any request
		return set.FromFile(filename)
	}
	// Cache the template
	cleanedFilename := set.resolveFilename(nil, filename)

	set.templateCacheMutex.Lock()
	defer set.templateCacheMutex.Unlock()

	tpl, has := set.templateCache[cleanedFilename]

	// Cache miss
	if !has {
		tpl, err := set.FromFile(cleanedFilename)
		if err != nil {
			return nil, err
		}
		set.templateCache[cleanedFilename] = tpl
		return tpl, nil
	}

	// Cache hit
	return tpl, nil
}

// FromString loads a template from string and returns a Template instance.
func (set *TemplateSet) FromString(tpl string) (*Template, error) {
	set.firstTemplateCreated = true

	return newTemplateString(set, []byte(tpl))
}

// FromBytes loads a template from bytes and returns a Template instance.
func (set *TemplateSet) FromBytes(tpl []byte) (*Template, error) {
	set.firstTemplateCreated = true

	return newTemplateString(set, tpl)
}

// FromFile loads a template from a filename and returns a Template instance.
func (set *TemplateSet) FromFile(filename string) (*Template, error) {
	set.firstTemplateCreated = true

	_, _, fd, err := set.resolveTemplate(nil, filename)
	if err != nil {
		return nil, &Error{
			Filename:  filename,
			Sender:    "fromfile",
			OrigError: err,
		}
	}
	buf, err := io.ReadAll(fd)
	if err != nil {
		return nil, &Error{
			Filename:  filename,
			Sender:    "fromfile",
			OrigError: err,
		}
	}

	return newTemplate(set, filename, false, buf)
}

// RenderTemplateString is a shortcut and renders a template string directly.
func (set *TemplateSet) RenderTemplateString(s string, ctx Context) (string, error) {
	set.firstTemplateCreated = true

	tpl := Must(set.FromString(s))
	result, err := tpl.Execute(ctx)
	if err != nil {
		return "", err
	}
	return result, nil
}

// RenderTemplateBytes is a shortcut and renders template bytes directly.
func (set *TemplateSet) RenderTemplateBytes(b []byte, ctx Context) (string, error) {
	set.firstTemplateCreated = true

	tpl := Must(set.FromBytes(b))
	result, err := tpl.Execute(ctx)
	if err != nil {
		return "", err
	}
	return result, nil
}

// RenderTemplateFile is a shortcut and renders a template file directly.
func (set *TemplateSet) RenderTemplateFile(fn string, ctx Context) (string, error) {
	set.firstTemplateCreated = true

	tpl := Must(set.FromFile(fn))
	result, err := tpl.Execute(ctx)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (set *TemplateSet) logf(format string, args ...any) {
	if set.Debug {
		logger.Printf(fmt.Sprintf("[template set: %s] %s", set.name, format), args...)
	}
}

// Logging function (internally used)
func logf(format string, items ...any) {
	if debug {
		logger.Printf(format, items...)
	}
}

var (
	debug  bool // internal debugging
	logger = log.New(os.Stdout, "[pongo2] ", log.LstdFlags|log.Lshortfile)

	// DefaultLoader allows the default un-sandboxed access to the local file
	// system and is being used by the DefaultSet.
	DefaultLoader = MustNewLocalFileSystemLoader("")

	// DefaultSet is a set created for you for convenience reasons.
	DefaultSet = NewSet("default", DefaultLoader)

	// FromString loads a template from string and returns a Template instance.
	// This is a convenience function that delegates to DefaultSet.FromString.
	FromString = DefaultSet.FromString

	// FromBytes loads a template from bytes and returns a Template instance.
	// This is a convenience function that delegates to DefaultSet.FromBytes.
	FromBytes = DefaultSet.FromBytes

	// FromFile loads a template from a filename and returns a Template instance.
	// This is a convenience function that delegates to DefaultSet.FromFile.
	FromFile = DefaultSet.FromFile

	// FromCache is a convenient method to cache templates. It is thread-safe
	// and will only compile the template associated with a filename once.
	// This is a convenience function that delegates to DefaultSet.FromCache.
	FromCache = DefaultSet.FromCache

	// RenderTemplateString is a shortcut and renders a template string directly.
	// This is a convenience function that delegates to DefaultSet.RenderTemplateString.
	RenderTemplateString = DefaultSet.RenderTemplateString

	// RenderTemplateFile is a shortcut and renders a template file directly.
	// This is a convenience function that delegates to DefaultSet.RenderTemplateFile.
	RenderTemplateFile = DefaultSet.RenderTemplateFile

	// RegisterFilter registers a new filter for the DefaultSet.
	// Returns an error if a filter with the same name already exists.
	RegisterFilter = DefaultSet.RegisterFilter

	// ReplaceFilter replaces an existing filter in the DefaultSet.
	// Use with caution since it changes existing filter behaviour.
	ReplaceFilter = DefaultSet.ReplaceFilter

	// RegisterTag registers a new tag for the DefaultSet.
	// Returns an error if a tag with the same name already exists.
	RegisterTag = DefaultSet.RegisterTag

	// ReplaceTag replaces an existing tag in the DefaultSet.
	// Use with caution since it changes existing tag behaviour.
	ReplaceTag = DefaultSet.ReplaceTag

	// Globals is the global context for the DefaultSet.
	// Variables added here will be available to all templates in the DefaultSet.
	Globals = DefaultSet.Globals

	// SetAutoescape configures the default autoescaping behavior for the DefaultSet.
	// When enabled (true), template output will be automatically HTML-escaped for safety.
	SetAutoescape = DefaultSet.SetAutoescape
)
