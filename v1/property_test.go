package v1

import (
	"math/rand"
	"sort"
	"testing"
)

// TestGetBoolIntMatchesGetBool checks the optimized, allocation-free GetBoolInt
// (which runs getInto, a second copy of get's probing loop) against the generic
// GetBool, for members and for mostly absent keys, with and without bloom
// functions and for 1-bit and 8-bit stored values.
func TestGetBoolIntMatchesGetBool(t *testing.T) {
	var members, absent int
	for seed := 0; seed < 20; seed++ {
		for _, bf := range []byte{0, 1, 3, 5} {
			for _, bitLimit := range []byte{1, 8} {
				n := 1 + seed*7
				bm := map[int]bool{}
				m8 := map[int]uint8{}
				for i := 0; i < n; i++ {
					k := seed*100003 + i*7919
					v := uint8((i*37 + seed) & 1)
					bm[k], m8[k] = v == 1, v
				}
				var f []byte
				if bitLimit == 1 {
					f = New(bm, 0, bf)
				} else {
					f = New(m8, bitLimit, bf)
				}
				for k, want := range bm {
					members++
					if got := GetBoolInt(f, k); got != want {
						t.Fatalf("seed=%d bloom=%d bitLimit=%d member %d: GetBoolInt=%v want %v", seed, bf, bitLimit, k, got, want)
					}
				}
				for j := 0; j < 400; j++ {
					k := seed*100003 + j*104729 + 1
					absent++
					if a, b := GetBoolInt(f, k), GetBool(f, k); a != b {
						t.Fatalf("seed=%d bloom=%d bitLimit=%d key %d: GetBoolInt=%v GetBool=%v", seed, bf, bitLimit, k, a, b)
					}
				}
			}
		}
	}
	t.Logf("%d member and %d absent-key lookups agree", members, absent)
}

// fixedPoint builds a filter body of the given size from pairs in the given
// order, running store over all pairs until a pass changes nothing, the way
// create does but without its load budget. It reports whether it converged.
func fixedPoint(pairs [][2][]byte, body int, bf byte) ([]byte, bool) {
	fs := make([]byte, body+2)
	fs[len(fs)-2], fs[len(fs)-1] = bf, 1
	if bf > 0 {
		for _, kv := range pairs {
			put(fs, kvPairKeyGet(kv[0]), bf)
		}
	}
	for pass := 0; pass < 150; pass++ {
		var total uint64
		for _, kv := range pairs {
			total += store(fs, kvPairKeyGet(kv[0]), *kvPairValue(&kv), 1)
		}
		if total == 0 {
			return fs, true
		}
	}
	return nil, false
}

// TestFixedPointIsIndependentOfInsertionOrder checks the property the whole
// construction rests on: for a given key set and filter size, the cells a build
// converges to do not depend on the order the keys are inserted in. (The count
// of mutations create measures against its load budget does depend on the
// order, which is why the budget, not the cells, is what can make two builds of
// the same map differ.)
func TestFixedPointIsIndependentOfInsertionOrder(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	var converged, builds int
	for _, n := range []int{3, 10, 50, 200} {
		for _, mult := range []float64{1.0, 1.5} {
			for _, bf := range []byte{0, 2} {
				for rep := 0; rep < 6; rep++ {
					m := map[int]bool{}
					for len(m) < n {
						m[int(r.Int63())] = r.Intn(2) == 1
					}
					var keys []int
					for k := range m {
						keys = append(keys, k)
					}
					sort.Ints(keys)
					var pairs [][2][]byte
					for _, k := range keys {
						var kv [2][]byte
						*kvPairKey(&kv) = comparableToBytes(k)
						*kvPairValue(&kv) = []byte{0}
						if m[k] {
							*kvPairValue(&kv) = []byte{1}
						}
						pairs = append(pairs, kv)
					}
					body := int(float64(byteSize(grow(uint64(n)))) * mult)
					ref, ok := fixedPoint(pairs, body, bf)
					if !ok {
						continue // too small to converge; nothing to compare
					}
					converged++
					for i := 0; i < 10; i++ {
						p := append([][2][]byte(nil), pairs...)
						r.Shuffle(len(p), func(a, b int) { p[a], p[b] = p[b], p[a] })
						got, ok := fixedPoint(p, body, bf)
						builds++
						if !ok || string(got) != string(ref) {
							t.Fatalf("n=%d body=%d bloom=%d: a different insertion order gave different cells (converged=%v)", n, body, bf, ok)
						}
					}
				}
			}
		}
	}
	if converged < 30 {
		t.Fatalf("only %d maps converged; the test is not exercising enough", converged)
	}
	t.Logf("%d converging maps x 10 shuffled orders = %d builds, all identical", converged, builds)
}
