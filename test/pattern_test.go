package cairo

import (
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试纯色 Pattern
func TestSolidPattern(t *testing.T) {
	// 测试 RGB
	pattern := cairo.NewPatternRGB(1.0, 0.5, 0.25)
	if pattern == nil {
		t.Fatal("Failed to create RGB pattern")
	}
	defer pattern.Destroy()

	if pattern.Status() != cairo.StatusSuccess {
		t.Errorf("Pattern status: %v", pattern.Status())
	}

	if pattern.GetType() != cairo.PatternTypeSolid {
		t.Errorf("Expected PatternTypeSolid, got %v", pattern.GetType())
	}

	// 测试 RGBA
	patternRGBA := cairo.NewPatternRGBA(1.0, 0.5, 0.25, 0.8)
	defer patternRGBA.Destroy()

	if solidPattern, ok := patternRGBA.(cairo.SolidPattern); ok {
		r, g, b, a := solidPattern.GetRGBA()
		if r != 1.0 || g != 0.5 || b != 0.25 || a != 0.8 {
			t.Errorf("RGBA mismatch: got (%f,%f,%f,%f)", r, g, b, a)
		}
	} else {
		t.Error("Pattern is not a SolidPattern")
	}
}

// 测试线性渐变 Pattern
func TestLinearGradientPattern(t *testing.T) {
	pattern := cairo.NewPatternLinear(0, 0, 100, 100)
	if pattern == nil {
		t.Fatal("Failed to create linear gradient pattern")
	}
	defer pattern.Destroy()

	if pattern.GetType() != cairo.PatternTypeLinear {
		t.Errorf("Expected PatternTypeLinear, got %v", pattern.GetType())
	}

	// 添加颜色停止点
	if gradPattern, ok := pattern.(cairo.LinearGradientPattern); ok {
		status := gradPattern.AddColorStopRGB(0.0, 1.0, 0.0, 0.0)
		if status != cairo.StatusSuccess {
			t.Errorf("Failed to add color stop: %v", status)
		}

		status = gradPattern.AddColorStopRGB(1.0, 0.0, 0.0, 1.0)
		if status != cairo.StatusSuccess {
			t.Errorf("Failed to add color stop: %v", status)
		}

		count := gradPattern.GetColorStopCount()
		if count != 2 {
			t.Errorf("Expected 2 color stops, got %d", count)
		}

		// 获取线性渐变点
		x0, y0, x1, y1 := gradPattern.GetLinearPoints()
		if x0 != 0 || y0 != 0 || x1 != 100 || y1 != 100 {
			t.Errorf("Linear points mismatch: (%f,%f) -> (%f,%f)", x0, y0, x1, y1)
		}
	} else {
		t.Error("Pattern is not a LinearGradientPattern")
	}
}

// 测试径向渐变 Pattern
func TestRadialGradientPattern(t *testing.T) {
	pattern := cairo.NewPatternRadial(50, 50, 10, 50, 50, 50)
	if pattern == nil {
		t.Fatal("Failed to create radial gradient pattern")
	}
	defer pattern.Destroy()

	if pattern.GetType() != cairo.PatternTypeRadial {
		t.Errorf("Expected PatternTypeRadial, got %v", pattern.GetType())
	}

	if gradPattern, ok := pattern.(cairo.RadialGradientPattern); ok {
		status := gradPattern.AddColorStopRGBA(0.0, 1.0, 1.0, 1.0, 1.0)
		if status != cairo.StatusSuccess {
			t.Errorf("Failed to add color stop: %v", status)
		}

		status = gradPattern.AddColorStopRGBA(1.0, 0.0, 0.0, 0.0, 0.0)
		if status != cairo.StatusSuccess {
			t.Errorf("Failed to add color stop: %v", status)
		}

		// 获取径向渐变圆
		cx0, cy0, r0, cx1, cy1, r1 := gradPattern.GetRadialCircles()
		if cx0 != 50 || cy0 != 50 || r0 != 10 || cx1 != 50 || cy1 != 50 || r1 != 50 {
			t.Errorf("Radial circles mismatch")
		}
	}
}

// 测试 Pattern 矩阵变换
func TestPatternMatrix(t *testing.T) {
	pattern := cairo.NewPatternRGB(1.0, 0.0, 0.0)
	defer pattern.Destroy()

	// 设置矩阵
	matrix := cairo.NewMatrix()
	matrix.InitScale(2.0, 2.0)
	pattern.SetMatrix(matrix)

	// 获取矩阵
	resultMatrix := pattern.GetMatrix()
	if resultMatrix.XX != 2.0 || resultMatrix.YY != 2.0 {
		t.Errorf("Matrix mismatch: XX=%f, YY=%f", resultMatrix.XX, resultMatrix.YY)
	}
}

// 测试 Pattern 扩展模式
func TestPatternExtend(t *testing.T) {
	pattern := cairo.NewPatternRGB(1.0, 0.0, 0.0)
	defer pattern.Destroy()

	// 测试各种扩展模式
	extends := []cairo.Extend{
		cairo.ExtendNone,
		cairo.ExtendRepeat,
		cairo.ExtendReflect,
		cairo.ExtendPad,
	}

	for _, extend := range extends {
		pattern.SetExtend(extend)
		if pattern.GetExtend() != extend {
			t.Errorf("Extend mode mismatch: expected %v, got %v", extend, pattern.GetExtend())
		}
	}
}

// 测试 Pattern 过滤模式
func TestPatternFilter(t *testing.T) {
	pattern := cairo.NewPatternRGB(1.0, 0.0, 0.0)
	defer pattern.Destroy()

	// 测试各种过滤模式
	filters := []cairo.Filter{
		cairo.FilterFast,
		cairo.FilterGood,
		cairo.FilterBest,
		cairo.FilterNearest,
		cairo.FilterBilinear,
	}

	for _, filter := range filters {
		pattern.SetFilter(filter)
		if pattern.GetFilter() != filter {
			t.Errorf("Filter mode mismatch: expected %v, got %v", filter, pattern.GetFilter())
		}
	}
}

// 测试 Mesh Pattern
func TestMeshPattern(t *testing.T) {
	pattern := cairo.NewPatternMesh()
	if pattern == nil {
		t.Fatal("Failed to create mesh pattern")
	}
	defer pattern.Destroy()

	if pattern.GetType() != cairo.PatternTypeMesh {
		t.Errorf("Expected PatternTypeMesh, got %v", pattern.GetType())
	}
}
