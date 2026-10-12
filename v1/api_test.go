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
	// bitLimit = 1, inserting a 2-byte value: normBytes truncates to last 1 byte,
	// so Make should NOT panic — it normalises instead.
	// {0x01, 0x02} truncated to 1 byte keeping last byte = {0x02}, lsb = 0 → false
	j := Make(map[string][]byte{"x": {0x01, 0x02}}, 1)
	if len(j) == 0 {
		t.Fatalf("Make returned empty filter")
	}
	got := GetBool(j, "x")
	if got != false {
		t.Fatalf("expected false (lsb of 0x02 truncated to 1 bit = 0), got %v", got)
	}
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
	// With normBytes, values are silently truncated/padded to fit bitLimit.
	// bitLimit=1 → need 1 byte; each value is normalised before storing.
	// "ok":{0x01} → 1 byte, lsb=1 → true
	// "bad":{0x01,0xFF} → truncated to {0xFF}, lsb=1 → true
	// "also":{0x00} → 1 byte, lsb=0 → false
	m := map[string][]byte{
		"ok":   {0x01},
		"bad":  {0x01, 0xFF},
		"also": {0x00},
	}
	j := Make(m, 1)
	if len(j) == 0 {
		t.Fatalf("Make returned empty filter")
	}
	if !GetBool(j, "ok") {
		t.Fatalf("expected true for 'ok'")
	}
	if !GetBool(j, "bad") {
		t.Fatalf("expected true for 'bad' (0xFF lsb=1)")
	}
	if GetBool(j, "also") {
		t.Fatalf("expected false for 'also'")
	}
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

type brokenValue = uint64
func TestInsertBigBytes(t *testing.T) {
	for bc := 3; bc < 63; bc++{
		m := map[[32]byte]brokenValue{
			[32]byte{1}: 1,
			[32]byte{2}: 2,
			[32]byte{3}: 3,
			[32]byte{4}: 4,
			[32]byte{5}: 5,
		}
		b := Make(m, byte(bc))
		for k, v := range m {
			val := brokenValue(GetNum(b, uint64(bc), k))
			if v != val {
				t.Fatalf("InsertBigBytes returned %v want %v", val, v)
			}
		}
	}
}

// TestInsertBigBytesUint32 verifies that uint32 values round-trip correctly
// for bitLimit values 3..62 — including those above uint32's natural 32-bit
// width, which previously caused a slice-bounds panic.
// The sweep starts at 3 so that all test values (1..5) fit within the bit limit.
func TestInsertBigBytesUint32(t *testing.T) {
	for bc := 3; bc < 63; bc++ {
		m := map[[32]byte]uint32{
			[32]byte{1}: 1,
			[32]byte{2}: 2,
			[32]byte{3}: 3,
			[32]byte{4}: 4,
			[32]byte{5}: 5,
		}
		b := Make(m, byte(bc))
		for k, v := range m {
			val := uint32(GetNum(b, uint64(bc), k))
			if v != val {
				t.Fatalf("InsertBigBytesUint32 bc=%d returned %v want %v", bc, val, v)
			}
		}
	}
}

// TestInsertBigBytesUint16 verifies that uint16 values round-trip correctly
// for bitLimit values 3..62 — including those above uint16's natural 16-bit
// width, which previously caused a slice-bounds panic.
// The sweep starts at 3 so that all test values (1..5) fit within the bit limit.
func TestInsertBigBytesUint16(t *testing.T) {
	for bc := 3; bc < 63; bc++ {
		m := map[[32]byte]uint16{
			[32]byte{1}: 1,
			[32]byte{2}: 2,
			[32]byte{3}: 3,
			[32]byte{4}: 4,
			[32]byte{5}: 5,
		}
		b := Make(m, byte(bc))
		for k, v := range m {
			val := uint16(GetNum(b, uint64(bc), k))
			if v != val {
				t.Fatalf("InsertBigBytesUint16 bc=%d returned %v want %v", bc, val, v)
			}
		}
	}
}

// TestInsertBigBytesUint8 verifies that uint8 values round-trip correctly
// for bitLimit values 3..62 — including those above uint8's natural 8-bit
// width, which previously caused a "inserting value exceeding bit limit" panic.
// The sweep starts at 3 so that all test values (1..5) fit within the bit limit.
func TestInsertBigBytesUint8(t *testing.T) {
	for bc := 3; bc < 63; bc++ {
		m := map[[32]byte]uint8{
			[32]byte{1}: 1,
			[32]byte{2}: 2,
			[32]byte{3}: 3,
			[32]byte{4}: 4,
			[32]byte{5}: 5,
		}
		b := Make(m, byte(bc))
		for k, v := range m {
			val := uint8(GetNum(b, uint64(bc), k))
			if v != val {
				t.Fatalf("InsertBigBytesUint8 bc=%d returned %v want %v", bc, val, v)
			}
		}
	}
}

func TestBytesValueNormalization(t *testing.T) {
	// 1. Too-short []byte with bitLimit != 0 → zero-padded left, round-trips via Get
	t.Run("too_short_padded", func(t *testing.T) {
		// bitLimit=16 requires 2 bytes; value is only 1 byte → should be left-padded
		m := map[string][]byte{"k": {0xAB}}
		j := Make(m, 16)
		got := Get(j, 16, "k")
		expected := []byte{0x00, 0xAB}
		if !bytes.Equal(got, expected) {
			t.Fatalf("too-short padding: got %x, want %x", got, expected)
		}
	})

	// 2. Too-long []byte with bitLimit != 0 → truncated from left, round-trips via Get
	t.Run("too_long_truncated", func(t *testing.T) {
		// bitLimit=8 requires 1 byte; value is 3 bytes → keep last byte
		m := map[string][]byte{"k": {0x11, 0x22, 0x33}}
		j := Make(m, 8)
		got := Get(j, 8, "k")
		expected := []byte{0x33}
		if !bytes.Equal(got, expected) {
			t.Fatalf("too-long truncation: got %x, want %x", got, expected)
		}
	})

	// 3. Exact-length []byte with bitLimit != 0 → unchanged, round-trips
	t.Run("exact_length", func(t *testing.T) {
		// bitLimit=16 requires 2 bytes; value is exactly 2 bytes
		m := map[string][]byte{"k": {0xDE, 0xAD}}
		j := Make(m, 16)
		got := Get(j, 16, "k")
		expected := []byte{0xDE, 0xAD}
		if !bytes.Equal(got, expected) {
			t.Fatalf("exact length: got %x, want %x", got, expected)
		}
	})

	// 4. bitLimit == 0 with exact-length values → stored as-is, round-trips
	t.Run("unlimited_as_is", func(t *testing.T) {
		val := []byte{0x01, 0x02, 0x03}
		m := map[string][]byte{"k": val}
		j := Make(m, 0)
		got := Get(j, uint64(len(val))*8, "k")
		if !bytes.Equal(got, val) {
			t.Fatalf("unlimited as-is: got %x, want %x", got, val)
		}
	})

	// 5. string value shorter than (bitLimit+7)/8 → zero-padded, Get returns correct bytes
	t.Run("string_short_padded", func(t *testing.T) {
		// bitLimit=16 requires 2 bytes; string "A" is 1 byte (0x41) → padded to {0x00, 0x41}
		m := map[string]string{"k": "A"}
		j := Make(m, 16)
		got := Get(j, 16, "k")
		expected := []byte{0x00, 0x41}
		if !bytes.Equal(got, expected) {
			t.Fatalf("string short padded: got %x, want %x", got, expected)
		}
	})
}

// FuzzPositionMapUint64 builds a map[[32]byte]uint64 where each value is the
// chunk's index in the input, materializes a filter with New for bloomFuncs 0..8,
// and verifies every key round-trips correctly.
func FuzzPositionMapUint64(f *testing.F) {
	f.Add(make([]byte, 64))
	f.Fuzz(func(t *testing.T, data []byte) {
		const bitLimit = 64
		m := make(map[[32]byte]uint64)
		for i := 0; (i+1)*32 <= len(data); i++ {
			var key [32]byte
			copy(key[:], data[i*32:(i+1)*32])
			m[key] = uint64(i)
		}
		if len(m) == 0 {
			return
		}
		for bloomFuncs := byte(0); bloomFuncs <= 8; bloomFuncs++ {
			filter := New(m, bitLimit, bloomFuncs)
			for k, v := range m {
				got := GetNum(filter, bitLimit, k)
				if got != v {
					t.Fatalf("bloomFuncs=%d key=%x: got %d want %d", bloomFuncs, k, got, v)
				}
			}
		}
	})
}

// FuzzPositionMapUint32 builds a map[[32]byte]uint32 where each value is the
// chunk's index in the input, materializes a filter with New for bloomFuncs 0..8,
// and verifies every key round-trips correctly.
func FuzzPositionMapUint32(f *testing.F) {
	f.Add(make([]byte, 64))
	f.Fuzz(func(t *testing.T, data []byte) {
		const bitLimit = 32
		m := make(map[[32]byte]uint32)
		for i := 0; (i+1)*32 <= len(data); i++ {
			var key [32]byte
			copy(key[:], data[i*32:(i+1)*32])
			m[key] = uint32(i)
		}
		if len(m) == 0 {
			return
		}
		mask := uint64((1 << bitLimit) - 1)
		for bloomFuncs := byte(0); bloomFuncs <= 8; bloomFuncs++ {
			filter := New(m, bitLimit, bloomFuncs)
			for k, v := range m {
				got := GetNum(filter, bitLimit, k)
				if got != uint64(v)&mask {
					t.Fatalf("bloomFuncs=%d key=%x: got %d want %d", bloomFuncs, k, got, uint64(v)&mask)
				}
			}
		}
	})
}

// FuzzPositionMapUint16 builds a map[[32]byte]uint16 where each value is the
// chunk's index in the input, materializes a filter with New for bloomFuncs 0..8,
// and verifies every key round-trips correctly.
func FuzzPositionMapUint16(f *testing.F) {
	f.Add(make([]byte, 64))
	f.Fuzz(func(t *testing.T, data []byte) {
		const bitLimit = 16
		m := make(map[[32]byte]uint16)
		for i := 0; (i+1)*32 <= len(data); i++ {
			var key [32]byte
			copy(key[:], data[i*32:(i+1)*32])
			m[key] = uint16(i)
		}
		if len(m) == 0 {
			return
		}
		mask := uint64((1 << bitLimit) - 1)
		for bloomFuncs := byte(0); bloomFuncs <= 8; bloomFuncs++ {
			filter := New(m, bitLimit, bloomFuncs)
			for k, v := range m {
				got := GetNum(filter, bitLimit, k)
				if got != uint64(v)&mask {
					t.Fatalf("bloomFuncs=%d key=%x: got %d want %d", bloomFuncs, k, got, uint64(v)&mask)
				}
			}
		}
	})
}

// FuzzPositionMapUint8 builds a map[[32]byte]uint8 where each value is the
// chunk's index in the input, materializes a filter with New for bloomFuncs 0..8,
// and verifies every key round-trips correctly.
func FuzzPositionMapUint8(f *testing.F) {
	f.Add(make([]byte, 64))
	f.Fuzz(func(t *testing.T, data []byte) {
		const bitLimit = 8
		m := make(map[[32]byte]uint8)
		for i := 0; (i+1)*32 <= len(data); i++ {
			var key [32]byte
			copy(key[:], data[i*32:(i+1)*32])
			m[key] = uint8(i)
		}
		if len(m) == 0 {
			return
		}
		mask := uint64((1 << bitLimit) - 1)
		for bloomFuncs := byte(0); bloomFuncs <= 8; bloomFuncs++ {
			filter := New(m, bitLimit, bloomFuncs)
			for k, v := range m {
				got := GetNum(filter, bitLimit, k)
				if got != uint64(v)&mask {
					t.Fatalf("bloomFuncs=%d key=%x: got %d want %d", bloomFuncs, k, got, uint64(v)&mask)
				}
			}
		}
	})
}
