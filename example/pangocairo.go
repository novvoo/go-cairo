package main

import (
	"fmt"
	"log"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	// Create a new image surface
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 400, 200)
	defer surface.Destroy()

	// Create a context
	context := cairo.NewContext(surface)
	defer context.Destroy()

	// Set background color
	context.SetSourceRGB(1, 1, 1) // White background
	context.Paint()

	// Set text color
	context.SetSourceRGB(0, 0, 0) // Black text

	// Simple text rendering
	context.SelectFontFace("sans", cairo.FontSlantNormal, cairo.FontWeightNormal)
	context.SetFontSize(24)

	// Move to position and show text
	context.MoveTo(50, 100)
	context.ShowText("Hello, Cairo!")

	// Save to PNG
	if imageSurface, ok := surface.(cairo.ImageSurface); ok {
		status := imageSurface.WriteToPNG("example/images/pangocairo.png")
		if status != cairo.StatusSuccess {
			log.Fatal("Failed to save PNG:", status)
		}
	} else {
		log.Fatal("Surface is not an ImageSurface")
	}

	fmt.Println("Simple test saved to example/images/pangocairo.png")
}
