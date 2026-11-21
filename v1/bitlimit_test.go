package v1

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// callWithTimeout runs fn and returns whatever it panicked with (nil if it did
// not panic). A call that does not return within d fails the test, so hangs in
// Make/New are reported as failures instead of stalling the whole test binary.
func callWithTimeout(t *testing.T, d time.Duration, desc string, fn func()) (panicked any) {
	t.Helper()
	ch := make(chan any, 1)
	go func() {
		defer func() { ch <- recover() }()
		fn()
	}()
	select {
	case p := <-ch:
		return p
	case <-time.After(d):
		t.Fatalf("%s: did not return within %v (hang)", desc, d)
		return nil
	}
}

// maskTopByte clears the bits of the most significant byte of v that lie above
// bitLimit, i.e. the bits a filter with that bitLimit does not store.
func maskTopByte(v []byte, bitLimit int) []byte {
	out := append([]byte(nil), v...)
	if r := bitLimit % 8; r != 0 && len(out) > 0 {
		out[0] &= byte(1<<uint(r) - 1)
	}
	return out
}

// TestBytesBitLimitHighRange covers bitLimit values around and above 248. The
// value width (bitLimit+7)/8 used to be computed in byte arithmetic, so for
// bitLimit >= 249 it wrapped to 0 and Make never returned.
func TestBytesBitLimitHighRange(t *testing.T) {
	for bl := 241; bl <= 255; bl++ {
		need := (bl + 7) / 8
		m := map[string][]byte{}
		want := map[string][]byte{}
		for i := 0; i < 3; i++ {
			v := make([]byte, need)
			for j := range v {
				v[j] = byte(0x5A + 37*i + 11*j)
			}
			k := fmt.Sprintf("key-%d", i)
			m[k] = v
			want[k] = maskTopByte(v, bl)
		}
		var f []byte
		desc := fmt.Sprintf("Make([]byte values, bitLimit=%d)", bl)
		if p := callWithTimeout(t, 10*time.Second, desc, func() { f = Make(m, byte(bl)) }); p != nil {
			t.Fatalf("%s panicked: %v", desc, p)
		}
		for k, v := range want {
			if got := Get(f, uint64(bl), k); string(got) != string(v) {
				t.Fatalf("bitLimit=%d key=%s: got %x want %x", bl, k, got, v)
			}
		}
	}
}

// TestUint64BitLimitAboveWidthPanics checks that a bitLimit wider than the
// 64-bit value type is rejected up front with a descriptive panic, for empty
// and non-empty maps alike. It used to surface as a raw slice-bounds panic (or
// not at all for an empty map).
func TestUint64BitLimitAboveWidthPanics(t *testing.T) {
	nonEmpty := map[[32]byte]uint64{{1}: 1, {2}: 2}
	empty := map[[32]byte]uint64{}
	for bl := 65; bl <= 255; bl++ {
		for name, m := range map[string]map[[32]byte]uint64{"non-empty": nonEmpty, "empty": empty} {
			desc := fmt.Sprintf("Make(%s uint64 map, bitLimit=%d)", name, bl)
			p := callWithTimeout(t, 10*time.Second, desc, func() { Make(m, byte(bl)) })
			msg, _ := p.(string)
			if !strings.Contains(msg, "exceeds the 64-bit width") {
				t.Fatalf("%s: want a panic mentioning \"exceeds the 64-bit width\", got %v", desc, p)
			}
		}
	}
	// the boundary itself must keep working
	f := Make(nonEmpty, 64)
	if got := GetNum(f, 64, [32]byte{2}); got != 2 {
		t.Fatalf("bitLimit=64: got %d want 2", got)
	}
}

func checkNumericBitLimitZero[V uint8 | uint16 | uint32 | uint64](t *testing.T, width int, vals []V) {
	t.Helper()
	for _, bloom := range []byte{0, 2} {
		m := map[[32]byte]V{}
		for i, v := range vals {
			m[[32]byte{byte(i + 1)}] = v
		}
		var f []byte
		desc := fmt.Sprintf("New(uint%d values, bitLimit=0, bloom=%d)", width, bloom)
		if p := callWithTimeout(t, 10*time.Second, desc, func() { f = New(m, 0, bloom) }); p != nil {
			t.Fatalf("%s panicked: %v", desc, p)
		}
		if got := int(f[len(f)-1]); got != width {
			t.Fatalf("%s: header bitLimit=%d, want the type width %d", desc, got, width)
		}
		if explicit := New(m, byte(width), bloom); string(explicit) != string(f) {
			t.Fatalf("%s: differs from the filter built with the explicit bitLimit=%d", desc, width)
		}
		for k, v := range m {
			if got := GetNum(f, uint64(width), k); got != uint64(v) {
				t.Fatalf("%s key %x: got %#x want %#x", desc, k[:2], got, uint64(v))
			}
			// reading fewer bits than were stored returns the low bits
			if got := GetNum(f, 8, k); got != uint64(v)&0xFF {
				t.Fatalf("%s key %x: 8-bit prefix got %#x want %#x", desc, k[:2], got, uint64(v)&0xFF)
			}
		}
	}
}

// TestNumericBitLimitZeroUsesFullWidth checks that bitLimit==0 on a fixed-width
// numeric value type means the full width of that type. It used to mean "zero
// bits", which made every value empty and Make/New never return.
func TestNumericBitLimitZeroUsesFullWidth(t *testing.T) {
	checkNumericBitLimitZero[uint8](t, 8, []uint8{0, 1, 0x7F, 0x80, 0xFF})
	checkNumericBitLimitZero[uint16](t, 16, []uint16{0, 1, 0x1234, 0x8000, 0xFFFF})
	checkNumericBitLimitZero[uint32](t, 32, []uint32{0, 1, 0xDEADBEEF, 0x80000000, 0xFFFFFFFF})
	checkNumericBitLimitZero[uint64](t, 64, []uint64{0, 1, 0xDEADBEEFCAFEF00D, 1 << 63, ^uint64(0)})
}

// TestEmptyMapKeepsHeader pins that an empty map still yields the header-only
// filter with the caller's bitLimit, also for numeric types.
func TestEmptyMapKeepsHeader(t *testing.T) {
	f := Make(map[[32]byte]uint64{}, 0)
	if len(f) != 2 || f[0] != 0 || f[1] != 0 {
		t.Fatalf("empty uint64 map, bitLimit=0: got %v, want [0 0]", f)
	}
	f = New(map[[32]byte]uint32{}, 12, 3)
	if len(f) != 2 || f[0] != 3 || f[1] != 12 {
		t.Fatalf("empty uint32 map, bitLimit=12 bloom=3: got %v, want [3 12]", f)
	}
}

// emptyValueIter yields n pairs whose values are all empty, so nothing can be
// stored.
func emptyValueIter(n int) Iterator {
	return func(yield func(kvPair [2][]byte) bool) {
		for i := 0; i < n; i++ {
			k := compatKey(i)
			if !yield([2][]byte{k[:], nil}) {
				return
			}
		}
	}
}

// TestCreateWithNothingToStoreReturnsEmptyFilter checks that create returns the
// header-only filter when the total size is 0 instead of retrying forever
// (growing a zero-sized filter leaves it zero-sized).
func TestCreateWithNothingToStoreReturnsEmptyFilter(t *testing.T) {
	for _, bitLimit := range []byte{0, 1, 2, 8, 64, 255} {
		for _, bloom := range []byte{0, 3} {
			for _, n := range []int{0, 1, 5} {
				desc := fmt.Sprintf("create(bitLimit=%d, bloom=%d, %d empty values)", bitLimit, bloom, n)
				var f []byte
				if p := callWithTimeout(t, 5*time.Second, desc, func() { f = create(emptyValueIter(n), bitLimit, bloom) }); p != nil {
					t.Fatalf("%s panicked: %v", desc, p)
				}
				if len(f) != 2 || f[0] != bloom || f[1] != bitLimit {
					t.Fatalf("%s: got %v, want the header-only filter [%d %d]", desc, f, bloom, bitLimit)
				}
				// and the empty filter answers every lookup as a miss
				if got := Get(f, 8, [32]byte{9}); got != nil {
					t.Fatalf("%s: lookup in the empty filter returned %x", desc, got)
				}
			}
		}
	}
}
