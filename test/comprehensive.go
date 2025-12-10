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
	fmt.Println("📝 Creating image surface (800x600 pixels)...")
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 800, 600)
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
	fmt.Println("   Drawing green rectangle at (720, 50)")
	ctx.Rectangle(720, 50, 30, 30) // Top-right quadrant
	ctx.Fill()

	ctx.SetSourceRGB(0, 0, 1) // Blue
	fmt.Println("   Drawing blue rectangle at (50, 520)")
	ctx.Rectangle(50, 520, 30, 30) // Bottom-left quadrant
	ctx.Fill()

	ctx.SetSourceRGB(1, 1, 0) // Yellow
	fmt.Println("   Drawing yellow rectangle at (720, 520)")
	ctx.Rectangle(720, 520, 30, 30) // Bottom-right quadrant
	ctx.Fill()

	// Test 2: Lines to show coordinate system orientation
	fmt.Println("📏 Drawing coordinate system diagonals...")
	ctx.SetSourceRGB(0.8, 0.8, 0.8) // Light gray
	ctx.SetLineWidth(1)

	// Diagonal from top-left to bottom-right
	fmt.Println("   Drawing diagonal from (0,0) to (800,600)")
	ctx.MoveTo(0, 0)
	ctx.LineTo(800, 600)
	ctx.Stroke()

	// Diagonal from bottom-left to top-right
	fmt.Println("   Drawing diagonal from (0,600) to (800,0)")
	ctx.MoveTo(0, 600)
	ctx.LineTo(800, 0)
	ctx.Stroke()

	// Test 3: Arcs and circles
	fmt.Println("⭕ Drawing circle at center...")
	ctx.SetSourceRGB(1, 0, 1)           // Magenta
	ctx.Arc(400, 300, 50, 0, 2*math.Pi) // Circle at center
	ctx.Stroke()
	fmt.Println("   Circle drawn at (400, 300) with radius 50")

	// Test 4: Text rendering using PangoCairo
	fmt.Println("🔤 Drawing text samples...")
	ctx.SetSourceRGB(0, 0, 0) // Black

	// Create PangoCairo layout
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)

	// Create font description with size 20
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("Go Regular")
	fontDesc.SetWeight(cairo.PangoWeightNormal)
	fontDesc.SetSize(20)
	layout.SetFontDescription(fontDesc)

	// Text at top-left
	fmt.Println("   Drawing 'Top Left' at (10, 30)")
	layout.SetText("Top Left")
	ctx.MoveTo(10, 30)
	ctx.PangoCairoShowText(layout)

	// Text at top-right
	fmt.Println("   Drawing 'Top Right' at right-aligned position")
	text := "Top Right"
	layout.SetText(text)
	extents := layout.GetPixelExtents()
	ctx.MoveTo(800-extents.Width-10, 30)
	ctx.PangoCairoShowText(layout)

	// Text at bottom-left
	fmt.Println("   Drawing 'Bottom Left' at bottom position")
	text = "Bottom Left"
	layout.SetText(text)
	extents = layout.GetPixelExtents()
	ctx.MoveTo(10, 600-10)
	ctx.PangoCairoShowText(layout)

	// Text at bottom-right
	fmt.Println("   Drawing 'Bottom Right' at bottom-right position")
	text = "Bottom Right"
	layout.SetText(text)
	extents = layout.GetPixelExtents()
	ctx.MoveTo(800-extents.Width-10, 600-10)
	ctx.PangoCairoShowText(layout)

	// Text at center with larger font
	fmt.Println("   Drawing 'Center' at centered position")
	fontDesc.SetSize(32)
	layout.SetFontDescription(fontDesc)
	text = "Center"
	layout.SetText(text)
	extents = layout.GetPixelExtents()
	fontExtents := layout.GetFontExtents()
	x := (800 - extents.Width) / 2
	y := (600-fontExtents.Height)/2 + fontExtents.Ascent
	ctx.MoveTo(x, y)
	ctx.PangoCairoShowText(layout)

	// Test 5: Bezier curves
	fmt.Println("➰ Drawing bezier curve...")
	ctx.SetSourceRGB(0, 1, 1) // Cyan
	ctx.SetLineWidth(3)
	ctx.MoveTo(100, 100)
	fmt.Println("   Drawing curve from (100,100) to (700,500) with control points")
	ctx.CurveTo(200, 50, 600, 550, 700, 500)
	ctx.Stroke()

	// Test 6: Multiple lines of text
	fmt.Println("📝 Drawing multiple lines of text...")
	ctx.SetSourceRGB(0, 0, 0) // Black
	fontDesc.SetSize(16)
	layout.SetFontDescription(fontDesc)

	lines := []string{
		"This is a comprehensive test",
		"of the Cairo graphics library",
		"with proper text rendering",
		"using PangoCairo integration",
	}

	startY := 150.0
	lineHeight := 25.0
	for i, line := range lines {
		y := startY + float64(i)*lineHeight
		layout.SetText(line)
		ctx.MoveTo(50, y)
		ctx.PangoCairoShowText(layout)
		fmt.Printf("   Drawing line %d: '%s' at y=%.1f\n", i+1, line, y)
	}

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
