package v1

import (
	"bytes"
	"sort"
)

// sortPairsByDigest sorts pairs by their 32-byte key digest, then repacks them.
//
// create inserts the pairs in the order it is given. The cells a build converges
// to do not depend on that order, but the number of mutations create counts
// against its load budget does, and with it whether an attempt is accepted or
// regrown. Feeding create a canonical order makes the whole build, including its
// final size, independent of Go's random map iteration order.
//
// Digests are uniformly distributed, so a counting sort on their first 16 bits
// leaves tiny buckets that are finished with a full comparison. That is close to
// linear, unlike a comparison sort over 32-byte keys.
func sortPairsByDigest(pairs [][2][]byte) {
	less := func(a, b [2][]byte) bool { return bytes.Compare(a[0], b[0]) < 0 }
	n := len(pairs)
	if n < 4096 {
		sort.Slice(pairs, func(i, j int) bool { return less(pairs[i], pairs[j]) })
		repackPairs(pairs)
		return
	}
	var start [65537]int
	for i := range pairs {
		start[int(pairs[i][0][0])<<8|int(pairs[i][0][1])+1]++
	}
	for b := 1; b <= 65536; b++ {
		start[b] += start[b-1]
	}
	next := start // copy of the bucket offsets
	out := make([][2][]byte, n)
	for i := range pairs {
		b := int(pairs[i][0][0])<<8 | int(pairs[i][0][1])
		out[next[b]] = pairs[i]
		next[b]++
	}
	copy(pairs, out)
	for b := 0; b < 65536; b++ {
		if lo, hi := start[b], start[b+1]; hi-lo > 1 {
			sub := pairs[lo:hi]
			sort.Slice(sub, func(i, j int) bool { return less(sub[i], sub[j]) })
		}
	}
	repackPairs(pairs)
}

// repackPairs copies every key and value into one contiguous block each, in the
// current order. Keys and values are separate small allocations, laid out roughly
// in map iteration order; after sorting they are scattered over the heap, and
// create walks all pairs once per pass, so without repacking every pass chases
// pointers (about 30% slower at a million keys). Contiguous blocks keep the
// passes sequential in memory.
func repackPairs(pairs [][2][]byte) {
	total := 0
	for i := range pairs {
		total += len(pairs[i][1])
	}
	keys := make([]byte, 32*len(pairs))
	vals := make([]byte, total)
	off := 0
	for i := range pairs {
		k := keys[32*i : 32*i+32 : 32*i+32]
		copy(k, pairs[i][0])
		l := len(pairs[i][1])
		v := vals[off : off+l : off+l]
		copy(v, pairs[i][1])
		off += l
		pairs[i][0], pairs[i][1] = k, v
	}
}
