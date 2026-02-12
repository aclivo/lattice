// Package lattice provides a compact, Z-order encoded multidimensional
// address type for use as Go map keys with zero allocations.
package lattice

import "fmt"

// Addr is a compact, Z-order encoded multidimensional address.
// It supports up to 12 dimensions with values ranging from 0 to 1,048,575.
//
// Bit layout:
//   - bits 0-3:   number of dimensions (max 15)
//   - bits 4-243: Z-order interleaved coordinates (20 bits each)
type Addr [4]uint64

const (
	// BitsPerCoord is the number of bits used per coordinate (20 bits).
	BitsPerCoord = 20

	// MaxDimensions is the maximum number of dimensions supported.
	MaxDimensions = 12

	// MaxCoordValue is the maximum value a coordinate can hold (2^20 - 1 = 1,048,575).
	MaxCoordValue = (1 << BitsPerCoord) - 1
)

// New creates a new Addr from the given coordinates using Z-order encoding.
// Panics if more than MaxDimensions coordinates are provided,
// or if any coordinate is out of range [0, MaxCoordValue].
func New(coords ...int) Addr {
	if len(coords) > MaxDimensions {
		panic(fmt.Sprintf("lattice: max %d dimensions supported", MaxDimensions))
	}

	for i, v := range coords {
		if v < 0 || v > MaxCoordValue {
			panic(fmt.Sprintf("lattice: coord[%d]=%d out of range [0,%d]", i, v, MaxCoordValue))
		}
	}

	var addr Addr
	addr[0] = uint64(len(coords))

	numDims := len(coords)
	for bitPos := 0; bitPos < BitsPerCoord; bitPos++ {
		for dimIdx := 0; dimIdx < numDims; dimIdx++ {
			bit := (coords[dimIdx] >> bitPos) & 1

			encodedBitPos := 4 + bitPos*numDims + dimIdx

			if bit == 1 {
				arrayIdx := encodedBitPos / 64
				bitInWord := encodedBitPos % 64
				addr[arrayIdx] |= 1 << bitInWord
			}
		}
	}

	return addr
}

// Dims returns the number of dimensions in this address.
func (a Addr) Dims() int {
	return int(a[0] & 0xF)
}

// Coords decodes and returns coordinates as a stack-allocated array.
// Use dims to know how many elements are valid.
// Zero allocations.
func (a Addr) Coords() ([MaxDimensions]int, int) {
	var coords [MaxDimensions]int
	dims := a.Dims()

	for bitPos := 0; bitPos < BitsPerCoord; bitPos++ {
		for dimIdx := 0; dimIdx < dims; dimIdx++ {
			encodedBitPos := 4 + bitPos*dims + dimIdx
			arrayIdx := encodedBitPos / 64
			bitInWord := encodedBitPos % 64

			if (a[arrayIdx]>>bitInWord)&1 == 1 {
				coords[dimIdx] |= 1 << bitPos
			}
		}
	}

	return coords, dims
}

// CoordsSlice decodes coordinates into the provided buffer.
// buf must be at least Dims() in length.
// Returns the filled slice with no allocations.
func (a Addr) CoordsSlice(buf []int) []int {
	coords, dims := a.Coords()
	if len(buf) < dims {
		panic(fmt.Sprintf("lattice: buf too small: need %d, got %d", dims, len(buf)))
	}
	buf = buf[:dims]
	for i := 0; i < dims; i++ {
		buf[i] = coords[i]
	}
	return buf
}

// Append returns a new Addr with extra coordinates added
// e.g. Addr{1,2}.Append(3) → Addr{1,2,3}
func (a Addr) Append(coords ...int) Addr {
	ac, dims := a.Coords()
	next := make([]int, dims+len(coords))
	for i := 0; i < dims; i++ {
		next[i] = ac[i]
	}
	copy(next[dims:], coords)
	return New(next...)
}

// At returns the coordinate value at a specific dimension
func (a Addr) At(dimIdx int) int {
	ac, dims := a.Coords()
	if dimIdx < 0 || dimIdx >= dims {
		panic(fmt.Sprintf("lattice: dimension index %d out of range [0:%d]", dimIdx, dims))
	}
	return ac[dimIdx]
}

// Contains checks if this address shares a prefix with another
// e.g. Addr{1,2} contains Addr{1,2,3}
func (a Addr) Contains(b Addr) bool {
	aDims := a.Dims()
	bDims := b.Dims()
	if aDims > bDims {
		return false
	}
	ac, _ := a.Coords()
	bc, _ := b.Coords()
	for i := 0; i < aDims; i++ {
		if ac[i] != bc[i] {
			return false
		}
	}
	return true
}

// Equal checks if two addresses are identical
func (a Addr) Equal(b Addr) bool {
	return a == b
}

// InRange checks if this address falls within the given coordinate ranges
// ranges: [2]int{min, max} per dimension, use {-1,-1} for "any"
func (a Addr) InRange(ranges ...[2]int) bool {
	ac, dims := a.Coords()
	for i, r := range ranges {
		if i >= dims {
			break
		}
		if r[0] != -1 && ac[i] < r[0] {
			return false
		}
		if r[1] != -1 && ac[i] > r[1] {
			return false
		}
	}
	return true
}

// IsZero checks if all coordinates are zero
func (a Addr) IsZero() bool {
	coords, dims := a.Coords()
	for i := 0; i < dims; i++ {
		if coords[i] != 0 {
			return false
		}
	}
	return true
}

// Slice returns a new Addr with a subset of dimensions
// e.g. Addr{1,2,3}.Slice(0,2) → Addr{1,2}
func (a Addr) Slice(from, to int) Addr {
	ac, dims := a.Coords()
	if from < 0 || to > dims || from > to {
		panic(fmt.Sprintf("lattice: slice [%d:%d] out of range [0:%d]", from, to, dims))
	}
	coords := make([]int, to-from)
	for i := range coords {
		coords[i] = ac[from+i]
	}
	return New(coords...)
}

// With returns a new Addr with one coordinate replaced
// e.g. Addr{1,2,3}.With(1, 99) → Addr{1,99,3}
func (a Addr) With(dimIdx int, value int) Addr {
	ac, dims := a.Coords()
	if dimIdx < 0 || dimIdx >= dims {
		panic(fmt.Sprintf("lattice: dimension index %d out of range [0:%d]", dimIdx, dims))
	}
	coords := make([]int, dims)
	for i := 0; i < dims; i++ {
		coords[i] = ac[i]
	}
	coords[dimIdx] = value
	return New(coords...)
}

// String returns a human-readable representation of the address.
func (a Addr) String() string {
	var buf [MaxDimensions]int
	return fmt.Sprintf("Addr%v", a.CoordsSlice(buf[:]))
}
