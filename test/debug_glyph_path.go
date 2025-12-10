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

	// 获取 M 的字形
	glyphsM, _, _, _ := scaledFont.TextToGlyphs(100, 100, "M")
	if len(glyphsM) == 0 {
		fmt.Println("无法获取 M 的字形")
		return
	}

	// 获取 M 的路径
	path, err := scaledFont.GlyphPath(glyphsM[0].Index)
	if err != nil {
		fmt.Printf("获取路径失败: %v\n", err)
		return
	}

	fmt.Println("=== M 字形路径分析 ===")
	fmt.Printf("字形位置: (%.2f, %.2f)\n", glyphsM[0].X, glyphsM[0].Y)
	fmt.Printf("路径段数: %d\n\n", len(path.Data))

	// 分析路径的实际边界
	var minX, maxX, minY, maxY float64
	firstPoint := true

	for i, pd := range path.Data {
		fmt.Printf("段 %d: 类型=%v, 点数=%d\n", i, pd.Type, len(pd.Points))
		for j, pt := range pd.Points {
			// 加上字形位置偏移
			absX := pt.X + glyphsM[0].X
			absY := pt.Y + glyphsM[0].Y

			fmt.Printf("  点 %d: 相对(%.2f, %.2f) -> 绝对(%.2f, %.2f)\n", j, pt.X, pt.Y, absX, absY)

			if firstPoint {
				minX, maxX = absX, absX
				minY, maxY = absY, absY
				firstPoint = false
			} else {
				if absX < minX {
					minX = absX
				}
				if absX > maxX {
					maxX = absX
				}
				if absY < minY {
					minY = absY
				}
				if absY > maxY {
					maxY = absY
				}
			}
		}
	}

	fmt.Printf("\n=== 路径实际边界 ===\n")
	fmt.Printf("X: %.2f 到 %.2f (宽度 %.2f)\n", minX, maxX, maxX-minX)
	fmt.Printf("Y: %.2f 到 %.2f (高度 %.2f)\n", minY, maxY, maxY-minY)

	// 对比字形度量
	metrics, _ := scaledFont.GetGlyphMetrics('M')
	fmt.Printf("\n=== 字形度量边界 ===\n")
	fmt.Printf("BoundingBox.XMin: %.2f\n", metrics.BoundingBox.XMin)
	fmt.Printf("BoundingBox.XMax: %.2f\n", metrics.BoundingBox.XMax)
	fmt.Printf("BoundingBox.YMin: %.2f\n", metrics.BoundingBox.YMin)
	fmt.Printf("BoundingBox.YMax: %.2f\n", metrics.BoundingBox.YMax)
	fmt.Printf("XBearing: %.2f\n", metrics.XBearing)
	fmt.Printf("YBearing: %.2f\n", metrics.YBearing)
	fmt.Printf("X: %.2f 到 %.2f (宽度 %.2f)\n",
		glyphsM[0].X+metrics.BoundingBox.XMin,
		glyphsM[0].X+metrics.BoundingBox.XMax,
		metrics.BoundingBox.XMax-metrics.BoundingBox.XMin)
	fmt.Printf("Y: %.2f 到 %.2f (高度 %.2f)\n",
		glyphsM[0].Y+metrics.BoundingBox.YMin,
		glyphsM[0].Y+metrics.BoundingBox.YMax,
		metrics.BoundingBox.YMax-metrics.BoundingBox.YMin)
	fmt.Printf("Advance: %.2f\n", metrics.XAdvance)
}
