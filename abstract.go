// Abstract layer for library
package iter

// Iterator interface
type Iterator[T any] func() (T, error)
