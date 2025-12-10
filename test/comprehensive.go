//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"math"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("🚀 Starting comprehensive Cairo demo...")

	// Create a new image surface
	fmt.Println("📝 Creating image surface (400x400 pixels)...")
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 400, 400)
	defer surface.Destroy()
	fmt.Printf("   Surface created with status: %v\n", surface.Status())

	// Create a context
	fmt.Println("✏️  Creating Cairo context...")
	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()
	fmt.Printf("   Context created with status: %v\n", ctx.Status())

	// 检查初始变换矩阵
	matrix := ctx.GetMatrix()
	fmt.Printf("   Initial matrix: XX=%.4f, YY=%.4f, XY=%.4f, YX=%.4f, X0=%.4f, Y0=%.4f\n",
		matrix.XX, matrix.YY, matrix.XY, matrix.YX, matrix.X0, matrix.Y0)

	// Set background to white
	fmt.Println("🎨 Setting background to white...")
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()
	fmt.Println("   Background painted")

	// Test 1: Basic shapes at different coordinates
	fmt.Println("🔷 Drawing basic colored rectangles...")
	ctx.SetSourceRGB(1, 0, 0) // Red
	fmt.Println("   Drawing red rectangle at (50, 50)")
	ctx.Rectangle(50, 50, 30, 30) // Top-left quadrant
	ctx.Fill()

	ctx.SetSourceRGB(0, 1, 0) // Green
	fmt.Println("   Drawing green rectangle at (320, 50)")
	ctx.Rectangle(320, 50, 30, 30) // Top-right quadrant
	ctx.Fill()

	ctx.SetSourceRGB(0, 0, 1) // Blue
	fmt.Println("   Drawing blue rectangle at (50, 320)")
	ctx.Rectangle(50, 320, 30, 30) // Bottom-left quadrant
	ctx.Fill()

	ctx.SetSourceRGB(1, 1, 0) // Yellow
	fmt.Println("   Drawing yellow rectangle at (320, 320)")
	ctx.Rectangle(320, 320, 30, 30) // Bottom-right quadrant
	ctx.Fill()

	// Test 2: Lines to show coordinate system orientation
	fmt.Println("📏 Drawing coordinate system diagonals...")
	ctx.SetSourceRGB(0, 0, 0) // Black
	ctx.SetLineWidth(2)

	// Diagonal from top-left to bottom-right
	fmt.Println("   Drawing diagonal from (0,0) to (400,400)")
	ctx.MoveTo(0, 0)
	ctx.LineTo(400, 400)
	ctx.Stroke()

	// Diagonal from bottom-left to top-right
	fmt.Println("   Drawing diagonal from (0,400) to (400,0)")
	ctx.MoveTo(0, 400)
	ctx.LineTo(400, 0)
	ctx.Stroke()

	// Test 3: Arcs and circles
	fmt.Println("⭕ Drawing circle at center...")

	// 检查绘制圆形前的变换矩阵
	matrix = ctx.GetMatrix()
	fmt.Printf("   Before circle - Matrix: XX=%.4f, YY=%.4f, XY=%.4f, YX=%.4f\n",
		matrix.XX, matrix.YY, matrix.XY, matrix.YX)

	// 检查设备到用户空间的转换
	devX1, devY1 := 200.0, 200.0
	userX1, userY1 := ctx.DeviceToUser(devX1, devY1)
	fmt.Printf("   Device (%.1f, %.1f) -> User (%.1f, %.1f)\n", devX1, devY1, userX1, userY1)

	// 检查用户到设备空间的转换
	userX2, userY2 := 200.0, 200.0
	devX2, devY2 := ctx.UserToDevice(userX2, userY2)
	fmt.Printf("   User (%.1f, %.1f) -> Device (%.1f, %.1f)\n", userX2, userY2, devX2, devY2)

	ctx.SetSourceRGB(1, 0, 1) // Magenta
	ctx.SetLineWidth(3)
	// Use DrawCircle for better precision
	ctx.DrawCircle(200, 200, 50) // Circle at center
	ctx.Stroke()
	fmt.Println("   Circle drawn at (200, 200) with radius 50 using DrawCircle")

	// Test 4: Text rendering using PangoCairo
	fmt.Println("🔤 Drawing text samples...")
	ctx.SetSourceRGB(0, 0, 0) // Black

	// Create PangoCairo layout
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)

	// Create font description with size 18
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("Go Regular")
	fontDesc.SetWeight(cairo.PangoWeightNormal)
	fontDesc.SetSize(18)
	layout.SetFontDescription(fontDesc)

	// Text at top-left
	fmt.Println("   Drawing 'Top Left' at (10, 20)")
	layout.SetText("Top Left")
	ctx.MoveTo(10, 20)
	ctx.PangoCairoShowText(layout)

	// Text at top-right
	fmt.Println("   Drawing 'Top Right' at right-aligned position")
	text := "Top Right"
	layout.SetText(text)
	extents := layout.GetPixelExtents()
	ctx.MoveTo(400-extents.Width-10, 20)
	ctx.PangoCairoShowText(layout)

	// Text at bottom-left
	fmt.Println("   Drawing 'Bottom Left' at bottom position")
	text = "Bottom Left"
	layout.SetText(text)
	extents = layout.GetPixelExtents()
	ctx.MoveTo(10, 390)
	ctx.PangoCairoShowText(layout)

	// Text at bottom-right
	fmt.Println("   Drawing 'Bottom Right' at bottom-right position")
	text = "Bottom Right"
	layout.SetText(text)
	extents = layout.GetPixelExtents()
	ctx.MoveTo(400-extents.Width-10, 390)
	ctx.PangoCairoShowText(layout)

	// Text at center with larger font
	fmt.Println("   Drawing 'Center' at centered position")
	fontDesc.SetSize(42)
	layout.SetFontDescription(fontDesc)
	text = "Center"
	layout.SetText(text)
	extents = layout.GetPixelExtents()
	fontExtents := layout.GetFontExtents()

	// 正确的居中计算：让文字的真实视觉中心位于图像中心
	// 打印调试信息
	fmt.Printf("\n🔍 调试 'Center' 文字宽度:\n")
	fmt.Printf("   extents.Width = %.2f\n", extents.Width)
	fmt.Printf("   extents.XBearing = %.2f\n", extents.X)
	fmt.Printf("   extents.Height = %.2f\n", extents.Height)
	fmt.Printf("   extents.YBearing = %.2f\n", extents.Y)

	// X轴：图像中心 - 文字宽度的一半 = 文字左边界
	// 但需要考虑 XBearing（左侧空白）
	x := 200 - extents.Width/2 - extents.X
	// Y轴：图像中心 + (Ascent - Descent) / 2 = 基线位置
	y := 200 + (fontExtents.Ascent-fontExtents.Descent)/2

	fmt.Printf("   计算的 x = %.2f (200 - %.2f/2 - %.2f)\n", x, extents.Width, extents.X)
	fmt.Printf("   计算的 y = %.2f\n\n", y)

	ctx.MoveTo(x, y)
	ctx.PangoCairoShowText(layout)

	// 打印图像中心和文字中心位置
	fmt.Println("\n📍 位置对比分析:")
	imageCenterX := 400.0 / 2
	imageCenterY := 400.0 / 2
	fmt.Printf("   图像中心: (%.2f, %.2f)\n", imageCenterX, imageCenterY)

	// 计算文字的中心位置
	// X轴：文字起点 + 宽度的一半
	textCenterX := x + extents.Width/2
	// Y轴：文字的真实视觉中心 = (顶部 + 底部) / 2
	textTop := y - fontExtents.Ascent
	textBottom := y + fontExtents.Descent
	textCenterY := (textTop + textBottom) / 2

	fmt.Printf("   'Center' 文字绘制起点(基线): (%.2f, %.2f)\n", x, y)
	fmt.Printf("   'Center' 文字中心: (%.2f, %.2f)\n", textCenterX, textCenterY)
	fmt.Printf("   'Center' 文字尺寸: 宽度=%.2f, Ascent=%.2f, 总高度=%.2f\n",
		extents.Width, fontExtents.Ascent, fontExtents.Height)

	// 详细分析
	fmt.Printf("\n   详细分析:\n")
	fmt.Printf("   - 文字左边界: %.2f\n", x)
	fmt.Printf("   - 文字右边界: %.2f\n", x+extents.Width)
	fmt.Printf("   - 文字顶部(基线-Ascent): %.2f\n", y-fontExtents.Ascent)
	fmt.Printf("   - 文字基线: %.2f\n", y)
	fmt.Printf("   - 文字底部(基线+Descent): %.2f\n", y+fontExtents.Descent)

	// 计算偏差
	deltaX := math.Abs(imageCenterX - textCenterX)
	deltaY := math.Abs(imageCenterY - textCenterY)
	fmt.Printf("\n   偏差分析:\n")
	fmt.Printf("   - X轴偏差: %.2f 像素\n", deltaX)
	fmt.Printf("   - Y轴偏差: %.2f 像素\n", deltaY)
	fmt.Printf("   - 总偏差: %.2f 像素\n", math.Sqrt(deltaX*deltaX+deltaY*deltaY))

	if deltaX < 1 && deltaY < 1 {
		fmt.Println("   ✅ 文字中心与图像中心基本一致!")
	} else {
		fmt.Println("   ⚠️  文字中心与图像中心存在偏差")
		fmt.Printf("   说明: 'Center' 的视觉中心应该在 'nt' 两个字母之间\n")
	}

	// Test 5: Bezier curves
	fmt.Println("➰ Drawing bezier curve...")
	ctx.SetSourceRGB(0, 1, 1) // Cyan
	ctx.SetLineWidth(4)
	fmt.Println("   起点: (100, 100)")
	ctx.MoveTo(100, 100)
	fmt.Println("   控制点1: (150, 20), 控制点2: (250, 380), 终点: (300, 300)")
	ctx.CurveTo(150, 20, 250, 380, 300, 300)
	ctx.Stroke()

	// Save to PNG
	fmt.Println("💾 Saving image to PNG...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("comprehensive_test.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("WriteToPNG failed: %v", status))
		}
		fmt.Println("✅ Comprehensive test image saved to comprehensive_test.png")
	} else {
		panic("Surface is not an ImageSurface")
	}

	fmt.Println("🎉 Comprehensive Cairo demo completed!")
}
