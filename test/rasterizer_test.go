package cairo

import (
	"image"
	"image/color"
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试创建光栅化器
func TestAdvancedRasterizerCreation(t *testing.T) {
	rast := cairo.NewAdvancedRasterizer(100, 100)
	if rast == nil {
		t.Fatal("Failed to create rasterizer")
	}
}

// 测试添加直线
func TestRasterizerAddLine(t *testing.T) {
	rast := cairo.NewAdvancedRasterizer(100, 100)

	// 添加一条直线
	rast.AddLine(10, 10, 90, 90)

	// 光栅化到图像
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	rast.Rasterize(img, color.Black, cairo.FillRuleWinding)
}

// 测试添加三角形
func TestRasterizerTriangle(t *testing.T) {
	rast := cairo.NewAdvancedRasterizer(100, 100)

	// 添加三角形的三条边
	rast.AddLine(50, 10, 90, 90)
	rast.AddLine(90, 90, 10, 90)
	rast.AddLine(10, 90, 50, 10)

	// 光栅化
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	rast.Rasterize(img, color.RGBA{R: 255, G: 0, B: 0, A: 255}, cairo.FillRuleWinding)
}

// 测试二次贝塞尔曲线
func TestRasterizerQuadraticBezier(t *testing.T) {
	rast := cairo.NewAdvancedRasterizer(100, 100)

	// 添加二次贝塞尔曲线
	rast.AddQuadraticBezier(10, 50, 50, 10, 90, 50)

	// 光栅化
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	rast.Rasterize(img, color.Black, cairo.FillRuleWinding)
}

// 测试三次贝塞尔曲线
func TestRasterizerCubicBezier(t *testing.T) {
	rast := cairo.NewAdvancedRasterizer(100, 100)

	// 添加三次贝塞尔曲线
	rast.AddCubicBezier(10, 50, 30, 10, 70, 90, 90, 50)

	// 光栅化
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	rast.Rasterize(img, color.Black, cairo.FillRuleWinding)
}

// 测试重置光栅化器
func TestRasterizerReset(t *testing.T) {
	rast := cairo.NewAdvancedRasterizer(100, 100)

	// 添加一些路径
	rast.AddLine(10, 10, 90, 90)
	rast.AddLine(90, 10, 10, 90)

	// 重置
	rast.Reset()

	// 添加新路径
	rast.AddLine(50, 10, 50, 90)

	// 光栅化应该只显示新路径
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	rast.Rasterize(img, color.Black, cairo.FillRuleWinding)
}

// 基准测试：光栅化直线
func BenchmarkRasterizeLine(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rast := cairo.NewAdvancedRasterizer(100, 100)
		rast.AddLine(10, 10, 90, 90)
		rast.Rasterize(img, color.Black, cairo.FillRuleWinding)
	}
}

// 基准测试：光栅化贝塞尔曲线
func BenchmarkRasterizeBezier(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rast := cairo.NewAdvancedRasterizer(100, 100)
		rast.AddCubicBezier(10, 50, 30, 10, 70, 90, 90, 50)
		rast.Rasterize(img, color.Black, cairo.FillRuleWinding)
	}
}
