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

func TestMultipleLoaders(t *testing.T) {
	// Create two separate filesystems with different templates
	primaryFS := fstest.MapFS{
		"shared.tpl":  {Data: []byte("primary shared")},
		"primary.tpl": {Data: []byte("primary only")},
	}
	fallbackFS := fstest.MapFS{
		"shared.tpl":   {Data: []byte("fallback shared")},
		"fallback.tpl": {Data: []byte("fallback only")},
	}

	primaryLoader := NewFSLoader(primaryFS)
	fallbackLoader := NewFSLoader(fallbackFS)

	set := NewSet("multi", primaryLoader, fallbackLoader)

	t.Run("primary loader takes precedence", func(t *testing.T) {
		tpl, err := set.FromFile("shared.tpl")
		if err != nil {
			t.Fatalf("FromFile failed: %v", err)
		}
		out, err := tpl.Execute(nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if out != "primary shared" {
			t.Errorf("got %q, want %q", out, "primary shared")
		}
	})

	t.Run("primary only template", func(t *testing.T) {
		tpl, err := set.FromFile("primary.tpl")
		if err != nil {
			t.Fatalf("FromFile failed: %v", err)
		}
		out, err := tpl.Execute(nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if out != "primary only" {
			t.Errorf("got %q, want %q", out, "primary only")
		}
	})

	t.Run("fallback loader used when primary fails", func(t *testing.T) {
		tpl, err := set.FromFile("fallback.tpl")
		if err != nil {
			t.Fatalf("FromFile failed: %v", err)
		}
		out, err := tpl.Execute(nil)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if out != "fallback only" {
			t.Errorf("got %q, want %q", out, "fallback only")
		}
	})

	t.Run("error when template not in any loader", func(t *testing.T) {
		_, err := set.FromFile("nonexistent.tpl")
		if err == nil {
			t.Error("FromFile should fail for template not in any loader")
		}
	})
}

func TestMultipleLoadersWithAddLoader(t *testing.T) {
	primaryFS := fstest.MapFS{
		"base.tpl": {Data: []byte("primary base")},
	}
	addedFS := fstest.MapFS{
		"added.tpl": {Data: []byte("added content")},
	}

	primaryLoader := NewFSLoader(primaryFS)
	set := NewSet("single", primaryLoader)

	// Template should not exist yet
	_, err := set.FromFile("added.tpl")
	if err == nil {
		t.Fatal("FromFile should fail before AddLoader")
	}

	// Add second loader
	addedLoader := NewFSLoader(addedFS)
	set.AddLoader(addedLoader)

	// Now the template should be found
	tpl, err := set.FromFile("added.tpl")
	if err != nil {
		t.Fatalf("FromFile failed after AddLoader: %v", err)
	}
	out, err := tpl.Execute(nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if out != "added content" {
		t.Errorf("got %q, want %q", out, "added content")
	}
}

func TestMultipleLoadersWithIncludes(t *testing.T) {
	// Test that includes work correctly across multiple loaders
	primaryFS := fstest.MapFS{
		"main.tpl": {Data: []byte("Main: {% include \"partial.tpl\" %}")},
	}
	fallbackFS := fstest.MapFS{
		"partial.tpl": {Data: []byte("Partial from fallback")},
	}

	primaryLoader := NewFSLoader(primaryFS)
	fallbackLoader := NewFSLoader(fallbackFS)

	set := NewSet("includes", primaryLoader, fallbackLoader)

	tpl, err := set.FromFile("main.tpl")
	if err != nil {
		t.Fatalf("FromFile failed: %v", err)
	}
	out, err := tpl.Execute(nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	expected := "Main: Partial from fallback"
	if out != expected {
		t.Errorf("got %q, want %q", out, expected)
	}
}
