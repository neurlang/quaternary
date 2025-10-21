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
	var mapping = make(map[string][]byte)
	for k, v := range m {
		switch val := any(v).(type) {
		case []byte:
			if len(val) == 0 {
				continue
			}
			mapping[comparableToString(k)] = val
		case string:
			if len(val) == 0 {
				continue
			}
			mapping[comparableToString(k)] = []byte(val)
		case bool:
			if val {
				mapping[comparableToString(k)] = []byte{1}
			} else {
				mapping[comparableToString(k)] = []byte{0}
			}
			bitLimit = 1
		case uint64:
			var b [8]byte
			binary.BigEndian.PutUint64(b[:], uint64(val))
			mapping[comparableToString(k)] = b[8-((bitLimit+7)/8):]
		case uint32:
			var b [4]byte
			binary.BigEndian.PutUint32(b[:], uint32(val))
			mapping[comparableToString(k)] = b[4-((bitLimit+7)/8):]
		case uint16:
			var b [2]byte
			binary.BigEndian.PutUint16(b[:], uint16(val))
			mapping[comparableToString(k)] = b[2-((bitLimit+7)/8):]
		case uint8:
			var b = []byte{byte(val)}
			mapping[comparableToString(k)] = b[:]
		}
	}

	// handle the empty map case
	if len(mapping) == 0 {
		return []byte{bloomFuncs, bitLimit}
	}

	// real impl
	return create(mapping, bitLimit, bloomFuncs)
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

// GetNum retrieves a number based on comparable key and value bit size
func GetNum[K comparable](f []byte, valBitSize uint64, key K) uint64 {
	var buf [8]byte
	k := comparableToString(key)
	// real impl
	b := get(f, []byte(k), valBitSize, f[len(f)-2])
	copy(buf[8-len(b):8], b)
	return binary.BigEndian.Uint64(buf[:])

}
