package pongo2

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type LocalFilesystemLoader struct {
	baseDir string
}

func MustNewLocalFileSystemLoader(baseDir string) *LocalFilesystemLoader {
	fs, err := NewLocalFileSystemLoader(baseDir)
	if err != nil {
		log.Panic(err)
	}
	return fs
}

func NewLocalFileSystemLoader(baseDir string) (*LocalFilesystemLoader, error) {
	fs := &LocalFilesystemLoader{}
	if baseDir != "" {
		if err := fs.SetBaseDir(baseDir); err != nil {
			return nil, err
		}
	}
	return fs, nil
}

// Use this function to set your template set's base directory. This directory will be used for any relative
// path in filters, tags and From*-functions to determine your template.
func (fs *LocalFilesystemLoader) SetBaseDir(path string) error {
	// Make the path absolute
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		path = abs
	}

	// Check for existence
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("The given path '%s' is not a directory.", path)
	}

	fs.baseDir = path
	return nil
}

func (fs *LocalFilesystemLoader) Get(path string) (io.Reader, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

// Resolves a filename relative to the base directory. Absolute paths are allowed.
// If sandbox restrictions are given (SandboxDirectories), they will be respected and checked.
// On sandbox restriction violation, resolveFilename() panics.

func (fs *LocalFilesystemLoader) Abs(base, name string) string {
	if filepath.IsAbs(name) {
		return name
	}

	// Our own base dir has always priority; if there's none
	// we use the path provided in base.
	var err error
	if fs.baseDir == "" {
		if base == "" {
			base, err = os.Getwd()
			if err != nil {
				panic(err)
			}
			return filepath.Join(base, name)
		}

		return filepath.Join(filepath.Dir(base), name)
	}

	return filepath.Join(fs.baseDir, name)
}

type SandboxedFilesystemLoader struct {
	*LocalFilesystemLoader
}

func NewSandboxedFilesystemLoader(baseDir string) (*SandboxedFilesystemLoader, error) {
	fs, err := NewLocalFileSystemLoader(baseDir)
	if err != nil {
		return nil, err
	}
	return &SandboxedFilesystemLoader{
		LocalFilesystemLoader: fs,
	}, nil
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
