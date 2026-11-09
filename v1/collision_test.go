package v1

import (
	"fmt"
	"strings"
	"testing"
)

func TestMakeRejectsConflictingKeyEncodings(t *testing.T) {
	type namedString string
	type record struct{ hidden, Exported int }
	cases := map[string]func(){
		"invalid UTF-8 array": func() { Make(map[[2]string]bool{{"\xff", "a"}: true, {"\xfe", "a"}: false}, 0) },
		"named string":        func() { Make(map[namedString]bool{"\xff": true, "\xfe": false}, 0) },
		"unexported field":    func() { Make(map[record]bool{{1, 7}: true, {2, 7}: false}, 0) },
		"byte values":         func() { New(map[namedString][]byte{"\xff": {1}, "\xfe": {2}}, 0, 3) },
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if got := recover(); got == nil || !strings.Contains(fmt.Sprint(got), "conflicting values for identical key encoding") {
					t.Fatalf("got panic %v, want conflicting key encoding diagnostic", got)
				}
			}()
			build()
		})
	}
}

func TestDuplicateKeyEncodingsWithEqualValues(t *testing.T) {
	for _, value := range []bool{false, true} {
		filter := Make(map[[2]string]bool{{"\xff", "a"}: value, {"\xfe", "a"}: value}, 0)
		for _, key := range [][2]string{{"\xff", "a"}, {"\xfe", "a"}} {
			if got := GetBool(filter, key); got != value {
				t.Fatalf("key %T: got %v, want %v", key, got, value)
			}
		}
	}
}
