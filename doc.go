// Package lattice provides a compact, Z-order encoded multidimensional
// address type for use as Go map keys with zero allocations.
//
// # Overview
//
// The core type [Addr] encodes up to 12 integer coordinates into a single
// 32-byte value using Z-order (Morton) encoding. Because Addr is a fixed-size
// array ([4]uint64), it is directly comparable and can be used as a map key
// without hashing or string conversion.
//
//	cells := map[lattice.Addr]float64{}
//	cells[lattice.New(1, 2, 3)] = 42.0
//
// # Encoding
//
// Coordinates are encoded using Z-order (Morton code), which interleaves the
// bits of each dimension into a single value. This preserves spatial locality:
// nearby coordinates in N-dimensional space produce nearby encoded values,
// which improves cache performance for range queries.
//
// The bit layout of an Addr is:
//
//	bits  0–3:   number of dimensions (max 15)
//	bits  4–243: Z-order interleaved coordinates (20 bits each)
//
// # Constraints
//
// Each coordinate must be in the range [0, MaxCoordValue] (0 to 1,048,575).
// A maximum of [MaxDimensions] (12) dimensions are supported.
// [New] panics if either constraint is violated.
//
// # Zero Allocations
//
// The hot path is entirely allocation-free:
//
//	lattice.New(1, 2, 3)       // 0 allocs - encodes to Addr
//	addr.Coords()              // 0 allocs - decodes to stack-allocated array
//	addr.CoordsSlice(buf)      // 0 allocs - decodes into caller-provided buffer
//	addr.Dims()                // 0 allocs - reads dimension count
//	addr.At(i)                 // 0 allocs - reads one coordinate
//	addr.Equal(b)              // 0 allocs - direct array comparison
//	addr.Contains(b)           // 0 allocs
//	addr.InRange(ranges...)    // 0 allocs
//	addr.IsZero()              // 0 allocs
//	addr.String()              // 0 allocs - uses stack buffer internally
//
// Methods that build a new coordinate slice ([Append], [Addr.Slice], [With])
// perform one allocation each, but the returned [Addr] is always 32 bytes and
// allocation-free to use as a map key.
//
// # Decoding
//
// Two decode methods are provided to suit different needs:
//
//	// Stack-allocated array — zero allocations, use in hot paths
//	coords, dims := addr.Coords()
//	for i := 0; i < dims; i++ {
//	    fmt.Println(coords[i])
//	}
//
//	// Caller-provided buffer — zero allocations, returns a slice
//	buf := make([]int, lattice.MaxDimensions) // allocate once, reuse forever
//	coords := addr.CoordsSlice(buf)
//
// # Navigation and Transformation
//
// Addr provides methods to navigate and transform addresses immutably.
// Each method returns a new Addr and never modifies the receiver:
//
//	addr := lattice.New(10, 20, 30)
//
//	addr.At(1)                          // 20
//	addr.With(1, 99)                    // Addr[10 99 30]
//	addr.Append(40)                     // Addr[10 20 30 40]
//	addr.Slice(0, 2)                    // Addr[10 20]
//
// # Querying
//
//	addr.IsZero()                       // false
//	addr.Equal(lattice.New(10, 20, 30)) // true
//
//	// Prefix containment: Addr{10,20} contains Addr{10,20,30}
//	lattice.New(10, 20).Contains(addr)  // true
//
//	// Range query: use -1 for "any" on a dimension
//	addr.InRange(
//	    [2]int{5, 15},    // dim 0: 5–15   ✓ (10)
//	    [2]int{-1, -1},   // dim 1: any    ✓
//	    [2]int{25, 35},   // dim 2: 25–35  ✓ (30)
//	)                                   // true
//
// # Capacity Planning
//
// Based on real measurements for map[Addr]float64:
//
//	lattice.Addr  = 32 bytes  (key)
//	float64       =  8 bytes  (value)
//	map overhead  = ~60 bytes (Go bucket structure + power-of-2 rounding)
//	─────────────────────────────────────────────────────────────────────
//	Actual cost   = ~100 bytes per cell
//
// Rule of thumb: approximately 10 million cells per gigabyte of memory.
//
// # References
//
// Z-order / Morton encoding:
//   - https://en.wikipedia.org/wiki/Z-order_curve
//   - https://en.wikipedia.org/wiki/Morton_code
//   - https://www.forceflow.be/2013/10/07/morton-encoding-decoding-through-bit-interleaving-implementations/
//   - https://fgiesen.wordpress.com/2009/12/13/decoding-morton-codes/
//
// Space-filling curves:
//   - https://en.wikipedia.org/wiki/Space-filling_curve
//   - https://en.wikipedia.org/wiki/Hilbert_curve
//
// OLAP and multidimensional data:
//   - https://en.wikipedia.org/wiki/OLAP_cube
//   - https://en.wikipedia.org/wiki/Sparse_matrix
//   - https://en.wikipedia.org/wiki/Array_DBMS
package lattice
