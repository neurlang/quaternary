package quaternary

import "crypto/sha512"

func byteSize(n int) int {
	return (3 + n) / 4
}

func cellSize(n int) int {
	return n * 4
}

func grow(n int) int {
	return (3 * n + 1) / 2
}

type Filter []byte

func (f Filter) store(data []byte, answer byte) (inserted int) {
	if len(f) == 0 {
		return 1
	}
	cells := cellSize(len(f))
	//println("insert", string(data), "size", len(f))
	for i := uint32(0); i < 512; i++ {
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
	for i := uint32(0); i < 64; i++ {
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

func (f Filter) GetInt(num int) bool {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetUint(num uint) bool {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetInt8(num int8) bool {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetUint8(num uint8) bool {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetInt16(num int16) bool {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetUint16(num uint16) bool {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetInt32(num int32) bool {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetUint32(num uint32) bool {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetInt64(num int64) bool {
	return f.GetUint64(uint64(num))
}
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

func (f Filter) GetBytes(data [64]byte) bool {
	if len(f) == 0 {
		return false
	}
	cells := cellSize(len(f))
	//println("insert", x, high, "size", len(f))
	for i := uint32(0); i < 512; i++ {
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

func (f Filter) GetString(str string) bool {
	if len(str) <= 7 {
		return f.GetUint64(stringToUint64(str))
	}
	return f.GetBytes(stringsToByte64(str))
}

func (f Filter) GetStrings(strs ...string) bool {
	return f.GetBytes(stringsToByte64(strs...))
}

type Number interface {
	int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}

func Make[T Number](numbers map[T]bool) Filter {
	return create(numbers, make(map[[64]byte]bool))[0]
}
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

func stringToUint64(s string) uint64 {
	var result uint64
	for i := 0; i < len(s) && i < 8; i++ {
		result |= uint64(s[i]) << (8 * i)
	}
	result |= uint64(len(s)) << 56
	return result
}

func stringsToByte64(s ...string) (ret [64]byte) {
	for _, str := range s {
		var hash = sha512.Sum512([]byte(str))
		for i := range hash {
			ret[i] += hash[i]
		}
	}
	return
}

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

func Make2Strings(string_map map[[2]string]bool) Filter {
	var data = make(map[[64]byte]bool)
	for k, v := range string_map {
		data[stringsToByte64(k[:]...)] = v
	}
	return create(make(map[int]bool), data)[0]
}
