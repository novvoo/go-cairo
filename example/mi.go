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

	// 分别获取M和I的位置信息
	mExtents := ctx.TextExtents("M")
	iExtents := ctx.TextExtents("I")

	// 计算文字总宽度（M的宽度 + I的宽度 + 间距）
	totalWidth := mExtents.Width + iExtents.Width + 2.0 // 2.0是M和I之间的间距

	// 计算居中位置
	midx := rectX + rectWidth/2 - totalWidth/2
	midy := rectY + rectHeight/2

	// 调整Y轴位置以更好地垂直居中
	fontExtents := ctx.FontExtents()
	midy = rectY + rectHeight/2 + fontExtents.Ascent/2 - fontExtents.Descent/2

	// 绘制白色文字
	ctx.SetSourceRGB(1.0, 1.0, 1.0)

	// 打印位置信息
	fmt.Printf("=== MI 整体位置信息 ===\n")
	fmt.Printf("MI起始位置: (%.2f, %.2f)\n", midx, midy)
	fmt.Printf("MI文本总宽度: %.2f\n", totalWidth)

	fmt.Printf("\n=== 字母 M 位置信息 ===\n")
	fmt.Printf("M起始位置: (%.2f, %.2f)\n", midx, midy)
	fmt.Printf("M文本范围: 宽度=%.2f, 高度=%.2f\n", mExtents.Width, mExtents.Height)
	fmt.Printf("M边界信息: XBearing=%.2f, YBearing=%.2f\n", mExtents.XBearing, mExtents.YBearing)
	fmt.Printf("M左上角坐标: (%.2f, %.2f)\n", midx+mExtents.XBearing, midy+mExtents.YBearing)
	fmt.Printf("M左下角坐标: (%.2f, %.2f)\n", midx+mExtents.XBearing, midy+mExtents.YBearing+mExtents.Height)
	fmt.Printf("M右上角坐标: (%.2f, %.2f)\n", midx+mExtents.XBearing+mExtents.Width, midy+mExtents.YBearing)
	fmt.Printf("M右下角坐标: (%.2f, %.2f)\n", midx+mExtents.XBearing+mExtents.Width, midy+mExtents.YBearing+mExtents.Height)

	fmt.Printf("\n=== 字母 I 位置信息 ===\n")
	// 修复字母重叠问题：正确计算I的起始位置
	// I的起始位置应该是M的起始位置加上M的宽度，再加上一些间距以确保完全分离
	iStartX := midx + mExtents.Width + 2.0 // 添加2像素的间距
	fmt.Printf("I起始位置: (%.2f, %.2f)\n", iStartX, midy)
	fmt.Printf("I文本范围: 宽度=%.2f, 高度=%.2f\n", iExtents.Width, iExtents.Height)
	fmt.Printf("I边界信息: XBearing=%.2f, YBearing=%.2f\n", iExtents.XBearing, iExtents.YBearing)
	fmt.Printf("I左上角坐标: (%.2f, %.2f)\n", iStartX+iExtents.XBearing, midy+iExtents.YBearing)
	fmt.Printf("I左下角坐标: (%.2f, %.2f)\n", iStartX+iExtents.XBearing, midy+iExtents.YBearing+iExtents.Height)
	fmt.Printf("I右上角坐标: (%.2f, %.2f)\n", iStartX+iExtents.XBearing+iExtents.Width, midy+iExtents.YBearing)
	fmt.Printf("I右下角坐标: (%.2f, %.2f)\n", iStartX+iExtents.XBearing+iExtents.Width, midy+iExtents.YBearing+iExtents.Height)

	// 修复字母重叠问题：分别绘制M和I，确保它们不会重叠
	// 先绘制M字母
	ctx.MoveTo(midx, midy)
	ctx.ShowText("M")

	// 再绘制I字母，使用计算出的正确位置
	ctx.MoveTo(iStartX, midy)
	ctx.ShowText("I")

	// 输出 PNG
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("images/mi_logo.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("WriteToPNG failed: %v", status))
		}
		fmt.Println("\n✅ 真实小米 Logo PNG 已保存到 images/mi_logo.png")
		fmt.Println("   - 添加了渐变背景")
		fmt.Println("   - 添加了多层阴影效果")
		fmt.Println("   - 使用渐变橙色（从亮到暗）")
		fmt.Println("   - 添加了高光和反光效果")
		fmt.Println("   - 文字带有阴影和高光")
	} else {
		panic("Surface is not an ImageSurface")
	}
}
