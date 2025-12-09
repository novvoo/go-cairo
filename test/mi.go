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
		bg.AddColorStopRGB(1, 1, 1, 1)
		ctx.SetSource(bg)
	}
	ctx.Paint()

	// === 2. 半透明黑色蒙版
	ctx.SetSourceRGBA(0, 0, 0, 0.2)
	ctx.Rectangle(0, 0, 500, 300)
	ctx.Fill()

	// === 3. 圆角矩形按钮
	rectX, rectY := 150.0, 100.0
	rectWidth, rectHeight := 200.0, 100.0
	rectRadius := 15.0

	// 绘制阴影效果
	ctx.SetSourceRGBA(0, 0, 0, 0.3)
	drawRoundedRect(ctx, rectX+3, rectY+3, rectWidth, rectHeight, rectRadius)
	ctx.Fill()

	// 绘制主按钮
	drawRoundedRect(ctx, rectX, rectY, rectWidth, rectHeight, rectRadius)
	// 渐变填充
	buttonGradient := cairo.NewPatternLinear(rectX, rectY, rectX, rectY+rectHeight)
	if grad, ok := buttonGradient.(cairo.LinearGradientPattern); ok {
		grad.AddColorStopRGB(0, 0.2, 0.6, 0.8) // 蓝色顶部
		grad.AddColorStopRGB(1, 0.1, 0.4, 0.6) // 深蓝色底部
		ctx.SetSource(grad)
	}
	ctx.FillPreserve()

	// 边框
	ctx.SetSourceRGBA(1, 1, 1, 0.8)
	ctx.SetLineWidth(2)
	ctx.Stroke()

	// === 4. 使用PangoCairo渲染"MI"文字
	ctx.SelectFontFace("Go Regular", cairo.FontSlantNormal, cairo.FontWeightBold)
	ctx.SetFontSize(48)

	// 获取文本尺寸
	mExtents := ctx.TextExtents("M")
	iExtents := ctx.TextExtents("I")
	totalWidth := mExtents.Width + iExtents.Width + 2.0 // 加上间距

	// 计算居中位置
	midx := rectX + rectWidth/2 - totalWidth/2
	midy := rectY + rectHeight/2 + ctx.FontExtents().Ascent/2 - ctx.FontExtents().Descent/2

	// 使用PangoCairo创建布局
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)

	// 设置字体描述
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("Go Regular")
	fontDesc.SetWeight(cairo.PangoWeightBold)
	fontDesc.SetSize(48.0)
	layout.SetFontDescription(fontDesc)

	// 渲染"M"
	layout.SetText("M")
	ctx.SetSourceRGB(1.0, 1.0, 1.0)
	ctx.MoveTo(midx, midy)
	ctx.PangoCairoShowText(layout)

	// 渲染"I"，确保不会重叠
	iStartX := midx + mExtents.Width + 2.0
	layout.SetText("I")
	ctx.MoveTo(iStartX, midy)
	ctx.PangoCairoShowText(layout)

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
	iStartX = midx + mExtents.Width + 2.0 // 添加2像素的间距
	fmt.Printf("I起始位置: (%.2f, %.2f)\n", iStartX, midy)
	fmt.Printf("I文本范围: 宽度=%.2f, 高度=%.2f\n", iExtents.Width, iExtents.Height)
	fmt.Printf("I边界信息: XBearing=%.2f, YBearing=%.2f\n", iExtents.XBearing, iExtents.YBearing)
	fmt.Printf("I左上角坐标: (%.2f, %.2f)\n", iStartX+iExtents.XBearing, midy+iExtents.YBearing)
	fmt.Printf("I左下角坐标: (%.2f, %.2f)\n", iStartX+iExtents.XBearing, midy+iExtents.YBearing+iExtents.Height)
	fmt.Printf("I右上角坐标: (%.2f, %.2f)\n", iStartX+iExtents.XBearing+iExtents.Width, midy+iExtents.YBearing)
	fmt.Printf("I右下角坐标: (%.2f, %.2f)\n", iStartX+iExtents.XBearing+iExtents.Width, midy+iExtents.YBearing+iExtents.Height)

	// === 5. 保存图片
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("mi_pango.png")
		if status != cairo.StatusSuccess {
			fmt.Printf("保存PNG失败: %v\n", status)
		} else {
			fmt.Println("\n✅ MI示例已保存到 mi_pango.png")
		}
	}
}
