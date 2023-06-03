// Reduce finalizer
package iter

// Reduce takes initial value and a function, and for each element in iterator
// executes fn(el, acc) (acc is value returned by fn in previous iteration and starts as initial value)
// after that it returns last returned value of fn
// You can find examples of Reduce usage in `number.Sum`, `number.Prod`, `iter.Len` etc.
func Reduce[T, K any](it Iterator[T], init K, fn func(T, K) K) (K, error) {
	result := init
	for it.Next() {
		v, err := it.Get()
		if err != nil {
			return result, err
		}
		result = fn(v, result)
	}
	return result, nil
}
