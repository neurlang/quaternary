// Package v1 provides a generic immutable map implementation with configurable bloom filters
// and variable-length value storage.
//
// This package extends the quaternary filter concept to support:
//   - Any comparable key type (strings, integers, floats, custom types)
//   - Variable-length values ([]byte, string, bool, uint8-uint64)
//   - Optional bloom filters to reduce false positive lookups
//   - Configurable bit limits for compact value storage
//
// # Performance Characteristics
//
// V1 trades performance for flexibility compared to the root package:
//   - Lookup: ~126 ns/op (13x slower than v0, but zero allocations with optimized functions)
//   - Creation: ~850 µs for 1000 entries (34x slower than v0)
//   - Memory: More compact than maps, but larger than v0 due to metadata
//
// # Basic Usage
//
//	// Create a filter with boolean values
//	m := map[string]bool{"key1": true, "key2": false}
//	filter := v1.Make(m, 1) // bitLimit=1 for booleans
//
//	// Lookup values
//	result := v1.GetBool(filter, "key1") // returns true
//
//	// Create with bloom filter (reduces false positives)
//	bloomFilter := v1.New(m, 1, 3) // 3 bloom filter functions
//
// # Optimized Functions
//
// For performance-critical integer lookups, use the optimized zero-allocation functions:
//
//	filter := v1.Make(map[int]bool{42: true, 99: false}, 1)
//	result := v1.GetBoolInt(filter, 42) // Zero allocations
//
// # Key Types
//
// Any comparable type can be used as a key. Keys are converted to bytes internally
// using a deterministic encoding, and a filter answers for that encoding, not for
// Go's notion of key equality. Where the two differ:
//   - Distinct keys can share an encoding: structs that differ only in unexported
//     fields, strings containing invalid UTF-8 inside composite keys or in named
//     string types, and interface keys holding different integer kinds (int(1) and
//     int64(1)). New and Make panic when two keys of the map share an encoding,
//     even if their values are equal. The Integration hook can tell such keys apart.
//   - Pointer keys are encoded by the value they point to, not by their address. Two
//     pointers to equal values share an encoding, and a lookup with any pointer to an
//     equal value hits.
//   - +0.0 and -0.0 are different keys. A Go map treats them as one key and keeps
//     whichever was assigned last, so the filter only recognizes that sign.
//   - Every NaN has the same encoding, so a map holding more than one NaN key panics.
//
// Existing key encodings and the serialized filter format are unchanged.
//
// # Value Types
//
// Supported value types:
//   - bool: Single bit storage
//   - uint8, uint16, uint32, uint64: Compact integer storage
//   - []byte, string: Variable-length data
//
// # Bit Limits
//
// The bitLimit parameter controls value storage size:
//   - 0 (Automatic): uint8/16/32/64 values use the full width of the type;
//     string and []byte values keep their own full length, which may differ
//     between values
//   - 1-255: Limit values to specified number of bits (at most the width of the
//     type for uint8/16/32/64)
//
// Bit limits enable more compact storage but require values to fit within the limit.
// Values that do not fit are silently truncated: numeric values keep their low
// bitLimit bits, and string and []byte values are cut from the left, keeping their
// trailing bytes. Checking that values fit is up to the caller.
//
// The filter is sized by the number of stored bits (bitLimit times the number of
// items, or the sum of the value lengths when bitLimit is 0), not by the value
// width rounded up to whole bytes.
//
// # Reading Values
//
// Get takes valBitSize, the number of bits to read. With a bitLimit above 0, any
// valBitSize up to bitLimit is valid and returns the low valBitSize bits of the
// value (a prefix read).
//
// With bitLimit 0 the filter does not know how long a value is: valBitSize must be
// exactly 8*len(value) of that key. Any other size returns wrong bytes without an
// error, and a prefix read is not possible. A key that was never inserted returns
// arbitrary bytes of the requested size (unless bloom filtering rejects it), so the
// caller has to know which keys exist and how long their values are. A length can
// be recovered in these ways:
//   - Store the length in the least significant byte of a fixed-width value (bitLimit
//     above 0, large enough for the longest value). Read 8 bits to get the length n,
//     then read 8*(1+n) bits to get the value followed by its length byte.
//   - Store the length as a one-byte value under a derived key of the same filter
//     (for example the key plus a suffix). Its size is known, so it can be read
//     exactly, and then the value can be read with its exact size.
//   - Keep the lengths in a second filter built with bitLimit 8.
//
// # Bloom Filters
//
// Optional bloom filters reduce lookup overhead for missing keys:
//
//	filter := v1.New(m, bitLimit, bloomFuncs)
//
// bloomFuncs=0 disables bloom filtering. Higher values (1-8) provide better
// false positive reduction at the cost of larger filter size: for 1-bit values and
// 20,000 keys, bloomFuncs 0, 3 and 8 need about 3, 10 and 23 bits per key. At most
// 56 hash functions are used; a larger bloomFuncs behaves exactly like 56.
//
// # Implementation Details
//
// V1 uses a quaternary (4-state) cell encoding with SHA256-based hashing for
// key distribution. The filter iterates through multiple rounds to resolve
// conflicts, growing the filter size if needed to accommodate all entries.
//
// The storage format includes:
//   - Quaternary cells (2 bits each)
//   - Optional bloom filter bits
//   - Metadata (bloomFuncs, bitLimit)
//
// # Thread Safety
//
// Filters are immutable after creation and safe for concurrent reads.
// Creation is not thread-safe and should be done in a single goroutine.
package v1
