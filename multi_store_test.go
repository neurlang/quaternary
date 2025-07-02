package quaternary

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

// Assumes Filter and Filters types and their store methods are defined in this package.

func TestMultiFilterStoreEquivalence(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Test parameters
	nFilters := 5
	filterSize := 256 // number of bytes per filter
	dataLen := 32
	nTrials := 100

	for trial := 0; trial < nTrials; trial++ {
		// Initialize random filters
		fs := make(Filters, nFilters)
		singles := make([]Filter, nFilters)
		for i := 0; i < nFilters; i++ {
			fs[i] = make([]byte, filterSize)
			singles[i] = make([]byte, filterSize)
		}

		// Random data and answer
		data := make([]byte, dataLen)
		rand.Read(data)
		// answer as bits for multi and a single byte for single store
		answerBits := rand.Uint64()

		// Clone filters for comparison after multi-store
		preMulti := make([]Filter, nFilters)
		for i := range fs {
			preMulti[i] = make([]byte, filterSize)
			copy(preMulti[i], fs[i])
		}

		// Run multi-filter store
		multiInserted := fs.store(data, answerBits)

		// Run single-filter stores on fresh copies
		singleInserted := make([]int, nFilters)
		for j := 0; j < nFilters; j++ {
			// pick bit j of answerBits
			ans := byte((answerBits >> uint(j)) & 1)
			singleInserted[j] = singles[j].store(data, ans)
		}

		// Compare inserted counts
		if !reflect.DeepEqual(multiInserted, singleInserted) {
			t.Errorf("Inserted counts differ on trial %d: multi=%v, single=%v", trial, multiInserted, singleInserted)
		}

		// Compare filter contents
		for j := 0; j < nFilters; j++ {
			if !bytes.Equal(fs[j], singles[j]) {
				t.Errorf("Filter data differ on trial %d, filter %d", trial, j)
			}
		}
	}
}

// TestMultiFilterInsertEquivalence verifies that Filters.insert behaves identically
// to calling Filter.insert individually.
func TestMultiFilterInsertEquivalence(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Test parameters
	nFilters := 5
	filterSize := 256
	nTrials := 100

	for trial := 0; trial < nTrials; trial++ {
		// Initialize random filters
		fs := make(Filters, nFilters)
		singles := make([]Filter, nFilters)
		for i := 0; i < nFilters; i++ {
			fs[i] = make([]byte, filterSize)
			singles[i] = make([]byte, filterSize)
		}

		// Random seed and answer
		num := rand.Uint64()
		answerBits := rand.Uint64()

		// Run multi-filter insert
		multiInserted := fs.insert(num, answerBits)

		// Run single-filter inserts on fresh copies
		singleInserted := make([]int, nFilters)
		for j := 0; j < nFilters; j++ {
			ans := byte((answerBits >> uint(j)) & 1)
			singleInserted[j] = singles[j].insert(num, ans)
		}

		// Compare inserted counts
		if !reflect.DeepEqual(multiInserted, singleInserted) {
			t.Errorf("Insert counts differ on trial %d: multi=%v, single=%v", trial, multiInserted, singleInserted)
		}

		// Compare filter contents
		for j := 0; j < nFilters; j++ {
			if !bytes.Equal(fs[j], singles[j]) {
				t.Errorf("Insert filter data differ on trial %d, filter %d", trial, j)
			}
		}
	}
}
