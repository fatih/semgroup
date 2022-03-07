package semgroup

import (
	"context"
	"errors"
	"os"
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

	wantErr := `2 error(s) occurred:
* foo
* bar`

	if wantErr != err.Error() {
		t.Errorf("error should be:\n%s\ngot:\n%s\n", wantErr, err.Error())
	}
}

func TestGroup_deadlock(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	g := NewGroup(canceledCtx, 1)

	g.Go(func() error { return nil })
	g.Go(func() error { return nil })

	err := g.Wait()
	if err == nil {
		t.Fatalf("g.Wait() should return an error")
	}

	wantErr := `1 error(s) occurred:
* couldn't acquire semaphore: context canceled`

	if wantErr != err.Error() {
		t.Errorf("error should be:\n%s\ngot:\n%s\n", wantErr, err.Error())
	}
}

func TestGroup_multiple_tasks_errors_Is(t *testing.T) {
	ctx := context.Background()
	g := NewGroup(ctx, 1)

	var (
		fooErr = errors.New("foo")
		barErr = errors.New("bar")
		bazErr = errors.New("baz")
	)

	g.Go(func() error { return fooErr })
	g.Go(func() error { return nil })
	g.Go(func() error { return barErr })
	g.Go(func() error { return nil })

	err := g.Wait()
	if err == nil {
		t.Fatalf("g.Wait() should return an error")
	}

	if !errors.Is(err, fooErr) {
		t.Errorf("error should be contained %v\n", fooErr)
	}

	if !errors.Is(err, barErr) {
		t.Errorf("error should be contained %v\n", barErr)
	}

	if errors.Is(err, bazErr) {
		t.Errorf("error should not be contained %v\n", bazErr)
	}
}

type foobarErr struct{ str string }

func (e foobarErr) Error() string {
	return "foobar"
}

func TestGroup_multiple_tasks_errors_As(t *testing.T) {
	ctx := context.Background()
	g := NewGroup(ctx, 1)

	g.Go(func() error { return foobarErr{"baz"} })
	g.Go(func() error { return nil })

	err := g.Wait()
	if err == nil {
		t.Fatalf("g.Wait() should return an error")
	}

	var (
		fbe foobarErr
		pe  *os.PathError
	)

	if !errors.As(err, &fbe) {
		t.Error("error should be matched foobarErr")
	}

	if errors.As(err, &pe) {
		t.Error("error should not be matched os.PathError")
	}
}
