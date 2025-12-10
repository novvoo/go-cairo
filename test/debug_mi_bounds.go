//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	// 创建字体
	fontFace := cairo.NewPangoCairoFont("Go Regular", cairo.FontSlantNormal, cairo.FontWeightNormal)
	defer fontFace.Destroy()

	fontMatrix := cairo.NewMatrix()
	fontMatrix.InitScale(48.0, 48.0)

	ctm := cairo.NewMatrix()
	ctm.InitIdentity()

	scaledFont := cairo.NewPangoCairoScaledFont(fontFace, fontMatrix, ctm, nil)
	defer scaledFont.Destroy()

	// 获取 M 和 I 的字形
	glyphsM, _, _, _ := scaledFont.TextToGlyphs(0, 0, "M")
	glyphsI, _, _, _ := scaledFont.TextToGlyphs(0, 0, "I")

	fmt.Println("=== M 字形信息 ===")
	if len(glyphsM) > 0 {
		scaledFont.PrintGlyphInfo(glyphsM[0], 'M')
	}

	fmt.Println("=== I 字形信息 ===")
	if len(glyphsI) > 0 {
		scaledFont.PrintGlyphInfo(glyphsI[0], 'I')
	}

	// 测试不同间距下的 MI
	fmt.Println("\n=== 测试不同间距 ===")
	for spacing := 0.0; spacing <= 20.0; spacing += 2.0 {
		// 手动计算 M 和 I 的位置
		mX := 100.0
		mY := 100.0

		// 获取 M 的度量
		metricsM, _ := scaledFont.GetGlyphMetrics('M')
		iX := mX + metricsM.XAdvance + spacing
		iY := mY

		// 创建字形
		glyphM := cairo.Glyph{Index: glyphsM[0].Index, X: mX, Y: mY}
		glyphI := cairo.Glyph{Index: glyphsI[0].Index, X: iX, Y: iY}

		// 检查碰撞
		collides, _ := scaledFont.CheckGlyphCollision(glyphM, glyphI, 'M', 'I')

		status := "✅ 无重叠"
		if collides {
			status = "❌ 重叠"
		}

		fmt.Printf("间距 %.1f: M右边界=%.2f, I左边界=%.2f, 间隙=%.2f %s\n",
			spacing,
			mX+metricsM.BoundingBox.XMax,
			iX+metricsM.BoundingBox.XMin,
			(iX+metricsM.BoundingBox.XMin)-(mX+metricsM.BoundingBox.XMax),
			status)
	}
}
