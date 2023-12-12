package pongo2

import "testing"

func BenchmarkCheckForValidIdentifiers(b *testing.B) {
	input := `valid_key`

	b.Run("regex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			isValidIdentifierRegex(input)
		}
	})
	b.Run("char_check", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			isValidIdentifierCharCheck(input)
		}
	})
}
