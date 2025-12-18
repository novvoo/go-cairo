//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"runtime"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("🔤 测试中文文字渲染...")
	fmt.Printf("操作系统: %s\n", runtime.GOOS)

	surface := cairo.NewImageSurface(cairo.FormatARGB32, 900, 700)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 白色背景
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()

	// 绘制标题
	fmt.Println("\n绘制标题...")
	ctx.SetSourceRGB(0.2, 0.2, 0.2)
	fontDesc.SetFamily("sans")
	fontDesc.SetSize(32)
	fontDesc.SetWeight(cairo.PangoWeightBold)
	layout.SetFontDescription(fontDesc)
	layout.SetText("中文字体渲染测试")
	
	extents := layout.GetPixelExtents()
	fontExtents := layout.GetFontExtents()
	titleX := 450.0 - extents.Width/2
	// 标题基线位置：让文字顶部在 y=20，所以基线 = 20 + Ascent
	titleY := 20.0 + fontExtents.Ascent
	
	fmt.Printf("标题位置: x=%.2f, y=%.2f (宽度=%.2f, Ascent=%.2f, 顶部=%.2f)\n", 
		titleX, titleY, extents.Width, fontExtents.Ascent, titleY-fontExtents.Ascent)
	
	ctx.MoveTo(titleX, titleY)
	ctx.PangoCairoShowText(layout)

	// 绘制分隔线（在标题下方留出空间）
	// 分隔线位置 = 标题基线 + Descent + 间距
	separatorY := titleY + fontExtents.Descent + 15
	ctx.SetSourceRGB(0.8, 0.8, 0.8)
	ctx.SetLineWidth(1)
	ctx.MoveTo(50, separatorY)
	ctx.LineTo(850, separatorY)
	ctx.Stroke()
	
	fmt.Printf("分隔线位置: y=%.2f\n", separatorY)

	// 测试不同的字体
	fonts := []struct {
		name    string
		display string
	}{
		{"Go Regular", "Go Regular (英文字体)"},
		{"sans", "sans (系统默认)"},
		{"PingFang SC", "PingFang SC (苹方)"},
		{"Hiragino Sans GB", "Hiragino Sans GB (冬青黑)"},
		{"STHeiti", "STHeiti (华文黑体)"},
		{"Arial Unicode MS", "Arial Unicode MS (通用)"},
	}

	// 从分隔线下方开始绘制字体测试
	y := separatorY + 20
	for _, font := range fonts {
		fmt.Printf("\n测试字体: %s\n", font.name)
		
		// 显示字体名称（小号灰色）
		ctx.SetSourceRGB(0.5, 0.5, 0.5)
		fontDesc.SetFamily("sans")
		fontDesc.SetSize(14)
		fontDesc.SetWeight(cairo.PangoWeightNormal)
		layout.SetFontDescription(fontDesc)
		layout.SetText(font.display)
		
		fontExtents = layout.GetFontExtents()
		labelY := y + fontExtents.Ascent
		ctx.MoveTo(50, labelY)
		ctx.PangoCairoShowText(layout)

		// 显示测试文本（使用指定字体）
		ctx.SetSourceRGB(0, 0, 0)
		fontDesc.SetFamily(font.name)
		fontDesc.SetSize(24)
		layout.SetFontDescription(fontDesc)
		layout.SetText("你好世界 Hello World 123 测试")
		
		fontExtents = layout.GetFontExtents()
		textY := y + 20 + fontExtents.Ascent
		ctx.MoveTo(50, textY)
		ctx.PangoCairoShowText(layout)

		y += 80
	}

	// 绘制分隔线
	ctx.SetSourceRGB(0.8, 0.8, 0.8)
	ctx.SetLineWidth(1)
	ctx.MoveTo(50, y + 10)
	ctx.LineTo(850, y + 10)
	ctx.Stroke()

	// 测试大号中文
	fmt.Println("\n测试大号中文...")
	ctx.SetSourceRGB(0.1, 0.3, 0.6)
	fontDesc.SetFamily("sans")
	fontDesc.SetSize(48)
	fontDesc.SetWeight(cairo.PangoWeightBold)
	layout.SetFontDescription(fontDesc)
	layout.SetText("Cairo 图形库")

	extents = layout.GetPixelExtents()
	fontExtents = layout.GetFontExtents()
	bigTextX := 450.0 - extents.Width/2
	bigTextY := y + 60 + fontExtents.Ascent
	
	fmt.Printf("大号文字位置: x=%.2f, y=%.2f\n", bigTextX, bigTextY)
	
	ctx.MoveTo(bigTextX, bigTextY)
	ctx.PangoCairoShowText(layout)

	// 保存
	fmt.Println("\n💾 保存图片...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("chinese_text_test.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("保存失败: %v", status))
		}
		fmt.Println("✅ 中文文字测试图片已保存到 chinese_text_test.png")
	}

	fmt.Println("🎉 测试完成!")
}
