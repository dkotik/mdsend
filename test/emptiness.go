package test

import (
	"iter"
	"testing"
)

func IteratorIsEmpty[T any](
	iterator iter.Seq2[T, error],
) func(*testing.T) {
	return func(t *testing.T) {
		for _, err := range iterator {
			if err != nil {
				t.Fatal(err)
			}
			t.Fatal("found one item")
		}
	}
}
