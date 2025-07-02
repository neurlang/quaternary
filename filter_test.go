package quaternary

import (
	"fmt"
	"runtime"
	"testing"
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

	Make(map[int]bool{
		0: true,
	})

	Make(map[int]bool{
		0: false,
	})

	var filter = Make(map[int]bool{
		5:  true,
		55: false,
	})
	if filter.GetInt(5) != true {
		panic("5 not true")
	}
	if filter.GetInt(55) != false {
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
func TestSanityStrings(t *testing.T) {
	{
		var filter = Make2Strings(map[[2]string]bool{
			{"a", ""}: true,
			{"b", ""}: false,
			{"", "0"}: true,
			{"", ""}:  false,
		})
		if filter.GetStrings("a", "") != true {
			panic("a not true")
		}
		if filter.GetStrings("b", "") != false {
			panic("b not false")
		}
		if filter.GetStrings("", "0") != true {
			panic("0 not true")
		}
		if filter.GetStrings("", "") != false {
			panic("empty string not false")
		}
	}
}
func TestSanityString(t *testing.T) {

	{
		var filter = MakeString(map[string]bool{
			"a": true,
			"b": false,
			"0": true,
			"":  false,
		})
		if filter.GetString("a") != true {
			panic("a not true")
		}
		if filter.GetString("b") != false {
			panic("b not false")
		}
		if filter.GetString("0") != true {
			panic("0 not true")
		}
		if filter.GetString("") != false {
			panic("empty string not false")
		}
	}
	const test = 10000
	{
		var m = make(map[string]bool)
		for i := 0; i < test; i++ {
			m[fmt.Sprint(i)] = false
		}
		for i := 0; i < test; i++ {
			m[fmt.Sprint(test+i)] = true
		}
		var f = MakeString(m)
		for i := 0; i < test; i++ {
			if f.GetString(fmt.Sprint(i)) == true {
				println(i)
				panic("")
			}
		}
		for i := 0; i < test; i++ {
			if f.GetString(fmt.Sprint(test+i)) == false {
				println(i)
				panic("")
			}
		}
	}
	{
		// long strings
		var mm = make(map[string]bool)
		var a, b string
		for i := 0; i < test; i++ {
			a += "a"
			mm[a] = false
		}
		for i := 0; i < test; i++ {
			b += "b"
			mm[b] = true
		}
		var ff = MakeString(mm)
		a, b = "", ""
		for i := 0; i < test; i++ {
			a += "a"
			if ff.GetString(a) == true {
				println(i)
				panic("this many times a")
			}
		}
		for i := 0; i < test; i++ {
			b += "b"
			if ff.GetString(b) == false {
				println(i)
				panic("this many times b")
			}
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
			m[i+j] = false
		}

		// Get memory usage after creating the map
		runtime.ReadMemStats(&m2)

		// Calculate the difference in memory usage
		memoryUsed := m2.Alloc - m1.Alloc

		var quarternaryMemoryUsed = uint64(len(Make(m)))
		// Print the memory used by the map
		fmt.Printf("[Numeric] Memory used by the %d element map: %d bytes\n", i, memoryUsed)
		// Print the memory used by the quarternary
		fmt.Printf("[Numeric] Memory used by the %d element quarternary: %d bytes\n", i, quarternaryMemoryUsed)

		fmt.Printf("[Numeric] Quarternary is: %dx smaller\n", memoryUsed/quarternaryMemoryUsed)

		if memoryUsed < quarternaryMemoryUsed {
			panic(fmt.Sprint(memoryUsed) + "<" + fmt.Sprint(quarternaryMemoryUsed))
		}
	}
}

func TestMapMemoryUsageString(t *testing.T) {
	for i := 10; i < 10000000; i *= 10 {
		// Force a GC to ensure we have a clean slate
		runtime.GC()

		// Get memory usage before creating the map
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create a map and populate it
		m := make(map[string]bool)
		for j := 0; j < i; j++ {
			m[fmt.Sprint(j)] = true
			m[fmt.Sprint(i+j)] = false
		}

		// Get memory usage after creating the map
		runtime.ReadMemStats(&m2)

		// Calculate the difference in memory usage
		memoryUsed := m2.Alloc - m1.Alloc

		var quarternaryMemoryUsed = uint64(len(MakeString(m)))
		// Print the memory used by the map
		fmt.Printf("[One string] Memory used by the %d element map: %d bytes\n", i, memoryUsed)
		// Print the memory used by the quarternary
		fmt.Printf("[One string] Memory used by the %d element quarternary: %d bytes\n", i, quarternaryMemoryUsed)

		fmt.Printf("[One string] Quarternary is: %dx smaller\n", memoryUsed/quarternaryMemoryUsed)

		if memoryUsed < quarternaryMemoryUsed {
			panic(fmt.Sprint(memoryUsed) + "<" + fmt.Sprint(quarternaryMemoryUsed))
		}
	}
}
func TestMapMemoryUsage2Strings(t *testing.T) {
	for i := 10; i < 10000000; i *= 10 {
		// Force a GC to ensure we have a clean slate
		runtime.GC()

		// Get memory usage before creating the map
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create a map and populate it
		m := make(map[[2]string]bool)
		for j := 0; j < i; j++ {
			m[[2]string{fmt.Sprint(j), fmt.Sprint(i + j)}] = true
			m[[2]string{fmt.Sprint(i + j), fmt.Sprint(j)}] = false
		}

		// Get memory usage after creating the map
		runtime.ReadMemStats(&m2)

		// Calculate the difference in memory usage
		memoryUsed := m2.Alloc - m1.Alloc

		var quarternaryMemoryUsed = uint64(len(Make2Strings(m)))
		// Print the memory used by the map
		fmt.Printf("[Two strings] Memory used by the %d element map: %d bytes\n", i, memoryUsed)
		// Print the memory used by the quarternary
		fmt.Printf("[Two strings] Memory used by the %d element quarternary: %d bytes\n", i, quarternaryMemoryUsed)

		fmt.Printf("[Two strings] Quarternary is: %dx smaller\n", memoryUsed/quarternaryMemoryUsed)

		if memoryUsed < quarternaryMemoryUsed {
			panic(fmt.Sprint(memoryUsed) + "<" + fmt.Sprint(quarternaryMemoryUsed))
		}
	}
}
func TestMapMemoryUsageBytes(t *testing.T) {
	for i := 10; i < 10000000; i *= 10 {
		// Force a GC to ensure we have a clean slate
		runtime.GC()

		// Get memory usage before creating the map
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Create a map and populate it
		m := make(map[[64]byte]bool)
		for j := 0; j < i; j++ {
			m[stringsToByte64(fmt.Sprint(j))] = true
			m[stringsToByte64(fmt.Sprint(i+j))] = false
		}

		// Get memory usage after creating the map
		runtime.ReadMemStats(&m2)

		// Calculate the difference in memory usage
		memoryUsed := m2.Alloc - m1.Alloc

		var quarternaryMemoryUsed = uint64(len(MakeBytes(m)))
		// Print the memory used by the map
		fmt.Printf("[Bytes] Memory used by the %d element map: %d bytes\n", i, memoryUsed)
		// Print the memory used by the quarternary
		fmt.Printf("[Bytes] Memory used by the %d element quarternary: %d bytes\n", i, quarternaryMemoryUsed)

		fmt.Printf("[Bytes] Quarternary is: %dx smaller\n", memoryUsed/quarternaryMemoryUsed)

		if memoryUsed < quarternaryMemoryUsed {
			panic(fmt.Sprint(memoryUsed) + "<" + fmt.Sprint(quarternaryMemoryUsed))
		}
	}
}

func FuzzMapIntBool(f *testing.F) {
	// Seed with example key/value pairs
	f.Add(42, true)
	f.Add(-3, false)
	m := make(map[int]bool)
	f.Fuzz(func(t *testing.T, key int, value bool) {

		f := Make(m)
		for k, v := range m {
			got := f.GetInt(k)
			if got != v {
				t.Fatalf("map[int]bool: expected %v for key %q, got %v", v, k, got)
			}
		}
		m[key] = value // Fuzzer controls the key and value
	})
}

func FuzzMapStringBool(f *testing.F) {
	// Seed with example key/value pairs
	f.Add("hello", true)
	f.Add("", false) // Empty string key
	m := make(map[string]bool)
	f.Fuzz(func(t *testing.T, key string, value bool) {

		f := MakeString(m)
		for k, v := range m {
			got := f.GetString(k)
			if got != v {
				t.Fatalf("map[string]bool: expected %v for key %q, got %v", v, k, got)
			}
		}
		m[key] = value // Fuzzer controls the key and value
	})
}
