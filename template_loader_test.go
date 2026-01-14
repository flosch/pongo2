package pongo2

import (
	"io"
	"path/filepath"
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
