package v1

import (
	"testing"
)

func TestMakeNumberMapping(t *testing.T) {
	for j := byte(100); j > 3; j-- {
		f := New(map[byte]bool{
			41: true,
			52: false,
		}, 1, j)

		a, b := GetBools(f, byte(41))

		if a != true || b != true {
			panic("ab")
		}

		c, d := GetBools(f, byte(52))

		if c != false || d != true {
			panic("cd")
		}

		for i := 0; i < 256; i++ {
			if i == 41 || i == 52 {
				continue
			}
			_, e := GetBools(f, byte(i))
			if e {
				panic("e")
			}
		}
	}
}

func TestMakeNumberMappingBuffers(t *testing.T) {
	for j := byte(100); j > 6; j-- {
		f := New(map[byte]string{
			41: "hello",
			52: "world",
		}, 0, j)

		if "hello" != (string(Get(f, 8*5, byte(41)))) {
			panic("ab")
		}
		if "world" != (string(Get(f, 8*5, byte(52)))) {
			panic("cd")
		}

		for i := 0; i < 256; i++ {
			if i == 41 || i == 52 {
				continue
			}
			e := Get(f, 8*5, byte(i))
			if e != nil {
				panic("e")
			}
		}
	}
}

func withHashstructureIntegration(h Hashstructure, f func()) {
	if Integration != nil {
		panic("pre-existing hashstructure integration: parallel tests?")
	}
	Integration = h
	// Clear the hook even if f panics, so a failing test cannot leak it into the
	// tests that run after it.
	defer func() { Integration = nil }()

	if f != nil {
		f()
	}

	if Integration == nil {
		panic("cleared hashstructure integration: parallel tests?")
	}
}

func TestMakeNumberMappingHashstructure(t *testing.T) {
	hashstructure := Hashstructure(func(interface{}) []byte {
		return []byte{1, 2, 3}
	})
	withHashstructureIntegration(hashstructure, func() {
		for j := byte(100); j > 3; j-- {
			f := New(map[byte]bool{
				41: true,
				52: false,
			}, 1, j)

			a, b := GetBools(f, byte(41))

			if a != true || b != true {
				panic("ab")
			}

			c, d := GetBools(f, byte(52))

			if c != false || d != true {
				panic("cd")
			}

			for i := 0; i < 256; i++ {
				if i == 41 || i == 52 {
					continue
				}
				_, e := GetBools(f, byte(i))
				if e {
					panic("e")
				}
			}
		}
	})
}

func TestMakeNumberMappingBuffersHashstructure(t *testing.T) {
	hashstructure := Hashstructure(func(v interface{}) []byte {
		return []byte{v.(byte)}
	})
	withHashstructureIntegration(hashstructure, func() {
		for j := byte(100); j > 6; j-- {
			f := New(map[byte]string{
				41: "hello",
				52: "world",
			}, 0, j)

			if "hello" != (string(Get(f, 8*5, byte(41)))) {
				panic("ab")
			}
			if "world" != (string(Get(f, 8*5, byte(52)))) {
				panic("cd")
			}

			for i := 0; i < 256; i++ {
				if i == 41 || i == 52 {
					continue
				}
				e := Get(f, 8*5, byte(i))
				if e != nil {
					panic("e")
				}
			}
		}
	})
}
