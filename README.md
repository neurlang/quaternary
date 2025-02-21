# quaternary
Quaternary Filter is a 100x smaller `map[int]bool`

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

## Memory efficiency/performance

```
$ go test --bench=ReadAll
Memory used by the 10 element map: 176 bytes
Memory used by the 10 element quarternary: 4 bytes
Quarternary is: 44x smaller
Memory used by the 100 element map: 3432 bytes
Memory used by the 100 element quarternary: 38 bytes
Quarternary is: 90x smaller
Memory used by the 1000 element map: 52888 bytes
Memory used by the 1000 element quarternary: 375 bytes
Quarternary is: 141x smaller
Memory used by the 10000 element map: 426648 bytes
Memory used by the 10000 element quarternary: 3750 bytes
Quarternary is: 113x smaller
Memory used by the 100000 element map: 1761288 bytes
Memory used by the 100000 element quarternary: 37500 bytes
Quarternary is: 46x smaller
Memory used by the 1000000 element map: 38995512 bytes
Memory used by the 1000000 element quarternary: 375000 bytes
Quarternary is: 103x smaller
goos: linux
goarch: amd64
pkg: github.com/neurlang/quaternary
cpu: AMD Ryzen 9 9950X 16-Core Processor            
BenchmarkReadAllQuarternary/Case1-32         	1000000000	         0.0000002 ns/op
BenchmarkReadAllQuarternary/Case2-32         	1000000000	         0.0000027 ns/op
BenchmarkReadAllQuarternary/Case3-32         	1000000000	         0.0000602 ns/op
BenchmarkReadAllQuarternary/Case4-32         	1000000000	         0.0001140 ns/op
BenchmarkReadAllQuarternary/Case5-32         	1000000000	         0.003177 ns/op
BenchmarkReadAllQuarternary/Case6-32         	1000000000	         0.02021 ns/op
BenchmarkReadAllMap/Case1-32                 	1000000000	         0.0000007 ns/op
BenchmarkReadAllMap/Case2-32                 	1000000000	         0.0000041 ns/op
BenchmarkReadAllMap/Case3-32                 	1000000000	         0.0000330 ns/op
BenchmarkReadAllMap/Case4-32                 	1000000000	         0.0003355 ns/op
BenchmarkReadAllMap/Case5-32                 	1000000000	         0.004093 ns/op
BenchmarkReadAllMap/Case6-32                 	1000000000	         0.1104 ns/op
PASS
ok  	github.com/neurlang/quaternary	3.463s
```

## License

MIT
