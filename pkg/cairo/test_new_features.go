package cairo

import (
	"fmt"
	"os"
	"testing"
	"math"
)

// TestNewFeatures demonstrates the usage of newly added features.
func TestNewFeatures(t *testing.T) {
	// 1. Test PostScript Surface (Phase 2)
	const psFilename = "test_output.ps"
	psSurface := NewPostScriptSurface(psFilename, 500, 500)
	if psSurface.Status() != StatusSuccess {
		t.Fatalf("NewPostScriptSurface failed: %v", psSurface.Status())
	}
	defer psSurface.Destroy()
	defer os.Remove(psFilename)

	ctx := NewContext(psSurface)
	if ctx.Status() != StatusSuccess {
		t.Fatalf("NewContext failed: %v", ctx.Status())
	}
	defer ctx.Destroy()

	// 2. Test Gradient with ExtendReflect (Phase 3)
	linearPattern := NewLinearGradient(0, 0, 500, 0)
	linearPattern.AddColorStopRGB(0.0, 1.0, 0.0, 0.0) // Red
	linearPattern.AddColorStopRGB(0.5, 0.0, 1.0, 0.0) // Green
	linearPattern.AddColorStopRGB(1.0, 0.0, 0.0, 1.0) // Blue
	linearPattern.SetExtend(ExtendReflect)
	
	ctx.SetSource(linearPattern)
	ctx.Rectangle(0, 0, 500, 500)
	ctx.Fill()

	// 3. Test ShowText (Phase 4 - HarfBuzz-like shaping)
	ctx.SetSourceRGB(0, 0, 0) // Black text
	ctx.SelectFontFace("sans", FontSlantNormal, FontWeightNormal)
	ctx.SetFontSize(40)
	ctx.MoveTo(50, 50)
	// This text would require shaping in a real-world scenario (e.g., Arabic, Devanagari)
	// For this test, we use a simple string to confirm the function call path.
	if err := ctx.ShowText("Hello Cairo 1.18.2 Features!"); err != nil {
		t.Errorf("ShowText failed: %v", err)
	}

	// 4. Test MatrixDecompose (Phase 5)
	matrix := NewMatrix()
	matrix.InitRotate(math.Pi / 4) // 45 degrees
	matrix.Scale(2.0, 1.5)
	matrix.Translate(10, 20)

	tx, ty, rot, sx, sy, shear, status := MatrixDecompose(matrix)
	if status != StatusSuccess {
		t.Errorf("MatrixDecompose failed: %v", status)
	}

	// Simple check to ensure values are non-zero and plausible
	if tx != 10 || ty != 20 {
		t.Errorf("MatrixDecompose translation incorrect: got (%f, %f), want (10, 20)", tx, ty)
	}
	if sx <= 1.0 || sy <= 1.0 {
		t.Errorf("MatrixDecompose scale incorrect: got (%f, %f)", sx, sy)
	}
	
	fmt.Printf("Matrix Decomposed: T=(%f, %f), R=%f, S=(%f, %f), Shear=%f\n", tx, ty, rot, sx, sy, shear)

	// Finish the surface
	psSurface.Finish()
}
