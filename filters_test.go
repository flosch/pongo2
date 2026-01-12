package pongo2

import "testing"

func TestReplaceFilter(t *testing.T) {
	t.Run("non-existent filter", func(t *testing.T) {
		err := ReplaceFilter("nonexistent_filter_xyz", func(in *Value, param *Value) (*Value, error) {
			return in, nil
		})
		if err == nil {
			t.Error("ReplaceFilter should return error for non-existent filter")
		}
	})

	t.Run("existing filter", func(t *testing.T) {
		originalFn := filters["upper"]
		defer func() { filters["upper"] = originalFn }()

		newFn := func(in *Value, param *Value) (*Value, error) {
			return AsValue("REPLACED"), nil
		}

		err := ReplaceFilter("upper", newFn)
		if err != nil {
			t.Errorf("ReplaceFilter failed for existing filter: %v", err)
		}

		result, err := ApplyFilter("upper", AsValue("test"), nil)
		if err != nil {
			t.Fatalf("ApplyFilter failed: %v", err)
		}
		if result.String() != "REPLACED" {
			t.Errorf("Filter was not replaced correctly, got %s", result.String())
		}
	})
}

func TestMustApplyFilter(t *testing.T) {
	t.Run("successful apply", func(t *testing.T) {
		result := MustApplyFilter("upper", AsValue("hello"), nil)
		if result.String() != "HELLO" {
			t.Errorf("MustApplyFilter returned %s, want HELLO", result.String())
		}
	})

	t.Run("panic on non-existent filter", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustApplyFilter should panic for non-existent filter")
			}
		}()
		MustApplyFilter("nonexistent_filter_xyz", AsValue("test"), nil)
	})
}
