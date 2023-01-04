package pongo2

import (
	"fmt"
)

// Range returns a list of integers from 1 till given input
func Range(in *Value, param *Value) (out *Value, err *Error) {
	if !in.IsInteger() {
		return nil, &Error{
			OrigError: fmt.Errorf("only strings should be sent to the scream filter"),
		}
	}

	n := in.Integer()
	l := []int{}
	for i := 1; i < n+1; i++ {
		l = append(l, i)
	}

	return AsValue(l), nil
}
