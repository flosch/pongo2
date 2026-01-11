package pongo2

import "testing"

func TestValueIterate(t *testing.T) {
	t.Run("array iteration", func(t *testing.T) {
		arr := []int{1, 2, 3}
		v := AsValue(arr)

		var items []int
		v.Iterate(func(idx, count int, key, value *Value) bool {
			items = append(items, key.Integer())
			return true
		}, func() {
			t.Error("Empty function should not be called for non-empty array")
		})

		if len(items) != 3 {
			t.Errorf("Iterate got %d items, want 3", len(items))
		}
	})

	t.Run("empty array", func(t *testing.T) {
		emptyArr := []int{}
		v := AsValue(emptyArr)
		emptyCalled := false
		v.Iterate(func(idx, count int, key, value *Value) bool {
			t.Error("Iteration function should not be called for empty array")
			return true
		}, func() {
			emptyCalled = true
		})

		if !emptyCalled {
			t.Error("Empty function should be called for empty array")
		}
	})
}

func TestValueSlice(t *testing.T) {
	t.Run("slice on array", func(t *testing.T) {
		arr := []int{1, 2, 3, 4, 5}
		v := AsValue(arr)
		sliced := v.Slice(1, 4)

		if sliced.Len() != 3 {
			t.Errorf("Slice len = %d, want 3", sliced.Len())
		}
	})

	t.Run("slice on string", func(t *testing.T) {
		str := "hello"
		v := AsValue(str)
		sliced := v.Slice(1, 3)

		if sliced.String() != "el" {
			t.Errorf("Slice string = %q, want %q", sliced.String(), "el")
		}
	})

	t.Run("slice on unsupported type", func(t *testing.T) {
		num := 42
		v := AsValue(num)
		sliced := v.Slice(0, 1)

		if sliced.Len() != 0 {
			t.Error("Slice on unsupported type should return empty slice")
		}
	})
}

func TestValueIndex(t *testing.T) {
	t.Run("index on array", func(t *testing.T) {
		arr := []string{"a", "b", "c"}
		v := AsValue(arr)

		item := v.Index(1)
		if item.String() != "b" {
			t.Errorf("Index(1) = %q, want %q", item.String(), "b")
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		arr := []string{"a", "b", "c"}
		v := AsValue(arr)

		item := v.Index(10)
		if !item.IsNil() {
			t.Error("Index out of bounds should return nil")
		}
	})

	t.Run("index on string", func(t *testing.T) {
		str := "hello"
		v := AsValue(str)
		item := v.Index(1)
		if item.String() != "e" {
			t.Errorf("String Index(1) = %q, want %q", item.String(), "e")
		}
	})

	t.Run("string index out of bounds", func(t *testing.T) {
		str := "hello"
		v := AsValue(str)
		item := v.Index(100)
		if item.String() != "" {
			t.Error("String Index out of bounds should return empty string")
		}
	})

	t.Run("index on unsupported type", func(t *testing.T) {
		num := 42
		v := AsValue(num)
		_ = v.Index(0) // Should not panic
	})
}

func TestValueBoolEdgeCases(t *testing.T) {
	t.Run("Bool on non-bool type", func(t *testing.T) {
		v := AsValue(42)
		if v.Bool() != false {
			t.Error("Bool() on non-bool should return false")
		}
	})

	t.Run("Bool on true", func(t *testing.T) {
		v := AsValue(true)
		if v.Bool() != true {
			t.Error("Bool() on true should return true")
		}
	})

	t.Run("Bool on false", func(t *testing.T) {
		v := AsValue(false)
		if v.Bool() != false {
			t.Error("Bool() on false should return false")
		}
	})
}

func TestValueTime(t *testing.T) {
	v := AsValue("not a time")
	tm := v.Time()
	if !tm.IsZero() {
		t.Error("Time() on non-time value should return zero time")
	}
}

func TestValueContains(t *testing.T) {
	t.Run("struct contains field", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		s := TestStruct{Name: "test", Age: 25}
		v := AsValue(s)

		if !v.Contains(AsValue("Name")) {
			t.Error("Contains should return true for existing field")
		}
		if v.Contains(AsValue("NonExistent")) {
			t.Error("Contains should return false for non-existing field")
		}
	})

	t.Run("map contains key", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		v := AsValue(m)

		if !v.Contains(AsValue("a")) {
			t.Error("Contains should return true for existing key")
		}
		if v.Contains(AsValue("z")) {
			t.Error("Contains should return false for non-existing key")
		}
	})

	t.Run("map with int keys", func(t *testing.T) {
		mi := map[int]string{1: "one", 2: "two"}
		v := AsValue(mi)

		if !v.Contains(AsValue(1)) {
			t.Error("Contains should return true for existing int key")
		}
	})

	t.Run("string contains", func(t *testing.T) {
		str := AsValue("hello world")
		if !str.Contains(AsValue("world")) {
			t.Error("Contains should return true for substring")
		}
		if str.Contains(AsValue("xyz")) {
			t.Error("Contains should return false for non-substring")
		}
	})

	t.Run("slice contains", func(t *testing.T) {
		slice := AsValue([]int{1, 2, 3})
		if !slice.Contains(AsValue(2)) {
			t.Error("Contains should return true for element in slice")
		}
		if slice.Contains(AsValue(5)) {
			t.Error("Contains should return false for element not in slice")
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		m := map[string]int{"a": 1}
		v := AsValue(m)
		if v.Contains(AsValue(nil)) {
			t.Error("Contains should return false for nil key")
		}
	})
}

func TestValueNegate(t *testing.T) {
	tests := []struct {
		value    any
		expected bool
	}{
		{1, false},
		{0, true},
		{1.5, false},
		{0.0, true},
		{[]int{1, 2}, false},
		{[]int{}, true},
		{"hello", false},
		{"", true},
		{true, false},
		{false, true},
	}

	for _, tt := range tests {
		v := AsValue(tt.value)
		negated := v.Negate()
		if negated.IsTrue() != tt.expected {
			t.Errorf("Negate(%v).IsTrue() = %v, want %v", tt.value, negated.IsTrue(), tt.expected)
		}
	}
}

func TestValueIterateOrder(t *testing.T) {
	t.Run("sorted iteration", func(t *testing.T) {
		arr := []int{3, 1, 2}
		v := AsValue(arr)

		var items []int
		v.IterateOrder(func(idx, count int, key, value *Value) bool {
			items = append(items, key.Integer())
			return true
		}, func() {}, false, true)

		if items[0] != 1 || items[1] != 2 || items[2] != 3 {
			t.Errorf("Sorted iteration order wrong: %v", items)
		}
	})

	t.Run("reverse iteration", func(t *testing.T) {
		arr := []int{3, 1, 2}
		v := AsValue(arr)

		var items []int
		v.IterateOrder(func(idx, count int, key, value *Value) bool {
			items = append(items, key.Integer())
			return true
		}, func() {}, true, false)

		if items[0] != 2 || items[1] != 1 || items[2] != 3 {
			t.Errorf("Reverse iteration order wrong: %v", items)
		}
	})

	t.Run("map sorted iteration", func(t *testing.T) {
		m := map[string]int{"c": 3, "a": 1, "b": 2}
		v := AsValue(m)

		var keys []string
		v.IterateOrder(func(idx, count int, key, value *Value) bool {
			keys = append(keys, key.String())
			return true
		}, func() {}, false, true)

		if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
			t.Errorf("Map sorted iteration order wrong: %v", keys)
		}
	})

	t.Run("string sorted iteration", func(t *testing.T) {
		str := "cba"
		v := AsValue(str)

		var chars []string
		v.IterateOrder(func(idx, count int, key, value *Value) bool {
			chars = append(chars, key.String())
			return true
		}, func() {}, false, true)

		if chars[0] != "a" || chars[1] != "b" || chars[2] != "c" {
			t.Errorf("String sorted iteration order wrong: %v", chars)
		}
	})
}
