// Unit tests for package
package tests

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/zkksch/iter"
)

// Check elements of a given iterator
func validateResult[T any](t *testing.T, idx int, it iter.Iterator[T], expected []T) {
	result, err := iter.ToSlice(it)
	if err != nil {
		t.Fatalf("[%v] %v", idx, err.Error())
	}
	if len(result) != len(expected) {
		t.Fatalf(
			"[%v] number of elements in the iterator doesn't match expected result:\n%v != %v\n%v != %v\n",
			idx, len(result), len(expected), result, expected)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf(
			"[%v] iterator elements are not equal to expected values:\n%v != %v\n",
			idx, result, expected)
	}
}

// Tests slice iterator and ToSlice finalizer (both safe and unsafe)
// For each slice creates iterator and finalizes it into slice, resulting slice should be the same
func TestFromToSlice(t *testing.T) {
	var empty []int
	cases := []struct {
		source   []int
		expected []int
	}{
		{
			source:   []int{2, 3, 1},
			expected: []int{2, 3, 1},
		},
		{
			source:   []int{},
			expected: []int{},
		},
		{
			source:   empty,
			expected: []int{},
		},
	}

	for i, testCase := range cases {
		it := iter.FromSlice(testCase.source)
		validateResult(t, i, it, testCase.expected)
	}

	for i, testCase := range cases {
		it := iter.FromSliceSafe(testCase.source)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests sequence iterator (both safe and unsafe)
// Start from start value and makes steps by adding step value to it
// Infinite so should handle MaxInt and MinInt
func TestSequence(t *testing.T) {
	cases := []struct {
		start    int
		step     int
		expected []int
		// Required to finalize results
		limit int
	}{
		{
			start:    0,
			step:     1,
			expected: []int{0, 1, 2, 3, 4, 5},
			limit:    6,
		},
		{
			start:    0,
			step:     -1,
			expected: []int{0, -1, -2, -3, -4, -5},
			limit:    6,
		},
		{
			start:    42,
			step:     4,
			expected: []int{42, 46, 50},
			limit:    3,
		},
		{
			start:    42,
			step:     -4,
			expected: []int{42, 38, 34},
			limit:    3,
		},
		{
			start:    math.MaxInt,
			step:     1,
			expected: []int{math.MaxInt, math.MinInt, math.MinInt + 1},
			limit:    3,
		},
		{
			start:    math.MinInt,
			step:     -1,
			expected: []int{math.MinInt, math.MaxInt, math.MaxInt - 1},
			limit:    3,
		},
		{
			start:    1,
			step:     0,
			expected: []int{1, 1, 1},
			limit:    3,
		},
	}

	for i, testCase := range cases {
		it := iter.Sequence(testCase.start, testCase.step)
		it = iter.Limit(it, testCase.limit)
		validateResult(t, i, it, testCase.expected)
	}

	for i, testCase := range cases {
		it := iter.SequenceSafe(testCase.start, testCase.step)
		it = iter.Limit(it, testCase.limit)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests generating iterator
// Takes generating function and returns new values from calling that function
func TestGenerator(t *testing.T) {
	cases := []struct {
		fn       func() func() int
		expected []int
		// Required to finalize results
		limit int
	}{
		{
			// Simple case
			fn: func() func() int {
				return func() int {
					return 0
				}
			},
			expected: []int{0, 0, 0, 0},
			limit:    4,
		},
		{
			// Generating function with closure
			fn: func() func() int {
				value := 1
				return func() int {
					value *= 2
					return value
				}
			},
			expected: []int{2, 4, 8, 16},
			limit:    4,
		},
	}

	for i, testCase := range cases {
		it := iter.Generator(testCase.fn())
		it = iter.Limit(it, testCase.limit)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests repeating iterator
// Repeats the same value indefinitely
func TestRepeat(t *testing.T) {
	cases := []struct {
		value    int
		expected []int
		limit    int
	}{
		{
			value:    0,
			expected: []int{0, 0, 0, 0},
			limit:    4,
		},
		{
			value:    1,
			expected: []int{1, 1, 1, 1},
			limit:    4,
		},
	}

	for i, testCase := range cases {
		it := iter.Repeat(testCase.value)
		it = iter.Limit(it, testCase.limit)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests filter pipe
// Filters values by using a filter function
func TestFilter(t *testing.T) {
	cases := []struct {
		source   []int
		filter   func(int) bool
		expected []int
	}{
		{
			source: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			filter: func(int) bool {
				return true
			},
			expected: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			source: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			filter: func(int) bool {
				return false
			},
			expected: []int{},
		},
		{
			source: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			filter: func(el int) bool {
				return el%2 == 0
			},
			expected: []int{0, 2, 4, 6, 8},
		},
	}

	for i, testCase := range cases {
		it := iter.FromSlice(testCase.source)
		it = iter.Filter(it, testCase.filter)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests map pipe
// Maps one values to another by using mapping function
func TestMap(t *testing.T) {
	cases := []struct {
		source   []int
		mapping  func(int) (string, error)
		expected []string
	}{
		{
			source: []int{0, 1, 2, 3, 4},
			mapping: func(el int) (string, error) {
				return fmt.Sprint(el), nil
			},
			expected: []string{"0", "1", "2", "3", "4"},
		},
		{
			source: []int{0, 1, 2, 3, 4},
			mapping: func(el int) (string, error) {
				return fmt.Sprint(el * 2), nil
			},
			expected: []string{"0", "2", "4", "6", "8"},
		},
		{
			// Checking returning errors from mapping function
			// ErrStopIt means that iterator should stop
			source: []int{0, 1, 2, 3, 4},
			mapping: func(el int) (string, error) {
				if el == 2 {
					return "", iter.ErrStopIt
				}
				return fmt.Sprint(el), nil
			},
			expected: []string{"0", "1"},
		},
	}

	for i, testCase := range cases {
		src := iter.FromSlice(testCase.source)
		it := iter.Map(src, testCase.mapping)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests limit pipe (both safe and unsafe)
// Limits amount of output elements from iterator
// Infinite cases are checked in generators tests
func TestLimit(t *testing.T) {
	cases := []struct {
		source   []int
		limit    int
		expected []int
	}{
		{
			// Limit higher than amount of elements ignored
			source:   []int{},
			limit:    5,
			expected: []int{},
		},
		{
			source:   []int{1, 2, 3, 4},
			limit:    0,
			expected: []int{},
		},
		{
			// Negative values should be treated the same as 0
			source:   []int{1, 2, 3, 4},
			limit:    -1,
			expected: []int{},
		},
		{
			source:   []int{1, 2, 3, 4},
			limit:    2,
			expected: []int{1, 2},
		},
		{
			source:   []int{1, 2, 3, 4},
			limit:    4,
			expected: []int{1, 2, 3, 4},
		},
		{
			source:   []int{1, 2, 3, 4},
			limit:    5,
			expected: []int{1, 2, 3, 4},
		},
	}

	for i, testCase := range cases {
		it := iter.FromSlice(testCase.source)
		it = iter.Limit(it, testCase.limit)
		validateResult(t, i, it, testCase.expected)
	}

	for i, testCase := range cases {
		it := iter.FromSlice(testCase.source)
		it = iter.LimitSafe(it, testCase.limit)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests pairs pipe (both safe and unsafe)
// Returns values from both iterators in pairs
func TestPairs(t *testing.T) {
	cases := []struct {
		lftSrc   []int
		rgtSrc   []int
		expected []iter.Pair[int, int]
	}{
		{
			lftSrc:   []int{},
			rgtSrc:   []int{},
			expected: []iter.Pair[int, int]{},
		},
		{
			// Always have the smallest amount of elements
			lftSrc:   []int{1, 2, 3, 4},
			rgtSrc:   []int{},
			expected: []iter.Pair[int, int]{},
		},
		{
			lftSrc:   []int{},
			rgtSrc:   []int{1, 2, 3, 4},
			expected: []iter.Pair[int, int]{},
		},
		{
			lftSrc: []int{1, 2, 3, 4},
			rgtSrc: []int{4, 3, 2, 1},
			expected: []iter.Pair[int, int]{
				{Left: 1, Right: 4},
				{Left: 2, Right: 3},
				{Left: 3, Right: 2},
				{Left: 4, Right: 1},
			},
		},
		{
			lftSrc: []int{1, 2, 3, 4, 5},
			rgtSrc: []int{5, 4, 3, 2},
			expected: []iter.Pair[int, int]{
				{Left: 1, Right: 5},
				{Left: 2, Right: 4},
				{Left: 3, Right: 3},
				{Left: 4, Right: 2},
			},
		},
	}

	for i, testCase := range cases {
		left := iter.FromSlice(testCase.lftSrc)
		right := iter.FromSlice(testCase.rgtSrc)
		it := iter.Pairs(left, right)
		validateResult(t, i, it, testCase.expected)
	}

	for i, testCase := range cases {
		left := iter.FromSlice(testCase.lftSrc)
		right := iter.FromSlice(testCase.rgtSrc)
		it := iter.PairsSafe(left, right)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests combine pipe (both safe and unsafe)
// Returns values from all iterators in slices
func TestCombine(t *testing.T) {
	cases := []struct {
		sources  [][]int
		expected [][]int
	}{
		{
			// Empty list of sources should stop iteration immediately
			sources:  [][]int{},
			expected: [][]int{},
		},
		{
			sources: [][]int{
				{},
				{},
				{},
			},
			expected: [][]int{},
		},
		{
			// Always have the smallest amount of elements
			sources: [][]int{
				{1, 2},
				{1},
				{},
			},
			expected: [][]int{},
		},
		{
			sources: [][]int{
				{1, 2, 3},
				{1, 2},
				{1},
			},
			expected: [][]int{
				{1, 1, 1},
			},
		},
		{
			sources: [][]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			expected: [][]int{
				{1, 4, 7},
				{2, 5, 8},
				{3, 6, 9},
			},
		},
		{
			sources: [][]int{
				{1, 2, 3, 4, 5},
				{6, 7, 8, 9},
				{10, 11, 12},
			},
			expected: [][]int{
				{1, 6, 10},
				{2, 7, 11},
				{3, 8, 12},
			},
		},
	}

	for i, testCase := range cases {
		sources := make([]iter.Iterator[int], len(testCase.sources))
		for j, src := range testCase.sources {
			sources[j] = iter.FromSlice(src)
		}
		it := iter.Combine(sources...)
		validateResult(t, i, it, testCase.expected)
	}

	for i, testCase := range cases {
		sources := make([]iter.Iterator[int], len(testCase.sources))
		for j, src := range testCase.sources {
			sources[j] = iter.FromSlice(src)
		}
		it := iter.CombineSafe(sources...)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests channel iterator that returns values recieved from channel
func TestFromChan(t *testing.T) {
	cases := []struct {
		source   func() chan int
		expected []int
	}{
		{
			source: func() chan int {
				c := make(chan int)
				go func() {
					defer close(c)
					for _, v := range []int{0, 1, 2, 3, 4} {
						c <- v
					}
				}()
				return c
			},
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			source: func() chan int {
				c := make(chan int)
				go func() {
					defer close(c)
					for _, v := range []int{} {
						c <- v
					}
				}()
				return c
			},
			expected: []int{},
		},
	}

	for i, testCase := range cases {
		c := testCase.source()
		it := iter.FromChan(context.Background(), c)
		validateResult(t, i, it, testCase.expected)
	}
}

// Tests that channel iterator can be stopped with context
func TestFromChanClose(t *testing.T) {
	var c chan int
	ctx, cancel := context.WithCancel(context.Background())
	it := iter.FromChan(ctx, c)

	success := make(chan struct{})
	fail := time.NewTimer(1 * time.Second)

	go func() {
		it.Next()
		close(success)
	}()

	cancel()

	select {
	case <-success:
	case <-fail.C:
		t.Fatal("timeout, iterator is not stopped by context")
	}
}

// Tests reduce finalizer
// Returns reduced value or error if occured
// If error is ErrStopIt just stops
func TestReduce(t *testing.T) {
	newError := errors.New("error")
	errorIt := iter.Map(iter.FromSlice([]int{1, 2, 3, 4}), func(el int) (int, error) {
		if el == 3 {
			return 0, newError
		}
		return el, nil
	})
	stopIt := iter.Map(iter.FromSlice([]int{1, 2, 3, 4}), func(el int) (int, error) {
		if el == 3 {
			return 0, iter.ErrStopIt
		}
		return el, nil
	})
	cases := []struct {
		iterator iter.Iterator[int]
		reduce   func(int, int) int
		init     int
		expected int
		err      error
	}{
		{
			iterator: iter.FromSlice([]int{0, 1, 2, 3, 4}),
			reduce: func(el, acc int) int {
				return acc + el
			},
			init:     0,
			expected: 10,
			err:      nil,
		},
		{
			iterator: iter.FromSlice([]int{1, 2, 3, 4}),
			reduce: func(el, acc int) int {
				return acc * el
			},
			init:     1,
			expected: 24,
			err:      nil,
		},
		{
			// Test error return from source iterator
			iterator: errorIt,
			reduce: func(el, acc int) int {
				return acc * el
			},
			init:     1,
			expected: 0,
			err:      newError,
		},
		{
			// Test stop iteration error from source iterator
			iterator: stopIt,
			reduce: func(el, acc int) int {
				return acc * el
			},
			init:     1,
			expected: 2,
			err:      nil,
		},
	}

	for i, testCase := range cases {
		result, err := iter.Reduce(testCase.iterator, testCase.init, testCase.reduce)
		if !errors.Is(err, testCase.err) {
			t.Fatalf("[%v] error not equal to expected\n%v != %v\n", i, err, testCase.err)
		}
		if result != testCase.expected {
			t.Fatalf("[%v] result not equal to expected\n%v != %v\n", i, result, testCase.expected)
		}
	}
}

func TestToSliceError(t *testing.T) {
	newError := errors.New("error")
	errorIt := iter.Map(iter.FromSlice([]int{0, 1, 2, 3, 4}), func(el int) (int, error) {
		if el == 3 {
			return 0, newError
		}
		return el, nil
	})
	stopIt := iter.Map(iter.FromSlice([]int{0, 1, 2, 3, 4}), func(el int) (int, error) {
		if el == 3 {
			return 0, iter.ErrStopIt
		}
		return el, nil
	})

	cases := []struct {
		iterator iter.Iterator[int]
		expected []int
		err      error
	}{
		{
			iterator: errorIt,
			expected: nil,
			err:      newError,
		},
		{
			iterator: stopIt,
			expected: []int{0, 1, 2},
		},
	}

	for i, testCase := range cases {
		result, err := iter.ToSlice(testCase.iterator)
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Fatalf("[%v] iterator elements are not equal to expected values\n%v != %v\n",
				i, result, testCase.expected)
		}
		if !errors.Is(err, testCase.err) {
			t.Fatalf("[%v] unexpected type of error\n%v != %v\n", i, err, testCase.err)
		}
	}
}

// Tests ToChan and ToChanSimple finalizers
// returns channel that will recieve values from iterator
func TestToChan(t *testing.T) {
	cases := []struct {
		source   []int
		expected []int
	}{
		{
			source:   []int{},
			expected: []int{},
		},
		{
			source:   []int{0, 1, 2, 3, 4},
			expected: []int{0, 1, 2, 3, 4},
		},
	}

	for i, testCase := range cases {
		source := iter.FromSlice(testCase.source)
		c := iter.ToChan(context.Background(), source)
		result := make([]int, 0, len(testCase.expected))
		for v := range c {
			result = append(result, v)
		}
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Fatalf(
				"[%v] channel has sent unexpected values\n%v != %v\n",
				i, result, testCase.expected)
		}
	}
}

func TestToChanCancel(t *testing.T) {
	it := iter.Generator(func() int { return 1 })
	ctx, cancel := context.WithCancel(context.Background())
	c := iter.ToChan(ctx, it)

	success := make(chan struct{})
	fail := time.NewTimer(1 * time.Second)

	go func() {
		for range c {
		}
		close(success)
	}()

	cancel()

	select {
	case <-success:
	case <-fail.C:
		t.Fatal("timeout, channel is not closed by context")
	}
}

func TestFinal(t *testing.T) {
	newError := errors.New("error")
	errorIt := iter.Map(iter.FromSlice([]int{0, 1, 2, 3, 4}), func(el int) (int, error) {
		if el == 3 {
			return 0, newError
		}
		return el, nil
	})
	stopIt := iter.Map(iter.FromSlice([]int{0, 1, 2, 3, 4}), func(el int) (int, error) {
		if el == 3 {
			return 0, iter.ErrStopIt
		}
		return el, nil
	})

	cases := []struct {
		iterator iter.Iterator[int]
		expected []int
		err      error
		stop     error
	}{
		{
			iterator: iter.FromSlice([]int{0, 1, 2, 3, 4}),
			expected: []int{0, 1, 2, 3, 4},
			err:      nil,
			stop:     iter.ErrStopIt,
		},
		{
			// Test error return from source iterator
			iterator: errorIt,
			expected: []int{0, 1, 2},
			err:      newError,
			stop:     newError,
		},
		{
			// Test stop iteration error from source iterator
			iterator: stopIt,
			expected: []int{0, 1, 2},
			err:      nil,
			stop:     iter.ErrStopIt,
		},
	}

	for i, testCase := range cases {
		it := iter.Final(testCase.iterator)
		result := make([]int, 0)
		for it.Next() {
			v := it.Get()
			result = append(result, v)
		}
		if !reflect.DeepEqual(result, testCase.expected) {
			t.Fatalf("[%v] iterator elements are not equal to expected values\n%v != %v\n",
				i, result, testCase.expected)
		}
		if !errors.Is(it.Stop(), testCase.stop) {
			t.Fatalf("[%v] unexpected stop error\n%v != %v\n", i, it.Stop(), testCase.stop)
		}
		if !errors.Is(it.Err(), testCase.err) {
			t.Fatalf("[%v] unexpected type of error\n%v != %v\n", i, it.Err(), testCase.err)
		}
		if it.Next() {
			t.Fatalf("[%v] iterator should be stopped at that point\n", i)
		}
	}
}
