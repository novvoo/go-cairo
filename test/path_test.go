package cairo

import (
	"math"
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试路径创建和清除
func TestPathCreation(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 创建路径
	ctx.MoveTo(10, 10)
	ctx.LineTo(90, 90)

	if ctx.HasCurrentPoint() != cairo.True {
		t.Error("Should have current point")
	}

	// 清除路径
	ctx.NewPath()
	if ctx.HasCurrentPoint() != cairo.False {
		t.Error("Should not have current point after NewPath")
	}
}

// 测试 ClosePath
func TestClosePath(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 创建一个三角形
	ctx.MoveTo(50, 10)
	ctx.LineTo(90, 90)
	ctx.LineTo(10, 90)
	ctx.ClosePath()

	// 填充应该成功
	ctx.SetSourceRGB(1, 0, 0)
	err := ctx.Fill()
	if err != nil {
		t.Errorf("Fill failed: %v", err)
	}
}

// 测试 CurveTo
func TestCurveTo(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	ctx.MoveTo(10, 50)
	ctx.CurveTo(30, 10, 70, 90, 90, 50)

	ctx.SetSourceRGB(0, 0, 1)
	err := ctx.Stroke()
	if err != nil {
		t.Errorf("Stroke failed: %v", err)
	}
}

// 测试相对路径操作
func TestRelativePathOperations(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 测试 RelMoveTo
	ctx.MoveTo(10, 10)
	ctx.RelMoveTo(10, 10)
	x, y := ctx.GetCurrentPoint()
	if x != 20 || y != 20 {
		t.Errorf("RelMoveTo failed: expected (20, 20), got (%f, %f)", x, y)
	}

	// 测试 RelLineTo
	ctx.RelLineTo(30, 30)
	x, y = ctx.GetCurrentPoint()
	if x != 50 || y != 50 {
		t.Errorf("RelLineTo failed: expected (50, 50), got (%f, %f)", x, y)
	}

	// 测试 RelCurveTo
	ctx.RelCurveTo(10, 10, 20, 20, 30, 30)
	x, y = ctx.GetCurrentPoint()
	if x != 80 || y != 80 {
		t.Errorf("RelCurveTo failed: expected (80, 80), got (%f, %f)", x, y)
	}
}

// 测试 Arc
func TestArc(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 绘制完整圆
	ctx.Arc(50, 50, 30, 0, 2*math.Pi)
	ctx.SetSourceRGB(1, 0, 0)
	err := ctx.Stroke()
	if err != nil {
		t.Errorf("Arc stroke failed: %v", err)
	}
}

// 测试 ArcNegative
func TestArcNegative(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 绘制逆时针圆弧
	ctx.ArcNegative(50, 50, 30, 0, -math.Pi)
	ctx.SetSourceRGB(0, 1, 0)
	err := ctx.Stroke()
	if err != nil {
		t.Errorf("ArcNegative stroke failed: %v", err)
	}
}

// 测试 DrawCircle
func TestDrawCircle(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	ctx.DrawCircle(50, 50, 30)
	ctx.SetSourceRGB(0, 0, 1)
	err := ctx.Fill()
	if err != nil {
		t.Errorf("DrawCircle fill failed: %v", err)
	}
}

// 测试复杂路径
func TestComplexPath(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 200)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 绘制星形
	ctx.MoveTo(100, 20)
	for i := 1; i < 10; i++ {
		angle := float64(i) * 2 * math.Pi / 10
		radius := 80.0
		if i%2 == 0 {
			radius = 40.0
		}
		x := 100 + radius*math.Sin(angle)
		y := 100 - radius*math.Cos(angle)
		ctx.LineTo(x, y)
	}
	ctx.ClosePath()

	ctx.SetSourceRGB(1, 1, 0)
	err := ctx.Fill()
	if err != nil {
		t.Errorf("Complex path fill failed: %v", err)
	}
}

// 测试 NewSubPath
func TestNewSubPath(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 第一个子路径
	ctx.MoveTo(10, 10)
	ctx.LineTo(30, 30)

	// 开始新的子路径
	ctx.NewSubPath()
	ctx.MoveTo(50, 50)
	ctx.LineTo(70, 70)

	ctx.SetSourceRGB(0, 0, 0)
	err := ctx.Stroke()
	if err != nil {
		t.Errorf("NewSubPath stroke failed: %v", err)
	}
}

// 测试 StrokePreserve 和 FillPreserve
func TestPreserveOperations(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 创建路径
	ctx.Rectangle(10, 10, 80, 80)

	// 先填充（保留路径）
	ctx.SetSourceRGB(1, 0, 0)
	err := ctx.FillPreserve()
	if err != nil {
		t.Errorf("FillPreserve failed: %v", err)
	}

	// 再描边（使用同一路径）
	ctx.SetSourceRGB(0, 0, 0)
	ctx.SetLineWidth(2)
	err = ctx.StrokePreserve()
	if err != nil {
		t.Errorf("StrokePreserve failed: %v", err)
	}
}

// 基准测试：路径创建
func BenchmarkPathCreation(b *testing.B) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.NewPath()
		ctx.MoveTo(10, 10)
		ctx.LineTo(90, 90)
		ctx.LineTo(90, 10)
		ctx.ClosePath()
	}
}

// 基准测试：圆弧绘制
func BenchmarkArcDrawing(b *testing.B) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.NewPath()
		ctx.Arc(50, 50, 30, 0, 2*math.Pi)
		ctx.Stroke()
	}
}
