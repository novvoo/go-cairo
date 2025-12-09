package main

import (
	"fmt"
	"math"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// SetFont 是一个辅助函数，用于设置字体系列和大小
func SetFont(ctx cairo.Context, face string, size float64) {
	ctx.SelectFontFace(face, cairo.FontSlantNormal, cairo.FontWeightNormal)
	ctx.SetFontSize(size)
	// 强制初始化 scaled font
	_ = ctx.GetScaledFont()
}

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
	ctx.SetSourceRGB(1, 0, 1)           // Magenta
	ctx.Arc(200, 200, 50, 0, 2*math.Pi) // Circle at center
	ctx.Stroke()
	fmt.Println("   Circle drawn at (200, 200) with radius 50")

	// Test 4: Text at different positions (优化文本显示)
	fmt.Println("🔤 Drawing text samples...")
	// 使用 "Go Regular" 字体而不是 "sans-serif"
	SetFont(ctx, "Go Regular", 18)

	// 手动触发一次 ScaledFont 创建
	_ = ctx.GetScaledFont()

	ctx.SetSourceRGB(0, 0, 0) // Black

	// Text at top-left
	fmt.Println("   Drawing 'Top Left' at (10, 30)")
	ctx.MoveTo(10, 30)
	ctx.ShowText("Top Left")

	// Text at top-right (手动计算位置)
	fmt.Println("   Drawing 'Top Right' at manually calculated position")
	text := "Top Right"
	extents := ctx.TextExtents(text)
	ctx.MoveTo(400-extents.XAdvance-10, 30)
	ctx.ShowText(text)

	// Text at bottom-left (手动计算垂直位置)
	fmt.Println("   Drawing 'Bottom Left' at manually calculated position")
	text = "Bottom Left"
	extents = ctx.TextExtents(text)
	ctx.MoveTo(10, 400-extents.Height-10)
	ctx.ShowText(text)

	// Text at bottom-right (手动计算位置)
	fmt.Println("   Drawing 'Bottom Right' at manually calculated position")
	text = "Bottom Right"
	extents = ctx.TextExtents(text)
	ctx.MoveTo(400-extents.XAdvance-10, 400-extents.Height-10)
	ctx.ShowText(text)

	// Text at center (增大字体以便更清楚显示)
	SetFont(ctx, "Go Regular", 24)
	fmt.Println("   Drawing 'Center' at manually calculated centered position")
	text = "Center"
	extents = ctx.TextExtents(text)
	x := (400 - extents.XAdvance) / 2
	y := (400-extents.Height)/2 + extents.Height
	ctx.MoveTo(x, y)
	ctx.ShowText(text)

	// Test 5: Bezier curves
	fmt.Println("➰ Drawing bezier curve...")
	ctx.SetSourceRGB(0, 1, 1) // Cyan
	ctx.SetLineWidth(3)
	ctx.MoveTo(100, 100)
	fmt.Println("   Drawing curve from (100,100) to (300,300) with control points")
	ctx.CurveTo(150, 50, 250, 350, 300, 300)
	ctx.Stroke()

	// Save to PNG with premultiplied alpha fix
	fmt.Println("💾 Saving image to PNG...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		// 应用反预乘 alpha 修复 PNG 透明度问题
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
