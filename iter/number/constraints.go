package number

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Float | constraints.Integer
}
