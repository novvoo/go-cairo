//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("🔍 调试文字渲染...")

	// Create surface and context
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// White background
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// Black text
	ctx.SetSourceRGB(0, 0, 0)

	// Create layout
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("Go Regular")
	fontDesc.SetSize(24)
	layout.SetFontDescription(fontDesc)
	layout.SetText("Test")

	// Move to position
	ctx.MoveTo(10, 50)

	// Get glyphs to debug
	fontFace := cairo.NewPangoCairoFont("Go Regular", cairo.FontSlantNormal, cairo.FontWeightNormal)
	defer fontFace.Destroy()

	fontMatrix := cairo.NewMatrix()
	fontMatrix.InitScale(24, 24)

	ctm := cairo.NewMatrix()
	ctm.InitIdentity()

	sf := cairo.NewPangoCairoScaledFont(fontFace, fontMatrix, ctm, nil)
	defer sf.Destroy()

	glyphs, _, _, status := sf.TextToGlyphs(10, 50, "Test")
	if status != cairo.StatusSuccess {
		fmt.Printf("❌ TextToGlyphs 失败: %v\n", status)
		return
	}

	fmt.Printf("✓ 获取到 %d 个字形\n", len(glyphs))
	for i, g := range glyphs {
		fmt.Printf("  字形 %d: Index=%d, X=%.2f, Y=%.2f\n", i, g.Index, g.X, g.Y)

		// Get glyph path
		path, err := sf.GlyphPath(g.Index)
		if err != nil {
			fmt.Printf("    ❌ 获取字形路径失败: %v\n", err)
			continue
		}

		if path == nil || len(path.Data) == 0 {
			fmt.Printf("    ⚠️  字形路径为空\n")
			continue
		}

		fmt.Printf("    ✓ 字形路径包含 %d 个段\n", len(path.Data))
		// Print first few path segments
		for j := 0; j < min(3, len(path.Data)); j++ {
			pd := path.Data[j]
			fmt.Printf("      段 %d: Type=%v, Points=%v\n", j, pd.Type, pd.Points)
		}
	}

	// Now render the text
	fmt.Println("\n📝 渲染文字...")
	ctx.PangoCairoShowText(layout)

	// Save
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("debug_text_render.png")
		if status != cairo.StatusSuccess {
			fmt.Printf("❌ 保存失败: %v\n", status)
			return
		}
		fmt.Println("✅ 已保存到 debug_text_render.png")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
