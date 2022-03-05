package semgroup_test

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/fatih/semgroup"
)

// This example increases a counter for each visit concurrently, using a
// SemGroup to block until all the visitors have finished. It only runs 2 tasks
// at any time.
func ExampleGroup_parallel() {
	const maxWorkers = 2
	s := semgroup.NewGroup(context.Background(), maxWorkers)

	var (
		counter int
		mu      sync.Mutex // protects visits
	)

	visitors := []int{5, 2, 10, 8, 9, 3, 1}

	for _, v := range visitors {
		v := v

		s.Go(func() error {
			mu.Lock()
			counter += v
			mu.Unlock()
			return nil
		})
	}

	// Wait for all visits to complete. Any errors are accumulated.
	if err := s.Wait(); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Counter: %d", counter)

	// Output:
	// Counter: 38
}

func ExampleGroup_withErrors() {
	const maxWorkers = 2
	s := semgroup.NewGroup(context.Background(), maxWorkers)

	visitors := []int{1, 1, 1, 1, 2, 2, 1, 1, 2}

	for _, v := range visitors {
		v := v

		s.Go(func() error {
			if v != 1 {
				return errors.New("only one visitor is allowed")
			}
			return nil
		})
	}

	// Wait for all visits to complete. Any errors are accumulated.
	if err := s.Wait(); err != nil {
		fmt.Println(err)
	}

	// Output:
	// 3 error(s) occurred:
	// * only one visitor is allowed
	// * only one visitor is allowed
	// * only one visitor is allowed
}
