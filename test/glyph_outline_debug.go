//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/opentype/api"
	"github.com/novvoo/go-cairo/pkg/cairo"
	"golang.org/x/image/math/fixed"
)

func main() {
	// Load a font
	pangoFont := cairo.NewPangoCairoFont("sans", cairo.FontSlantNormal, cairo.FontWeightNormal)
	defer pangoFont.Destroy()

	// Get the real face
	var realFace font.Face
	fontKey := "sans-normal-normal"
	face, _, err := cairo.LoadEmbeddedFont(fontKey)
	if err != nil {
		log.Fatal("Failed to load font:", err)
	}
	realFace = face

	// Get glyph ID for 'H'
	gid, ok := realFace.NominalGlyph('H')
	if !ok {
		log.Fatal("Failed to get glyph for 'H'")
	}

	fmt.Printf("字符 'H' 的字形 ID: %d\n\n", gid)

	// Get glyph data (outline)
	glyphData := realFace.GlyphData(gid)
	outline, ok := glyphData.(api.GlyphOutline)
	if !ok {
		log.Fatal("Glyph has no outline")
	}

	// Print first few outline points
	fmt.Println("轮廓点（原始格式）:")
	pointCount := 0
	for _, seg := range outline.Segments {
		for _, arg := range seg.Args {
			if pointCount < 5 {
				fmt.Printf("  点 %d: X=%d, Y=%d (类型: %T)\n", pointCount, arg.X, arg.Y, arg.X)
				fmt.Printf("         X/64=%.2f, Y/64=%.2f\n", float64(arg.X)/64.0, float64(arg.Y)/64.0)
			}
			pointCount++
		}
	}

	// Get font metrics
	unitsPerEm := realFace.Upem()
	fmt.Printf("\nUnits Per Em: %d\n", unitsPerEm)

	// Get horizontal advance
	advance := realFace.HorizontalAdvance(gid)
	fmt.Printf("Horizontal Advance (字体单位): %d\n", advance)
	fmt.Printf("Horizontal Advance (归一化): %.4f\n", float64(advance)/float64(unitsPerEm))

	// Calculate what the advance should be at 24pt
	advanceAt24pt := (float64(advance) / float64(unitsPerEm)) * 24.0
	fmt.Printf("Horizontal Advance (24pt): %.2f\n", advanceAt24pt)

	// Now let's check what fixed.I(24) means
	size := fixed.I(24)
	fmt.Printf("\nfixed.I(24) = %d (26.6 格式)\n", size)
	fmt.Printf("fixed.I(24) / 64 = %.2f\n", float64(size)/64.0)
}
