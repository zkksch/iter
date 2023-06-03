// Abstractions used in the library
package iter

// Iterator represents iterator interface
// Using separate Next and Get methods because Go doesn't have foreach loops for custom types
type Iterator[T any] interface {
	Next() bool
	Get() (T, error)
}
