// Package quaternary implements a smaller but immutable map which can't be iterated
package quaternary

import "crypto/sha512"

func byteSize(n int) int {
	return (3 + n) / 4
}

func cellSize(n int) int {
	return n * 4
}

func grow(n int) int {
	return (3*n + 1) / 2
}

const ROUNDS = 64

/*
type Bloom []byte

func (f Bloom) Put(num, salt uint32) {
	idx := hash64(num, salt, uint64(len(f)*8))
	(f)[idx>>3] |= 1 << (idx & 7)
}

func (f Bloom) Get(num, salt uint32) bool {
	idx := hash64(num, salt, uint64(len(f)*8))
	return ((f)[idx>>3]>>(idx&7))&1 == 1
}
*/
// Filter is an immutable map without iterating capability.
// It is used to store keys of various types and retrieve the corresponding booleans.
type Filter []byte

// Filters are multiple Filters of the same size
type Filters [][]byte

func (fs Filters) store(data []byte, answer uint64) (inserted []int) {
	if len(fs) == 0 {
		return nil
	}

	// Validate all filters have same size
	baseSize := len(fs[0])
	for _, f := range fs {
		if len(f) != baseSize {
			panic("all filters must have same size")
		}
	}

	cells := cellSize(baseSize)
	inserted = make([]int, len(fs))

	// Track active filters and their insertion counts
	active := make([]bool, len(fs))
	counts := make([]int, len(fs))
	for i := range fs {
		if len(fs[i]) > 0 {
			active[i] = true
		} else {
			inserted[i] = 1 // Mark empty filters
		}
	}

	// Process rounds
	for round := uint32(0); round < ROUNDS; round++ {
		anyActive := false
		h := hash64(dataHash(round, data), uint32(cells), uint64(cells)<<1)

		for i, f := range fs {
			if !active[i] {
				continue
			}
			anyActive = true

			pos := h >> 3
			shift := h & 6
			state := (f[pos] >> shift) & 3
			bit := byte((answer >> uint(i)) & 1)

			switch state {
			case 0:
				if bit == byte(h&1) {
					active[i] = false
				} else {
					f[pos] |= ((bit & 1) + 1) << shift
					counts[i] = 1
					active[i] = false
				}
			case 1:
				if bit == 0 {
					active[i] = false
				} else {
					f[pos] |= 3 << shift
					counts[i]++
				}
			case 2:
				if bit == 1 {
					active[i] = false
				} else {
					f[pos] |= 3 << shift
					counts[i]++
				}
			}
		}

		if !anyActive {
			break
		}
	}

	// Finalize insertion counts
	for i := range fs {
		if inserted[i] == 0 {
			if active[i] {
				inserted[i] = counts[i] + 1
			} else {
				inserted[i] = counts[i]
			}
		}
	}
	return inserted
}

func (fs Filters) insert(num uint64, answer uint64) (inserted []int) {
	if len(fs) == 0 {
		return nil
	}

	// Validate all filters have same size
	baseSize := len(fs[0])
	for _, f := range fs {
		if len(f) != baseSize {
			panic("all filters must have same size")
		}
	}

	cells := cellSize(baseSize)
	inserted = make([]int, len(fs))

	x0 := uint32(num)
	high := uint32(num >> 32)

	// Track active filters and insertion counts
	active := make([]bool, len(fs))
	counts := make([]int, len(fs))
	for i := range fs {
		if len(fs[i]) > 0 {
			active[i] = true
		} else {
			inserted[i] = 1 // Mark empty filters
		}
	}

	// Process rounds
	for r := uint32(0); r < ROUNDS; r++ {
		anyActive := false
		shift := r & 31
		xr := (x0 >> shift) | (x0 << (32 - shift))
		h := hash64(xr, high^r, uint64(cells))
		lb := byte(xr & 1)

		for i, f := range fs {
			if !active[i] {
				continue
			}
			anyActive = true

			pos := h >> 2
			shift := (h & 3) * 2
			state := (f[pos] >> shift) & 3
			bit := byte((answer >> uint(i)) & 1)

			switch state {
			case 0:
				if bit == lb {
					active[i] = false
				} else {
					f[pos] |= ((bit & 1) + 1) << shift
					counts[i] = 1
					active[i] = false
				}
			case 1:
				if bit == 0 {
					active[i] = false
				} else {
					f[pos] |= 3 << shift
					counts[i]++
				}
			case 2:
				if bit == 1 {
					active[i] = false
				} else {
					f[pos] |= 3 << shift
					counts[i]++
				}
			}
		}

		if !anyActive {
			break
		}
	}

	// Finalize insertion counts
	for i := range fs {
		if inserted[i] == 0 {
			if active[i] {
				inserted[i] = counts[i] + 1
			} else {
				inserted[i] = counts[i]
			}
		}
	}
	return inserted
}

func (f Filter) store(data []byte, answer byte) (inserted int) {
	if len(f) == 0 {
		return 1
	}
	cells := cellSize(len(f))
	//println("insert", string(data), "size", len(f))
	for i := uint32(0); i < ROUNDS; i++ {
		h := hash64(dataHash(i, data), uint32(cells), uint64(cells)<<1)
		switch ((f)[h>>3] >> (h & 6)) & 3 {
		case 0:
			if answer == byte(h&1) {
				return inserted
			}
			(f)[h>>3] |= ((answer & 1) + 1) << (h & 6)
			inserted++
			return inserted
		case 1:
			if answer == 0 {
				return inserted
			}
		case 2:
			if answer == 1 {
				return inserted
			}
		default:
			continue
		}
		(f)[h>>3] |= 3 << (h & 6)
		inserted++
	}
	return inserted + 1
}

func (f Filter) insert(num uint64, answer byte) (inserted int) {
	if len(f) == 0 {
		return 1
	}
	cells := cellSize(len(f))
	x := uint32(num)
	high := uint32(num >> 32)
	//println("insert", x, high, "size", len(f))
	for i := uint32(0); i < ROUNDS; i++ {
		h := hash64(x, high^i, uint64(cells))
		switch ((f)[h>>2] >> ((h & 3) * 2)) & 3 {
		case 0:
			if answer == byte(x&1) {
				return inserted
			}
			(f)[h>>2] |= ((answer & 1) + 1) << ((h & 3) * 2)
			inserted++
			return inserted
		case 1:
			if answer == 0 {
				return inserted
			}
		case 2:
			if answer == 1 {
				return inserted
			}
		case 3:
			x = (x >> 1) | (x << 31)
			continue
		}
		(f)[h>>2] |= 3 << ((h & 3) * 2)
		x = (x >> 1) | (x << 31)
		inserted++

	}
	return inserted + 1
}

// GetInt checks if an int value exists in the Filter.
func (f Filter) GetInt(num int) bool {
	return f.GetUint64(uint64(num))
}

// GetUint checks if an uint value exists in the Filter.
func (f Filter) GetUint(num uint) bool {
	return f.GetUint64(uint64(num))
}

// GetInt8 checks if an int8 value exists in the Filter.
func (f Filter) GetInt8(num int8) bool {
	return f.GetUint64(uint64(num))
}

// GetUint8 checks if an uint8 value exists in the Filter.
func (f Filter) GetUint8(num uint8) bool {
	return f.GetUint64(uint64(num))
}

// GetInt16 checks if an int16 value exists in the Filter.
func (f Filter) GetInt16(num int16) bool {
	return f.GetUint64(uint64(num))
}

// GetUint16 checks if an uint16 value exists in the Filter.
func (f Filter) GetUint16(num uint16) bool {
	return f.GetUint64(uint64(num))
}

// GetInt32 checks if an int32 value exists in the Filter.
func (f Filter) GetInt32(num int32) bool {
	return f.GetUint64(uint64(num))
}

// GetUint32 checks if an uint32 value exists in the Filter.
func (f Filter) GetUint32(num uint32) bool {
	return f.GetUint64(uint64(num))
}

// GetInt64 checks if an int64 value exists in the Filter.
func (f Filter) GetInt64(num int64) bool {
	return f.GetUint64(uint64(num))
}

// GetUint64 checks if an uint64 value exists in the Filter.
func (f Filter) GetUint64(num uint64) bool {
	if len(f) == 0 {
		return num&1 == 1
	}
	cells := cellSize(len(f))
	x := uint32(num)
	high := uint32(num >> 32)
	for i := uint32(0); i < 64; i++ {
		h := hash64(x, high^i, uint64(cells))
		switch ((f)[h>>2] >> ((h & 3) * 2)) & 3 {
		case 0:
			//println("return parity", x & 1 == 1)
			return x&1 == 1
		case 1:
			//println("return false")
			return false
		case 2:
			//println("return true")
			return true
		case 3:
			x = (x >> 1) | (x << 31)
		}
		//println("hop")
	}
	//println("won't happen")
	// won't happen
	return false
}

// GetUint64Multi checks if a uint64 value exists in the Filters.
func (f Filters) GetUint64Multi(num uint64) (ret uint64) {
	n := len(f)
	if n == 0 {
		return 0
	}

	// Validate uniform filter size
	baseLen := len(f[0])
	for i := 1; i < n; i++ {
		if len(f[i]) != baseLen {
			panic("filters must be the same size")
		}
	}

	// Handle empty filters uniformly
	if baseLen == 0 {
		if num&1 == 1 {
			return 1<<n - 1
		}
		return 0
	}

	cells := cellSize(baseLen)
	high := uint32(num >> 32)
	x := uint32(num)            // Global state
	done := uint64(0)           // Bitmap tracking completed filters
	allDone := uint64(1<<n - 1) // Precomputed completion mask

	// Process 64 hops
	for j := uint32(0); j < 64; j++ {
		if done == allDone {
			break
		}

		rotate := false
		h := hash64(x, high^j, uint64(cells))
		index := h >> 2
		shift := (h & 3) * 2

		for i := 0; i < n; i++ {
			mask := uint64(1 << i)
			if done&mask != 0 {
				continue
			}

			// Directly access filter data
			val := (f[i][index] >> shift) & 3
			switch val {
			case 0:
				if x&1 == 1 {
					ret |= mask
				}
				done |= mask
			case 1:
				// Leave bit unset (false)
				done |= mask
			case 2:
				ret |= mask
				done |= mask
			case 3:
				rotate = true
			}
		}

		if rotate {
			x = (x >> 1) | (x << 31) // Rotate state
		}
	}
	return
}

// GetBytesMulti checks if a 64-byte array exists in the Filters.
func (f Filters) GetBytesMulti(data [64]byte) (ret uint64) {
	n := len(f)
	if n == 0 {
		return 0
	}

	// Validate uniform filter size
	baseLen := len(f[0])
	for i := 1; i < n; i++ {
		if len(f[i]) != baseLen {
			panic("filters must be the same size")
		}
	}

	if baseLen == 0 {
		return 0
	}

	cells := cellSize(baseLen)
	done := uint64(0)           // Bitmap tracking completed filters
	allDone := uint64(1<<n - 1) // Precomputed completion mask

	// Process 64 hops
	for j := uint32(0); j < ROUNDS; j++ {
		if done == allDone {
			break
		}
		hh := hash64(dataHash(j, data[:]), uint32(cells), uint64(cells)<<1)
		index := hh >> 3
		shift := hh & 6
		parity := byte(hh&1) == 1

		for i := 0; i < n; i++ {
			mask := uint64(1 << i)
			if done&mask != 0 {
				continue
			}

			// Directly access filter data
			val := (f[i][index] >> shift) & 3
			switch val {
			case 0:
				if parity {
					ret |= mask
				}
				done |= mask
			case 1:
				// Leave bit unset (false)
				done |= mask
			case 2:
				ret |= mask
				done |= mask
			case 3:
				// Continue processing
			}
		}
	}
	return
}

// GetBytes checks if a 64-byte array exists in the Filter.
func (f Filter) GetBytes(data [64]byte) bool {
	if len(f) == 0 {
		return false
	}
	cells := cellSize(len(f))
	//println("insert", x, high, "size", len(f))
	for i := uint32(0); i < ROUNDS; i++ {
		h := hash64(dataHash(i, data[:]), uint32(cells), uint64(cells)<<1)
		switch ((f)[h>>3] >> (h & 6)) & 3 {
		case 0:
			return byte(h&1) == 1
		case 1:
			return false
		case 2:
			return true
		case 3:
			continue
		}
	}
	return false
}

// GetString checks if a string exists in the Filter created by MakeString.
func (f Filter) GetString(str string) bool {
	if len(str) <= 7 {
		return f.GetUint64(stringToUint64(str))
	}
	return f.GetBytes(stringsToByte64(str))
}

// GetStringMulti checks if a string exists in the Filters created by MakeStringMulti.
func (f Filters) GetStringMulti(str string) uint64 {
	if len(str) <= 7 {
		return f.GetUint64Multi(stringToUint64(str))
	}
	return f.GetBytesMulti(stringsToByte64(str))
}

// GetStrings checks the two provided strings exist in the Filter created by MakeStrings.
func (f Filter) GetStrings(strs ...string) bool {
	return f.GetBytes(stringsToByte64(strs...))
}

// Number is a type constraint that represents any numeric type.
type Number interface {
	int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}

// Make creates a new Filter from a map of numeric values.
// The type T must satisfy the Number constraint.
func Make[T Number](numbers map[T]bool) Filter {
	return create(numbers, make(map[[64]byte]bool))[0]
}

// MakeBytes creates a new Filter from a map of 64-byte arrays.
func MakeBytes(data map[[64]byte]bool) Filter {
	return create(make(map[int]bool), data)[0]
}
func create[T Number](numbers map[T]bool, data map[[64]byte]bool) []Filter {
	if len(data)+len(numbers) == 0 {
		return []Filter{nil}
	}
	bytes := byteSize(grow(len(data) + len(numbers)))
	filter := make([]byte, bytes, bytes)
	var maxLoad = len(data) + len(numbers)
	for {
		var is_mutated = true
		var load int
		for is_mutated && load < maxLoad {
			var new_inserted int
			for k, v := range data {
				if v {
					new_inserted += Filter(filter).store((k[:]), 1)
				} else {
					new_inserted += Filter(filter).store((k[:]), 0)
				}
				if load+new_inserted >= maxLoad {
					break
				}
			}
			for k, v := range numbers {
				if v {
					new_inserted += Filter(filter).insert(uint64(k), 1)
				} else {
					new_inserted += Filter(filter).insert(uint64(k), 0)
				}
				if load+new_inserted >= maxLoad {
					break
				}
			}
			is_mutated = is_mutated && new_inserted > 0
			load += new_inserted
			//println("inserted", new_inserted, "is_mutated", is_mutated, "load", load)
		}
		if is_mutated {
			bytes = byteSize(grow(cellSize(bytes)))
			filter = make([]byte, bytes, bytes)
			maxLoad = grow(maxLoad)
			//println("bytes", bytes, "maxLoad", maxLoad)
		} else {
			break
		}
	}
	return []Filter{Filter(filter)}
}
func create64[T Number](filters byte, numbers map[T]uint64, data map[[64]byte]uint64) (filter []Filter) {
	if len(data)+len(numbers) == 0 {
		return make([]Filter, filters, filters)
	}
	bytes := byteSize(grow(len(data) + len(numbers)))
	filter = make([]Filter, filters, filters)
	for i := byte(0); i < filters; i++ {
		filter[i] = make([]byte, bytes, bytes)
	}
	var maxLoad = len(data) + len(numbers)
	for {
		var fs Filters
		for i := byte(0); i < filters; i++ {
			fs = append(fs, filter[i])
		}
		var is_mutated = true
		var load = make([]int, filters, filters)
	outer:
		for is_mutated {
			var new_inserted = make([]int, filters, filters)
		inner1:
			for k, v := range data {
				ins := fs.store((k[:]), v)
				for i := byte(0); i < filters; i++ {
					new_inserted[i] += ins[i]
					//new_inserted[i] += filter[i].store(k[:], byte(v >> i) & 1)
				}
				for i := byte(0); i < filters; i++ {
					if load[i]+new_inserted[i] >= maxLoad {
						break inner1
					}
				}

			}
		inner2:
			for k, v := range numbers {
				ins := fs.insert(uint64(k), v)
				for i := byte(0); i < filters; i++ {
					new_inserted[i] += ins[i]
					//new_inserted[i] += filter[i].insert(uint64(k), byte(v >> i) & 1)

				}
				for i := byte(0); i < filters; i++ {
					if load[i]+new_inserted[i] >= maxLoad {
						break inner2
					}
				}
			}
			var orInserted bool
			for i := byte(0); i < filters; i++ {
				orInserted = orInserted || new_inserted[i] > 0
				load[i] += new_inserted[i]
			}
			is_mutated = is_mutated && orInserted
			//println("inserted", new_inserted, "is_mutated", is_mutated, "load", load)
			for i := byte(0); i < filters; i++ {
				if load[i] >= maxLoad {
					break outer
				}
			}
		}
		if is_mutated {
			bytes = byteSize(grow(cellSize(bytes)))
			for i := byte(0); i < filters; i++ {
				filter[i] = make([]byte, bytes, bytes)
			}
			maxLoad = grow(maxLoad)
			//println("bytes", bytes, "maxLoad", maxLoad)
		} else {
			break
		}
	}
	return
}

func stringToUint64(s string) uint64 {
	var result uint64
	for i := 0; i < len(s) && i < 8; i++ {
		result |= uint64(s[i]) << (8 * i)
	}
	result |= uint64(len(s)) << 56
	return result
}

// stringsToByte64 builds a 64‑byte key.
// If called with exactly one string of length ≤ 63, it does:
//
//	ret[0..len-1] = str bytes
//	ret[len]      = byte(len)
//	all other ret[i] == 0
//
// Otherwise it sums SHA‑512 hashes as before.
func stringsToByte64(parts ...string) (ret [64]byte) {
	if len(parts) == 1 {
		str := parts[0]
		if len(str) <= 63 {
			// fast path for short keys
			n := copy(ret[:], str)
			ret[n] = byte(n)
			return ret
		}
	}
	// fallback: sum SHA‑512 over all parts
	for _, str := range parts {
		hash := sha512.Sum512([]byte(str))
		for i := range hash {
			ret[i] += hash[i]
		}
	}
	return ret
}

// MakeString creates a new Filter from a map of strings.
func MakeString(string_map map[string]bool) Filter {
	var data = make(map[[64]byte]bool)
	var nums = make(map[uint64]bool)
	for k, v := range string_map {
		if len(k) <= 7 {
			nums[stringToUint64(k)] = v
		} else {
			data[stringsToByte64(k)] = v
		}
	}
	return create(nums, data)[0]
}

// MakeStringMulti creates a new Filters from a map of strings.
func MakeStringMulti(multi byte, string_map map[string]uint64) []Filter {
	var data = make(map[[64]byte]uint64)
	var nums = make(map[uint64]uint64)
	for k, v := range string_map {
		if len(k) <= 7 {
			nums[stringToUint64(k)] = v
		} else {
			data[stringsToByte64(k)] = v
		}
	}
	r := create64(multi, nums, data)

	return r
}

// Make2Strings creates a new Filter from a map of 2-string arrays.
func Make2Strings(string_map map[[2]string]bool) Filter {
	var data = make(map[[64]byte]bool)
	for k, v := range string_map {
		data[stringsToByte64(k[:]...)] = v
	}
	return create(make(map[int]bool), data)[0]
}
