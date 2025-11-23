package v1

import (
	"testing"
)

func TestMakeNumberMapping(t *testing.T) {
	for j := byte(100); j > 3; j-- {
		println(j)
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
		println(j)
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
