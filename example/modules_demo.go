//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"image/color"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	// 创建 800x600 的图像表面
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 800, 600)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 1. 测试 Pixman 图像后端
	fmt.Println("测试 Pixman 图像后端...")
	pixmanImg := cairo.NewPixmanImage(cairo.PixmanFormatARGB32, 200, 200)
	pixmanImg.Fill(50, 50, 100, 100, color.NRGBA{R: 255, G: 100, B: 50, A: 255})
	fmt.Println("✓ Pixman 图像操作完成")

	// 2. 测试 Porter-Duff 混合模式
	fmt.Println("\n测试 Porter-Duff 混合模式...")

	// 绘制背景
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// 测试不同的混合模式
	operators := []cairo.Operator{
		cairo.OperatorOver,
		cairo.OperatorMultiply,
		cairo.OperatorScreen,
		cairo.OperatorOverlay,
	}

	x := 50.0
	for i, op := range operators {
		ctx.Save()
		ctx.SetOperator(op)

		// 绘制红色圆
		ctx.SetSourceRGBA(1, 0, 0, 0.7)
		ctx.Arc(x, 100, 40, 0, 2*3.14159)
		ctx.Fill()

		// 绘制蓝色圆
		ctx.SetSourceRGBA(0, 0, 1, 0.7)
		ctx.Arc(x+30, 100, 40, 0, 2*3.14159)
		ctx.Fill()

		ctx.Restore()
		x += 150

		fmt.Printf("✓ 混合模式 %d 完成\n", i+1)
	}

	// 3. 测试颜色空间转换
	fmt.Println("\n测试颜色空间转换...")

	// RGB -> HSL -> RGB
	r, g, b := 0.8, 0.3, 0.5
	h, s, l := cairo.RgbToHSL(r, g, b)
	r2, g2, b2 := cairo.HslToRGB(h, s, l)
	fmt.Printf("RGB(%.2f, %.2f, %.2f) -> HSL(%.2f, %.2f, %.2f) -> RGB(%.2f, %.2f, %.2f)\n",
		r, g, b, h, s, l, r2, g2, b2)

	// 绘制 HSL 色轮
	centerX, centerY := 400.0, 350.0
	radius := 80.0

	for angle := 0.0; angle < 360; angle += 5 {
		hue := angle / 360.0
		rr, gg, bb := cairo.HslToRGB(hue, 1.0, 0.5)

		ctx.SetSourceRGB(rr, gg, bb)
		ctx.MoveTo(centerX, centerY)
		ctx.Arc(centerX, centerY, radius,
			angle*3.14159/180, (angle+5)*3.14159/180)
		ctx.ClosePath()
		ctx.Fill()
	}
	fmt.Println("✓ HSL 色轮绘制完成")

	// 4. 测试高级光栅化器
	fmt.Println("\n测试高级光栅化器...")

	// 绘制平滑的贝塞尔曲线
	ctx.SetSourceRGB(0, 0, 0)
	ctx.SetLineWidth(3)
	ctx.MoveTo(50, 400)
	ctx.CurveTo(150, 300, 250, 500, 350, 400)
	ctx.Stroke()
	fmt.Println("✓ 贝塞尔曲线光栅化完成")

	// 5. 测试图像后端
	fmt.Println("\n测试图像后端...")
	backend := cairo.NewImageBackend(200, 100)
	backend.Clear(color.RGBA{R: 240, G: 240, B: 240, A: 255})
	backend.FillRect(20, 20, 160, 60, color.RGBA{R: 100, G: 150, B: 200, A: 255})
	fmt.Println("✓ 图像后端操作完成")

	// 保存结果
	imgSurface := surface.(cairo.ImageSurface)
	status := imgSurface.WriteToPNG("modules_demo.png")
	if status == cairo.StatusSuccess {
		fmt.Println("\n✓ 图像已保存到 modules_demo.png")
	} else {
		fmt.Printf("\n✗ 保存失败: %v\n", status)
	}

	// 输出模块状态
	fmt.Println("\n=== Go-Cairo 模块状态 ===")
	fmt.Println("✓ Pixman (图像后端)")
	fmt.Println("✓ Rasterizer (光栅化器)")
	fmt.Println("✓ Alpha Blend (Porter-Duff 混合)")
	fmt.Println("✓ Colorspace (颜色空间转换)")
	fmt.Println("========================")
}
