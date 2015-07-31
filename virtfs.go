package pongo2

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalFilesystemLoader struct {
	baseDir string
}

func NewLocalFileSystemLoader(baseDir string) *LocalFilesystemLoader {
	return &LocalFilesystemLoader{
		baseDir: baseDir,
	}
}

// Use this function to set your template set's base directory. This directory will be used for any relative
// path in filters, tags and From*-functions to determine your template.
func (fs *LocalFilesystemLoader) SetBaseDirectory(name string) error {
	// Make the path absolute
	if !filepath.IsAbs(name) {
		abs, err := filepath.Abs(name)
		if err != nil {
			return err
		}
		name = abs
	}

	// Check for existence
	fi, err := os.Stat(name)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("The given path '%s' is not a directory.", name)
	}

	set.baseDirectory = name
	return nil
}

func (fs *LocalFilesystemLoader) Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func (fs *LocalFilesystemLoader) IsDir(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func (fs *LocalFilesystemLoader) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// Resolves a filename relative to the base directory. Absolute paths are allowed.
// If sandbox restrictions are given (SandboxDirectories), they will be respected and checked.
// On sandbox restriction violation, resolveFilename() panics.

func (fs *LocalFilesystemLoader) Abs(relativeTo, name string) string {
	if filepath.IsAbs(name) {
		return name
	}

	// No base directory given
	if fs.baseDir == "" {
		// Since we asked from inside a
		if relativeTo != "" {
			base := filepath.Dir(relativeTo)
			return filepath.Join(base, name)
		}
		return name
	}

	// Base directory given
	return filepath.Join(fs.baseDir, name)

}

type SandboxedFilesystemLoader struct {
	*LocalFilesystemLoader
}

func NewSandboxedFilesystemLoader(baseDir string) *SandboxedFilesystemLoader {
	return &SandboxedFilesystemLoader{
		LocalFilesystemLoader: NewLocalFileSystemLoader(baseDir),
	}
}

// Move sandbox to a virtual fs

/*
if len(set.SandboxDirectories) > 0 {
    defer func() {
        // Remove any ".." or other crap
        resolvedPath = filepath.Clean(resolvedPath)

        // Make the path absolute
        absPath, err := filepath.Abs(resolvedPath)
        if err != nil {
            panic(err)
        }
        resolvedPath = absPath

        // Check against the sandbox directories (once one pattern matches, we're done and can allow it)
        for _, pattern := range set.SandboxDirectories {
            matched, err := filepath.Match(pattern, resolvedPath)
            if err != nil {
                panic("Wrong sandbox directory match pattern (see http://golang.org/pkg/path/filepath/#Match).")
            }
            if matched {
                // OK!
                return
            }
        }

        // No pattern matched, we have to log+deny the request
        set.logf("Access attempt outside of the sandbox directories (blocked): '%s'", resolvedPath)
        resolvedPath = ""
    }()
}
*/
