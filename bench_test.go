package each_test

import (
	"testing"

	"github.com/bold-minds/each"
)

var benchUsers = func() []User {
	out := make([]User, 0, 1000)
	for i := 0; i < 1000; i++ {
		role := "editor"
		if i%5 == 0 {
			role = "admin"
		}
		out = append(out, User{
			ID:     i,
			Name:   "user",
			Role:   role,
			Active: i%3 != 0,
		})
	}
	return out
}()

func BenchmarkFind_Hit(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = each.Find(benchUsers, func(u User) bool { return u.ID == 500 })
	}
}

func BenchmarkFind_Miss(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = each.Find(benchUsers, func(u User) bool { return u.ID == -1 })
	}
}

func BenchmarkFilter_Half(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = each.Filter(benchUsers, func(u User) bool { return u.Active })
	}
}

func BenchmarkGroupBy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = each.GroupBy(benchUsers, func(u User) string { return u.Role })
	}
}

func BenchmarkKeyBy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = each.KeyBy(benchUsers, func(u User) int { return u.ID })
	}
}

func BenchmarkPartition(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = each.Partition(benchUsers, func(u User) bool { return u.Active })
	}
}

func BenchmarkCount(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = each.Count(benchUsers, func(u User) bool { return u.Role == "admin" })
	}
}

func BenchmarkEvery_AllTrue(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = each.Every(benchUsers, func(u User) bool { return u.ID >= 0 })
	}
}

func BenchmarkEvery_EarlyFail(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = each.Every(benchUsers, func(u User) bool { return u.ID > 500 })
	}
}
