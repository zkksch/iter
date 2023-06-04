// Contains pipes to construct pipelines based on iterators
package iter

// Filter iterator implementation
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

func (it *mapIterator[T, K]) Next() bool {
	return it.base.Next()
}

func (it *mapIterator[T, K]) Get() (K, error) {
	original, err := it.base.Get()
	if err != nil {
		var value K
		return value, err
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

func (it *limitIterator[T]) Next() bool {
	if it.remain <= 0 {
		return false
	}
	next := it.base.Next()
	if next {
		it.remain--
	}
	return next
}

func (it *limitIterator[T]) Get() (T, error) {
	return it.base.Get()
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

type Pair[T, K any] struct {
	Left  T
	Right K
}

type pairIterator[T, K any] struct {
	left    Iterator[T]
	right   Iterator[K]
	stopped bool
}

func (it *pairIterator[T, K]) Next() bool {
	if it.stopped {
		return false
	}
	left := it.left.Next()
	right := it.right.Next()
	next := left && right
	it.stopped = !next
	return next
}

func (it *pairIterator[T, K]) Get() (Pair[T, K], error) {
	value := Pair[T, K]{}
	if it.stopped {
		return value, ErrStopIt
	}

	left, err := it.left.Get()
	if err != nil {
		return value, err
	}

	right, err := it.right.Get()
	if err != nil {
		return value, err
	}

	value.Left = left
	value.Right = right

	return value, nil
}

func Pairs[T, K any](left Iterator[T], right Iterator[K]) Iterator[Pair[T, K]] {
	return &pairIterator[T, K]{
		left:  left,
		right: right,
	}
}
