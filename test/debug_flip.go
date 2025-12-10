//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("=== 调试字形翻转问题 ===\n")

	// 创建一个小的测试表面
	width, height := 300, 150
	surface := cairo.NewImageSurface(cairo.FormatARGB32, width, height)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 白色背景
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// 黑色文字
	ctx.SetSourceRGB(0, 0, 0)

	// 测试单个字母 "A"
	text := "A"
	fontSize := 48.0

	// 创建字体
	fontFamily := "sans"
	pangoFont := cairo.NewPangoCairoFont(fontFamily, cairo.FontSlantNormal, cairo.FontWeightNormal)
	defer pangoFont.Destroy()

	// 创建字体矩阵
	fontMatrix := cairo.NewMatrix()
	fontMatrix.InitScale(fontSize, fontSize)

	ctm := cairo.NewMatrix()
	ctm.InitIdentity()

	scaledFont := cairo.NewPangoCairoScaledFont(pangoFont, fontMatrix, ctm, nil)
	defer scaledFont.Destroy()

	// 获取上下文矩阵
	ctxMatrix := ctx.GetMatrix()
	fmt.Printf("上下文矩阵 YY: %.4f\n", ctxMatrix.YY)
	fmt.Printf("字体矩阵 YY: %.4f\n", fontMatrix.YY)

	// 判断翻转逻辑
	flipY := fontMatrix.YY > 0
	fmt.Printf("GlyphPath 中 flipY 应该是: %v\n", flipY)
	fmt.Printf("解释: 字体矩阵YY=%.4f > 0，所以 flipY=%v\n\n", fontMatrix.YY, flipY)

	// 渲染文字
	x, y := 50.0, 100.0
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily(fontFamily)
	fontDesc.SetSize(fontSize)
	layout.SetFontDescription(fontDesc)
	layout.SetText(text)

	ctx.MoveTo(x, y)
	ctx.PangoCairoShowText(layout)

	// 获取字形路径来检查
	glyphs, _, _, status := scaledFont.TextToGlyphs(x, y, text)
	if status == cairo.StatusSuccess && len(glyphs) > 0 {
		glyph := glyphs[0]
		fmt.Printf("字形 '%s' 的位置: (%.2f, %.2f)\n", text, glyph.X, glyph.Y)

		// 获取字形路径
		glyphPath, err := scaledFont.GlyphPath(glyph.Index)
		if err == nil && glyphPath != nil {
			fmt.Printf("\n字形路径点数: %d\n", len(glyphPath.Data))
			
			// 打印前几个路径点
			fmt.Println("\n前5个路径操作:")
			for i := 0; i < 5 && i < len(glyphPath.Data); i++ {
				pd := glyphPath.Data[i]
				fmt.Printf("  [%d] 类型: %v, 点数: %d\n", i, pd.Type, len(pd.Points))
				for j, pt := range pd.Points {
					fmt.Printf("      点[%d]: (%.2f, %.2f)\n", j, pt.X, pt.Y)
				}
			}

			// 检查Y坐标的符号
			hasPositiveY := false
			hasNegativeY := false
			for _, pd := range glyphPath.Data {
				for _, pt := range pd.Points {
					if pt.Y > 0 {
						hasPositiveY = true
					}
					if pt.Y < 0 {
						hasNegativeY = true
					}
				}
			}

			fmt.Printf("\n路径Y坐标分析:\n")
			fmt.Printf("  有正Y坐标: %v\n", hasPositiveY)
			fmt.Printf("  有负Y坐标: %v\n", hasNegativeY)

			if hasPositiveY && !hasNegativeY {
				fmt.Println("  → 字形路径全部为正Y，说明字形是正向的（未翻转）")
			} else if !hasPositiveY && hasNegativeY {
				fmt.Println("  → 字形路径全部为负Y，说明字形已翻转")
			} else if hasPositiveY && hasNegativeY {
				fmt.Println("  → 字形路径有正有负，这是正常的字形轮廓")
			}
		}
	}

	// 保存图像
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		filename := "debug_flip.png"
		status := imgSurf.WriteToPNG(filename)
		if status != cairo.StatusSuccess {
			fmt.Printf("保存PNG失败: %v\n", status)
			os.Exit(1)
		}
		fmt.Printf("\n✓ 调试图像已保存到: %s\n", filename)
		fmt.Println("请检查图像中的字母 'A' 是否上下翻转")
	}
}
