package semgroup

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestGroup_single_task(t *testing.T) {
	ctx := context.Background()
	g := NewGroup(ctx, 1)

	g.Go(func() error { return nil })

	err := g.Wait()
	if err != nil {
		t.Errorf("g.Wait() should not return an error")
	}
}

func TestGroup_multiple_tasks(t *testing.T) {
	ctx := context.Background()
	g := NewGroup(ctx, 1)

	count := 0
	var mu sync.Mutex

	inc := func() error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}

	g.Go(func() error { return inc() })
	g.Go(func() error { return inc() })
	g.Go(func() error { return inc() })
	g.Go(func() error { return inc() })

	err := g.Wait()
	if err != nil {
		t.Errorf("g.Wait() should not return an error")
	}

	if count != 4 {
		t.Errorf("count should be %d, got: %d", 4, count)
	}
}

func TestGroup_multiple_tasks_errors(t *testing.T) {
	ctx := context.Background()
	g := NewGroup(ctx, 1)

	g.Go(func() error { return errors.New("foo") })
	g.Go(func() error { return nil })
	g.Go(func() error { return errors.New("bar") })
	g.Go(func() error { return nil })

	err := g.Wait()
	if err == nil {
		t.Fatalf("g.Wait() should return an error")
	}

	wantErr := `2 error(s) occured:
* foo
* bar`

	if wantErr != err.Error() {
		t.Errorf("error should be:\n%s\ngot:\n%s\n", wantErr, err.Error())
	}
}
