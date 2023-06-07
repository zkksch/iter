// Contains pipes to construct pipelines based on iterators
package iter

import (
	"sync"
	"sync/atomic"
)

// Filter iterator implementation (thread safe)
type filterIterator[T any] struct {
	source Iterator[T]
	fn     func(T) bool
}

func (it *filterIterator[T]) Next() (T, error) {
	var (
		v   T
		err error
	)
	for v, err = it.source.Next(); err == nil; v, err = it.source.Next() {
		if it.fn(v) {
			return v, nil
		}
	}
	return v, err
}

// Function Filter returns a thread safe filter pipe
// For each element in a given iterator executes filter function and
// if a return value is true that element will be included in a resulting iterator
func Filter[T any](it Iterator[T], fn func(T) bool) Iterator[T] {
	return &filterIterator[T]{
		source: it,
		fn:     fn,
	}
}

// Map iterator implementation (thread safe)
type mapIterator[T, K any] struct {
	source Iterator[T]
	fn     func(T) (K, error)
}

func (it *mapIterator[T, K]) Next() (K, error) {
	original, err := it.source.Next()
	if err != nil {
		var empty K
		return empty, err
	}
	return it.fn(original)
}

// Function Map returns a thread safe map pipe
// For each element in iterator executes mapping function
// and includes result of that function in a resulting iterator
func Map[T, K any](it Iterator[T], fn func(T) (K, error)) Iterator[K] {
	return &mapIterator[T, K]{
		source: it,
		fn:     fn,
	}
}

// Limit iterator implementation
type limitIterator[T any] struct {
	source Iterator[T]
	remain int
}

func (it *limitIterator[T]) Next() (T, error) {
	if it.remain <= 0 {
		var empty T
		return empty, ErrStopIt
	}
	next, err := it.source.Next()
	if err == nil {
		it.remain--
	} else {
		it.remain = 0
	}
	return next, err
}

// Limit iterator implementation (thread safe)
type safeLimitIterator[T any] struct {
	source  Iterator[T]
	remain  *atomic.Int64
	stopped *atomic.Bool
}

func (it *safeLimitIterator[T]) Next() (T, error) {
	stop := it.stopped.Load()
	if stop {
		var empty T
		return empty, ErrStopIt
	}

	remain := it.remain.Add(-1)
	if remain < 0 {
		var empty T
		it.stopped.Store(true)
		return empty, ErrStopIt
	}
	next, err := it.source.Next()
	if err != nil {
		it.stopped.Store(true)
	}
	return next, err
}

// Function Limit returns a limit pipe
// Accepts limit number as a parametr and
// only includes n <= limit elements in a resulting iterator
func Limit[T any](it Iterator[T], limit int) Iterator[T] {
	return &limitIterator[T]{
		source: it,
		remain: limit,
	}
}

// Function Limit returns a thread safe limit pipe
// Accepts limit number as a parametr and
// only includes n <= limit elements in a resulting iterator
func LimitSafe[T any](it Iterator[T], limit int) Iterator[T] {
	v := &atomic.Int64{}
	v.Add(int64(limit))
	return &safeLimitIterator[T]{
		source:  it,
		remain:  v,
		stopped: &atomic.Bool{},
	}
}

// Pair of values
type Pair[T, K any] struct {
	Left  T
	Right K
}

// Pairs iterator implementation
type pairsIterator[T, K any] struct {
	left  Iterator[T]
	right Iterator[K]
}

func (it *pairsIterator[T, K]) Next() (Pair[T, K], error) {
	left, err := it.left.Next()
	if err != nil {
		return Pair[T, K]{}, err
	}
	right, err := it.right.Next()
	if err != nil {
		return Pair[T, K]{}, err
	}
	return Pair[T, K]{
		Left:  left,
		Right: right,
	}, nil
}

// Pairs iterator implementation (thread safe)
type safePairsIterator[T, K any] struct {
	sync.Mutex
	left  Iterator[T]
	right Iterator[K]
}

func (it *safePairsIterator[T, K]) Next() (Pair[T, K], error) {
	it.Lock()
	defer it.Unlock()
	left, err := it.left.Next()
	if err != nil {
		return Pair[T, K]{}, err
	}
	right, err := it.right.Next()
	if err != nil {
		return Pair[T, K]{}, err
	}
	return Pair[T, K]{
		Left:  left,
		Right: right,
	}, nil
}

// Function Pairs combines 2 iterators into one iterator that will provide Pair values
func Pairs[T, K any](left Iterator[T], right Iterator[K]) Iterator[Pair[T, K]] {
	return &pairsIterator[T, K]{
		left:  left,
		right: right,
	}
}

// Function Pairs combines 2 iterators into one thread safe iterator that will provide Pair values
func PairsSafe[T, K any](left Iterator[T], right Iterator[K]) Iterator[Pair[T, K]] {
	return &safePairsIterator[T, K]{
		left:  left,
		right: right,
	}
}

// Combine iterator implementation
type combineIterator[T any] struct {
	sources []Iterator[T]
}

func (it *combineIterator[T]) Next() ([]T, error) {
	values := make([]T, 0, len(it.sources))
	for _, base := range it.sources {
		v, err := base.Next()
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

// Combine iterator implementation (thread safe)
type safeCombineIterator[T any] struct {
	sync.Mutex
	sources []Iterator[T]
}

func (it *safeCombineIterator[T]) Next() ([]T, error) {
	it.Lock()
	defer it.Unlock()
	values := make([]T, 0, len(it.sources))
	for _, base := range it.sources {
		v, err := base.Next()
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

// Function Combine combines several same typed iterators
// into one iterator that will provide slices as values
func Combine[T any](iterators ...Iterator[T]) Iterator[[]T] {
	return &combineIterator[T]{
		sources: iterators,
	}
}

// Function Combine combines several same typed iterators
// into one thread safe iterator that will provide slices as values
func CombineSafe[T any](iterators ...Iterator[T]) Iterator[[]T] {
	return &safeCombineIterator[T]{
		sources: iterators,
	}
}
