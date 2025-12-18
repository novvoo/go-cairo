//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	// Create output directory if it doesn't exist

	// Test 1: Simple text analysis
	testSimpleText()

	// Test 2: Text with potential collisions
	testTextWithCollisions()

}

func testSimpleText() {
	fmt.Println("=== 测试1: 简单文本分析 ===")

	// Create a new image surface
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 500, 300)
	defer surface.Destroy()

	// Create a context
	context := cairo.NewContext(surface)
	defer context.Destroy()

	// Set background color
	context.SetSourceRGB(1, 1, 1) // White background
	context.Paint()

	// Set text color
	context.SetSourceRGB(0, 0, 0) // Black text

	// 使用PangoCairo创建布局
	layout := context.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)

	// 设置字体描述
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("sans")
	fontDesc.SetWeight(cairo.PangoWeightNormal)
	fontDesc.SetSize(32.0)
	layout.SetFontDescription(fontDesc)

	// 设置文本
	text := "Hello World 你好世界!"
	layout.SetText(text)

	// Move to position and show text
	context.MoveTo(50, 100)
	context.PangoCairoShowText(layout)

	// 获取文本范围信息
	extents := layout.GetPixelExtents()
	fmt.Println("字体类型检查:")
	fmt.Printf("  使用 PangoCairo 渲染\n")

	// 打印文本范围信息
	fmt.Printf("  文本范围: 宽度=%.2f, 高度=%.2f\n", extents.Width, extents.Height)
	fmt.Printf("  边界信息: X=%.2f, Y=%.2f\n", extents.X, extents.Y)
	fmt.Println()

	// Save to PNG
	if imageSurface, ok := surface.(cairo.ImageSurface); ok {
		status := imageSurface.WriteToPNG("glyph_simple.png")
		if status != cairo.StatusSuccess {
			log.Fatal("Failed to save PNG:", status)
		}
	} else {
		log.Fatal("Surface is not an ImageSurface")
	}

	fmt.Println("简单文本测试结果保存到 example/glyph_simple.png")
	fmt.Println()
}

func testTextWithCollisions() {
	fmt.Println("=== 测试2: 字符重叠检测 ===")

	// Create a new image surface
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 500, 300)
	defer surface.Destroy()

	// Create a context
	context := cairo.NewContext(surface)
	defer context.Destroy()

	// Set background color
	context.SetSourceRGB(1, 1, 1) // White background
	context.Paint()

	// Set text color
	context.SetSourceRGB(0, 0, 0) // Black text

	// 使用PangoCairo创建布局
	layout := context.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)

	// 设置字体描述
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("monospace")
	fontDesc.SetWeight(cairo.PangoWeightBold)
	fontDesc.SetSize(48.0)
	layout.SetFontDescription(fontDesc)

	// Show normal text at top
	layout.SetText("Test")
	context.MoveTo(50, 100)
	context.PangoCairoShowText(layout)

	// Show the same text with manual positioning to demonstrate potential overlap
	layout.SetText("T")
	context.MoveTo(50, 200)
	context.PangoCairoShowText(layout)

	layout.SetText("e")
	context.MoveTo(70, 200) // Very close positioning to show potential overlap
	context.PangoCairoShowText(layout)

	layout.SetText("s")
	context.MoveTo(90, 200)
	context.PangoCairoShowText(layout)

	layout.SetText("t")
	context.MoveTo(110, 200)
	context.PangoCairoShowText(layout)

	// Save to PNG
	if imageSurface, ok := surface.(cairo.ImageSurface); ok {
		status := imageSurface.WriteToPNG("./glyph_collision.png")
		if status != cairo.StatusSuccess {
			log.Fatal("Failed to save PNG:", status)
		}
	} else {
		log.Fatal("Surface is not an ImageSurface")
	}

	fmt.Println("字符重叠测试结果保存到 /glyph_collision.png")
	fmt.Println("注意: 在第二行中，字符被故意放置得很近以演示潜在的重叠情况")
	fmt.Println()
}
