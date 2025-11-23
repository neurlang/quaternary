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
//	filter := v1.New(m, 1, 3) // 3 bloom filter functions
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
// Any comparable type can be used as a key. Keys are converted to strings internally
// using a deterministic encoding that preserves uniqueness.
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
//   - 0 (Unlimited): Store full value length
//   - 1-255: Limit values to specified number of bits
//
// Bit limits enable more compact storage but require values to fit within the limit.
//
// # Bloom Filters
//
// Optional bloom filters reduce lookup overhead for missing keys:
//
//	filter := v1.New(m, bitLimit, bloomFuncs)
//
// bloomFuncs=0 disables bloom filtering. Higher values (1-8) provide better
// false positive reduction at the cost of larger filter size.
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
