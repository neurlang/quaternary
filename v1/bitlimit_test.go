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
