package v1

import "reflect"
import "fmt"
import "encoding/binary"
import "encoding/json"
import "math"

const Unlimited byte = 0

func comparableToString[T comparable](v T) string {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return ""
	}
	switch val := any(v).(type) {
	case string:
		return val
	case bool:
		if val == true {
			return "\x01"
		}
		return "\x00"
	case int:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case int8:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case int16:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case int32:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case int64:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case uint:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case uint8:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case uint16:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case uint32:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case uint64:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case uintptr:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(val))
		return string(b[:])
	case float32:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(math.Float32bits(val)))
		return string(b[:])
	case float64:
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(math.Float64bits(val)))
		return string(b[:])
	}

	bytes, err := json.Marshal(v)
	if err == nil {
		return string(bytes)
	}

	return fmt.Sprintf("%#v", v)
}

// New generates the map based on map m with garbage rate dependent on bloomFuncs
func New[K comparable, V string | []byte | bool | uint64 | uint32 | uint16 | uint8](m map[K]V, bitLimit, bloomFuncs byte) []byte {
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

	// Materialize key-value pairs once to avoid repeated conversions
	pairs := make([]KVPair, 0, len(m))
	for k, v := range m {
		var kv KVPair
		kv.Key = comparableToString(k)
		
		switch val := any(v).(type) {
		case []byte:
			if len(val) == 0 {
				continue
			}
			kv.Value = val
		case string:
			if len(val) == 0 {
				continue
			}
			kv.Value = []byte(val)
		case bool:
			if val {
				kv.Value = []byte{1}
			} else {
				kv.Value = []byte{0}
			}
		case uint64:
			var b [8]byte
			binary.BigEndian.PutUint64(b[:], uint64(val))
			kv.Value = b[8-((bitLimit+7)/8):]
		case uint32:
			var b [4]byte
			binary.BigEndian.PutUint32(b[:], uint32(val))
			kv.Value = b[4-((bitLimit+7)/8):]
		case uint16:
			var b [2]byte
			binary.BigEndian.PutUint16(b[:], uint16(val))
			kv.Value = b[2-((bitLimit+7)/8):]
		case uint8:
			var b = []byte{byte(val)}
			kv.Value = b[:]
		default:
			continue
		}
		pairs = append(pairs, kv)
	}

	// handle the empty pairs case (all values were empty)
	if len(pairs) == 0 {
		return []byte{bloomFuncs, bitLimit}
	}

	// Create iterator over the materialized pairs
	iter := func(yield func(KVPair) bool) {
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

// Bools retrieves a bool and the probabilistic membership based on comparable key
func GetBools[K comparable](f []byte, key K) (bool, bool) {
	k := comparableToString(key)
	// real impl
	data := get(f, []byte(k), 1, f[len(f)-2])
	return (len(data) > 0) && (data[0] == 1), data != nil
}

// Get retrieves an item based on comparable key and value bit size
func Get[K comparable](f []byte, valBitSize uint64, key K) []byte {
	k := comparableToString(key)
	// real impl
	return get(f, []byte(k), valBitSize, f[len(f)-2])
}

// GetBool retrieves a bool based on comparable key
func GetBool[K comparable](f []byte, key K) bool {
	k := comparableToString(key)
	// real impl
	data := get(f, []byte(k), 1, f[len(f)-2])
	return (len(data) > 0) && (data[0] == 1)
}

// GetBoolInt retrieves a bool based on int key (optimized, no allocations)
func GetBoolInt(f []byte, key int) bool {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(key))
	return getBoolBytes(f, b[:])
}

// getBoolBytes retrieves a bool based on byte key (internal, optimized)
func getBoolBytes(f []byte, data []byte) bool {
	var ret [1]byte
	var done [1]byte
	ok := getInto(f, data, ret[:], done[:], 1, f[len(f)-2])
	return ok && (ret[0]&1 == 1)
}

// GetNum retrieves a number based on comparable key and value bit size
func GetNum[K comparable](f []byte, valBitSize uint64, key K) uint64 {
	var buf [8]byte
	k := comparableToString(key)
	// real impl
	b := get(f, []byte(k), valBitSize, f[len(f)-2])
	copy(buf[8-len(b):8], b)
	return binary.BigEndian.Uint64(buf[:])

}
