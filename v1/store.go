package v1

import "encoding/binary"
import "crypto/sha512"

const ROUNDS = 16

func store(fs []byte, data []byte, answer []byte, bitLimit byte) uint64 {
	if len(fs) == 0 {
		return 0
	}
	if len(answer) == 0 {
		return 0
	}

	var datb [64]byte
	datb = sha512.Sum512(data)

	baseSize := uint64(len(fs))
	cells := cellSize(baseSize - 2)
	storedBits := uint64(bitLimit)
	if storedBits == 0 {
		storedBits = uint64(len(answer)) * 8
		if bitLimit != 0 {
			if uint64(bitLimit) < storedBits {
				storedBits = uint64(bitLimit)
			}
		}
	}
	if storedBits == 0 {
		return 0
	}
	cells -= storedBits - 1

	// Track active filters and their insertion counts
	active := make([]bool, storedBits, storedBits)
	mutated := make([]bool, storedBits, storedBits)

	//println(storedBits, cells, baseSize)

	// Process rounds
outer:
	for roundx := uint32(1); roundx < ROUNDS; roundx++ {
		for roundy := uint32(0); roundy < roundx; roundy++ {
			anyActive := false
			x := binary.BigEndian.Uint32(datb[4*roundx:])
			y := binary.BigEndian.Uint32(datb[4*roundy:])
			hh := hash64(x, y, uint64(cells)<<1)

			//println("hh:", hh, string(data), x, y, "store", storedBits)

			for i := uint64(0); i < storedBits; i++ {
				if active[i] {
					continue
				}
				anyActive = true
				h := hh + (i << 1)

				pos := h >> 3
				shift := h & 6
				state := (fs[pos] >> shift) & 3
				bit := byte((answer[len(answer)-int(i>>3)-1] >> (i & 7)) & 1)

				switch state {
				case 0:
					if bit == byte(h&1) {
						active[i] = true
					} else {
						fs[pos] |= (bit + 1) << shift
						mutated[i] = true
					}
				case 1:
					if bit == 0 {
						active[i] = true
					} else {
						fs[pos] |= 3 << shift
						mutated[i] = true
					}
				case 2:
					if bit == 1 {
						active[i] = true
					} else {
						fs[pos] |= 3 << shift
						mutated[i] = true
					}
				}
			}

			if !anyActive {
				break outer
			}
		}
	}
	var totalInserted uint64
	for i := uint64(0); i < storedBits; i++ {
		if mutated[i] {
			totalInserted++
		}
	}
	// Finalize insertion counts
	return totalInserted
}

// bloom put
func put(fs []byte, data []byte, funcs byte) (ret byte) {
	if len(fs) == 0 {
		return 0
	}
	if funcs == 0 {
		return 0
	}

	var datb [64]byte
	datb = sha512.Sum512(data)

	baseSize := uint64(len(fs))
	cells := bitSize(baseSize - 2)

	for roundx := uint32(1); roundx < ROUNDS; roundx++ {
		for roundy := uint32(0); roundy < roundx; roundy++ {
			x := binary.BigEndian.Uint32(datb[4*roundx:])
			y := binary.BigEndian.Uint32(datb[4*roundy:])
			hh := hash64(x, y, uint64(cells))
			mask := byte(1) << (byte(hh) & 7)
			pos := hh >> 3
			if fs[pos]&mask == 0 {
				fs[pos] |= mask
				ret++
			}
			funcs--
			if funcs == 0 {
				return
			}
		}
	}
	return
}
