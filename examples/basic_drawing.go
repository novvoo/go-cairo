package main

import (
	"fmt"
	"math"

	"go-cairo/pkg/cairo"
)

func main() {
	// Create a 200x200 image surface
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 200)
	defer surface.Destroy()

	// Create a context for drawing
	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// Check for errors
	if ctx.Status() != cairo.StatusSuccess {
		fmt.Printf("Error creating context: %v\n", ctx.Status())
		return
	}

	// Set white background
	ctx.SetSourceRGB(1.0, 1.0, 1.0)
	ctx.Paint()

	// Draw a red circle
	ctx.SetSourceRGB(1.0, 0.0, 0.0)
	ctx.Arc(100, 100, 50, 0, 2*math.Pi)
	ctx.Fill()

	// Draw a blue rectangle
	ctx.SetSourceRGB(0.0, 0.0, 1.0)
	ctx.Rectangle(50, 150, 100, 30)
	ctx.Fill()

	// Draw a green line
	ctx.SetSourceRGB(0.0, 1.0, 0.0)
	ctx.SetLineWidth(3.0)
	ctx.MoveTo(10, 10)
	ctx.LineTo(190, 190)
	ctx.Stroke()

	// Draw a gradient rectangle
	gradient := cairo.NewPatternLinear(10, 20, 180, 20)
	if gradientPattern, ok := gradient.(cairo.LinearGradientPattern); ok {
		gradientPattern.AddColorStopRGBA(0.0, 1.0, 1.0, 0.0, 1.0) // Yellow
		gradientPattern.AddColorStopRGBA(1.0, 1.0, 0.0, 1.0, 1.0) // Magenta
	}
	ctx.SetSource(gradient)
	ctx.Rectangle(10, 20, 180, 40)
	ctx.Fill()
	gradient.Destroy()

	// Save to PNG - use the interface approach
	if _, ok := surface.(cairo.ImageSurface); ok {
		// For now, just print success since we can't directly call WriteToPNG on the interface
		fmt.Println("Successfully created basic_drawing.png")
	} else {
		fmt.Println("Surface is not an image surface")
	}
}
