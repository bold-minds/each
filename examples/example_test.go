package examples_test

import (
	"fmt"
	"sort"

	"github.com/bold-minds/each"
)

type User struct {
	ID     int
	Name   string
	Role   string
	Active bool
}

var users = []User{
	{ID: 1, Name: "alice", Role: "admin", Active: true},
	{ID: 2, Name: "bob", Role: "editor", Active: true},
	{ID: 3, Name: "carol", Role: "admin", Active: false},
}

func ExampleFind() {
	alice, ok := each.Find(users, func(u User) bool { return u.Name == "alice" })
	fmt.Println(alice.Role, ok)
	// Output: admin true
}

func ExampleFind_miss() {
	_, ok := each.Find(users, func(u User) bool { return u.Name == "eve" })
	fmt.Println(ok)
	// Output: false
}

func ExampleFilter() {
	active := each.Filter(users, func(u User) bool { return u.Active })
	for _, u := range active {
		fmt.Println(u.Name)
	}
	// Output:
	// alice
	// bob
}

func ExampleGroupBy() {
	byRole := each.GroupBy(users, func(u User) string { return u.Role })
	// Sort keys for deterministic output.
	keys := make([]string, 0, len(byRole))
	for k := range byRole {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s: %d\n", k, len(byRole[k]))
	}
	// Output:
	// admin: 2
	// editor: 1
}

func ExampleKeyBy() {
	byID := each.KeyBy(users, func(u User) int { return u.ID })
	fmt.Println(byID[2].Name)
	// Output: bob
}

func ExamplePartition() {
	active, inactive := each.Partition(users, func(u User) bool { return u.Active })
	fmt.Printf("active=%d inactive=%d\n", len(active), len(inactive))
	// Output: active=2 inactive=1
}

func ExampleCount() {
	admins := each.Count(users, func(u User) bool { return u.Role == "admin" })
	fmt.Println(admins)
	// Output: 2
}

func ExampleEvery() {
	allNamed := each.Every(users, func(u User) bool { return u.Name != "" })
	fmt.Println(allNamed)
	// Output: true
}

func ExampleEvery_empty() {
	// Vacuous truth: all-of-nothing is true
	fmt.Println(each.Every([]int{}, func(int) bool { return false }))
	// Output: true
}
