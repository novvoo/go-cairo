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
	fmt.Println("=== PangoCairo 测试开始 ===\n")

	// Create a new image surface
	width, height := 400, 200
	surface := cairo.NewImageSurface(cairo.FormatARGB32, width, height)
	defer surface.Destroy()
	fmt.Printf("✓ 创建图像表面: %dx%d\n", width, height)

	// Create a context
	context := cairo.NewContext(surface)
	defer context.Destroy()
	fmt.Printf("✓ 创建 Cairo 上下文\n")

	// Set background color to light blue for better visibility
	context.SetSourceRGB(0.9, 0.95, 1.0) // Light blue background
	context.Paint()
	fmt.Printf("✓ 设置背景颜色: 浅蓝色 (0.9, 0.95, 1.0)\n")

	// Set text color to dark blue
	context.SetSourceRGB(0.0, 0.0, 0.5) // Dark blue text
	fmt.Printf("✓ 设置文字颜色: 深蓝色 (0.0, 0.0, 0.5)\n\n")

	// Create a PangoCairo font directly to get the proper exported ScaledFont type
	fontFamily := "sans"
	fontSize := 24.0
	pangoFont := cairo.NewPangoCairoFont(fontFamily, cairo.FontSlantNormal, cairo.FontWeightNormal)
	defer pangoFont.Destroy()
	fmt.Printf("✓ 创建 PangoCairo 字体: %s\n", fontFamily)

	// Create a font matrix for scaling
	// Use positive Y scale - the context already has Y-flip applied
	fontMatrix := cairo.NewMatrix()
	fontMatrix.InitScale(fontSize, fontSize)
	fmt.Printf("✓ 设置字体大小: %.1f\n", fontSize)

	// Create CTM (Current Transformation Matrix)
	ctm := cairo.NewMatrix()
	ctm.InitIdentity()

	// Create font options
	fontOptions := cairo.NewFontOptions()

	// Create a PangoCairo scaled font (this will be the exported type)
	scaledFont := cairo.NewPangoCairoScaledFont(pangoFont, fontMatrix, ctm, fontOptions)
	defer scaledFont.Destroy()

	// Set the scaled font on the context
	context.SetScaledFont(scaledFont)

	// Now we have a proper exported PangoCairoScaledFont
	fmt.Printf("✓ ScaledFont 类型: %T\n", scaledFont)
	fmt.Printf("✓ 成功创建 PangoCairoScaledFont\n\n")

	// === 文字翻转诊断 ===
	printTextOrientationDiagnostics(context, fontMatrix, scaledFont)

	// Move to position and show text
	startX, startY := 50.0, 100.0
	context.MoveTo(startX, startY)
	fmt.Printf("=== 文字渲染 ===\n")
	fmt.Printf("起始位置: (%.1f, %.1f)\n", startX, startY)

	// Get current point BEFORE showing text
	x, y := context.GetCurrentPoint()
	fmt.Printf("当前点: (%.1f, %.1f)\n", x, y)

	// Show text and get glyphs for analysis using PangoCairo
	text := "Hello, Cairo!"
	fmt.Printf("渲染文本: \"%s\"\n\n", text)

	layout := context.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily(fontFamily)
	fontDesc.SetSize(fontSize)
	layout.SetFontDescription(fontDesc)
	layout.SetText(text)

	// Show the text
	context.PangoCairoShowText(layout)
	fmt.Printf("✓ 文字已渲染到画布\n\n")

	// Verify rendering by checking if pixels were modified
	if imageSurface, ok := surface.(cairo.ImageSurface); ok {
		data := imageSurface.GetData()
		hasNonBackground := false
		// Check if any pixels are not the background color
		for i := 0; i < len(data); i += 4 {
			b, g, r := data[i], data[i+1], data[i+2]
			// Check if pixel is not background color (allowing some tolerance)
			if r < 220 || g < 235 || b < 250 {
				hasNonBackground = true
				break
			}
		}
		if hasNonBackground {
			fmt.Printf("✓ 验证成功: 检测到文字像素已渲染\n\n")
		} else {
			fmt.Printf("⚠ 警告: 未检测到文字像素，可能渲染失败\n\n")
		}
	}

	// Print glyph information using the built-in method
	printGlyphInformation(context, scaledFont, text, x, y)

	// Save to PNG with absolute path
	fmt.Println("=== 保存图像 ===")
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("获取工作目录失败:", err)
	}

	filename := filepath.Join(wd, "pangocairo.png")
	fmt.Printf("保存路径: %s\n", filename)

	if imageSurface, ok := surface.(cairo.ImageSurface); ok {
		status := imageSurface.WriteToPNG(filename)
		if status != cairo.StatusSuccess {
			log.Fatal("保存 PNG 失败:", status)
		}
		fmt.Printf("✓ 图像已成功保存\n")
	} else {
		log.Fatal("Surface 不是 ImageSurface 类型")
	}

	fmt.Println("\n=== PangoCairo 测试完成 ===")
	fmt.Printf("请查看生成的图像: pangocairo.png\n")
}

// printTextOrientationDiagnostics 打印文字方向和翻转的详细诊断信息
func printTextOrientationDiagnostics(context cairo.Context, fontMatrix *cairo.Matrix, scaledFont cairo.ScaledFont) {
	fmt.Println("=== 文字方向诊断 ===")

	// 1. 获取上下文变换矩阵
	ctxMatrix := context.GetMatrix()
	fmt.Println("\n【上下文变换矩阵】")
	fmt.Printf("  XX (X轴缩放): %8.4f\n", ctxMatrix.XX)
	fmt.Printf("  YX (X轴倾斜): %8.4f\n", ctxMatrix.YX)
	fmt.Printf("  XY (Y轴倾斜): %8.4f\n", ctxMatrix.XY)
	fmt.Printf("  YY (Y轴缩放): %8.4f\n", ctxMatrix.YY)
	fmt.Printf("  X0 (X平移):   %8.4f\n", ctxMatrix.X0)
	fmt.Printf("  Y0 (Y平移):   %8.4f\n", ctxMatrix.Y0)

	// 2. 字体矩阵信息
	fmt.Println("\n【字体矩阵】")
	fmt.Printf("  XX (X轴缩放): %8.4f\n", fontMatrix.XX)
	fmt.Printf("  YX (X轴倾斜): %8.4f\n", fontMatrix.YX)
	fmt.Printf("  XY (Y轴倾斜): %8.4f\n", fontMatrix.XY)
	fmt.Printf("  YY (Y轴缩放): %8.4f\n", fontMatrix.YY)
	fmt.Printf("  X0 (X平移):   %8.4f\n", fontMatrix.X0)
	fmt.Printf("  Y0 (Y平移):   %8.4f\n", fontMatrix.Y0)

	// 3. 分析坐标系状态
	fmt.Println("\n【坐标系状态分析】")
	ctxFlippedX := ctxMatrix.XX < 0
	ctxFlippedY := ctxMatrix.YY < 0
	fontFlippedX := fontMatrix.XX < 0
	fontFlippedY := fontMatrix.YY < 0

	if ctxFlippedX {
		fmt.Printf("  ❌ 上下文X轴翻转: 是 (XX=%.4f < 0)\n", ctxMatrix.XX)
	} else {
		fmt.Printf("  ✅ 上下文X轴翻转: 否 (XX=%.4f >= 0)\n", ctxMatrix.XX)
	}

	if ctxFlippedY {
		fmt.Printf("  ⚠️  上下文Y轴翻转: 是 (YY=%.4f < 0)\n", ctxMatrix.YY)
		fmt.Printf("      → 这是Cairo的标准行为，用于匹配图像坐标系\n")
	} else {
		fmt.Printf("  ✅ 上下文Y轴翻转: 否 (YY=%.4f >= 0)\n", ctxMatrix.YY)
	}

	if fontFlippedX {
		fmt.Printf("  ❌ 字体X轴翻转: 是 (XX=%.4f < 0)\n", fontMatrix.XX)
	} else {
		fmt.Printf("  ✅ 字体X轴翻转: 否 (XX=%.4f >= 0)\n", fontMatrix.XX)
	}

	if fontFlippedY {
		fmt.Printf("  ❌ 字体Y轴翻转: 是 (YY=%.4f < 0)\n", fontMatrix.YY)
		fmt.Printf("      → 这会导致字形上下颠倒！\n")
	} else {
		fmt.Printf("  ✅ 字体Y轴翻转: 否 (YY=%.4f >= 0)\n", fontMatrix.YY)
		fmt.Printf("      → 正确：字体矩阵使用正Y缩放，配合上下文的Y翻转\n")
	}

	// 4. 获取字体度量
	fontExtents := scaledFont.Extents()
	fmt.Println("\n【字体度量信息】")
	fmt.Printf("  Ascent (上升高度):  %.2f\n", fontExtents.Ascent)
	fmt.Printf("  Descent (下降高度): %.2f\n", fontExtents.Descent)
	fmt.Printf("  Height (总高度):    %.2f\n", fontExtents.Height)
	fmt.Printf("  LineGap (行间距):   %.2f\n", fontExtents.LineGap)

	// 5. 综合诊断
	fmt.Println("\n【综合诊断结果】")
	hasIssue := false

	// 检查是否有问题的配置
	if fontFlippedY && ctxFlippedY {
		fmt.Println("  ❌ 错误配置: 字体和上下文都进行了Y轴翻转")
		fmt.Println("     → 这会导致双重翻转，文字显示正常但逻辑错误")
		hasIssue = true
	} else if fontFlippedY && !ctxFlippedY {
		fmt.Println("  ❌ 错误配置: 只有字体进行了Y轴翻转")
		fmt.Println("     → 这会导致文字上下颠倒")
		hasIssue = true
	} else if !fontFlippedY && !ctxFlippedY {
		fmt.Println("  ⚠️  非标准配置: 字体和上下文都没有Y轴翻转")
		fmt.Println("     → 文字可能显示正常，但不符合Cairo标准")
		hasIssue = true
	} else if !fontFlippedY && ctxFlippedY {
		fmt.Println("  ✅ 正确配置: 上下文Y轴翻转，字体使用正Y缩放")
		fmt.Println("     → 这是标准的Cairo文本渲染配置")
		fmt.Println("     → 上下文的Y翻转将图像坐标系转换为Cairo坐标系")
		fmt.Println("     → 字体的正Y缩放在Cairo坐标系中正确渲染字形")
	}

	if ctxFlippedX || fontFlippedX {
		fmt.Println("  ⚠️  检测到X轴翻转，这会导致文字左右镜像")
		hasIssue = true
	}

	// 6. 提供修复建议
	if hasIssue {
		fmt.Println("\n【修复建议】")
		if fontFlippedY {
			fmt.Println("  1. 修改字体矩阵创建代码:")
			fmt.Println("     fontMatrix.InitScale(fontSize, fontSize)  // 使用正Y缩放")
			fmt.Println()
			fmt.Println("  2. 确保GlyphPath函数中的翻转逻辑:")
			fmt.Println("     flipY := s.fontMatrix.YY > 0  // 当字体矩阵YY为正时翻转")
		}
		if fontFlippedY && ctxFlippedY {
			fmt.Println("  3. 移除字体矩阵中的负Y缩放，让上下文处理Y轴翻转")
		}
	}

	fmt.Println()
}

// printGlyphInformation prints detailed information about glyphs using the built-in PangoCairo methods
func printGlyphInformation(_ cairo.Context, scaledFont cairo.ScaledFont, text string, startX, startY float64) {
	fmt.Println("=== 字形分析 ===")

	// Get glyphs for the text
	glyphs, _, _, status := scaledFont.TextToGlyphs(startX, startY, text)
	if status != cairo.StatusSuccess {
		log.Fatal("获取字形失败:", status)
	}
	fmt.Printf("✓ 成功获取 %d 个字形\n\n", len(glyphs))

	// Since we know this is a PangoCairoScaledFont, we can cast it
	if pangoCairoFont, ok := scaledFont.(*cairo.PangoCairoScaledFont); ok {
		// Print detailed information for each glyph
		fmt.Println("=== 每个字符的详细信息 ===")
		runes := []rune(text)
		collisionCount := 0

		for i, glyph := range glyphs {
			var char rune
			if i < len(runes) {
				char = runes[i]
			} else {
				char = rune(glyph.Index)
			}

			fmt.Printf("--- 字符 #%d: '%c' ---\n", i+1, char)

			// Get and print metrics
			metrics, status := pangoCairoFont.GetGlyphMetrics(char)
			if status == cairo.StatusSuccess {
				fmt.Printf("  位置: (%.2f, %.2f)\n", glyph.X, glyph.Y)
				fmt.Printf("  边界框: [%.2f, %.2f] -> [%.2f, %.2f]\n",
					glyph.X+metrics.BoundingBox.XMin, glyph.Y+metrics.BoundingBox.YMin,
					glyph.X+metrics.BoundingBox.XMax, glyph.Y+metrics.BoundingBox.YMax)
				fmt.Printf("  宽度: %.2f, 高度: %.2f\n",
					metrics.BoundingBox.XMax-metrics.BoundingBox.XMin,
					metrics.BoundingBox.YMax-metrics.BoundingBox.YMin)
				fmt.Printf("  Advance: %.2f\n", metrics.XAdvance)
			}

			// Check for collisions with subsequent glyphs
			hasCollision := false
			for j := i + 1; j < len(glyphs); j++ {
				var nextChar rune
				if j < len(runes) {
					nextChar = runes[j]
				} else {
					nextChar = rune(glyphs[j].Index)
				}
				collides, status := pangoCairoFont.CheckGlyphCollision(glyph, glyphs[j], char, nextChar)
				if status == cairo.StatusSuccess && collides {
					fmt.Printf("  ⚠ 警告: 与字符 '%c' 发生重叠!\n", nextChar)
					hasCollision = true
					collisionCount++
				}
			}

			if !hasCollision {
				fmt.Printf("  ✓ 无重叠冲突\n")
			}
			fmt.Println()
		}

		// Summary
		fmt.Println("=== 碰撞检测总结 ===")
		if collisionCount == 0 {
			fmt.Printf("✓ 所有字符正常排列，无重叠\n")
		} else {
			fmt.Printf("⚠ 检测到 %d 处字符重叠\n", collisionCount)
		}
		fmt.Println()
	}
}
