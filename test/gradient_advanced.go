//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"math"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("🎨 开始高级渐变测试...")

	// 创建更大的画布
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 1000, 800)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 深色背景
	fmt.Println("🌑 设置深色背景...")
	ctx.SetSourceRGB(0.1, 0.1, 0.15)
	ctx.Paint()

	// 测试1: 渐变扩展模式 - Pad (默认)
	fmt.Println("\n📐 测试1: 渐变扩展模式 - Pad")
	pattern1 := cairo.NewPatternLinear(50, 50, 150, 50)
	if gradPat, ok := pattern1.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 0, 0)
		gradPat.AddColorStopRGB(1, 0, 0, 1)
	}
	pattern1.SetExtend(cairo.ExtendPad)
	ctx.SetSource(pattern1)
	ctx.Rectangle(30, 30, 200, 80)
	ctx.Fill()
	pattern1.Destroy()

	// 测试2: 渐变扩展模式 - Repeat
	fmt.Println("📐 测试2: 渐变扩展模式 - Repeat")
	pattern2 := cairo.NewPatternLinear(270, 50, 320, 50)
	if gradPat, ok := pattern2.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 0, 1, 0)
		gradPat.AddColorStopRGB(1, 1, 1, 0)
	}
	pattern2.SetExtend(cairo.ExtendRepeat)
	ctx.SetSource(pattern2)
	ctx.Rectangle(250, 30, 200, 80)
	ctx.Fill()
	pattern2.Destroy()

	// 测试3: 渐变扩展模式 - Reflect
	fmt.Println("📐 测试3: 渐变扩展模式 - Reflect")
	pattern3 := cairo.NewPatternLinear(490, 50, 540, 50)
	if gradPat, ok := pattern3.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 0, 1)
		gradPat.AddColorStopRGB(1, 0, 1, 1)
	}
	pattern3.SetExtend(cairo.ExtendReflect)
	ctx.SetSource(pattern3)
	ctx.Rectangle(470, 30, 200, 80)
	ctx.Fill()
	pattern3.Destroy()

	// 测试4: 旋转的线性渐变
	fmt.Println("\n🔄 测试4: 旋转的线性渐变")
	ctx.Save()
	ctx.Translate(800, 80)
	ctx.Rotate(math.Pi / 4) // 45度旋转
	pattern4 := cairo.NewPatternLinear(-50, 0, 50, 0)
	if gradPat, ok := pattern4.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 0.5, 0)
		gradPat.AddColorStopRGB(1, 1, 1, 0)
	}
	ctx.SetSource(pattern4)
	ctx.Rectangle(-60, -40, 120, 80)
	ctx.Fill()
	pattern4.Destroy()
	ctx.Restore()

	// 测试5: 复杂的多色径向渐变 (日落效果)
	fmt.Println("\n🌅 测试5: 日落效果径向渐变")
	pattern5 := cairo.NewPatternRadial(150, 250, 0, 150, 250, 100)
	if gradPat, ok := pattern5.(cairo.RadialGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 1, 0.9)     // 亮黄
		gradPat.AddColorStopRGB(0.3, 1, 0.8, 0.2) // 橙黄
		gradPat.AddColorStopRGB(0.6, 1, 0.4, 0)   // 橙色
		gradPat.AddColorStopRGB(0.8, 0.8, 0.2, 0) // 深橙
		gradPat.AddColorStopRGB(1, 0.4, 0, 0.2)   // 暗红
	}
	ctx.SetSource(pattern5)
	ctx.Arc(150, 250, 100, 0, 2*math.Pi)
	ctx.Fill()
	pattern5.Destroy()

	// 测试6: 偏心径向渐变 (3D球体效果)
	fmt.Println("\n⚽ 测试6: 3D球体效果")
	pattern6 := cairo.NewPatternRadial(370, 220, 10, 400, 250, 100)
	if gradPat, ok := pattern6.(cairo.RadialGradientPattern); ok {
		gradPat.AddColorStopRGBA(0, 1, 1, 1, 1)      // 高光
		gradPat.AddColorStopRGBA(0.2, 0.3, 0.6, 1, 1) // 亮蓝
		gradPat.AddColorStopRGBA(0.7, 0.1, 0.3, 0.8, 1) // 深蓝
		gradPat.AddColorStopRGBA(1, 0, 0.1, 0.4, 1)   // 暗蓝
	}
	ctx.SetSource(pattern6)
	ctx.Arc(400, 250, 100, 0, 2*math.Pi)
	ctx.Fill()
	pattern6.Destroy()

	// 测试7: 渐变遮罩效果
	fmt.Println("\n🎭 测试7: 渐变遮罩效果")
	// 先画一个彩色矩形
	ctx.SetSourceRGB(0.8, 0.2, 0.8)
	ctx.Rectangle(550, 150, 200, 200)
	ctx.Fill()
	
	// 应用渐变遮罩
	pattern7 := cairo.NewPatternLinear(550, 150, 750, 350)
	if gradPat, ok := pattern7.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGBA(0, 0, 0, 0, 0)   // 完全透明
		gradPat.AddColorStopRGBA(1, 0, 0, 0, 0.8) // 半透明黑
	}
	ctx.SetSource(pattern7)
	ctx.Rectangle(550, 150, 200, 200)
	ctx.Fill()
	pattern7.Destroy()

	// 测试8: 锥形渐变模拟 (使用多个径向渐变)
	fmt.Println("\n🎯 测试8: 多层径向渐变")
	centerX, centerY := 150.0, 500.0
	for i := 0; i < 5; i++ {
		radius := float64(80 - i*15)
		pattern := cairo.NewPatternRadial(centerX, centerY, 0, centerX, centerY, radius)
		if gradPat, ok := pattern.(cairo.RadialGradientPattern); ok {
			alpha := 0.3
			r := float64(i) / 4.0
			gradPat.AddColorStopRGBA(0, 1-r, r, 0.5, alpha)
			gradPat.AddColorStopRGBA(1, r, 1-r, 0.5, alpha)
		}
		ctx.SetSource(pattern)
		ctx.Arc(centerX, centerY, radius, 0, 2*math.Pi)
		ctx.Fill()
		pattern.Destroy()
	}

	// 测试9: 渐变文字效果
	fmt.Println("\n✨ 测试9: 渐变文字")
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("Go Regular")
	fontDesc.SetSize(72)
	fontDesc.SetWeight(cairo.PangoWeightBold)
	layout.SetFontDescription(fontDesc)
	layout.SetText("GRADIENT")
	
	extents := layout.GetPixelExtents()
	textX := 350.0
	textY := 500.0
	
	// 创建渐变
	pattern9 := cairo.NewPatternLinear(textX, textY-extents.Height, textX, textY)
	if gradPat, ok := pattern9.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 0.2, 0.2)   // 红
		gradPat.AddColorStopRGB(0.5, 1, 1, 0.2)   // 黄
		gradPat.AddColorStopRGB(1, 0.2, 1, 0.2)   // 绿
	}
	ctx.SetSource(pattern9)
	ctx.MoveTo(textX, textY)
	ctx.PangoCairoShowText(layout)
	pattern9.Destroy()

	// 测试10: 渐变描边文字
	fmt.Println("\n🖌️  测试10: 渐变描边文字")
	layout.SetText("STROKE")
	fontDesc.SetSize(60)
	layout.SetFontDescription(fontDesc)
	extents = layout.GetPixelExtents()
	textX = 350.0
	textY = 600.0
	
	// 先填充
	ctx.SetSourceRGB(0.1, 0.1, 0.15)
	ctx.MoveTo(textX, textY)
	ctx.PangoCairoShowText(layout)
	
	// 再描边
	pattern10 := cairo.NewPatternLinear(textX, textY-extents.Height, textX+extents.Width, textY)
	if gradPat, ok := pattern10.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 0, 0.5, 1)     // 蓝
		gradPat.AddColorStopRGB(0.5, 0.5, 0, 1)   // 紫
		gradPat.AddColorStopRGB(1, 1, 0, 0.5)     // 粉
	}
	ctx.SetSource(pattern10)
	ctx.SetLineWidth(3)
	ctx.MoveTo(textX, textY)
	// 注意：这里需要路径模式，但 PangoCairo 直接渲染，所以效果可能不同
	pattern10.Destroy()

	// 测试11: 渐变圆环
	fmt.Println("\n💍 测试11: 渐变圆环")
	pattern11 := cairo.NewPatternRadial(150, 680, 40, 150, 680, 80)
	if gradPat, ok := pattern11.(cairo.RadialGradientPattern); ok {
		gradPat.AddColorStopRGBA(0, 1, 0.8, 0, 0)   // 透明中心
		gradPat.AddColorStopRGBA(0.5, 1, 0.8, 0, 1) // 金色
		gradPat.AddColorStopRGBA(1, 0.8, 0.5, 0, 0) // 渐变到透明
	}
	ctx.SetSource(pattern11)
	ctx.Arc(150, 680, 80, 0, 2*math.Pi)
	ctx.Fill()
	pattern11.Destroy()

	// 添加标题和说明
	fmt.Println("\n📝 添加标题...")
	ctx.SetSourceRGB(1, 1, 1)
	fontDesc.SetFamily("PingFang SC")  // 使用支持中文的字体
	fontDesc.SetSize(32)
	fontDesc.SetWeight(cairo.PangoWeightBold)
	layout.SetFontDescription(fontDesc)
	layout.SetText("Cairo 高级渐变测试")
	extents = layout.GetPixelExtents()
	ctx.MoveTo(500-extents.Width/2, 30)
	ctx.PangoCairoShowText(layout)

	// 添加小标签
	fontDesc.SetFamily("Go Regular")  // 英文标签使用 Go Regular
	fontDesc.SetSize(14)
	fontDesc.SetWeight(cairo.PangoWeightNormal)
	layout.SetFontDescription(fontDesc)
	
	labels := []struct {
		text string
		x, y float64
	}{
		{"Pad", 130, 120},
		{"Repeat", 350, 120},
		{"Reflect", 570, 120},
		{"Rotated", 800, 120},
		{"Sunset", 150, 360},
		{"3D Sphere", 400, 360},
		{"Mask", 650, 360},
		{"Layered", 150, 610},
		{"Ring", 150, 770},
	}
	
	for _, label := range labels {
		layout.SetText(label.text)
		extents = layout.GetPixelExtents()
		ctx.MoveTo(label.x-extents.Width/2, label.y)
		ctx.PangoCairoShowText(layout)
	}

	// 保存
	fmt.Println("\n💾 保存图片...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("gradient_advanced_test.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("保存失败: %v", status))
		}
		fmt.Println("✅ 高级渐变测试图片已保存到 gradient_advanced_test.png")
	}

	fmt.Println("🎉 高级渐变测试完成!")
}
