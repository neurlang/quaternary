package v1

func byteSize(n uint64) uint64 {
	return (3 + n) / 4
}

func cellSize(n uint64) uint64 {
	return n * 4
}

func grow(n uint64) uint64 {
	return (3*n + 1) / 2
}

func create(data map[string][]byte, bitLimit byte) (filter []byte) {
	var size uint64
	var maxb uint64
	if bitLimit == 1 {
		for _, v := range data {
			if len(v) > 0 {
				size++
			}
		}
		maxb = 1
	} else if bitLimit > 1 {
		for _, v := range data {
			length := uint64(len(v)) * 8
			size += length
			if maxb < length {
				maxb = length
			}
		}
	} else {
		for _, v := range data {
			size += uint64(len(v)) * 8
			if maxb < uint64(len(v))*8 {
				maxb = uint64(len(v)) * 8
			}
		}
	}
	bytes := byteSize(grow(size))
	filter = make([]byte, bytes+1, bytes+1)
	filter[bytes] = bitLimit
	var maxLoad = size
	for {
		var is_mutated = true
		var load uint64
		for is_mutated {
			var new_inserted uint64
		inner1:
			for k, v := range data {
				if bitLimit != 0 && (len(v)) != int(bitLimit+7)/8 {
					panic("inserting value exceeding bit limit when bit limit set")
				}
				stored := bitLimit
				if maxb < 256 && byte(maxb) < stored {
					stored = byte(maxb)
				}
				if 8*len(v) < 256 && byte(8*len(v)) < stored {
					stored = byte(8 * len(v))
				}
				ins := store(filter, []byte(k), v, stored)
				new_inserted += ins
				if load+new_inserted >= maxLoad {
					break inner1
				}
			}
			is_mutated = is_mutated && new_inserted > 0
			//println("inserted", new_inserted, "is_mutated", is_mutated, "load", load, "maxLoad", maxLoad)
			load += new_inserted
		}
		if is_mutated {
			bytes = byteSize(grow(cellSize(bytes)))
			filter = make([]byte, bytes+1, bytes+1)
			filter[bytes] = bitLimit
			maxLoad = grow(uint64(maxLoad))
			//println("bytes", bytes, "maxLoad", maxLoad)
		} else {
			break
		}
	}
	return
}
