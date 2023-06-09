// Contains pipes to construct pipelines based on iterators
package iter

import (
	"sync"
	"sync/atomic"
)

// Function Filter returns a thread safe filter pipe
// For each element in a given iterator executes filter function and
// if a return value is true that element will be included in a resulting iterator
func Filter[T any](source Iterator[T], filter func(T) bool) Iterator[T] {
	return func() (T, error) {
		var (
			v   T
			err error
		)
		for v, err = source(); err == nil; v, err = source() {
			if filter(v) {
				return v, nil
			}
		}
		return v, err
	}
}

// Function Map returns a thread safe map pipe
// For each element in iterator executes mapping function
// and includes result of that function in a resulting iterator
func Map[T, K any](source Iterator[T], mapping func(T) (K, error)) Iterator[K] {
	return func() (K, error) {
		v, err := source()
		if err != nil {
			var empty K
			return empty, err
		}
		return mapping(v)
	}
}

// Function Limit returns a limit pipe
// Accepts limit number as a parametr and
// only includes n <= limit elements in a resulting iterator
func Limit[T any](source Iterator[T], limit int) Iterator[T] {
	remain := limit
	return func() (T, error) {
		if remain <= 0 {
			var empty T
			return empty, ErrStopIt
		}
		v, err := source()
		if err == nil {
			remain--
		} else {
			remain = 0
		}
		return v, err
	}
}

// Function Limit returns a thread safe limit pipe
// Accepts limit number as a parametr and
// only includes n <= limit elements in a resulting iterator
func LimitSafe[T any](source Iterator[T], limit int) Iterator[T] {
	remain := &atomic.Int64{}
	remain.Add(int64(limit))
	return func() (T, error) {
		remainLocal := remain.Add(-1)
		if remainLocal < 0 {
			var empty T
			remain.Store(0)
			return empty, ErrStopIt
		}
		v, err := source()
		if err != nil {
			remain.Store(0)
		}
		return v, err
	}
}

// Pair of values
type Pair[T, K any] struct {
	Left  T
	Right K
}

// Function Pairs combines 2 iterators into one iterator that will provide Pair values
func Pairs[T, K any](left Iterator[T], right Iterator[K]) Iterator[Pair[T, K]] {
	return func() (Pair[T, K], error) {
		leftEl, err := left()
		if err != nil {
			return Pair[T, K]{}, err
		}
		rightEl, err := right()
		if err != nil {
			return Pair[T, K]{}, err
		}
		return Pair[T, K]{
			Left:  leftEl,
			Right: rightEl,
		}, nil
	}
}

// Function Pairs combines 2 iterators into one thread safe iterator that will provide Pair values
func PairsSafe[T, K any](left Iterator[T], right Iterator[K]) Iterator[Pair[T, K]] {
	var mutex sync.Mutex
	return func() (Pair[T, K], error) {
		mutex.Lock()
		defer mutex.Unlock()
		leftEl, err := left()
		if err != nil {
			return Pair[T, K]{}, err
		}
		rightEl, err := right()
		if err != nil {
			return Pair[T, K]{}, err
		}
		return Pair[T, K]{
			Left:  leftEl,
			Right: rightEl,
		}, nil
	}
}

// Function Combine combines several same typed iterators
// into one iterator that will provide slices as values
func Combine[T any](iterators ...Iterator[T]) Iterator[[]T] {
	return func() ([]T, error) {
		values := make([]T, 0, len(iterators))
		if len(iterators) == 0 {
			return values, ErrStopIt
		}
		for _, source := range iterators {
			v, err := source()
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		return values, nil
	}
}

// Function Combine combines several same typed iterators
// into one thread safe iterator that will provide slices as values
func CombineSafe[T any](iterators ...Iterator[T]) Iterator[[]T] {
	var mutex sync.Mutex
	return func() ([]T, error) {
		mutex.Lock()
		defer mutex.Unlock()
		values := make([]T, 0, len(iterators))
		if len(iterators) == 0 {
			return values, ErrStopIt
		}
		for _, source := range iterators {
			v, err := source()
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		return values, nil
	}
}
