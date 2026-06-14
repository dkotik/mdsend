package queue

import (
	"context"
	"iter"
)

func CollectMostOf[T any](ctx context.Context, count int) func(iter.Seq2[T, error]) iter.Seq2[T, error] {
	return func(in iter.Seq2[T, error]) iter.Seq2[T, error] {
		return func(yield func(item T, err error) bool) {
			limit := count
			for item, err := range in {
				if !yield(item, err) {
					return
				}
				limit--
				if limit == 0 {
					return
				}
			}
		}
	}
}
