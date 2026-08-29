package pongo2

import (
	"strings"
	"testing"
)

func TestURLizeEscapesPlainTextAndAnchorTextOnce(t *testing.T) {
	t.Parallel()
	input := `<b> https://x.y/?a=1&b=2`
	got, err := filterUrlizeHelper(input, true, -1)
	if err != nil {
		t.Fatalf("filterUrlizeHelper: %v", err)
	}
	want := `&lt;b&gt; <a href="https://x.y/?a=1&b=2" rel="nofollow">` +
		`https://x.y/?a=1&amp;b=2</a>`
	if got != want {
		t.Fatalf("filterUrlizeHelper(%q) = %q, want %q", input, got, want)
	}
	if strings.Contains(got, "&amp;amp;") {
		t.Fatalf("anchor text was escaped twice: %q", got)
	}
}

func TestURLizeTruncationCountsTheUnescapedURL(t *testing.T) {
	t.Parallel()
	input := `https://example.com/?q="1"&r='2'&s=3`
	got, err := filterUrlizeHelper(input, true, len([]rune(input)))
	if err != nil {
		t.Fatalf("filterUrlizeHelper: %v", err)
	}
	if strings.Contains(got, ellipsis) {
		t.Fatalf("exact-length URL was truncated after escaping: %q", got)
	}
	wantTitle := htmlEscapeReplacer.Replace(input)
	if !strings.Contains(got, `>`+wantTitle+`</a>`) {
		t.Fatalf("anchor title = %q, want escaped %q", got, wantTitle)
	}
}

func TestURLizeMailPassSkipsGeneratedURLAnchors(t *testing.T) {
	t.Parallel()
	input := `https://example.com/bob@example.com then bob@example.com`
	got, err := filterUrlizeHelper(input, true, -1)
	if err != nil {
		t.Fatalf("filterUrlizeHelper: %v", err)
	}
	if strings.Count(got, `<a href=`) != 2 {
		t.Fatalf("anchor count = %d, want URL plus trailing mail: %q",
			strings.Count(got, `<a href=`), got)
	}
	if strings.Contains(got, `href="https://example.com/<a`) ||
		strings.Contains(got, `href="mailto:example.com`) {
		t.Fatalf("mail pass rewrote the generated URL anchor: %q", got)
	}
	wantURL := `<a href="https://example.com/bob@example.com" rel="nofollow">` +
		`https://example.com/bob@example.com</a>`
	if !strings.Contains(got, wantURL) ||
		!strings.Contains(got, `<a href="mailto:bob@example.com">bob@example.com</a>`) {
		t.Fatalf("unexpected URL/mail rendering: %q", got)
	}
}

func TestURLizeCanLeaveTrustedTextUnescaped(t *testing.T) {
	t.Parallel()
	input := `<b>www.example.com</b>`
	got, err := filterUrlizeHelper(input, false, -1)
	if err != nil {
		t.Fatalf("filterUrlizeHelper: %v", err)
	}
	if !strings.HasPrefix(got, "<b>") {
		t.Fatalf("autoescape=false changed trusted prefix: %q", got)
	}
}
