// Aggregate finalizers
package iter

// Len returns number of elements in iterator
func Len[T any](it Iterator[T]) (int, error) {
	return Reduce[T, int](it, 0, func(_ T, i int) int {
		return i + 1
	})
}
