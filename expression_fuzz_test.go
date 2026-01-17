package pongo2

import (
	"bytes"
	"testing"
)

// FuzzExpressionParsing fuzzes complex expression parsing and evaluation.
func FuzzExpressionParsing(f *testing.F) {
	// Basic arithmetic
	f.Add("1 + 1")
	f.Add("10 - 5")
	f.Add("3 * 4")
	f.Add("10 / 2")
	f.Add("10 % 3")
	f.Add("2 ^ 3")

	// Negative numbers
	f.Add("-1")
	f.Add("-1 + 1")
	f.Add("1 + -1")
	f.Add("-1 * -1")
	f.Add("--1")
	f.Add("---1")

	// Float operations
	f.Add("1.5 + 1.5")
	f.Add("3.14 * 2")
	f.Add("10.0 / 3.0")
	f.Add("0.1 + 0.2")

	// Division edge cases
	f.Add("10 / 0")
	f.Add("10.0 / 0.0")
	f.Add("0 / 0")
	f.Add("10 % 0")
	f.Add("1 / 0.0000001")

	// Comparison operators
	f.Add("1 == 1")
	f.Add("1 != 1")
	f.Add("1 < 2")
	f.Add("1 > 2")
	f.Add("1 <= 1")
	f.Add("1 >= 1")
	f.Add("1 <> 1")

	// Logical operators
	f.Add("true && true")
	f.Add("true || false")
	f.Add("!true")
	f.Add("not true")
	f.Add("true and false")
	f.Add("true or false")

	// Complex expressions
	f.Add("1 + 2 * 3")
	f.Add("(1 + 2) * 3")
	f.Add("1 + (2 * 3)")
	f.Add("((1 + 2) * (3 + 4))")
	f.Add("1 + 2 + 3 + 4 + 5")
	f.Add("1 * 2 * 3 * 4 * 5")
	f.Add("2 ^ 2 ^ 2")
	f.Add("(2 ^ 2) ^ 2")
	f.Add("2 ^ (2 ^ 2)")

	// Mixed types
	f.Add(`"a" + "b"`)
	f.Add("1 + \"a\"")
	f.Add("\"a\" + 1")
	f.Add(`"hello" == "hello"`)
	f.Add(`"a" < "b"`)

	// Nested parentheses
	f.Add("((((1))))")
	f.Add("(((1 + 2)))")
	f.Add("((1) + (2))")
	f.Add("(((((((1)))))))")

	// In operator
	f.Add("1 in [1, 2, 3]")
	f.Add(`"a" in "abc"`)

	// Chained comparisons
	f.Add("1 < 2 && 2 < 3")
	f.Add("1 == 1 && 2 == 2")
	f.Add("1 != 2 || 2 != 3")

	// Large numbers
	f.Add("999999999 + 1")
	f.Add("999999999 * 999999999")
	f.Add("2 ^ 30")
	f.Add("2 ^ 62")

	// Edge cases
	f.Add("")
	f.Add(" ")
	f.Add("   ")
	f.Add("()")
	f.Add("(")
	f.Add(")")
	f.Add("+ +")
	f.Add("1 +")
	f.Add("+ 1")
	f.Add("1 1")
	f.Add("* 1")
	f.Add("1 *")

	// Unicode in expressions
	f.Add(`"你好" + "世界"`)
	f.Add(`"🎉" + "🎊"`)

	// Boolean edge cases
	f.Add("true")
	f.Add("false")
	f.Add("true && false || true")
	f.Add("!(!(!true))")
	f.Add("not not not true") //nolint:dupword // intentional triple negation test case

	// Power operations
	f.Add("2 ^ 0")
	f.Add("2 ^ 1")
	f.Add("2 ^ -1")
	f.Add("0 ^ 0")
	f.Add("0 ^ 1")
	f.Add("1 ^ 0")
	f.Add("-1 ^ 2")
	f.Add("-1 ^ 3")
	f.Add("0.5 ^ 2")
	f.Add("2 ^ 0.5")

	f.Fuzz(func(t *testing.T, expr string) {
		// Wrap expression in template syntax
		tplStr := "{{ " + expr + " }}"

		ts := NewSet("fuzz-test", &DummyLoader{})
		tpl, err := ts.FromString(tplStr)
		if err != nil {
			// Parse errors are expected for malformed expressions
			return
		}

		// Execute with empty context
		var buf bytes.Buffer
		err = tpl.ExecuteWriter(Context{}, &buf)
		if err != nil {
			// Execution errors (like division by zero) are expected
			return
		}

		// If we get here, the expression was valid and executed
		_ = buf.String()
	})
}
