package pongo2

import (
	"bytes"
	"math"
	"strings"
	"testing"
)

// FuzzValueOperations fuzzes the Value type's conversion and operation methods.
func FuzzValueOperations(f *testing.F) {
	// Integers
	f.Add("0")
	f.Add("1")
	f.Add("-1")
	f.Add("123")
	f.Add("-123")
	f.Add("2147483647")
	f.Add("-2147483648")
	f.Add("9223372036854775807")
	f.Add("-9223372036854775808")
	f.Add("18446744073709551615")

	// Floats
	f.Add("0.0")
	f.Add("1.0")
	f.Add("-1.0")
	f.Add("3.14159265358979323846")
	f.Add("-3.14159265358979323846")
	f.Add("1.7976931348623157e+308")
	f.Add("-1.7976931348623157e+308")
	f.Add("2.2250738585072014e-308")
	f.Add("1e10")
	f.Add("1e-10")
	f.Add("1.5e10")
	f.Add("1.5e-10")
	f.Add(".5")
	f.Add("5.")
	f.Add("-.5")
	f.Add("-5.")
	f.Add("0.000000001")
	f.Add("999999999.999999999")

	// Boolean-like
	f.Add("true")
	f.Add("false")
	f.Add("True")
	f.Add("False")
	f.Add("TRUE")
	f.Add("FALSE")
	f.Add("yes")
	f.Add("no")
	f.Add("on")
	f.Add("off")

	// Strings
	f.Add("")
	f.Add(" ")
	f.Add("   ")
	f.Add("hello")
	f.Add("HELLO")
	f.Add("Hello World")
	f.Add("hello\nworld")
	f.Add("hello\tworld")
	f.Add("hello\rworld")
	f.Add("hello\x00world")

	// Unicode strings
	f.Add("你好")
	f.Add("こんにちは")
	f.Add("Привіт")
	f.Add("Ελληνικά")
	f.Add("🎉🎊🎁")
	f.Add("café")
	f.Add("naïve")
	f.Add("日本語テスト")
	f.Add("مرحبا")
	f.Add("שלום")
	f.Add("हिंदी")

	// Special numeric strings
	f.Add("NaN")
	f.Add("Inf")
	f.Add("-Inf")
	f.Add("+Inf")
	f.Add("infinity")
	f.Add("-infinity")
	f.Add("nan")

	// Edge case strings
	f.Add("   123   ")
	f.Add("123abc")
	f.Add("abc123")
	f.Add("12.34.56")
	f.Add("1,234")
	f.Add("1,234.56")
	f.Add("$100")
	f.Add("100%")
	f.Add("+123")
	f.Add("++123")
	f.Add("--123")
	f.Add("1e")
	f.Add("e1")
	f.Add("1e+")
	f.Add("1e-")
	f.Add(".")
	f.Add("..")
	f.Add("-")
	f.Add("+")
	f.Add("+-")
	f.Add("-+")

	// Control characters
	f.Add("\x00")
	f.Add("\x01\x02\x03")
	f.Add("\n\r\t")
	f.Add("\v\f")

	// Long strings
	f.Add(strings.Repeat("a", 100))
	f.Add(strings.Repeat("1", 100))
	f.Add(strings.Repeat("9", 50))
	f.Add(strings.Repeat("0", 100))

	f.Fuzz(func(t *testing.T, input string) {
		v := AsValue(input)

		// Test all Value operations - none should panic
		_ = v.IsString()
		_ = v.IsBool()
		_ = v.IsFloat()
		_ = v.IsInteger()
		_ = v.IsNumber()
		_ = v.IsTime()
		_ = v.IsNil()
		_ = v.IsTrue()
		_ = v.String()
		_ = v.Integer()
		_ = v.Float()
		_ = v.Len()
		_ = v.CanSlice()
		_ = v.IsSliceOrArray()
		_ = v.IsMap()
		_ = v.IsStruct()
		_ = v.Negate()
		_ = v.Interface()

		// Test EqualValueTo with self
		_ = v.EqualValueTo(v)

		// Test EqualValueTo with nil
		_ = v.EqualValueTo(AsValue(nil))

		// Test Contains
		_ = v.Contains(AsValue("a"))
		_ = v.Contains(AsValue(1))
		_ = v.Contains(v)

		// Test Slice and Index if possible
		if v.CanSlice() && v.Len() > 0 {
			_ = v.Index(0)
			if v.Len() > 1 {
				_ = v.Slice(0, 1)
			}
		}
	})
}

// FuzzValueIterate fuzzes the Value.Iterate and IterateOrder methods.
func FuzzValueIterate(f *testing.F) {
	// Arrays/slices
	f.Add("[]")
	f.Add("[1]")
	f.Add("[1, 2, 3]")
	f.Add("[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]")

	// Strings
	f.Add("")
	f.Add("a")
	f.Add("abc")
	f.Add("hello world")
	f.Add("你好世界")
	f.Add("🎉🎊🎁")

	// Long strings
	f.Add(strings.Repeat("a", 100))
	f.Add(strings.Repeat("ab", 50))

	f.Fuzz(func(t *testing.T, input string) {
		v := AsValue(input)

		// Test Iterate
		count := 0
		v.Iterate(func(idx, cnt int, key, value *Value) bool {
			count++
			if count > 10000 {
				// Prevent infinite loops
				return false
			}
			return true
		}, func() {})

		// Test IterateOrder with different options
		count = 0
		v.IterateOrder(func(idx, cnt int, key, value *Value) bool {
			count++
			return count < 10000
		}, func() {}, false, false)

		count = 0
		v.IterateOrder(func(idx, cnt int, key, value *Value) bool {
			count++
			return count < 10000
		}, func() {}, true, false)

		count = 0
		v.IterateOrder(func(idx, cnt int, key, value *Value) bool {
			count++
			return count < 10000
		}, func() {}, false, true)

		count = 0
		v.IterateOrder(func(idx, cnt int, key, value *Value) bool {
			count++
			return count < 10000
		}, func() {}, true, true)
	})
}

// FuzzValueNumericConversions tests numeric conversion edge cases.
func FuzzValueNumericConversions(f *testing.F) {
	// Integer edge cases
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(-1))
	f.Add(int64(math.MaxInt64))
	f.Add(int64(math.MinInt64))
	f.Add(int64(math.MaxInt32))
	f.Add(int64(math.MinInt32))

	f.Fuzz(func(t *testing.T, input int64) {
		v := AsValue(input)

		// All operations should be safe
		_ = v.Integer()
		_ = v.Float()
		_ = v.String()
		_ = v.IsInteger()
		_ = v.IsFloat()
		_ = v.IsNumber()
		_ = v.IsTrue()
		_ = v.Negate()
	})
}

// FuzzValueFloatConversions tests float conversion edge cases.
func FuzzValueFloatConversions(f *testing.F) {
	f.Add(0.0)
	f.Add(1.0)
	f.Add(-1.0)
	f.Add(math.MaxFloat64)
	f.Add(-math.MaxFloat64)
	f.Add(math.SmallestNonzeroFloat64)
	f.Add(-math.SmallestNonzeroFloat64)
	f.Add(math.Inf(1))
	f.Add(math.Inf(-1))
	f.Add(math.NaN())
	f.Add(0.1)
	f.Add(0.2)
	f.Add(0.3)
	f.Add(1e308)
	f.Add(1e-308)

	f.Fuzz(func(t *testing.T, input float64) {
		v := AsValue(input)

		// All operations should be safe (even with NaN/Inf)
		_ = v.Integer()
		_ = v.Float()
		_ = v.String()
		_ = v.IsInteger()
		_ = v.IsFloat()
		_ = v.IsNumber()
		_ = v.IsTrue()
		_ = v.Negate()
	})
}

// FuzzUTF8Handling tests UTF-8 edge cases across various functions.
func FuzzUTF8Handling(f *testing.F) {
	// Valid UTF-8
	f.Add("hello")
	f.Add("你好世界")
	f.Add("🎉🎊🎁")
	f.Add("Ελληνικά")
	f.Add("العربية")
	f.Add("עברית")
	f.Add("日本語")
	f.Add("한국어")

	// Mixed scripts
	f.Add("Hello 你好 🎉")
	f.Add("Test مرحبا тест")
	f.Add("1234 αβγδ ＡＢＣ")

	// Special Unicode
	f.Add("\u0000")               // Null
	f.Add("\u0001")               // SOH
	f.Add("\u001F")               // US
	f.Add("\u007F")               // DEL
	f.Add("\u0080")               // Padding
	f.Add("\u00A0")               // NBSP
	f.Add("\u2028")               // Line separator
	f.Add("\u2029")               // Paragraph separator
	f.Add("\uFEFF")               // BOM
	f.Add("\uFFFD")               // Replacement
	f.Add("\U0001F600")           // Emoji
	f.Add("\U0001F1FA\U0001F1F8") // Flag

	// Invalid UTF-8 sequences (as raw bytes)
	f.Add(string([]byte{0x80}))
	f.Add(string([]byte{0xFF}))
	f.Add(string([]byte{0xC0, 0x80}))
	f.Add(string([]byte{0xED, 0xA0, 0x80}))
	f.Add(string([]byte{0xF4, 0x90, 0x80, 0x80}))

	// Combining characters
	f.Add("e\u0301")      // é as e + combining acute
	f.Add("n\u0303")      // ñ as n + combining tilde
	f.Add("\u0041\u030A") // Å as A + combining ring

	// Zero-width characters
	f.Add("\u200B")   // Zero-width space
	f.Add("\u200C")   // ZWNJ
	f.Add("\u200D")   // ZWJ
	f.Add("a\u200Bb") // ZWS between chars

	f.Fuzz(func(t *testing.T, input string) {
		// Test that various operations handle the input safely
		v := AsValue(input)

		_ = v.String()
		_ = v.Len()

		if v.CanSlice() && v.Len() > 0 {
			_ = v.Index(0)
			if v.Len() > 1 {
				_ = v.Slice(0, 1)
			}
		}

		// Test with filters
		_, _ = filterUpper(v, nil)
		_, _ = filterLower(v, nil)
		_, _ = filterTitle(v, nil)
		_, _ = filterLength(v, nil)
		_, _ = filterMakelist(v, nil)

		// Test template parsing with this string
		tplStr := "{{ x }}"
		ts := NewSet("fuzz-test", &DummyLoader{})
		tpl, err := ts.FromString(tplStr)
		if err != nil {
			return
		}

		var buf bytes.Buffer
		_ = tpl.ExecuteWriter(Context{"x": input}, &buf)
	})
}
