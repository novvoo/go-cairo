//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("🌍 多语言文本渲染测试...")

	surface := cairo.NewImageSurface(cairo.FormatARGB32, 1000, 1200)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 白色背景
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// 标题
	ctx.SetSourceRGB(0.1, 0.1, 0.3)
	ctx.MoveTo(50, 60)
	
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("sans")
	fontDesc.SetWeight(cairo.PangoWeightBold)
	fontDesc.SetSize(36)
	layout.SetFontDescription(fontDesc)
	layout.SetText("多语言文本渲染")
	ctx.PangoCairoShowText(layout)

	// 分隔线
	ctx.SetSourceRGB(0.7, 0.7, 0.7)
	ctx.SetLineWidth(2)
	ctx.MoveTo(50, 80)
	ctx.LineTo(950, 80)
	ctx.Stroke()

	y := 130.0

	// 测试各种语言
	languages := []struct {
		name   string
		text   string
		color  [3]float64
		size   float64
	}{
		{
			name:  "英语 (English)",
			text:  "The quick brown fox jumps over the lazy dog",
			color: [3]float64{0.2, 0.2, 0.2},
			size:  24,
		},
		{
			name:  "阿拉伯语 (Arabic) - RTL",
			text:  "مرحبا بك في عالم الرسومات الجميلة",
			color: [3]float64{0.8, 0.3, 0.1},
			size:  24,
		},
		{
			name:  "希伯来语 (Hebrew) - RTL",
			text:  "שלום לכולם ברוכים הבאים",
			color: [3]float64{0.1, 0.4, 0.8},
			size:  24,
		},
		{
			name:  "中文 (Chinese)",
			text:  "春眠不觉晓，处处闻啼鸟",
			color: [3]float64{0.8, 0.1, 0.3},
			size:  28,
		},
		{
			name:  "日语 (Japanese)",
			text:  "こんにちは、世界！美しいグラフィックス",
			color: [3]float64{0.6, 0.2, 0.6},
			size:  24,
		},
		{
			name:  "韩语 (Korean)",
			text:  "안녕하세요 아름다운 세상",
			color: [3]float64{0.2, 0.6, 0.4},
			size:  24,
		},
		{
			name:  "俄语 (Russian)",
			text:  "Привет мир! Красивая графика",
			color: [3]float64{0.3, 0.3, 0.7},
			size:  24,
		},
		{
			name:  "希腊语 (Greek)",
			text:  "Γεια σου κόσμε! Όμορφα γραφικά",
			color: [3]float64{0.1, 0.5, 0.5},
			size:  24,
		},
		{
			name:  "印地语 (Hindi)",
			text:  "नमस्ते दुनिया सुंदर ग्राफिक्स",
			color: [3]float64{0.9, 0.5, 0.1},
			size:  24,
		},
		{
			name:  "泰语 (Thai)",
			text:  "สวัสดีชาวโลก กราฟิกที่สวยงาม",
			color: [3]float64{0.5, 0.1, 0.7},
			size:  24,
		},
	}

	for _, lang := range languages {
		fmt.Printf("\n渲染: %s\n", lang.name)
		fmt.Printf("  文本: %s\n", lang.text)

		// 绘制语言名称
		ctx.SetSourceRGB(0.4, 0.4, 0.4)
		ctx.MoveTo(50, y)
		
		layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
		fontDesc := cairo.NewPangoFontDescription()
		fontDesc.SetFamily("sans")
		fontDesc.SetWeight(cairo.PangoWeightBold)
		fontDesc.SetSize(16)
		layout.SetFontDescription(fontDesc)
		layout.SetText(lang.name)
		ctx.PangoCairoShowText(layout)
		y += 25

		// 自动检测文本属性
		direction := cairo.DetectTextDirection(lang.text)
		language := cairo.DetectLanguage(lang.text)
		script := cairo.DetectScript(lang.text)
		needsComplex := cairo.NeedsComplexShaping(lang.text)

		fmt.Printf("  方向: %v, 语言: %s, 文字: %s, 复杂: %v\n",
			direction, language, script, needsComplex)

		// 创建 shaping options
		options := cairo.NewShapingOptions()
		options.Direction = direction
		options.Language = language
		options.Script = script

		// 绘制文本
		ctx.SetSourceRGB(lang.color[0], lang.color[1], lang.color[2])
		
		// 对于RTL文本，从左边开始但使用右对齐
		x := 70.0
		
		ctx.MoveTo(x, y)
		
		layout2 := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
		fontDesc2 := cairo.NewPangoFontDescription()
		fontDesc2.SetFamily("sans")
		fontDesc2.SetWeight(cairo.PangoWeightNormal)
		fontDesc2.SetSize(lang.size)
		layout2.SetFontDescription(fontDesc2)
		
		// 对于RTL文本使用右对齐，设置宽度让文本在指定区域内右对齐
		// 这是正确的国际化文本显示方式
		if direction == cairo.TextDirectionRTL {
			layout2.SetAlignment(cairo.PangoAlignRight)
			// 设置宽度为可用区域（从70到930）
			availableWidth := 860.0
			width := int(availableWidth * 1024) // Pango使用1024为单位
			layout2.SetWidth(width)
		}
		
		// 如果想强制所有文本都从左边显示（不推荐），可以注释掉上面的if块
		
		layout2.SetText(lang.text)
		
		ctx.PangoCairoShowText(layout2)
		y += 40

		// 绘制信息框
		ctx.SetSourceRGBA(0.9, 0.9, 0.9, 0.5)
		ctx.Rectangle(70, y-15, 860, 20)
		ctx.Fill()

		ctx.SetSourceRGB(0.5, 0.5, 0.5)
		ctx.MoveTo(75, y)
		
		layout3 := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
		fontDesc3 := cairo.NewPangoFontDescription()
		fontDesc3.SetFamily("mono")
		fontDesc3.SetWeight(cairo.PangoWeightNormal)
		fontDesc3.SetSize(12)
		layout3.SetFontDescription(fontDesc3)
		info := fmt.Sprintf("Dir: %v | Lang: %s | Script: %s | Complex: %v",
			direction, language, script, needsComplex)
		layout3.SetText(info)
		ctx.PangoCairoShowText(layout3)
		y += 35
	}

	// 底部说明
	y += 20
	ctx.SetSourceRGB(0.3, 0.3, 0.3)
	
	layout4 := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc4 := cairo.NewPangoFontDescription()
	fontDesc4.SetFamily("sans")
	fontDesc4.SetStyle(cairo.PangoStyleItalic)
	fontDesc4.SetWeight(cairo.PangoWeightNormal)
	fontDesc4.SetSize(14)
	layout4.SetFontDescription(fontDesc4)
	
	ctx.MoveTo(50, y)
	layout4.SetText("✨ 自动检测文本方向、语言和文字系统")
	ctx.PangoCairoShowText(layout4)
	y += 25
	
	ctx.MoveTo(50, y)
	layout4.SetText("✨ 支持 LTR、RTL 和复杂文字系统")
	ctx.PangoCairoShowText(layout4)
	y += 25
	
	ctx.MoveTo(50, y)
	layout4.SetText("✨ 使用 HarfBuzz 进行高质量文本塑形")
	ctx.PangoCairoShowText(layout4)

	// 保存
	fmt.Println("\n💾 保存图片...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("multilingual_text.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("保存失败: %v", status))
		}
		fmt.Println("✅ 多语言文本测试图片已保存到 multilingual_text.png")
	}

	fmt.Println("🎉 测试完成!")
}
