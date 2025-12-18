//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("🎨 OpenType 特性测试...")

	surface := cairo.NewImageSurface(cairo.FormatARGB32, 1200, 1400)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 白色背景
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	y := 50.0

	// 测试 1: 自动检测文本方向
	fmt.Println("\n📊 测试 1: 自动检测文本方向")
	testAutoDirection(ctx, &y)

	// 测试 2: RTL 文本（阿拉伯文）
	fmt.Println("\n📊 测试 2: RTL 文本（阿拉伯文）")
	testRTLText(ctx, &y)

	// 测试 3: 混合方向文本
	fmt.Println("\n📊 测试 3: 混合方向文本")
	testMixedDirection(ctx, &y)

	// 测试 4: OpenType 特性 - 连字
	fmt.Println("\n📊 测试 4: OpenType 特性 - 连字")
	testLigatures(ctx, &y)

	// 测试 5: OpenType 特性 - 小型大写字母
	fmt.Println("\n📊 测试 5: OpenType 特性 - 小型大写字母")
	testSmallCaps(ctx, &y)

	// 测试 6: 复杂文字系统检测
	fmt.Println("\n📊 测试 6: 复杂文字系统检测")
	testComplexScripts(ctx, &y)

	// 测试 7: 语言和文字系统检测
	fmt.Println("\n📊 测试 7: 语言和文字系统检测")
	testLanguageDetection(ctx, &y)

	// 保存
	fmt.Println("\n💾 保存图片...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("opentype_features_test.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("保存失败: %v", status))
		}
		fmt.Println("✅ OpenType 特性测试图片已保存到 opentype_features_test.png")
	}

	fmt.Println("🎉 测试完成!")
}

func drawTitle(ctx cairo.Context, title string, y *float64) {
	ctx.SetSourceRGB(0.2, 0.2, 0.2)
	ctx.MoveTo(50, *y)
	
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("sans")
	fontDesc.SetWeight(cairo.PangoWeightBold)
	fontDesc.SetSize(20)
	layout.SetFontDescription(fontDesc)
	layout.SetText(title)
	ctx.PangoCairoShowText(layout)
	
	*y += 35
}

func drawText(ctx cairo.Context, text string, y *float64, color [3]float64) {
	ctx.SetSourceRGB(color[0], color[1], color[2])
	ctx.MoveTo(70, *y)
	
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("sans")
	fontDesc.SetWeight(cairo.PangoWeightNormal)
	fontDesc.SetSize(16)
	layout.SetFontDescription(fontDesc)
	layout.SetText(text)
	ctx.PangoCairoShowText(layout)
	
	*y += 25
}

func testAutoDirection(ctx cairo.Context, y *float64) {
	drawTitle(ctx, "1. 自动检测文本方向", y)

	tests := []struct {
		text string
		desc string
	}{
		{"Hello World", "英文 (LTR)"},
		{"مرحبا بالعالم", "阿拉伯文 (RTL)"},
		{"שלום עולם", "希伯来文 (RTL)"},
		{"你好世界", "中文 (LTR)"},
		{"Привет мир", "俄文 (LTR)"},
	}

	for _, test := range tests {
		direction := cairo.DetectTextDirection(test.text)
		dirStr := "LTR"
		if direction == cairo.TextDirectionRTL {
			dirStr = "RTL"
		}
		info := fmt.Sprintf("%s: %s → %s", test.desc, test.text, dirStr)
		drawText(ctx, info, y, [3]float64{0, 0, 0})
		fmt.Printf("  %s\n", info)
	}

	*y += 10
}

func testRTLText(ctx cairo.Context, y *float64) {
	drawTitle(ctx, "2. RTL 文本渲染", y)

	// 创建 RTL 文本的 shaping options
	options := cairo.NewShapingOptions()
	options.Direction = cairo.TextDirectionRTL
	options.Language = "ar"
	options.Script = "Arab"

	rtlTexts := []string{
		"مرحبا",      // Hello
		"العربية",    // Arabic
		"القاهرة",    // Cairo
	}

	for _, text := range rtlTexts {
		drawText(ctx, text, y, [3]float64{0, 0.3, 0.6})
		fmt.Printf("  RTL: %s\n", text)
	}

	*y += 10
}

func testMixedDirection(ctx cairo.Context, y *float64) {
	drawTitle(ctx, "3. 混合方向文本", y)

	mixedTexts := []string{
		"Hello مرحبا World",
		"English עברית Mixed",
		"中文 English 混合",
	}

	for _, text := range mixedTexts {
		// 分析双向文本
		runs := cairo.SplitBidiRuns(text)
		info := fmt.Sprintf("%s → %d runs", text, len(runs))
		drawText(ctx, info, y, [3]float64{0.3, 0, 0.6})
		fmt.Printf("  Mixed: %s\n", info)
		
		for i, run := range runs {
			level := "LTR"
			if run.Level == 1 {
				level = "RTL"
			}
			fmt.Printf("    Run %d: '%s' (%s)\n", i+1, run.Text, level)
		}
	}

	*y += 10
}

func testLigatures(ctx cairo.Context, y *float64) {
	drawTitle(ctx, "4. 连字特性", y)

	// 启用连字
	optionsOn := cairo.NewShapingOptions()
	cairo.SetDefaultFeatures(optionsOn, "default")

	// 禁用连字
	optionsOff := cairo.NewShapingOptions()
	cairo.SetDefaultFeatures(optionsOff, "no-ligatures")

	ligatureTexts := []string{
		"fi fl ffi ffl",
		"office difficult",
	}

	for _, text := range ligatureTexts {
		drawText(ctx, fmt.Sprintf("连字开启: %s", text), y, [3]float64{0, 0.5, 0})
		drawText(ctx, fmt.Sprintf("连字关闭: %s", text), y, [3]float64{0.5, 0.5, 0.5})
		fmt.Printf("  Ligatures: %s\n", text)
	}

	*y += 10
}

func testSmallCaps(ctx cairo.Context, y *float64) {
	drawTitle(ctx, "5. 小型大写字母", y)

	options := cairo.NewShapingOptions()
	cairo.SetDefaultFeatures(options, "small-caps")

	texts := []string{
		"Hello World",
		"Small Caps Test",
	}

	for _, text := range texts {
		drawText(ctx, fmt.Sprintf("普通: %s", text), y, [3]float64{0, 0, 0})
		drawText(ctx, fmt.Sprintf("小型大写: %s", text), y, [3]float64{0, 0.3, 0.6})
		fmt.Printf("  Small Caps: %s\n", text)
	}

	*y += 10
}

func testComplexScripts(ctx cairo.Context, y *float64) {
	drawTitle(ctx, "6. 复杂文字系统", y)

	tests := []struct {
		text    string
		desc    string
		complex bool
	}{
		{"Hello", "英文", false},
		{"مرحبا", "阿拉伯文", true},
		{"नमस्ते", "印地语", true},
		{"สวัสดี", "泰文", true},
		{"你好", "中文", false},
	}

	for _, test := range tests {
		needsComplex := cairo.NeedsComplexShaping(test.text)
		status := "简单"
		if needsComplex {
			status = "复杂"
		}
		info := fmt.Sprintf("%s (%s): %s", test.text, test.desc, status)
		
		color := [3]float64{0, 0.5, 0}
		if needsComplex {
			color = [3]float64{0.8, 0.3, 0}
		}
		
		drawText(ctx, info, y, color)
		fmt.Printf("  %s\n", info)
	}

	*y += 10
}

func testLanguageDetection(ctx cairo.Context, y *float64) {
	drawTitle(ctx, "7. 语言和文字系统检测", y)

	tests := []string{
		"Hello World",
		"مرحبا بالعالم",
		"שלום עולם",
		"Привет мир",
		"你好世界",
		"こんにちは",
		"안녕하세요",
		"नमस्ते",
		"สวัสดี",
	}

	for _, text := range tests {
		lang := cairo.DetectLanguage(text)
		script := cairo.DetectScript(text)
		info := fmt.Sprintf("%s → Lang: %s, Script: %s", text, lang, script)
		drawText(ctx, info, y, [3]float64{0.2, 0.2, 0.2})
		fmt.Printf("  %s\n", info)
	}

	*y += 10
}
