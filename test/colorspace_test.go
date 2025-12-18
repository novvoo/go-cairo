package cairo

import (
	"math"
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试 RGB 到 HSL 转换
func TestRGBToHSL(t *testing.T) {
	r, g, b := 1.0, 0.5, 0.0
	h, s, l := cairo.RgbToHSL(r, g, b)

	// 验证转换结果在合理范围内
	if h < 0 || h > 1 {
		t.Errorf("Hue out of range: %f", h)
	}
	if s < 0 || s > 1 {
		t.Errorf("Saturation out of range: %f", s)
	}
	if l < 0 || l > 1 {
		t.Errorf("Lightness out of range: %f", l)
	}
}

// 测试 HSL 到 RGB 转换
func TestHSLToRGB(t *testing.T) {
	h, s, l := 0.5, 0.8, 0.6
	r, g, b := cairo.HslToRGB(h, s, l)

	// 验证转换结果在合理范围内
	if r < 0 || r > 1 {
		t.Errorf("Red out of range: %f", r)
	}
	if g < 0 || g > 1 {
		t.Errorf("Green out of range: %f", g)
	}
	if b < 0 || b > 1 {
		t.Errorf("Blue out of range: %f", b)
	}
}

// 测试 RGB -> HSL -> RGB 往返转换
func TestRGBHSLRoundTrip(t *testing.T) {
	testCases := []struct {
		r, g, b float64
	}{
		{1.0, 0.0, 0.0},  // 纯红
		{0.0, 1.0, 0.0},  // 纯绿
		{0.0, 0.0, 1.0},  // 纯蓝
		{1.0, 1.0, 1.0},  // 白色
		{0.0, 0.0, 0.0},  // 黑色
		{0.5, 0.5, 0.5},  // 灰色
		{1.0, 0.5, 0.25}, // 橙色
	}

	for _, tc := range testCases {
		h, s, l := cairo.RgbToHSL(tc.r, tc.g, tc.b)
		r2, g2, b2 := cairo.HslToRGB(h, s, l)

		if math.Abs(tc.r-r2) > 0.01 || math.Abs(tc.g-g2) > 0.01 || math.Abs(tc.b-b2) > 0.01 {
			t.Errorf("RGB->HSL->RGB roundtrip failed for (%f,%f,%f): got (%f,%f,%f)",
				tc.r, tc.g, tc.b, r2, g2, b2)
		}
	}
}

// 测试灰度颜色的 HSL 转换
func TestGrayscaleHSL(t *testing.T) {
	// 灰度颜色的饱和度应该为 0
	r, g, b := 0.5, 0.5, 0.5
	_, s, l := cairo.RgbToHSL(r, g, b)

	if s != 0.0 {
		t.Errorf("Grayscale should have saturation 0, got %f", s)
	}

	if math.Abs(l-0.5) > 0.01 {
		t.Errorf("Grayscale lightness should be 0.5, got %f", l)
	}
}

// 基准测试：RGB 到 HSL
func BenchmarkRGBToHSL(b *testing.B) {
	r, g, bl := 0.5, 0.3, 0.8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cairo.RgbToHSL(r, g, bl)
	}
}

// 基准测试：HSL 到 RGB
func BenchmarkHSLToRGB(b *testing.B) {
	h, s, l := 0.5, 0.8, 0.6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cairo.HslToRGB(h, s, l)
	}
}

// 基准测试：RGB->HSL->RGB 往返
func BenchmarkColorSpaceRoundTrip(b *testing.B) {
	r, g, bl := 0.5, 0.3, 0.8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h, s, l := cairo.RgbToHSL(r, g, bl)
		cairo.HslToRGB(h, s, l)
	}
}
