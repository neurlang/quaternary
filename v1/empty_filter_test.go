package v1

import "testing"

func TestEmptyAndNilFiltersAreAbsent(t *testing.T) {
	emptyBool := Make(map[int]bool{}, 0)
	if len(emptyBool) != 2 {
		t.Fatalf("empty bool filter length = %d, want 2", len(emptyBool))
	}
	if got := GetBool(emptyBool, 1); got {
		t.Fatalf("GetBool(empty) = true, want false")
	}
	if got := GetBoolInt(emptyBool, 1); got {
		t.Fatalf("GetBoolInt(empty) = true, want false")
	}
	if value, found := GetBools(emptyBool, 1); value || found {
		t.Fatalf("GetBools(empty) = %v, %v, want false, false", value, found)
	}
	if got := Get(emptyBool, 1, 1); got != nil {
		t.Fatalf("Get(empty) = %x, want nil", got)
	}
	if got := GetNum(emptyBool, 1, 1); got != 0 {
		t.Fatalf("GetNum(empty) = %d, want 0", got)
	}

	key32 := [32]byte{1}
	emptyNum := Make(map[[32]byte]uint64{}, 64)
	if got := GetNum(emptyNum, 64, key32); got != 0 {
		t.Fatalf("GetNum(empty uint64) = %d, want 0", got)
	}

	for _, bloomFuncs := range []byte{0, 1, 2, 3, 4} {
		f := New(map[int]bool{}, 0, bloomFuncs)
		if value, found := GetBools(f, 1); value || found {
			t.Fatalf("GetBools(bloom %d) = %v, %v, want false, false", bloomFuncs, value, found)
		}
		if got := GetBool(f, 1); got {
			t.Fatalf("GetBool(bloom %d) = true, want false", bloomFuncs)
		}
		if got := GetBoolInt(f, 1); got {
			t.Fatalf("GetBoolInt(bloom %d) = true, want false", bloomFuncs)
		}
	}

	if value, found := GetBools[int](nil, 1); value || found {
		t.Fatalf("GetBools(nil) = %v, %v, want false, false", value, found)
	}
	if got := GetBool[int](nil, 1); got {
		t.Fatalf("GetBool(nil) = true, want false")
	}
	if got := GetBoolInt(nil, 1); got {
		t.Fatalf("GetBoolInt(nil) = true, want false")
	}
	if got := Get[int](nil, 8, 1); got != nil {
		t.Fatalf("Get(nil) = %x, want nil", got)
	}
	if got := GetNum[int](nil, 64, 1); got != 0 {
		t.Fatalf("GetNum(nil) = %d, want 0", got)
	}
	if got := GetNum[[32]byte](nil, 64, key32); got != 0 {
		t.Fatalf("GetNum(nil, 32-byte key) = %d, want 0", got)
	}

	short := []byte{0}
	if value, found := GetBools(short, 1); value || found {
		t.Fatalf("GetBools(short) = %v, %v, want false, false", value, found)
	}
	if got := GetBool(short, 1); got {
		t.Fatalf("GetBool(short) = true, want false")
	}
}

func TestCorruptShortBodyStillPanics(t *testing.T) {
	// One body byte is four cells. A bit limit of 4 does not fit, so this
	// stays a corrupt filter and must still panic.
	f := []byte{0, 0, 4}
	defer func() {
		if recover() == nil {
			t.Fatal("GetBool on a short body did not panic")
		}
	}()
	_ = GetBool(f, 1)
}
