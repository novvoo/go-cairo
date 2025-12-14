//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"math"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("🌈 开始渐变测试...")

	// 创建画布
	fmt.Println("📝 创建 800x600 像素的画布...")
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 800, 600)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 白色背景
	fmt.Println("🎨 设置白色背景...")
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// 测试1: 水平线性渐变
	fmt.Println("\n📊 测试1: 水平线性渐变 (红->蓝)")
	pattern1 := cairo.NewPatternLinear(50, 0, 250, 0)
	if gradPat, ok := pattern1.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 0, 0)    // 红色
		gradPat.AddColorStopRGB(1, 0, 0, 1)    // 蓝色
	}
	ctx.SetSource(pattern1)
	ctx.Rectangle(50, 50, 200, 100)
	ctx.Fill()
	pattern1.Destroy()
	fmt.Println("   ✓ 绘制完成")

	// 测试2: 垂直线性渐变
	fmt.Println("\n📊 测试2: 垂直线性渐变 (绿->黄)")
	pattern2 := cairo.NewPatternLinear(0, 200, 0, 400)
	if gradPat, ok := pattern2.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 0, 1, 0)    // 绿色
		gradPat.AddColorStopRGB(1, 1, 1, 0)    // 黄色
	}
	ctx.SetSource(pattern2)
	ctx.Rectangle(50, 200, 200, 100)
	ctx.Fill()
	pattern2.Destroy()
	fmt.Println("   ✓ 绘制完成")

	// 测试3: 对角线性渐变
	fmt.Println("\n📊 测试3: 对角线性渐变 (青->洋红)")
	pattern3 := cairo.NewPatternLinear(50, 350, 250, 550)
	if gradPat, ok := pattern3.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 0, 1, 1)    // 青色
		gradPat.AddColorStopRGB(1, 1, 0, 1)    // 洋红色
	}
	ctx.SetSource(pattern3)
	ctx.Rectangle(50, 350, 200, 100)
	ctx.Fill()
	pattern3.Destroy()
	fmt.Println("   ✓ 绘制完成")

	// 测试4: 多色线性渐变
	fmt.Println("\n📊 测试4: 多色线性渐变 (彩虹)")
	pattern4 := cairo.NewPatternLinear(300, 50, 700, 50)
	if gradPat, ok := pattern4.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0.0, 1, 0, 0)  // 红
		gradPat.AddColorStopRGB(0.2, 1, 1, 0)  // 黄
		gradPat.AddColorStopRGB(0.4, 0, 1, 0)  // 绿
		gradPat.AddColorStopRGB(0.6, 0, 1, 1)  // 青
		gradPat.AddColorStopRGB(0.8, 0, 0, 1)  // 蓝
		gradPat.AddColorStopRGB(1.0, 1, 0, 1)  // 洋红
	}
	ctx.SetSource(pattern4)
	ctx.Rectangle(300, 50, 400, 100)
	ctx.Fill()
	pattern4.Destroy()
	fmt.Println("   ✓ 绘制完成")

	// 测试5: 径向渐变 (从中心向外)
	fmt.Println("\n⭕ 测试5: 径向渐变 (中心白色->边缘红色)")
	pattern5 := cairo.NewPatternRadial(400, 275, 10, 400, 275, 80)
	if gradPat, ok := pattern5.(cairo.RadialGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 1, 1)    // 白色
		gradPat.AddColorStopRGB(1, 1, 0, 0)    // 红色
	}
	ctx.SetSource(pattern5)
	ctx.Arc(400, 275, 80, 0, 2*math.Pi)
	ctx.Fill()
	pattern5.Destroy()
	fmt.Println("   ✓ 绘制完成")

	// 测试6: 径向渐变 (偏心)
	fmt.Println("\n⭕ 测试6: 偏心径向渐变 (光照效果)")
	pattern6 := cairo.NewPatternRadial(580, 275, 5, 600, 275, 80)
	if gradPat, ok := pattern6.(cairo.RadialGradientPattern); ok {
		gradPat.AddColorStopRGBA(0, 1, 1, 0.8, 1)  // 浅黄
		gradPat.AddColorStopRGBA(0.5, 1, 0.5, 0, 1) // 橙色
		gradPat.AddColorStopRGBA(1, 0.5, 0, 0, 1)   // 深红
	}
	ctx.SetSource(pattern6)
	ctx.Arc(600, 275, 80, 0, 2*math.Pi)
	ctx.Fill()
	pattern6.Destroy()
	fmt.Println("   ✓ 绘制完成")

	// 测试7: 带透明度的渐变
	fmt.Println("\n📊 测试7: 透明度渐变 (不透明->透明)")
	// 先画一个彩色背景
	ctx.SetSourceRGB(0.9, 0.9, 0.9)
	ctx.Rectangle(300, 380, 200, 100)
	ctx.Fill()
	
	pattern7 := cairo.NewPatternLinear(300, 380, 500, 380)
	if gradPat, ok := pattern7.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGBA(0, 0, 0, 1, 1)    // 不透明蓝色
		gradPat.AddColorStopRGBA(1, 0, 0, 1, 0)    // 透明蓝色
	}
	ctx.SetSource(pattern7)
	ctx.Rectangle(300, 380, 200, 100)
	ctx.Fill()
	pattern7.Destroy()
	fmt.Println("   ✓ 绘制完成")

	// 测试8: 圆形渐变填充
	fmt.Println("\n⭕ 测试8: 圆形多色径向渐变")
	pattern8 := cairo.NewPatternRadial(400, 450, 0, 400, 450, 60)
	if gradPat, ok := pattern8.(cairo.RadialGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 1, 1)      // 白色中心
		gradPat.AddColorStopRGB(0.3, 1, 1, 0)    // 黄色
		gradPat.AddColorStopRGB(0.6, 1, 0.5, 0)  // 橙色
		gradPat.AddColorStopRGB(1, 1, 0, 0)      // 红色边缘
	}
	ctx.SetSource(pattern8)
	ctx.Arc(400, 450, 60, 0, 2*math.Pi)
	ctx.Fill()
	pattern8.Destroy()
	fmt.Println("   ✓ 绘制完成")

	// 测试9: 渐变描边
	fmt.Println("\n📊 测试9: 渐变描边效果")
	pattern9 := cairo.NewPatternLinear(550, 380, 750, 480)
	if gradPat, ok := pattern9.(cairo.LinearGradientPattern); ok {
		gradPat.AddColorStopRGB(0, 1, 0, 0)      // 红
		gradPat.AddColorStopRGB(0.5, 0, 1, 0)    // 绿
		gradPat.AddColorStopRGB(1, 0, 0, 1)      // 蓝
	}
	ctx.SetSource(pattern9)
	ctx.SetLineWidth(10)
	ctx.Rectangle(560, 390, 180, 80)
	ctx.Stroke()
	pattern9.Destroy()
	fmt.Println("   ✓ 绘制完成")

	// 添加标题文字
	fmt.Println("\n🔤 添加标题...")
	ctx.SetSourceRGB(0, 0, 0)
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("PingFang SC")  // 使用支持中文的字体
	fontDesc.SetSize(24)
	layout.SetFontDescription(fontDesc)
	
	layout.SetText("Cairo 渐变测试")
	extents := layout.GetPixelExtents()
	ctx.MoveTo(400-extents.Width/2, 20)
	ctx.PangoCairoShowText(layout)

	// 保存图片
	fmt.Println("\n💾 保存图片...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("gradient_test.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("保存失败: %v", status))
		}
		fmt.Println("✅ 渐变测试图片已保存到 gradient_test.png")
	}

	fmt.Println("🎉 渐变测试完成!")
}
