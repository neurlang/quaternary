package v1

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

// This file pins the observable behaviour of filters built by v0.2.5 so that
// later changes to sizing and bitLimit handling cannot silently change the data
// structure:
//
//   - TestCompatPinnedFilters: byte-aligned configurations must keep producing
//     byte-identical filters.
//   - TestCompatLegacyFiltersReadable: filters built by v0.2.5 with a bitLimit
//     that is not a multiple of 8 (sized by the padded byte width) must keep
//     answering correctly, because get derives its layout from len(filter).
//
// The expected values live in compat_golden_test.go. To regenerate them from a
// known-good revision run:
//
//	QUATERNARY_PRINT_GOLDEN=1 go test ./v1 -run TestCompatPrintGolden -v

type compatGoldenEntry struct {
	len int
	sha string
}

func compatKey(i int) [32]byte {
	return sha256.Sum256([]byte(fmt.Sprintf("quaternary-compat-key-%d", i)))
}

func compatBits(i, tag int) uint64 {
	h := sha256.Sum256([]byte(fmt.Sprintf("quaternary-compat-val-%d-%d", tag, i)))
	return binary.BigEndian.Uint64(h[:8])
}

func compatLow(v uint64, bits int) uint64 {
	if bits >= 64 {
		return v
	}
	return v & (uint64(1)<<uint(bits) - 1)
}

type compatCase struct {
	name   string
	build  func() []byte
	verify func(f []byte) error
}

func numCompat[V uint8 | uint16 | uint32 | uint64](name string, n, bits int, bloom byte) compatCase {
	m := make(map[[32]byte]V, n)
	want := make(map[[32]byte]uint64, n)
	for i := 0; i < n; i++ {
		v := V(compatLow(compatBits(i, bits), bits))
		m[compatKey(i)] = v
		want[compatKey(i)] = uint64(v)
	}
	return compatCase{
		name:  name,
		build: func() []byte { return New(m, byte(bits), bloom) },
		verify: func(f []byte) error {
			for k, v := range want {
				if got := GetNum(f, uint64(bits), k); got != v {
					return fmt.Errorf("key %x: got %#x want %#x", k[:4], got, v)
				}
			}
			return nil
		},
	}
}

func boolCompat(name string, n int, bloom byte) compatCase {
	m := make(map[[32]byte]bool, n)
	for i := 0; i < n; i++ {
		m[compatKey(i)] = compatBits(i, 1)&1 == 1
	}
	return compatCase{
		name:  name,
		build: func() []byte { return New(m, 0, bloom) },
		verify: func(f []byte) error {
			for k, v := range m {
				if got, found := GetBools(f, k); got != v || !found {
					return fmt.Errorf("key %x: got %v/%v want %v/true", k[:4], got, found, v)
				}
			}
			return nil
		},
	}
}

func bytesCompat(name string, n, bits int) compatCase {
	need := (bits + 7) / 8
	m := make(map[[32]byte][]byte, n)
	want := make(map[[32]byte][]byte, n)
	for i := 0; i < n; i++ {
		v := make([]byte, need)
		for j := range v {
			v[j] = byte(compatBits(i*64+j, 2) >> 8)
		}
		m[compatKey(i)] = v
		want[compatKey(i)] = v
	}
	return compatCase{
		name:  name,
		build: func() []byte { return New(m, byte(bits), 0) },
		verify: func(f []byte) error {
			for k, v := range want {
				if got := Get(f, uint64(bits), k); string(got) != string(v) {
					return fmt.Errorf("key %x: got %x want %x", k[:4], got, v)
				}
			}
			return nil
		},
	}
}

func dynStringCompat(name string, n int) compatCase {
	m := make(map[[32]byte]string, n)
	for i := 0; i < n; i++ {
		l := 1 + int(compatBits(i, 3)%20)
		b := make([]byte, l)
		for j := range b {
			b[j] = byte(compatBits(i*64+j, 4))
		}
		m[compatKey(i)] = string(b)
	}
	return compatCase{
		name:  name,
		build: func() []byte { return New(m, 0, 0) },
		verify: func(f []byte) error {
			for k, v := range m {
				if got := Get(f, uint64(8*len(v)), k); string(got) != v {
					return fmt.Errorf("key %x: got %x want %x", k[:4], got, []byte(v))
				}
			}
			return nil
		},
	}
}

// compatPinned lists configurations whose filters must stay byte-identical:
// bool, bitLimit==1, and every bitLimit that is a multiple of 8 or dynamic (0).
func compatPinned() []compatCase {
	return []compatCase{
		boolCompat("bool/n50/bf0", 50, 0),
		boolCompat("bool/n50/bf2", 50, 2),
		numCompat[uint64]("u64/b1/n50/bf0", 50, 1, 0),
		numCompat[uint64]("u64/b8/n50/bf0", 50, 8, 0),
		numCompat[uint64]("u64/b24/n30/bf1", 30, 24, 1),
		numCompat[uint64]("u64/b64/n50/bf0", 50, 64, 0),
		numCompat[uint64]("u64/b64/n50/bf3", 50, 64, 3),
		numCompat[uint32]("u32/b16/n50/bf0", 50, 16, 0),
		numCompat[uint16]("u16/b16/n50/bf0", 50, 16, 0),
		numCompat[uint8]("u8/b8/n50/bf0", 50, 8, 0),
		bytesCompat("bytes/b24/n30", 30, 24),
		bytesCompat("bytes/b248/n5", 5, 248),
		dynStringCompat("string/dyn/n30", 30),
	}
}

// compatLegacy lists non-byte-aligned configurations. v0.2.5 sized these by the
// padded byte width; the filters it produced must remain readable.
func compatLegacy() []compatCase {
	return []compatCase{
		numCompat[uint64]("u64/b2/n24/bf0", 24, 2, 0),
		numCompat[uint64]("u64/b5/n24/bf0", 24, 5, 0),
		numCompat[uint64]("u64/b12/n24/bf0", 24, 12, 0),
		numCompat[uint64]("u64/b12/n24/bf1", 24, 12, 1),
		numCompat[uint64]("u64/b33/n24/bf0", 24, 33, 0),
	}
}

func TestCompatPinnedFilters(t *testing.T) {
	for _, c := range compatPinned() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			want, ok := compatGolden[c.name]
			if !ok {
				t.Fatalf("no golden entry for %s", c.name)
			}
			f := c.build()
			sum := sha256.Sum256(f)
			if got := hex.EncodeToString(sum[:8]); len(f) != want.len || got != want.sha {
				t.Fatalf("filter bytes changed: len=%d sha=%s, pinned len=%d sha=%s", len(f), got, want.len, want.sha)
			}
			if err := c.verify(f); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCompatLegacyFiltersReadable(t *testing.T) {
	for _, c := range compatLegacy() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			h, ok := compatLegacyHex[c.name]
			if !ok {
				t.Fatalf("no legacy fixture for %s", c.name)
			}
			f, err := hex.DecodeString(h)
			if err != nil {
				t.Fatal(err)
			}
			if err := c.verify(f); err != nil {
				t.Fatalf("legacy filter no longer readable: %v", err)
			}
		})
	}
}

// TestCompatPrintGolden regenerates compat_golden_test.go. Skipped by default.
func TestCompatPrintGolden(t *testing.T) {
	if os.Getenv("QUATERNARY_PRINT_GOLDEN") == "" {
		t.Skip("set QUATERNARY_PRINT_GOLDEN=1 to print the golden table")
	}
	var b strings.Builder
	b.WriteString("package v1\n\n// Generated by TestCompatPrintGolden from a known-good revision. Do not edit by hand.\n\n")
	b.WriteString("var compatGolden = map[string]compatGoldenEntry{\n")
	for _, c := range compatPinned() {
		f := c.build()
		sum := sha256.Sum256(f)
		fmt.Fprintf(&b, "\t%q: {len: %d, sha: %q},\n", c.name, len(f), hex.EncodeToString(sum[:8]))
	}
	b.WriteString("}\n\nvar compatLegacyHex = map[string]string{\n")
	cs := compatLegacy()
	sort.Slice(cs, func(i, j int) bool { return cs[i].name < cs[j].name })
	for _, c := range cs {
		fmt.Fprintf(&b, "\t%q: %q,\n", c.name, hex.EncodeToString(c.build()))
	}
	b.WriteString("}\n")
	fmt.Print("=====GOLDEN-BEGIN=====\n" + b.String() + "=====GOLDEN-END=====\n")
}
