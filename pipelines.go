// Contains pipes to construct pipelines based on iterators
package iter

import (
	"sync"
	"sync/atomic"
)

// Function Filter returns a thread safe filter pipe
// Iterator returns only values for which the filter function returns true
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
// Iterator returns values obtained by applying mapping function
// on elements returned by the source
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
// Iterator will return limited amount of elements from the source
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
// Iterator will return limited amount of elements from the source
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

// Function Pairs combines 2 iterators into one iterator that returns Pair values
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

// Function PairsSafe combines 2 iterators into one thread safe iterator that returns Pair values
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
// into one iterator that returns slices of values combined from all sources
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

// Function CombineSafe combines several same typed iterators
// into one thread safe iterator that returns slices of values combined from all sources
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
