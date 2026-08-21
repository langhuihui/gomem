package gomem

import "sync"

// Pool is a generic, type-safe wrapper around [sync.Pool].
// It eliminates the type assertions that are otherwise required when using
// sync.Pool directly, and provides a more ergonomic API for strongly-typed
// object pools.
//
// T should be a pointer type or a non-pointer type whose zero value is
// meaningful, since sync.Pool may return any previously Put value or call
// New to create a fresh one.
type Pool[T any] struct {
	p sync.Pool
}

// NewPool creates a new Pool whose New function calls the provided
// constructor whenever the pool is empty.
func NewPool[T any](newFn func() T) *Pool[T] {
	pool := &Pool[T]{}
	pool.p.New = func() any { return newFn() }
	return pool
}

// Get selects an arbitrary item from the pool, removes it, and returns it
// to the caller.  If the pool is empty, Get calls the constructor supplied
// to NewPool or SetNew.  If neither has been called, Get returns the zero
// value of T rather than panicking on a nil type assertion.
func (p *Pool[T]) Get() T {
	if v, ok := p.p.Get().(T); ok {
		return v
	}
	var zero T
	return zero
}

// Put adds v to the pool.
func (p *Pool[T]) Put(v T) {
	p.p.Put(v)
}

// SetNew configures the constructor used when the pool is empty.  Use this
// instead of NewPool when the constructor needs to close over the pool
// variable itself (e.g. to return objects to this pool via Put), which would
// otherwise create a package-level initialisation cycle.
func (p *Pool[T]) SetNew(newFn func() T) {
	p.p.New = func() any { return newFn() }
}
