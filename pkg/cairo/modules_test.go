package cairo

import (
	"image/color"
	"math"
	"testing"
)

// 测试 Pixman 图像操作
func TestPixmanImage(t *testing.T) {
	img := NewPixmanImage(PixmanFormatARGB32, 100, 100)

	// 测试设置和获取像素
	testColor := color.NRGBA{R: 255, G: 128, B: 64, A: 200}
	img.SetPixel(50, 50, testColor)

	result := img.GetPixel(50, 50)
	if result.R != testColor.R || result.G != testColor.G ||
		result.B != testColor.B || result.A != testColor.A {
		t.Errorf("Pixel mismatch: got %v, want %v", result, testColor)
	}
}

// 测试 Porter-Duff 混合
func TestPorterDuffBlend(t *testing.T) {
	src := color.NRGBA{R: 255, G: 0, B: 0, A: 128}
	dst := color.NRGBA{R: 0, G: 0, B: 255, A: 128}

	// 测试 Over 操作
	result := PorterDuffBlend(src, dst, OperatorOver)
	if result.A == 0 {
		t.Error("Blend result should not be transparent")
	}

	// 测试 Clear 操作
	result = PorterDuffBlend(src, dst, OperatorClear)
	if result.A != 0 {
		t.Error("Clear should produce transparent result")
	}
}

// 测试颜色空间转换
func TestColorSpaceConversion(t *testing.T) {
	// RGB to HSL and back
	r, g, b := 0.5, 0.3, 0.8
	h, s, l := rgbToHSL(r, g, b)
	r2, g2, b2 := hslToRGB(h, s, l)

	if math.Abs(r-r2) > 0.01 || math.Abs(g-g2) > 0.01 || math.Abs(b-b2) > 0.01 {
		t.Errorf("RGB->HSL->RGB conversion failed: (%f,%f,%f) -> (%f,%f,%f)",
			r, g, b, r2, g2, b2)
	}
}

// 测试 RGB to HSV 转换
func TestRGBToHSV(t *testing.T) {
	r, g, b := 1.0, 0.5, 0.0
	h, s, v := rgbToHSV(r, g, b)
	r2, g2, b2 := hsvToRGB(h, s, v)

	if math.Abs(r-r2) > 0.01 || math.Abs(g-g2) > 0.01 || math.Abs(b-b2) > 0.01 {
		t.Errorf("RGB->HSV->RGB conversion failed")
	}
}

// 测试 RGB to LAB 转换
func TestRGBToLAB(t *testing.T) {
	r, g, b := 0.5, 0.5, 0.5
	l, a, bVal := rgbToLAB(r, g, b)
	r2, g2, b2 := labToRGB(l, a, bVal)

	if math.Abs(r-r2) > 0.05 || math.Abs(g-g2) > 0.05 || math.Abs(b-b2) > 0.05 {
		t.Errorf("RGB->LAB->RGB conversion failed: (%f,%f,%f) -> (%f,%f,%f)",
			r, g, b, r2, g2, b2)
	}
}

// 测试高级光栅化器
func TestAdvancedRasterizer(t *testing.T) {
	rast := NewAdvancedRasterizer(100, 100)

	// 添加一个三角形
	rast.AddLine(10, 10, 90, 10)
	rast.AddLine(90, 10, 50, 90)
	rast.AddLine(50, 90, 10, 10)

	if len(rast.edges) != 3 {
		t.Errorf("Expected 3 edges, got %d", len(rast.edges))
	}
}

// 测试贝塞尔曲线细分
func TestBezierSubdivision(t *testing.T) {
	rast := NewAdvancedRasterizer(100, 100)

	// 添加三次贝塞尔曲线
	rast.AddCubicBezier(0, 0, 30, 60, 70, 60, 100, 0)

	// 应该生成多条边
	if len(rast.edges) < 2 {
		t.Error("Bezier curve should be subdivided into multiple edges")
	}
}

// 测试图像后端
func TestImageBackend(t *testing.T) {
	backend := NewImageBackend(100, 100)

	// 测试清空
	backend.Clear(color.White)
	c := backend.GetImage().At(50, 50)
	r, g, b, _ := c.RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Error("Clear failed")
	}

	// 测试填充矩形
	backend.FillRect(10, 10, 20, 20, color.Black)
	c = backend.GetImage().At(15, 15)
	r, g, b, _ = c.RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 0 {
		t.Error("FillRect failed")
	}
}

// 基准测试
func BenchmarkPorterDuffBlend(b *testing.B) {
	src := color.NRGBA{R: 255, G: 128, B: 64, A: 200}
	dst := color.NRGBA{R: 64, G: 128, B: 255, A: 200}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PorterDuffBlend(src, dst, OperatorOver)
	}
}

func BenchmarkColorSpaceConversion(b *testing.B) {
	r, g, bl := 0.5, 0.3, 0.8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h, s, l := rgbToHSL(r, g, bl)
		hslToRGB(h, s, l)
	}
}
