// Filter pipe
package iter

type filterIterator[T any] struct {
	base Iterator[T]
	fn   func(T) bool
}

func (it *filterIterator[T]) Next() bool {
	for it.base.Next() {
		value, err := it.base.Get()
		if err != nil {
			return false
		}
		if it.fn(value) {
			return true
		}
	}
	return false
}

func (it *filterIterator[T]) Get() (T, error) {
	return it.base.Get()
}

// Returns Filter pipe
// For each element in given iterator executes fn and if a return value
// is true that element will be included in resulting iterator
func Filter[T any](i Iterator[T], fn func(T) bool) Iterator[T] {
	return &filterIterator[T]{
		base: i,
		fn:   fn,
	}
}
