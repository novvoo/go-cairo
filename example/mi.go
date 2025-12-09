package main

import (
	"fmt"
	"math"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 绘制圆角矩形路径
func drawRoundedRect(ctx cairo.Context, x, y, width, height, radius float64) {
	ctx.NewPath()
	ctx.Arc(x+radius, y+radius, radius, math.Pi, 1.5*math.Pi)         // 左上
	ctx.Arc(x+width-radius, y+radius, radius, 1.5*math.Pi, 2*math.Pi) // 右上
	ctx.Arc(x+width-radius, y+height-radius, radius, 0, 0.5*math.Pi)  // 右下
	ctx.Arc(x+radius, y+height-radius, radius, 0.5*math.Pi, math.Pi)  // 左下
	ctx.ClosePath()
}

func main() {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 500, 300)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// === 1. 渐变背景 (淡灰色到白色)
	bgGradient := cairo.NewPatternLinear(0, 0, 0, 300)
	if bg, ok := bgGradient.(cairo.LinearGradientPattern); ok {
		bg.AddColorStopRGB(0, 0.95, 0.95, 0.95)
		bg.AddColorStopRGB(1, 1.0, 1.0, 1.0)
	}
	ctx.SetSource(bgGradient)
	if err := ctx.Paint(); err != nil {
		panic(err)
	}
	bgGradient.Destroy()

	// Logo 参数
	radius := 18.0
	rectX, rectY := 125.0, 75.0
	rectWidth, rectHeight := 250.0, 150.0

	// === 2. 阴影效果 (多层模糊阴影)
	ctx.Save()
	// 外层大阴影
	shadowOffset := 8.0
	drawRoundedRect(ctx, rectX+shadowOffset, rectY+shadowOffset, rectWidth, rectHeight, radius)
	ctx.SetSourceRGBA(0, 0, 0, 0.15)
	if err := ctx.Fill(); err != nil {
		panic(err)
	}
	// 内层小阴影（更暗）
	shadowOffset = 3.0
	drawRoundedRect(ctx, rectX+shadowOffset, rectY+shadowOffset, rectWidth, rectHeight, radius)
	ctx.SetSourceRGBA(0, 0, 0, 0.25)
	if err := ctx.Fill(); err != nil {
		panic(err)
	}
	ctx.Restore()

	// === 3. 小米橙色矩形 (标准色 #FF6700)
	drawRoundedRect(ctx, rectX, rectY, rectWidth, rectHeight, radius)
	// 小米橙色 #FF6700 = RGB(255, 103, 0) = (1.0, 0.404, 0.0)
	ctx.SetSourceRGB(1.0, 0.404, 0.0)
	if err := ctx.Fill(); err != nil {
		panic(err)
	}

	// === 4. 顶部高光效果 (轻微)
	ctx.Save()
	highlightGradient := cairo.NewPatternLinear(rectX, rectY, rectX, rectY+rectHeight*0.4)
	if hl, ok := highlightGradient.(cairo.LinearGradientPattern); ok {
		hl.AddColorStopRGBA(0, 1.0, 1.0, 1.0, 0.15)
		hl.AddColorStopRGBA(1, 1.0, 1.0, 1.0, 0.0)
	}

	drawRoundedRect(ctx, rectX, rectY, rectWidth, rectHeight*0.4, radius)
	ctx.SetSource(highlightGradient)
	if err := ctx.Fill(); err != nil {
		panic(err)
	}
	highlightGradient.Destroy()
	ctx.Restore()

	// === 6. 白色粗体 "MI" 文字 (已修复字体渲染堆叠问题)
	ctx.SelectFontFace("sans-serif", cairo.FontSlantNormal, cairo.FontWeightBold)
	ctx.SetFontSize(110)
	ctx.SetAntialias(cairo.AntialiasSubpixel)

	// 计算文字居中位置
	extents := ctx.TextExtents("MI")
	midx := rectX + rectWidth/2 - extents.Width/2 - extents.XBearing
	midy := rectY + rectHeight/2 - extents.Height/2 - extents.YBearing

	// 绘制白色文字
	ctx.SetSourceRGB(1.0, 1.0, 1.0)
	ctx.MoveTo(midx, midy)
	ctx.ShowText("MI")

	// 输出 PNG
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("images/mi_logo.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("WriteToPNG failed: %v", status))
		}
		fmt.Println("✅ 真实小米 Logo PNG 已保存到 images/mi_logo.png")
		fmt.Println("   - 添加了渐变背景")
		fmt.Println("   - 添加了多层阴影效果")
		fmt.Println("   - 使用渐变橙色（从亮到暗）")
		fmt.Println("   - 添加了高光和反光效果")
		fmt.Println("   - 文字带有阴影和高光")
	} else {
		panic("Surface is not an ImageSurface")
	}
}
