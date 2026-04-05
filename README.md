# each

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/bold-minds/each.svg)](https://pkg.go.dev/github.com/bold-minds/each)
[![Go Version](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/bold-minds/each/main/.github/badges/go-version.json)](https://golang.org/doc/go1.21)
[![Latest Release](https://img.shields.io/github/v/release/bold-minds/each?logo=github&color=blueviolet)](https://github.com/bold-minds/each/releases)
[![Last Updated](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/bold-minds/each/main/.github/badges/last-updated.json)](https://github.com/bold-minds/each/commits)
[![golangci-lint](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/bold-minds/each/main/.github/badges/golangci-lint.json)](https://github.com/bold-minds/each/actions/workflows/test.yaml)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/bold-minds/each/main/.github/badges/coverage.json)](https://github.com/bold-minds/each/actions/workflows/test.yaml)
[![Dependabot](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/bold-minds/each/main/.github/badges/dependabot.json)](https://github.com/bold-minds/each/security/dependabot)

**Find, filter, group — slice operations Go stdlib skipped.**

Go's `slices` package (1.21+) covers sorting, searching, and mutation, but it deliberately omits the predicate-and-key-function operations that show up constantly in real code: find by field, group by category, partition active vs inactive, count matching elements. `each` fills exactly those gaps with seven one-line operations.

```go
// Before — typical Go pattern: find element by ID
index := -1
found := false
for i, env := range envs {
    if env.Id == envID {
        index = i
        found = true
        break
    }
}
if !found {
    return fmt.Errorf("env not found")
}
env := envs[index]

// After
env, ok := each.Find(envs, func(e *Env) bool { return e.Id == envID })
if !ok {
    return fmt.Errorf("env not found")
}
```

## ✨ Why each?

- 🔍 **`Find` returns the value**, not just an index. `slices.IndexFunc` gives you an `int` and then you bounds-check and slice yourself. `each.Find` gives you `(T, bool)` in one call.
- 🗂️ **`GroupBy` and `KeyBy`** — the reshape operations stdlib lacks entirely
- ✂️ **Non-mutating `Filter`** — `slices.DeleteFunc` is in-place and inverse; `each.Filter` returns a new slice of elements matching your predicate
- 🪓 **`Partition`** splits a slice by predicate in one call — no two-pass filter dance
- 🔢 **`Count` and `Every`** for predicate-based queries stdlib doesn't cover
- 🚫 **No `Map`, no `Reduce`** — deliberate non-goal; Go's for-range loop is usually clearer for transformations
- 🪶 **Seven functions, one file, zero dependencies** — only what stdlib genuinely skipped

## 📦 Installation

```bash
go get github.com/bold-minds/each
```

Requires Go 1.21 or later.

## 🚀 Quick Start

```go
package main

import (
    "fmt"

    "github.com/bold-minds/each"
)

type User struct {
    ID     int
    Name   string
    Role   string
    Active bool
}

func main() {
    users := []User{
        {ID: 1, Name: "alice", Role: "admin", Active: true},
        {ID: 2, Name: "bob",   Role: "editor", Active: true},
        {ID: 3, Name: "carol", Role: "admin", Active: false},
        {ID: 4, Name: "dave",  Role: "editor", Active: true},
    }

    // Find by predicate — always check the ok return
    alice, ok := each.Find(users, func(u User) bool { return u.Name == "alice" })
    if !ok {
        panic("alice missing")
    }
    fmt.Println(alice.Role) // "admin"

    // Filter to a new slice (non-mutating)
    active := each.Filter(users, func(u User) bool { return u.Active })
    fmt.Println(len(active)) // 3

    // Group by role
    byRole := each.GroupBy(users, func(u User) string { return u.Role })
    fmt.Println(len(byRole["admin"])) // 2

    // Index by ID
    byID := each.KeyBy(users, func(u User) int { return u.ID })
    fmt.Println(byID[3].Name) // "carol"

    // Partition into active and inactive
    on, off := each.Partition(users, func(u User) bool { return u.Active })
    fmt.Println(len(on), len(off)) // 3 1

    // Count matches
    admins := each.Count(users, func(u User) bool { return u.Role == "admin" })
    fmt.Println(admins) // 2

    // All must match
    allNamed := each.Every(users, func(u User) bool { return u.Name != "" })
    fmt.Println(allNamed) // true
}
```

## 🔧 Core Features

### `Find` — first matching element

Returns the first element where `pred` returns true, along with a `bool` indicating whether any match was found. Unlike `slices.IndexFunc`, you get the value directly — no manual bounds check + slice access.

```go
env, ok := each.Find(envs, func(e *Env) bool { return e.Id == envID })
if !ok {
    return ErrNotFound
}
// use env directly
```

### `Filter` — new slice of matching elements

Returns a new slice containing elements for which `pred` returns true. Non-mutating — the input slice is never modified.

```go
active := each.Filter(users, func(u User) bool { return u.Active })
errors := each.Filter(logs, func(l LogLine) bool { return l.Level == "error" })
recent := each.Filter(events, func(e Event) bool { return e.Time.After(cutoff) })
```

Returns an empty (non-nil) slice if nothing matches. Safe to range over without a nil check.

### `GroupBy` — slice to `map[K][]T`

Groups elements by the key returned by `keyFn`. Every element appears in exactly one group. Elements within a group preserve their relative order from the input.

```go
byRole := each.GroupBy(users, func(u User) string { return u.Role })
// map["admin"][User, User], "editor"[User, User]]

byDate := each.GroupBy(events, func(e Event) string { return e.Date.Format("2006-01-02") })
```

Common for aggregation, reporting, and histogram-style analysis.

### `KeyBy` — slice to `map[K]T`

Indexes elements by the key returned by `keyFn`. Each key maps to exactly one element — **if two elements produce the same key, the later one wins.** Use `GroupBy` if you need all elements per key.

```go
byID := each.KeyBy(users, func(u User) int { return u.ID })
alice := byID[1] // direct lookup
```

The common pattern for "I have a slice but I want O(1) lookup by field."

### `Partition` — split by predicate into two slices

Splits `s` into two slices: one containing elements matching `pred`, one containing elements that don't. Iterates `s` exactly once and calls `pred` once per element, which is cheaper in CPU than two `Filter` calls with opposite predicates. It does pre-allocate `len(s)` capacity for each half (so the total reserved capacity is `2×len(s)`); for strongly skewed splits on very large inputs, two `Filter` calls or a manual loop may use less peak memory.

```go
active, inactive := each.Partition(users, func(u User) bool { return u.Active })
valid, invalid  := each.Partition(records, isValid)
```

Each returned slice is non-nil, even if empty.

### `Count` — how many match

Returns the number of elements for which `pred` returns true. Stdlib has no direct equivalent.

```go
errors  := each.Count(logs, func(l LogLine) bool { return l.Level == "error" })
admins  := each.Count(users, func(u User) bool { return u.Role == "admin" })
overdue := each.Count(tasks, func(t Task) bool { return t.Due.Before(now) })
```

### `Every` — all must match

Returns `true` if `pred` returns true for every element, or if the slice is empty. Short-circuits on the first false.

```go
allValid := each.Every(records, isValid)
allPaid  := each.Every(invoices, func(i Invoice) bool { return i.PaidAt != nil })

// Empty slice is vacuously true — common convention
each.Every([]int{}, func(int) bool { return false }) // true
```

For the "any element matches" case, use [`slices.ContainsFunc`](https://pkg.go.dev/slices#ContainsFunc) from stdlib — `each` doesn't duplicate it.

## 🚫 What's deliberately NOT here

`each` does **not** include `Map`, `Reduce`, `Fold`, `FlatMap`, `Zip`, or other functional-programming primitives. This is a deliberate scope decision.

- **`Map`** (transform each element to a different type) is cleaner as a `for`-loop in Go. The stdlib has no `slices.Map` for the same reason: Go's generics don't support method-level type parameters, and the two-line loop is clearer at call sites than a generic helper.
- **`Reduce` / `Fold`** requires you to name an accumulator type, a reducer function, and an initial value — three things to get right. A loop with a named variable is easier to read.
- **`FindIndex`** duplicates [`slices.IndexFunc`](https://pkg.go.dev/slices#IndexFunc). Use stdlib.
- **`Any`** duplicates [`slices.ContainsFunc`](https://pkg.go.dev/slices#ContainsFunc). Use stdlib.

If you want these, [`samber/lo`](https://github.com/samber/lo) has all of them. `each` is scoped to the seven operations that are hit constantly in real code, are not well-served by stdlib, and read better as one call than as a loop.

## 🛡️ Safety guarantees

- **Never panics on valid input.** Nil slices, empty slices, and empty variadic calls all return non-nil empty slices or maps.
- **Immutable.** `each` never modifies input slices. `Filter`, `Partition`, `GroupBy`, and `KeyBy` allocate new slices or maps.
- **Nil-safe.** All functions accept `nil` slices and return sensible defaults: empty slices for `Filter`/`Partition`, empty maps for `GroupBy`/`KeyBy`, zero values for `Find`/`Count`, `true` for `Every`.
- **Zero dependencies.** Pure stdlib.
- **No reflection.** Pure generic functions.

### Non-comparable key values

`GroupBy` and `KeyBy` use the key function's return value as a map key. If the key function returns a non-comparable dynamic type (e.g., a slice stored in an `any`), Go's map implementation will panic at runtime. `each` does not recover from these panics — it is the caller's responsibility to ensure the key function returns a comparable value.

## 🏎️ Performance

Measured on Go 1.26 (Intel Ultra 9 275HX) on 1000-element slices. The library targets Go 1.21+; generics codegen has improved across minor releases, so these numbers are upper bounds for current Go and likely slower on older toolchains.

```
BenchmarkFind_Hit-24           773994     856.5 ns/op        0 B/op     0 allocs/op
BenchmarkFind_Miss-24          342858    1517   ns/op        0 B/op     0 allocs/op
BenchmarkFilter_Half-24         60860   11493   ns/op    49152 B/op     1 allocs/op
BenchmarkGroupBy-24             14418   52569   ns/op   154275 B/op    20 allocs/op
BenchmarkKeyBy-24               10000   61584   ns/op   131155 B/op     5 allocs/op
BenchmarkPartition-24           20832   25655   ns/op    98305 B/op     2 allocs/op
BenchmarkCount-24              291494    2327   ns/op        0 B/op     0 allocs/op
BenchmarkEvery_AllTrue-24      290486    1852   ns/op        0 B/op     0 allocs/op
BenchmarkEvery_EarlyFail-24  377931465       1.6 ns/op        0 B/op     0 allocs/op
```

`Find`, `Count`, and `Every` are zero-allocation (they only need a loop counter or local variable). `Every` short-circuits on the first false — the 1.6ns result shows that when the predicate fails on element 0, evaluation stops immediately.

`Filter`, `GroupBy`, `KeyBy`, and `Partition` allocate because they must return new collections. The allocation counts are minimal: one slice for `Filter`, two for `Partition`, and map-growth allocations for the reshape operations.

## 🧪 Testing

```bash
go test ./...                      # unit tests
go test -race ./...                # race detection
go test -bench=. -benchmem ./...   # benchmarks
```

Current coverage: 100%.

## 📚 API Reference

```go
// Find returns the first element of s for which pred returns true, and true.
// If no element matches, returns (zero, false).
func Find[T any](s []T, pred func(T) bool) (T, bool)

// Filter returns a new slice containing all elements of s for which pred
// returns true. The input slice is not modified.
func Filter[T any](s []T, pred func(T) bool) []T

// GroupBy groups slice elements into a map keyed by keyFn's result.
// Each key maps to a slice of matching elements in original order.
func GroupBy[T any, K comparable](s []T, keyFn func(T) K) map[K][]T

// KeyBy indexes slice elements into a map keyed by keyFn's result.
// If multiple elements produce the same key, the later element wins.
func KeyBy[T any, K comparable](s []T, keyFn func(T) K) map[K]T

// Partition splits s into two slices: elements matching pred, and
// elements not matching. Done in a single pass.
func Partition[T any](s []T, pred func(T) bool) (matched, unmatched []T)

// Count returns the number of elements in s for which pred returns true.
func Count[T any](s []T, pred func(T) bool) int

// Every returns true if pred returns true for every element of s.
// Returns true for an empty slice (vacuously true). Short-circuits.
func Every[T any](s []T, pred func(T) bool) bool
```

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Bold Minds Go libraries follow a shared set of design principles; read [PRINCIPLES.md](https://github.com/bold-minds/oss/blob/main/PRINCIPLES.md) before opening a PR.

## 📄 License

MIT. See [LICENSE](LICENSE).

## 🔗 Related Projects

- Go standard library [`slices`](https://pkg.go.dev/slices) — the mechanical foundation. `each` complements `slices`, it doesn't replace it. Use `slices.IndexFunc` for "give me the index," `slices.ContainsFunc` for "is there any match," `slices.Sort`, `slices.Reverse`, etc.
- [`bold-minds/list`](https://github.com/bold-minds/list) — set operations on slices (Unique, Union, Intersect, Minus, Without). `each` handles per-element predicates and key-function ops; `list` handles multi-slice set semantics.
- [`bold-minds/dig`](https://github.com/bold-minds/dig) — nested data navigation. Common pattern: dig out a slice, then filter/group it with `each`.
- [`samber/lo`](https://github.com/samber/lo) — comprehensive Go utility library with ~200 helpers. `each` is a focused subset: only the per-element predicate and reshape operations, only what stdlib genuinely lacks.
