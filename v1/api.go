package v1

import "reflect"
import "fmt"
import "encoding/binary"
import "encoding/json"
import "math"
import sha256 "github.com/minio/sha256-simd"

const Unlimited byte = 0

// normBytes adjusts b to exactly need bytes: truncate from left if too long,
// zero-pad on the left if too short.
// bytesForBits returns the number of bytes needed to hold bitLimit bits.
//
// The arithmetic must be done in int: bitLimit is a byte, so evaluating
// bitLimit+7 in byte arithmetic wraps around for bitLimit >= 249 and yields a
// width of 0 bytes.
func bytesForBits(bitLimit byte) int {
	return (int(bitLimit) + 7) / 8
}

// fixedBitWidth reports the natural width in bits of V when V is a fixed-width
// numeric type (uint8, uint16, uint32, uint64). ok is false for string, []byte
// and bool.
func fixedBitWidth[V string | []byte | bool | uint64 | uint32 | uint16 | uint8]() (width int, ok bool) {
	var zero V
	switch any(zero).(type) {
	case uint64:
		return 64, true
	case uint32:
		return 32, true
	case uint16:
		return 16, true
	case uint8:
		return 8, true
	}
	return 0, false
}

func normBytes(b []byte, need int) []byte {
	if len(b) > need {
		return b[len(b)-need:]
	}
	if len(b) < need {
		padded := make([]byte, need)
		copy(padded[need-len(b):], b)
		return padded
	}
	return b
}

// Hashstructure is the integration callback type for custom map key hashing library e.g.:
// https://github.com/mitchellh/hashstructure
type Hashstructure func(interface{}) []byte

// Integration is the integration to the custom map key hashing library e.g.:
// https://github.com/mitchellh/hashstructure
var Integration Hashstructure

func comparableToBytes[T comparable](v T) []byte {
	var data = comparableToBytesInside(v)
	if Integration != nil {
		data = append(data, Integration(v)...)
	}
	var ret = sha256.Sum256(data)
	return ret[:]
}

func comparableToBytesInsideHashstructure[T comparable](v T) []byte {
	var data = comparableToBytesInside(v)
	if Integration != nil {
		data = append(data, Integration(v)...)
	}
	return data
}

func comparableToBytesInside[T comparable](v T) []byte {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return nil
	}
	switch val := any(v).(type) {
	case string:
		return []byte(val)
	case bool:
		if val == true {
			return []byte{1}
		}
		return []byte{0}
	case int:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case int8:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case int16:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case int32:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case int64:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case uint:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case uint8:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case uint16:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case uint32:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case uint64:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case uintptr:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return b[:]
	case float32:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(math.Float32bits(val)))
		return b[:]
	case float64:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(math.Float64bits(val)))
		return b[:]
	}

	bytes, err := json.Marshal(v)
	if err == nil {
		return bytes
	}

	return []byte(fmt.Sprintf("%#v", v))
}

// New generates the map based on map m with garbage rate dependent on bloomFuncs.
// It panics if distinct keys encode to identical bytes with conflicting stored values.
// It also panics if bitLimit exceeds the width of a fixed-width value type (8, 16, 32
// or 64 bits for uint8/16/32/64). For those types a bitLimit of 0 means the full width.
func New[K comparable, V string | []byte | bool | uint64 | uint32 | uint16 | uint8](m map[K]V, bitLimit, bloomFuncs byte) []byte {
	// Reject a bitLimit the value type cannot represent. This is argument
	// validation, so it does not depend on whether m happens to be empty.
	if width, ok := fixedBitWidth[V](); ok && int(bitLimit) > width {
		panic(fmt.Sprintf("quaternary: bitLimit %d exceeds the %d-bit width of the value type", bitLimit, width))
	}

	// Check if map is empty
	if len(m) == 0 {
		return []byte{bloomFuncs, bitLimit}
	}

	// Adjust bitLimit for bool type
	for _, v := range m {
		if _, ok := any(v).(bool); ok {
			bitLimit = 1
			break
		}
	}

	// A fixed-width numeric type has a natural width, so bitLimit 0 ("automatic")
	// means the full width of the type. Only string and []byte values have no
	// fixed width and stay dynamic (bitLimit 0) with their own length per value.
	if width, ok := fixedBitWidth[V](); ok && bitLimit == 0 {
		bitLimit = byte(width)
	}

	// Materialize key-value pairs once to avoid repeated conversions
	pairs := make([][2][]byte, 0, len(m))
	encoded := make(map[[32]byte]struct{}, len(m))
	for k, v := range m {
		var kv [2][]byte
		*kvPairKey(&kv) = comparableToBytes(k)

		switch val := any(v).(type) {
		case []byte:
			if len(val) == 0 {
				continue
			}
			if bitLimit != 0 {
				need := bytesForBits(bitLimit)
				*kvPairValue(&kv) = normBytes(val, need)
			} else {
				*kvPairValue(&kv) = val
			}
		case string:
			if len(val) == 0 {
				continue
			}
			b := []byte(val)
			if bitLimit != 0 {
				need := bytesForBits(bitLimit)
				*kvPairValue(&kv) = normBytes(b, need)
			} else {
				*kvPairValue(&kv) = b
			}
		case bool:
			if val {
				*kvPairValue(&kv) = []byte{1}
			} else {
				*kvPairValue(&kv) = []byte{0}
			}
		case uint64:
			var b [8]byte
			binary.BigEndian.PutUint64(b[:], uint64(val))
			*kvPairValue(&kv) = b[8-bytesForBits(bitLimit):]
		case uint32:
			var b [4]byte
			binary.BigEndian.PutUint32(b[:], uint32(val))
			need := bytesForBits(bitLimit)
			start := 4 - need
			if start < 0 {
				start = 0
			}
			*kvPairValue(&kv) = normBytes(b[start:], need)
		case uint16:
			var b [2]byte
			binary.BigEndian.PutUint16(b[:], uint16(val))
			need := bytesForBits(bitLimit)
			start := 2 - need
			if start < 0 {
				start = 0
			}
			*kvPairValue(&kv) = normBytes(b[start:], need)
		case uint8:
			b := []byte{byte(val)}
			need := bytesForBits(bitLimit)
			*kvPairValue(&kv) = normBytes(b, need)
		default:
			continue
		}
		key := kvPairKeyGet(kv[0])
		if _, ok := encoded[key]; ok {
			panic("quaternary: conflicting values for identical key encoding: use Hashstructure Integration in quaternary package level to fix")
		}
		encoded[key] = struct{}{}
		pairs = append(pairs, kv)
	}

	// handle the empty pairs case (all values were empty)
	if len(pairs) == 0 {
		return []byte{bloomFuncs, bitLimit}
	}

	// Create iterator over the materialized pairs
	iter := func(yield func(kvPair [2][]byte) bool) {
		for i := range pairs {
			if !yield(pairs[i]) {
				return
			}
		}
	}

	// real impl
	return create(iter, bitLimit, bloomFuncs)
}

// Make generates the filter based on map m
func Make[K comparable, V string | []byte | bool | uint64 | uint32 | uint16 | uint8](m map[K]V, bitLimit byte) []byte {
	return New(m, bitLimit, 0)
}

// Hashstructure

// GetBools retrieves a bool and the probabilistic membership based on comparable key
func GetBools[K comparable](f []byte, key K) (bool, bool) {
	if len(f) < 2 {
		return false, false
	}
	k := comparableToBytesInsideHashstructure(key)
	// real impl
	data := get(f, k, 1, f[len(f)-2])
	return (len(data) > 0) && (data[0] == 1), data != nil
}

// Get retrieves an item based on comparable key and value bit size
func Get[K comparable](f []byte, valBitSize uint64, key K) []byte {
	if len(f) < 2 {
		return nil
	}
	k := comparableToBytesInsideHashstructure(key)
	// real impl
	return get(f, k, valBitSize, f[len(f)-2])
}

// GetBool retrieves a bool based on comparable key
func GetBool[K comparable](f []byte, key K) bool {
	if len(f) < 2 {
		return false
	}
	k := comparableToBytesInsideHashstructure(key)
	// real impl
	data := get(f, k, 1, f[len(f)-2])
	return (len(data) > 0) && (data[0] == 1)
}

// GetBoolInt retrieves a bool based on int key (optimized, no allocations)
func GetBoolInt(f []byte, key int) bool {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(key))
	var data = b[:]
	if Integration != nil {
		data = append(data, Integration(key)...)
	}
	return hashstructureGetBoolBytes(f, data)
}

// hashstructureGetBoolBytes retrieves a bool based on byte key (internal, optimized)
func hashstructureGetBoolBytes(f []byte, data []byte) bool {
	if len(f) < 2 {
		return false
	}
	var ret [1]byte
	var done [1]byte
	ok := getInto(f, data, ret[:], done[:], 1, f[len(f)-2])
	return ok && (ret[0]&1 == 1)
}

// GetNum retrieves a number based on comparable key and value bit size
func GetNum[K comparable](f []byte, valBitSize uint64, key K) uint64 {
	if len(f) < 2 {
		return 0
	}
	var buf [8]byte
	k := comparableToBytesInsideHashstructure(key)
	// real impl
	b := get(f, k, valBitSize, f[len(f)-2])
	copy(buf[8-len(b):8], b)
	return binary.BigEndian.Uint64(buf[:])

}
