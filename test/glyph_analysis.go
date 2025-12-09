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

	fmt.Println("✅ 所有测试已完成，结果保存在 example/")
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

	// Set font properties
	context.SelectFontFace("sans", cairo.FontSlantNormal, cairo.FontWeightNormal)
	context.SetFontSize(32)

	// Move to position and show text
	context.MoveTo(50, 100)

	// Show text
	text := "Hello World"
	context.ShowText(text)

	// Get the scaled font from context for our new functionality
	sf := context.GetScaledFont()
	if sf == nil {
		log.Fatal("Failed to get scaled font")
	}
	defer sf.Destroy()

	// Try to cast to PangoCairoScaledFont, but handle the case where it's not
	fmt.Println("字体类型检查:")
	fmt.Printf("  字体类型: %v\n", sf.GetType())

	// Always print information about text extents
	extents := context.TextExtents(text)
	fmt.Printf("  文本范围: 宽度=%.2f, 高度=%.2f\n", extents.Width, extents.Height)
	fmt.Printf("  边界信息: XBearing=%.2f, YBearing=%.2f\n", extents.XBearing, extents.YBearing)
	fmt.Printf("  前进距离: X=%.2f, Y=%.2f\n", extents.XAdvance, extents.YAdvance)
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

	// Set font properties for clearer collision demonstration
	context.SelectFontFace("monospace", cairo.FontSlantNormal, cairo.FontWeightBold)
	context.SetFontSize(48)

	// Show normal text at top
	context.MoveTo(50, 100)
	context.ShowText("Test")

	// Show the same text with manual positioning to demonstrate potential overlap
	context.MoveTo(50, 200)
	context.ShowText("T")
	context.MoveTo(70, 200) // Very close positioning to show potential overlap
	context.ShowText("e")
	context.MoveTo(90, 200)
	context.ShowText("s")
	context.MoveTo(110, 200)
	context.ShowText("t")

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
