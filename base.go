// Base types
package iter

import "errors"

// Iterator is just a function that will return next value or error
type Iterator[T any] func() (T, error)

// Default error to signal that iterator has no elements to return
var ErrStopIt = errors.New("stop iteration")
