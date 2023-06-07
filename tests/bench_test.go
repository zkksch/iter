/*
Testing perfomance against other possible solutions for creating pipelines

Test is pretty simple, we have slice of ints and pipeline that:

	Removes ints that are not divisible by 2
	Divides them by 2
	Removes ints that are not divisible by 3
	Divides them by 3
	Counts remaining ints that are divisible by 5

We're testing 4 possible solutions:

	Fastest solution: hardcoded for loop, without pipeline
	Slice solution: transforming slice into a new slice on each stage
	Channel solution: using channels and goroutines to transport values
	Iterator solution: using functionality of that library

Here results of one of the runs on my PC:

	BenchmarkVs/for_loop-1000-12         	  739795	      1465 ns/op	       0 B/op	       0 allocs/op
	BenchmarkVs/slice_pipes-1000-12      	  107056	     10457 ns/op	   17776 B/op	      21 allocs/op
	BenchmarkVs/chan_pipes-1000-12       	    1384	    901248 ns/op	     744 B/op	      10 allocs/op
	BenchmarkVs/iter_pipes-1000-12       	   80746	     15336 ns/op	     144 B/op	       5 allocs/op
	BenchmarkVs/for_loop-10000-12        	   21633	     55425 ns/op	       0 B/op	       0 allocs/op
	BenchmarkVs/slice_pipes-10000-12     	    8270	    132290 ns/op	  222321 B/op	      31 allocs/op
	BenchmarkVs/chan_pipes-10000-12      	     133	   8992839 ns/op	     803 B/op	      10 allocs/op
	BenchmarkVs/iter_pipes-10000-12      	    6889	    180489 ns/op	     144 B/op	       5 allocs/op
	BenchmarkVs/for_loop-100000-12       	    1916	    625871 ns/op	       0 B/op	       0 allocs/op
	BenchmarkVs/slice_pipes-100000-12    	     759	   1617881 ns/op	 3156492 B/op	      48 allocs/op
	BenchmarkVs/chan_pipes-100000-12     	      13	  89800554 ns/op	     760 B/op	      10 allocs/op
	BenchmarkVs/iter_pipes-100000-12     	     727	   1667509 ns/op	     144 B/op	       5 allocs/op

As we can see channel solution, that pretty often mentioned, is very slow, and slice solution allocates O(n) memory
For loop is much faster but it doesn't provide flexibility of pipelines

Iterators not that much slower than a slice solution, but have constant amount of memory usage and allocations (only need to allocate Iterators)
*/
package tests

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/zkksch/iter"
)

// The fastest solution
func forLoop(s []int) int {
	var cnt int
	for _, el := range s {
		if el%2 != 0 {
			continue
		}
		el /= 2
		if el%3 != 0 {
			continue
		}
		el /= 3
		if el%5 == 0 {
			cnt++
		}
	}
	return cnt
}

// Filter pipe for a slice solution
func sliceFilter[T any](s []T, fn func(T) bool) []T {
	result := make([]T, 0)
	for _, el := range s {
		if fn(el) {
			result = append(result, el)
		}
	}
	return result
}

// Map pipe for a slice solution
func sliceMap[T, K any](s []T, fn func(T) (K, error)) ([]K, error) {
	// We already know size of resulting slice,
	// so we can allocate enough memory from the start
	result := make([]K, 0, len(s))
	for _, el := range s {
		v, err := fn(el)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

// Count finalizer for a slice solution
func sliceCnt[T any](s []T, fn func(T) bool) int {
	var cnt int
	for _, el := range s {
		if fn(el) {
			cnt++
		}
	}
	return cnt
}

// Slice solution
func slicePipe(s []int) (int, error) {
	s = sliceFilter[int](s, func(i int) bool { return i%2 == 0 })
	s, err := sliceMap[int, int](s, func(i int) (int, error) { return i / 2, nil })
	if err != nil {
		return 0, err
	}
	s = sliceFilter[int](s, func(i int) bool { return i%3 == 0 })
	s, err = sliceMap[int, int](s, func(i int) (int, error) { return i / 3, nil })
	if err != nil {
		return 0, err
	}
	result := sliceCnt[int](s, func(i int) bool { return i%5 == 0 })
	return result, nil
}

// Creates a channel from a slice
func sliceChan[T any](s []T) <-chan T {
	result := make(chan T)
	go func() {
		defer close(result)
		for _, el := range s {
			result <- el
		}
	}()
	return result
}

// Filter pipe for a channel solution
func filterChan[T any](c <-chan T, fn func(T) bool) <-chan T {
	result := make(chan T)
	go func() {
		defer close(result)
		for el := range c {
			if fn(el) {
				result <- el
			}
		}
	}()
	return result
}

// Map pipe for a channel solution
func mapChan[T, K any](c <-chan T, fn func(T) (K, error)) <-chan K {
	result := make(chan K)
	go func() {
		defer close(result)
		for el := range c {
			v, err := fn(el)
			if err != nil {
				return
			}
			result <- v
		}
	}()
	return result
}

// Count finalizer for a channel solution
func cntChan[T any](c <-chan T, fn func(T) bool) int {
	var cnt int
	for el := range c {
		if fn(el) {
			cnt++
		}
	}

	return cnt
}

// Channel solution
func chanPipe(s []int) (int, error) {
	sch := sliceChan[int](s)
	sch = filterChan[int](sch, func(i int) bool { return i%2 == 0 })
	sch = mapChan[int, int](sch, func(i int) (int, error) { return i / 2, nil })
	sch = filterChan[int](sch, func(i int) bool { return i%3 == 0 })
	sch = mapChan[int, int](sch, func(i int) (int, error) { return i / 3, nil })
	return cntChan[int](sch, func(i int) bool { return i%5 == 0 }), nil
}

// Iterator solution
func iterPipe(s []int) (int, error) {
	sit := iter.FromSlice[int](s)
	sit = iter.Filter[int](sit, func(i int) bool { return i%2 == 0 })
	sit = iter.Map[int, int](sit, func(i int) (int, error) { return i / 2, nil })
	sit = iter.Filter[int](sit, func(i int) bool { return i%3 == 0 })
	sit = iter.Map[int, int](sit, func(i int) (int, error) { return i / 3, nil })
	// Don't have count finalizer, but can use Reduce to achieve same result
	result, err := iter.Reduce[int, int](sit, 0, func(el, acc int) int {
		if el%5 == 0 {
			acc += 1
		}
		return acc
	})
	return result, err
}

// Comparing solutions in 3 cases (N = 1'000, 10'000 and 100'000)
// Generates slice with size N, fills it with random numbers (from 0 to N)
// We also need to check that all of them done the same work by comparing results
func BenchmarkVs(b *testing.B) {
	cases := []int{
		1000,
		10000,
		100000,
	}

	for _, n := range cases {
		gen := iter.Generator[int](func() int { return rand.Intn(n) })
		gen = iter.Limit[int](gen, n)
		data, err := iter.ToSlice[int](gen)
		if err != nil {
			b.Fatal(err)
		}
		valid := forLoop(data)

		b.Run(fmt.Sprintf("for loop-%v", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result := forLoop(data)
				if result != valid {
					b.Fatalf("wrong result %v != %v", result, valid)
				}
			}
		})

		b.Run(fmt.Sprintf("slice pipes-%v", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result, err := slicePipe(data)
				if err != nil {
					b.Fatal(err)
				}
				if result != valid {
					b.Fatalf("wrong result %v != %v", result, valid)
				}
			}
		})

		b.Run(fmt.Sprintf("chan pipes-%v", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result, err := chanPipe(data)
				if err != nil {
					b.Fatal(err)
				}
				if result != valid {
					b.Fatalf("wrong result %v != %v", result, valid)
				}
			}
		})

		b.Run(fmt.Sprintf("iter pipes-%v", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result, err := iterPipe(data)
				if err != nil {
					b.Fatal(err)
				}
				if result != valid {
					b.Fatalf("wrong result %v != %v", result, valid)
				}
			}
		})
	}
}
