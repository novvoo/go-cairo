//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"math"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("🔵 Circle Comparison Test: Arc vs DrawCircle")

	// Create a new image surface
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 600, 300)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// White background
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// Left side: Using Arc method
	fmt.Println("📍 Drawing circle using Arc method (left)")
	ctx.SetSourceRGB(1, 0, 1) // Magenta
	ctx.SetLineWidth(3)
	ctx.Arc(150, 150, 80, 0, 2*math.Pi)
	ctx.Stroke()

	// Add label
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("Go Regular")
	fontDesc.SetSize(16)
	layout.SetFontDescription(fontDesc)
	layout.SetText("Arc Method")
	ctx.SetSourceRGB(0, 0, 0)
	ctx.MoveTo(110, 250)
	ctx.PangoCairoShowText(layout)

	// Right side: Using DrawCircle method
	fmt.Println("📍 Drawing circle using DrawCircle method (right)")
	ctx.SetSourceRGB(0, 0.5, 1) // Blue
	ctx.SetLineWidth(3)
	ctx.DrawCircle(450, 150, 80)
	ctx.Stroke()

	// Add label
	layout.SetText("DrawCircle Method")
	ctx.SetSourceRGB(0, 0, 0)
	ctx.MoveTo(380, 250)
	ctx.PangoCairoShowText(layout)

	// Draw crosshairs at centers for reference
	ctx.SetSourceRGB(0.7, 0.7, 0.7)
	ctx.SetLineWidth(1)
	
	// Left crosshair
	ctx.MoveTo(130, 150)
	ctx.LineTo(170, 150)
	ctx.Stroke()
	ctx.MoveTo(150, 130)
	ctx.LineTo(150, 170)
	ctx.Stroke()

	// Right crosshair
	ctx.MoveTo(430, 150)
	ctx.LineTo(470, 150)
	ctx.Stroke()
	ctx.MoveTo(450, 130)
	ctx.LineTo(450, 170)
	ctx.Stroke()

	// Save to PNG
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("circle_comparison.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("WriteToPNG failed: %v", status))
		}
		fmt.Println("✅ Circle comparison saved to circle_comparison.png")
	}

	fmt.Println("🎉 Test completed!")
}
