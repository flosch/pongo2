package pongo2

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Value struct {
	v reflect.Value
}

func AsValue(i interface{}) *Value {
	return &Value{
		v: reflect.ValueOf(i),
	}
}

func (v *Value) getResolvedValue() reflect.Value {
	if v.v.IsValid() && v.v.Kind() == reflect.Ptr {
		return v.v.Elem()
	}
	return v.v
}

func (v *Value) IsString() bool {
	return v.getResolvedValue().Kind() == reflect.String
}

func (v *Value) IsFloat() bool {
	return v.getResolvedValue().Kind() == reflect.Float32 ||
		v.getResolvedValue().Kind() == reflect.Float64
}

func (v *Value) IsInteger() bool {
	return v.getResolvedValue().Kind() == reflect.Int ||
		v.getResolvedValue().Kind() == reflect.Int8 ||
		v.getResolvedValue().Kind() == reflect.Int16 ||
		v.getResolvedValue().Kind() == reflect.Int32 ||
		v.getResolvedValue().Kind() == reflect.Int64 ||
		v.getResolvedValue().Kind() == reflect.Uint ||
		v.getResolvedValue().Kind() == reflect.Uint8 ||
		v.getResolvedValue().Kind() == reflect.Uint16 ||
		v.getResolvedValue().Kind() == reflect.Uint32 ||
		v.getResolvedValue().Kind() == reflect.Uint64
}

func (v *Value) IsNumber() bool {
	return v.IsInteger() || v.IsFloat()
}

func (v *Value) IsNil() bool {
	//fmt.Printf("%+v\n", v.getResolvedValue().Type().String())
	return !v.getResolvedValue().IsValid()
}

func (v *Value) String() string {
	switch v.getResolvedValue().Kind() {
	case reflect.String:
		return v.getResolvedValue().String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return strconv.FormatInt(v.getResolvedValue().Int(), 10)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.getResolvedValue().Float())
	case reflect.Bool:
		if v.Bool() {
			return "True"
		} else {
			return "False"
		}
	default:
		if v.IsNil() {
			return ""
		}
		logf("Value.String() not implemented for type: %s\n", v.getResolvedValue().Kind().String())
		return v.getResolvedValue().String()
	}
}

func (v *Value) Integer() int {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return int(v.getResolvedValue().Int())
	case reflect.Float32, reflect.Float64:
		return int(v.getResolvedValue().Float())
	case reflect.String:
		// Try to convert from string to int (base 10)
		i, err := strconv.Atoi(v.getResolvedValue().String())
		if err != nil {
			return 0
		}
		return i
	default:
		logf("Value.Integer() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return 0
	}
}

func (v *Value) Float() float64 {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return float64(v.getResolvedValue().Int())
	case reflect.Float32, reflect.Float64:
		return v.getResolvedValue().Float()
	case reflect.String:
		// Try to convert from string to float64 (base 10)
		f, err := strconv.ParseFloat(v.getResolvedValue().String(), 64)
		if err != nil {
			return 0.0
		}
		return f
	default:
		logf("Value.Float() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return 0.0
	}
}

func (v *Value) Bool() bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Bool:
		return v.getResolvedValue().Bool()
	default:
		logf("Value.Bool() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return false
	}
}

func (v *Value) IsTrue() bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return v.getResolvedValue().Int() != 0
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.getResolvedValue().Len() > 0
	case reflect.Bool:
		return v.getResolvedValue().Bool()
	default:
		logf("Value.IsTrue() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return false
	}
}

func (v *Value) Negate() *Value {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		if v.getResolvedValue().Int() != 0 {
			return AsValue(0)
		} else {
			return AsValue(1)
		}
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return AsValue(v.getResolvedValue().Len() == 0)
	case reflect.Bool:
		return AsValue(!v.getResolvedValue().Bool())
	default:
		logf("Value.IsTrue() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return AsValue(true)
	}
}

func (v *Value) Len() int {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.getResolvedValue().Len()
	default:
		logf("Value.Len() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return 0
	}
}

func (v *Value) Slice(i, j int) *Value {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		return AsValue(v.getResolvedValue().Slice(i, j).Interface())
	default:
		logf("Value.Slice() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return AsValue([]int{})
	}
}

func (v *Value) Contains(other *Value) bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Struct:
		field_value := v.getResolvedValue().FieldByName(other.String())
		return field_value.IsValid()
	case reflect.Map:
		var map_value reflect.Value
		switch other.Interface().(type) {
		case int:
			map_value = v.getResolvedValue().MapIndex(other.getResolvedValue())
		case string:
			map_value = v.getResolvedValue().MapIndex(other.getResolvedValue())
		default:
			fmt.Printf("Value.Contains() does not support lookup type '%s'\n", other.getResolvedValue().Kind().String())
			return false
		}

		return map_value.IsValid()
	case reflect.String:
		return strings.Contains(v.getResolvedValue().String(), other.String())

	// TODO: reflect.Array, reflect.Slice

	default:
		logf("Value.Contains() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return false
	}
}

func (v *Value) CanSlice() bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		return true
	}
	return false
}

func (v *Value) Iterate(fn func(idx, count int, key, value *Value) bool, empty func()) {
	switch v.getResolvedValue().Kind() {
	case reflect.Map:
		keys := v.getResolvedValue().MapKeys()
		keyLen := len(keys)
		for idx, key := range keys {
			value := v.getResolvedValue().MapIndex(key)
			if !fn(idx, keyLen, &Value{key}, &Value{value}) {
				return
			}
		}
		if keyLen == 0 {
			empty()
		}
		return // done
	case reflect.Array, reflect.Slice, reflect.String:
		itemCount := v.getResolvedValue().Len()
		if itemCount > 0 {
			for i := 0; i < itemCount; i++ {
				if !fn(i, itemCount, &Value{v.getResolvedValue().Index(i)}, nil) {
					return
				}
			}
		} else {
			empty()
		}
		return // done
	default:
		logf("Value.Iterate() not available for type: %s\n", v.getResolvedValue().Kind().String())
	}
	empty()
}

func (v *Value) Interface() interface{} {
	if v.v.IsValid() {
		return v.v.Interface()
	}
	return nil
}

func (v *Value) EqualValueTo(other *Value) bool {
	return v.Interface() == other.Interface()
}
