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
	if len(data) > 63 {
		datb = sha512.Sum512(data)
	} else {
		copy(datb[:], data)
		datb[63] = byte(len(data))
		for i := range datb {
			datb[(i+1)&63] ^= datb[i]
		}
	}

	baseSize := uint64(len(fs))
	cells := cellSize(baseSize - 1)
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

	var totalInserted uint64
	//println(storedBits, cells, baseSize)

	// Process rounds
outer:
	for roundx := uint32(0); roundx < ROUNDS; roundx++ {
		for roundy := roundx + 1; roundy < ROUNDS; roundy++ {
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
						totalInserted++
						active[i] = true
					}
				case 1:
					if bit == 0 {
						active[i] = true
					} else {
						fs[pos] |= 3 << shift
						totalInserted++
					}
				case 2:
					if bit == 1 {
						active[i] = true
					} else {
						fs[pos] |= 3 << shift
						totalInserted++
					}
				}
			}

			if !anyActive {
				break outer
			}
		}
	}

	// Finalize insertion counts
	return totalInserted
}
