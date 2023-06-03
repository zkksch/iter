// Aggregate finalizers for number types
package number

import (
	"sort"

	"github.com/zkksch/iter/iter"
)

// Average finalizer calculates average value for given iterator
func Average(it iter.Iterator[float64]) (float64, error) {
	var total, sum float64
	for it.Next() {
		v, err := it.Get()
		if err != nil {
			return 0, err
		}
		total++
		sum += v
	}

	if total == 0 {
		return 0, ErrEmptyIterator
	}

	return sum / total, nil
}

// Mediana finalizer calculates mediana value for given iterator
func Mediana(it iter.Iterator[float64]) (float64, error) {
	var result float64

	slice, err := iter.Finalize[float64](it)
	if err != nil {
		return result, err
	}

	sort.Float64s(slice)

	l := len(slice)

	if l == 0 {
		return result, ErrEmptyIterator
	}

	if l == 1 {
		return slice[0], nil
	}

	if l%2 == 0 {
		result = (slice[l/2] + slice[l/2-1]) / 2
	} else {
		result = slice[l/2]
	}

	return result, nil
}

// Sum finalizer returns sum of elements in iterator
func Sum[T Number](it iter.Iterator[T]) (T, error) {
	return iter.Reduce[T, T](it, 0, func(a, b T) T {
		return a + b
	})
}

// Prod finalizer returns production of elements in iterator
func Prod[T Number](it iter.Iterator[T]) (T, error) {
	return iter.Reduce[T, T](it, 1, func(a, b T) T {
		return a * b
	})
}
