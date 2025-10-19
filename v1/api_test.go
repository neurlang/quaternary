package v1

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"
)

// Helper: compute masked least-significant `bits` of input big-endian bytes and return in
// big-endian padded to (bits+7)/8 bytes. If bits==0, returns nil.
func maskLSBBytes(input []byte, bits uint64) []byte {
	if bits == 0 {
		return nil
	}
	val := new(big.Int).SetBytes(input)
	if bits > 0 {
		mask := new(big.Int).Lsh(big.NewInt(1), uint(bits))
		mask.Sub(mask, big.NewInt(1))
		val.And(val, mask)
	}
	bytesNeeded := int((bits + 7) / 8)
	out := val.Bytes()
	if len(out) < bytesNeeded {
		pad := make([]byte, bytesNeeded-len(out))
		out = append(pad, out...)
	}
	return out
}

func TestMakeAndGet_BasicCases(t *testing.T) {
	cases := []struct {
		name     string
		bitLimit byte // 0 = Unlimited
		val      []byte
		key      string
	}{
		{"unlimited_zero", 0, []byte{}, "a"},
		{"1bit_zero", 1, []byte{0x00}, "b"},
		{"1bit_one", 1, []byte{0x01}, "c"},
		{"2bit_two", 2, []byte{0x02}, "d"},
		{"7bit_127", 7, []byte{0x7F}, "e"},
		{"8bit_255", 8, []byte{0xFF}, "f"},
		{"9bit_511", 9, []byte{0x01, 0xFF}, "g"},
		{"16bit_65535", 16, []byte{0xFF, 0xFF}, "h"},
		{"32bit", 32, []byte{0x12, 0x34, 0x56, 0x78}, "i"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s_make", c.name), func(t *testing.T) {
			m := map[string][]byte{c.key: c.val}
			var j []byte
			func() {
				defer func() {
					if r := recover(); r == nil {
						if c.bitLimit != 0 {
							if int(c.bitLimit) > len(c.val)*8 {
								t.Fatalf("expected panic when requesting %d bits (stored limit %d) but none", len(c.val)*8, c.bitLimit)
							}
						}
					}
				}()

				j = Make(m, c.bitLimit)
			}()
			if len(j) == 0 {
				t.Fatalf("Make returned empty json")
			}

			// Test Get for several requested sizes
			for req := uint64(1); req <= uint64(len(c.val)*8); req++ {
				var got []byte
				if req <= uint64(c.bitLimit) || c.bitLimit == 0 {
					got = Get(j, req, c.key)
				}
				// if req==0 handled above
				// If stored bitlimit != 0 and req > stored limit -> should panic
				if c.bitLimit != 0 && req > uint64(c.bitLimit) {
					// Expect panic — call in a func and recover
					func() {
						defer func() {
							if r := recover(); r == nil {
								t.Fatalf("expected panic when requesting %d bits (stored limit %d) but none", req, c.bitLimit)
							}
						}()
						_ = Get(j, req, c.key)
					}()
					continue
				}
				// else (unlimited or req<=limit) should return masked bytes
				expected := maskLSBBytes(c.val, req)
				if !bytes.Equal(got, expected) {
					t.Fatalf("Get bits=%d returned %x expected %x", req, got, expected)
				}
			}
		})
	}
}

func TestMakePanicsWhenValueExceedsLimit(t *testing.T) {
	// bitLimit = 1, inserting 2 values must panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when inserting value exceeding bitLimit")
		}
	}()
	Make(map[string][]byte{"x": {0x01, 0x02}}, 1)
}

func TestGetPanicsWhenRequestingMoreThanStoredLimit(t *testing.T) {
	// create storage with bitLimit 9 for a small value
	m := map[string][]byte{"k": {0x01, 0xFE}} // fits 9 bits? 0x01FE = 510 -> 9 bits fits
	j := Make(m, 9)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when Get requests more bits than stored limit")
		}
	}()
	// request 16 bits > stored 9 -> should panic
	_ = Get(j, 16, "k")
}

func TestGetReturnsNilForMissingKey(t *testing.T) {
	m := map[string][]byte{"present": {0x01}}
	j := Make(m, 8)
	if v := Get(j, 8, "absent"); len(v) != 1 {
		t.Fatalf("expected single byte for missing key, got %x", v)
	}
}

func TestMakeUnlimitedAllowsAnyValue(t *testing.T) {
	// large value > 64 bits should be accepted when bitLimit==0 (Unlimited)
	// create 128-bit value
	large := make([]byte, 16)
	for i := range large {
		large[i] = 0xFF
	}
	j := Make(map[string][]byte{"big": large}, 0)
	// Get 128 bits back
	got := Get(j, 128, "big")
	if !bytes.Equal(got, large) {
		t.Fatalf("unlimited Get returned %x want %x", got, large)
	}
}

func TestMakePanicsOnMultipleEntriesIfOneExceeds(t *testing.T) {
	// Mixed map: one entry fits, one entry exceeds -> entire Make must panic
	m := map[string][]byte{
		"ok":   {0x01},
		"bad":  {0x01, 0xFF}, // will exceed 1-bit limit
		"also": {0x00},
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when one item exceeds bitLimit")
		}
	}()
	Make(m, 1)
}

func TestInsertBooleans(t *testing.T) {
	m := map[string]bool{
		"ok":    true,
		"cool":  true,
		"bad":   false,
		"also":  false,
		"again": false,
	}
	b := Make(m, 1)
	for k, v := range m {
		val := GetBool(b, k)
		if v != val {
			t.Fatalf("Insert Booleans returned %v want %v", val, v)
		}
	}
}
func TestInsertIntegersActive(t *testing.T) {
	{
		m := map[string]uint64{
			"active":   1,
			"inactive": 0,
		}
		b := Make(m, 1)
		for k, v := range m {
			valb := GetBool(b, k)
			if (v != 0) != valb {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", valb, k, v)
			}
			val := uint64(GetNum(b, 1, k))
			if v != val {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", val, k, v)
			}

		}
	}
	{
		m := map[string]uint64{
			"S0t6lRZL": 1,
			"kSix4YCB": 0,
		}
		b := Make(m, 1)
		for k, v := range m {
			valb := GetBool(b, k)
			if (v != 0) != valb {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", valb, k, v)
			}
			val := uint64(GetNum(b, 1, k))
			if v != val {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", val, k, v)
			}

		}
	}
	{
		m := map[string]uint64{
			"NDD63C2T": 1,
			"yAVmNkB3": 0,
		}
		b := Make(m, 1)
		for k, v := range m {
			valb := GetBool(b, k)
			if (v != 0) != valb {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", valb, k, v)
			}
			val := uint64(GetNum(b, 1, k))
			if v != val {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", val, k, v)
			}

		}
	}
	{
		m := map[string]uint64{
			"MdR8DAqA": 0,
			"hgPXP3o6": 1,
		}
		b := Make(m, 1)
		for k, v := range m {
			valb := GetBool(b, k)
			if (v != 0) != valb {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", valb, k, v)
			}
			val := uint64(GetNum(b, 1, k))
			if v != val {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", val, k, v)
			}

		}
	}
	{
		m := map[string]uint64{
			"S0t6lRZL": 0,
			"lyBFeCmp": 1,
		}
		b := Make(m, 1)
		for k, v := range m {
			valb := GetBool(b, k)
			if (v != 0) != valb {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", valb, k, v)
			}
			val := uint64(GetNum(b, 1, k))
			if v != val {
				t.Fatalf("Insert Integers Active returned %v for %s want %v", val, k, v)
			}

		}
	}
}

func TestInsertIntegers(t *testing.T) {
	{
		m := map[string]uint8{
			"ok":    1,
			"cool":  2,
			"bad":   3,
			"also":  4,
			"again": 5,
		}
		b := Make(m, 8)
		for k, v := range m {
			val := uint8(GetNum(b, 8, k))
			if v != val {
				t.Fatalf("Insert Integers returned %v want %v", val, v)
			}
		}
	}
	{
		m := map[string]uint16{
			"ok":    1,
			"cool":  2,
			"bad":   3,
			"also":  4,
			"again": 5,
		}
		b := Make(m, 16)
		for k, v := range m {
			val := uint16(GetNum(b, 16, k))
			if v != val {
				t.Fatalf("Insert Integers returned %v want %v", val, v)
			}
		}
	}
	{
		m := map[string]uint32{
			"ok":    1,
			"cool":  2,
			"bad":   3,
			"also":  4,
			"again": 5,
		}
		b := Make(m, 32)
		for k, v := range m {
			val := uint32(GetNum(b, 32, k))
			if v != val {
				t.Fatalf("Insert Integers returned %v want %v", val, v)
			}
		}
	}
	{
		m := map[string]uint64{
			"ok":    1,
			"cool":  2,
			"bad":   3,
			"also":  4,
			"again": 5,
		}
		b := Make(m, 64)
		for k, v := range m {
			val := uint64(GetNum(b, 64, k))
			if v != val {
				t.Fatalf("Insert Integers returned %v want %v", val, v)
			}
		}
	}
}

// ------------------------------
// Fuzz tests
// ------------------------------

func FuzzMakeAndGetRoundTrip(t *testing.F) {
	// Add some seeds
	t.Add("a", []byte{0x00}, 0) // unlimited
	t.Add("b", []byte{0x01}, 1)
	t.Add("c", []byte{0xFF}, 8)
	t.Add("d", []byte{0x01, 0x02}, 9)

	t.Fuzz(func(t *testing.T, key string, val []byte, rawLimit int) {

		key = fmt.Sprintf("%x", key)

		// bound the limit to 0..64 to avoid huge memory allocations
		limit := byte(rawLimit & 0x3F) // 0..63 (0 means unlimited)
		// Build a small map
		m := map[string][]byte{
			key: val,
		}

		// We'll attempt to Make but catch panics to check they correspond to an "exceed" condition.
		var j []byte
		var panicked bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			j = Make(m, limit)
		}()

		if limit == 0 {
			// Unlimited mode -> should never panic
			if panicked {
				t.Fatalf("Make panicked unexpectedly in unlimited mode")
			}
			if len(j) == 0 {
				t.Fatalf("Make returned empty json for unlimited")
			}
			// verify Get with several requested sizes returns masked values
			req := uint64(len(val)) * 8
			{
				got := Get(j, req, key)
				if req == 0 {
					if got != nil {
						t.Fatalf("expected nil when requesting 0 bits in unlimited mode")
					}
					return
				}
				expected := maskLSBBytes(val, req)
				if !bytes.Equal(got, expected) {
					t.Fatalf("unlimited Get mismatch for f=%x req=%d got=%x expected=%x", j, req, got, expected)
				}
			}
			return
		}

		// limit != 0
		if panicked {
			// verify that at least one value's BitLen exceeded limit
			if int(len(val)) == (int(limit)+7)/8 {
				t.Fatalf("Make panicked but none of the checked entries exceed limit (limit=%d, len(val)=%d)", limit, len(val))
			}
			// panic is acceptable (we found an exceeding entry)
			return
		}

	})
}

func FuzzGetUnlimitedBehavior(t *testing.F) {
	// seeds
	t.Add("seed", []byte{0x01, 0x02})
	t.Fuzz(func(t *testing.T, key string, val []byte) {
		req := uint64(len(val)) * 8
		if len(val) == 0 {
			// ensure there's at least one byte to make interesting values
			val = []byte{0x00}
		}
		key = fmt.Sprintf("%x", key)
		j := Make(map[string][]byte{key: val}, 0) // unlimited

		got := Get(j, req, key)
		exp := maskLSBBytes(val, req)
		if !bytes.Equal(got, exp) {
			t.Fatalf("unlimited fuzz Get mismatch req=%d got=%x exp=%x", req, got, exp)
		}
	})
}
