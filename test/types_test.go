package cairo

import (
	"math"
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试矩阵创建和初始化
func TestMatrixCreation(t *testing.T) {
	m := cairo.NewMatrix()
	if m == nil {
		t.Fatal("Failed to create matrix")
	}

	// 验证单位矩阵
	if m.XX != 1.0 || m.YY != 1.0 || m.XY != 0.0 || m.YX != 0.0 || m.X0 != 0.0 || m.Y0 != 0.0 {
		t.Error("NewMatrix should create identity matrix")
	}
}

// 测试矩阵平移
func TestMatrixTranslate(t *testing.T) {
	m := cairo.NewMatrix()
	m.InitTranslate(10.0, 20.0)

	if m.X0 != 10.0 || m.Y0 != 20.0 {
		t.Errorf("Translation failed: X0=%f, Y0=%f", m.X0, m.Y0)
	}
}

// 测试矩阵缩放
func TestMatrixScale(t *testing.T) {
	m := cairo.NewMatrix()
	m.InitScale(2.0, 3.0)

	if m.XX != 2.0 || m.YY != 3.0 {
		t.Errorf("Scale failed: XX=%f, YY=%f", m.XX, m.YY)
	}
}

// 测试矩阵旋转
func TestMatrixRotate(t *testing.T) {
	m := cairo.NewMatrix()
	angle := math.Pi / 4 // 45 degrees

	m.InitRotate(angle)

	expectedCos := math.Cos(angle)
	expectedSin := math.Sin(angle)

	if math.Abs(m.XX-expectedCos) > 0.0001 || math.Abs(m.YX-expectedSin) > 0.0001 {
		t.Errorf("Rotation failed: XX=%f (expected %f), YX=%f (expected %f)",
			m.XX, expectedCos, m.YX, expectedSin)
	}
}

// 测试矩阵变换点
func TestMatrixTransformPoint(t *testing.T) {
	m := cairo.NewMatrix()
	m.InitTranslate(10.0, 20.0)

	x, y := cairo.MatrixTransformPoint(m, 5.0, 5.0)

	if x != 15.0 || y != 25.0 {
		t.Errorf("Transform point failed: got (%f, %f), expected (15.0, 25.0)", x, y)
	}
}

// 测试矩阵变换距离
func TestMatrixTransformDistance(t *testing.T) {
	m := cairo.NewMatrix()
	m.InitScale(2.0, 3.0)

	dx, dy := cairo.MatrixTransformDistance(m, 10.0, 10.0)

	if dx != 20.0 || dy != 30.0 {
		t.Errorf("Transform distance failed: got (%f, %f), expected (20.0, 30.0)", dx, dy)
	}
}

// 测试矩阵求逆
func TestMatrixInvert(t *testing.T) {
	m := cairo.NewMatrix()
	m.InitScale(2.0, 2.0)

	status := cairo.MatrixInvert(m)
	if status != cairo.StatusSuccess {
		t.Errorf("Matrix invert failed: %v", status)
	}

	if math.Abs(m.XX-0.5) > 0.0001 || math.Abs(m.YY-0.5) > 0.0001 {
		t.Errorf("Inverted matrix incorrect: XX=%f, YY=%f", m.XX, m.YY)
	}
}

// 测试矩阵乘法
func TestMatrixMultiply(t *testing.T) {
	m1 := cairo.NewMatrix()
	m1.InitScale(2.0, 2.0)

	m2 := cairo.NewMatrix()
	m2.InitTranslate(10.0, 20.0)

	result := cairo.NewMatrix()
	cairo.MatrixMultiply(result, m1, m2)

	// 验证结果
	if result.XX != 2.0 || result.YY != 2.0 {
		t.Errorf("Matrix multiply failed: XX=%f, YY=%f", result.XX, result.YY)
	}
}

// 测试 Status 字符串
func TestStatusString(t *testing.T) {
	tests := []struct {
		status   cairo.Status
		expected string
	}{
		{cairo.StatusSuccess, "success"},
		{cairo.StatusNoMemory, "no memory"},
		{cairo.StatusInvalidMatrix, "invalid matrix"},
		{cairo.StatusNullPointer, "null pointer"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("Status %v: expected '%s', got '%s'", tt.status, tt.expected, tt.status.String())
		}
	}
}
