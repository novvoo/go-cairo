//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"math"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func drawRoundedRect(ctx cairo.Context, x, y, width, height, radius float64) {
	ctx.NewPath()
	ctx.Arc(x+radius, y+radius, radius, math.Pi, 1.5*math.Pi)
	ctx.Arc(x+width-radius, y+radius, radius, 1.5*math.Pi, 2*math.Pi)
	ctx.Arc(x+width-radius, y+height-radius, radius, 0, 0.5*math.Pi)
	ctx.Arc(x+radius, y+height-radius, radius, 0.5*math.Pi, math.Pi)
	ctx.ClosePath()
}

func main() {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 500, 300)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 白色背景
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// 绘制按钮
	rectX, rectY := 150.0, 100.0
	rectWidth, rectHeight := 200.0, 100.0
	rectRadius := 15.0

	drawRoundedRect(ctx, rectX, rectY, rectWidth, rectHeight, rectRadius)
	ctx.SetSourceRGB(0.2, 0.6, 0.8)
	ctx.Fill()

	// 创建布局
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("Go Regular")
	fontDesc.SetWeight(cairo.PangoWeightBold)
	fontDesc.SetSize(48.0)
	layout.SetFontDescription(fontDesc)

	// 获取文本尺寸
	layout.SetText("M")
	mExtents := layout.GetPixelExtents()
	layout.SetText("I")
	iExtents := layout.GetPixelExtents()
	
	letterSpacing := 8.0
	totalWidth := mExtents.Width + iExtents.Width + letterSpacing

	// 计算居中位置
	fontExtents := layout.GetFontExtents()
	midx := rectX + rectWidth/2 - totalWidth/2
	midy := rectY + rectHeight/2 + fontExtents.Ascent/2 - fontExtents.Descent/2

	// 渲染 M
	layout.SetText("M")
	ctx.SetSourceRGB(1.0, 1.0, 1.0)
	ctx.MoveTo(midx, midy)
	ctx.PangoCairoShowText(layout)

	// 绘制 M 的边界框（红色）
	ctx.SetSourceRGBA(1.0, 0.0, 0.0, 0.5)
	ctx.SetLineWidth(2)
	ctx.Rectangle(midx+mExtents.X, midy+mExtents.Y, mExtents.Width, mExtents.Height)
	ctx.Stroke()

	// 渲染 I
	iStartX := midx + mExtents.Width + letterSpacing
	layout.SetText("I")
	ctx.SetSourceRGB(1.0, 1.0, 1.0)
	ctx.MoveTo(iStartX, midy)
	ctx.PangoCairoShowText(layout)

	// 绘制 I 的边界框（绿色）
	ctx.SetSourceRGBA(0.0, 1.0, 0.0, 0.5)
	ctx.SetLineWidth(2)
	ctx.Rectangle(iStartX+iExtents.X, midy+iExtents.Y, iExtents.Width, iExtents.Height)
	ctx.Stroke()

	// 保存图片
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("mi_with_bounds.png")
		if status != cairo.StatusSuccess {
			fmt.Printf("保存PNG失败: %v\n", status)
		} else {
			fmt.Println("✅ 已保存到 mi_with_bounds.png")
			fmt.Printf("M边界框: (%.2f, %.2f) 宽%.2f 高%.2f\n", midx+mExtents.X, midy+mExtents.Y, mExtents.Width, mExtents.Height)
			fmt.Printf("I边界框: (%.2f, %.2f) 宽%.2f 高%.2f\n", iStartX+iExtents.X, midy+iExtents.Y, iExtents.Width, iExtents.Height)
			fmt.Printf("M右边界: %.2f, I左边界: %.2f, 间隙: %.2f\n",
				midx+mExtents.X+mExtents.Width,
				iStartX+iExtents.X,
				(iStartX+iExtents.X)-(midx+mExtents.X+mExtents.Width))
		}
	}
}
