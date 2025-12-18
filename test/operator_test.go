package cairo

import (
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试设置和获取操作符
func TestSetGetOperator(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

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
		cairo.OperatorMultiply,
		cairo.OperatorScreen,
		cairo.OperatorOverlay,
		cairo.OperatorDarken,
		cairo.OperatorLighten,
	}

	for _, op := range operators {
		ctx.SetOperator(op)
		if ctx.GetOperator() != op {
			t.Errorf("Operator mismatch: expected %v, got %v", op, ctx.GetOperator())
		}
	}
}

// 测试线条属性
func TestLineProperties(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 测试线宽
	ctx.SetLineWidth(5.0)
	if ctx.GetLineWidth() != 5.0 {
		t.Errorf("Line width mismatch: expected 5.0, got %f", ctx.GetLineWidth())
	}

	// 测试线帽样式
	lineCaps := []cairo.LineCap{
		cairo.LineCapButt,
		cairo.LineCapRound,
		cairo.LineCapSquare,
	}
	for _, cap := range lineCaps {
		ctx.SetLineCap(cap)
		if ctx.GetLineCap() != cap {
			t.Errorf("Line cap mismatch: expected %v, got %v", cap, ctx.GetLineCap())
		}
	}

	// 测试线连接样式
	lineJoins := []cairo.LineJoin{
		cairo.LineJoinMiter,
		cairo.LineJoinRound,
		cairo.LineJoinBevel,
	}
	for _, join := range lineJoins {
		ctx.SetLineJoin(join)
		if ctx.GetLineJoin() != join {
			t.Errorf("Line join mismatch: expected %v, got %v", join, ctx.GetLineJoin())
		}
	}

	// 测试斜接限制
	ctx.SetMiterLimit(15.0)
	if ctx.GetMiterLimit() != 15.0 {
		t.Errorf("Miter limit mismatch: expected 15.0, got %f", ctx.GetMiterLimit())
	}
}

// 测试虚线
func TestDash(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 设置虚线模式
	dashes := []float64{10.0, 5.0, 2.0, 5.0}
	offset := 3.0
	ctx.SetDash(dashes, offset)

	// 获取并验证
	resultDashes, resultOffset := ctx.GetDash()
	if len(resultDashes) != len(dashes) {
		t.Errorf("Dash count mismatch: expected %d, got %d", len(dashes), len(resultDashes))
	}

	for i, d := range dashes {
		if resultDashes[i] != d {
			t.Errorf("Dash[%d] mismatch: expected %f, got %f", i, d, resultDashes[i])
		}
	}

	if resultOffset != offset {
		t.Errorf("Dash offset mismatch: expected %f, got %f", offset, resultOffset)
	}

	// 测试虚线数量
	if ctx.GetDashCount() != len(dashes) {
		t.Errorf("GetDashCount mismatch: expected %d, got %d", len(dashes), ctx.GetDashCount())
	}
}

// 测试填充规则
func TestFillRule(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	fillRules := []cairo.FillRule{
		cairo.FillRuleWinding,
		cairo.FillRuleEvenOdd,
	}

	for _, rule := range fillRules {
		ctx.SetFillRule(rule)
		if ctx.GetFillRule() != rule {
			t.Errorf("Fill rule mismatch: expected %v, got %v", rule, ctx.GetFillRule())
		}
	}
}

// 测试容差
func TestTolerance(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	tolerances := []float64{0.01, 0.1, 0.5, 1.0}
	for _, tol := range tolerances {
		ctx.SetTolerance(tol)
		if ctx.GetTolerance() != tol {
			t.Errorf("Tolerance mismatch: expected %f, got %f", tol, ctx.GetTolerance())
		}
	}
}

// 测试抗锯齿
func TestAntialias(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	antialiases := []cairo.Antialias{
		cairo.AntialiasDefault,
		cairo.AntialiasNone,
		cairo.AntialiasGray,
		cairo.AntialiasSubpixel,
		cairo.AntialiasFast,
		cairo.AntialiasGood,
		cairo.AntialiasBest,
	}

	for _, aa := range antialiases {
		ctx.SetAntialias(aa)
		if ctx.GetAntialias() != aa {
			t.Errorf("Antialias mismatch: expected %v, got %v", aa, ctx.GetAntialias())
		}
	}
}

// 测试不同操作符的绘制效果
func TestOperatorDrawing(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 先绘制背景
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// 使用不同操作符绘制
	operators := []cairo.Operator{
		cairo.OperatorOver,
		cairo.OperatorSource,
		cairo.OperatorMultiply,
	}

	for i, op := range operators {
		ctx.SetOperator(op)
		ctx.SetSourceRGBA(1, 0, 0, 0.5)
		ctx.Rectangle(float64(10+i*30), 10, 20, 80)
		ctx.Fill()
	}
}

// 测试虚线绘制
func TestDashDrawing(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 绘制实线
	ctx.SetLineWidth(2)
	ctx.MoveTo(10, 20)
	ctx.LineTo(190, 20)
	ctx.Stroke()

	// 绘制虚线
	ctx.SetDash([]float64{10, 5}, 0)
	ctx.MoveTo(10, 40)
	ctx.LineTo(190, 40)
	ctx.Stroke()

	// 绘制点划线
	ctx.SetDash([]float64{10, 5, 2, 5}, 0)
	ctx.MoveTo(10, 60)
	ctx.LineTo(190, 60)
	ctx.Stroke()
}
