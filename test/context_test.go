package cairo

import (
	"math"
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试创建 Context
func TestContextCreation(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	if ctx == nil {
		t.Fatal("Failed to create context")
	}
	defer ctx.Destroy()

	if ctx.Status() != cairo.StatusSuccess {
		t.Errorf("Context status: %v", ctx.Status())
	}
}

// 测试 Context 引用计数
func TestContextReferenceCount(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	initialCount := ctx.GetReferenceCount()
	if initialCount != 1 {
		t.Errorf("Expected initial reference count 1, got %d", initialCount)
	}

	ref := ctx.Reference()
	if ctx.GetReferenceCount() != 2 {
		t.Errorf("Expected reference count 2, got %d", ctx.GetReferenceCount())
	}

	ref.Destroy()
}

// 测试 Save/Restore
func TestContextSaveRestore(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 设置初始状态
	ctx.SetLineWidth(5.0)
	ctx.Save()

	// 修改状态
	ctx.SetLineWidth(10.0)
	if ctx.GetLineWidth() != 10.0 {
		t.Error("Line width should be 10.0")
	}

	// 恢复状态
	ctx.Restore()
	if ctx.GetLineWidth() != 5.0 {
		t.Error("Line width should be restored to 5.0")
	}
}

// 测试设置源颜色
func TestContextSetSource(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 测试 RGB
	ctx.SetSourceRGB(1.0, 0.5, 0.25)
	source := ctx.GetSource()
	defer source.Destroy()

	if solidPattern, ok := source.(cairo.SolidPattern); ok {
		r, g, b, a := solidPattern.GetRGBA()
		if r != 1.0 || g != 0.5 || b != 0.25 || a != 1.0 {
			t.Errorf("Source color mismatch: (%f,%f,%f,%f)", r, g, b, a)
		}
	}
}

// 测试路径操作
func TestContextPath(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 测试 MoveTo/LineTo
	ctx.MoveTo(10, 10)
	if ctx.HasCurrentPoint() != cairo.True {
		t.Error("Should have current point after MoveTo")
	}

	x, y := ctx.GetCurrentPoint()
	if x != 10 || y != 10 {
		t.Errorf("Current point mismatch: (%f, %f)", x, y)
	}

	ctx.LineTo(90, 90)
	x, y = ctx.GetCurrentPoint()
	if x != 90 || y != 90 {
		t.Errorf("Current point after LineTo: (%f, %f)", x, y)
	}
}

// 测试矩形
func TestContextRectangle(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	ctx.Rectangle(10, 10, 80, 80)
	ctx.SetSourceRGB(1.0, 0.0, 0.0)
	err := ctx.Fill()
	if err != nil {
		t.Errorf("Fill failed: %v", err)
	}
}

// 测试圆弧
func TestContextArc(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	ctx.Arc(50, 50, 30, 0, 2*math.Pi)
	ctx.SetSourceRGB(0.0, 0.0, 1.0)
	err := ctx.Fill()
	if err != nil {
		t.Errorf("Fill failed: %v", err)
	}
}

// 测试变换
func TestContextTransform(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 测试平移
	ctx.Translate(10, 20)
	matrix := ctx.GetMatrix()
	if matrix.X0 != 10 || matrix.Y0 != 20 {
		t.Errorf("Translation failed: X0=%f, Y0=%f", matrix.X0, matrix.Y0)
	}

	// 测试缩放
	ctx.IdentityMatrix()
	ctx.Scale(2.0, 3.0)
	matrix = ctx.GetMatrix()
	if matrix.XX != 2.0 || matrix.YY != 3.0 {
		t.Errorf("Scale failed: XX=%f, YY=%f", matrix.XX, matrix.YY)
	}

	// 测试旋转
	ctx.IdentityMatrix()
	ctx.Rotate(math.Pi / 4)
	matrix = ctx.GetMatrix()
	expectedCos := math.Cos(math.Pi / 4)
	if math.Abs(matrix.XX-expectedCos) > 0.0001 {
		t.Errorf("Rotation failed: XX=%f", matrix.XX)
	}
}
