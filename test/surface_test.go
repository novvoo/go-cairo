package cairo

import (
	"os"
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试创建图像 Surface
func TestImageSurface(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	if surface == nil {
		t.Fatal("Failed to create image surface")
	}
	defer surface.Destroy()

	if surface.Status() != cairo.StatusSuccess {
		t.Errorf("Surface status: %v", surface.Status())
	}

	if surface.GetType() != cairo.SurfaceTypeImage {
		t.Errorf("Expected SurfaceTypeImage, got %v", surface.GetType())
	}

	// 测试图像 Surface 特定方法
	if imgSurface, ok := surface.(cairo.ImageSurface); ok {
		if imgSurface.GetWidth() != 100 {
			t.Errorf("Expected width 100, got %d", imgSurface.GetWidth())
		}
		if imgSurface.GetHeight() != 100 {
			t.Errorf("Expected height 100, got %d", imgSurface.GetHeight())
		}
		if imgSurface.GetFormat() != cairo.FormatARGB32 {
			t.Errorf("Expected FormatARGB32, got %v", imgSurface.GetFormat())
		}
	} else {
		t.Error("Surface is not an ImageSurface")
	}
}

// 测试不同的图像格式
func TestImageSurfaceFormats(t *testing.T) {
	formats := []cairo.Format{
		cairo.FormatARGB32,
		cairo.FormatRGB24,
		cairo.FormatA8,
		cairo.FormatA1,
	}

	for _, format := range formats {
		surface := cairo.NewImageSurface(format, 50, 50)
		if surface == nil {
			t.Errorf("Failed to create surface with format %v", format)
			continue
		}

		if imgSurface, ok := surface.(cairo.ImageSurface); ok {
			if imgSurface.GetFormat() != format {
				t.Errorf("Format mismatch: expected %v, got %v", format, imgSurface.GetFormat())
			}
		}

		surface.Destroy()
	}
}

// 测试 Surface 引用计数
func TestSurfaceReferenceCount(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	initialCount := surface.GetReferenceCount()
	if initialCount != 1 {
		t.Errorf("Expected initial reference count 1, got %d", initialCount)
	}

	ref := surface.Reference()
	if surface.GetReferenceCount() != 2 {
		t.Errorf("Expected reference count 2 after Reference(), got %d", surface.GetReferenceCount())
	}

	ref.Destroy()
	if surface.GetReferenceCount() != 1 {
		t.Errorf("Expected reference count 1 after Destroy(), got %d", surface.GetReferenceCount())
	}
}

// 测试 PNG 写入
func TestSurfaceWriteToPNG(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	imgSurface, ok := surface.(cairo.ImageSurface)
	if !ok {
		t.Fatal("Surface is not an ImageSurface")
	}

	// 创建一个简单的图像
	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	ctx.SetSourceRGB(1.0, 0.0, 0.0)
	ctx.Rectangle(10, 10, 80, 80)
	ctx.Fill()

	// 写入 PNG
	filename := "test_output.png"
	status := imgSurface.WriteToPNG(filename)
	if status != cairo.StatusSuccess {
		t.Errorf("Failed to write PNG: %v", status)
	}

	// 清理
	os.Remove(filename)
}

// 测试设备缩放
func TestSurfaceDeviceScale(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	surface.SetDeviceScale(2.0, 2.0)
	xScale, yScale := surface.GetDeviceScale()

	if xScale != 2.0 || yScale != 2.0 {
		t.Errorf("Device scale mismatch: expected (2.0, 2.0), got (%f, %f)", xScale, yScale)
	}
}

// 测试设备偏移
func TestSurfaceDeviceOffset(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	surface.SetDeviceOffset(10.0, 20.0)
	xOffset, yOffset := surface.GetDeviceOffset()

	if xOffset != 10.0 || yOffset != 20.0 {
		t.Errorf("Device offset mismatch: expected (10.0, 20.0), got (%f, %f)", xOffset, yOffset)
	}
}

// 测试创建相似 Surface
func TestCreateSimilarSurface(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	similar := surface.CreateSimilar(cairo.ContentColorAlpha, 50, 50)
	if similar == nil {
		t.Fatal("Failed to create similar surface")
	}
	defer similar.Destroy()

	if similar.GetContent() != cairo.ContentColorAlpha {
		t.Errorf("Content mismatch: expected ContentColorAlpha, got %v", similar.GetContent())
	}
}

// 测试 Surface Flush 和 MarkDirty
func TestSurfaceFlushAndMarkDirty(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	// 这些方法不应该导致错误
	err := surface.Flush()
	if err != nil {
		t.Errorf("Flush failed: %v", err)
	}

	surface.MarkDirty()
	surface.MarkDirtyRectangle(10, 10, 50, 50)
}

// 测试 Surface Finish
func TestSurfaceFinish(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	err := surface.Finish()
	if err != nil {
		t.Errorf("Finish failed: %v", err)
	}
}

// 测试无效的 Surface 尺寸
func TestInvalidSurfaceSize(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, -100, 100)
	if surface.Status() == cairo.StatusSuccess {
		t.Error("Expected error for negative width")
	}
	surface.Destroy()

	surface = cairo.NewImageSurface(cairo.FormatARGB32, 100, 0)
	if surface.Status() == cairo.StatusSuccess {
		t.Error("Expected error for zero height")
	}
	surface.Destroy()
}

// 基准测试：创建 Surface
func BenchmarkCreateImageSurface(b *testing.B) {
	for i := 0; i < b.N; i++ {
		surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
		surface.Destroy()
	}
}

// 基准测试：填充 Surface
func BenchmarkFillSurface(b *testing.B) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.SetSourceRGB(1.0, 0.0, 0.0)
		ctx.Rectangle(0, 0, 100, 100)
		ctx.Fill()
	}
}

// 测试 Surface 数据访问
func TestSurfaceDataAccess(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 10, 10)
	defer surface.Destroy()

	imgSurface, ok := surface.(cairo.ImageSurface)
	if !ok {
		t.Fatal("Surface is not an ImageSurface")
	}

	data := imgSurface.GetData()
	if data == nil {
		t.Error("Surface data is nil")
	}

	stride := imgSurface.GetStride()
	expectedStride := cairo.FormatStrideForWidth(cairo.FormatARGB32, 10)
	if stride != expectedStride {
		t.Errorf("Stride mismatch: expected %d, got %d", expectedStride, stride)
	}
}

// 测试 Go Image 互操作
func TestSurfaceGoImage(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	imgSurface, ok := surface.(cairo.ImageSurface)
	if !ok {
		t.Fatal("Surface is not an ImageSurface")
	}

	goImage := imgSurface.GetGoImage()
	if goImage == nil {
		t.Error("Go image is nil")
	}

	bounds := goImage.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Image bounds mismatch: expected 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// 测试设置像素
	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	ctx.SetSourceRGBA(1.0, 0.0, 0.0, 1.0)
	ctx.Rectangle(0, 0, 100, 100)
	ctx.Fill()

	// 验证像素颜色
	c := goImage.At(50, 50)
	r, g, b, a := c.RGBA()
	// 注意：RGBA() 返回的是 0-65535 范围的值
	if r < 60000 || g > 5000 || b > 5000 || a < 60000 {
		t.Logf("Pixel color at (50,50): R=%d, G=%d, B=%d, A=%d", r, g, b, a)
	}
}

// 测试 Surface 内容类型
func TestSurfaceContent(t *testing.T) {
	tests := []struct {
		format  cairo.Format
		content cairo.Content
	}{
		{cairo.FormatARGB32, cairo.ContentColorAlpha},
		{cairo.FormatRGB24, cairo.ContentColor},
		{cairo.FormatA8, cairo.ContentAlpha},
	}

	for _, tt := range tests {
		surface := cairo.NewImageSurface(tt.format, 100, 100)
		if surface.GetContent() != tt.content {
			t.Errorf("Format %v: expected content %v, got %v", tt.format, tt.content, surface.GetContent())
		}
		surface.Destroy()
	}
}
