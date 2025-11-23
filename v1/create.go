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

func bitSize(n uint64) uint64 {
	return n * 8
}

func kvPairKey(kvPair *[2][]byte) *[]byte {
	return &(kvPair[0])
}
func kvPairValue(kvPair *[2][]byte) *[]byte {
	return &(kvPair[1])
}

// Iterator is a function that yields key-value pairs
// It can be called multiple times to restart iteration
type Iterator func(yield func(kvPair [2][]byte) bool)

func create(iter Iterator, bitLimit, bloomFuncs byte) (filter []byte) {
	var size uint64
	var maxb uint64
	if bitLimit == 1 {
		iter(func(kvPair [2][]byte) bool {
			if len(*kvPairValue(&kvPair)) > 0 {
				size++
			}
			return true
		})
		maxb = 1
	} else if bitLimit > 1 {
		iter(func(kv [2][]byte) bool {
			length := uint64(len(*kvPairValue(&kv))) * 8
			size += length
			if maxb < length {
				maxb = length
			}
			return true
		})
	} else {
		iter(func(kv [2][]byte) bool {
			size += uint64(len(*kvPairValue(&kv))) * 8
			if maxb < uint64(len(*kvPairValue(&kv)))*8 {
				maxb = uint64(len(*kvPairValue(&kv))) * 8
			}
			return true
		})
	}
	bytes := byteSize(grow(size))
	filter = make([]byte, bytes+2, bytes+2)
	filter[bytes+1] = bitLimit
	filter[bytes] = bloomFuncs
	var maxLoad = size
	for {
		// BLOOM STAGE
		var is_mutated = bloomFuncs > 0
		var load uint64
		for is_mutated && load < maxLoad {
			var bloom_inserted uint64
			iter(func(kv [2][]byte) bool {
				ins := put(filter, *kvPairKey(&kv), bloomFuncs)
				bloom_inserted += uint64(ins)
				if load+bloom_inserted >= maxLoad {
					return false
				}
				return true
			})
			is_mutated = is_mutated && bloom_inserted > 0
			//println("inserted", bloom_inserted, "is_mutated", is_mutated, "load", load, "maxLoad", maxLoad)
			load += bloom_inserted
		}
		if is_mutated {
			bytes = byteSize(grow(cellSize(bytes)))
			filter = make([]byte, bytes+2, bytes+2)
			filter[bytes+1] = bitLimit
			filter[bytes] = bloomFuncs
			maxLoad = grow(uint64(maxLoad))
			//println("bytes", bytes, "maxLoad", maxLoad)
			continue
		}
		// QUATERNARY STAGE
		is_mutated = true
		for is_mutated && load < maxLoad {
			var new_inserted uint64
			iter(func(kv [2][]byte) bool {
				if bitLimit != 0 && (len(*kvPairValue(&kv))) != int(bitLimit+7)/8 {
					panic("inserting value exceeding bit limit when bit limit set")
				}
				stored := bitLimit
				if maxb < 256 && byte(maxb) < stored {
					stored = byte(maxb)
				}
				if 8*len(*kvPairValue(&kv)) < 256 && byte(8*len(*kvPairValue(&kv))) < stored {
					stored = byte(8 * len(*kvPairValue(&kv)))
				}
				ins := store(filter, *kvPairKey(&kv), *kvPairValue(&kv), stored)
				new_inserted += ins
				if load+new_inserted >= maxLoad {
					return false
				}
				return true
			})
			is_mutated = is_mutated && new_inserted > 0
			//println("inserted", new_inserted, "is_mutated", is_mutated, "load", load, "maxLoad", maxLoad)
			load += new_inserted
		}
		if is_mutated {
			bytes = byteSize(grow(cellSize(bytes)))
			filter = make([]byte, bytes+2, bytes+2)
			filter[bytes+1] = bitLimit
			filter[bytes] = bloomFuncs
			maxLoad = grow(uint64(maxLoad))
			//println("bytes", bytes, "maxLoad", maxLoad)
			continue
		} else {
			break
		}
	}
	return
}
