// Abstract layer for library
package iter

// Iterator interface
type Iterator[T any] interface {
	// Method Next switches an iterator to a next value
	// also returns whether an iterator has stopped or not
	Next() bool
	// Method Get returns current value of iterator and error
	// if value is not available (for example if iterator has stopped)
	Get() (T, error)
}
