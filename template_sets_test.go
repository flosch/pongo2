package pongo2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateSetAddLoader(t *testing.T) {
	loader1 := MustNewLocalFileSystemLoader("")
	set := NewSet("test", loader1)

	loader2 := MustNewLocalFileSystemLoader("")
	set.AddLoader(loader2)

	if len(set.loaders) != 2 {
		t.Errorf("AddLoader: got %d loaders, want 2", len(set.loaders))
	}
}

func TestTemplateSetCleanCache(t *testing.T) {
	tmpDir := t.TempDir()

	tplPath1 := filepath.Join(tmpDir, "file1.tpl")
	tplPath2 := filepath.Join(tmpDir, "file2.tpl")
	if err := os.WriteFile(tplPath1, []byte("template 1"), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := os.WriteFile(tplPath2, []byte("template 2"), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	loader := MustNewLocalFileSystemLoader(tmpDir)
	set := NewSet("test-cache", loader)

	if _, err := set.FromCache("file1.tpl"); err != nil {
		t.Fatalf("FromCache file1 failed: %v", err)
	}
	if _, err := set.FromCache("file2.tpl"); err != nil {
		t.Fatalf("FromCache file2 failed: %v", err)
	}

	set.templateCacheMutex.Lock()
	initialCacheSize := len(set.templateCache)
	set.templateCacheMutex.Unlock()

	if initialCacheSize != 2 {
		t.Errorf("Cache should have 2 entries, got %d", initialCacheSize)
	}

	set.CleanCache("file1.tpl")
	set.templateCacheMutex.Lock()
	cacheAfterClean := len(set.templateCache)
	set.templateCacheMutex.Unlock()

	if cacheAfterClean != 1 {
		t.Errorf("Cache should have 1 entry after cleaning file1, got %d", cacheAfterClean)
	}

	set.CleanCache()
	set.templateCacheMutex.Lock()
	if len(set.templateCache) != 0 {
		t.Error("CleanCache() did not clear all cache")
	}
	set.templateCacheMutex.Unlock()
}

func TestTemplateSetFromCache(t *testing.T) {
	tmpDir := t.TempDir()

	tplPath := filepath.Join(tmpDir, "test.tpl")
	if err := os.WriteFile(tplPath, []byte("cached template"), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	loader := MustNewLocalFileSystemLoader(tmpDir)
	set := NewSet("test-fromcache", loader)

	tpl1, err := set.FromCache("test.tpl")
	if err != nil {
		t.Fatalf("FromCache failed: %v", err)
	}

	tpl2, err := set.FromCache("test.tpl")
	if err != nil {
		t.Fatalf("FromCache second call failed: %v", err)
	}

	if tpl1 != tpl2 {
		t.Error("FromCache should return cached template")
	}

	set.Debug = true
	tpl3, err := set.FromCache("test.tpl")
	if err != nil {
		t.Fatalf("FromCache with debug failed: %v", err)
	}

	if tpl3 == tpl2 {
		t.Error("FromCache with Debug=true should recompile")
	}
}

func TestTemplateSetRenderTemplateString(t *testing.T) {
	loader := MustNewLocalFileSystemLoader("")
	set := NewSet("test-render", loader)

	result, err := set.RenderTemplateString("Hello {{ name }}!", Context{"name": "World"})
	if err != nil {
		t.Fatalf("RenderTemplateString failed: %v", err)
	}

	if result != "Hello World!" {
		t.Errorf("RenderTemplateString = %q, want %q", result, "Hello World!")
	}
}

func TestTemplateSetRenderTemplateBytes(t *testing.T) {
	loader := MustNewLocalFileSystemLoader("")
	set := NewSet("test-render-bytes", loader)

	result, err := set.RenderTemplateBytes([]byte("Hello {{ name }}!"), Context{"name": "World"})
	if err != nil {
		t.Fatalf("RenderTemplateBytes failed: %v", err)
	}

	if result != "Hello World!" {
		t.Errorf("RenderTemplateBytes = %q, want %q", result, "Hello World!")
	}
}

func TestTemplateSetRenderTemplateFile(t *testing.T) {
	tmpDir := t.TempDir()

	tplPath := filepath.Join(tmpDir, "test.tpl")
	if err := os.WriteFile(tplPath, []byte("Hello {{ name }}!"), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	loader := MustNewLocalFileSystemLoader(tmpDir)
	set := NewSet("test-render-file", loader)

	result, err := set.RenderTemplateFile("test.tpl", Context{"name": "World"})
	if err != nil {
		t.Fatalf("RenderTemplateFile failed: %v", err)
	}

	if result != "Hello World!" {
		t.Errorf("RenderTemplateFile = %q, want %q", result, "Hello World!")
	}
}

func TestNewSetPanicWithNoLoaders(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewSet should panic with no loaders")
		}
	}()
	NewSet("test")
}

func TestBanTagAndFilter(t *testing.T) {
	loader := MustNewLocalFileSystemLoader("")
	set := NewSet("test-ban", loader)

	t.Run("ban existing tag", func(t *testing.T) {
		err := set.BanTag("for")
		if err != nil {
			t.Fatalf("BanTag failed: %v", err)
		}
	})

	t.Run("ban non-existent tag", func(t *testing.T) {
		err := set.BanTag("nonexistent")
		if err == nil {
			t.Error("BanTag should fail for non-existent tag")
		}
	})

	t.Run("ban already banned tag", func(t *testing.T) {
		err := set.BanTag("for")
		if err == nil {
			t.Error("BanTag should fail for already banned tag")
		}
	})

	t.Run("ban existing filter", func(t *testing.T) {
		err := set.BanFilter("upper")
		if err != nil {
			t.Fatalf("BanFilter failed: %v", err)
		}
	})

	t.Run("ban non-existent filter", func(t *testing.T) {
		err := set.BanFilter("nonexistent")
		if err == nil {
			t.Error("BanFilter should fail for non-existent filter")
		}
	})

	t.Run("ban already banned filter", func(t *testing.T) {
		err := set.BanFilter("upper")
		if err == nil {
			t.Error("BanFilter should fail for already banned filter")
		}
	})
}

func TestOptions(t *testing.T) {
	loader := MustNewLocalFileSystemLoader("")
	set := NewSet("test-options", loader)

	if set.Options == nil {
		t.Error("Options should not be nil")
	}

	set.Options.TrimBlocks = true
	set.Options.LStripBlocks = true

	tpl, err := set.FromString("{% if true %}\nyes\n{% endif %}")
	if err != nil {
		t.Fatalf("FromString failed: %v", err)
	}

	tpl.Options.TrimBlocks = true
	tpl.Options.LStripBlocks = true

	result, err := tpl.Execute(Context{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if strings.HasPrefix(result, "\n") {
		t.Errorf("TrimBlocks should remove leading newline, got %q", result)
	}
}
