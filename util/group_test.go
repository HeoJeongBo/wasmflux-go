package util

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGroup_Success(t *testing.T) {
	g, cancel := NewGroup(context.Background())
	defer cancel()

	g.Go(func(_ context.Context) error { return nil })
	g.Go(func(_ context.Context) error { return nil })

	errs := g.Wait()
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

func TestGroup_Errors(t *testing.T) {
	g, cancel := NewGroup(context.Background())
	defer cancel()

	g.Go(func(_ context.Context) error { return errors.New("err1") })
	g.Go(func(_ context.Context) error { return nil })
	g.Go(func(_ context.Context) error { return errors.New("err2") })

	errs := g.Wait()
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}
}

func TestGroup_PanicRecovery(t *testing.T) {
	g, cancel := NewGroup(context.Background())
	defer cancel()

	g.Go(func(_ context.Context) error {
		panic("crash!")
	})

	errs := g.Wait()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "crash!") {
		t.Errorf("error should contain panic value, got %q", errs[0])
	}
}

func TestGroup_WaitFirst(t *testing.T) {
	g, cancel := NewGroup(context.Background())
	defer cancel()

	g.Go(func(_ context.Context) error { return errors.New("first") })
	g.Go(func(_ context.Context) error { return nil })

	err := g.WaitFirst()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGroup_WaitFirst_NoError(t *testing.T) {
	g, cancel := NewGroup(context.Background())
	defer cancel()

	g.Go(func(_ context.Context) error { return nil })

	err := g.WaitFirst()
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestGroup_Cancel(t *testing.T) {
	g, cancel := NewGroup(context.Background())

	done := make(chan struct{})
	g.Go(func(ctx context.Context) error {
		<-ctx.Done()
		close(done)
		return ctx.Err()
	})

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("goroutine should have been cancelled")
	}

	errs := g.Wait()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
}

func TestGroup_EmptyWait(t *testing.T) {
	g, cancel := NewGroup(context.Background())
	defer cancel()

	errs := g.Wait()
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

func TestGroup_Context(t *testing.T) {
	g, cancel := NewGroup(context.Background())
	defer cancel()

	ctx := g.Context()
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
	if ctx.Err() != nil {
		t.Error("context should not be cancelled yet")
	}
}
