package v1

import (
	"fmt"
	"testing"
	"time"
)

// TestFilterSizeFollowsBitLimit checks that the filter is sized by
// bitLimit x items, not by the byte-padded value width. Without bloom
// functions the filter needs about 3 bits per stored bit, so a 2-bit value must
// cost about 6 bits per key, not the 24 bits of an 8-bit value.
func TestFilterSizeFollowsBitLimit(t *testing.T) {
	const n = 2000
	for _, b := range []int{1, 2, 3, 5, 7, 9, 12, 17, 33, 63, 64} {
		m := make(map[[32]byte]uint64, n)
		for i := 0; i < n; i++ {
			m[compatKey(i)] = compatLow(compatBits(i, b), b)
		}
		f := Make(m, byte(b))
		if perBit := float64(len(f)*8) / float64(n*b); perBit > 3.1 {
			t.Errorf("bitLimit=%d: %d bytes for %d keys is %.2f bits per stored bit, want about 3", b, len(f), n, perBit)
		}
		for k, v := range m {
			if got := GetNum(f, uint64(b), k); got != v {
				t.Fatalf("bitLimit=%d key %x: got %#x want %#x", b, k[:4], got, v)
			}
		}
	}
}

// TestSmallFiltersAllBitLimits builds tiny filters (down to a single key) for
// every bitLimit, the case where the initial size is tightest. Lookups must not
// hit the "cells <= storedBits" guard.
func TestSmallFiltersAllBitLimits(t *testing.T) {
	for b := 1; b <= 255; b++ {
		need := (b + 7) / 8
		for _, n := range []int{1, 2, 3, 4, 8} {
			m := make(map[[32]byte][]byte, n)
			for i := 0; i < n; i++ {
				v := make([]byte, need)
				for j := range v {
					v[j] = byte(compatBits(i*64+j, b) >> 7)
				}
				m[compatKey(i)] = v
			}
			desc := fmt.Sprintf("Make([]byte values, bitLimit=%d, n=%d)", b, n)
			var f []byte
			if p := callWithTimeout(t, 10*time.Second, desc, func() { f = Make(m, byte(b)) }); p != nil {
				t.Fatalf("%s panicked: %v", desc, p)
			}
			for k, v := range m {
				if got := Get(f, uint64(b), k); string(got) != string(maskTopByte(v, b)) {
					t.Fatalf("%s key %x: got %x want %x", desc, k[:4], got, maskTopByte(v, b))
				}
			}
		}
	}
}
