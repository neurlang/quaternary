package quaternary

func byteSize(n int) int {
	return (3+n)/4
}

func cellSize(n int) int {
	return n * 4
}

func grow(n int) int {
	return (3*n) / 2
}

type Filter []byte

func (f Filter) insert(num uint64, answer byte) (inserted int) {
	if len(f) == 0 {
		return 1
	}
	cells := cellSize(len(f))
	x := uint32(num)
	high := uint32(num >> 64)
	//println("insert", x, high, "size", len(f))
	for i := uint32(0); i < 64; i++ {
		h := hash64(x, high ^ i, uint64(cells))
		switch ((f)[h >> 2] >> ((h & 3) * 2)) & 3 {
		case 0:
			if answer == byte(x & 1) {
				return inserted
			}
			(f)[h >> 2] |= ((answer & 1) + 1) << ((h & 3) * 2)
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
		(f)[h >> 2] |= 3 << ((h & 3) * 2)
		x = (x >> 1) | (x << 31)
		inserted++
		
	}
	return inserted + 1
}

func (f Filter) GetInt(num int) (bool) {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetUint(num uint) (bool) {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetInt8(num int8) (bool) {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetUint8(num uint8) (bool) {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetInt16(num int16) (bool) {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetUint16(num uint16) (bool) {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetInt32(num int32) (bool) {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetUint32(num uint32) (bool) {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetInt64(num int64) (bool) {
	return f.GetUint64(uint64(num))
}
func (f Filter) GetUint64(num uint64) bool {
	if len(f) == 0 {
		return num & 1 == 1
	}
	cells := cellSize(len(f))
	x := uint32(num)
	high := uint32(num >> 64)
	for i := uint32(0); i < 64; i++ {
		h := hash64(x, high ^ i, uint64(cells))
		switch ((f)[h >> 2] >> ((h & 3) * 2)) & 3 {
		case 0:
			//println("return parity", x & 1 == 1)
			return x & 1 == 1
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

type Number interface {
	int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}

func Make[T Number](data map[T]bool) Filter {
	if len(data) == 0 {
		return Filter(nil)
	}
	bytes := byteSize(grow(len(data)))
	filter := make([]byte, bytes, bytes)
	var maxLoad = len(data)
	for {
		var is_mutated = true
		var load int
		for is_mutated && load < maxLoad {
			var new_inserted int
			for k, v := range data {
				if v {
					new_inserted += Filter(filter).insert(uint64(k), 1)
				} else {
					new_inserted += Filter(filter).insert(uint64(k), 0)
				}
			}
			is_mutated = new_inserted > 0
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
	return Filter(filter)
}
