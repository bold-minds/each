# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] — Initial release

### Added
- `Find[T any](s []T, pred func(T) bool) (T, bool)` — first matching element, returns value + bool
- `Filter[T any](s []T, pred func(T) bool) []T` — new slice of matching elements, non-mutating
- `GroupBy[T any, K comparable](s []T, keyFn func(T) K) map[K][]T` — group slice elements by key
- `KeyBy[T any, K comparable](s []T, keyFn func(T) K) map[K]T` — index slice elements by key (last-wins)
- `Partition[T any](s []T, pred func(T) bool) (matched, unmatched []T)` — split by predicate in one pass
- `Count[T any](s []T, pred func(T) bool) int` — count matching elements
- `Every[T any](s []T, pred func(T) bool) bool` — all-match with short-circuit (vacuously true for empty)
- Adversarial test coverage verifying nil-return guarantees, immutability, result non-aliasing, custom comparable key types, struct keys, stateful predicates, and short-circuit evaluation
- Integration tests mirroring real-world patterns (find-by-ID, log analysis, role grouping)
- 100% test coverage
- Zero external dependencies — pure stdlib

### Deliberate non-goals
- No `Map`, `Reduce`, `Fold`, or `FlatMap` — use a Go for-loop
- No `FindIndex` — stdlib `slices.IndexFunc` covers this
- No `Any` — stdlib `slices.ContainsFunc` covers this
- No set operations (those live in [`bold-minds/list`](https://github.com/bold-minds/list))
- No iterator (`iter.Seq`) support in v1 — works on `[]T` only
- No in-place mutation — every function is non-mutating

### Requires
- Go 1.21 or later
