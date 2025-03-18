package terror_test

import (
	"errors"
	"github.com/towgo/towgo/errors/tcode"
	"github.com/towgo/towgo/errors/terror"

	"testing"
)

var (
	// base error for benchmark testing of Wrap* functions.
	baseError = errors.New("test")
)

func Benchmark_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.New("test")
	}
}

func Benchmark_Newf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.
			Newf("%s", "test")
	}
}

func Benchmark_Wrap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.Wrap(baseError, "test")
	}
}

func Benchmark_Wrapf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.Wrapf(baseError, "%s", "test")
	}
}

func Benchmark_NewSkip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.NewSkip(1, "test")
	}
}

func Benchmark_NewSkipf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.NewSkipf(1, "%s", "test")
	}
}

func Benchmark_NewCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.NewCode(tcode.New(500, "", nil), "test")
	}
}

func Benchmark_NewCodef(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.NewCodef(tcode.New(500, "", nil), "%s", "test")
	}
}

func Benchmark_NewCodeSkip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.NewCodeSkip(tcode.New(1, "", nil), 500, "test")
	}
}

func Benchmark_NewCodeSkipf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.NewCodeSkipf(tcode.New(1, "", nil), 500, "%s", "test")
	}
}

func Benchmark_WrapCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.WrapCode(tcode.New(500, "", nil), baseError, "test")
	}
}

func Benchmark_WrapCodef(b *testing.B) {
	for i := 0; i < b.N; i++ {
		terror.WrapCodef(tcode.New(500, "", nil), baseError, "test")
	}
}
