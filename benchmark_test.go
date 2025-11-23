package quaternary

import (
	"fmt"
	"testing"
	v1 "github.com/neurlang/quaternary/v1"
)

// BenchmarkV0Lookup benchmarks the v0 (root) implementation for integer to boolean lookups
func BenchmarkV0Lookup(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	
	for _, n := range sizes {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			// Setup: create map with integers 0-N mapped to whether they are odd
			m := make(map[int]bool, n)
			for i := 0; i < n; i++ {
				m[i] = i%2 == 1
			}
			
			// Create filter
			filter := Make(m)
			
			// Benchmark lookups
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				key := i % n
				_ = filter.GetInt(key)
			}
		})
	}
}

// BenchmarkV1Lookup benchmarks the v1 implementation for integer to boolean lookups
func BenchmarkV1Lookup(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	
	for _, n := range sizes {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			// Setup: create map with integers 0-N mapped to whether they are odd
			m := make(map[int]bool, n)
			for i := 0; i < n; i++ {
				m[i] = i%2 == 1
			}
			
			// Create filter
			filter := v1.Make(m, 1)
			
			// Benchmark lookups
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				key := i % n
				_ = v1.GetBool(filter, key)
			}
		})
	}
}

// BenchmarkV0Creation benchmarks the v0 filter creation time
func BenchmarkV0Creation(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	
	for _, n := range sizes {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			// Setup: create map with integers 0-N mapped to whether they are odd
			m := make(map[int]bool, n)
			for i := 0; i < n; i++ {
				m[i] = i%2 == 1
			}
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				_ = Make(m)
			}
		})
	}
}

// BenchmarkV1Creation benchmarks the v1 filter creation time
func BenchmarkV1Creation(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	
	for _, n := range sizes {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			// Setup: create map with integers 0-N mapped to whether they are odd
			m := make(map[int]bool, n)
			for i := 0; i < n; i++ {
				m[i] = i%2 == 1
			}
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				_ = v1.Make(m, 1)
			}
		})
	}
}

// BenchmarkV1LookupOptimized benchmarks the optimized v1 implementation for integer to boolean lookups
func BenchmarkV1LookupOptimized(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			// Setup: create map with integers 0-N mapped to whether they are odd
			m := make(map[int]bool, n)
			for i := 0; i < n; i++ {
				m[i] = i%2 == 1
			}

			// Create filter
			filter := v1.Make(m, 1)

			// Benchmark lookups
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				key := i % n
				_ = v1.GetBoolInt(filter, key)
			}
		})
	}
}
