package v1

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"testing"
)

// hugeRoundTrip drives the real put, store and get code on a filter body of the
// given size holding n keys. The buffer is mostly untouched, so the keys do not
// conflict and a build converges after a pass or two; the point is the index
// arithmetic and the hash branch at that size. It reports whether the 64-bit
// modulo branch of hash64 was used.
func hugeRoundTrip(t *testing.T, body int, bitLimit, bf byte, n int) (wide bool) {
	t.Helper()
	fs := make([]byte, body+2)
	fs[len(fs)-2], fs[len(fs)-1] = bf, bitLimit
	cells := cellSize(uint64(len(fs)) - 2)
	m := (cells - (uint64(bitLimit) - 1)) << 1
	wide = m >= 1<<32

	need := (int(bitLimit) + 7) / 8
	rng := rand.New(rand.NewSource(int64(body)*31 + int64(bitLimit)))
	keys := make([][]byte, n)
	vals := make([][]byte, n)
	for i := range keys {
		keys[i] = []byte(fmt.Sprintf("huge-%d-%d", body, i))
		v := make([]byte, need)
		rng.Read(v)
		vals[i] = maskTopByte(v, int(bitLimit))
	}
	if bf > 0 {
		for _, k := range keys {
			put(fs, sha256.Sum256(k), bf)
		}
	}
	for pass := 0; ; pass++ {
		var total uint64
		for i, k := range keys {
			total += store(fs, sha256.Sum256(k), vals[i], bitLimit)
		}
		if total == 0 {
			break
		}
		if pass > 100 {
			t.Fatalf("body=%d bitLimit=%d: no fixed point", body, bitLimit)
		}
	}
	for i, k := range keys {
		if got := get(fs, k, uint64(bitLimit), bf); string(got) != string(vals[i]) {
			t.Errorf("body=%d bitLimit=%d bloom=%d (m=%d, wide=%v): key %d reads %x, want %x", body, bitLimit, bf, m, wide, i, got, vals[i])
			return wide
		}
	}
	return wide
}

// TestHugeFilterBodyRoundTrip covers filter bodies around 512 MiB, where
// 2*cells reaches 2^32 and hash64 switches from its 32-bit multiply-shift branch
// to the 64-bit modulo branch; nothing else in the suite gets near that size.
// It needs about 1.5 GB of memory, so it only runs with
// QUATERNARY_BIG_TESTS=1 on a 64-bit platform.
func TestHugeFilterBodyRoundTrip(t *testing.T) {
	if os.Getenv("QUATERNARY_BIG_TESTS") == "" {
		t.Skip("set QUATERNARY_BIG_TESTS=1 to run; needs about 1.5 GB of memory")
	}
	if strconv.IntSize < 64 {
		t.Skip("needs a 64-bit platform")
	}
	var narrow, wide int
	// 1<<29-1 and 1<<29 straddle the switch for bitLimit 1; 1<<29+100 is past it
	// for bitLimit 64, whose smaller usable cell count moves the switch up a little.
	for _, body := range []int{1<<29 - 1, 1 << 29, 1<<29 + 100, 3 << 28} {
		for _, bitLimit := range []byte{1, 64} {
			for _, bf := range []byte{0, 3} {
				if hugeRoundTrip(t, body, bitLimit, bf, 2000) {
					wide++
				} else {
					narrow++
				}
			}
		}
		runtime.GC()
		debug.FreeOSMemory()
	}
	if narrow == 0 || wide == 0 {
		t.Fatalf("the test no longer exercises both hash branches (32-bit: %d cases, 64-bit: %d cases)", narrow, wide)
	}
	t.Logf("%d cases below and %d cases above the 2^32 switch, all round-trip", narrow, wide)
}
