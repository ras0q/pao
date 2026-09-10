package pao

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Resource is a handle to async data managed by Bubble Tea commands.
type Resource[T any] struct {
	cell *resourceCell[T]
}

// NewResource creates a resource that fetches data with fetch. Canceling ctx cancels an in-flight fetch.
func NewResource[T any](ctx context.Context, fetch func(context.Context) (T, error)) Resource[T] {
	c := &resourceCell[T]{
		ctx:   ctx,
		fetch: fetch,
	}
	c.state.Store(&State[T]{})
	context.AfterFunc(ctx, c.abort)

	return Resource[T]{cell: c}
}

// Snapshot returns a copy of the current resource state.
func (r Resource[T]) Snapshot() State[T] {
	return *r.cell.state.Load()
}

// Ensure starts a fetch when data is missing or invalidated.
func (r Resource[T]) Ensure() tea.Cmd {
	return r.cell.startFetch(false)
}

// Refresh starts a fetch when one is not already in progress.
func (r Resource[T]) Refresh() tea.Cmd {
	return r.cell.startFetch(true)
}

// Invalidate marks data stale and cancels any in-flight fetch.
func (r Resource[T]) Invalidate() {
	r.cell.invalidate()
}

type resourceCell[T any] struct {
	ctx   context.Context
	fetch func(context.Context) (T, error)
	state atomic.Pointer[State[T]]

	mu         sync.Mutex
	generation uint64
	cancel     context.CancelFunc
}

func (c *resourceCell[T]) startFetch(force bool) tea.Cmd {
	c.mu.Lock()
	state := *c.state.Load()
	if state.Fetching || (!force && state.HasData && !state.Invalidated) {
		c.mu.Unlock()
		return nil
	}

	c.generation++
	gen := c.generation
	state.Fetching = true
	c.state.Store(&state)

	ctx, cancel := context.WithCancel(c.ctx)
	c.cancel = cancel
	c.mu.Unlock()

	return func() tea.Msg {
		defer cancel()

		data, err := c.fetch(ctx)

		c.mu.Lock()
		defer c.mu.Unlock()

		if gen != c.generation {
			return nil
		}

		c.cancel = nil
		state := *c.state.Load()
		state.Fetching = false

		if ctx.Err() != nil {
			c.state.Store(&state)
			return nil
		}

		if err != nil {
			state.Err = err
			c.state.Store(&state)
			return LoadedMsg[T]{
				Resource: Resource[T]{cell: c},
				State:    state,
			}
		}

		state.Data = data
		state.HasData = true
		state.Err = nil
		state.UpdatedAt = time.Now()
		state.Invalidated = false
		c.state.Store(&state)
		return LoadedMsg[T]{
			Resource: Resource[T]{cell: c},
			State:    state,
		}
	}
}

func (c *resourceCell[T]) invalidate() {
	c.stop(true)
}

func (c *resourceCell[T]) abort() {
	c.stop(false)
}

func (c *resourceCell[T]) stop(invalidate bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.generation++
	if invalidate {
		if c.cancel != nil {
			c.cancel()
		}
	}
	c.cancel = nil

	state := *c.state.Load()
	state.Fetching = false
	if invalidate {
		state.Invalidated = true
	}
	c.state.Store(&state)
}

// LoadedMsg is sent when a fetch completes. Handle it in Update, or read Snapshot in View.
type LoadedMsg[T any] struct {
	Resource Resource[T]
	State    State[T]
}
