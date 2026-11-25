# quaternary v1

**Quaternary Filter** is a compact, immutable, `[]byte`-backed replacement for read-only Go maps.

It turns many `map[K]V` into a tiny, serialized `[]byte` structure with predictable lookups and dramatic memory savings.

---

## Why

* Up to **1000× smaller** than native Go maps
* Still allows **fast O(1)** lookups by key
* Trivial to **serialize** and **transmit** (`[]byte`)
* Works with numbers, booleans, short strings, and byte arrays

Use this when you want the benefits of a map but:

* You **don’t need to insert** more keys after construction
* You **don’t need to iterate** over keys
* You **only need lookups**

---

## Example

```go
package main

import (
    "fmt"
    quaternary "github.com/neurlang/quaternary/v1"
)

func main() {
    filter := quaternary.Make(map[int]bool{
        5:  true,
        55: false,
    }, 0) // bitLimit=0 → auto

    if quaternary.GetBool(filter, 5) != true {
        panic("5 not true")
    }
    if quaternary.GetBool(filter, 55) != false {
        panic("55 not false")
    }

    fmt.Println("5:", quaternary.GetBool(filter, 5),
                "55:", quaternary.GetBool(filter, 55),
                "size:", len(filter), "B")
}
```

Output:

```
5: true 55: false size: 3 B
```

---

## API

```go
// New generates the filter based on map m
func New[K comparable, V string | []byte | bool | uint64 | uint32 | uint16 | uint8](m map[K]V, bitLimit, bloomFuncs byte) []byte

// Make generates the filter based on map m
func Make[K comparable, V string | []byte | bool | uint64 | uint32 | uint16 | uint8](m map[K]V, bitLimit byte) []byte

// Get retrieves raw value bytes for a given key
func Get[K comparable](f []byte, valBitSize uint64, key K) []byte

// GetBool retrieves a bool value
func GetBool[K comparable](f []byte, key K) bool

// GetNum retrieves a numeric value (uintX) of a given bit size
func GetNum[K comparable](f []byte, valBitSize uint64, key K) uint64
```

---

## BitLimit

- For fine grained control of bit size of stored values
  - Such as: storing integers with only few least significant bits
- Use 0 for byte-aligned automatic control
  - `uint8/16/32/64` values: the full width of the type (8/16/32/64 bits)
  - `string` / `[]byte` values: every value keeps its own length (it may vary between values)
  - `bool` values always use 1 bit
- `bitLimit` may not exceed the width of fixed-width numeric values (8/16/32/64 bits for `uint8/16/32/64`)
- Memory is proportional to the number of stored bits: `bitLimit` x number of items (or the sum of the value lengths when `bitLimit` is 0)
- Values that do not fit in `bitLimit` are **silently truncated** (numbers keep their low bits; `string`/`[]byte` keep their trailing bytes)
- With `bitLimit` above 0, `Get`/`GetNum` can read any prefix (the low `valBitSize` bits) of a value
- With `bitLimit` 0 the filter does not know a value's length: `valBitSize` must be exactly `8*len(value)` of that key, otherwise you silently get wrong bytes. Keep the length yourself: in the least significant byte of a fixed-width value, under a derived key, or in a second filter (see the package documentation)

---

## BloomFuncs

- Setting to higher values than 0 enables returning dummy value (e.g. 0 / false) if value wasn't inserted
- Probabilistic only (may fail - silently return garbage value with small probability)
- Every function makes the filter larger; at most 56 functions are used (more behaves like 56)

---

## Supported key/value types

* **Keys**: any `comparable` (int, string, fixed byte arrays, etc.). A filter answers for the key's encoding, not for Go's key equality:
  * two keys with the same encoding (structs differing only in unexported fields, invalid UTF-8 strings in composite keys, `int(1)` vs `int64(1)` in a `map[any]`) make `New`/`Make` panic, even if their values are equal
  * pointer keys are encoded by the value they point to, not by address
  * `+0.0` and `-0.0` are different keys; several `NaN` keys panic
* **Values**: `bool`, `uint8/16/32/64`, `string`, `[]byte`

> ⚠️ If you look up a key that wasn’t inserted, you’ll get a garbage value.
> This is by design for speed and size. Validate your keys externally if needed.

---

## Memory efficiency: Ω(⅀values_bits_size)

From benchmarks (`go test --bench=ReadAll`):

```
map[int]bool       → 62x – 113x smaller
map[string]bool    → 69x – 439x smaller
map[[2]string]bool → 604x – 1094x smaller
map[[64]byte]bool  → 438x – 887x smaller
```

Example:

* `map[string]bool` with 1,000,000 entries → **165 MB**
* `quaternary` filter with same entries → **0.75 MB**
* That’s a **220× reduction**.

---

## Use cases

* Embedding **static lookup tables** in binaries
* Compact **configuration tables**
* **On-disk file formats** (direct `[]byte` storage)
* **Networking protocols** (transmit prebuilt filter)

---

## Status

* **Version**: v1.0 (stable API)
* **Format**: still subject to change in `v1.x.0` if major optimizations are found.
* **Quality**: production-ready, with alpha-quality experimental features.

---

## License

MIT
