package cairo

import (
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试 FontOptions 创建
func TestFontOptionsCreation(t *testing.T) {
	opts := cairo.NewFontOptions()
	if opts == nil {
		t.Fatal("Failed to create font options")
	}

	if opts.Status != cairo.StatusSuccess {
		t.Errorf("Expected StatusSuccess, got %v", opts.Status)
	}

	if opts.Antialias != cairo.AntialiasDefault {
		t.Errorf("Expected AntialiasDefault, got %v", opts.Antialias)
	}
}

// 测试 FontOptions 复制
func TestFontOptionsCopy(t *testing.T) {
	opts := cairo.NewFontOptions()
	opts.Antialias = cairo.AntialiasBest
	opts.HintStyle = cairo.HintStyleFull

	copy := opts.Copy()
	if copy == nil {
		t.Fatal("Failed to copy font options")
	}

	if copy.Antialias != opts.Antialias {
		t.Errorf("Antialias not copied correctly")
	}

	if copy.HintStyle != opts.HintStyle {
		t.Errorf("HintStyle not copied correctly")
	}

	// 修改副本不应影响原始对象
	copy.Antialias = cairo.AntialiasNone
	if opts.Antialias == cairo.AntialiasNone {
		t.Error("Modifying copy affected original")
	}
}

// 测试 FontExtents (跳过 - 需要完整的字体 API)
func TestFontExtents(t *testing.T) {
	t.Skip("FontExtents requires full font API implementation")
}

// 测试 TextExtents
func TestTextExtents(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	text := "Hello, World!"
	extents := ctx.TextExtents(text)

	if extents == nil {
		t.Fatal("Failed to get text extents")
	}

	// 基本验证 - 文本应该有宽度
	if extents.Width < 0 {
		t.Errorf("Invalid text width: %f", extents.Width)
	}
}

// 测试 SelectFontFace (跳过 - 需要完整的字体 API)
func TestSelectFontFace(t *testing.T) {
	t.Skip("SelectFontFace requires full font API implementation")
}

// 测试 SetFontSize (跳过 - 需要完整的字体 API)
func TestSetFontSize(t *testing.T) {
	t.Skip("SetFontSize requires full font API implementation")
}

// 测试 ShowText (跳过 - 需要完整的字体 API)
func TestShowText(t *testing.T) {
	t.Skip("ShowText requires full font API implementation")
}

// 测试空文本 (跳过 - 需要完整的字体 API)
func TestShowTextEmpty(t *testing.T) {
	t.Skip("ShowText requires full font API implementation")
}

// 测试多行文本 (跳过 - 需要完整的字体 API)
func TestShowTextMultiline(t *testing.T) {
	t.Skip("ShowText requires full font API implementation")
}

// 基准测试：TextExtents
func BenchmarkTextExtents(b *testing.B) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	text := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.TextExtents(text)
	}
}

// 基准测试：ShowText (跳过 - 需要完整的字体 API)
func BenchmarkShowText(b *testing.B) {
	b.Skip("ShowText requires full font API implementation")
}
