package main

import (
	"fmt"
	"math"

	"go-cairo/pkg/cairo"
)

func main() {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 400, 200)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// === 1. 白色背景
	ctx.SetSourceRGB(1.0, 1.0, 1.0)
	if err := ctx.Paint(); err != nil {
		panic(err)
	}

	// === 2. 小米橙色圆角矩形 (MI logo 背景)
	ctx.SetSourceRGB(1.0, 0.41, 0.0) // #FF6700
	radius := 20.0
	rectX, rectY := 80.0, 40.0
	rectWidth, rectHeight := 240.0, 120.0
	// 圆角矩形路径
	ctx.NewPath()
	ctx.Arc(rectX+radius, rectY+radius, radius, math.Pi, 1.5*math.Pi)                // 左上
	ctx.Arc(rectX+rectWidth-radius, rectY+radius, radius, 1.5*math.Pi, math.Pi)      // 右上
	ctx.Arc(rectX+rectWidth-radius, rectY+rectHeight-radius, radius, 0, 0.5*math.Pi) // 右下
	ctx.Arc(rectX+radius, rectY+rectHeight-radius, radius, 0.5*math.Pi, 0)           // 左下
	ctx.ClosePath()
	if err := ctx.Fill(); err != nil {
		panic(err)
	}

	// === 3. 白色粗体 "MI" 文字精确居中 (使用extents)
	ctx.SelectFontFace("serif", cairo.FontSlantNormal, cairo.FontWeightBold)
	ctx.SetFontSize(90)
	ctx.SetAntialias(cairo.AntialiasSubpixel)
	extents := ctx.TextExtents("MI")
	midx := rectX + rectWidth/2 - extents.Width/2 - extents.XBearing
	midy := rectY + rectHeight/2 - extents.Height/2 - extents.YBearing
	ctx.SetSourceRGB(1.0, 1.0, 1.0) // 白色
	ctx.MoveTo(midx, midy)
	ctx.ShowText("MI")

	// 输出 PNG
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("images/xiaomi_logo.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("WriteToPNG failed: %v", status))
		}
		fmt.Println("MI Logo PNG saved successfully to images/xiaomi_logo.png")
	} else {
		panic("Surface is not an ImageSurface")
	}
}
