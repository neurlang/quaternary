package quaternary

import (
	"testing"
    	"runtime"
	"fmt"
)

func BenchmarkReadAllQuarternary(b *testing.B) {
	cases := []struct {
		name  string
		input int
	}{
		{"Case1", 10},
		{"Case2", 100},
		{"Case3", 1000},
		{"Case4", 10000},
		{"Case5", 100000},
		{"Case6", 1000000},
	}
	for _, tc := range cases {
		var m = make(map[int]bool)
		for i := 0; i < tc.input; i++ {
			m[i] = false
		}
		for i := 0; i < tc.input; i++ {
			m[tc.input+i] = true
		}
		var f = Make(m)
	
		b.Run(tc.name, func(b *testing.B) {

			for i := 0; i < tc.input; i++ {
				if f.GetInt(i) == true {
					println(i)
					panic("")
				}
			}
			for i := 0; i < tc.input; i++ {
				if f.GetInt(tc.input+i) == false {
					println(i)
					panic("")
				}
			}
		})
	}
}
func BenchmarkReadAllMap(b *testing.B) {
	cases := []struct {
		name  string
		input int
	}{
		{"Case1", 10},
		{"Case2", 100},
		{"Case3", 1000},
		{"Case4", 10000},
		{"Case5", 100000},
		{"Case6", 1000000},
	}
	for _, tc := range cases {
		var m = make(map[int]bool)
		for i := 0; i < tc.input; i++ {
			m[i] = false
		}
		for i := 0; i < tc.input; i++ {
			m[tc.input+i] = true
		}
		b.Run(tc.name, func(b *testing.B) {

			for i := 0; i < tc.input; i++ {
				if m[i] == true {
					println(i)
					panic("")
				}
			}
			for i := 0; i < tc.input; i++ {
				if m[tc.input+i] == false {
					println(i)
					panic("")
				}
			}
		})
	}
}

func TestSanity(t *testing.T) {

	var filter = Make(map[int]bool{
		5: true,
		55: false,
	})
	if (filter.GetInt(5) != true) {
		panic("5 not true")
	}
	if (filter.GetInt(55) != false) {
		panic("55 not false")
	}

	const test = 10000

	var m = make(map[int]bool)
	for i := 0; i < test; i++ {
		m[i] = false
	}
	for i := 0; i < test; i++ {
		m[test+i] = true
	}
	var f = Make(m)
	for i := 0; i < test; i++ {
		if f.GetInt(i) == true {
			println(i)
			panic("")
		}
	}
	for i := 0; i < test; i++ {
		if f.GetInt(test+i) == false {
			println(i)
			panic("")
		}
	}

}

func TestMapMemoryUsage(t *testing.T) {
	for i := 10; i < 10000000; i *= 10 {
		// Force a GC to ensure we have a clean slate
		runtime.GC()

		// Get memory usage before creating the map
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create a map and populate it
		m := make(map[int]bool)
		for j := 0; j < i; j++ {
			m[j] = true
		}

		// Get memory usage after creating the map
		runtime.ReadMemStats(&m2)

		// Calculate the difference in memory usage
		memoryUsed := m2.Alloc - m1.Alloc

		var quarternaryMemoryUsed = uint64(len(Make(m)))
		// Print the memory used by the map
		fmt.Printf("Memory used by the %d element map: %d bytes\n", i, memoryUsed)
		// Print the memory used by the quarternary
		fmt.Printf("Memory used by the %d element quarternary: %d bytes\n", i, quarternaryMemoryUsed)

		fmt.Printf("Quarternary is: %dx smaller\n", memoryUsed / quarternaryMemoryUsed)

		if memoryUsed < quarternaryMemoryUsed {
			panic(fmt.Sprint(memoryUsed) + "<" + fmt.Sprint(quarternaryMemoryUsed))
		}
	}
}
