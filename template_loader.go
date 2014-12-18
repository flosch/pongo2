package pongo2

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// The TemplateLoader allows pongo2 to look up templates.
type TemplateLoader interface {
	// TODO: Opens a file ....
	OpenFile(name string) (io.ReadCloser, error)
	
	// TODO: Returns the canonical path to ...
	AbsPath(tpl *Template, filename string) string
}

type filesystemLoader struct {
	base      string
	sandboxed []string
}

// Creates a new filesystem template loader.
//
// If you set the base directory (string is non-empty), all filename lookups in tags/filters are
// relative to this directory. If it's empty, all lookups are relative to the current filename which is importing.
//
// You can limit file accesses (for all tags/filters which are using pongo2's file resolver technique)
// to these sandbox directories. All default pongo2 filters/tags are respecting these restrictions.
// For example, if you only have your base directory in the list, a {% ssi "/etc/passwd" %} will not work.
// No items in SandboxDirectories means no restrictions at all.
//
// SandboxDirectories can be changed at runtime. Please synchronize the access to it if you need to change it
// after you've added your first template to the set. You *must* use this match pattern for your directories:
// http://golang.org/pkg/path/filepath/#Match
func NewFilesystemLoader(base string, sandboxed []string) *filesystemLoader {
	// Make the path absolute
	if !filepath.IsAbs(base) {
		abs, err := filepath.Abs(base)
		if err != nil {
			panic(fmt.Errorf("pongo2: %s", err))
		}
		base = abs
	}

	// Check for existence
	fi, err := os.Stat(base)
	if err != nil {
		panic(fmt.Errorf("pongo2: %s", err))
	}
	if !fi.IsDir() {
		panic(fmt.Errorf("The given path '%s' is not a directory."))
	}

	return &filesystemLoader{
		base: base,
	}
}

func (fs *filesystemLoader) OpenFile(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// Resolves a filename relative to the base directory. Absolute paths are allowed.
// If sandbox restrictions are given (SandboxDirectories), they will be respected and checked.
// On sandbox restriction violation, resolveFilename() panics.
func (fs *filesystemLoader) AbsPath(tpl *Template, filename string) (resolved_path string) {
	if len(fs.sandboxed) > 0 {
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
			for _, pattern := range fs.sandboxed {
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
			// set.logf("Access attempt outside of the sandbox directories (blocked): '%s'", resolved_path)
			resolved_path = ""
		}()
	}

	if filepath.IsAbs(filename) {
		return filename
	}

	if fs.base == "" {
		if tpl != nil {
			if tpl.is_tpl_string {
				return filename
			}
			base := filepath.Dir(tpl.name)
			return filepath.Join(base, filename)
		}
		return filename
	} else {
		return filepath.Join(fs.base, filename)
	}
}
