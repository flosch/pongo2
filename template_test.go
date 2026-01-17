package pongo2

import (
	"bytes"
	"strings"
	"testing"
)

func TestTemplateExecuteWriter(t *testing.T) {
	tpl, err := FromString("Hello {{ name }}!")
	if err != nil {
		t.Fatalf("FromString failed: %v", err)
	}

	var buf bytes.Buffer
	err = tpl.ExecuteWriter(Context{"name": "World"}, &buf)
	if err != nil {
		t.Fatalf("ExecuteWriter failed: %v", err)
	}

	if buf.String() != "Hello World!" {
		t.Errorf("ExecuteWriter result = %q, want %q", buf.String(), "Hello World!")
	}
}

func TestTemplateExecuteWriterUnbuffered(t *testing.T) {
	tpl, err := FromString("Hello {{ name }}!")
	if err != nil {
		t.Fatalf("FromString failed: %v", err)
	}

	var buf bytes.Buffer
	err = tpl.ExecuteWriterUnbuffered(Context{"name": "World"}, &buf)
	if err != nil {
		t.Fatalf("ExecuteWriterUnbuffered failed: %v", err)
	}

	if buf.String() != "Hello World!" {
		t.Errorf("ExecuteWriterUnbuffered result = %q, want %q", buf.String(), "Hello World!")
	}
}

func TestMust(t *testing.T) {
	t.Run("successful Must", func(t *testing.T) {
		tpl := Must(FromString("test"))
		if tpl == nil {
			t.Error("Must should return template on success")
		}
	})

	t.Run("Must with error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Must should panic on error")
			}
		}()
		Must(FromString("{% invalid %}"))
	})
}

func TestTokenString(t *testing.T) {
	token := &Token{
		Typ:      TokenHTML,
		Val:      "test",
		Line:     1,
		Col:      5,
		Filename: "test.tpl",
	}

	str := token.String()
	if !strings.Contains(str, "test") {
		t.Errorf("Token.String() should contain value, got %q", str)
	}
}
