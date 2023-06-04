// Contains pipes to construct pipelines based on iterators
package iter

// Filter iterator implementation
type filterIterator[T any] struct {
	base Iterator[T]
	fn   func(T) bool
}

func (it *filterIterator[T]) Next() (T, error) {
	var (
		v   T
		err error
	)
	for v, err = it.base.Next(); err == nil; v, err = it.base.Next() {
		if it.fn(v) {
			return v, nil
		}
	}
	return v, err
}

// Function Filter returns filter pipe
// For each element in a given iterator executes filter function and
// if a return value is true that element will be included in a resulting iterator
func Filter[T any](it Iterator[T], fn func(T) bool) Iterator[T] {
	return &filterIterator[T]{
		base: it,
		fn:   fn,
	}
}

// Map iterator implementation
type mapIterator[T, K any] struct {
	base Iterator[T]
	fn   func(T) (K, error)
}

func (it *mapIterator[T, K]) Next() (K, error) {
	original, err := it.base.Next()
	if err != nil {
		var empty K
		return empty, err
	}
	return it.fn(original)
}

// Function Map returns map pipe
// For each element in iterator executes mapping function
// and includes result of that function in a resulting iterator
func Map[T, K any](it Iterator[T], fn func(T) (K, error)) Iterator[K] {
	return &mapIterator[T, K]{
		base: it,
		fn:   fn,
	}
}

// Limit iterator implementation
type limitIterator[T any] struct {
	base   Iterator[T]
	remain int
}

func (it *limitIterator[T]) Next() (T, error) {
	if it.remain <= 0 {
		var empty T
		return empty, ErrStopIt
	}
	next, err := it.base.Next()
	if err == nil {
		it.remain--
	} else {
		it.remain = 0
	}
	return next, err
}

// Function Limit returns limit pipe
// Accepts limit number as a parament and
// only includes n <= limit elements in a resulting iterator
func Limit[T any](it Iterator[T], limit int) Iterator[T] {
	return &limitIterator[T]{
		base:   it,
		remain: limit,
	}
}
