package v1

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type payload struct {
	BitLimit byte              `json:"bitLimit"`
	Items    map[string][]byte `json:"items"`
}

// MakeScaffold generates the filter based on map m
func MakeScaffold(m map[string][]byte, bitLimit byte) []byte {
	out := payload{
		BitLimit: bitLimit,
		Items:    make(map[string][]byte, len(m)),
	}

	for k, v := range m {
		// Treat nil slice as zero
		if v == nil {
			v = []byte{}
		}

		if bitLimit == Unlimited {
			// store as-is
			out.Items[k] = append([]byte(nil), v...)
			continue
		}

		// build big.Int from the provided bytes (interpret as big-endian unsigned)
		val := new(big.Int).SetBytes(v)

		// If the value needs more than bitLimit bits -> panic
		if val.BitLen() > int(bitLimit) {
			panic(fmt.Sprintf("Make: value for key %q exceeds bitLimit (%d bits required > %d)", k, val.BitLen(), bitLimit))
		}

		// Mask to the lower bitLimit bits (not strictly necessary because we checked above,
		// but ensures consistent stored representation)
		mask := new(big.Int).Lsh(big.NewInt(1), uint(bitLimit)) // 1 << bitLimit
		mask.Sub(mask, big.NewInt(1))                           // mask = (1<<bitLimit)-1
		masked := new(big.Int).And(val, mask)

		// Convert masked to exactly the number of bytes needed to hold bitLimit bits
		bytesNeeded := int((int(bitLimit) + 7) / 8)
		bs := masked.Bytes() // big-endian, may be shorter (or empty for zero)

		// pad with leading zeros to bytesNeeded
		if len(bs) < bytesNeeded {
			pad := make([]byte, bytesNeeded-len(bs))
			bs = append(pad, bs...)
		}
		out.Items[k] = bs
	}

	j, err := json.Marshal(out)
	if err != nil {
		// per spec: panic on unexpected serialization error
		panic(fmt.Sprintf("Make: json.Marshal failed: %v", err))
	}
	return j
}

// GetScaffold retrieves an item based on string key and value bit size
func GetScaffold(f []byte, valBitSize uint64, key string) []byte {
	// valBitSize == 0 returns nil
	if valBitSize == 0 {
		return nil
	}

	var p payload
	if err := json.Unmarshal(f, &p); err != nil {
		panic(fmt.Sprintf("Get: json.Unmarshal failed: %v", err))
	}

	// If stored structure has a limit, ensure requested valBitSize <= stored bitLimit
	if p.BitLimit != Unlimited {
		if valBitSize > uint64(p.BitLimit) {
			panic(fmt.Sprintf("Get: requested valBitSize %d exceeds stored bitLimit %d", valBitSize, p.BitLimit))
		}
	}

	// find key
	stored, ok := p.Items[key]
	if !ok {
		// not found -> per spec we return zeroes
		return make([]byte, int((int(valBitSize)+7)/8), int((int(valBitSize)+7)/8))
	}

	// Interpret stored bytes as big-endian unsigned integer
	val := new(big.Int).SetBytes(stored)

	// Mask to lower valBitSize bits
	mask := new(big.Int).Lsh(big.NewInt(1), uint(valBitSize))
	mask.Sub(mask, big.NewInt(1))
	masked := new(big.Int).And(val, mask)

	// Return exactly bytesNeeded bytes (big-endian)
	bytesNeeded := int((int(valBitSize) + 7) / 8)
	bs := masked.Bytes()
	if len(bs) < bytesNeeded {
		pad := make([]byte, bytesNeeded-len(bs))
		bs = append(pad, bs...)
	}
	return bs
}
