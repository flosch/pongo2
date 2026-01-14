package pongo2

import (
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestFSLoader(t *testing.T) {
	testFS := fstest.MapFS{
		"templates/base.tpl":  {Data: []byte("base content")},
		"templates/child.tpl": {Data: []byte("child content")},
	}

	loader := NewFSLoader(testFS)

	t.Run("Abs", func(t *testing.T) {
		absPath := loader.Abs("templates/base.tpl", "child.tpl")
		expected := filepath.Join("templates", "child.tpl")
		if absPath != expected {
			t.Errorf("Abs = %q, want %q", absPath, expected)
		}
	})

	t.Run("Get existing file", func(t *testing.T) {
		reader, err := loader.Get("templates/base.tpl")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}

		if string(content) != "base content" {
			t.Errorf("Get content = %q, want %q", string(content), "base content")
		}
	})

	t.Run("Get non-existent file", func(t *testing.T) {
		_, err := loader.Get("nonexistent.tpl")
		if err == nil {
			t.Error("Get should fail for non-existent file")
		}
	})
}

type mockHTTPFileSystem struct {
	files map[string]string
}

func (m *mockHTTPFileSystem) Open(name string) (http.File, error) {
	content, ok := m.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return &mockHTTPFile{Reader: strings.NewReader(content), name: name}, nil
}

type mockHTTPFile struct {
	*strings.Reader
	name string
}

func (f *mockHTTPFile) Close() error                       { return nil }
func (f *mockHTTPFile) Readdir(int) ([]fs.FileInfo, error) { return nil, nil }
func (f *mockHTTPFile) Stat() (fs.FileInfo, error)         { return nil, nil }

func TestNewHttpFileSystemLoader(t *testing.T) {
	t.Run("nil filesystem", func(t *testing.T) {
		_, err := NewHttpFileSystemLoader(nil, "")
		if err == nil {
			t.Error("NewHttpFileSystemLoader should fail with nil filesystem")
		}
	})

	t.Run("valid filesystem", func(t *testing.T) {
		mockFS := &mockHTTPFileSystem{
			files: map[string]string{
				"/templates/test.tpl": "test content",
			},
		}

		loader, err := NewHttpFileSystemLoader(mockFS, "")
		if err != nil {
			t.Fatalf("NewHttpFileSystemLoader failed: %v", err)
		}

		absPath := loader.Abs("/some/path", "test.tpl")
		if absPath != "test.tpl" {
			t.Errorf("Abs = %q, want %q", absPath, "test.tpl")
		}

		reader, err := loader.Get("/templates/test.tpl")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}

		if string(content) != "test content" {
			t.Errorf("Get content = %q, want %q", string(content), "test content")
		}
	})

	t.Run("with base directory", func(t *testing.T) {
		mockFS := &mockHTTPFileSystem{
			files: map[string]string{
				"/templates/test.tpl": "test content",
			},
		}

		loader, err := NewHttpFileSystemLoader(mockFS, "/templates")
		if err != nil {
			t.Fatalf("NewHttpFileSystemLoader with baseDir failed: %v", err)
		}

		reader, err := loader.Get("test.tpl")
		if err != nil {
			t.Fatalf("Get with baseDir failed: %v", err)
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}

		if string(content) != "test content" {
			t.Errorf("Get with baseDir content = %q, want %q", string(content), "test content")
		}
	})
}

func TestMustNewHttpFileSystemLoader(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockFS := &mockHTTPFileSystem{files: map[string]string{}}
		loader := MustNewHttpFileSystemLoader(mockFS, "")
		if loader == nil {
			t.Error("MustNewHttpFileSystemLoader returned nil")
		}
	})

	t.Run("panic on nil filesystem", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNewHttpFileSystemLoader should panic with nil filesystem")
			}
		}()
		MustNewHttpFileSystemLoader(nil, "")
	})
}
