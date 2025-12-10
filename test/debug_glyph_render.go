//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("=== 字形渲染调试 (使用 PangoCairo) ===\n")

	// Create surface and context
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 400, 200)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// White background
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// Black text
	ctx.SetSourceRGB(0, 0, 0)
	fmt.Println("✓ 设置颜色: 黑色")

	// Draw a simple rectangle first to verify drawing works
	ctx.Rectangle(10, 10, 50, 30)
	ctx.Fill()
	fmt.Println("✓ 绘制测试矩形")

	// Use PangoCairo for text rendering
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)

	// Set font description with larger size for easier debugging
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("sans")
	fontDesc.SetWeight(cairo.PangoWeightNormal)
	fontDesc.SetSize(48)
	layout.SetFontDescription(fontDesc)

	// Render text "H"
	text := "H"
	layout.SetText(text)

	// Get text extents
	extents := layout.GetPixelExtents()
	fontExtents := layout.GetFontExtents()

	fmt.Printf("✓ 文本: \"%s\"\n", text)
	fmt.Printf("  文本范围: 宽度=%.2f, 高度=%.2f\n", extents.Width, extents.Height)
	fmt.Printf("  字体度量: Ascent=%.2f, Descent=%.2f, Height=%.2f\n",
		fontExtents.Ascent, fontExtents.Descent, fontExtents.Height)

	// Move to position and render
	x, y := 100.0, 100.0
	ctx.MoveTo(x, y)
	ctx.PangoCairoShowText(layout)

	fmt.Printf("✓ 使用 PangoCairo 渲染文本于位置 (%.2f, %.2f)\n", x, y)
	fmt.Printf("  文本边界框: [%.2f, %.2f] -> [%.2f, %.2f]\n",
		x+extents.X, y+extents.Y,
		x+extents.X+extents.Width, y+extents.Y+extents.Height)

	// Save
	wd, _ := os.Getwd()
	filename := filepath.Join(wd, "debug_glyph_render.png")

	if imageSurface, ok := surface.(cairo.ImageSurface); ok {
		status := imageSurface.WriteToPNG(filename)
		if status != cairo.StatusSuccess {
			log.Fatal("保存失败:", status)
		}
		fmt.Printf("\n✓ 图像已保存: %s\n", filename)
	}

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("注意: 现在使用 PangoCairo 进行文本渲染，避免了字形路径的复杂性")
}
