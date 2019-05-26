package gofft

import (
	ktyefft "github.com/ktye/fft"
	dspfft "github.com/mjibson/go-dsp/fft"
	"math/cmplx"
	"math/rand"
	"reflect"
	"testing"
)

func complexRand(N int) []complex128 {
	x := make([]complex128, N)
	for i := 0; i < N; i++ {
		x[i] = complex(2.0*rand.Float64()-1.0, 2.0*rand.Float64()-1.0)
	}
	return x
}

func copyVector(v []complex128) []complex128 {
	y := make([]complex128, len(v))
	copy(y, v)
	return y
}

func TestFFT(t *testing.T) {
	N := 1 << 10
	x := complexRand(N)
	slow := slow{}
	slowPre := newSlowPre(N)
	err := Prepare(N)
	if err != nil {
		t.Errorf("Prepare error: %v", err)
	}

	y1 := slow.Transform(copyVector(x))
	y2 := slowPre.Transform(copyVector(x))
	y3 := copyVector(x)
	err = FFT(y3)
	if err != nil {
		t.Errorf("FFT error: %v", err)
	}
	for i := 0; i < N; i++ {
		if e := cmplx.Abs(y1[i] - y2[i]); e > 1E-9 {
			t.Errorf("slow and slowPre differ: i=%d diff=%v\n", i, e)
		}
		if e := cmplx.Abs(y1[i] - y3[i]); e > 1E-9 {
			t.Errorf("slow and fast differ: i=%d diff=%v\n", i, e)
		}
	}
}

func TestIFFT(t *testing.T) {
	N := 256
	x := complexRand(N)
	err := Prepare(N)
	if err != nil {
		t.Errorf("Prepare error: %v", err)
	}
	y := copyVector(x)
	err = FFT(y)
	if err != nil {
		t.Errorf("FFT error: %v", err)
	}
	err = IFFT(y)
	if err != nil {
		t.Errorf("IFFT error: %v", err)
	}
	for i := range x {
		if e := cmplx.Abs(x[i] - y[i]); e > 1E-9 {
			t.Errorf("inverse differs %d: %v %v\n", i, x[i], y[i])
		}
	}
}

func TestPermutationIndex(t *testing.T) {
	tab := [][]int{
		[]int{0},
		[]int{0, 1},
		[]int{0, 2, 1, 3},
		[]int{0, 4, 2, 6, 1, 5, 3, 7},
		[]int{0, 8, 4, 12, 2, 10, 6, 14, 1, 9, 5, 13, 3, 11, 7, 15},
	}
	for i := 0; i < len(tab); i++ {
		got := permutationIndex(1 << uint32(i))
		expect := tab[i]
		if !reflect.DeepEqual(got, expect) {
			t.Errorf("%d expected: %v, got: %v\n", i, expect, got)
		}
	}
}

var (
	benchmarks = []struct {
		size int
		name string
	}{
		{4, "Tiny (4)"},
		{128, "Small (128)"},
		{4096, "Medium (4096)"},
		{131072, "Large (131072)"},
		{4194304, "Huge (4194304)"},
	}
)

func BenchmarkSlowFFT000(b *testing.B) {
	for _, bm := range benchmarks {
		if bm.size > 10000 {
			// Don't run sizes too big for slow
			continue
		}
		b.Run(bm.name, func(b *testing.B) {
			slow := slow{}
			x := complexRand(bm.size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = slow.Transform(copyVector(x))
			}
		})
	}
}

func BenchmarkSlowFFTPre(b *testing.B) {
	for _, bm := range benchmarks {
		if bm.size > 10000 {
			// Don't run sizes too big for slow
			continue
		}
		b.Run(bm.name, func(b *testing.B) {
			slowPre := newSlowPre(bm.size)
			x := complexRand(bm.size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = slowPre.Transform(copyVector(x))
			}
		})
	}
}

func BenchmarkKtyeFFT(b *testing.B) {
	for _, bm := range benchmarks {
		if bm.size > 1048576 {
			// Max size for ktye's fft
			continue
		}
		b.Run(bm.name, func(b *testing.B) {
			f, err := ktyefft.New(bm.size)
			if err != nil {
				b.Errorf("fft.New error: %v", err)
			}
			x := complexRand(bm.size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f.Transform(copyVector(x))
			}
		})
	}
}

func BenchmarkDSPFFT(b *testing.B) {
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			x := complexRand(bm.size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dspfft.FFT(copyVector(x))
			}
		})
	}
}

func BenchmarkFFT(b *testing.B) {
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			err := Prepare(bm.size)
			if err != nil {
				b.Errorf("Prepare error: %v", err)
			}
			x := complexRand(bm.size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := FFT(copyVector(x))
				if err != nil {
					b.Errorf("FFT error: %v", err)
				}
			}
		})
	}
}
