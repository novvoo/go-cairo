package cairo

import (
	"math"
)

// TestA1Fill - Ported from a1-fill.c
// Tests filling of an a1-surface and use as mask
func TestA1Fill() *TestCase {
	return NewTestCase(
		"a1_fill",
		"Test filling of an a1-surface and use as mask",
		"a1, alpha, fill, mask",
		100, 100,
		func(ctx Context, width, height int) TestStatus {
			// Create A1 surface  
			a1 := NewImageSurface(FormatA1, 100, 100)
			if a1.Status() != StatusSuccess {
				return TestError
			}
			defer a1.Destroy()
			
			ctx2 := NewContext(a1)
			if ctx2.Status() != StatusSuccess {
				return TestError
			}
			defer ctx2.Destroy()
			
			ctx2.SetOperator(OperatorSource)
			ctx2.Rectangle(10, 10, 80, 80)
			ctx2.SetSourceRGB(1, 1, 1)
			ctx2.Fill()
			ctx2.Rectangle(20, 20, 60, 60)
			ctx2.SetSourceRGB(0, 0, 0)
			ctx2.Fill()
			
			ctx.SetSourceRGB(1, 1, 1)
			ctx.Paint()
			
			ctx.SetSourceRGB(1, 0, 0)
			ctx.MaskSurface(a1, 0, 0)
			
			return TestSuccess
		},
	)
}

// TestArcDirection - Ported from arc-direction.c  
// Tests drawing positive/negative arcs
func TestArcDirectionAdvanced() *TestCase {
	const SIZE = 2 * 20
	const PAD = 2
	
	drawArcs := func(ctx Context) {
		start := math.Pi / 12
		stop := 2 * start
		
		ctx.MoveTo(SIZE/2, SIZE/2)
		ctx.Arc(SIZE/2, SIZE/2, SIZE/2, start, stop)
		ctx.Fill()
		
		ctx.Translate(SIZE+PAD, 0)
		ctx.MoveTo(SIZE/2, SIZE/2)
		ctx.Arc(SIZE/2, SIZE/2, SIZE/2, 2*math.Pi-stop, 2*math.Pi-start)
		ctx.Fill()
		
		ctx.Translate(0, SIZE+PAD)
		ctx.MoveTo(SIZE/2, SIZE/2)
		ctx.ArcNegative(SIZE/2, SIZE/2, SIZE/2, 2*math.Pi-stop, 2*math.Pi-start)
		ctx.Fill()
		
		ctx.Translate(-SIZE-PAD, 0)
		ctx.MoveTo(SIZE/2, SIZE/2)
		ctx.ArcNegative(SIZE/2, SIZE/2, SIZE/2, start, stop)
		ctx.Fill()
	}
	
	return NewTestCase(
		"arc_direction_advanced",
		"Test drawing positive/negative arcs with transformations",
		"arc, fill, transform",
		2*(3*PAD+2*SIZE), 2*(3*PAD+2*SIZE),
		func(ctx Context, width, height int) TestStatus {
			ctx.Save()
			ctx.SetSourceRGB(1.0, 1.0, 1.0) // white background
			ctx.Paint()
			ctx.Restore()
			
			ctx.Save()
			ctx.Translate(PAD, PAD)
			drawArcs(ctx)
			ctx.Restore()
			
			ctx.SetSourceRGB(1, 0, 0) // red
			ctx.Translate(2*SIZE+3*PAD, 0)
			ctx.Save()
			ctx.Translate(2*SIZE+2*PAD, PAD)
			ctx.Scale(-1, 1)
			drawArcs(ctx)
			ctx.Restore()
			
			ctx.SetSourceRGB(1, 0, 1) // magenta
			ctx.Translate(0, 2*SIZE+3*PAD)
			ctx.Save()
			ctx.Translate(2*SIZE+2*PAD, 2*SIZE+2*PAD)
			ctx.Scale(-1, -1)
			drawArcs(ctx)
			ctx.Restore()
			
			ctx.SetSourceRGB(0, 0, 1) // blue
			ctx.Translate(-(2*SIZE+3*PAD), 0)
			ctx.Save()
			ctx.Translate(PAD, 2*SIZE+2*PAD)
			ctx.Scale(1, -1)
			drawArcs(ctx)
			ctx.Restore()
			
			return TestSuccess
		},
	)
}

// TestClipAll - Ported from clip-all.c
// Test clipping with everything clipped out
func TestClipAll() *TestCase {
	const SIZE = 10
	const CLIP_SIZE = 2
	
	return NewTestCase(
		"clip_all",
		"Test clipping with everything clipped out",
		"clip",
		SIZE, SIZE,
		func(ctx Context, width, height int) TestStatus {
			ctx.Rectangle(0, 0, SIZE, SIZE)
			ctx.SetSourceRGB(0, 0, 1)
			ctx.Fill()
			
			ctx.ResetClip()
			ctx.Rectangle(CLIP_SIZE, CLIP_SIZE, CLIP_SIZE, CLIP_SIZE)
			ctx.Clip()
			ctx.Rectangle(3*CLIP_SIZE, 3*CLIP_SIZE, CLIP_SIZE, CLIP_SIZE)
			ctx.Clip()
			
			ctx.Translate(0.5, 0.5)
			
			ctx.ResetClip()
			ctx.Rectangle(CLIP_SIZE, CLIP_SIZE, CLIP_SIZE, CLIP_SIZE)
			ctx.Clip()
			ctx.Rectangle(3*CLIP_SIZE, 3*CLIP_SIZE, CLIP_SIZE, CLIP_SIZE)
			ctx.Clip()
			
			ctx.Rectangle(0, 0, SIZE, SIZE)
			ctx.SetSourceRGB(1, 1, 0)
			ctx.Fill()
			
			return TestSuccess
		},
	)
}

// TestClipContexts - Ported from clip-contexts.c
// Test clipping with 2 separate contexts  
func TestClipContexts() *TestCase {
	const SIZE = 10
	const CLIP_SIZE = 2
	
	return NewTestCase(
		"clip_contexts",
		"Test clipping with 2 separate contexts",
		"clip",
		SIZE, SIZE,
		func(ctx Context, width, height int) TestStatus {
			// Opaque background
			ctx.SetSourceRGB(0, 0, 0)
			ctx.Paint()
			
			// First create an empty, non-overlapping clip
			ctx2 := NewContext(ctx.GetTarget())
			if ctx2.Status() != StatusSuccess {
				return TestError
			}
			defer ctx2.Destroy()
			
			ctx2.Rectangle(0, 0, SIZE/2-2, SIZE/2-2)
			ctx2.Clip()
			
			ctx2.Rectangle(SIZE/2+2, SIZE/2+2, SIZE/2-2, SIZE/2-2)
			ctx2.Clip()
			
			// Apply the clip onto the surface, empty nothing should be painted
			ctx2.SetSourceRGBA(1, 0, 0, 0.5)
			ctx2.Paint()
			
			// Switch back to the original, and set only the last clip
			ctx.Rectangle(SIZE/2+2, SIZE/2+2, SIZE/2-2, SIZE/2-2)
			ctx.Clip()
			
			ctx.SetSourceRGBA(0, 0, 1, 0.5)
			ctx.Paint()
			
			return TestSuccess
		},
	)
}

// TestLinearGradientOneStop - Tests linear gradient with single stop
func TestLinearGradientOneStop() *TestCase {
	return NewTestCase(
		"linear_gradient_one_stop",
		"Test linear gradient with single color stop",
		"gradient, linear",
		120, 60,
		func(ctx Context, width, height int) TestStatus {
			// White background
			ctx.SetSourceRGB(1, 1, 1)
			ctx.Paint()
			
			// Create linear gradient with single stop
			gradient := NewPatternLinear(0, 0, 120, 0)
			defer gradient.Destroy()
			
			if linearGrad, ok := gradient.(LinearGradientPattern); ok {
				linearGrad.AddColorStopRGB(0.5, 1, 0, 0) // Single red stop at middle
				
				ctx.SetSource(gradient)
				ctx.Rectangle(10, 10, 100, 40)
				ctx.Fill()
			} else {
				return TestError
			}
			
			return TestSuccess
		},
	)
}

// TestRadialGradientOneStop - Tests radial gradient with single stop
func TestRadialGradientOneStop() *TestCase {
	return NewTestCase(
		"radial_gradient_one_stop", 
		"Test radial gradient with single color stop",
		"gradient, radial",
		120, 120,
		func(ctx Context, width, height int) TestStatus {
			// White background
			ctx.SetSourceRGB(1, 1, 1)
			ctx.Paint()
			
			// Create radial gradient with single stop
			gradient := NewPatternRadial(60, 60, 0, 60, 60, 40)
			defer gradient.Destroy()
			
			if radialGrad, ok := gradient.(RadialGradientPattern); ok {
				radialGrad.AddColorStopRGBA(0.0, 0, 1, 0, 0.8) // Single green stop
				
				ctx.SetSource(gradient)
				ctx.Arc(60, 60, 50, 0, 2*math.Pi)
				ctx.Fill()
			} else {
				return TestError
			}
			
			return TestSuccess
		},
	)
}

// TestOperatorClear - Tests CLEAR operator
func TestOperatorClear() *TestCase {
	return NewTestCase(
		"operator_clear",
		"Test CLEAR compositing operator", 
		"operator, compositing",
		100, 100,
		func(ctx Context, width, height int) TestStatus {
			// Red background
			ctx.SetSourceRGB(1, 0, 0)
			ctx.Paint()
			
			// Clear a rectangle in the middle
			ctx.SetOperator(OperatorClear)
			ctx.Rectangle(25, 25, 50, 50)
			ctx.Fill()
			
			return TestSuccess
		},
	)
}

// TestOperatorSource - Tests SOURCE operator
func TestOperatorSource() *TestCase {
	return NewTestCase(
		"operator_source",
		"Test SOURCE compositing operator",
		"operator, compositing", 
		100, 100,
		func(ctx Context, width, height int) TestStatus {
			// Semi-transparent red background
			ctx.SetSourceRGBA(1, 0, 0, 0.5)
			ctx.Paint()
			
			// Blue rectangle with SOURCE operator (should completely replace)
			ctx.SetOperator(OperatorSource)
			ctx.SetSourceRGBA(0, 0, 1, 0.7)
			ctx.Rectangle(25, 25, 50, 50)
			ctx.Fill()
			
			return TestSuccess
		},
	)
}

// TestDashOffset - Tests line dashing with offset
func TestDashOffset() *TestCase {
	return NewTestCase(
		"dash_offset",
		"Test line dashing with various offsets",
		"line, dash",
		150, 100,
		func(ctx Context, width, height int) TestStatus {
			// White background
			ctx.SetSourceRGB(1, 1, 1)
			ctx.Paint()
			
			ctx.SetSourceRGB(0, 0, 0)
			ctx.SetLineWidth(4)
			
			dashes := []float64{10, 5}
			
			// Line with offset 0
			ctx.SetDash(dashes, 0)
			ctx.MoveTo(10, 20)
			ctx.LineTo(140, 20)
			ctx.Stroke()
			
			// Line with offset 5  
			ctx.SetDash(dashes, 5)
			ctx.MoveTo(10, 40)
			ctx.LineTo(140, 40)
			ctx.Stroke()
			
			// Line with offset 10
			ctx.SetDash(dashes, 10)
			ctx.MoveTo(10, 60)
			ctx.LineTo(140, 60)
			ctx.Stroke()
			
			// Line with offset 15
			ctx.SetDash(dashes, 15)
			ctx.MoveTo(10, 80)
			ctx.LineTo(140, 80)
			ctx.Stroke()
			
			return TestSuccess
		},
	)
}

// TestLineCapStyles - Tests different line cap styles
func TestLineCapStyles() *TestCase {
	return NewTestCase(
		"line_cap_styles",
		"Test different line cap styles",
		"line, cap",
		150, 120,
		func(ctx Context, width, height int) TestStatus {
			// White background
			ctx.SetSourceRGB(1, 1, 1)
			ctx.Paint()
			
			ctx.SetSourceRGB(0, 0, 0)
			ctx.SetLineWidth(10)
			
			// Butt caps
			ctx.SetLineCap(LineCapButt)
			ctx.MoveTo(20, 30)
			ctx.LineTo(130, 30)
			ctx.Stroke()
			
			// Round caps
			ctx.SetLineCap(LineCapRound)
			ctx.MoveTo(20, 60)
			ctx.LineTo(130, 60)
			ctx.Stroke()
			
			// Square caps  
			ctx.SetLineCap(LineCapSquare)
			ctx.MoveTo(20, 90)
			ctx.LineTo(130, 90)
			ctx.Stroke()
			
			return TestSuccess
		},
	)
}

// TestLineJoinStyles - Tests different line join styles
func TestLineJoinStyles() *TestCase {
	return NewTestCase(
		"line_join_styles", 
		"Test different line join styles",
		"line, join",
		180, 120,
		func(ctx Context, width, height int) TestStatus {
			// White background
			ctx.SetSourceRGB(1, 1, 1)
			ctx.Paint()
			
			ctx.SetSourceRGB(0, 0, 0)
			ctx.SetLineWidth(8)
			
			// Miter joins
			ctx.SetLineJoin(LineJoinMiter)
			ctx.MoveTo(20, 20)
			ctx.LineTo(60, 50)
			ctx.LineTo(20, 80)
			ctx.Stroke()
			
			// Round joins
			ctx.SetLineJoin(LineJoinRound)
			ctx.MoveTo(80, 20)
			ctx.LineTo(120, 50)
			ctx.LineTo(80, 80)
			ctx.Stroke()
			
			// Bevel joins
			ctx.SetLineJoin(LineJoinBevel)
			ctx.MoveTo(140, 20)
			ctx.LineTo(180, 50)
			ctx.LineTo(140, 80)
			ctx.Stroke()
			
			return TestSuccess
		},
	)
}