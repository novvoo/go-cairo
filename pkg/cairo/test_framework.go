package cairo

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// TestStatus represents the result of a cairo test
type TestStatus int

const (
	TestSuccess TestStatus = iota
	TestNoMemory
	TestFailure
	TestNew
	TestXFailure
	TestError
	TestCrashed
	TestUntested = 77 // match automake's skipped exit status
)

// TestContext holds the test environment and state
type TestContext struct {
	Name        string
	Description string
	Keywords    string
	Width       int
	Height      int
	OutputDir   string
	RefDir      string
}

// DrawFunction is the signature for test drawing functions
type DrawFunction func(ctx Context, width, height int) TestStatus

// TestCase represents a complete cairo test case
type TestCase struct {
	Name        string
	Description string
	Keywords    string
	Width       int
	Height      int
	Draw        DrawFunction
}

// NewTestCase creates a new test case
func NewTestCase(name, description, keywords string, width, height int, draw DrawFunction) *TestCase {
	return &TestCase{
		Name:        name,
		Description: description,
		Keywords:    keywords,
		Width:       width,
		Height:      height,
		Draw:        draw,
	}
}

// RunTest executes a cairo test case and saves the output
func (tc *TestCase) RunTest(t *testing.T) {
	// Create output directory
	outputDir := "test_output"
	os.MkdirAll(outputDir, 0755)

	// Create surface and context
	surface := NewImageSurface(FormatARGB32, tc.Width, tc.Height)
	if surface.Status() != StatusSuccess {
		t.Fatalf("Failed to create surface: %v", surface.Status())
	}
	defer surface.Destroy()

	ctx := NewContext(surface)
	if ctx.Status() != StatusSuccess {
		t.Fatalf("Failed to create context: %v", ctx.Status())
	}
	defer ctx.Destroy()

	// Run the drawing function
	status := tc.Draw(ctx, tc.Width, tc.Height)

	// Check cairo status
	if ctx.Status() != StatusSuccess {
		t.Errorf("Cairo error during draw: %v", ctx.Status())
	}

	// Save output image
	if imageSurface, ok := surface.(ImageSurface); ok {
		filename := filepath.Join(outputDir, tc.Name+".png")
		writeStatus := imageSurface.WriteToPNG(filename)
		if writeStatus != StatusSuccess {
			t.Errorf("Failed to write PNG %s: %v", filename, writeStatus)
		} else {
			t.Logf("Saved test output: %s", filename)
		}
	}

	// Check test result
	switch status {
	case TestSuccess:
		// Test passed
	case TestUntested:
		t.Skipf("Test %s was skipped", tc.Name)
	case TestFailure, TestError:
		t.Errorf("Test %s failed with status: %v", tc.Name, status)
	default:
		t.Errorf("Test %s returned unexpected status: %v", tc.Name, status)
	}
}

// Helper functions for common test operations

// TestFillRectangle draws a filled rectangle for testing
func TestFillRectangle(ctx Context, x, y, width, height float64, r, g, b, a float64) {
	ctx.Rectangle(x, y, width, height)
	ctx.SetSourceRGBA(r, g, b, a)
	ctx.Fill()
}

// TestStrokeRectangle draws a stroked rectangle for testing
func TestStrokeRectangle(ctx Context, x, y, width, height float64, lineWidth float64, r, g, b, a float64) {
	ctx.Rectangle(x, y, width, height)
	ctx.SetLineWidth(lineWidth)
	ctx.SetSourceRGBA(r, g, b, a)
	ctx.Stroke()
}

// TestFillCircle draws a filled circle for testing
func TestFillCircle(ctx Context, x, y, radius float64, r, g, b, a float64) {
	ctx.Arc(x, y, radius, 0, 2*math.Pi)
	ctx.SetSourceRGBA(r, g, b, a)
	ctx.Fill()
}

// TestSetBackground sets a solid background color
func TestSetBackground(ctx Context, r, g, b, a float64) {
	ctx.SetSourceRGBA(r, g, b, a)
	ctx.Paint()
}

// Common test patterns

// TestPatternSolidFill tests solid color fills
func TestPatternSolidFill() *TestCase {
	return NewTestCase(
		"solid_fill",
		"Test solid color fills",
		"fill, solid",
		100, 100,
		func(ctx Context, width, height int) TestStatus {
			// White background
			TestSetBackground(ctx, 1, 1, 1, 1)

			// Red rectangle
			TestFillRectangle(ctx, 10, 10, 30, 30, 1, 0, 0, 1)

			// Green circle
			TestFillCircle(ctx, 70, 30, 15, 0, 1, 0, 1)

			// Blue semi-transparent rectangle
			TestFillRectangle(ctx, 30, 50, 40, 30, 0, 0, 1, 0.5)

			return TestSuccess
		},
	)
}

// TestPatternLinearGradient tests linear gradients
func TestPatternLinearGradient() *TestCase {
	return NewTestCase(
		"linear_gradient",
		"Test linear gradient patterns",
		"pattern, gradient, linear",
		150, 100,
		func(ctx Context, width, height int) TestStatus {
			// White background
			TestSetBackground(ctx, 1, 1, 1, 1)

			// Create linear gradient
			gradient := NewPatternLinear(0, 0, 150, 0)
			defer gradient.Destroy()

			if linearGrad, ok := gradient.(LinearGradientPattern); ok {
				linearGrad.AddColorStopRGB(0.0, 1, 0, 0) // Red
				linearGrad.AddColorStopRGB(0.5, 0, 1, 0) // Green
				linearGrad.AddColorStopRGB(1.0, 0, 0, 1) // Blue

				ctx.SetSource(gradient)
				ctx.Rectangle(10, 10, 130, 30)
				ctx.Fill()
			} else {
				return TestError
			}

			// Create radial gradient
			radial := NewPatternRadial(75, 70, 0, 75, 70, 30)
			defer radial.Destroy()

			if radialGrad, ok := radial.(RadialGradientPattern); ok {
				radialGrad.AddColorStopRGBA(0.0, 1, 1, 1, 1)   // White center
				radialGrad.AddColorStopRGBA(1.0, 0, 0, 0, 0.8) // Dark edge

				ctx.SetSource(radial)
				ctx.Arc(75, 70, 25, 0, 2*math.Pi)
				ctx.Fill()
			} else {
				return TestError
			}

			return TestSuccess
		},
	)
}

// TestPatternTransforms tests pattern transformations
func TestPatternTransforms() *TestCase {
	return NewTestCase(
		"pattern_transforms",
		"Test pattern matrix transformations",
		"pattern, transform, matrix",
		120, 120,
		func(ctx Context, width, height int) TestStatus {
			// White background
			TestSetBackground(ctx, 1, 1, 1, 1)

			// Create gradient pattern
			gradient := NewPatternLinear(0, 0, 50, 50)
			defer gradient.Destroy()

			if linearGrad, ok := gradient.(LinearGradientPattern); ok {
				linearGrad.AddColorStopRGB(0.0, 1, 0, 0)
				linearGrad.AddColorStopRGB(1.0, 0, 0, 1)

				// First rectangle - no transformation
				ctx.SetSource(gradient)
				ctx.Rectangle(10, 10, 40, 40)
				ctx.Fill()

				// Second rectangle - scaled pattern
				matrix := NewMatrix()
				matrix.InitScale(2.0, 2.0)
				gradient.SetMatrix(matrix)

				ctx.SetSource(gradient)
				ctx.Rectangle(60, 10, 40, 40)
				ctx.Fill()

				// Third rectangle - rotated pattern
				matrix.InitRotate(math.Pi / 4)
				gradient.SetMatrix(matrix)

				ctx.SetSource(gradient)
				ctx.Rectangle(10, 60, 40, 40)
				ctx.Fill()

				// Fourth rectangle - translated pattern
				matrix.InitTranslate(10, 10)
				gradient.SetMatrix(matrix)

				ctx.SetSource(gradient)
				ctx.Rectangle(60, 60, 40, 40)
				ctx.Fill()
			} else {
				return TestError
			}

			return TestSuccess
		},
	)
}

// TestClipBasic tests basic clipping functionality
func TestClipBasic() *TestCase {
	return NewTestCase(
		"clip_basic",
		"Test basic clipping operations",
		"clip",
		100, 100,
		func(ctx Context, width, height int) TestStatus {
			// White background
			TestSetBackground(ctx, 1, 1, 1, 1)

			// Set clipping region
			ctx.Rectangle(20, 20, 60, 60)
			ctx.Clip()

			// Draw red rectangle that extends beyond clip
			TestFillRectangle(ctx, 10, 10, 80, 80, 1, 0, 0, 1)

			// Reset clip and draw blue border to show clip boundary
			ctx.ResetClip()
			TestStrokeRectangle(ctx, 20, 20, 60, 60, 2, 0, 0, 1, 1)

			return TestSuccess
		},
	)
}

// TestStateManagement tests save/restore functionality
func TestStateManagement() *TestCase {
	return NewTestCase(
		"state_management",
		"Test graphics state save/restore",
		"state, save, restore",
		150, 100,
		func(ctx Context, width, height int) TestStatus {
			// White background
			TestSetBackground(ctx, 1, 1, 1, 1)

			// Set initial state
			ctx.SetSourceRGB(1, 0, 0) // Red
			ctx.SetLineWidth(3)

			// Draw with initial state
			ctx.Rectangle(10, 10, 30, 30)
			ctx.Stroke()

			// Save state and modify
			ctx.Save()
			ctx.SetSourceRGB(0, 1, 0) // Green
			ctx.SetLineWidth(1)
			ctx.Translate(50, 0)

			// Draw with modified state
			ctx.Rectangle(10, 10, 30, 30)
			ctx.Stroke()

			// Restore state
			ctx.Restore()

			// Draw with restored state (should be red, width 3, no translation)
			ctx.Rectangle(10, 50, 30, 30)
			ctx.Stroke()

			return TestSuccess
		},
	)
}

// TestArcDirection tests arc drawing in different directions
func TestArcDirection() *TestCase {
	return NewTestCase(
		"arc_direction",
		"Test drawing positive/negative arcs",
		"arc, fill",
		100, 100,
		func(ctx Context, width, height int) TestStatus {
			// White background
			TestSetBackground(ctx, 1, 1, 1, 1)

			start := math.Pi / 12
			stop := 2 * start

			ctx.SetSourceRGB(1, 0, 0) // Red

			// Positive arc top-left
			ctx.MoveTo(25, 25)
			ctx.Arc(25, 25, 20, start, stop)
			ctx.Fill()

			// Negative arc top-right
			ctx.MoveTo(75, 25)
			ctx.ArcNegative(75, 25, 20, start, stop)
			ctx.Fill()

			// Positive arc bottom-left
			ctx.MoveTo(25, 75)
			ctx.Arc(25, 75, 20, 2*math.Pi-stop, 2*math.Pi-start)
			ctx.Fill()

			// Negative arc bottom-right
			ctx.MoveTo(75, 75)
			ctx.ArcNegative(75, 75, 20, 2*math.Pi-stop, 2*math.Pi-start)
			ctx.Fill()

			return TestSuccess
		},
	)
}

// TestPatternSolidFillGo is a standard Go test function
func TestPatternSolidFillGo(t *testing.T) {
	tc := TestPatternSolidFill()
	tc.RunTest(t)
}
