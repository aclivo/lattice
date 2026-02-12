# lattice

A compact, Z-order encoded multidimensional address for Go.

## Install
```bash
go get github.com/aclivo/lattice
```

## Usage
```go
import "github.com/aclivo/lattice"

// Create an address
addr := lattice.New(100, 200, 300)

// Decode back to coordinates - zero allocations
coords, dims := addr.Coords()   // [MaxDimensions]int, 3
for i := 0; i < dims; i++ {
    fmt.Println(coords[i])      // 100, 200, 300
}

// Decode into a reusable buffer - zero allocations
buf := make([]int, lattice.MaxDimensions) // allocate once
coords := addr.CoordsSlice(buf)           // []int{100, 200, 300}

// Number of dimensions
dims := addr.Dims() // 3

// Use as a map key
cells := map[lattice.Addr]float64{}
cells[lattice.New(1, 2, 3)] = 42.0

// Navigate and transform
addr.At(1)                          // 200
addr.With(1, 999)                   // Addr[100 999 300]
addr.Append(400)                    // Addr[100 200 300 400]
addr.Slice(0, 2)                    // Addr[100 200]

// Query
addr.Contains(lattice.New(100, 200)) // false (reversed - shorter doesn't contain longer)
lattice.New(100, 200).Contains(addr) // true  (prefix match)
addr.InRange([2]int{50,150}, [2]int{150,250}, [2]int{250,350}) // true
addr.IsZero()                        // false
addr.Equal(lattice.New(100, 200, 300)) // true
```

## API
```go
// New creates a new Addr from the given coordinates.
// Panics if len(coords) > MaxDimensions or any coord is out of [0, MaxCoordValue].
func New(coords ...int) Addr

// Dims returns the number of dimensions.
func (a Addr) Dims() int

// Coords decodes coordinates into a stack-allocated array.
// Returns the array and the number of valid dimensions.
// Zero allocations.
func (a Addr) Coords() ([MaxDimensions]int, int)

// CoordsSlice decodes coordinates into the provided buffer.
// buf must be at least Dims() in length.
// Zero allocations.
func (a Addr) CoordsSlice(buf []int) []int

// Append returns a new Addr with extra coordinates added.
// e.g. Addr{1,2}.Append(3) → Addr{1,2,3}
func (a Addr) Append(coords ...int) Addr

// At returns the coordinate value at a specific dimension.
func (a Addr) At(dimIdx int) int

// Contains checks if this address shares a prefix with another.
// e.g. Addr{1,2}.Contains(Addr{1,2,3}) → true
func (a Addr) Contains(b Addr) bool

// Equal checks if two addresses are identical.
func (a Addr) Equal(b Addr) bool

// InRange checks if this address falls within the given coordinate ranges.
// Use {-1,-1} for "any" on a dimension.
func (a Addr) InRange(ranges ...[2]int) bool

// IsZero checks if all coordinates are zero.
func (a Addr) IsZero() bool

// Slice returns a new Addr with a subset of dimensions.
// e.g. Addr{1,2,3}.Slice(0,2) → Addr{1,2}
func (a Addr) Slice(from, to int) Addr

// With returns a new Addr with one coordinate replaced.
// e.g. Addr{1,2,3}.With(1, 99) → Addr{1,99,3}
func (a Addr) With(dimIdx int, value int) Addr

// String returns a human-readable representation e.g. "Addr[1 2 3]".
func (a Addr) String() string
```

## Specs

| Property                | Value                   |
|-------------------------|-------------------------|
| Memory per address      | 32 bytes                |
| Max dimensions          | 12                      |
| Max value per dimension | 1,048,575 (0 to 2²⁰-1) |
| Encoding                | Z-order (Morton code)   |
| Map key compatible      | ✅                      |

## Why Z-order?

Z-order encoding interleaves bits from each dimension, preserving
spatial locality. Nearby coordinates produce nearby encoded values,
which improves cache performance for range queries.

## Performance

### Addressing Strategies Comparison

| Strategy | Encode | Decode | Map Lookup | Memory per Address | Collision Risk |
|----------|--------|--------|------------|--------------------|----------------|
| **`lattice.Addr`** | O(d·b) | O(d·b) | O(1) | 32 bytes | None |
| String key `"1,2,3"` | O(d) | O(d) | O(d) | 24 + d·chars bytes | None |
| `uint64` hash | O(d) | ❌ | O(1) | 8 bytes | Yes |
| `[]int` slice | O(1) | O(1) | ❌ not comparable | 24 + d·8 bytes | None |
| Nested maps `map[int]map[int]...` | O(1) | O(1) | O(d) | d · map overhead | None |

Where:
- **d** = number of dimensions
- **b** = bits per coordinate (20 in this package)
- **O(d·b)** is a small constant in practice (max 12 × 20 = 240 iterations)

### Why O(d·b) Is Fast In Practice

Although encode/decode is O(d·b), the constant is tiny and **bounded**:
```
Worst case: 12 dimensions × 20 bits = 240 iterations
Best case:   1 dimension  × 20 bits =  20 iterations
```

This means encode/decode is effectively **O(1)** from the caller's perspective -
the upper bound never changes regardless of data size.

### Map Lookup Detail

| Strategy | Hash Cost | Comparison Cost | Total |
|----------|-----------|-----------------|-------|
| **`lattice.Addr`** | O(1) · 4 words | O(1) · 32 bytes | **O(1)** |
| String key | O(d) chars | O(d) chars | **O(d)** |
| `uint64` hash | O(1) | O(1) | **O(1)** but collisions |
| Nested maps | O(1) per level | O(1) per level | **O(d)** |

### Zero Allocations
```
lattice.New(1, 2, 3)         →  0 allocs/op
addr.Coords()                →  0 allocs/op  (stack-allocated array)
addr.CoordsSlice(buf)        →  0 allocs/op  (reusable buffer)
addr.Dims()                  →  0 allocs/op
addr.At(i)                   →  0 allocs/op
addr.Contains(b)             →  0 allocs/op
addr.Equal(b)                →  0 allocs/op
addr.InRange(ranges...)      →  0 allocs/op
addr.IsZero()                →  0 allocs/op
addr.String()                →  0 allocs/op  (stack buffer internally)
addr.Append(coords...)       →  1 alloc/op   (new coord slice)
addr.Slice(from, to)         →  1 alloc/op   (new coord slice)
addr.With(dimIdx, value)     →  1 alloc/op   (new coord slice)
map[Addr]float64 lookup      →  0 allocs/op
```

> **Note**: `Append`, `Slice` and `With` allocate because they build a new
> coordinate slice internally. The returned `Addr` is still 32 bytes and
> allocation-free to use as a map key.

## Benchmarks
```
$ go test -bench=. -benchmem
goos: linux
goarch: amd64
pkg: github.com/aclivo/lattice
cpu: i5 @ 1.60GHz

BenchmarkNew_1D-8               18826022        58.36 ns/op      0 B/op   0 allocs/op
BenchmarkNew_3D-8                8817471       135.5  ns/op      0 B/op   0 allocs/op
BenchmarkNew_6D-8                4666989       256.3  ns/op      0 B/op   0 allocs/op
BenchmarkNew_12D-8               3131702       381.2  ns/op      0 B/op   0 allocs/op
BenchmarkNew_12D_Large-8         2806237       440.5  ns/op      0 B/op   0 allocs/op

BenchmarkCoords_3D-8             8990538       123.6  ns/op      0 B/op   0 allocs/op
BenchmarkCoords_12D-8            2698065       456.7  ns/op      0 B/op   0 allocs/op
BenchmarkCoordsSlice_3D-8        8883076       139.4  ns/op      0 B/op   0 allocs/op
BenchmarkCoordsSlice_12D-8       2633020       471.7  ns/op      0 B/op   0 allocs/op
BenchmarkDims-8               1000000000         0.28 ns/op      0 B/op   0 allocs/op

BenchmarkRoundTrip_3D-8          4775402       261.8  ns/op      0 B/op   0 allocs/op
BenchmarkRoundTrip_12D-8         1462904       812.0  ns/op      0 B/op   0 allocs/op

BenchmarkMapInsert_3D-8          8052862       183.9  ns/op      0 B/op   0 allocs/op
BenchmarkMapLookup_3D-8          5807565       197.8  ns/op      0 B/op   0 allocs/op
BenchmarkMapLookup_12D-8         2420226       506.7  ns/op      0 B/op   0 allocs/op
BenchmarkMapIteration_100k-8        1224    963144    ns/op      0 B/op   0 allocs/op

BenchmarkAppend_One-8            4111620       293.5  ns/op      0 B/op   0 allocs/op
BenchmarkAt-8                    5847494       204.8  ns/op      0 B/op   0 allocs/op
BenchmarkContains-8              4053244       300.0  ns/op      0 B/op   0 allocs/op
BenchmarkEqual-8              1000000000         1.14 ns/op      0 B/op   0 allocs/op
BenchmarkInRange_3D-8            8667261       134.8  ns/op      0 B/op   0 allocs/op
BenchmarkIsZero-8                5705386       204.6  ns/op      0 B/op   0 allocs/op
BenchmarkSlice-8                 3112213       386.7  ns/op      0 B/op   0 allocs/op
BenchmarkWith-8                  2837587       434.1  ns/op     48 B/op   1 allocs/op
```

## References

### Z-order / Morton Encoding
- [Z-order curve - Wikipedia](https://en.wikipedia.org/wiki/Z-order_curve)
- [Morton code - Wikipedia](https://en.wikipedia.org/wiki/Morton_code)
- [Morton Encoding/Decoding through Bit Interleaving](https://www.forceflow.be/2013/10/07/morton-encoding-decoding-through-bit-interleaving-implementations/)
- [Fast Morton Codes - Fabian Giesen](https://fgiesen.wordpress.com/2009/12/13/decoding-morton-codes/)

### Space-Filling Curves (broader context)
- [Space-filling curve - Wikipedia](https://en.wikipedia.org/wiki/Space-filling_curve)
- [Hilbert curve - Wikipedia](https://en.wikipedia.org/wiki/Hilbert_curve)

### OLAP and Multidimensional Data
- [OLAP cube - Wikipedia](https://en.wikipedia.org/wiki/OLAP_cube)
- [Sparse matrix - Wikipedia](https://en.wikipedia.org/wiki/Sparse_matrix)
- [Array DBMS - Wikipedia](https://en.wikipedia.org/wiki/Array_DBMS)