package pao

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestEnsureFetchesOnce(t *testing.T) {
	var calls atomic.Int32
	res := NewResource(t.Context(), func(ctx context.Context) (int, error) {
		calls.Add(1)
		return 42, nil
	})

	mustLoaded[int](t, run(res.Ensure()))
	if run(res.Ensure()) != nil {
		t.Fatal("second Ensure should not fetch")
	}
	if calls.Load() != 1 {
		t.Fatalf("fetch calls = %d, want 1", calls.Load())
	}

	assertState(t, res.Snapshot(), State[int]{Data: 42, HasData: true})
}

func TestRefreshRefetches(t *testing.T) {
	var calls atomic.Int32
	res := NewResource(t.Context(), func(ctx context.Context) (int, error) {
		return int(calls.Add(1)), nil
	})

	run(res.Ensure())
	mustLoaded[int](t, run(res.Refresh()))
	if calls.Load() != 2 {
		t.Fatalf("fetch calls = %d, want 2", calls.Load())
	}
}

func TestInvalidateRefetchesOnEnsure(t *testing.T) {
	var calls atomic.Int32
	res := NewResource(t.Context(), func(ctx context.Context) (int, error) {
		calls.Add(1)
		return 1, nil
	})

	run(res.Ensure())
	res.Invalidate()
	assertState(t, res.Snapshot(), State[int]{Data: 1, HasData: true, Invalidated: true})

	loaded := mustLoaded[int](t, run(res.Ensure()))
	if loaded.State.Invalidated {
		t.Fatalf("loaded state remains invalidated: %+v", loaded.State)
	}
	if calls.Load() != 2 {
		t.Fatalf("fetch calls = %d, want 2", calls.Load())
	}

	assertState(t, res.Snapshot(), State[int]{Data: 1, HasData: true})
}

func TestFetchErrorKeepsPreviousData(t *testing.T) {
	errBoom := errors.New("boom")
	var fail atomic.Bool
	res := NewResource(t.Context(), func(ctx context.Context) (int, error) {
		if fail.Load() {
			return 0, errBoom
		}
		return 7, nil
	})

	run(res.Ensure())

	fail.Store(true)
	loaded := mustLoaded[int](t, run(res.Refresh()))
	assertState(t, loaded.State, State[int]{Data: 7, HasData: true, Err: errBoom})
	assertState(t, res.Snapshot(), State[int]{Data: 7, HasData: true, Err: errBoom})
}

func TestEnsureWaitsForInFlightFetch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		started := make(chan struct{})
		release := make(chan struct{})
		res := NewResource(t.Context(), func(ctx context.Context) (int, error) {
			close(started)
			select {
			case <-release:
				return 1, nil
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		})

		go run(res.Ensure())

		<-started
		if run(res.Ensure()) != nil {
			t.Fatal("Ensure during fetch should return nil")
		}

		close(release)
		synctest.Wait()

		assertState(t, res.Snapshot(), State[int]{Data: 1, HasData: true})
	})
}

func TestContextCancelDropsStaleResult(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		started := make(chan struct{})
		ctx, cancel := context.WithCancel(t.Context())

		res := NewResource(ctx, func(ctx context.Context) (int, error) {
			close(started)
			<-ctx.Done()
			return 99, ctx.Err()
		})

		go func() {
			assertNoMsg(t, run(res.Ensure()))
		}()

		<-started
		cancel()
		synctest.Wait()

		assertState(t, res.Snapshot(), State[int]{})
	})
}

func TestContextCancelDropsSuccessfulResultWhenParentCancelsDuringFetch(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	res := NewResource(ctx, func(ctx context.Context) (int, error) {
		cancel()
		return 99, nil
	})

	assertNoMsg(t, run(res.Ensure()))
	assertState(t, res.Snapshot(), State[int]{})
}

func TestFetchContextCanceledErrorIsReportedWhileContextIsActive(t *testing.T) {
	res := NewResource(t.Context(), func(ctx context.Context) (int, error) {
		return 0, context.Canceled
	})

	loaded := mustLoaded[int](t, run(res.Ensure()))
	assertState(t, loaded.State, State[int]{Err: context.Canceled})
	assertState(t, res.Snapshot(), State[int]{Err: context.Canceled})
}

func TestSnapshotDoesNotWaitForWriterLock(t *testing.T) {
	res := NewResource(t.Context(), func(ctx context.Context) (int, error) {
		return 42, nil
	})

	res.cell.mu.Lock()
	defer res.cell.mu.Unlock()

	done := make(chan State[int], 1)
	go func() {
		done <- res.Snapshot()
	}()

	select {
	case snap := <-done:
		assertState(t, snap, State[int]{})
	case <-time.After(time.Second):
		t.Fatal("Snapshot waited for the writer lock")
	}
}
