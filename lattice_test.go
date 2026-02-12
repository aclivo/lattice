package lattice

import (
	"fmt"
	"reflect"
	"testing"
)

// ============================================================
// New
// ============================================================

func TestNew_Dimensions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		coords  []int
		wantDim int
	}{
		{"empty", []int{}, 0},
		{"one dimension", []int{1}, 1},
		{"two dimensions", []int{1, 2}, 2},
		{"three dimensions", []int{1, 2, 3}, 3},
		{"six dimensions", []int{1, 2, 3, 4, 5, 6}, 6},
		{"max dimensions", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, 12},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			addr := New(testCase.coords...)
			if got := addr.Dims(); got != testCase.wantDim {
				t.Errorf("Dims() = %v, want %v", got, testCase.wantDim)
			}
		})
	}
}

func TestNew_PanicTooManyDimensions(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with 13 dimensions")
		}
	}()

	New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13)
}

func TestNew_PanicNegativeCoord(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with negative coordinate")
		}
	}()

	New(1, -1, 3)
}

func TestNew_PanicCoordTooLarge(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with coordinate exceeding MaxCoordValue")
		}
	}()

	New(1, MaxCoordValue+1, 3)
}

func TestNew_PanicMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		coords  []int
		wantMsg string
	}{
		{
			"too many dimensions",
			make([]int, MaxDimensions+1),
			fmt.Sprintf("lattice: max %d dimensions supported", MaxDimensions),
		},
		{
			"coord out of range",
			[]int{0, MaxCoordValue + 1},
			fmt.Sprintf("lattice: coord[1]=%d out of range [0,%d]", MaxCoordValue+1, MaxCoordValue),
		},
		{
			"negative coord",
			[]int{0, -5},
			fmt.Sprintf("lattice: coord[1]=%d out of range [0,%d]", -5, MaxCoordValue),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				rec := recover()

				if rec == nil {
					t.Error("expected panic")

					return
				}

				if got := fmt.Sprintf("%v", rec); got != testCase.wantMsg {
					t.Errorf("panic message = %q, want %q", got, testCase.wantMsg)
				}
			}()

			New(testCase.coords...)
		})
	}
}

// ============================================================
// Coords round-trip
// ============================================================

func TestCoords_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		coords []int
	}{
		{"zeros", []int{0, 0, 0}},
		{"small values", []int{1, 2, 3}},
		{"medium values", []int{1000, 2000, 3000}},
		{"large values", []int{500000, 123456, 999999}},
		{"max values", []int{MaxCoordValue, MaxCoordValue, MaxCoordValue}},
		{"single zero", []int{0}},
		{"single max", []int{MaxCoordValue}},
		{"mixed", []int{0, MaxCoordValue, 1, MaxCoordValue - 1}},
		{"powers of two", []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024}},
		{"alternating high low", []int{MaxCoordValue, 0, MaxCoordValue, 0}},
		{"twelve small", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
		{"twelve large", []int{
			100000, 200000, 300000, 400000, 500000, 600000,
			700000, 800000, 900000, 1000000, 1048575, 999999,
		}},
		{"checkerboard bits", []int{0xAAAAA & MaxCoordValue, 0x55555 & MaxCoordValue}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			addr := New(testCase.coords...)
			got, dims := addr.Coords()

			if dims != len(testCase.coords) {
				t.Errorf("dims = %d, want %d", dims, len(testCase.coords))
			}

			for i := range dims {
				if got[i] != testCase.coords[i] {
					t.Errorf("coord[%d] = %d, want %d", i, got[i], testCase.coords[i])
				}
			}
		})
	}
}

func TestCoords_Sequential(t *testing.T) {
	t.Parallel()

	for testIndex := range 1000 {
		coords := []int{testIndex, testIndex + 1, testIndex + 2}
		addr := New(coords...)
		got, dims := addr.Coords()

		if dims != 3 {
			t.Fatalf("dims = %d, want 3", dims)
		}

		for j := range dims {
			if got[j] != coords[j] {
				t.Errorf("i=%d coord[%d] = %d, want %d", testIndex, j, got[j], coords[j])
			}
		}
	}
}

func TestCoords_MaxCapacity(t *testing.T) {
	t.Parallel()

	coords := make([]int, MaxDimensions)
	for i := range coords {
		coords[i] = MaxCoordValue
	}

	addr := New(coords...)
	got, dims := addr.Coords()

	if dims != MaxDimensions {
		t.Errorf("dims = %d, want %d", dims, MaxDimensions)
	}

	for i := range dims {
		if i >= len(got) {
			t.Fatalf("index out of range")
		}
		if got[i] != MaxCoordValue {
			t.Errorf("coord[%d] = %d, want %d", i, got[i], MaxCoordValue)
		}
	}
}

// ============================================================
// CoordsSlice
// ============================================================

func TestCoordsSlice_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		coords []int
	}{
		{"three dims", []int{100, 200, 300}},
		{"one dim", []int{42}},
		{"max dims", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
		{"large values", []int{500000, 999999, 1048575}},
	}

	buf := make([]int, MaxDimensions)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			addr := New(testCase.coords...)
			got := addr.CoordsSlice(buf)

			if !reflect.DeepEqual(got, testCase.coords) {
				t.Errorf("CoordsSlice() = %v, want %v", got, testCase.coords)
			}
		})
	}
}

func TestCoordsSlice_PanicBufferTooSmall(t *testing.T) {
	t.Parallel()

	addr := New(1, 2, 3)
	buf := make([]int, 2) // too small for 3 dims

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with buffer too small")
		}
	}()

	addr.CoordsSlice(buf)
}

func TestCoordsSlice_PanicMessage(t *testing.T) {
	t.Parallel()

	addr := New(1, 2, 3)
	buf := make([]int, 2)

	defer func() {
		rec := recover()
		if rec == nil {
			t.Error("expected panic")

			return
		}

		want := "lattice: buf too small: need 3, got 2"
		if got := fmt.Sprintf("%v", rec); got != want {
			t.Errorf("panic message = %q, want %q", got, want)
		}
	}()

	addr.CoordsSlice(buf)
}

func TestCoordsSlice_ReusableBuffer(t *testing.T) {
	t.Parallel()

	buf := make([]int, MaxDimensions)

	// Use same buffer for multiple addresses
	addrs := [][]int{
		{1, 2, 3},
		{100, 200},
		{MaxCoordValue, 0, MaxCoordValue},
	}

	for _, want := range addrs {
		addr := New(want...)
		got := addr.CoordsSlice(buf)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("CoordsSlice() = %v, want %v", got, want)
		}
	}
}

// ============================================================
// Dims
// ============================================================

func TestDims(t *testing.T) {
	t.Parallel()

	for d := 0; d <= MaxDimensions; d++ {
		coords := make([]int, d)
		addr := New(coords...)

		if got := addr.Dims(); got != d {
			t.Errorf("Dims() = %d, want %d", got, d)
		}
	}
}

func TestDims_ZeroValue(t *testing.T) {
	t.Parallel()

	var addr Addr
	if got := addr.Dims(); got != 0 {
		t.Errorf("zero value Dims() = %d, want 0", got)
	}
}

// ============================================================
// Uniqueness
// ============================================================

func TestNew_Uniqueness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		a, b         []int
		shouldBeSame bool
	}{
		{"identical", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"different last", []int{1, 2, 3}, []int{1, 2, 4}, false},
		{"reversed", []int{1, 2, 3}, []int{3, 2, 1}, false},
		{"different length", []int{1, 2}, []int{1, 2, 3}, false},
		{"swapped", []int{100, 200}, []int{200, 100}, false},
		{"differ by one", []int{1000000, 500000}, []int{1000000, 500001}, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			same := New(testCase.a...) == New(testCase.b...)
			if same != testCase.shouldBeSame {
				t.Errorf("%v == %v is %v, want %v", testCase.a, testCase.b, same, testCase.shouldBeSame)
			}
		})
	}
}

// ============================================================
// Map key usage
// ============================================================

func TestAddr_AsMapKey(t *testing.T) {
	t.Parallel()

	mAddrs := make(map[Addr]float64)

	addr1 := New(1, 2, 3)
	addr2 := New(4, 5, 6)
	addr3 := New(1, 2, 3) // Same as addr1

	mAddrs[addr1] = 1.0
	mAddrs[addr2] = 2.0
	mAddrs[addr3] = 3.0 // Overwrites addr1

	if len(mAddrs) != 2 {
		t.Errorf("map len = %d, want 2", len(mAddrs))
	}

	if mAddrs[addr1] != 3.0 {
		t.Errorf("mAddrs[addr1] = %f, want 3.0", mAddrs[addr1])
	}

	if mAddrs[addr2] != 2.0 {
		t.Errorf("mAddrs[addr2] = %f, want 2.0", mAddrs[addr2])
	}
}

func TestAddr_MapKeyNotFound(t *testing.T) {
	t.Parallel()

	mAddrs := make(map[Addr]float64)
	mAddrs[New(1, 2, 3)] = 42.0

	if _, ok := mAddrs[New(9, 9, 9)]; ok {
		t.Error("expected key not to be found")
	}
}

func TestAddr_MapDelete(t *testing.T) {
	t.Parallel()

	mAddrs := make(map[Addr]float64)

	addr := New(1, 2, 3)
	mAddrs[addr] = 42.0
	delete(mAddrs, addr)

	if _, ok := mAddrs[addr]; ok {
		t.Error("expected key to be deleted")
	}
}

// ============================================================
// String
// ============================================================

func TestAddr_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		coords []int
		want   string
	}{
		{[]int{}, "Addr[]"},
		{[]int{1}, "Addr[1]"},
		{[]int{1, 2, 3}, "Addr[1 2 3]"},
		{[]int{100, 200, 300}, "Addr[100 200 300]"},
		{[]int{0, MaxCoordValue}, fmt.Sprintf("Addr[0 %d]", MaxCoordValue)},
	}

	for _, testCase := range tests {
		t.Run(testCase.want, func(t *testing.T) {
			t.Parallel()

			got := New(testCase.coords...).String()
			if got != testCase.want {
				t.Errorf("String() = %q, want %q", got, testCase.want)
			}
		})
	}
}

// ============================================================
// Constants
// ============================================================

func TestConstants(t *testing.T) {
	t.Parallel()

	if BitsPerCoord != 20 {
		t.Errorf("BitsPerCoord = %d, want 20", BitsPerCoord)
	}

	if MaxDimensions != 12 {
		t.Errorf("MaxDimensions = %d, want 12", MaxDimensions)
	}

	if MaxCoordValue != 1048575 {
		t.Errorf("MaxCoordValue = %d, want 1048575", MaxCoordValue)
	}
}

// ============================================================
// Benchmarks
// ============================================================

func BenchmarkNew_1D(b *testing.B) {
	for b.Loop() {
		_ = New(12345)
	}
}

func BenchmarkNew_3D(b *testing.B) {
	for b.Loop() {
		_ = New(100, 200, 300)
	}
}

func BenchmarkNew_6D(b *testing.B) {
	for b.Loop() {
		_ = New(100, 200, 300, 400, 500, 600)
	}
}

func BenchmarkNew_12D(b *testing.B) {
	for b.Loop() {
		_ = New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)
	}
}

func BenchmarkNew_12D_Large(b *testing.B) {
	for b.Loop() {
		_ = New(500000, 600000, 700000, 800000, 900000, 1000000,
			100000, 200000, 300000, 400000, 500000, 600000)
	}
}

func BenchmarkCoords_3D(b *testing.B) {
	addr := New(100, 200, 300)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = addr.Coords()
	}
}

func BenchmarkCoords_12D(b *testing.B) {
	addr := New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = addr.Coords()
	}
}

func BenchmarkCoordsSlice_3D(b *testing.B) {
	addr := New(100, 200, 300)
	buf := make([]int, MaxDimensions)

	b.ReportAllocs()

	for b.Loop() {
		_ = addr.CoordsSlice(buf)
	}
}

func BenchmarkCoordsSlice_12D(b *testing.B) {
	addr := New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)
	buf := make([]int, MaxDimensions)

	b.ReportAllocs()

	for b.Loop() {
		_ = addr.CoordsSlice(buf)
	}
}

func BenchmarkDims(b *testing.B) {
	addr := New(1, 2, 3, 4, 5)

	for b.Loop() {
		_ = addr.Dims()
	}
}

func BenchmarkRoundTrip_3D(b *testing.B) {
	coords := []int{100, 200, 300}

	b.ReportAllocs()

	for b.Loop() {
		addr := New(coords...)
		_, _ = addr.Coords()
	}
}

func BenchmarkRoundTrip_12D(b *testing.B) {
	coords := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	b.ReportAllocs()

	for b.Loop() {
		addr := New(coords...)
		_, _ = addr.Coords()
	}
}

func BenchmarkMapInsert_3D(b *testing.B) {
	mAddrs := make(map[Addr]float64, b.N)
	addrs := make([]Addr, b.N)

	for i := range addrs {
		v := i % (MaxCoordValue - 10)
		addrs[i] = New(v, v+1, v+2)
	}

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		mAddrs[addrs[i]] = float64(i)
	}
}

func BenchmarkMapLookup_3D(b *testing.B) {
	mAddrs := make(map[Addr]float64, 10000)
	for i := range 10000 {
		mAddrs[New(i, i+1, i+2)] = float64(i)
	}

	for i := 0; b.Loop(); i++ {
		idx := i % 10000
		_ = mAddrs[New(idx, idx+1, idx+2)]
	}
}

func BenchmarkMapLookup_12D(b *testing.B) {
	mAddrs := make(map[Addr]float64, 10000)
	for i := range 10000 {
		mAddrs[New(i, i+1, i+2, i+3, i+4, i+5, i+6, i+7, i+8, i+9, i+10, i+11)] = float64(i)
	}

	for i := 0; b.Loop(); i++ {
		idx := i % 10000
		_ = mAddrs[New(idx, idx+1, idx+2, idx+3, idx+4, idx+5,
			idx+6, idx+7, idx+8, idx+9, idx+10, idx+11)]
	}
}

func BenchmarkMapIteration_100k(b *testing.B) {
	mAddrs := make(map[Addr]float64, 100000)

	for i := range 100000 {
		v := i % (MaxCoordValue - 10)
		mAddrs[New(v, v+1, v+2)] = float64(i)
	}

	for b.Loop() {
		var sum float64
		for _, v := range mAddrs {
			sum += v
		}

		_ = sum
	}
}

// ============================================================
// Append
// ============================================================

func TestAppend_Basic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		base   []int
		append []int
		want   []int
	}{
		{"append one", []int{1, 2}, []int{3}, []int{1, 2, 3}},
		{"append many", []int{1, 2}, []int{3, 4, 5}, []int{1, 2, 3, 4, 5}},
		{"append to empty", []int{}, []int{1, 2, 3}, []int{1, 2, 3}},
		{"append single", []int{1}, []int{2}, []int{1, 2}},
		{"append large values", []int{500000}, []int{999999, 1048575}, []int{500000, 999999, 1048575}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			addr := New(testCase.base...).Append(testCase.append...)
			got, dims := addr.Coords()

			if dims != len(testCase.want) {
				t.Fatalf("dims = %d, want %d", dims, len(testCase.want))
			}

			for i := range dims {
				if got[i] != testCase.want[i] {
					t.Errorf("coord[%d] = %d, want %d", i, got[i], testCase.want[i])
				}
			}
		})
	}
}

func TestAppend_PreservesOriginal(t *testing.T) {
	t.Parallel()

	original := New(1, 2, 3)
	_ = original.Append(4, 5)

	// Original should be unchanged
	if original.Dims() != 3 {
		t.Errorf("original dims = %d, want 3", original.Dims())
	}

	if original.At(2) != 3 {
		t.Errorf("original coord[2] = %d, want 3", original.At(2))
	}
}

func TestAppend_PanicTooManyDimensions(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when exceeding MaxDimensions")
		}
	}()

	// Start with 10 dimensions, append 3 more = 13 (exceeds MaxDimensions)
	base := make([]int, 10)
	New(base...).Append(1, 2, 3)
}

func TestAppend_Chaining(t *testing.T) {
	t.Parallel()

	addr := New(1).Append(2).Append(3).Append(4)
	want := []int{1, 2, 3, 4}

	got, dims := addr.Coords()
	if dims != len(want) {
		t.Fatalf("dims = %d, want %d", dims, len(want))
	}

	for i := range dims {
		if got[i] != want[i] {
			t.Errorf("coord[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

// ============================================================
// At
// ============================================================

func TestAt_Basic(t *testing.T) {
	t.Parallel()

	addr := New(10, 20, 30)

	tests := []struct {
		dimIdx int
		want   int
	}{
		{0, 10},
		{1, 20},
		{2, 30},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("dim%d", tt.dimIdx), func(t *testing.T) {
			t.Parallel()

			if got := addr.At(tt.dimIdx); got != tt.want {
				t.Errorf("At(%d) = %d, want %d", tt.dimIdx, got, tt.want)
			}
		})
	}
}

func TestAt_LargeValues(t *testing.T) {
	t.Parallel()

	addr := New(500000, 999999, MaxCoordValue)

	if got := addr.At(0); got != 500000 {
		t.Errorf("At(0) = %d, want 500000", got)
	}

	if got := addr.At(1); got != 999999 {
		t.Errorf("At(1) = %d, want 999999", got)
	}

	if got := addr.At(2); got != MaxCoordValue {
		t.Errorf("At(2) = %d, want %d", got, MaxCoordValue)
	}
}

func TestAt_PanicNegativeIndex(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with negative index")
		}
	}()

	New(1, 2, 3).At(-1)
}

func TestAt_PanicIndexOutOfRange(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with index out of range")
		}
	}()

	New(1, 2, 3).At(3)
}

func TestAt_PanicMessage(t *testing.T) {
	t.Parallel()

	defer func() {
		rec := recover()

		if rec == nil {
			t.Error("expected panic")

			return
		}

		want := "lattice: dimension index 5 out of range [0:3]"
		if got := fmt.Sprintf("%v", rec); got != want {
			t.Errorf("panic message = %q, want %q", got, want)
		}
	}()

	New(1, 2, 3).At(5)
}

// ============================================================
// Contains
// ============================================================

func TestContains_Basic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b []int
		want bool
	}{
		{"prefix match", []int{1, 2}, []int{1, 2, 3}, true},
		{"exact match", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"empty contains all", []int{}, []int{1, 2, 3}, true},
		{"longer does not contain shorter", []int{1, 2, 3}, []int{1, 2}, false},
		{"different values", []int{1, 2}, []int{1, 3, 4}, false},
		{"single dim prefix", []int{1}, []int{1, 2, 3}, true},
		{"single dim mismatch", []int{2}, []int{1, 2, 3}, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := New(testCase.a...).Contains(New(testCase.b...))
			if got != testCase.want {
				t.Errorf("Contains() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestContains_NotCommutative(t *testing.T) {
	t.Parallel()

	aAddr := New(1, 2)
	bAddr := New(1, 2, 3)

	if !aAddr.Contains(bAddr) {
		t.Error("a.Contains(b) should be true")
	}

	if bAddr.Contains(aAddr) {
		t.Error("b.Contains(a) should be false")
	}
}

func TestContains_LargeValues(t *testing.T) {
	t.Parallel()

	aAddr := New(500000, 999999)
	bAddr := New(500000, 999999, 1048575)

	if !aAddr.Contains(bAddr) {
		t.Error("expected a to contain b")
	}
}

// ============================================================
// Equal
// ============================================================

func TestEqual_Basic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b []int
		want bool
	}{
		{"identical", []int{1, 2, 3}, []int{1, 2, 3}, true},
		{"different values", []int{1, 2, 3}, []int{1, 2, 4}, false},
		{"different length", []int{1, 2}, []int{1, 2, 3}, false},
		{"empty equal", []int{}, []int{}, true},
		{"single equal", []int{42}, []int{42}, false},
		{"reversed", []int{1, 2, 3}, []int{3, 2, 1}, false},
	}

	// Fix: single value equal should be true
	tests[4].want = true

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := New(testCase.a...).Equal(New(testCase.b...))
			if got != testCase.want {
				t.Errorf("Equal() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestEqual_Reflexive(t *testing.T) {
	t.Parallel()

	addr := New(1, 2, 3)
	if !addr.Equal(addr) {
		t.Error("addr should equal itself")
	}
}

func TestEqual_Symmetric(t *testing.T) {
	t.Parallel()

	a := New(1, 2, 3)
	b := New(1, 2, 3)

	if a.Equal(b) != b.Equal(a) {
		t.Error("Equal should be symmetric")
	}
}

// ============================================================
// InRange
// ============================================================

func TestInRange_Basic(t *testing.T) {
	t.Parallel()

	addr := New(10, 20, 30)

	tests := []struct {
		name   string
		ranges [][2]int
		want   bool
	}{
		{"all match", [][2]int{{5, 15}, {15, 25}, {25, 35}}, true},
		{"first fails", [][2]int{{15, 20}, {15, 25}, {25, 35}}, false},
		{"second fails", [][2]int{{5, 15}, {25, 30}, {25, 35}}, false},
		{"third fails", [][2]int{{5, 15}, {15, 25}, {35, 40}}, false},
		{"any wildcard", [][2]int{{-1, -1}, {-1, -1}, {-1, -1}}, true},
		{"exact match", [][2]int{{10, 10}, {20, 20}, {30, 30}}, true},
		{"partial ranges", [][2]int{{5, 15}}, true},
		{"only min", [][2]int{{5, -1}, {15, -1}, {25, -1}}, true},
		{"only max", [][2]int{{-1, 15}, {-1, 25}, {-1, 35}}, true},
		{"min fails", [][2]int{{11, -1}}, false},
		{"max fails", [][2]int{{-1, 9}}, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := addr.InRange(testCase.ranges...)
			if got != testCase.want {
				t.Errorf("InRange() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestInRange_FewerRangesThanDims(t *testing.T) {
	t.Parallel()

	addr := New(10, 20, 30)

	// Only check first two dimensions
	if !addr.InRange([2]int{5, 15}, [2]int{15, 25}) {
		t.Error("expected true with fewer ranges than dims")
	}
}

func TestInRange_MoreRangesThanDims(t *testing.T) {
	t.Parallel()

	addr := New(10, 20)

	// Extra ranges are ignored
	if !addr.InRange([2]int{5, 15}, [2]int{15, 25}, [2]int{25, 35}) {
		t.Error("expected true when extra ranges are ignored")
	}
}

func TestInRange_BoundaryValues(t *testing.T) {
	t.Parallel()

	addr := New(0, MaxCoordValue)

	if !addr.InRange([2]int{0, 0}, [2]int{MaxCoordValue, MaxCoordValue}) {
		t.Error("expected boundary values to match exactly")
	}
}

// ============================================================
// IsZero
// ============================================================

func TestIsZero_Basic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		coords []int
		want   bool
	}{
		{"all zeros", []int{0, 0, 0}, true},
		{"single zero", []int{0}, true},
		{"has nonzero", []int{0, 0, 1}, false},
		{"all nonzero", []int{1, 2, 3}, false},
		{"first nonzero", []int{1, 0, 0}, false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := New(testCase.coords...).IsZero()
			if got != testCase.want {
				t.Errorf("IsZero() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestIsZero_EmptyAddr(t *testing.T) {
	t.Parallel()

	// Empty address has no coordinates - all zero vacuously
	if !New().IsZero() {
		t.Error("empty address should be zero")
	}
}

func TestIsZero_ZeroValue(t *testing.T) {
	t.Parallel()

	var addr Addr
	if !addr.IsZero() {
		t.Error("zero value Addr should be zero")
	}
}

// ============================================================
// Slice
// ============================================================

func TestSlice_Basic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		coords   []int
		from, to int
		want     []int
	}{
		{"first two", []int{1, 2, 3, 4}, 0, 2, []int{1, 2}},
		{"last two", []int{1, 2, 3, 4}, 2, 4, []int{3, 4}},
		{"middle", []int{1, 2, 3, 4, 5}, 1, 4, []int{2, 3, 4}},
		{"all", []int{1, 2, 3}, 0, 3, []int{1, 2, 3}},
		{"single", []int{1, 2, 3}, 1, 2, []int{2}},
		{"empty slice", []int{1, 2, 3}, 1, 1, []int{}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			addr := New(testCase.coords...).Slice(testCase.from, testCase.to)
			got, dims := addr.Coords()

			if dims != len(testCase.want) {
				t.Fatalf("dims = %d, want %d", dims, len(testCase.want))
			}

			for i := range dims {
				if got[i] != testCase.want[i] {
					t.Errorf("coord[%d] = %d, want %d", i, got[i], testCase.want[i])
				}
			}
		})
	}
}

func TestSlice_PanicFromNegative(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with negative from")
		}
	}()

	New(1, 2, 3).Slice(-1, 2)
}

func TestSlice_PanicToOutOfRange(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with to out of range")
		}
	}()

	New(1, 2, 3).Slice(0, 4)
}

func TestSlice_PanicFromGreaterThanTo(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with from > to")
		}
	}()

	New(1, 2, 3).Slice(2, 1)
}

func TestSlice_PanicMessage(t *testing.T) {
	t.Parallel()

	defer func() {
		rec := recover()
		if rec == nil {
			t.Error("expected panic")

			return
		}

		want := "lattice: slice [1:5] out of range [0:3]"
		if got := fmt.Sprintf("%v", rec); got != want {
			t.Errorf("panic message = %q, want %q", got, want)
		}
	}()

	New(1, 2, 3).Slice(1, 5)
}

func TestSlice_PreservesOriginal(t *testing.T) {
	t.Parallel()

	original := New(1, 2, 3, 4, 5)
	_ = original.Slice(1, 3)

	if original.Dims() != 5 {
		t.Errorf("original dims = %d, want 5", original.Dims())
	}
}

// ============================================================
// With
// ============================================================

func TestWith_Basic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		coords []int
		dimIdx int
		value  int
		want   []int
	}{
		{"replace first", []int{1, 2, 3}, 0, 99, []int{99, 2, 3}},
		{"replace middle", []int{1, 2, 3}, 1, 99, []int{1, 99, 3}},
		{"replace last", []int{1, 2, 3}, 2, 99, []int{1, 2, 99}},
		{"replace with zero", []int{1, 2, 3}, 1, 0, []int{1, 0, 3}},
		{"replace with max", []int{1, 2, 3}, 1, MaxCoordValue, []int{1, MaxCoordValue, 3}},
		{"replace single dim", []int{42}, 0, 99, []int{99}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			addr := New(testCase.coords...).With(testCase.dimIdx, testCase.value)
			got, dims := addr.Coords()

			if dims != len(testCase.want) {
				t.Fatalf("dims = %d, want %d", dims, len(testCase.want))
			}

			for i := range dims {
				if got[i] != testCase.want[i] {
					t.Errorf("coord[%d] = %d, want %d", i, got[i], testCase.want[i])
				}
			}
		})
	}
}

func TestWith_PanicNegativeIndex(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with negative index")
		}
	}()

	New(1, 2, 3).With(-1, 99)
}

func TestWith_PanicIndexOutOfRange(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with index out of range")
		}
	}()

	New(1, 2, 3).With(3, 99)
}

func TestWith_PanicMessage(t *testing.T) {
	t.Parallel()

	defer func() {
		rec := recover()

		if rec == nil {
			t.Error("expected panic")

			return
		}

		want := "lattice: dimension index 5 out of range [0:3]"
		if got := fmt.Sprintf("%v", rec); got != want {
			t.Errorf("panic message = %q, want %q", got, want)
		}
	}()

	New(1, 2, 3).With(5, 99)
}

func TestWith_PreservesOriginal(t *testing.T) {
	t.Parallel()

	original := New(1, 2, 3)
	_ = original.With(1, 99)

	if original.At(1) != 2 {
		t.Errorf("original coord[1] = %d, want 2", original.At(1))
	}
}

func TestWith_Chaining(t *testing.T) {
	t.Parallel()

	addr := New(0, 0, 0).
		With(0, 10).
		With(1, 20).
		With(2, 30)

	want := []int{10, 20, 30}
	got, dims := addr.Coords()

	if dims != len(want) {
		t.Fatalf("dims = %d, want %d", dims, len(want))
	}

	for i := range dims {
		if i >= len(got) {
			t.Fatalf("index out of range")
		}
		if got[i] != want[i] {
			t.Errorf("coord[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

// ============================================================
// Method interactions
// ============================================================

func TestMethodInteractions(t *testing.T) {
	t.Run("Append then At", func(t *testing.T) {
		t.Parallel()

		addr := New(1, 2).Append(3)
		if got := addr.At(2); got != 3 {
			t.Errorf("At(2) = %d, want 3", got)
		}
	})

	t.Run("With then Equal", func(t *testing.T) {
		t.Parallel()

		a := New(1, 2, 3).With(1, 99)
		b := New(1, 99, 3)

		if !a.Equal(b) {
			t.Error("expected addresses to be equal after With")
		}
	})

	t.Run("Slice then Contains", func(t *testing.T) {
		t.Parallel()

		original := New(1, 2, 3, 4, 5)
		prefix := original.Slice(0, 3)

		if !prefix.Contains(original) {
			t.Error("expected sliced prefix to contain original")
		}
	})

	t.Run("Append then Slice", func(t *testing.T) {
		t.Parallel()

		addr := New(1, 2).Append(3, 4, 5).Slice(1, 4)
		want := []int{2, 3, 4}
		got, dims := addr.Coords()

		if dims != len(want) {
			t.Fatalf("dims = %d, want %d", dims, len(want))
		}

		for i := range dims {
			if i >= len(got) {
				t.Fatalf("index out of range")
			}
			if got[i] != want[i] {
				t.Errorf("coord[%d] = %d, want %d", i, got[i], want[i])
			}
		}
	})

	t.Run("IsZero after With", func(t *testing.T) {
		t.Parallel()

		addr := New(0, 0, 0).With(1, 1)
		if addr.IsZero() {
			t.Error("expected IsZero to be false after With")
		}
	})

	t.Run("InRange after Append", func(t *testing.T) {
		t.Parallel()

		addr := New(10, 20).Append(30)
		if !addr.InRange([2]int{5, 15}, [2]int{15, 25}, [2]int{25, 35}) {
			t.Error("expected InRange to be true after Append")
		}
	})
}

// ============================================================
// Benchmarks
// ============================================================

func BenchmarkAppend_One(b *testing.B) {
	addr := New(1, 2, 3)

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		_ = addr.Append(4)
	}
}

func BenchmarkAt(b *testing.B) {
	addr := New(10, 20, 30, 40, 50)

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		_ = addr.At(2)
	}
}

func BenchmarkContains(b *testing.B) {
	aAddr := New(1, 2)
	bAddr := New(1, 2, 3, 4, 5)

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		_ = aAddr.Contains(bAddr)
	}
}

func BenchmarkEqual(b *testing.B) {
	aAddr := New(1, 2, 3, 4, 5)
	bAddr := New(1, 2, 3, 4, 5)

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		_ = aAddr.Equal(bAddr)
	}
}

func BenchmarkInRange_3D(b *testing.B) {
	addr := New(100, 200, 300)
	ranges := [][2]int{{50, 150}, {150, 250}, {250, 350}}

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		_ = addr.InRange(ranges...)
	}
}

func BenchmarkIsZero(b *testing.B) {
	addr := New(0, 0, 0, 0, 0)

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		_ = addr.IsZero()
	}
}

func BenchmarkSlice(b *testing.B) {
	addr := New(1, 2, 3, 4, 5, 6)

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		_ = addr.Slice(1, 4)
	}
}

func BenchmarkWith(b *testing.B) {
	addr := New(1, 2, 3, 4, 5)

	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		_ = addr.With(2, 99)
	}
}
