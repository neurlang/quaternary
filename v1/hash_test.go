package v1

import (
	"math/rand"
	"testing"
)

func chiSquare(counts []int, total int) float64 {
	exp := float64(total) / float64(len(counts))
	var c float64
	for _, k := range counts {
		d := float64(k) - exp
		c += d * d / exp
	}
	return c
}

// TestHash64Distribution checks both branches of hash64, the 32-bit
// multiply-shift one and the 64-bit modulo one used once 2*cells reaches 2^32
// (a filter body of about 512 MiB), for range, uniformity over the table, the
// parity bit (which selects the cell state) and its independence from the cell
// index. The thresholds are the 99.9th percentiles of the chi-square
// distribution; the generator is seeded, so the outcome is deterministic.
func TestHash64Distribution(t *testing.T) {
	const samples, buckets = 1_000_000, 64
	r := rand.New(rand.NewSource(1))
	for _, m := range []uint64{1<<32 - 8, 1 << 32, 1<<32 + 6, 3 << 31, 1 << 33, 1<<34 + 10} {
		pos := make([]int, buckets)
		par := make([]int, 2)
		joint := make([]int, 4)
		for i := 0; i < samples; i++ {
			hh := hash64(r.Uint32(), r.Uint32(), m)
			if hh >= m {
				t.Fatalf("m=%d: hash %d out of range", m, hh)
			}
			p, c := int(hh&1), int((hh>>1)&1)
			pos[int(hh*buckets/m)]++
			par[p]++
			joint[p*2+c]++
		}
		if x := chiSquare(pos, samples); x > 103.4 { // 63 degrees of freedom
			t.Errorf("m=%d: positions not uniform, chi2=%.1f", m, x)
		}
		if x := chiSquare(par, samples); x > 10.83 { // 1 degree of freedom
			t.Errorf("m=%d: parity not balanced, chi2=%.2f", m, x)
		}
		if x := chiSquare(joint, samples); x > 16.27 { // 3 degrees of freedom
			t.Errorf("m=%d: parity and cell index not independent, chi2=%.2f", m, x)
		}
	}
}
