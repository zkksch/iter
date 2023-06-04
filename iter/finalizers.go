// Contains finalizers for iterators
package iter

// Function Reduce takes an iterator, initial value and a function,
// after that it initializes accumulating value with init value,
// and for each element in an iterator executes function from element value and
// accumulating value that should return a new accumulating value
func Reduce[T, K any](it Iterator[T], init K, fn func(T, K) K) (K, error) {
	acc := init
	for it.Next() {
		v, err := it.Get()
		if err != nil {
			return init, err
		}
		acc = fn(v, acc)
	}
	return acc, nil
}

// Function ToSlice makes a slice from elements of an iterator
func ToSlice[T any](it Iterator[T]) ([]T, error) {
	slice := make([]T, 0)
	for it.Next() {
		v, err := it.Get()
		if err != nil {
			return nil, err
		}
		slice = append(slice, v)
	}

	return slice, nil
}
