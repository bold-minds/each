package each_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/bold-minds/each"
)

// Test fixture: a realistic domain type that appears throughout the tests.
type User struct {
	ID     int
	Name   string
	Role   string
	Active bool
}

var testUsers = []User{
	{ID: 1, Name: "alice", Role: "admin", Active: true},
	{ID: 2, Name: "bob", Role: "editor", Active: true},
	{ID: 3, Name: "carol", Role: "admin", Active: false},
	{ID: 4, Name: "dave", Role: "editor", Active: true},
}

// =============================================================================
// Find
// =============================================================================

func TestFind_Hit(t *testing.T) {
	got, ok := each.Find(testUsers, func(u User) bool { return u.Name == "bob" })
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got.ID != 2 {
		t.Errorf("got ID=%d, want 2", got.ID)
	}
}

func TestFind_FirstMatchWins(t *testing.T) {
	got, ok := each.Find(testUsers, func(u User) bool { return u.Role == "admin" })
	if !ok || got.Name != "alice" {
		t.Errorf("got (%+v, %v), want first admin (alice, true)", got, ok)
	}
}

func TestFind_Miss(t *testing.T) {
	_, ok := each.Find(testUsers, func(u User) bool { return u.Name == "eve" })
	if ok {
		t.Error("expected ok=false for no match")
	}
}

func TestFind_EmptySlice(t *testing.T) {
	_, ok := each.Find([]User{}, func(u User) bool { return true })
	if ok {
		t.Error("expected ok=false for empty slice")
	}
}

func TestFind_NilSlice(t *testing.T) {
	_, ok := each.Find[User](nil, func(u User) bool { return true })
	if ok {
		t.Error("expected ok=false for nil slice")
	}
}

func TestFind_ReturnsValueNotPointer(t *testing.T) {
	// Find returns a copy of the element, not a pointer into the slice.
	// Mutating the returned value must not affect the source slice.
	users := []User{{ID: 1, Name: "alice"}}
	got, _ := each.Find(users, func(u User) bool { return u.ID == 1 })
	got.Name = "mutated"
	if got.Name != "mutated" {
		t.Error("local copy was not modified (unexpected)")
	}
	if users[0].Name != "alice" {
		t.Errorf("source slice was mutated: %q", users[0].Name)
	}
}

// =============================================================================
// Filter
// =============================================================================

func TestFilter_Basic(t *testing.T) {
	active := each.Filter(testUsers, func(u User) bool { return u.Active })
	if len(active) != 3 {
		t.Errorf("got %d active users, want 3", len(active))
	}
}

func TestFilter_NoMatch(t *testing.T) {
	got := each.Filter(testUsers, func(u User) bool { return u.Name == "eve" })
	if got == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("got len %d, want 0", len(got))
	}
}

func TestFilter_AllMatch(t *testing.T) {
	got := each.Filter(testUsers, func(u User) bool { return u.ID > 0 })
	if len(got) != len(testUsers) {
		t.Errorf("got %d, want %d", len(got), len(testUsers))
	}
}

func TestFilter_NilSlice(t *testing.T) {
	got := each.Filter[int](nil, func(v int) bool { return true })
	if got == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(got) != 0 {
		t.Errorf("got len %d, want 0", len(got))
	}
}

func TestFilter_EmptySlice(t *testing.T) {
	got := each.Filter([]int{}, func(v int) bool { return true })
	if got == nil {
		t.Error("expected non-nil empty slice")
	}
}

// =============================================================================
// GroupBy
// =============================================================================

func TestGroupBy_Basic(t *testing.T) {
	byRole := each.GroupBy(testUsers, func(u User) string { return u.Role })
	if len(byRole["admin"]) != 2 {
		t.Errorf("got %d admins, want 2", len(byRole["admin"]))
	}
	if len(byRole["editor"]) != 2 {
		t.Errorf("got %d editors, want 2", len(byRole["editor"]))
	}
}

func TestGroupBy_PreservesOrderWithinGroup(t *testing.T) {
	byRole := each.GroupBy(testUsers, func(u User) string { return u.Role })
	admins := byRole["admin"]
	if admins[0].ID != 1 || admins[1].ID != 3 {
		t.Errorf("admin order wrong: %v", admins)
	}
}

func TestGroupBy_NilSlice(t *testing.T) {
	got := each.GroupBy[int, string](nil, func(v int) string { return "" })
	if got == nil {
		t.Error("expected non-nil empty map")
	}
	if len(got) != 0 {
		t.Errorf("got %d keys, want 0", len(got))
	}
}

func TestGroupBy_EmptySlice(t *testing.T) {
	got := each.GroupBy([]int{}, func(v int) string { return "" })
	if got == nil || len(got) != 0 {
		t.Errorf("got %v, want non-nil empty map", got)
	}
}

func TestGroupBy_IntKey(t *testing.T) {
	byID := each.GroupBy(testUsers, func(u User) int { return u.ID })
	for id := 1; id <= 4; id++ {
		if len(byID[id]) != 1 {
			t.Errorf("id %d: got %d, want 1", id, len(byID[id]))
		}
	}
}

// =============================================================================
// KeyBy
// =============================================================================

func TestKeyBy_Basic(t *testing.T) {
	byID := each.KeyBy(testUsers, func(u User) int { return u.ID })
	if len(byID) != 4 {
		t.Errorf("got %d keys, want 4", len(byID))
	}
	if byID[3].Name != "carol" {
		t.Errorf("got %q, want carol", byID[3].Name)
	}
}

func TestKeyBy_LastWinsOnDuplicateKey(t *testing.T) {
	// Duplicate keys: last element wins
	dupes := []User{
		{ID: 1, Name: "first"},
		{ID: 2, Name: "second"},
		{ID: 1, Name: "third"}, // same ID as first
	}
	byID := each.KeyBy(dupes, func(u User) int { return u.ID })
	if len(byID) != 2 {
		t.Errorf("got %d keys, want 2 (duplicates collapse)", len(byID))
	}
	if byID[1].Name != "third" {
		t.Errorf("got %q, want third (last-wins)", byID[1].Name)
	}
}

func TestKeyBy_NilSlice(t *testing.T) {
	got := each.KeyBy[int, string](nil, func(v int) string { return "" })
	if got == nil || len(got) != 0 {
		t.Errorf("got %v, want non-nil empty map", got)
	}
}

// =============================================================================
// Partition
// =============================================================================

func TestPartition_Basic(t *testing.T) {
	active, inactive := each.Partition(testUsers, func(u User) bool { return u.Active })
	if len(active) != 3 {
		t.Errorf("got %d active, want 3", len(active))
	}
	if len(inactive) != 1 {
		t.Errorf("got %d inactive, want 1", len(inactive))
	}
}

func TestPartition_AllMatch(t *testing.T) {
	yes, no := each.Partition([]int{1, 2, 3}, func(v int) bool { return v > 0 })
	if len(yes) != 3 {
		t.Errorf("got %d matched, want 3", len(yes))
	}
	if no == nil {
		t.Error("unmatched must be non-nil even when empty")
	}
	if len(no) != 0 {
		t.Errorf("got %d unmatched, want 0", len(no))
	}
}

func TestPartition_NoneMatch(t *testing.T) {
	yes, no := each.Partition([]int{1, 2, 3}, func(v int) bool { return v > 100 })
	if yes == nil {
		t.Error("matched must be non-nil even when empty")
	}
	if len(yes) != 0 {
		t.Errorf("got %d matched, want 0", len(yes))
	}
	if len(no) != 3 {
		t.Errorf("got %d unmatched, want 3", len(no))
	}
}

func TestPartition_NilSlice(t *testing.T) {
	yes, no := each.Partition[int](nil, func(v int) bool { return true })
	if yes == nil || no == nil {
		t.Error("Partition must return non-nil slices even for nil input")
	}
	if len(yes) != 0 || len(no) != 0 {
		t.Errorf("got (%d, %d), want (0, 0)", len(yes), len(no))
	}
}

// =============================================================================
// Count
// =============================================================================

func TestCount_Basic(t *testing.T) {
	if got := each.Count(testUsers, func(u User) bool { return u.Active }); got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

func TestCount_NoMatch(t *testing.T) {
	if got := each.Count(testUsers, func(u User) bool { return false }); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestCount_AllMatch(t *testing.T) {
	if got := each.Count(testUsers, func(u User) bool { return true }); got != 4 {
		t.Errorf("got %d, want 4", got)
	}
}

func TestCount_NilSlice(t *testing.T) {
	if got := each.Count[int](nil, func(v int) bool { return true }); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

// =============================================================================
// Every
// =============================================================================

func TestEvery_AllMatch(t *testing.T) {
	got := each.Every(testUsers, func(u User) bool { return u.ID > 0 })
	if !got {
		t.Error("expected true")
	}
}

func TestEvery_OneFails(t *testing.T) {
	got := each.Every(testUsers, func(u User) bool { return u.Active })
	if got {
		t.Error("expected false (carol is inactive)")
	}
}

func TestEvery_EmptySlice(t *testing.T) {
	// Vacuous truth: all-of-nothing is true.
	got := each.Every([]int{}, func(v int) bool { return false })
	if !got {
		t.Error("expected true for empty slice (vacuous truth)")
	}
}

func TestEvery_NilSlice(t *testing.T) {
	got := each.Every[int](nil, func(v int) bool { return false })
	if !got {
		t.Error("expected true for nil slice (vacuous truth)")
	}
}

func TestEvery_ShortCircuits(t *testing.T) {
	// Counter ensures Every stops evaluating as soon as a false is seen.
	calls := 0
	each.Every([]int{1, 2, 3, 4, 5}, func(v int) bool {
		calls++
		return v < 3
	})
	if calls != 3 {
		t.Errorf("got %d predicate calls, want 3 (should short-circuit on first false)", calls)
	}
}

func TestEvery_PredicateNotCalledOnEmpty(t *testing.T) {
	// Documented invariant: Every returns true on empty/nil input
	// without invoking the predicate at all.
	for _, tc := range []struct {
		name string
		in   []int
	}{
		{"nil", nil},
		{"empty", []int{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			got := each.Every(tc.in, func(int) bool {
				calls++
				return false
			})
			if !got {
				t.Errorf("Every(%s) = false, want true (vacuous truth)", tc.name)
			}
			if calls != 0 {
				t.Errorf("Every(%s) called predicate %d times, want 0", tc.name, calls)
			}
		})
	}
}

// =============================================================================
// Adversarial — edge cases the to-review lessons told me to test
// =============================================================================

// TestNilReturnGuarantees verifies every function returns non-nil slices
// or non-nil maps for nil/empty input where documented.
func TestNilReturnGuarantees(t *testing.T) {
	// Filter
	if got := each.Filter[int](nil, func(v int) bool { return true }); got == nil {
		t.Error("Filter(nil) returned nil")
	}
	if got := each.Filter([]int{}, func(v int) bool { return true }); got == nil {
		t.Error("Filter([]) returned nil")
	}

	// GroupBy
	if got := each.GroupBy[int, int](nil, func(v int) int { return v }); got == nil {
		t.Error("GroupBy(nil) returned nil")
	}

	// KeyBy
	if got := each.KeyBy[int, int](nil, func(v int) int { return v }); got == nil {
		t.Error("KeyBy(nil) returned nil")
	}

	// Partition — BOTH halves must be non-nil
	matched, unmatched := each.Partition[int](nil, func(v int) bool { return true })
	if matched == nil {
		t.Error("Partition matched is nil")
	}
	if unmatched == nil {
		t.Error("Partition unmatched is nil")
	}
}

// TestImmutability verifies that functions never mutate their inputs.
func TestImmutability(t *testing.T) {
	// Filter
	in := []int{1, 2, 3, 4, 5}
	snapshot := append([]int{}, in...)
	_ = each.Filter(in, func(v int) bool { return v%2 == 0 })
	if !reflect.DeepEqual(in, snapshot) {
		t.Errorf("Filter mutated input: %v vs %v", in, snapshot)
	}

	// GroupBy
	in2 := []User{{ID: 1}, {ID: 2}}
	snap2 := append([]User{}, in2...)
	_ = each.GroupBy(in2, func(u User) int { return u.ID })
	if !reflect.DeepEqual(in2, snap2) {
		t.Error("GroupBy mutated input")
	}

	// KeyBy
	in3 := []User{{ID: 1}, {ID: 2}}
	snap3 := append([]User{}, in3...)
	_ = each.KeyBy(in3, func(u User) int { return u.ID })
	if !reflect.DeepEqual(in3, snap3) {
		t.Error("KeyBy mutated input")
	}

	// Partition
	in4 := []int{1, 2, 3}
	snap4 := append([]int{}, in4...)
	_, _ = each.Partition(in4, func(v int) bool { return v > 1 })
	if !reflect.DeepEqual(in4, snap4) {
		t.Error("Partition mutated input")
	}
}

// TestResultIsNotAliased verifies that mutating the returned slice/map
// does not affect the input slice.
func TestResultIsNotAliased(t *testing.T) {
	t.Run("Filter", func(t *testing.T) {
		in := []User{{ID: 1, Name: "alice"}, {ID: 2, Name: "bob"}}
		filtered := each.Filter(in, func(u User) bool { return true })
		if len(filtered) == 0 {
			t.Fatal("expected non-empty result")
		}
		filtered[0].Name = "mutated"
		if in[0].Name != "alice" {
			t.Errorf("Filter result shares backing storage with input: in[0].Name = %q", in[0].Name)
		}
	})

	t.Run("Partition", func(t *testing.T) {
		in := []User{{ID: 1, Name: "alice"}, {ID: 2, Name: "bob"}}
		matched, unmatched := each.Partition(in, func(u User) bool { return u.ID == 1 })
		if len(matched) == 0 || len(unmatched) == 0 {
			t.Fatal("expected both halves non-empty")
		}
		matched[0].Name = "mutated-matched"
		unmatched[0].Name = "mutated-unmatched"
		if in[0].Name != "alice" || in[1].Name != "bob" {
			t.Errorf("Partition result shares backing storage with input: %+v", in)
		}
	})
}

// TestCustomComparableKeyTypes verifies that GroupBy and KeyBy work with
// named comparable types, not just primitives.
func TestCustomComparableKeyTypes(t *testing.T) {
	type UserID int
	type Role string

	type Record struct {
		ID   UserID
		Role Role
	}
	records := []Record{
		{ID: 1, Role: "admin"},
		{ID: 2, Role: "editor"},
		{ID: 3, Role: "admin"},
	}

	byID := each.KeyBy(records, func(r Record) UserID { return r.ID })
	if len(byID) != 3 {
		t.Errorf("got %d keys, want 3", len(byID))
	}

	byRole := each.GroupBy(records, func(r Record) Role { return r.Role })
	if len(byRole["admin"]) != 2 {
		t.Errorf("got %d admins, want 2", len(byRole["admin"]))
	}
}

// TestStructKey verifies that GroupBy/KeyBy work with struct-typed keys
// (composite keys) as long as the struct is comparable.
func TestStructKey(t *testing.T) {
	type Point struct {
		X, Y int
	}
	type PointVal struct {
		P Point
		V string
	}
	records := []PointVal{
		{Point{0, 0}, "origin"},
		{Point{1, 1}, "diag"},
		{Point{0, 0}, "also origin"},
	}

	byPoint := each.GroupBy(records, func(pv PointVal) Point { return pv.P })
	if len(byPoint) != 2 {
		t.Errorf("got %d groups, want 2", len(byPoint))
	}
	if len(byPoint[Point{0, 0}]) != 2 {
		t.Errorf("got %d records at origin, want 2", len(byPoint[Point{0, 0}]))
	}
}

// TestNonComparableKeyPanics pins the documented behavior that GroupBy
// and KeyBy panic when the key function returns a non-comparable dynamic
// type stored in an `any`. This is surfaced by Go's map implementation —
// each does not recover, and callers must ensure keys are comparable.
func TestNonComparableKeyPanics(t *testing.T) {
	type Row struct {
		Tags []string
	}
	rows := []Row{{Tags: []string{"a"}}, {Tags: []string{"b"}}}

	t.Run("GroupBy", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic from non-comparable key in GroupBy, got none")
			}
		}()
		_ = each.GroupBy(rows, func(r Row) any { return r.Tags })
	})

	t.Run("KeyBy", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic from non-comparable key in KeyBy, got none")
			}
		}()
		_ = each.KeyBy(rows, func(r Row) any { return r.Tags })
	})
}

// TestPredicateCapturesState verifies that a predicate with captured state
// (closure over a counter) works correctly. This is important because
// stateful predicates are a common real-world pattern.
func TestPredicateCapturesState(t *testing.T) {
	// Keep every other element
	i := 0
	got := each.Filter([]int{10, 20, 30, 40, 50}, func(v int) bool {
		keep := i%2 == 0
		i++
		return keep
	})
	want := []int{10, 30, 50}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// =============================================================================
// Integration — realistic patterns
// =============================================================================

func TestIntegration_TraverzerStyleFind(t *testing.T) {
	// Mirrors the real 7-line "find by ID" loop from traverzer-admin's
	// customer.go that motivated each.Find's existence.
	type Env struct {
		ID   string
		Name string
	}
	envs := []Env{
		{ID: "dev-1", Name: "development"},
		{ID: "stg-2", Name: "staging"},
		{ID: "prd-3", Name: "production"},
	}
	env, ok := each.Find(envs, func(e Env) bool { return e.ID == "stg-2" })
	if !ok || env.Name != "staging" {
		t.Errorf("got (%+v, %v), want staging environment", env, ok)
	}
}

func TestIntegration_LogAnalysis(t *testing.T) {
	type LogLine struct {
		Level string
		Msg   string
	}
	logs := []LogLine{
		{"info", "startup"},
		{"error", "connection failed"},
		{"info", "retry"},
		{"error", "timeout"},
		{"warn", "slow query"},
	}

	// Group by level
	byLevel := each.GroupBy(logs, func(l LogLine) string { return l.Level })
	if len(byLevel["error"]) != 2 {
		t.Errorf("got %d errors, want 2", len(byLevel["error"]))
	}

	// Count errors
	errCount := each.Count(logs, func(l LogLine) bool { return l.Level == "error" })
	if errCount != 2 {
		t.Errorf("Count: got %d, want 2", errCount)
	}

	// Any timeout?
	timeout, found := each.Find(logs, func(l LogLine) bool { return strings.Contains(l.Msg, "timeout") })
	if !found || timeout.Level != "error" {
		t.Errorf("got (%+v, %v), want timeout log", timeout, found)
	}
}
