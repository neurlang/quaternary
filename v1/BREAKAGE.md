# Value-type narrowing breakage in `v1`

This document records what happens when the value type `V` in the v1 `Make`/`New`
generic API is narrowed from `uint64` to `uint8`/`uint16`/`uint32`, using the
`TestInsertBigBytes` test in `v1/api_test.go` as the probe.

## Setup

```go
type brokenValue = uint64  // the type under test (V in Make[K, V])
func TestInsertBigBytes(t *testing.T) {
    for bc := 3; bc < 63; bc++ {
        m := map[[32]byte]brokenValue{
            [32]byte{1}: 1, [32]byte{2}: 2, [32]byte{3}: 3,
            [32]byte{4}: 4, [32]byte{5}: 5,
        }
        b := Make(m, byte(bc))          // bitLimit = bc (3..62)
        for k, v := range m {
            val := brokenValue(GetNum(b, uint64(bc), k))
            if v != val { t.Fatalf(...) }
        }
    }
}
```

The keys are `[32]byte`, values are `1..5`, and `bitLimit` sweeps `3..62`.
With `V = uint64` this test passes (verified: `ok`).

## Summary of breakage

| Value type `V` | Failure mode | Failing site | Threshold |
|---|---|---|---|
| `uint64` | none — passes | — | — |
| `uint32` | `panic: slice bounds out of range [255:4]` | `v1/api.go:131` | `bitLimit` > 32 |
| `uint16` | `panic: slice bounds out of range` | `v1/api.go:135` | `bitLimit` > 16 |
| `uint8` | `panic: inserting value exceeding bit limit when bit limit set` | `v1/create.go:97` | `bitLimit` > 8 |

## Root cause 1: out-of-range slice in `api.go` (uint16 / uint32)

In `v1.New`, numeric value cases truncate the big-endian encoding to the bit limit:

```go
// api.go:128-135
case uint32:
    var b [4]byte
    binary.BigEndian.PutUint32(b[:], uint32(val))
    *kvPairValue(&kv) = b[4-((bitLimit+7)/8):]
case uint16:
    var b [2]byte
    binary.BigEndian.PutUint16(b[:], uint16(val))
    *kvPairValue(&kv) = b[2-((bitLimit+7)/8):]
```

The slice start `N - ((bitLimit+7)/8)` (N = 4 or 2) goes **negative** once
`(bitLimit+7)/8` exceeds the value width `N`. Since `bitLimit` sweeps up to 62,
`(bitLimit+7)/8` reaches 8, so `b[4-8:]` / `b[2-8:]` panics with
`slice bounds out of range`.

This means **`bitLimit` must not exceed the width of the value type** for
`uint16`/`uint32`. The test's `bitLimit` range is what triggers it; the values
themselves (`1..5`) are well within the type width.

## Root cause 2: untruncated `uint8` value vs. `create.go` length check

The `uint8` case stores the value as a full byte with **no bit-limit truncation**:

```go
// api.go:136-138
case uint8:
    var b = []byte{byte(val)}
    *kvPairValue(&kv) = b[:]
```

Then `v1.create` enforces that stored values match the bit-limit byte length:

```go
// create.go:96-98
if bitLimit != 0 && (len(*kvPairValue(&kv))) != int(bitLimit+7)/8 {
    panic("inserting value exceeding bit limit when bit limit set")
}
```

A `uint8` value is always 1 byte, but `(bitLimit+7)/8` is 2 or more once
`bitLimit > 8`. So the length check panics for `bitLimit` in `9..62`. Unlike the
other types, the `uint8` path can't represent a bit limit wider than 8 bits at
all.

## Implications

- The v1 API advertises `V = uint8 | uint16 | uint32 | uint64` as supported, but
  the implementation **only works over the full `bitLimit` sweep when `V = uint64`**.
- `bitLimit` semantics differ per value type:
  - `uint8` is effectively limited to `bitLimit <= 8`.
  - `uint16` is effectively limited to `bitLimit <= 16`.
  - `uint32` is effectively limited to `bitLimit <= 32`.
  - `uint64` accepts the full `3..62` sweep (and beyond up to 64).
- The breakage is **silent for small `bitLimit` values** — the test only catches it
  once the sweep crosses the type width, which makes it easy to miss with fixed,
  in-range bit limits in real usage.

## Fix directions (not applied)

- Guard the slice bounds in the `uint16`/`uint32` cases (e.g. clamp
  `(bitLimit+7)/8` to `N`, or reject `bitLimit` exceeding the type width).
- For `uint8`, either truncate/validate against `bitLimit` like the wider types,
  or document/limit the supported `bitLimit` range.
