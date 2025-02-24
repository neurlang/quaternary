# quaternary
Quaternary Filter is a:

* **62**x to **112**x smaller `map[int]bool`, constructed using generic `Make(...)`
* **69**x to **439**x smaller `map[string]bool`, constructed using `MakeString(...)`
* **604**x to **1094**x smaller `map[[2]string]bool`, constructed using `Make2Strings(...)`
* **438**x to **887**x smaller `map[[64]byte]bool`, constructed using `MakeBytes(...)`

## Use this

* If you don't need to insert any more elements into the map
* If you don't need to loop over the map
* If you don't need to know which keys exist

## Example

```go
package main

import "github.com/neurlang/quaternary"

func main() {
	var filter = quaternary.Make(map[int]bool{
		5: true,
		55: false,
	})

	if (filter.GetInt(5) != true) {
		panic("5 not true")
	}
	if (filter.GetInt(55) != false) {
		panic("55 not false")
	}

	println("5:", filter.GetInt(5), "55:", filter.GetInt(55), "size:", len(filter), "B")
}
```

## Usage in file formats / networking protocols

Since Quaternary Filter is []byte internally (this won't change), it can be
easily serialized to disk/transmitted over the wire.
We reserve the right to break the format in the future, but will do so only if bumping the 0.X.0 version.
The Quaternary Filter is currently alpha-quality. Patches welcome.

## Advices

* If you lookup a value which wasn't inserted, you get a garbage boolean. This is a known feature and won't be fixed.
* Use MakeBytes() to key uuids for optimal speed. You can pack up to 4 uuids into single `[64]byte`.
* Use MakeString() to key strings with string shorter than <= 7 bytes for optimal speeds (avoid longer strings).
* Make2Strings() is slow, but hey, it's here. Fixme.

## Memory efficiency/performance

```
$ go test --bench=ReadAll
[Numeric] Memory used by the 10 element map: 528 bytes
[Numeric] Memory used by the 10 element quarternary: 8 bytes
[Numeric] Quarternary is: 66x smaller
[Numeric] Memory used by the 100 element map: 7016 bytes
[Numeric] Memory used by the 100 element quarternary: 113 bytes
[Numeric] Quarternary is: 62x smaller
[Numeric] Memory used by the 1000 element map: 105912 bytes
[Numeric] Memory used by the 1000 element quarternary: 1688 bytes
[Numeric] Quarternary is: 62x smaller
[Numeric] Memory used by the 10000 element map: 848840 bytes
[Numeric] Memory used by the 10000 element quarternary: 7500 bytes
[Numeric] Quarternary is: 113x smaller
[Numeric] Memory used by the 100000 element map: 5317304 bytes
[Numeric] Memory used by the 100000 element quarternary: 112500 bytes
[Numeric] Quarternary is: 47x smaller
[Numeric] Memory used by the 1000000 element map: 77999736 bytes
[Numeric] Memory used by the 1000000 element quarternary: 1125000 bytes
[Numeric] Quarternary is: 69x smaller
[One string] Memory used by the 10 element map: 6024 bytes
[One string] Memory used by the 10 element quarternary: 8 bytes
[One string] Quarternary is: 753x smaller
[One string] Memory used by the 100 element map: 16992 bytes
[One string] Memory used by the 100 element quarternary: 113 bytes
[One string] Quarternary is: 150x smaller
[One string] Memory used by the 1000 element map: 218848 bytes
[One string] Memory used by the 1000 element quarternary: 750 bytes
[One string] Quarternary is: 291x smaller
[One string] Memory used by the 10000 element map: 1745504 bytes
[One string] Memory used by the 10000 element quarternary: 7500 bytes
[One string] Quarternary is: 232x smaller
[One string] Memory used by the 100000 element map: 12061184 bytes
[One string] Memory used by the 100000 element quarternary: 75000 bytes
[One string] Quarternary is: 160x smaller
[One string] Memory used by the 1000000 element map: 165179824 bytes
[One string] Memory used by the 1000000 element quarternary: 750000 bytes
[One string] Quarternary is: 220x smaller
[Two strings] Memory used by the 10 element map: 6856 bytes
[Two strings] Memory used by the 10 element quarternary: 4 bytes
[Two strings] Quarternary is: 1714x smaller
[Two strings] Memory used by the 100 element map: 27072 bytes
[Two strings] Memory used by the 100 element quarternary: 38 bytes
[Two strings] Quarternary is: 712x smaller
[Two strings] Memory used by the 1000 element map: 392800 bytes
[Two strings] Memory used by the 1000 element quarternary: 375 bytes
[Two strings] Quarternary is: 1047x smaller
[Two strings] Memory used by the 10000 element map: 3242816 bytes
[Two strings] Memory used by the 10000 element quarternary: 3750 bytes
[Two strings] Quarternary is: 864x smaller
[Two strings] Memory used by the 100000 element map: 22651728 bytes
[Two strings] Memory used by the 100000 element quarternary: 37500 bytes
[Two strings] Quarternary is: 604x smaller
[Two strings] Memory used by the 1000000 element map: 309011104 bytes
[Two strings] Memory used by the 1000000 element quarternary: 375000 bytes
[Two strings] Quarternary is: 824x smaller
[Bytes] Memory used by the 10 element map: 8568 bytes
[Bytes] Memory used by the 10 element quarternary: 8 bytes
[Bytes] Quarternary is: 1071x smaller
[Bytes] Memory used by the 100 element map: 46240 bytes
[Bytes] Memory used by the 100 element quarternary: 75 bytes
[Bytes] Quarternary is: 616x smaller
[Bytes] Memory used by the 1000 element map: 665920 bytes
[Bytes] Memory used by the 1000 element quarternary: 750 bytes
[Bytes] Quarternary is: 887x smaller
[Bytes] Memory used by the 10000 element map: 3805464 bytes
[Bytes] Memory used by the 10000 element quarternary: 7500 bytes
[Bytes] Quarternary is: 507x smaller
[Bytes] Memory used by the 100000 element map: 32998808 bytes
[Bytes] Memory used by the 100000 element quarternary: 75000 bytes
[Bytes] Quarternary is: 439x smaller
[Bytes] Memory used by the 1000000 element map: 475530664 bytes
[Bytes] Memory used by the 1000000 element quarternary: 750000 bytes
[Bytes] Quarternary is: 634x smaller
PASS
ok  	github.com/neurlang/quaternary	24.628s
```

## License

MIT
