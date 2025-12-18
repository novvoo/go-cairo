//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("🎨 中文渐变效果测试...")

	surface := cairo.NewImageSurface(cairo.FormatARGB32, 1000, 700)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 深色背景
	ctx.SetSourceRGB(0.05, 0.05, 0.1)
	ctx.Paint()

	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()

	// 测试1: 中文标题 - 水平渐变
	fmt.Println("\n📊 测试1: 中文标题 - 水平渐变")
	fontDesc.SetFamily("sans")
	fontDesc.SetSize(64)
	fontDesc.SetWeight(cairo.PangoWeightBold)
	layout.SetFontDescription(fontDesc)
	layout.SetText("渐变效果")

	extents := layout.GetPixelExtents()
	textX := 500.0 - extents.Width/2
	textY := 80.0

	pattern1 := cairo.NewPatternLinear(textX, textY-extents.Height, textX+extents.Width, textY)
	if gradPat, ok := pattern1.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 0.3, 0.3) // 红
		gradPat.AddColorStopRGB(0.5, 1, 1, 0.3) // 黄
		gradPat.AddColorStopRGB(1, 0.3, 1, 0.3) // 绿
	}
	ctx.SetSource(pattern1)
	ctx.MoveTo(textX, textY)
	ctx.PangoCairoShowText(layout)
	pattern1.Destroy()

	// 测试2: 中文副标题 - 垂直渐变
	fmt.Println("📊 测试2: 中文副标题 - 垂直渐变")
	fontDesc.SetSize(36)
	fontDesc.SetWeight(cairo.PangoWeightNormal)
	layout.SetFontDescription(fontDesc)
	layout.SetText("Cairo 图形库测试")

	extents = layout.GetPixelExtents()
	textX = 500.0 - extents.Width/2
	textY = 150.0

	pattern2 := cairo.NewPatternLinear(textX, textY-extents.Height, textX, textY)
	if gradPat, ok := pattern2.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 0.3, 0.8, 1) // 亮蓝
		gradPat.AddColorStopRGB(1, 0.5, 0.3, 1) // 紫
	}
	ctx.SetSource(pattern2)
	ctx.MoveTo(textX, textY)
	ctx.PangoCairoShowText(layout)
	pattern2.Destroy()

	// 测试3: 多行中文 - 彩虹渐变
	fmt.Println("📊 测试3: 多行中文 - 彩虹渐变")
	texts := []string{
		"春眠不觉晓",
		"处处闻啼鸟",
		"夜来风雨声",
		"花落知多少",
	}

	fontDesc.SetSize(32)
	layout.SetFontDescription(fontDesc)

	y := 230.0
	colors := [][3]float64{
		{1, 0.3, 0.3}, // 红
		{1, 0.8, 0.3}, // 橙
		{0.3, 1, 0.3}, // 绿
		{0.3, 0.5, 1}, // 蓝
	}

	for i, text := range texts {
		layout.SetText(text)
		extents = layout.GetPixelExtents()
		textX = 150.0

		pattern := cairo.NewPatternLinear(textX, y, textX+extents.Width, y)
		if gradPat, ok := pattern.(cairo.LinearGradientPattern); ok {
			c := colors[i]
			gradPat.AddColorStopRGB(0, c[0], c[1], c[2])
			gradPat.AddColorStopRGB(1, c[0]*0.5, c[1]*0.5, c[2]*0.5)
		}
		ctx.SetSource(pattern)
		ctx.MoveTo(textX, y)
		ctx.PangoCairoShowText(layout)
		pattern.Destroy()

		y += 50
	}

	// 测试4: 中文 + 英文混合 - 对角渐变
	fmt.Println("📊 测试4: 中英混合 - 对角渐变")
	fontDesc.SetSize(28)
	layout.SetFontDescription(fontDesc)
	layout.SetText("Hello 世界 · 你好 World")

	extents = layout.GetPixelExtents()
	textX = 500.0
	textY = 280.0

	pattern4 := cairo.NewPatternLinear(textX, textY-extents.Height, textX+extents.Width, textY)
	if gradPat, ok := pattern4.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 1, 0.3)     // 黄
		gradPat.AddColorStopRGB(0.5, 1, 0.5, 0.8) // 粉
		gradPat.AddColorStopRGB(1, 0.5, 1, 1)     // 青
	}
	ctx.SetSource(pattern4)
	ctx.MoveTo(textX, textY)
	ctx.PangoCairoShowText(layout)
	pattern4.Destroy()

	// 测试5: 径向渐变背景 + 中文
	fmt.Println("⭕ 测试5: 径向渐变背景 + 中文")

	// 绘制径向渐变圆形背景
	pattern5bg := cairo.NewPatternRadial(500, 480, 0, 500, 480, 150)
	if gradPat, ok := pattern5bg.(cairo.RadialGradientPattern); ok {
		gradPat.AddColorStopRGBA(0, 1, 0.8, 0.2, 0.8)   // 金色中心
		gradPat.AddColorStopRGBA(0.7, 1, 0.4, 0.1, 0.5) // 橙色
		gradPat.AddColorStopRGBA(1, 0.8, 0.2, 0, 0)     // 透明边缘
	}
	ctx.SetSource(pattern5bg)
	ctx.Arc(500, 480, 150, 0, 6.28)
	ctx.Fill()
	pattern5bg.Destroy()

	// 在圆形上绘制文字
	fontDesc.SetSize(48)
	fontDesc.SetWeight(cairo.PangoWeightBold)
	layout.SetFontDescription(fontDesc)
	layout.SetText("圆满")

	extents = layout.GetPixelExtents()
	textX = 500.0 - extents.Width/2
	textY = 490.0

	// 文字使用白色到透明的渐变
	pattern5text := cairo.NewPatternLinear(textX, textY-extents.Height, textX, textY)
	if gradPat, ok := pattern5text.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGBA(0, 1, 1, 1, 1)     // 白色
		gradPat.AddColorStopRGBA(1, 1, 0.9, 0.8, 1) // 浅黄
	}
	ctx.SetSource(pattern5text)
	ctx.MoveTo(textX, textY)
	ctx.PangoCairoShowText(layout)
	pattern5text.Destroy()

	// 测试6: 数字和中文 - 多色渐变
	fmt.Println("📊 测试6: 数字和中文 - 多色渐变")
	fontDesc.SetSize(40)
	fontDesc.SetWeight(cairo.PangoWeightNormal)
	layout.SetFontDescription(fontDesc)
	layout.SetText("2024年 · 新年快乐")

	extents = layout.GetPixelExtents()
	textX = 500.0 - extents.Width/2
	textY = 620.0

	pattern6 := cairo.NewPatternLinear(textX, textY, textX+extents.Width, textY)
	if gradPat, ok := pattern6.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 0.2, 0.2)    // 红
		gradPat.AddColorStopRGB(0.25, 1, 0.6, 0.2) // 橙
		gradPat.AddColorStopRGB(0.5, 1, 1, 0.2)    // 黄
		gradPat.AddColorStopRGB(0.75, 0.2, 1, 0.5) // 绿
		gradPat.AddColorStopRGB(1, 0.5, 0.5, 1)    // 紫
	}
	ctx.SetSource(pattern6)
	ctx.MoveTo(textX, textY)
	ctx.PangoCairoShowText(layout)
	pattern6.Destroy()

	// 保存
	fmt.Println("\n💾 保存图片...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("chinese_gradient_test.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("保存失败: %v", status))
		}
		fmt.Println("✅ 中文渐变测试图片已保存到 chinese_gradient_test.png")
	}

	fmt.Println("🎉 中文渐变测试完成!")
}
