// Abstract layer for library
package iter

// Iterator interface
type Iterator[T any] interface {
	// Method Next switches an iterator to a next value
	// and returns it, also returns error if occured
	// (including stop iteration error)
	Next() (T, error)
}
