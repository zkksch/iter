package tests

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/zkksch/iter/iter"
)

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

func sliceFilter[T any](s []T, fn func(T) bool) []T {
	result := make([]T, 0)
	for _, el := range s {
		if fn(el) {
			result = append(result, el)
		}
	}
	return result
}

func sliceMap[T, K any](s []T, fn func(T) (K, error)) ([]K, error) {
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

func sliceCnt[T any](s []T, fn func(T) bool) int {
	var cnt int
	for _, el := range s {
		if fn(el) {
			cnt++
		}
	}
	return cnt
}

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

func cntChan[T any](c <-chan T, fn func(T) bool) int {
	var cnt int
	for el := range c {
		if fn(el) {
			cnt++
		}
	}

	return cnt
}

func chanPipe(s []int) (int, error) {
	sch := sliceChan[int](s)
	sch = filterChan[int](sch, func(i int) bool { return i%2 == 0 })
	sch = mapChan[int, int](sch, func(i int) (int, error) { return i / 2, nil })
	sch = filterChan[int](sch, func(i int) bool { return i%3 == 0 })
	sch = mapChan[int, int](sch, func(i int) (int, error) { return i / 3, nil })
	return cntChan[int](sch, func(i int) bool { return i%5 == 0 }), nil
}

func iterPipe(s []int) (int, error) {
	sit := iter.FromSlice[int](s)
	sit = iter.Filter[int](sit, func(i int) bool { return i%2 == 0 })
	sit = iter.Map[int, int](sit, func(i int) (int, error) { return i / 2, nil })
	sit = iter.Filter[int](sit, func(i int) bool { return i%3 == 0 })
	sit = iter.Map[int, int](sit, func(i int) (int, error) { return i / 3, nil })
	result, err := iter.Reduce[int, int](sit, 0, func(el, acc int) int {
		if el%5 == 0 {
			acc += 1
		}
		return acc
	})
	return result, err
}

func BenchmarkVs(b *testing.B) {
	cases := []int{
		1000,
		10000,
		100000,
	}

	for _, n := range cases {
		gen := iter.Generate[int](func(i int) (int, error) { return rand.Intn(n), nil })
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
					b.Fatalf("wrong length of resulting array %v != %v", result, valid)
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
					b.Fatalf("wrong length of resulting array %v != %v", result, valid)
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
					b.Fatalf("wrong length of resulting array %v != %v", result, valid)
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
					b.Fatalf("wrong length of resulting array %v != %v", result, valid)
				}
			}
		})
	}
}
