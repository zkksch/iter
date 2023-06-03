// Map pipe
package iter

type mapIterator[T, K any] struct {
	base Iterator[T]
	fn   func(T) (K, error)
}

func (it *mapIterator[T, K]) Next() bool {
	return it.base.Next()
}

func (it *mapIterator[T, K]) Get() (K, error) {
	var value K
	original, err := it.base.Get()
	if err != nil {
		return value, err
	}
	value, err = it.fn(original)
	return value, err
}

// Returns Map pipe
// For each element in iterator executes fn and includes result in resulting iterator
func Map[T, K any](i Iterator[T], fn func(T) (K, error)) Iterator[K] {
	return &mapIterator[T, K]{
		base: i,
		fn:   fn,
	}
}
