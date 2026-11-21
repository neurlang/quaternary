package v1

import (
	"fmt"
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
