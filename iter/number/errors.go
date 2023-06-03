// Common errors
package number

import "errors"

// For most arithmetic finalizers having an empty iterator is an error case
var ErrEmptyIterator = errors.New("empty iterator")
