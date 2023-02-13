package tools

import "fmt"

func NilPanic(args ...interface{}) {
	for i, v := range args {
		if v == nil {
			panic(
				fmt.Errorf("nil arg #%d", i),
			)
		}
	}
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func Pointer[T any](v T) *T {
	return &v
}
