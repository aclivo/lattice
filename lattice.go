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

	// dimsBits is the number of bits used to store the number of dimensions.
	dimsBits = 4

	// dimsMask selects the dimension count bits from the header word.
	dimsMask = 0xF

	// bitsPerWord is the number of bits in a uint64 word.
	bitsPerWord = 64
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
	for bitPos := range BitsPerCoord {
		for dimIdx := range numDims {
			bit := (coords[dimIdx] >> bitPos) & 1

			encodedBitPos := dimsBits + bitPos*numDims + dimIdx

			if bit == 1 {
				arrayIdx := encodedBitPos / bitsPerWord
				bitInWord := encodedBitPos % bitsPerWord
				addr[arrayIdx] |= 1 << bitInWord
			}
		}
	}

	return addr
}

// Dims returns the number of dimensions in this address.
func (a Addr) Dims() int {
	return int(a[0] & dimsMask) //nolint:gosec // dimsMask ensures value fits in [0,15]
}

// Coords decodes and returns coordinates as a stack-allocated array.
// Use dims to know how many elements are valid.
// Zero allocations.
func (a Addr) Coords() (Buffer, int) {
	var coords Buffer

	dims := a.Dims()

	for bitPos := range BitsPerCoord {
		for dimIdx := range dims {
			encodedBitPos := dimsBits + bitPos*dims + dimIdx
			arrayIdx := encodedBitPos / bitsPerWord
			bitInWord := encodedBitPos % bitsPerWord

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

	for i := range dims {
		buf[i] = coords[i] //nolint:gosec // coords is Buffer, dims <= MaxDimensions
	}

	return buf
}

// Append returns a new Addr with extra coordinates added
// e.g. Addr{1,2}.Append(3) → Addr{1,2,3}.
func (a Addr) Append(coords ...int) Addr {
	ac, dims := a.Coords()
	next := make([]int, dims+len(coords))

	for i := range dims {
		next[i] = ac[i]
	}

	copy(next[dims:], coords)

	return New(next...)
}

// At returns the coordinate value at a specific dimension
// e.g. Addr{1,2,3}.At(1) → 2.
func (a Addr) At(dimIdx int) int {
	ac, dims := a.Coords()
	if dimIdx < 0 || dimIdx >= dims {
		panic(fmt.Sprintf("lattice: dimension index %d out of range [0:%d]", dimIdx, dims))
	}

	return ac[dimIdx]
}

// Contains checks if this address shares a prefix with another
// e.g. Addr{1,2} contains Addr{1,2,3}.
func (a Addr) Contains(bAddr Addr) bool {
	aDims := a.Dims()
	bDims := bAddr.Dims()

	if aDims > bDims {
		return false
	}

	aCoords, _ := a.Coords()
	bCoords, _ := bAddr.Coords()

	for i := range aDims {
		if aCoords[i] != bCoords[i] { //nolint:gosec // i < aDims <= MaxDimensions == len(aCoords)
			return false
		}
	}

	return true
}

// InRange checks if this address falls within the given coordinate ranges.
// ranges: each element is [min, max] for the corresponding dimension.
// A value of -1 for min or max means no bound in that direction.
// e.g. Addr{10,20,30}.InRange({5,15}, {15,25}, {25,35}) → true.
func (a Addr) InRange(ranges ...AddrRange) bool {
	aCoords, dims := a.Coords()

	for index, _range := range ranges {
		if index >= dims {
			break
		}

		coord := aCoords[index]

		if _range[0] != -1 && coord < _range[0] { //nolint:gosec // _range is [2]int, indexes 0 and 1 always valid
			return false
		}

		if _range[1] != -1 && coord > _range[1] { //nolint:gosec // _range is [2]int, indexes 0 and 1 always valid
			return false
		}
	}

	return true
}

// IsZero checks if all coordinates are zero.
func (a Addr) IsZero() bool {
	coords, dims := a.Coords()
	for i := range dims {
		if coords[i] != 0 { //nolint:gosec // i < dims <= MaxDimensions == len(coords)
			return false
		}
	}

	return true
}

// Slice returns a new Addr with a subset of dimensions
// e.g. Addr{1,2,3}.Slice(0,2) → Addr{1,2}.
func (a Addr) Slice(fromAddr, toAddr int) Addr {
	aCoords, dims := a.Coords()
	if fromAddr < 0 || toAddr > dims || fromAddr > toAddr {
		panic(fmt.Sprintf("lattice: slice [%d:%d] out of range [0:%d]", fromAddr, toAddr, dims))
	}

	coords := make([]int, toAddr-fromAddr)
	for i := range coords {
		coords[i] = aCoords[fromAddr+i]
	}

	return New(coords...)
}

// With returns a new Addr with one coordinate replaced
// e.g. Addr{1,2,3}.With(1, 99) → Addr{1,99,3}.
func (a Addr) With(dimIdx int, value int) Addr {
	aCoords, dims := a.Coords()
	if dimIdx < 0 || dimIdx >= dims {
		panic(fmt.Sprintf("lattice: dimension index %d out of range [0:%d]", dimIdx, dims))
	}

	coords := make([]int, dims)

	for i := range dims {
		coords[i] = aCoords[i] //nolint:gosec // i < dims <= MaxDimensions == len(aCoords)
	}

	coords[dimIdx] = value

	return New(coords...)
}

// String returns a human-readable representation of the address.
func (a Addr) String() string {
	var buf Buffer

	return fmt.Sprintf("Addr%v", a.CoordsSlice(buf[:]))
}

type AddrRange [2]int
type Buffer [MaxDimensions]int
