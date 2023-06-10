// Includes tests for thread safiness of iterators
// Only need to check thread safiness of stateful iterators or
// iterators that combine values from multiple sources
package tests

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/zkksch/iter"
)

// Function iterateAll iterates until an iterator stops in given number of goroutines
func iterateAll[T any](goroutines int, it iter.Iterator[T]) int {
	start := make(chan struct{})
	total := &atomic.Int64{}
	wg := &sync.WaitGroup{}
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			a := 0
			for _, err := it(); err == nil; _, err = it() {
				a++
			}
			total.Add(int64(a))
		}()
	}

	close(start)
	wg.Wait()
	return int(total.Load())
}

// Function iterateN iterates over an iterator in given number of goroutines (n times each)
func iterateN[T any](goroutines int, it iter.Iterator[T], n int) int {
	start := make(chan struct{})
	total := &atomic.Int64{}
	wg := &sync.WaitGroup{}
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			a := 0
			for j := 0; j < n; j++ {
				_, err := it()
				if err != nil {
					break
				}
				a++
			}
			total.Add(int64(a))
		}()
	}

	close(start)
	wg.Wait()

	return int(total.Load())
}

// Function iterateCheck iterates over an iterator and checks that check function from an element returns true
func iterateCheck[T any](goroutines int, it iter.Iterator[T], check func(T) bool) (int, bool) {
	start := make(chan struct{})
	total := &atomic.Int64{}
	failed := &atomic.Bool{}
	wg := &sync.WaitGroup{}
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			a := 0
			for val, err := it(); err == nil; val, err = it() {
				a++
				if !check(val) {
					failed.Store(true)
				}
			}
			total.Add(int64(a))
		}()
	}

	close(start)
	wg.Wait()
	return int(total.Load()), failed.Load()
}

const elements = 1000000
const goroutines = 20

// Tests that iterator with slice source will have
// same amount of iterations as amount of elements in the slice
func TestSafeFromSliceSafe(t *testing.T) {
	it := iter.FromSliceSafe(make([]int, elements))
	k := iterateAll(goroutines, it)
	if k != elements {
		t.Fatalf("wrong number of iterations %v != %v\n", elements, k)
	}
}

// Tests sequence iterator, checks that after N iterations the next element of iterator will be N
func TestSafeSequenceSafe(t *testing.T) {
	elements := (elements / goroutines) * goroutines
	it := iter.SequenceSafe(0, 1)
	k := iterateN(goroutines, it, elements/goroutines)
	if k != elements {
		t.Fatalf("wrong number of iterations %v != %v\n", elements, k)
	}
	next, err := it()
	if err != nil {
		t.Fatal(err)
	}
	if next != elements {
		t.Fatalf("wrong next element %v != %v\n", next, elements)
	}
}

// Tests cycling iterator, checks that after N iterations (N divisible by amount of values)
// it returned the same amount of each value from the set
func TestSafeCycleSafe(t *testing.T) {
	elements := (elements / 10) * 10
	it := iter.CycleSafe(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	it = iter.LimitSafe(it, elements)
	counters := []*atomic.Int32{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}}
	k, _ := iterateCheck(goroutines, it, func(el int) bool {
		counters[el].Add(1)
		return true
	})
	if k != elements {
		t.Fatalf("wrong number of iterations %v != %v\n", elements, k)
	}
	countersVerbose := make([]int, 10)
	for i, c := range counters {
		countersVerbose[i] = int(c.Load())
	}
	for i := 1; i < 10; i++ {
		if countersVerbose[i-1] != countersVerbose[i] {
			t.Fatalf("cycle repeated each element unevenly\n%v\n", countersVerbose)
		}
	}
}

// Tests that limit iterator with limit of N will perfom N iterations
func TestSafeLimitSafe(t *testing.T) {
	it := iter.Generator(func() int { return 0 })
	it = iter.LimitSafe(it, elements)
	k := iterateAll(goroutines, it)
	if k != elements {
		t.Fatalf("wrong number of iterations %v != %v\n", elements, k)
	}
}

// Tests that pairs iterator from 2 iterators with the same source
// will iterate over pairs of the same values
func TestSafePairsSafe(t *testing.T) {
	gen := iter.Sequence(0, 1)
	sl, err := iter.ToSlice(iter.Limit(gen, elements))
	if err != nil {
		t.Fatal(err)
	}

	sourceLeft := iter.FromSliceSafe(sl)
	sourceRight := iter.FromSliceSafe(sl)

	it := iter.PairsSafe(sourceLeft, sourceRight)
	k, failed := iterateCheck(goroutines, it, func(p iter.Pair[int, int]) bool {
		// Each pair should contain the same int from left and right sources
		return p.Left == p.Right
	})
	if failed {
		t.Fatalf("Pairs are not synchronized")
	}
	if k != elements {
		t.Fatalf("wrong number of iterations %v != %v\n", elements, k)
	}
}

// Tests that combine iterator from iterators with the same source
// will iterate over groups of the same values
func TestSafeCombineSafe(t *testing.T) {
	n := 5
	gen := iter.Sequence(0, 1)
	sl, err := iter.ToSlice(iter.Limit(gen, elements))
	if err != nil {
		t.Fatal(err)
	}

	iterators := make([]iter.Iterator[int], n)

	for i := 0; i < n; i++ {
		iterators[i] = iter.FromSliceSafe(sl)
	}

	it := iter.CombineSafe(iterators...)
	k, failed := iterateCheck(goroutines, it, func(g []int) bool {
		// Each group should contain same ints
		if len(g) != n {
			return false
		}
		for i := 1; i < len(g); i++ {
			if g[i-1] != g[i] {
				return false
			}
		}
		return true
	})
	if failed {
		t.Fatalf("Groups are not synchronized")
	}
	if k != elements {
		t.Fatalf("wrong number of iterations %v != %v\n", elements, k)
	}
}

// Tests that amount of iterations in chain will be equal
// to total amount of iterations in iterators used in chain
func TestSafeChainSafe(t *testing.T) {
	elements := (elements / 10) * 10
	gen := iter.Generator(func() int { return 0 })
	sources := make([]iter.Iterator[int], 10)
	for i := 0; i < 10; i++ {
		// Source also should be thread safe
		sources[i] = iter.LimitSafe(gen, elements/10)
	}

	it := iter.ChainSafe(sources...)
	k := iterateAll(goroutines, it)
	if k != elements {
		t.Fatalf("wrong number of iterations %v != %v\n", elements, k)
	}
}
