package pongo2

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// FuzzFilterEscapejs fuzzes the escapejs filter for JavaScript escaping security.
func FuzzFilterEscapejs(f *testing.F) {
	// Basic strings
	f.Add("")
	f.Add("hello")
	f.Add("Hello World")
	f.Add("hello world")

	// JavaScript-sensitive characters
	f.Add("'")
	f.Add("\"")
	f.Add("\\")
	f.Add("/")
	f.Add("<")
	f.Add(">")
	f.Add("&")
	f.Add("=")
	f.Add("`")
	f.Add("$")

	// Escape sequences
	f.Add("\n")
	f.Add("\r")
	f.Add("\t")
	f.Add("\\n")
	f.Add("\\r")
	f.Add("\\t")
	f.Add("\\\n")
	f.Add("\\\\n")
	f.Add("\n\r\t")
	f.Add("\\n\\r\\t")
	f.Add("\\\\\n")
	f.Add("line1\nline2")
	f.Add("line1\\nline2")
	f.Add("col1\tcol2")
	f.Add("return\rhere")

	// XSS attack vectors
	f.Add("<script>alert('xss')</script>")
	f.Add(`<script>alert("xss")</script>`)
	f.Add("</script><script>alert('xss')</script>")
	f.Add("';alert('xss');//")
	f.Add(`";alert("xss");//`)
	f.Add("javascript:alert('xss')")
	f.Add("onerror=alert('xss')")
	f.Add("<img src=x onerror=alert('xss')>")
	f.Add("<svg/onload=alert('xss')>")

	// Template injection attempts
	f.Add("${alert('xss')}")
	f.Add("{{alert('xss')}}")
	f.Add("{{constructor.constructor('alert(1)')()}}")

	// Unicode
	f.Add("你好世界")
	f.Add("こんにちは")
	f.Add("Привіт світ")
	f.Add("🎉🎊🎁")
	f.Add("\u0000")
	f.Add("\u001F")
	f.Add("\u007F")
	f.Add("\u0080")
	f.Add("\u00FF")
	f.Add("\u0100")
	f.Add("\u2028")                         // Line separator
	f.Add("\u2029")                         // Paragraph separator
	f.Add("\uFEFF")                         // BOM
	f.Add(string([]byte{0xED, 0xA0, 0x80})) // High surrogate (invalid UTF-8)
	f.Add(string([]byte{0xED, 0xBF, 0xBF})) // Low surrogate (invalid UTF-8)
	f.Add("\uFFFD")                         // Replacement char

	// Control characters
	f.Add("\x00")
	f.Add("\x01")
	f.Add("\x1F")
	f.Add("\x7F")
	f.Add("\x00\x01\x02\x03\x04\x05")
	f.Add("hello\x00world")

	// Mixed content
	f.Add("hello'world")
	f.Add(`hello"world`)
	f.Add("hello\\world")
	f.Add("hello\nworld")
	f.Add("hello'\"\\world")
	f.Add("It's a \"test\"")
	f.Add(`She said "hello"`)
	f.Add("path\\to\\file")
	f.Add("C:\\Users\\test")
	f.Add("\\\\server\\share")

	// Edge cases
	f.Add("\\")
	f.Add("\\\\")
	f.Add("\\\\\\")
	f.Add("'")
	f.Add("''")
	f.Add("'''")
	f.Add("\"")
	f.Add("\"\"")
	f.Add("\"\"\"")
	f.Add("\\'\"\\/")

	// Long strings
	f.Add(strings.Repeat("a", 1000))
	f.Add(strings.Repeat("'", 100))
	f.Add(strings.Repeat("\\", 100))
	f.Add(strings.Repeat("\\n", 100))
	f.Add(strings.Repeat("\n", 100))

	// Invalid UTF-8
	f.Add(string([]byte{0x80}))
	f.Add(string([]byte{0xFF}))
	f.Add(string([]byte{0xC0, 0x80}))
	f.Add(string([]byte{0xED, 0xA0, 0x80}))

	f.Fuzz(func(t *testing.T, input string) {
		result, err := filterEscapejs(AsValue(input), nil)
		if err != nil {
			t.Fatalf("escapejs returned error: %v", err)
		}

		output := result.String()

		// Verify basic safety: no unescaped dangerous characters
		// (except alphanumeric, space, and /)
		for i := 0; i < len(output); i++ {
			c := output[i]
			if c == '\\' && i+1 < len(output) {
				// Skip the next char (it's escaped)
				i++
				continue
			}
			// These should be escaped in JS strings
			if c == '\n' || c == '\r' || c == '\'' || c == '"' {
				t.Errorf("escapejs output contains unescaped dangerous char %q at position %d: input=%q, output=%q",
					c, i, input, output)
			}
		}
	})
}

// FuzzFilterUrlize fuzzes the urlize filter for URL/email detection.
func FuzzFilterUrlize(f *testing.F) {
	// Basic URLs
	f.Add("http://example.com")
	f.Add("https://example.com")
	f.Add("http://www.example.com")
	f.Add("https://www.example.com")
	f.Add("www.example.com")
	f.Add("example.com")

	// URLs with paths
	f.Add("http://example.com/path")
	f.Add("http://example.com/path/to/page")
	f.Add("http://example.com/path?query=1")
	f.Add("http://example.com/path?a=1&b=2")
	f.Add("http://example.com/path#anchor")
	f.Add("http://example.com/path?q=1#anchor")

	// URLs with ports
	f.Add("http://example.com:8080")
	f.Add("http://example.com:8080/path")
	f.Add("http://localhost:3000")

	// URLs with authentication
	f.Add("http://user@example.com")
	f.Add("http://user:pass@example.com")

	// Email addresses
	f.Add("user@example.com")
	f.Add("user.name@example.com")
	f.Add("user+tag@example.com")
	f.Add("user@subdomain.example.com")
	f.Add("user@example.co.uk")

	// Various TLDs
	f.Add("example.org")
	f.Add("example.net")
	f.Add("example.io")
	f.Add("example.dev")
	f.Add("example.ai")
	f.Add("example.app")
	f.Add("example.co")
	f.Add("example.me")
	f.Add("example.tv")
	f.Add("example.info")
	f.Add("example.biz")
	f.Add("example.edu")
	f.Add("example.gov")

	// Country code TLDs
	f.Add("example.uk")
	f.Add("example.de")
	f.Add("example.fr")
	f.Add("example.jp")
	f.Add("example.cn")
	f.Add("example.ru")
	f.Add("example.br")
	f.Add("example.au")

	// Mixed content
	f.Add("Visit http://example.com today!")
	f.Add("Check out www.example.com for more.")
	f.Add("Contact user@example.com for help.")
	f.Add("See http://a.com and http://b.com")
	f.Add("Email me at a@b.com or c@d.com")
	f.Add("Visit example.com or example.org")

	// Edge cases
	f.Add("")
	f.Add(" ")
	f.Add("   ")
	f.Add("no urls here")
	f.Add("almost.notld")
	f.Add("http://")
	f.Add("https://")
	f.Add("www.")
	f.Add("@")
	f.Add("user@")
	f.Add("@example.com")

	// Unicode domains
	f.Add("http://例え.jp")
	f.Add("http://münchen.de")
	f.Add("http://україна.укр")

	// URLs with spaces (should not match)
	f.Add("http://example .com")
	f.Add("example .com")

	// Special characters
	f.Add("http://example.com/path with spaces")
	f.Add("http://example.com/path%20encoded")
	f.Add("http://example.com/<script>")
	f.Add("javascript:alert('xss')")

	// Long inputs
	f.Add(strings.Repeat("http://example.com ", 50))
	f.Add(strings.Repeat("user@example.com ", 50))
	f.Add("http://" + strings.Repeat("a", 100) + ".com")

	// Malformed
	f.Add("http://example..com")
	f.Add("http://.example.com")
	f.Add("http://example.com.")
	f.Add("user@@example.com")
	f.Add("user@example@com")
	f.Add("...")
	f.Add("http:///")
	f.Add("http://./")

	f.Fuzz(func(t *testing.T, input string) {
		// Test urlize with default autoescape
		result, err := filterUrlize(AsValue(input), AsValue(true))
		if err != nil {
			// Some inputs may cause errors (very long inputs, etc.)
			return
		}
		_ = result.String()

		// Test urlize without autoescape
		result2, err := filterUrlize(AsValue(input), AsValue(false))
		if err != nil {
			return
		}
		_ = result2.String()
	})
}

// FuzzFilterSlice fuzzes the slice filter for Python-style slicing.
func FuzzFilterSlice(f *testing.F) {
	// Input strings with various slice parameters
	f.Add("hello", ":")
	f.Add("hello", "0:")
	f.Add("hello", ":5")
	f.Add("hello", "0:5")
	f.Add("hello", "1:4")
	f.Add("hello", "1:")
	f.Add("hello", ":4")
	f.Add("hello", "0:0")
	f.Add("hello", "5:5")

	// Negative indices
	f.Add("hello world", "-1:")
	f.Add("hello world", ":-1")
	f.Add("hello world", "-5:")
	f.Add("hello world", ":-5")
	f.Add("hello world", "-5:-1")
	f.Add("hello world", "-1:-5")
	f.Add("hello world", "-100:")
	f.Add("hello world", ":-100")
	f.Add("hello world", "-100:-100")

	// Out of bounds
	f.Add("hello", "0:100")
	f.Add("hello", "100:")
	f.Add("hello", ":100")
	f.Add("hello", "100:200")
	f.Add("hello", "-100:100")

	// Edge cases
	f.Add("", ":")
	f.Add("", "0:")
	f.Add("", ":0")
	f.Add("", "0:0")
	f.Add("a", ":")
	f.Add("a", "0:")
	f.Add("a", ":1")
	f.Add("a", "0:1")
	f.Add("a", "-1:")
	f.Add("a", ":-1")

	// Invalid formats
	f.Add("hello", "")
	f.Add("hello", "1")
	f.Add("hello", "1:2:3")
	f.Add("hello", "a:b")
	f.Add("hello", "::")
	f.Add("hello", ":::")
	f.Add("hello", " : ")
	f.Add("hello", "1 : 2")

	// Unicode strings
	f.Add("你好世界", ":")
	f.Add("你好世界", "0:2")
	f.Add("你好世界", "2:")
	f.Add("你好世界", "-2:")
	f.Add("你好世界", ":-2")
	f.Add("こんにちは", "1:4")
	f.Add("🎉🎊🎁", "0:2")
	f.Add("🎉🎊🎁", "-1:")

	// Long strings
	f.Add(strings.Repeat("a", 1000), "0:100")
	f.Add(strings.Repeat("a", 1000), "900:")
	f.Add(strings.Repeat("a", 1000), "-100:")
	f.Add(strings.Repeat("a", 1000), ":-100")

	// Whitespace
	f.Add("hello", " 0 : 5 ")
	f.Add("hello", "\t0\t:\t5\t")

	// Special numeric values
	f.Add("hello", "0:9999999999")
	f.Add("hello", "-9999999999:")

	f.Fuzz(func(t *testing.T, input, sliceParam string) {
		result, err := filterSlice(AsValue(input), AsValue(sliceParam))
		if err != nil {
			// Errors are expected for invalid slice formats
			return
		}

		// Result should never be longer than input
		if result.Len() > AsValue(input).Len() {
			t.Errorf("slice result longer than input: input=%q, param=%q, result=%q",
				input, sliceParam, result.String())
		}
	})
}

// FuzzFilterFloatformat fuzzes the floatformat filter for number formatting.
func FuzzFilterFloatformat(f *testing.F) {
	// Basic floats with various decimal places
	f.Add("3.14159", "0")
	f.Add("3.14159", "1")
	f.Add("3.14159", "2")
	f.Add("3.14159", "3")
	f.Add("3.14159", "4")
	f.Add("3.14159", "5")
	f.Add("3.14159", "10")

	// Negative decimal places (trim mode)
	f.Add("3.14159", "-1")
	f.Add("3.14159", "-2")
	f.Add("3.14159", "-3")
	f.Add("3.0", "-1")
	f.Add("3.0", "-2")

	// Zero and near-zero
	f.Add("0", "2")
	f.Add("0.0", "2")
	f.Add("0.00", "2")
	f.Add("-0", "2")
	f.Add("-0.0", "2")
	f.Add("0.001", "2")
	f.Add("0.001", "3")
	f.Add("0.001", "4")

	// Integers
	f.Add("42", "0")
	f.Add("42", "2")
	f.Add("-42", "2")
	f.Add("100", "2")
	f.Add("100", "-2")

	// Large numbers
	f.Add("1234567890", "2")
	f.Add("9999999999", "2")
	f.Add("12345678901234567890", "2")

	// Very small numbers
	f.Add("0.000000001", "10")
	f.Add("0.000000001", "5")
	f.Add("1e-10", "15")
	f.Add("1e-10", "5")

	// Scientific notation
	f.Add("1e10", "2")
	f.Add("1.5e10", "2")
	f.Add("-1e10", "2")
	f.Add("1e-5", "10")
	f.Add("1.23e5", "2")
	f.Add("1.23E5", "2")

	// Special float values
	f.Add("inf", "2")
	f.Add("-inf", "2")
	f.Add("nan", "2")
	f.Add("Inf", "2")
	f.Add("-Inf", "2")
	f.Add("NaN", "2")

	// Non-numeric strings
	f.Add("hello", "2")
	f.Add("", "2")
	f.Add("   ", "2")
	f.Add("3.14abc", "2")
	f.Add("abc3.14", "2")

	// Edge case parameters
	f.Add("3.14159", "")
	f.Add("3.14159", "abc")
	f.Add("3.14159", "999")
	f.Add("3.14159", "1000")
	f.Add("3.14159", "1001")
	f.Add("3.14159", "-999")
	f.Add("3.14159", "-1000")

	// Rounding edge cases
	f.Add("1.5", "0")
	f.Add("2.5", "0")
	f.Add("1.25", "1")
	f.Add("1.35", "1")
	f.Add("1.45", "1")
	f.Add("1.55", "1")
	f.Add("0.5", "0")
	f.Add("0.05", "1")
	f.Add("0.005", "2")

	// Precision edge cases
	f.Add("0.1", "20")
	f.Add("0.2", "20")
	f.Add("0.3", "20")
	f.Add("0.1", "50")

	f.Fuzz(func(t *testing.T, input, decimals string) {
		result, err := filterFloatformat(AsValue(input), AsValue(decimals))
		if err != nil {
			// Errors are expected for very large decimal values
			return
		}
		_ = result.String()
	})
}

// FuzzFilterTruncate fuzzes the truncatechars and truncatewords filters.
func FuzzFilterTruncate(f *testing.F) {
	// Basic truncation
	f.Add("Hello World", "5")
	f.Add("Hello World", "10")
	f.Add("Hello World", "15")
	f.Add("Hello World", "20")

	// Edge cases
	f.Add("Hello World", "0")
	f.Add("Hello World", "1")
	f.Add("Hello World", "2")
	f.Add("Hello World", "-1")
	f.Add("Hello World", "-10")
	f.Add("", "5")
	f.Add("", "0")
	f.Add("a", "1")
	f.Add("a", "0")
	f.Add("a", "2")

	// Long strings
	f.Add(strings.Repeat("hello ", 100), "10")
	f.Add(strings.Repeat("hello ", 100), "50")
	f.Add(strings.Repeat("hello ", 100), "100")
	f.Add(strings.Repeat("a", 1000), "100")
	f.Add(strings.Repeat("a", 1000), "10")

	// Unicode
	f.Add("你好世界测试", "3")
	f.Add("你好世界测试", "5")
	f.Add("こんにちは世界", "4")
	f.Add("🎉🎊🎁🎀🎁", "3")
	f.Add("Ελληνικά", "4")
	f.Add("Привіт світ", "5")

	// Mixed content
	f.Add("Hello 你好 World", "7")
	f.Add("Test 🎉 emoji", "6")
	f.Add("Mixed 日本語 content", "10")

	// HTML content (for truncatechars_html)
	f.Add("<p>Hello World</p>", "5")
	f.Add("<p><b>Bold</b> text</p>", "5")
	f.Add("<div><p>Nested</p></div>", "5")
	f.Add("<a href='test'>Link</a>", "3")
	f.Add("<script>alert(1)</script>", "5")
	f.Add("<p>你好<b>世界</b></p>", "3")

	// Malformed HTML
	f.Add("<p>Unclosed", "3")
	f.Add("No closing</p>", "5")
	f.Add("<>Empty tags</>", "5")
	f.Add("<<nested>>", "3")

	// Special parameters
	f.Add("Hello", "")
	f.Add("Hello", "abc")
	f.Add("Hello", "1.5")
	f.Add("Hello", "99999")

	// Words truncation
	f.Add("one two three four five", "1")
	f.Add("one two three four five", "2")
	f.Add("one two three four five", "3")
	f.Add("one two three four five", "4")
	f.Add("one two three four five", "5")
	f.Add("one two three four five", "10")
	f.Add("single", "1")
	f.Add("   spaced   words   ", "2")
	f.Add("multiple\nlines\nhere", "2")

	f.Fuzz(func(t *testing.T, input, length string) {
		// Test truncatechars
		result1, err1 := filterTruncatechars(AsValue(input), AsValue(length))
		if err1 == nil {
			out := result1.String()
			// Verify output length is reasonable
			inputLen := len([]rune(input))
			paramLen := AsValue(length).Integer()
			if paramLen > 0 && len([]rune(out)) > paramLen && len([]rune(out)) > inputLen {
				t.Errorf("truncatechars output too long: input=%q, param=%q, output=%q",
					input, length, out)
			}
		}

		// Test truncatewords
		result2, err2 := filterTruncatewords(AsValue(input), AsValue(length))
		if err2 == nil {
			_ = result2.String()
		}

		// Test truncatechars_html
		result3, err3 := filterTruncatecharsHTML(AsValue(input), AsValue(length))
		if err3 == nil {
			_ = result3.String()
		}

		// Test truncatewords_html
		result4, err4 := filterTruncatewordsHTML(AsValue(input), AsValue(length))
		if err4 == nil {
			_ = result4.String()
		}
	})
}

// FuzzFilterAdd fuzzes the add filter for number and string addition.
func FuzzFilterAdd(f *testing.F) {
	// Integer addition
	f.Add("0", "0")
	f.Add("1", "1")
	f.Add("-1", "1")
	f.Add("1", "-1")
	f.Add("-1", "-1")
	f.Add("100", "200")
	f.Add("2147483647", "1")
	f.Add("-2147483648", "-1")

	// Float addition
	f.Add("1.5", "2.5")
	f.Add("0.1", "0.2")
	f.Add("-1.5", "1.5")
	f.Add("3.14159", "2.71828")
	f.Add("1e10", "1")
	f.Add("1e-10", "1")

	// String concatenation
	f.Add("hello", "world")
	f.Add("", "test")
	f.Add("test", "")
	f.Add("", "")
	f.Add("a", "b")
	f.Add("你好", "世界")
	f.Add("🎉", "🎊")

	// Mixed types
	f.Add("123", "abc")
	f.Add("abc", "123")
	f.Add("1.5", "abc")
	f.Add("abc", "1.5")

	// Edge cases
	f.Add("nan", "1")
	f.Add("inf", "1")
	f.Add("-inf", "1")
	f.Add("   ", "test")
	f.Add("test", "   ")

	// Large numbers
	f.Add("99999999999999999999", "1")
	f.Add("1", "99999999999999999999")

	f.Fuzz(func(t *testing.T, value, param string) {
		result, err := filterAdd(AsValue(value), AsValue(param))
		if err != nil {
			t.Fatalf("add filter error: %v", err)
		}
		_ = result.String()
	})
}

// FuzzFilterPluralize fuzzes the pluralize filter.
func FuzzFilterPluralize(f *testing.F) {
	// Basic pluralization
	f.Add("0", "")
	f.Add("1", "")
	f.Add("2", "")
	f.Add("-1", "")
	f.Add("100", "")

	// With suffix
	f.Add("0", "s")
	f.Add("1", "s")
	f.Add("2", "s")
	f.Add("0", "es")
	f.Add("1", "es")
	f.Add("2", "es")

	// With singular,plural format
	f.Add("0", "y,ies")
	f.Add("1", "y,ies")
	f.Add("2", "y,ies")
	f.Add("0", "child,children")
	f.Add("1", "child,children")
	f.Add("2", "child,children")

	// Edge cases
	f.Add("1.5", "s")
	f.Add("0.5", "s")
	f.Add("-1.5", "s")
	f.Add("abc", "s")
	f.Add("", "s")
	f.Add("1", "")
	f.Add("1", ",")
	f.Add("1", ",,")
	f.Add("1", ",,,")
	f.Add("1", "a,b,c,d")

	// Unicode
	f.Add("1", "猫,猫们")
	f.Add("2", "猫,猫们")

	f.Fuzz(func(t *testing.T, count, suffix string) {
		result, err := filterPluralize(AsValue(count), AsValue(suffix))
		if err != nil {
			// Errors expected for invalid suffix formats
			return
		}
		_ = result.String()
	})
}

// FuzzFilterYesno fuzzes the yesno filter.
func FuzzFilterYesno(f *testing.F) {
	// Boolean values
	f.Add("true", "")
	f.Add("false", "")
	f.Add("True", "")
	f.Add("False", "")

	// With custom values
	f.Add("true", "yes,no")
	f.Add("false", "yes,no")
	f.Add("true", "yes,no,maybe")
	f.Add("false", "yes,no,maybe")
	f.Add("", "yes,no,maybe")

	// Truthy/falsy values
	f.Add("1", "yes,no")
	f.Add("0", "yes,no")
	f.Add("hello", "yes,no")
	f.Add("", "yes,no")
	f.Add("   ", "yes,no")

	// Edge cases
	f.Add("true", ",")
	f.Add("true", "a,")
	f.Add("true", ",b")
	f.Add("true", "a,b,c,d")
	f.Add("true", "x")
	f.Add("true", "")

	// Unicode
	f.Add("true", "是,否")
	f.Add("false", "是,否")
	f.Add("true", "はい,いいえ,たぶん")

	f.Fuzz(func(t *testing.T, value, options string) {
		result, err := filterYesno(AsValue(value), AsValue(options))
		if err != nil {
			// Errors expected for invalid option formats
			return
		}
		_ = result.String()
	})
}

// FuzzFilterStringformat fuzzes the stringformat filter.
func FuzzFilterStringformat(f *testing.F) {
	// Integer formats
	f.Add("42", "%d")
	f.Add("42", "%5d")
	f.Add("42", "%-5d")
	f.Add("42", "%05d")
	f.Add("-42", "%d")
	f.Add("0", "%d")

	// Float formats
	f.Add("3.14159", "%f")
	f.Add("3.14159", "%.2f")
	f.Add("3.14159", "%10.2f")
	f.Add("3.14159", "%-10.2f")
	f.Add("3.14159", "%e")
	f.Add("3.14159", "%E")
	f.Add("3.14159", "%g")
	f.Add("3.14159", "%G")

	// String formats
	f.Add("hello", "%s")
	f.Add("hello", "%10s")
	f.Add("hello", "%-10s")
	f.Add("hello", "%q")
	f.Add("hello", "%.3s")

	// Hex/octal/binary
	f.Add("255", "%x")
	f.Add("255", "%X")
	f.Add("255", "%o")
	f.Add("255", "%b")
	f.Add("255", "%#x")
	f.Add("255", "%#X")
	f.Add("255", "%#o")
	f.Add("255", "%#b")

	// Character
	f.Add("65", "%c")
	f.Add("97", "%c")

	// Percent
	f.Add("test", "%%")
	f.Add("test", "100%%")

	// Unicode
	f.Add("你好", "%s")
	f.Add("你好", "%q")
	f.Add("🎉", "%s")

	// Edge cases
	f.Add("", "%s")
	f.Add("test", "")
	f.Add("test", "no format")
	f.Add("test", "%")
	f.Add("test", "%%")
	f.Add("test", "%%%")
	f.Add("test", "%z")
	f.Add("test", "%100s")
	f.Add("test", "%.100s")

	// Complex formats
	f.Add("42", "Value: %d")
	f.Add("3.14", "Pi is approximately %f")
	f.Add("hello", "The word is: %s")

	f.Fuzz(func(t *testing.T, value, format string) {
		result, err := filterStringformat(AsValue(value), AsValue(format))
		if err != nil {
			return
		}
		_ = result.String()
	})
}

// FuzzFilterWordwrap fuzzes the wordwrap filter.
func FuzzFilterWordwrap(f *testing.F) {
	// Basic wrapping
	f.Add("one two three four five", "2")
	f.Add("one two three four five", "3")
	f.Add("one two three four five", "5")
	f.Add("one two three four five", "10")

	// Edge cases
	f.Add("", "5")
	f.Add("single", "5")
	f.Add("one two", "0")
	f.Add("one two", "-1")
	f.Add("one two", "1")
	f.Add("one two", "100")

	// Long words
	f.Add("supercalifragilisticexpialidocious", "5")
	f.Add("a b c d e f g h i j", "1")

	// Multiple spaces
	f.Add("one  two   three    four", "2")
	f.Add("   leading spaces", "2")
	f.Add("trailing spaces   ", "2")

	// Newlines
	f.Add("one\ntwo\nthree", "2")
	f.Add("one two\nthree four", "2")

	// Unicode
	f.Add("你好 世界 测试", "2")
	f.Add("🎉 🎊 🎁 🎀", "2")

	f.Fuzz(func(t *testing.T, input, wrapAt string) {
		result, err := filterWordwrap(AsValue(input), AsValue(wrapAt))
		if err != nil {
			return
		}
		_ = result.String()
	})
}

// FuzzFilterCenter fuzzes the center filter.
func FuzzFilterCenter(f *testing.F) {
	f.Add("test", "10")
	f.Add("test", "20")
	f.Add("test", "4")
	f.Add("test", "3")
	f.Add("test", "0")
	f.Add("test", "-1")
	f.Add("test", "1")
	f.Add("", "10")
	f.Add("a", "10")
	f.Add("你好", "10")
	f.Add("🎉", "10")
	f.Add(strings.Repeat("a", 100), "10")
	f.Add("test", "1000")
	f.Add("test", "10000")
	f.Add("test", "100000")

	f.Fuzz(func(t *testing.T, input, width string) {
		result, err := filterCenter(AsValue(input), AsValue(width))
		if err != nil {
			// Errors expected for very large widths
			return
		}
		_ = result.String()
	})
}

// FuzzFilterLjustRjust fuzzes the ljust and rjust filters.
func FuzzFilterLjustRjust(f *testing.F) {
	f.Add("test", "10")
	f.Add("test", "4")
	f.Add("test", "3")
	f.Add("test", "0")
	f.Add("test", "-1")
	f.Add("", "10")
	f.Add("你好", "10")
	f.Add("🎉", "10")
	f.Add("test", "100")
	f.Add("test", "1000")
	f.Add("test", "10000")

	f.Fuzz(func(t *testing.T, input, width string) {
		// Test ljust
		result1, err1 := filterLjust(AsValue(input), AsValue(width))
		if err1 == nil {
			_ = result1.String()
		}

		// Test rjust
		result2, err2 := filterRjust(AsValue(input), AsValue(width))
		if err2 == nil {
			_ = result2.String()
		}
	})
}

// FuzzFilterIriencode fuzzes the iriencode filter.
func FuzzFilterIriencode(f *testing.F) {
	// Basic URLs
	f.Add("/path/to/page")
	f.Add("/search?q=test")
	f.Add("/path?a=1&b=2")
	f.Add("/path#anchor")

	// URLs with spaces
	f.Add("/path with spaces")
	f.Add("/search?q=hello world")
	f.Add("path/file name.txt")

	// Special characters
	f.Add("/path<script>")
	f.Add("/path\"quotes\"")
	f.Add("/path'apostrophe'")
	f.Add("/path&ampersand")

	// Unicode
	f.Add("/你好世界")
	f.Add("/こんにちは")
	f.Add("/Привіт")
	f.Add("/🎉🎊")
	f.Add("/path/日本語/test")

	// IRI-safe characters that should be preserved
	f.Add("/#%[]=:;$&()+,!?*@'~")
	f.Add("/path#fragment")
	f.Add("/path?query=value")
	f.Add("/path[0]")
	f.Add("/path:8080")

	// Empty and whitespace
	f.Add("")
	f.Add(" ")
	f.Add("   ")

	// Long paths
	f.Add(strings.Repeat("/path", 100))
	f.Add("/" + strings.Repeat("a", 1000))

	f.Fuzz(func(t *testing.T, input string) {
		result, err := filterIriencode(AsValue(input), nil)
		if err != nil {
			t.Fatalf("iriencode error: %v", err)
		}
		_ = result.String()
	})
}

// FuzzFilterGetdigit fuzzes the get_digit filter.
func FuzzFilterGetdigit(f *testing.F) {
	// Basic numbers
	f.Add("123456789", "1")
	f.Add("123456789", "2")
	f.Add("123456789", "5")
	f.Add("123456789", "9")

	// Edge cases
	f.Add("123456789", "0")
	f.Add("123456789", "-1")
	f.Add("123456789", "10")
	f.Add("123456789", "100")
	f.Add("0", "1")
	f.Add("1", "1")
	f.Add("", "1")

	// Non-numeric strings
	f.Add("abc", "1")
	f.Add("12abc34", "1")
	f.Add("12.34", "1")
	f.Add("-123", "1")

	// Unicode
	f.Add("一二三", "1")
	f.Add("١٢٣", "1")

	// Long numbers
	f.Add(strings.Repeat("1", 100), "50")
	f.Add(strings.Repeat("9", 100), "1")

	f.Fuzz(func(t *testing.T, input, digit string) {
		result, err := filterGetdigit(AsValue(input), AsValue(digit))
		if err != nil {
			return
		}
		_ = result.String()
	})
}

// FuzzFilterSplit fuzzes the split filter.
func FuzzFilterSplit(f *testing.F) {
	// Basic splitting
	f.Add("a,b,c", ",")
	f.Add("a|b|c", "|")
	f.Add("a b c", " ")
	f.Add("a::b::c", "::")
	f.Add("a---b---c", "---")

	// Edge cases
	f.Add("", ",")
	f.Add("abc", ",")
	f.Add(",", ",")
	f.Add(",,", ",")
	f.Add(",,,", ",")
	f.Add("a,", ",")
	f.Add(",a", ",")
	f.Add("a,,b", ",")

	// Empty delimiter
	f.Add("abc", "")
	f.Add("hello", "")

	// Unicode
	f.Add("你好,世界", ",")
	f.Add("🎉|🎊|🎁", "|")
	f.Add("日本語テスト", "")

	// Long inputs
	f.Add(strings.Repeat("a,", 100)+"a", ",")
	f.Add(strings.Repeat("word ", 100), " ")

	f.Fuzz(func(t *testing.T, input, delimiter string) {
		result, err := filterSplit(AsValue(input), AsValue(delimiter))
		if err != nil {
			return
		}
		_ = result.String()
	})
}

// FuzzFilterLinebreaks fuzzes the linebreaks filter.
func FuzzFilterLinebreaks(f *testing.F) {
	// Basic line breaks
	f.Add("hello\nworld")
	f.Add("hello\r\nworld")
	f.Add("hello\rworld")
	f.Add("line1\nline2\nline3")
	f.Add("para1\n\npara2")
	f.Add("para1\n\n\npara2")

	// Edge cases
	f.Add("")
	f.Add("\n")
	f.Add("\n\n")
	f.Add("\n\n\n")
	f.Add("no line breaks")
	f.Add("\nleading")
	f.Add("trailing\n")
	f.Add("\nboth\n")

	// Unicode
	f.Add("你好\n世界")
	f.Add("🎉\n🎊")

	// Mixed line endings
	f.Add("a\nb\r\nc\rd")
	f.Add("\r\n\r\n")
	f.Add("\n\r\n\r")

	f.Fuzz(func(t *testing.T, input string) {
		// Test linebreaks
		result1, err1 := filterLinebreaks(AsValue(input), nil)
		if err1 == nil {
			_ = result1.String()
		}

		// Test linebreaksbr
		result2, err2 := filterLinebreaksbr(AsValue(input), nil)
		if err2 == nil {
			_ = result2.String()
		}
	})
}

// FuzzFilterMakelist fuzzes the make_list filter.
func FuzzFilterMakelist(f *testing.F) {
	f.Add("")
	f.Add("a")
	f.Add("abc")
	f.Add("hello world")
	f.Add("你好世界")
	f.Add("🎉🎊🎁")
	f.Add(strings.Repeat("a", 100))
	f.Add("12345")
	f.Add("   ")
	f.Add("\n\t\r")

	f.Fuzz(func(t *testing.T, input string) {
		result, err := filterMakelist(AsValue(input), nil)
		if err != nil {
			return
		}
		// Verify result length matches rune count
		inputRunes := []rune(input)
		if result.Len() != len(inputRunes) {
			t.Errorf("make_list length mismatch: input runes=%d, result len=%d",
				len(inputRunes), result.Len())
		}
	})
}

// FuzzFilterRemovetags fuzzes the removetags filter for security testing.
func FuzzFilterRemovetags(f *testing.F) {
	// Add seed corpus with known problematic inputs
	f.Add("<script>alert(1)</script>", "script")
	f.Add("<sc<script>ript>alert(1)</sc</script>ript>", "script")
	f.Add("<scr<scr<script>ipt>ipt>x</scr</scr</script>ipt>ipt>", "script")
	f.Add("<SCRIPT>alert(1)</SCRIPT>", "script")
	f.Add("<script src='x'>", "script")
	f.Add("<script>", "script")
	f.Add("</script>", "script")
	f.Add("<script >alert(1)</script >", "script")
	f.Add("<a href='javascript:alert(1)'>click</a>", "a")
	f.Add("<img src=x onerror=alert(1)>", "img")
	f.Add("<<script>script>alert(1)<</script>/script>", "script")
	f.Add("<svg onload=alert(1)>", "svg")
	f.Add("<div><script>alert(1)</script></div>", "script")
	f.Add("normal text", "script")
	f.Add("", "script")
	f.Add("<b>bold</b>", "b,i,script")

	// Case variations
	f.Add("<ScRiPt>alert(1)</ScRiPt>", "script")
	f.Add("<script  >alert(1)</script  >", "script")
	f.Add("<script\t>alert(1)</script\t>", "script")
	f.Add("<script\n>alert(1)</script\n>", "script")
	f.Add("<script\r>alert(1)</script\r>", "script")
	f.Add("<script/>", "script")
	f.Add("<script type='text/javascript'>", "script")
	f.Add("<script language='javascript'>", "script")
	f.Add("<script defer>alert(1)</script>", "script")
	f.Add("<script async>alert(1)</script>", "script")

	// More nested injection attempts
	f.Add("<script<script>>alert(1)</script</script>>", "script")
	f.Add("<scri<script>pt>alert(1)</scri</script>pt>", "script")
	f.Add("<s<script>cript>alert(1)</s</script>cript>", "script")
	f.Add("<scr\x00ipt>alert(1)</scr\x00ipt>", "script")

	// Encoding bypass attempts
	f.Add("<\x00script>alert(1)</script>", "script")
	f.Add("<script\x00>alert(1)</script>", "script")
	f.Add("<%73cript>alert(1)</%73cript>", "script")
	f.Add("<&#115;cript>alert(1)</&#115;cript>", "script")
	f.Add("<&#x73;cript>alert(1)</&#x73;cript>", "script")
	f.Add("<script&#x09;>alert(1)</script>", "script")
	f.Add("<script&#x0A;>alert(1)</script>", "script")
	f.Add("<script&#x0D;>alert(1)</script>", "script")

	// Other dangerous tags - img
	f.Add("<IMG SRC=x ONERROR=alert(1)>", "img")
	f.Add("<img/src=x/onerror=alert(1)>", "img")
	f.Add("<img src=x onerror='alert(1)'>", "img")
	f.Add("<img src=1 onerror=alert(1) />", "img")

	// SVG variations
	f.Add("<SVG ONLOAD=alert(1)>", "svg")
	f.Add("<svg/onload=alert(1)>", "svg")
	f.Add("<svg onload='alert(1)'/>", "svg")
	f.Add("<svg><script>alert(1)</script></svg>", "svg,script")
	f.Add("<svg><animate onbegin=alert(1)></svg>", "svg")

	// Anchor tag variations
	f.Add("<A HREF='javascript:alert(1)'>click</A>", "a")
	f.Add("<a href='  javascript:alert(1)'>click</a>", "a")
	f.Add("<a href=javascript:alert(1)>click</a>", "a")

	// Iframe
	f.Add("<iframe src='javascript:alert(1)'></iframe>", "iframe")
	f.Add("<IFRAME SRC='javascript:alert(1)'></IFRAME>", "iframe")
	f.Add("<iframe src=javascript:alert(1)>", "iframe")
	f.Add("<iframe srcdoc='<script>alert(1)</script>'>", "iframe")

	// Other dangerous elements
	f.Add("<object data='javascript:alert(1)'>", "object")
	f.Add("<embed src='javascript:alert(1)'>", "embed")
	f.Add("<form action='javascript:alert(1)'>", "form")
	f.Add("<input onfocus=alert(1) autofocus>", "input")
	f.Add("<select onfocus=alert(1) autofocus>", "select")
	f.Add("<textarea onfocus=alert(1) autofocus>", "textarea")
	f.Add("<video><source onerror=alert(1)></video>", "video,source")
	f.Add("<audio><source onerror=alert(1)></audio>", "audio,source")
	f.Add("<body onload=alert(1)>", "body")
	f.Add("<marquee onstart=alert(1)>", "marquee")

	// Event handler variations
	f.Add("<div onclick=alert(1)>click</div>", "div")
	f.Add("<div onmouseover=alert(1)>hover</div>", "div")
	f.Add("<div onmouseout=alert(1)>leave</div>", "div")
	f.Add("<div onfocus=alert(1) tabindex=0>focus</div>", "div")
	f.Add("<div onblur=alert(1) tabindex=0>blur</div>", "div")
	f.Add("<div onkeydown=alert(1)>key</div>", "div")
	f.Add("<div ondblclick=alert(1)>double</div>", "div")
	f.Add("<div oncontextmenu=alert(1)>right</div>", "div")

	// Style-based XSS
	f.Add("<style>*{background:url('javascript:alert(1)')}</style>", "style")
	f.Add("<div style='background:url(javascript:alert(1))'>", "div")
	f.Add("<link rel=stylesheet href='javascript:alert(1)'>", "link")

	// HTML5 elements
	f.Add("<details open ontoggle=alert(1)>", "details")
	f.Add("<base href='javascript:alert(1)//'>", "base")

	// Multiple tags to remove
	f.Add("<b><i><script>alert(1)</script></i></b>", "b,i,script")
	f.Add("<p><script>x</script></p>", "p,script")
	f.Add("<span><b><i>text</i></b></span>", "span,b,i")

	// Edge cases
	f.Add("<>", "script")
	f.Add("</>", "script")
	f.Add("< script>alert(1)</ script>", "script")
	f.Add("<script >alert(1)< /script>", "script")
	f.Add("<!--<script>alert(1)</script>-->", "script")
	f.Add("<![CDATA[<script>alert(1)</script>]]>", "script")

	// Unicode variations
	f.Add("\u003cscript\u003ealert(1)\u003c/script\u003e", "script")

	// Malformed HTML
	f.Add("<script<", "script")
	f.Add(">script<", "script")
	f.Add("<script", "script")
	f.Add("script>", "script")
	f.Add("<script>>alert(1)</script>", "script")
	f.Add("<<script>alert(1)</script>>", "script")
	f.Add("<script>alert(1)<script>", "script")
	f.Add("</script>alert(1)<script>", "script")
	f.Add("<script x>alert(1)</script y>", "script")
	f.Add("<script/x>alert(1)</script/y>", "script")

	// Long inputs
	f.Add("<script>"+string(make([]byte, 1000))+"</script>", "script")
	f.Add(string(make([]byte, 100))+"<script>alert(1)</script>"+string(make([]byte, 100)), "script")

	// Multiple tag names with different formats
	f.Add("<script>x</script><style>y</style><iframe>z</iframe>", "script,style,iframe")
	f.Add("<a>1</a><b>2</b><c>3</c>", "a, b, c")
	f.Add("<a>1</a><b>2</b><c>3</c>", " a , b , c ")
	f.Add("<tag>test</tag>", "tag")
	f.Add("<customtag>test</customtag>", "customtag")

	// Empty and whitespace
	f.Add("<script></script>", "script")
	f.Add("<script>   </script>", "script")
	f.Add("<script>\n\t\r</script>", "script")
	f.Add("   <script>x</script>   ", "script")

	// Self-closing
	f.Add("<script />", "script")
	f.Add("<br/>", "br")
	f.Add("<br />", "br")
	f.Add("<hr/>", "hr")
	f.Add("<img/>", "img")
	f.Add("<input/>", "input")

	f.Fuzz(func(t *testing.T, input, tags string) {
		// Skip invalid tag names
		if tags == "" {
			return
		}

		result, err := filterRemovetags(AsValue(input), AsValue(tags))
		if err != nil {
			// Errors are expected for invalid tag names
			return
		}

		// The result should not contain any of the specified tags
		tagList := strings.Split(tags, ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			// Skip invalid tag names (same validation as the filter)
			if !reTagName.MatchString(tag) {
				continue
			}
			// Check for opening/closing tags using the same regex pattern as the filter
			// This ensures we only match actual tags, not substrings (e.g., <s vs <svg)
			re := regexp.MustCompile(fmt.Sprintf(`(?i)</?%s(?:\s[^>]*)?/?>`, regexp.QuoteMeta(tag)))
			if re.MatchString(result.String()) {
				t.Errorf("result still contains <%s> tag: input=%q, tags=%q, result=%q",
					tag, input, tags, result.String())
			}
		}
	})
}

// FuzzBuiltinFilters fuzzes all built-in filters with various inputs.
func FuzzBuiltinFilters(f *testing.F) {
	// Original seed corpus
	f.Add("foobar", "123")
	f.Add("foobar", `123,456`)
	f.Add("foobar", `123,456,"789"`)
	f.Add("foobar", `"test","test"`)
	f.Add("foobar", `123,"test"`)
	f.Add("foobar", "")
	f.Add("123", "foobar")

	// Empty and whitespace values
	f.Add("", "")
	f.Add("", "test")
	f.Add("   ", "")
	f.Add("   ", "test")
	f.Add("\t\n\r", "")
	f.Add("\t\n\r", "test")

	// String values
	f.Add("hello", "")
	f.Add("hello world", "")
	f.Add("hello world", " ")
	f.Add("HELLO", "")
	f.Add("Hello World", "")
	f.Add("hello\nworld", "")
	f.Add("hello\tworld", "")
	f.Add("hello\r\nworld", "")
	f.Add("  hello  ", "")
	f.Add("a", "")
	f.Add("ab", "")
	f.Add("abc", "")

	// Numeric string values
	f.Add("0", "")
	f.Add("1", "")
	f.Add("-1", "")
	f.Add("123", "")
	f.Add("-123", "")
	f.Add("123.456", "")
	f.Add("-123.456", "")
	f.Add("0.0", "")
	f.Add("0.001", "")
	f.Add(".5", "")
	f.Add("5.", "")
	f.Add("1e10", "")
	f.Add("1E10", "")
	f.Add("1e-10", "")
	f.Add("1.5e10", "")
	f.Add("999999999999999999", "")
	f.Add("-999999999999999999", "")

	// Numeric filter arguments
	f.Add("test", "0")
	f.Add("test", "1")
	f.Add("test", "-1")
	f.Add("test", "100")
	f.Add("test", "-100")
	f.Add("test", "0.5")
	f.Add("test", "-0.5")
	f.Add("test", "1.5")
	f.Add("test", "999999999")
	f.Add("test", "-999999999")

	// Boolean-like values
	f.Add("true", "")
	f.Add("false", "")
	f.Add("True", "")
	f.Add("False", "")
	f.Add("TRUE", "")
	f.Add("FALSE", "")
	f.Add("yes", "")
	f.Add("no", "")
	f.Add("on", "")
	f.Add("off", "")
	f.Add("1", "true")
	f.Add("0", "false")

	// HTML content
	f.Add("<b>bold</b>", "")
	f.Add("<script>alert(1)</script>", "")
	f.Add("<p>paragraph</p>", "")
	f.Add("<div><p>nested</p></div>", "")
	f.Add("&lt;script&gt;", "")
	f.Add("&amp;&lt;&gt;&quot;", "")
	f.Add("<a href='test'>link</a>", "")
	f.Add("<img src='test.png'>", "")

	// URL content
	f.Add("http://example.com", "")
	f.Add("https://example.com/path?query=1&other=2", "")
	f.Add("http://example.com/path with spaces", "")
	f.Add("ftp://files.example.com", "")
	f.Add("mailto:test@example.com", "")
	f.Add("/path/to/file", "")
	f.Add("../relative/path", "")

	// Special characters
	f.Add("hello'world", "")
	f.Add(`hello"world`, "")
	f.Add("hello\\world", "")
	f.Add("hello/world", "")
	f.Add("hello|world", "")
	f.Add("hello:world", "")
	f.Add("test@example.com", "")
	f.Add("price: $100", "")
	f.Add("100%", "")
	f.Add("foo & bar", "")
	f.Add("a < b > c", "")
	f.Add("(parentheses)", "")
	f.Add("[brackets]", "")
	f.Add("{braces}", "")

	// Unicode content
	f.Add("你好世界", "")
	f.Add("こんにちは", "")
	f.Add("Привіт світ", "")
	f.Add("مرحبا", "")
	f.Add("שלום", "")
	f.Add("Ελληνικά", "")
	f.Add("🎉🎊🎁", "")
	f.Add("café", "")
	f.Add("naïve", "")
	f.Add("日本語テスト", "")
	f.Add("你好", "2")
	f.Add("🎉🎊", "1")

	// Slice filter arguments
	f.Add("hello world", ":5")
	f.Add("hello world", "5:")
	f.Add("hello world", "2:5")
	f.Add("hello world", "-5:")
	f.Add("hello world", ":-5")
	f.Add("hello world", "-5:-2")
	f.Add("hello world", ":0")
	f.Add("hello world", "0:")
	f.Add("hello world", ":")
	f.Add("hello world", "::")
	f.Add("hello world", "100:")
	f.Add("hello world", ":100")
	f.Add("hello world", "-100:")
	f.Add("hello world", ":-100")

	// Date/time format arguments
	f.Add("2024-01-15", "Y-m-d")
	f.Add("2024-01-15", "d/m/Y")
	f.Add("2024-01-15", "F j, Y")
	f.Add("15:30:45", "H:i:s")
	f.Add("15:30:45", "g:i A")
	f.Add("now", "Y-m-d H:i:s")

	// Pluralize arguments
	f.Add("1", "y,ies")
	f.Add("2", "y,ies")
	f.Add("0", "y,ies")
	f.Add("1", "es")
	f.Add("2", "es")
	f.Add("1", "")
	f.Add("2", "")

	// Yesno arguments
	f.Add("true", "yes,no")
	f.Add("false", "yes,no")
	f.Add("", "yes,no,maybe")
	f.Add("true", "ja,nein,vielleicht")
	f.Add("false", "是,否,也许")

	// Float format arguments
	f.Add("3.14159", "0")
	f.Add("3.14159", "1")
	f.Add("3.14159", "2")
	f.Add("3.14159", "3")
	f.Add("3.14159", "-1")
	f.Add("3.14159", "-2")
	f.Add("3.14159", "-3")
	f.Add("100.00", "2")
	f.Add("100.00", "-2")

	// Truncation arguments
	f.Add("hello world test", "5")
	f.Add("hello world test", "10")
	f.Add("hello world test", "0")
	f.Add("hello world test", "1")
	f.Add("hello world test", "100")
	f.Add("hello world test", "-1")

	// Width/alignment arguments
	f.Add("test", "10")
	f.Add("test", "20")
	f.Add("test", "1")
	f.Add("test", "0")
	f.Add("test", "100")

	// Stringformat arguments
	f.Add("42", "%d")
	f.Add("42", "%05d")
	f.Add("42", "%-5d")
	f.Add("3.14", "%.2f")
	f.Add("3.14", "%10.2f")
	f.Add("hello", "%s")
	f.Add("hello", "%10s")
	f.Add("hello", "%-10s")
	f.Add("65", "%c")
	f.Add("255", "%x")
	f.Add("255", "%X")
	f.Add("255", "%o")
	f.Add("255", "%b")

	// List-like values
	f.Add("a,b,c", ",")
	f.Add("a|b|c", "|")
	f.Add("a b c", " ")
	f.Add("a  b  c", " ")
	f.Add("a::b::c", "::")
	f.Add("", ",")
	f.Add(",,,", ",")
	f.Add("a,", ",")
	f.Add(",a", ",")

	// Edge cases for specific filters
	f.Add("1234567890", "1")
	f.Add("1234567890", "5")
	f.Add("1234567890", "10")
	f.Add("1234567890", "0")
	f.Add("1234567890", "15")

	// Filesizeformat values
	f.Add("0", "")
	f.Add("1", "")
	f.Add("1023", "")
	f.Add("1024", "")
	f.Add("1048576", "")
	f.Add("1073741824", "")
	f.Add("1099511627776", "")

	// Phone2numeric values
	f.Add("1-800-TEST", "")
	f.Add("1-800-PONGO", "")
	f.Add("555-1234", "")
	f.Add("ABC-DEFG", "")

	// Slugify values
	f.Add("Hello World", "")
	f.Add("Hello, World!", "")
	f.Add("  spaces  around  ", "")
	f.Add("multiple---hyphens", "")
	f.Add("UPPERCASE", "")
	f.Add("Special!@#$%Characters", "")
	f.Add("MixedCase123Numbers", "")

	// JSON values
	f.Add(`{"key": "value"}`, "")
	f.Add(`[1, 2, 3]`, "")
	f.Add(`"string"`, "")
	f.Add(`123`, "")
	f.Add(`true`, "")
	f.Add(`null`, "")

	// Complex filter arguments with quotes
	f.Add("test", `"quoted"`)
	f.Add("test", `'single'`)
	f.Add("test", `"with spaces"`)
	f.Add("test", `"with,comma"`)
	f.Add("test", `"with:colon"`)
	f.Add("test", `"with|pipe"`)

	// Very long values
	f.Add(string(make([]byte, 100)), "")
	f.Add(string(make([]byte, 1000)), "")
	f.Add("test", string(make([]byte, 100)))

	// Null byte and control characters
	f.Add("hello\x00world", "")
	f.Add("test\x01\x02\x03", "")
	f.Add("\x00", "")
	f.Add("", "\x00")

	// Mixed content
	f.Add("Hello 世界 мир 🌍", "")
	f.Add("<b>Hello</b> 世界", "")
	f.Add("123 test 456", "")
	f.Add("test\n<br>\nmore", "")

	f.Fuzz(func(t *testing.T, value, filterArg string) {
		ts := NewSet("fuzz-test", &DummyLoader{})
		for name := range builtinFilters {
			tpl, err := ts.FromString(fmt.Sprintf("{{ %v|%v:%v }}", value, name, filterArg))
			if tpl != nil && err != nil {
				t.Errorf("filter=%q value=%q, filterArg=%q, err=%v", name, value, filterArg, err)
			}
			if err == nil {
				tpl.Execute(nil) //nolint:errcheck
			}
		}
	})
}
