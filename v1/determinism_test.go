package v1

import (
	"bytes"
	"math/rand"
	"sort"
	"testing"
)

// TestBuildIsDeterministic builds the same map repeatedly. New used to hand the
// pairs to create in Go's random map iteration order, and while the cells a
// build converges to do not depend on that order, the load budget create
// compares its mutation count against does, so the same map could come out at
// different sizes (and therefore different bytes) from one build to the next.
func TestBuildIsDeterministic(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	var maps, varying int
	for _, cfg := range [][2]int{{10, 2}, {10, 0}, {30, 2}} {
		for rep := 0; rep < 60; rep++ {
			m := map[int]bool{}
			for len(m) < cfg[0] {
				m[int(r.Int63())] = r.Intn(2) == 1
			}
			first := New(m, 0, byte(cfg[1]))
			maps++
			for i := 0; i < 15; i++ {
				if !bytes.Equal(New(m, 0, byte(cfg[1])), first) {
					varying++
					break
				}
			}
		}
	}
	if varying != 0 {
		t.Fatalf("%d of %d maps produced different filters on repeated builds of the same map", varying, maps)
	}
}

// TestSortPairsByDigestMatchesReference checks the bucket sort used to fix the
// build order against a plain comparison sort, including exact ties on the
// 16-bit bucket prefix and the size where it switches from the small-input path.
func TestSortPairsByDigestMatchesReference(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for _, n := range []int{0, 1, 2, 100, 4095, 4096, 5000, 70000} {
		pairs := make([][2][]byte, n)
		for i := range pairs {
			d := make([]byte, 32)
			r.Read(d)
			if n > 4000 && i%50 == 0 && i > 0 {
				copy(d[:2], pairs[i-1][0][:2])
			}
			pairs[i] = [2][]byte{d, {byte(i), byte(i >> 8)}}
		}
		want := append([][2][]byte(nil), pairs...)
		sort.Slice(want, func(i, j int) bool { return bytes.Compare(want[i][0], want[j][0]) < 0 })
		sortPairsByDigest(pairs)
		for i := range pairs {
			if !bytes.Equal(pairs[i][0], want[i][0]) || !bytes.Equal(pairs[i][1], want[i][1]) {
				t.Fatalf("n=%d: position %d differs from the reference sort (key or its value moved)", n, i)
			}
		}
	}
}
