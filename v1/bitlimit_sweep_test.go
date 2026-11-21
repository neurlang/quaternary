package v1

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
	"time"
)

// This file sweeps every bitLimit 0..255 over every supported value type. Each
// combination must either round-trip all values or be rejected with a
// descriptive panic; none may hang. Builds run under a timeout.

// sweepRejects reports whether New must reject bitLimit b for a value type of
// the given fixed bit width (0 for types without a fixed width).
func sweepRejects(width, b int) bool {
	return width != 0 && b > width
}

// sweepNumVals returns five values that fit in the effective number of bits.
func sweepNumVals(width, b int) []uint64 {
	bits := b
	if bits == 0 || bits > width {
		bits = width
	}
	return []uint64{
		0,
		compatLow(1, bits),
		compatLow(^uint64(0), bits),
		compatLow(0xA5A5A5A5A5A5A5A5, bits),
		compatLow(uint64(1)<<uint(bits-1), bits),
	}
}

// sweepExpectedBytes is the value zero-extended big-endian to the stored width
// of effB bits, with the bits above effB cleared.
func sweepExpectedBytes(v uint64, effB int) []byte {
	need := (effB + 7) / 8
	var full [8]byte
	binary.BigEndian.PutUint64(full[:], v)
	out := make([]byte, need)
	n := need
	if n > 8 {
		n = 8
	}
	copy(out[need-n:], full[8-n:])
	return maskTopByte(out, effB)
}

func sweepNumeric[V uint8 | uint16 | uint32 | uint64](t *testing.T, width int) {
	t.Helper()
	for b := 0; b <= 255; b++ {
		for _, n := range []int{1, 5} {
			vals := sweepNumVals(width, b)[:n]
			m := make(map[[32]byte]V, n)
			for i, v := range vals {
				m[compatKey(i)] = V(v)
			}
			desc := fmt.Sprintf("New(uint%d values, bitLimit=%d, n=%d)", width, b, n)
			var f []byte
			p := callWithTimeout(t, 10*time.Second, desc, func() { f = New(m, byte(b), 0) })
			if sweepRejects(width, b) {
				if msg, _ := p.(string); !strings.Contains(msg, "exceeds the") {
					t.Fatalf("%s: want a panic containing \"exceeds the\", got %v", desc, p)
				}
				continue
			}
			if p != nil {
				t.Fatalf("%s panicked: %v", desc, p)
			}
			eff := b
			if eff == 0 {
				eff = width
			}
			for i, v := range vals {
				k := compatKey(i)
				if got, want := Get(f, uint64(eff), k), sweepExpectedBytes(v, eff); string(got) != string(want) {
					t.Fatalf("%s key %d: Get got %x want %x", desc, i, got, want)
				}
				if eff <= 64 {
					if got := GetNum(f, uint64(eff), k); got != v {
						t.Fatalf("%s key %d: GetNum got %#x want %#x", desc, i, got, v)
					}
				}
			}
		}
	}
}

func sweepBytes[V string | []byte](t *testing.T, name string, conv func([]byte) V) {
	t.Helper()
	for b := 0; b <= 255; b++ {
		for _, n := range []int{1, 5} {
			lens := []int{1, 2, 3, 7, 40}[:n]
			m := make(map[[32]byte]V, n)
			raw := make([][]byte, n)
			for i := 0; i < n; i++ {
				l := lens[i]
				if b != 0 {
					l = (b + 7) / 8
				}
				v := make([]byte, l)
				for j := range v {
					v[j] = byte(0x5A + 37*i + 11*j)
				}
				raw[i] = v
				m[compatKey(i)] = conv(v)
			}
			desc := fmt.Sprintf("New(%s values, bitLimit=%d, n=%d)", name, b, n)
			var f []byte
			if p := callWithTimeout(t, 10*time.Second, desc, func() { f = New(m, byte(b), 0) }); p != nil {
				t.Fatalf("%s panicked: %v", desc, p)
			}
			for i, v := range raw {
				k := compatKey(i)
				want, bits := v, uint64(8*len(v))
				if b != 0 {
					want, bits = maskTopByte(v, b), uint64(b)
				}
				if got := Get(f, bits, k); string(got) != string(want) {
					t.Fatalf("%s key %d: got %x want %x", desc, i, got, want)
				}
			}
		}
	}
}

func TestBitLimitSweepAllValueTypes(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		for b := 0; b <= 255; b++ {
			for _, n := range []int{1, 5} {
				want := []bool{true, false, true, true, false}[:n]
				m := make(map[[32]byte]bool, n)
				for i, v := range want {
					m[compatKey(i)] = v
				}
				desc := fmt.Sprintf("New(bool values, bitLimit=%d, n=%d)", b, n)
				var f []byte
				if p := callWithTimeout(t, 10*time.Second, desc, func() { f = New(m, byte(b), 0) }); p != nil {
					t.Fatalf("%s panicked: %v", desc, p)
				}
				for i, v := range want {
					if got, found := GetBools(f, compatKey(i)); got != v || !found {
						t.Fatalf("%s key %d: got %v/%v want %v/true", desc, i, got, found, v)
					}
				}
			}
		}
	})
	t.Run("string", func(t *testing.T) { sweepBytes(t, "string", func(b []byte) string { return string(b) }) })
	t.Run("[]byte", func(t *testing.T) { sweepBytes(t, "[]byte", func(b []byte) []byte { return b }) })
	t.Run("uint8", func(t *testing.T) { sweepNumeric[uint8](t, 8) })
	t.Run("uint16", func(t *testing.T) { sweepNumeric[uint16](t, 16) })
	t.Run("uint32", func(t *testing.T) { sweepNumeric[uint32](t, 32) })
	t.Run("uint64", func(t *testing.T) { sweepNumeric[uint64](t, 64) })
}
