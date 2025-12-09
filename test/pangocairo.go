package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	// Create a new image surface
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 400, 200)
	defer surface.Destroy()

	// Create a context
	context := cairo.NewContext(surface)
	defer context.Destroy()

	// Set background color
	context.SetSourceRGB(1, 1, 1) // White background
	context.Paint()

	// Set text color
	context.SetSourceRGB(0, 0, 0) // Black text

	// Simple text rendering
	context.SelectFontFace("sans", cairo.FontSlantNormal, cairo.FontWeightNormal)
	context.SetFontSize(24)

	// Move to position and show text
	context.MoveTo(50, 100)

	// Get the scaled font from context for our new functionality
	sf := context.GetScaledFont()
	if sf == nil {
		log.Fatal("Failed to get scaled font")
	}
	defer sf.Destroy()

	// 打印实际的类型信息
	fmt.Printf("ScaledFont 实际类型: %T\n", sf)
	fmt.Printf("ScaledFont 反射类型: %s\n", reflect.TypeOf(sf).String())

	// 检查是否是 PangoCairoScaledFont 类型
	if pangoFont, ok := sf.(*cairo.PangoCairoScaledFont); ok {
		fmt.Println("成功断言为 PangoCairoScaledFont")
		printGlyphInfo(context, pangoFont, "Hello, Cairo!")
	} else {
		// 对于其他类型的 ScaledFont，我们只需要检查它是否实现了接口
		// 由于所有实现都满足 cairo.ScaledFont 接口，我们可以直接使用
		fmt.Println("是标准 ScaledFont 类型，使用通用方法")
		printGenericGlyphInfo(context, sf, "Hello, Cairo!")
	}

	// Save to PNG
	if imageSurface, ok := surface.(cairo.ImageSurface); ok {
		// 确保目录存在
		status := imageSurface.WriteToPNG("pangocairo.png")
		if status != cairo.StatusSuccess {
			log.Fatal("Failed to save PNG:", status)
		}
	} else {
		log.Fatal("Surface is not an ImageSurface")
	}

	fmt.Println("Simple test saved to pangocairo.png")
}

func printGlyphInfo(context cairo.Context, pangoFont *cairo.PangoCairoScaledFont, text string) {
	// Get current point BEFORE showing text
	x, y := context.GetCurrentPoint()

	// Show text and get glyphs for analysis
	context.ShowText(text)

	// Get glyphs for the text
	glyphs, _, _, status := pangoFont.TextToGlyphs(x, y, text)
	if status != cairo.StatusSuccess {
		log.Fatal("Failed to get glyphs:", status)
	}

	// Print information about each glyph
	fmt.Println("=== 字符坐标和碰撞检测信息 ===")
	pangoFont.PrintTextGlyphsInfo(text, glyphs)

	// 额外打印每个字母的详细位置信息和冲突检测
	fmt.Println("=== 每个字母的详细位置信息 ===")
	runes := []rune(text)
	for i, glyph := range glyphs {
		var char rune
		if i < len(runes) {
			char = runes[i]
		} else {
			char = rune(glyph.Index)
		}

		// 获取字形的角落坐标
		coords, status := pangoFont.GetGlyphCornerCoordinates(glyph)
		if status != cairo.StatusSuccess {
			fmt.Printf("无法获取字符 '%c' 的坐标信息: %v\n", char, status)
			continue
		}

		// 打印字形的详细位置信息
		fmt.Printf("字符 '%c':\n", char)
		fmt.Printf("  位置: (%.2f, %.2f)\n", glyph.X, glyph.Y)
		fmt.Printf("  左上角: (%.2f, %.2f)\n", coords.TopLeftX, coords.TopLeftY)
		fmt.Printf("  右上角: (%.2f, %.2f)\n", coords.TopRightX, coords.TopRightY)
		fmt.Printf("  左下角: (%.2f, %.2f)\n", coords.BottomLeftX, coords.BottomLeftY)
		fmt.Printf("  右下角: (%.2f, %.2f)\n", coords.BottomRightX, coords.BottomRightY)

		// 检查与其他字符的冲突
		hasCollision := false
		for j := i + 1; j < len(glyphs); j++ {
			collides, status := pangoFont.CheckGlyphCollision(glyph, glyphs[j])
			if status == cairo.StatusSuccess && collides {
				var nextChar rune
				if j < len(runes) {
					nextChar = runes[j]
				} else {
					nextChar = rune(glyphs[j].Index)
				}
				fmt.Printf("  警告: 与字符 '%c' 发生重叠!\n", nextChar)
				hasCollision = true
			}
		}

		if !hasCollision {
			fmt.Printf("  无重叠冲突\n")
		}
		fmt.Println()
	}
}

func printGenericGlyphInfo(context cairo.Context, sf cairo.ScaledFont, text string) {
	// Get current point BEFORE showing text
	x, y := context.GetCurrentPoint()

	// Show text and get glyphs for analysis
	context.ShowText(text)

	// Get glyphs for the text
	glyphs, _, _, status := sf.TextToGlyphs(x, y, text)
	if status != cairo.StatusSuccess {
		log.Fatal("Failed to get glyphs:", status)
	}

	// 打印每个字母的详细位置信息
	fmt.Println("=== 每个字母的详细位置信息 ===")
	runes := []rune(text)
	for i, glyph := range glyphs {
		var char rune
		if i < len(runes) {
			char = runes[i]
		} else {
			char = rune(glyph.Index)
		}

		// 使用 TextExtents 获取文本范围信息
		extents := sf.TextExtents(string(char))

		// 计算四个角的坐标（基于字形位置和文本范围）
		topLeftX := glyph.X + extents.XBearing
		topLeftY := glyph.Y + extents.YBearing
		topRightX := glyph.X + extents.XBearing + extents.Width
		topRightY := glyph.Y + extents.YBearing
		bottomLeftX := glyph.X + extents.XBearing
		bottomLeftY := glyph.Y + extents.YBearing + extents.Height
		bottomRightX := glyph.X + extents.XBearing + extents.Width
		bottomRightY := glyph.Y + extents.YBearing + extents.Height

		fmt.Printf("字符 '%c':\n", char)
		fmt.Printf("  位置: (%.2f, %.2f)\n", glyph.X, glyph.Y)
		fmt.Printf("  左上角: (%.2f, %.2f)\n", topLeftX, topLeftY)
		fmt.Printf("  右上角: (%.2f, %.2f)\n", topRightX, topRightY)
		fmt.Printf("  左下角: (%.2f, %.2f)\n", bottomLeftX, bottomLeftY)
		fmt.Printf("  右下角: (%.2f, %.2f)\n", bottomRightX, bottomRightY)

		// 检查与其他字符的冲突（简化版）
		hasCollision := false
		for j := i + 1; j < len(glyphs); j++ {
			// 简单的边界框重叠检测
			otherGlyph := glyphs[j]
			var otherChar rune
			if j < len(runes) {
				otherChar = runes[j]
			} else {
				otherChar = rune(otherGlyph.Index)
			}

			// 获取另一个字符的范围信息
			otherExtents := sf.TextExtents(string(otherChar))

			// 计算当前字符的边界框
			currentLeft := glyph.X + extents.XBearing
			currentRight := glyph.X + extents.XBearing + extents.Width
			currentTop := glyph.Y + extents.YBearing
			currentBottom := glyph.Y + extents.YBearing + extents.Height

			// 计算另一个字符的边界框
			otherLeft := otherGlyph.X + otherExtents.XBearing
			otherRight := otherGlyph.X + otherExtents.XBearing + otherExtents.Width
			otherTop := otherGlyph.Y + otherExtents.YBearing
			otherBottom := otherGlyph.Y + otherExtents.YBearing + otherExtents.Height

			// 检查边界框是否重叠（增加一个小的容差值）
			tolerance := 1.0
			if currentLeft < otherRight-tolerance && currentRight-tolerance > otherLeft &&
				currentTop < otherBottom-tolerance && currentBottom-tolerance > otherTop {
				fmt.Printf("  警告: 与字符 '%c' 发生重叠!\n", otherChar)
				hasCollision = true
			}
		}

		if !hasCollision {
			fmt.Printf("  无重叠冲突\n")
		}
		fmt.Println()
	}
}
