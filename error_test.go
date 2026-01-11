package pongo2

import (
	"errors"
	"os"
	"testing"
)

func TestErrorUnwrap(t *testing.T) {
	origErr := errors.New("original error")
	pErr := &Error{
		Sender:    "test",
		OrigError: origErr,
	}

	unwrapped := pErr.Unwrap()
	if unwrapped != origErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, origErr)
	}

	if !errors.Is(pErr, origErr) {
		t.Error("errors.Is should return true for the original error")
	}
}

func TestErrorRawLine(t *testing.T) {
	t.Run("line <= 0", func(t *testing.T) {
		e := &Error{Line: 0}
		line, available, err := e.RawLine()
		if available || err != nil || line != "" {
			t.Error("RawLine should return empty for Line <= 0")
		}
	})

	t.Run("filename is <string>", func(t *testing.T) {
		e := &Error{Line: 1, Filename: "<string>"}
		line, available, err := e.RawLine()
		if available || err != nil || line != "" {
			t.Error("RawLine should return empty for <string> filename")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		e := &Error{Line: 1, Filename: "/nonexistent/file.tpl"}
		_, _, err := e.RawLine()
		if err == nil {
			t.Error("RawLine should return error for non-existent file")
		}
	})

	t.Run("valid file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.tpl")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		content := "line 1\nline 2\nline 3"
		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		e := &Error{Line: 2, Filename: tmpFile.Name()}
		line, available, err := e.RawLine()
		if err != nil {
			t.Fatalf("RawLine returned error: %v", err)
		}
		if !available {
			t.Error("RawLine should return available=true for valid file")
		}
		if line != "line 2" {
			t.Errorf("RawLine = %q, want %q", line, "line 2")
		}
	})

	t.Run("line exceeds file length", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test*.tpl")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString("line 1\nline 2"); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		tmpFile.Close()

		e := &Error{Line: 100, Filename: tmpFile.Name()}
		_, available, err := e.RawLine()
		if err != nil {
			t.Fatalf("RawLine returned error: %v", err)
		}
		if available {
			t.Error("RawLine should return available=false when line exceeds file length")
		}
	})
}
