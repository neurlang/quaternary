package v1

import (
	"fmt"
	"strings"
	"testing"
)

// TestGetNumRejectsMoreThan64Bits checks that GetNum names the problem when asked
// for more bits than a uint64 can hold. It used to fail with a raw
// "slice bounds out of range [-1:]" from copy(buf[8-len(b):8], b), and only when
// the key's value was non-empty.
func TestGetNumRejectsMoreThan64Bits(t *testing.T) {
	wide := Make(map[string][]byte{"k": make([]byte, 9)}, 72)
	empty := Make(map[string][]byte{}, 0)
	for name, f := range map[string][]byte{"72-bit filter": wide, "empty filter": empty, "nil filter": nil} {
		for _, bits := range []uint64{65, 72, 100, 255, 256, 1 << 20} {
			func() {
				defer func() {
					msg := fmt.Sprint(recover())
					if !strings.Contains(msg, "exceeds the 64 bits") {
						t.Errorf("%s, valBitSize=%d: want a panic mentioning \"exceeds the 64 bits\", got %q", name, bits, msg)
					}
				}()
				GetNum(f, bits, "k")
			}()
		}
	}
	// 64 bits is still fine, and reads the low 64 bits of a wider value
	v := make([]byte, 9)
	v[8] = 0x2A
	f := Make(map[string][]byte{"k": v}, 72)
	if got := GetNum(f, 64, "k"); got != 0x2A {
		t.Fatalf("GetNum(f, 64, k) = %#x, want 0x2a", got)
	}
}
