// Package each provides per-element slice operations that Go's stdlib
// skipped.
//
// Find, Filter, GroupBy, KeyBy, Partition, Count, and Every all take a
// slice and a predicate or key-function, and return a concrete result
// without mutating the input. Every operation is nil-safe, never panics
// on valid input, and allocates new slices or maps for its output.
//
// This package deliberately does NOT include Map, Reduce, Fold, or
// FlatMap. A two-line for-loop in Go is usually clearer than a lambda,
// and the stdlib has no slices.Map for the same reason: Go's generics
// do not allow method-level type parameters, making idiomatic use of
// a generic Map function awkward.
//
// # Concurrency
//
// All functions are safe to call concurrently on the same input slice
// as long as the input slice is not modified during the call. They
// neither read from nor write to package-level state.
//
// # Panic safety from non-comparable interface values
//
// GroupBy and KeyBy use the key-function's result as a map key. If the
// caller supplies a key function that returns a non-comparable dynamic
// type (e.g., a slice stored in an any), Go's map implementation will
// panic at runtime. each does not recover from these panics — it is the
// caller's responsibility to ensure the key function returns a
// comparable value.
//
// # Nil predicates
//
// A nil predicate or key function will panic with a nil pointer
// dereference on the first element, matching the convention of
// slices.IndexFunc and similar stdlib helpers. Empty or nil input
// slices do not invoke the predicate at all and so are safe even
// with a nil function value.
//
// # Large value types
//
// Find returns T by value, which copies the element. For large value
// types (e.g., structs containing arrays or many fields), callers who
// want to avoid the copy should use []*T instead of []T so Find returns
// a pointer.
//
// For documentation and examples, see https://github.com/bold-minds/each.
package each

// Find returns the first element of s for which pred returns true, and true.
// Returns (zero, false) if no element matches, or if s is nil/empty.
func Find[T any](s []T, pred func(T) bool) (T, bool) {
	for _, v := range s {
		if pred(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// Filter returns a new slice containing all elements of s for which pred
// returns true. The input slice is not modified. Returns a non-nil empty
// slice if no elements match or if s is nil/empty.
//
// Filter pre-allocates a result slice with capacity equal to len(s) so
// it never has to grow. For highly selective predicates on very large
// inputs (e.g., 0.1% match rate on a million-element slice) this may
// reserve more memory than strictly necessary; callers who care can
// pass the result through slices.Clip.
func Filter[T any](s []T, pred func(T) bool) []T {
	result := make([]T, 0, len(s))
	for _, v := range s {
		if pred(v) {
			result = append(result, v)
		}
	}
	return result
}

// GroupBy groups the elements of s by the key returned by keyFn, returning
// a map from key to a slice of matching elements. Elements within each
// group preserve their relative order from the input slice.
//
// Returns a non-nil empty map for nil or empty input.
func GroupBy[T any, K comparable](s []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range s {
		k := keyFn(v)
		result[k] = append(result[k], v)
	}
	return result
}

// KeyBy indexes the elements of s by the key returned by keyFn. If two
// elements produce the same key, the later element overwrites the earlier
// (last-wins). Use GroupBy if you need to preserve all elements per key.
//
// Returns a non-nil empty map for nil or empty input.
func KeyBy[T any, K comparable](s []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T, len(s))
	for _, v := range s {
		result[keyFn(v)] = v
	}
	return result
}

// Partition splits s into two slices: elements for which pred returns
// true, and elements for which it returns false. Both returned slices
// are non-nil, even if empty. Elements preserve their relative order
// from the input within each output slice.
//
// Partition iterates s exactly once and calls pred once per element,
// which is cheaper in CPU than calling Filter twice with opposite
// predicates. It does pre-allocate len(s) capacity for each of the two
// output slices (2×len(s) total reserved capacity) so that neither
// half has to grow; callers with strongly skewed splits on very large
// inputs may prefer two Filter calls or a manual loop if peak memory
// matters more than iteration count.
func Partition[T any](s []T, pred func(T) bool) (matched, unmatched []T) {
	matched = make([]T, 0, len(s))
	unmatched = make([]T, 0, len(s))
	for _, v := range s {
		if pred(v) {
			matched = append(matched, v)
		} else {
			unmatched = append(unmatched, v)
		}
	}
	return matched, unmatched
}

// Count returns the number of elements in s for which pred returns true.
// Returns 0 for nil or empty input.
func Count[T any](s []T, pred func(T) bool) int {
	n := 0
	for _, v := range s {
		if pred(v) {
			n++
		}
	}
	return n
}

// Every returns true if pred returns true for every element of s, or if
// s is nil/empty (vacuously true). Short-circuits on the first false,
// so pred is called at most once per element and is never called at
// all for nil or empty input.
func Every[T any](s []T, pred func(T) bool) bool {
	for _, v := range s {
		if !pred(v) {
			return false
		}
	}
	return true
}
