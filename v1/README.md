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
5: true 55: false size: 2 B
```

---

## API

```go
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

## Supported key/value types

* **Keys**: any `comparable` (int, string, fixed byte arrays, etc.)
* **Values**: `bool`, `uint8/16/32/64`, `string`, `[]byte`

> ⚠️ If you look up a key that wasn’t inserted, you’ll get a garbage value.
> This is by design for speed and size. Validate your keys externally if needed.

---

## Memory efficiency

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
