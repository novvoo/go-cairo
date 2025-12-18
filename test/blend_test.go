package cairo

import (
	"image/color"
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试 Porter-Duff Over 操作
func TestPorterDuffOver(t *testing.T) {
	src := color.NRGBA{R: 255, G: 0, B: 0, A: 128}
	dst := color.NRGBA{R: 0, G: 0, B: 255, A: 128}

	result := cairo.PorterDuffBlend(src, dst, cairo.OperatorOver)

	if result.A == 0 {
		t.Error("Over blend should not produce transparent result")
	}
}

// 测试 Porter-Duff Clear 操作
func TestPorterDuffClear(t *testing.T) {
	src := color.NRGBA{R: 255, G: 0, B: 0, A: 128}
	dst := color.NRGBA{R: 0, G: 0, B: 255, A: 128}

	result := cairo.PorterDuffBlend(src, dst, cairo.OperatorClear)

	if result.A != 0 {
		t.Error("Clear should produce transparent result")
	}
}

// 测试 Porter-Duff Source 操作
func TestPorterDuffSource(t *testing.T) {
	src := color.NRGBA{R: 255, G: 0, B: 0, A: 200}
	dst := color.NRGBA{R: 0, G: 0, B: 255, A: 128}

	result := cairo.PorterDuffBlend(src, dst, cairo.OperatorSource)

	// Source 操作应该返回源颜色
	if result.R != src.R || result.A != src.A {
		t.Errorf("Source blend failed: got R=%d A=%d, expected R=%d A=%d",
			result.R, result.A, src.R, src.A)
	}
}

// 测试所有 Porter-Duff 操作符
func TestAllPorterDuffOperators(t *testing.T) {
	src := color.NRGBA{R: 255, G: 128, B: 64, A: 200}
	dst := color.NRGBA{R: 64, G: 128, B: 255, A: 200}

	operators := []cairo.Operator{
		cairo.OperatorClear,
		cairo.OperatorSource,
		cairo.OperatorOver,
		cairo.OperatorIn,
		cairo.OperatorOut,
		cairo.OperatorAtop,
		cairo.OperatorDest,
		cairo.OperatorDestOver,
		cairo.OperatorDestIn,
		cairo.OperatorDestOut,
		cairo.OperatorDestAtop,
		cairo.OperatorXor,
		cairo.OperatorAdd,
		cairo.OperatorSaturate,
	}

	for _, op := range operators {
		result := cairo.PorterDuffBlend(src, dst, op)
		// 只验证不会 panic
		_ = result
	}
}

// 基准测试：Porter-Duff Over
func BenchmarkPorterDuffOver(b *testing.B) {
	src := color.NRGBA{R: 255, G: 128, B: 64, A: 200}
	dst := color.NRGBA{R: 64, G: 128, B: 255, A: 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cairo.PorterDuffBlend(src, dst, cairo.OperatorOver)
	}
}
