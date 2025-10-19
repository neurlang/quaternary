package v1

import "encoding/binary"
import "crypto/sha512"

// get checks if an array exists in the Filters.
func get(f []byte, data []byte, anslen uint64) (ret []byte) {
	if len(f) <= 0 {
		return nil
	}
	if anslen == 0 {
		return nil
	}
	var datb [64]byte
	datb = sha512.Sum512(data)

	baseSize := uint64(len(f))
	bitLimit := f[len(f)-1]
	if bitLimit != 0 {
		if uint64(bitLimit) < anslen {
			panic("stored bit limit smaller than required answer")
		}
	}

	cells := cellSize(baseSize - 1)
	storedBits := uint64(bitLimit)
	if storedBits == 0 {
		storedBits = anslen
	}
	if storedBits >= cells {
		panic("bad")
	}
	cells -= storedBits - 1

	//println(baseSize, bitLimit, storedBits, cells)
	//println(storedBits, anslen)

	if storedBits > anslen {
		storedBits = anslen
	}

	ret = make([]byte, (storedBits+7)/8, (storedBits+7)/8)
	done := make([]byte, (storedBits+7)/8, (storedBits+7)/8)

	// Process rounds
outer:
	for roundx := uint32(1); roundx < ROUNDS; roundx++ {
		for roundy := uint32(0); roundy < roundx; roundy++ {
			var allDone uint64
			for i := range done {
				for j := 0; j < 8; j++ {
					if (done[i]>>j)&1 == 1 {
						allDone++
					}
				}
			}
			//println(storedBits, allDone)
			if storedBits == allDone {
				break outer
			}
			x := binary.BigEndian.Uint32(datb[4*roundx:])
			y := binary.BigEndian.Uint32(datb[4*roundy:])
			hh := hash64(x, y, uint64(cells)<<1)

			//println("hh:", hh, string(data), x, y, "load", storedBits)
			parity := byte(hh&1) == 1

			for i := uint64(0); i < storedBits; i++ {
				mask := byte(1 << (i & 7))
				if done[i>>3]&mask != 0 {
					continue
				}
				h := hh + (i << 1)

				pos := h >> 3
				shift := h & 6
				// Directly access filter data
				val := (f[pos] >> shift) & 3

				switch val {
				case 0:
					if parity {
						ret[len(ret)-int(i>>3)-1] |= mask
					}
					done[i>>3] |= mask

				case 1:
					// Leave bit unset (false)
					done[i>>3] |= mask

				case 2:
					ret[len(ret)-int(i>>3)-1] |= mask
					done[i>>3] |= mask
				case 3:
					// Continue processing
				}
			}
		}
	}
	return
}
