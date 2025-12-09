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
	ctx.SelectFontFace("sans-serif", cairo.FontSlantNormal, cairo.FontWeightBold)
	ctx.SetFontSize(18)
	ctx.SetSourceRGB(0, 0, 0) // Black

	// Text at top-left
	fmt.Println("   Drawing 'Top Left' at (10, 30)")
	ctx.MoveTo(10, 30)
	ctx.ShowText("Top Left")

	// Text at top-right
	fmt.Println("   Drawing 'Top Right' at (280, 30)")
	ctx.MoveTo(280, 30)
	ctx.ShowText("Top Right")

	// Text at bottom-left
	fmt.Println("   Drawing 'Bottom Left' at (10, 390)")
	ctx.MoveTo(10, 390)
	ctx.ShowText("Bottom Left")

	// Text at bottom-right
	fmt.Println("   Drawing 'Bottom Right' at (250, 390)")
	ctx.MoveTo(250, 390)
	ctx.ShowText("Bottom Right")

	// Text at center (增大字体以便更清楚显示)
	ctx.SetFontSize(24)
	fmt.Println("   Drawing 'Center' at (170, 200)")
	ctx.MoveTo(170, 200)
	ctx.ShowText("Center")

	// Test 5: Bezier curves
	fmt.Println("➰ Drawing bezier curve...")
	ctx.SetSourceRGB(0, 1, 1) // Cyan
	ctx.SetLineWidth(3)
	ctx.MoveTo(100, 100)
	fmt.Println("   Drawing curve from (100,100) to (300,300) with control points")
	ctx.CurveTo(150, 50, 250, 350, 300, 300)
	ctx.Stroke()

	// Save to PNG
	fmt.Println("💾 Saving image to PNG...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("images/comprehensive_test.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("WriteToPNG failed: %v", status))
		}
		fmt.Println("✅ Comprehensive test image saved to images/comprehensive_test.png")
	} else {
		panic("Surface is not an ImageSurface")
	}

	fmt.Println("🎉 Comprehensive Cairo demo completed!")
}
