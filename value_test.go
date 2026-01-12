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

func TestValueIsSliceOrArray(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		// Slices
		{"string slice", []string{"a", "b", "c"}, true},
		{"int slice", []int{1, 2, 3}, true},
		{"empty slice", []string{}, true},
		{"interface slice", []any{"a", 1, true}, true},
		{"byte slice", []byte{1, 2, 3}, true},

		// Arrays
		{"string array", [3]string{"a", "b", "c"}, true},
		{"int array", [3]int{1, 2, 3}, true},
		{"empty array", [0]int{}, true},

		// Pointers to slices/arrays
		{"pointer to slice", &[]string{"a", "b"}, true},
		{"pointer to array", &[2]int{1, 2}, true},

		// Non-slice/array types
		{"string", "not a slice", false},
		{"empty string", "", false},
		{"integer", 42, false},
		{"float", 3.14, false},
		{"bool", true, false},
		{"nil", nil, false},
		{"map", map[string]int{"a": 1}, false},
		{"struct", struct{ Name string }{"test"}, false},

		// Edge cases
		{"pointer to string", func() any { s := "test"; return &s }(), false},
		{"pointer to int", func() any { i := 42; return &i }(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := AsValue(tt.input)
			result := v.IsSliceOrArray()
			if result != tt.expected {
				t.Errorf("IsSliceOrArray() = %v, want %v for input %T(%v)", result, tt.expected, tt.input, tt.input)
			}
		})
	}
}

func TestValueIsMap(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		// Maps
		{"string to int map", map[string]int{"a": 1}, true},
		{"int to string map", map[int]string{1: "one"}, true},
		{"empty map", map[string]int{}, true},
		{"interface map", map[string]any{"a": 1}, true},

		// Pointers to maps
		{"pointer to map", &map[string]int{"a": 1}, true},

		// Non-map types
		{"string", "not a map", false},
		{"integer", 42, false},
		{"slice", []int{1, 2, 3}, false},
		{"array", [3]int{1, 2, 3}, false},
		{"struct", struct{ Name string }{"test"}, false},
		{"nil", nil, false},
		{"bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := AsValue(tt.input)
			result := v.IsMap()
			if result != tt.expected {
				t.Errorf("IsMap() = %v, want %v for input %T(%v)", result, tt.expected, tt.input, tt.input)
			}
		})
	}
}

func TestValueIsStruct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		// Structs
		{"named struct", TestStruct{Name: "test", Age: 25}, true},
		{"anonymous struct", struct{ Name string }{"test"}, true},
		{"empty struct", struct{}{}, true},

		// Pointers to structs
		{"pointer to struct", &TestStruct{Name: "test"}, true},

		// Non-struct types
		{"string", "not a struct", false},
		{"integer", 42, false},
		{"slice", []int{1, 2, 3}, false},
		{"map", map[string]int{"a": 1}, false},
		{"nil", nil, false},
		{"bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := AsValue(tt.input)
			result := v.IsStruct()
			if result != tt.expected {
				t.Errorf("IsStruct() = %v, want %v for input %T(%v)", result, tt.expected, tt.input, tt.input)
			}
		})
	}
}

func TestValueGetItem(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	t.Run("map with string key", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		v := AsValue(m)

		result := v.GetItem(AsValue("b"))
		if result.IsNil() {
			t.Error("GetItem should return a value for existing key")
		}
		if result.Integer() != 2 {
			t.Errorf("GetItem(\"b\") = %d, want 2", result.Integer())
		}
	})

	t.Run("map with non-existent key", func(t *testing.T) {
		m := map[string]int{"a": 1}
		v := AsValue(m)

		result := v.GetItem(AsValue("z"))
		if !result.IsNil() {
			t.Error("GetItem should return nil for non-existent key")
		}
	})

	t.Run("map with int key", func(t *testing.T) {
		m := map[int]string{1: "one", 2: "two", 3: "three"}
		v := AsValue(m)

		result := v.GetItem(AsValue(2))
		if result.IsNil() {
			t.Error("GetItem should return a value for existing int key")
		}
		if result.String() != "two" {
			t.Errorf("GetItem(2) = %q, want %q", result.String(), "two")
		}
	})

	t.Run("struct field access", func(t *testing.T) {
		p := Person{Name: "Alice", Age: 30}
		v := AsValue(p)

		nameResult := v.GetItem(AsValue("Name"))
		if nameResult.IsNil() {
			t.Error("GetItem should return value for existing field")
		}
		if nameResult.String() != "Alice" {
			t.Errorf("GetItem(\"Name\") = %q, want %q", nameResult.String(), "Alice")
		}

		ageResult := v.GetItem(AsValue("Age"))
		if ageResult.Integer() != 30 {
			t.Errorf("GetItem(\"Age\") = %d, want 30", ageResult.Integer())
		}
	})

	t.Run("struct non-existent field", func(t *testing.T) {
		p := Person{Name: "Alice", Age: 30}
		v := AsValue(p)

		result := v.GetItem(AsValue("NonExistent"))
		if !result.IsNil() {
			t.Error("GetItem should return nil for non-existent field")
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		p := &Person{Name: "Bob", Age: 25}
		v := AsValue(p)

		result := v.GetItem(AsValue("Name"))
		if result.IsNil() {
			t.Error("GetItem should work with pointer to struct")
		}
		if result.String() != "Bob" {
			t.Errorf("GetItem(\"Name\") = %q, want %q", result.String(), "Bob")
		}
	})

	t.Run("pointer to map", func(t *testing.T) {
		m := &map[string]int{"x": 10}
		v := AsValue(m)

		result := v.GetItem(AsValue("x"))
		if result.IsNil() {
			t.Error("GetItem should work with pointer to map")
		}
		if result.Integer() != 10 {
			t.Errorf("GetItem(\"x\") = %d, want 10", result.Integer())
		}
	})

	t.Run("nil key", func(t *testing.T) {
		m := map[string]int{"a": 1}
		v := AsValue(m)

		result := v.GetItem(AsValue(nil))
		if !result.IsNil() {
			t.Error("GetItem with nil key should return nil")
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		v := AsValue("not a map or struct")
		result := v.GetItem(AsValue("key"))
		if !result.IsNil() {
			t.Error("GetItem on unsupported type should return nil")
		}
	})

	t.Run("slice type", func(t *testing.T) {
		v := AsValue([]int{1, 2, 3})
		result := v.GetItem(AsValue(0))
		if !result.IsNil() {
			t.Error("GetItem on slice should return nil (use Index instead)")
		}
	})

	t.Run("integer type", func(t *testing.T) {
		v := AsValue(42)
		result := v.GetItem(AsValue("key"))
		if !result.IsNil() {
			t.Error("GetItem on integer should return nil")
		}
	})

	t.Run("map with float64 key - convertible", func(t *testing.T) {
		m := map[float64]string{1.5: "one-half", 2.5: "two-half"}
		v := AsValue(m)

		// Float key access - should work with direct conversion
		result := v.GetItem(AsValue(1.5))
		if result.IsNil() {
			t.Error("GetItem should return value for existing float64 key")
		}
		if result.String() != "one-half" {
			t.Errorf("GetItem(1.5) = %q, want %q", result.String(), "one-half")
		}
	})

	t.Run("map with float64 key - non-existent", func(t *testing.T) {
		m := map[float64]string{1.5: "one-half"}
		v := AsValue(m)

		result := v.GetItem(AsValue(9.9))
		if !result.IsNil() {
			t.Error("GetItem should return nil for non-existent float64 key")
		}
	})

	t.Run("map with bool key - convertible", func(t *testing.T) {
		m := map[bool]string{true: "yes", false: "no"}
		v := AsValue(m)

		result := v.GetItem(AsValue(true))
		if result.IsNil() {
			t.Error("GetItem should return value for existing bool key")
		}
		if result.String() != "yes" {
			t.Errorf("GetItem(true) = %q, want %q", result.String(), "yes")
		}
	})

	t.Run("map with uint key", func(t *testing.T) {
		m := map[uint]string{1: "one", 2: "two"}
		v := AsValue(m)

		// uint keys are handled by the default case with direct conversion
		result := v.GetItem(AsValue(uint(1)))
		if result.IsNil() {
			t.Error("GetItem should return value for existing uint key")
		}
		if result.String() != "one" {
			t.Errorf("GetItem(uint(1)) = %q, want %q", result.String(), "one")
		}
	})

	t.Run("map with incompatible key type", func(t *testing.T) {
		// Create a map with a complex key type that can't be converted from string/int
		type customKey struct {
			id int
		}
		m := map[customKey]string{{id: 1}: "value"}
		v := AsValue(m)

		// String key can't be converted to customKey
		result := v.GetItem(AsValue("key"))
		if !result.IsNil() {
			t.Error("GetItem should return nil when key can't be converted")
		}
	})
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
