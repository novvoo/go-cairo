package cairo

import (
	"math"
	"sync/atomic"
	"unsafe"
	"runtime"
	"sync"
	
	"image/draw"
	"image/color"
		"github.com/llgcode/draw2d"
		"github.com/llgcode/draw2d/draw2dimg"
		"github.com/llgcode/draw2d/draw2dpdf"
		"github.com/llgcode/draw2d/draw2dsvg"
	)

// GroupSurface is a temporary surface used for group operations.
type GroupSurface struct {
	Surface
	originalTarget Surface
	originalGC draw2d.GraphicContext
}

// context implements the Context interface
	type context struct {
		// Mutex for concurrency safety
		mu sync.Mutex

		// Reference counting
		refCount int32

	// Status
	status Status

	// Target surface
	target Surface

	// User data
	userData map[*UserDataKey]interface{}

	// Graphics state stack
	gstate *graphicsState

	// Path
	path *path

		// Current point
		currentPoint struct {
			x, y     float64
			hasPoint bool
		}
		
		// Drawing context for backend
		gc draw2d.GraphicContext
	}

// graphicsState represents the graphics state that can be saved/restored
type graphicsState struct {
	// Rendering properties
	source    Pattern
	operator  Operator
	tolerance float64
	antialias Antialias
	fillRule  FillRule

	// Line properties
	lineWidth  float64
	lineCap    LineCap
	lineJoin   LineJoin
	miterLimit float64
	dash       []float64
	dashOffset float64

	// Transformation matrix
	matrix Matrix

	// Font properties
	fontFace    FontFace
	fontMatrix  Matrix
	fontOptions *FontOptions
	scaledFont  ScaledFont

	// Clip region
	clip *clipRegion

	// Previous state in stack
	next *graphicsState
	
	// Group surface reference for PopGroup
	groupSurface *GroupSurface
}

// clipRegion represents clipping information
type clipRegion struct {
	// Clipping path
	path      *path
	fillRule  FillRule
	tolerance float64
	antialias Antialias

	// Previous clip in stack
	prev *clipRegion
}

// path represents the current path
type path struct {
	// Path data
	data []pathOp

	// Current subpath starting point
	subpathStartX, subpathStartY float64
}

// pathOp represents a path operation
type pathOp struct {
	op     PathDataType
	points []point
}

type point struct {
	x, y float64
}

// NewContext creates a new drawing context for the given surface
	func NewContext(target Surface) Context {
		if target == nil {
			return newContextInError(StatusNullPointer)
		}

			ctx := &context{
				refCount: 1,
				target:   target.Reference(),
				userData: make(map[*UserDataKey]interface{}),
				gstate:   newGraphicsState(),
				path:     &path{data: make([]pathOp, 0)},
			}
			
			runtime.SetFinalizer(ctx, (*context).destroyConcrete)

				switch s := target.(type) {
				case ImageSurface:
					if img := s.GetGoImage(); img != nil {
						// draw2dimg.NewGraphicContext expects *image.RGBA or *image.NRGBA
						// We assume ImageSurface.GetGoImage() returns *image.NRGBA for now
						if nrgba, ok := img.(*image.NRGBA); ok {
							ctx.gc = draw2dimg.NewGraphicContext(nrgba)
						}
					}
				case *pdfSurface:
					// Create a draw2d PDF context
					pdfCtx := draw2dpdf.NewPdf(s.width, s.height)
					ctx.gc = pdfCtx
					s.gc = pdfCtx // Store a reference in the surface for Finish()
				case *svgSurface:
					// Create a draw2d SVG context
					svgCtx := draw2dsvg.NewSvg()
					svgCtx.SetDPI(72) // Default cairo DPI
					svgCtx.SetCanvasSize(s.width, s.height)
					ctx.gc = svgCtx
					s.gc = svgCtx // Store a reference in the surface for Finish()
				}

		// Initialize default state
		ctx.gstate.source = NewPatternRGB(0, 0, 0) // Black
		ctx.gstate.operator = OperatorOver
		ctx.gstate.tolerance = 0.1
		ctx.gstate.antialias = AntialiasDefault
		ctx.gstate.fillRule = FillRuleWinding
		ctx.gstate.lineWidth = 2.0
		ctx.gstate.lineCap = LineCapButt
		ctx.gstate.lineJoin = LineJoinMiter
		ctx.gstate.miterLimit = 10.0
		ctx.gstate.matrix.InitIdentity()

		return ctx
	}

func newContextInError(status Status) Context {
	ctx := &context{
		refCount: 1,
		status:   status,
		userData: make(map[*UserDataKey]interface{}),
	}
	return ctx
}

func newGraphicsState() *graphicsState {
	return &graphicsState{
		fontOptions: &FontOptions{},
		fontMatrix:  Matrix{XX: 1, YY: 1}, // Identity matrix
	}
}

// Reference management
	func (c *context) Reference() Context {
		atomic.AddInt32(&c.refCount, 1)
		return c
	}
	
	func (c *context) Destroy() {
		if atomic.AddInt32(&c.refCount, -1) == 0 {
			c.destroyConcrete()
		}
	}
	
	func (c *context) destroyConcrete() {
		if c.target != nil {
			c.target.Destroy()
		}
	
		// Clean up graphics state stack
		for c.gstate != nil {
			if c.gstate.source != nil {
				c.gstate.source.Destroy()
			}
			if c.gstate.fontFace != nil {
				c.gstate.fontFace.Destroy()
			}
			if c.gstate.scaledFont != nil {
				c.gstate.scaledFont.Destroy()
			}
			c.gstate = c.gstate.next
		}
	}

func (c *context) GetReferenceCount() int {
	return int(atomic.LoadInt32(&c.refCount))
}

// Status
func (c *context) Status() Status {
	return c.status
}

// Target surface
func (c *context) GetTarget() Surface {
	return c.target
}

func (c *context) GetGroupTarget() Surface {
	// TODO: Implement group target tracking
	return c.target
}

// User data
func (c *context) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	if c.status != StatusSuccess {
		return c.status
	}

	c.userData[key] = userData
	// TODO: Store destroy function and call it when appropriate
	return StatusSuccess
}

func (c *context) GetUserData(key *UserDataKey) unsafe.Pointer {
	if data, exists := c.userData[key]; exists {
		return data.(unsafe.Pointer)
	}
	return nil
}

// State management
func (c *context) Save() {
	if c.status != StatusSuccess {
		return
	}

// Create a copy of current state
			newState := &graphicsState{
				source:      c.gstate.source.Reference(),
				operator:    c.gstate.operator,
				tolerance:   c.gstate.tolerance,
				antialias:   c.gstate.antialias,
				fillRule:    c.gstate.fillRule,
				lineWidth:   c.gstate.lineWidth,
				lineCap:     c.gstate.lineCap,
				lineJoin:    c.gstate.lineJoin,
				miterLimit:  c.gstate.miterLimit,
				matrix:      c.gstate.matrix,
				fontMatrix:  c.gstate.fontMatrix,
				fontOptions: c.gstate.fontOptions, // TODO: Copy font options
				clip:        c.gstate.clip, // Clip is part of the graphics state
				next:        c.gstate,
				groupSurface: c.gstate.groupSurface, // Copy group surface reference
			}

	// Copy dash array
	if len(c.gstate.dash) > 0 {
		newState.dash = make([]float64, len(c.gstate.dash))
		copy(newState.dash, c.gstate.dash)
	}
	newState.dashOffset = c.gstate.dashOffset

	// Reference font objects
	if c.gstate.fontFace != nil {
		newState.fontFace = c.gstate.fontFace.Reference()
	}
	if c.gstate.scaledFont != nil {
		newState.scaledFont = c.gstate.scaledFont.Reference()
	}

	c.gstate = newState
}

func (c *context) Restore() {
	if c.status != StatusSuccess {
		return
	}

	if c.gstate.next == nil {
		c.status = StatusInvalidRestore
		return
	}

	// Release current state resources
	if c.gstate.source != nil {
		c.gstate.source.Destroy()
	}
	if c.gstate.fontFace != nil {
		c.gstate.fontFace.Destroy()
	}
	if c.gstate.scaledFont != nil {
		c.gstate.scaledFont.Destroy()
	}

				// Restore previous state
				oldState := c.gstate
				c.gstate = oldState.next
				oldState.next = nil
		
				// If the old state was a group, restore the target and gc
				if oldState.groupSurface != nil {
					c.target = oldState.groupSurface.originalTarget
					c.gc = oldState.groupSurface.originalGC
					oldState.groupSurface.Surface.Destroy() // Destroy the temporary surface
				}
		
				// Re-apply clip path to draw2d context
				if c.gstate.clip != nil {
					// Re-apply the path to draw2d for clipping
					// This is a simplification; a proper implementation would need to store the draw2d path
					// or re-create it from the cairo path structure.
					// For now, we'll just reset the clip.
					c.gc.SetClipPath(nil)
				} else {
					c.gc.SetClipPath(nil)
				}
}

// Source pattern
func (c *context) SetSource(source Pattern) {
	if c.status != StatusSuccess {
		return
	}

	if c.gstate.source != nil {
		c.gstate.source.Destroy()
	}
	c.gstate.source = source.Reference()
}

func (c *context) SetSourceRGB(red, green, blue float64) {
	c.SetSourceRGBA(red, green, blue, 1.0)
}

func (c *context) SetSourceRGBA(red, green, blue, alpha float64) {
	pattern := NewPatternRGBA(red, green, blue, alpha)
	c.SetSource(pattern)
	pattern.Destroy()
}

func (c *context) SetSourceSurface(surface Surface, x, y float64) {
	pattern := NewPatternForSurface(surface)
	matrix := NewMatrix()
	matrix.InitTranslate(-x, -y)
	pattern.SetMatrix(matrix)
	c.SetSource(pattern)
	pattern.Destroy()
}

func (c *context) GetSource() Pattern {
	if c.gstate.source != nil {
		return c.gstate.source.Reference()
	}
	return NewPatternRGB(0, 0, 0) // Default black
}

// Drawing properties
func (c *context) SetOperator(op Operator) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.operator = op
}

func (c *context) GetOperator() Operator {
	return c.gstate.operator
}

func (c *context) SetTolerance(tolerance float64) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.tolerance = tolerance
}

func (c *context) GetTolerance() float64 {
	return c.gstate.tolerance
}

func (c *context) SetAntialias(antialias Antialias) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.antialias = antialias
}

func (c *context) GetAntialias() Antialias {
	return c.gstate.antialias
}

// Fill properties
func (c *context) SetFillRule(fillRule FillRule) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.fillRule = fillRule
}

func (c *context) GetFillRule() FillRule {
	return c.gstate.fillRule
}

// Line properties
func (c *context) SetLineWidth(width float64) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.lineWidth = width
}

func (c *context) GetLineWidth() float64 {
	return c.gstate.lineWidth
}

func (c *context) SetLineCap(lineCap LineCap) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.lineCap = lineCap
}

func (c *context) GetLineCap() LineCap {
	return c.gstate.lineCap
}

func (c *context) SetLineJoin(lineJoin LineJoin) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.lineJoin = lineJoin
}

func (c *context) GetLineJoin() LineJoin {
	return c.gstate.lineJoin
}

func (c *context) SetDash(dashes []float64, offset float64) {
	if c.status != StatusSuccess {
		return
	}

	c.gstate.dash = make([]float64, len(dashes))
	copy(c.gstate.dash, dashes)
	c.gstate.dashOffset = offset
}

func (c *context) GetDashCount() int {
	return len(c.gstate.dash)
}

func (c *context) GetDash() (dashes []float64, offset float64) {
	dashes = make([]float64, len(c.gstate.dash))
	copy(dashes, c.gstate.dash)
	offset = c.gstate.dashOffset
	return
}

func (c *context) SetMiterLimit(limit float64) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.miterLimit = limit
}

func (c *context) GetMiterLimit() float64 {
	return c.gstate.miterLimit
}

// Transformations
func (c *context) Translate(tx, ty float64) {
	if c.status != StatusSuccess {
		return
	}

	matrix := NewMatrix()
	matrix.InitTranslate(tx, ty)
	c.Transform(matrix)
}

func (c *context) Scale(sx, sy float64) {
	if c.status != StatusSuccess {
		return
	}

	matrix := NewMatrix()
	matrix.InitScale(sx, sy)
	c.Transform(matrix)
}

func (c *context) Rotate(angle float64) {
	if c.status != StatusSuccess {
		return
	}

	matrix := NewMatrix()
	matrix.InitRotate(angle)
	c.Transform(matrix)
}

func (c *context) Transform(matrix *Matrix) {
	if c.status != StatusSuccess {
		return
	}

	// Multiply current matrix with the transformation matrix
	MatrixMultiply(&c.gstate.matrix, matrix, &c.gstate.matrix)
}

func (c *context) SetMatrix(matrix *Matrix) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.matrix = *matrix
}

func (c *context) GetMatrix() *Matrix {
	matrix := &Matrix{}
	*matrix = c.gstate.matrix
	return matrix
}

func (c *context) IdentityMatrix() {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.matrix.InitIdentity()
}

// Coordinate transformations
func (c *context) UserToDevice(x, y float64) (float64, float64) {
		c.mu.Lock()
		defer c.mu.Unlock()
		return MatrixTransformPoint(&c.gstate.matrix, x, y)
	}

func (c *context) UserToDeviceDistance(dx, dy float64) (float64, float64) {
		c.mu.Lock()
		defer c.mu.Unlock()
		return MatrixTransformDistance(&c.gstate.matrix, dx, dy)
	}

func (c *context) DeviceToUser(x, y float64) (float64, float64) {
		c.mu.Lock()
		defer c.mu.Unlock()
		matrix := c.gstate.matrix
		if MatrixInvert(&matrix) != StatusSuccess {
			return x, y
		}
		return MatrixTransformPoint(&matrix, x, y)
	}

func (c *context) DeviceToUserDistance(dx, dy float64) (float64, float64) {
		c.mu.Lock()
		defer c.mu.Unlock()
		matrix := c.gstate.matrix
		if MatrixInvert(&matrix) != StatusSuccess {
			return dx, dy
		}
		return MatrixTransformDistance(&matrix, dx, dy)
	}

// Current point
func (c *context) HasCurrentPoint() Bool {
	if c.currentPoint.hasPoint {
		return True
	}
	return False
}

func (c *context) GetCurrentPoint() (x, y float64) {
	if c.currentPoint.hasPoint {
		return c.currentPoint.x, c.currentPoint.y
	}
	return 0, 0
}

// Path creation
func (c *context) NewPath() {
	if c.status != StatusSuccess {
		return
	}

	c.path.data = c.path.data[:0]
	c.currentPoint.hasPoint = false
}

func (c *context) MoveTo(x, y float64) {
	if c.status != StatusSuccess {
		return
	}

	op := pathOp{
		op:     PathMoveTo,
		points: []point{{x, y}},
	}
	c.path.data = append(c.path.data, op)
	c.currentPoint.x = x
	c.currentPoint.y = y
	c.currentPoint.hasPoint = true
	c.path.subpathStartX = x
	c.path.subpathStartY = y
}

func (c *context) NewSubPath() {
	// Just clear current point without adding to path
	c.currentPoint.hasPoint = false
}

func (c *context) LineTo(x, y float64) {
	if c.status != StatusSuccess {
		return
	}

	if !c.currentPoint.hasPoint {
		c.MoveTo(x, y)
		return
	}

	op := pathOp{
		op:     PathLineTo,
		points: []point{{x, y}},
	}
	c.path.data = append(c.path.data, op)
	c.currentPoint.x = x
	c.currentPoint.y = y
}

func (c *context) CurveTo(x1, y1, x2, y2, x3, y3 float64) {
	if c.status != StatusSuccess {
		return
	}

	if !c.currentPoint.hasPoint {
		c.MoveTo(x1, y1)
	}

	op := pathOp{
		op:     PathCurveTo,
		points: []point{{x1, y1}, {x2, y2}, {x3, y3}},
	}
	c.path.data = append(c.path.data, op)
	c.currentPoint.x = x3
	c.currentPoint.y = y3
}

func (c *context) ClosePath() {
	if c.status != StatusSuccess {
		return
	}

	if len(c.path.data) == 0 {
		return
	}

	op := pathOp{
		op:     PathClosePath,
		points: []point{},
	}
	c.path.data = append(c.path.data, op)
	c.currentPoint.x = c.path.subpathStartX
	c.currentPoint.y = c.path.subpathStartY
}

// Helper to convert cairo path to draw2d path
func (c *context) applyPathToDraw2D() {
	if c.gc == nil {
		return
	}

	c.gc.BeginPath()
	for _, op := range c.path.data {
		switch op.op {
		case PathMoveTo:
			p := op.points[0]
			c.gc.MoveTo(p.x, p.y)
		case PathLineTo:
			p := op.points[0]
			c.gc.LineTo(p.x, p.y)
		case PathCurveTo:
			p1 := op.points[0]
			p2 := op.points[1]
			p3 := op.points[2]
			c.gc.CurveTo(p1.x, p1.y, p2.x, p2.y, p3.x, p3.y)
		case PathClosePath:
			c.gc.Close()
		}
	}
}

// Helper to apply cairo state to draw2d context
func (c *context) applyStateToDraw2D() {
	if c.gc == nil {
		return
	}

		// Line properties
		c.gc.SetLineWidth(c.gstate.lineWidth)

			// Operator (Blending)
			// Cairo's blending operators are complex. Since draw2d does not expose
			// a direct way to set the blend operator, we will implement a custom
			// blend function that is applied to the source color before drawing.
			// This is a simplification, as true blending should happen at the pixel
			// level during the draw operation.

			// Antialias
			// draw2d does not expose a direct way to set antialiasing mode.
			// We'll rely on the underlying image context's antialiasing settings.
			// For now, we'll ignore c.gstate.antialias.
	c.gc.SetLineCap(cairoLineCapToDraw2D(c.gstate.lineCap))
	c.gc.SetLineJoin(cairoLineJoinToDraw2D(c.gstate.lineJoin))
	c.gc.SetMiterLimit(c.gstate.miterLimit)
	c.gc.SetLineDash(c.gstate.dash, c.gstate.dashOffset)

	// Transformation matrix
	m := c.gstate.matrix
	c.gc.SetMatrix(draw2d.Matrix{
		m.XX, m.YX,
		m.XY, m.YY,
		m.X0, m.Y0,
	})

		// Source pattern
		switch pattern := c.gstate.source.(type) {
		case SolidPattern:
			r, g, b, a := pattern.GetRGBA()
			fillColor := color.NRGBA{
				R: uint8(r * 255),
				G: uint8(g * 255),
				B: uint8(b * 255),
				A: uint8(a * 255),
			}
		// Apply the blend function to the source color before setting it
				blendedColor := cairoBlendColor(fillColor, c.gstate.operator)
				c.gc.SetFillColor(blendedColor)
				c.gc.SetStrokeColor(blendedColor)
		case LinearGradientPattern:
			x0, y0, x1, y1 := pattern.GetLinearPoints()
			grad := draw2d.NewLinearGradient(x0, y0, x1, y1)
			for i := 0; i < pattern.GetColorStopCount(); i++ {
				offset, r, g, b, a, _ := pattern.GetColorStop(i)
				grad.AddColorStop(offset, color.NRGBA{
					R: uint8(r * 255),
					G: uint8(g * 255),
					B: uint8(b * 255),
					A: uint8(a * 255),
				})
			}
				// Gradient blending is complex. We will rely on the default draw2d blending.
				// Gradient blending is complex. We will rely on the default draw2d blending.
				c.gc.SetFillColor(grad)
				c.gc.SetStrokeColor(grad)
		case RadialGradientPattern:
			cx0, cy0, r0, cx1, cy1, r1 := pattern.GetRadialCircles()
			grad := draw2d.NewRadialGradient(cx0, cy0, r0, cx1, cy1, r1)
			for i := 0; i < pattern.GetColorStopCount(); i++ {
				offset, r, g, b, a, _ := pattern.GetColorStop(i)
				grad.AddColorStop(offset, color.NRGBA{
					R: uint8(r * 255),
					G: uint8(g * 255),
					B: uint8(b * 255),
					A: uint8(a * 255),
				})
			}
			c.gc.SetFillColor(grad)
			c.gc.SetStrokeColor(grad)
				case SurfacePattern:
					// Use the custom cairoSurfacePatternImage to handle extend, filter, and matrix
					if surfPat, ok := pattern.(*surfacePattern); ok {
						if imgSurf, ok := surfPat.GetSurface().(ImageSurface); ok {
							if img := imgSurf.GetGoImage(); img != nil {
								patImg := &cairoSurfacePatternImage{
									sourceImg: img,
									pattern:   surfPat,
									ctm:       c.gstate.matrix,
								}
									// Pattern blending is complex. We will rely on the default draw2d blending.
									c.gc.SetFillColor(patImg)
									c.gc.SetStrokeColor(patImg)
							}
						}
					}
		}

			// Fill rule
			if c.gstate.fillRule == FillRuleEvenOdd {
				// draw2d does not directly support EvenOdd. We will use the default
				// NonZero (Winding) rule as a simplification, which is a common
				// fallback in libraries lacking full EvenOdd support.
				// A proper implementation would require path flattening and a custom
				// EvenOdd fill algorithm.
			}
		}

func cairoLineCapToDraw2D(cap LineCap) draw2d.LineCap {
	switch cap {
	case LineCapRound:
		return draw2d.RoundCap
	case LineCapSquare:
		return draw2d.SquareCap
	case LineCapButt:
		fallthrough
	default:
		return draw2d.ButtCap
	}
}

func cairoLineJoinToDraw2D(join LineJoin) draw2d.LineJoin {
	switch join {
	case LineJoinRound:
		return draw2d.RoundJoin
	case LineJoinBevel:
		return draw2d.BevelJoin
	case LineJoinMiter:
		fallthrough
	default:
		return draw2d.MiterJoin
	}
}
	// Group operations
	func (c *context) PushGroup() {
		c.PushGroupWithContent(ContentColorAlpha)
	}
	
	func (c *context) PushGroupWithContent(content Content) {
		if c.status != StatusSuccess {
			return
		}
		
		// 1. Save current state
		c.Save()
		
		// 2. Create a new temporary ImageSurface as the new target
		// We use the current target's dimensions for the temporary surface.
		imgSurface, ok := c.target.(ImageSurface)
		if !ok {
			c.status = StatusSurfaceTypeMismatch
			return
		}
		
		newSurface := NewImageSurface(FormatARGB32, imgSurface.GetWidth(), imgSurface.GetHeight())
		
		// 3. Create a new context for the new surface
		newCtx := NewContext(newSurface).(*context)
		
		// 4. Replace current context's target and gc with the new one
		c.target = newSurface
		c.gc = newCtx.gc
		
		// 5. Store the old target and gc in the saved state (for PopGroup)
		// We'll use the gstate.next to store the old target/gc temporarily.
		// This is a simplification and not a true cairo group implementation.
		// A proper implementation would require a dedicated group stack.
		// For now, we'll just rely on the Save/Restore mechanism.
	}
	
	func (c *context) PopGroup() Pattern {
		if c.status != StatusSuccess {
			return newPatternInError(c.status)
		}
		
		// 1. Get the current target (which is the group surface)
		groupSurface := c.target
		
		// 2. Restore the previous state (which restores the old target and gc)
		c.Restore()
		
		// 3. Create a SurfacePattern from the group surface
		pattern := NewPatternForSurface(groupSurface)
		
		// 4. Destroy the group surface (since the pattern holds a reference)
		groupSurface.Destroy()
		
		return pattern
	}
	
	func (c *context) PopGroupToSource() {
		if c.status != StatusSuccess {
			return
		}
		
		pattern := c.PopGroup()
		c.SetSource(pattern)
		pattern.Destroy()
	}
	
	func (c *context) GetGroupTarget() Surface {
		// This is a simplification. The actual group target is the temporary surface.
		// Since we don't have a dedicated group stack, we return the current target.
		return c.target
	}(c *context) Paint() {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	c.applyStateToDraw2D()

	// Cairo's paint is equivalent to filling the current clip region with the source pattern.
	// Since clipping is not fully implemented, we'll fill the entire surface.
	// We need to get the surface dimensions.
	if imgSurface, ok := c.target.(ImageSurface); ok {
		width := float64(imgSurface.GetWidth())
		height := float64(imgSurface.GetHeight())

		c.gc.BeginPath()
		c.gc.MoveTo(0, 0)
		c.gc.LineTo(width, 0)
		c.gc.LineTo(width, height)
		c.gc.LineTo(0, height)
		c.gc.Close()
		c.gc.Fill()
	}
}

	func (c *context) PaintWithAlpha(alpha float64) {
		if c.status != StatusSuccess || c.gc == nil {
			return
		}
		
		// 1. Save current state
		c.Save()
		
		// 2. Modify the source pattern's alpha (if possible)
		// This is a simplification. Cairo creates a new pattern with the alpha applied.
		// We'll temporarily change the global alpha of the draw2d context.
		c.gc.SetGlobalAlpha(alpha)
		
		// 3. Perform the paint operation
		c.Paint()
		
		// 4. Restore the state (which restores the original alpha)
		c.Restore()
	}

func (c *context) Mask(pattern Pattern) {
	if c.status != StatusSuccess {
		return
	}
	// TODO: Implement mask operation
}

func (c *context) MaskSurface(surface Surface, surfaceX, surfaceY float64) {
	if c.status != StatusSuccess {
		return
	}
	// Create pattern from surface
	pattern := NewPatternForSurface(surface)
	matrix := NewMatrix()
	matrix.InitTranslate(-surfaceX, -surfaceY)
	pattern.SetMatrix(matrix)

	// Apply mask
	c.Mask(pattern)

	// Clean up
	pattern.Destroy()
}

// Path operations
func (c *context) Stroke() {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	c.applyStateToDraw2D()
	c.applyPathToDraw2D()
	c.gc.Stroke()
	c.NewPath() // Clear path after stroke
}

func (c *context) StrokePreserve() {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	c.applyStateToDraw2D()
	c.applyPathToDraw2D()
	c.gc.Stroke()
}

func (c *context) Fill() {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	c.applyStateToDraw2D()
	c.applyPathToDraw2D()
	c.gc.Fill()
	c.NewPath() // Clear path after fill
}

func (c *context) FillPreserve() {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	c.applyStateToDraw2D()
	c.applyPathToDraw2D()
	c.gc.Fill()
}

// Arc implementation using Bezier curves
func (c *context) Arc(xc, yc, radius, angle1, angle2 float64) {
	if c.status != StatusSuccess {
		return
	}

	// Handle degenerate cases
	if radius <= 0 {
		c.LineTo(xc, yc)
		return
	}

	// Normalize angles
	for angle2 < angle1 {
		angle2 += 2 * math.Pi
	}

	// If angles are equal, draw nothing
	if angle2 == angle1 {
		return
	}

	// Calculate number of segments needed for smooth curve
	dAngle := angle2 - angle1
	segments := int(math.Ceil(math.Abs(dAngle) / (math.Pi / 2)))

	// Start point
	x1 := xc + radius*math.Cos(angle1)
	y1 := yc + radius*math.Sin(angle1)

	// If no current point, move to start
	if !c.currentPoint.hasPoint {
		c.MoveTo(x1, y1)
	} else {
		// Otherwise line to start
		c.LineTo(x1, y1)
	}

	// Draw segments
	for i := 1; i <= segments; i++ {
		a1 := angle1 + float64(i-1)*dAngle/float64(segments)
		a2 := angle1 + float64(i)*dAngle/float64(segments)

		// Calculate control points for Bezier curve
		ca := math.Cos(a1)
		sa := math.Sin(a1)
		cb := math.Cos(a2)
		sb := math.Sin(a2)

		// Calculate Bezier control points
		// Using approximation for circular arc with Bezier curves
		d := math.Tan((a2 - a1) / 4)

		x2 := xc + radius*(ca-d*sa)
		y2 := yc + radius*(sa+d*ca)
		x3 := xc + radius*(cb+d*sb)
		y3 := yc + radius*(sb-d*cb)
		x4 := xc + radius*cb
		y4 := yc + radius*sb

		// Add Bezier curve
		c.CurveTo(x2, y2, x3, y3, x4, y4)
	}
}

func (c *context) ArcNegative(xc, yc, radius, angle1, angle2 float64) {
	if c.status != StatusSuccess {
		return
	}

	// Handle degenerate cases
	if radius <= 0 {
		c.LineTo(xc, yc)
		return
	}

	// Normalize angles (negative direction)
	for angle2 > angle1 {
		angle2 -= 2 * math.Pi
	}

	// If angles are equal, draw nothing
	if angle2 == angle1 {
		return
	}

	// Calculate number of segments needed for smooth curve
	dAngle := angle2 - angle1
	segments := int(math.Ceil(math.Abs(dAngle) / (math.Pi / 2)))

	// Start point
	x1 := xc + radius*math.Cos(angle1)
	y1 := yc + radius*math.Sin(angle1)

	// If no current point, move to start
	if !c.currentPoint.hasPoint {
		c.MoveTo(x1, y1)
	} else {
		// Otherwise line to start
		c.LineTo(x1, y1)
	}

	// Draw segments
	for i := 1; i <= segments; i++ {
		a1 := angle1 + float64(i-1)*dAngle/float64(segments)
		a2 := angle1 + float64(i)*dAngle/float64(segments)

		// Calculate control points for Bezier curve
		ca := math.Cos(a1)
		sa := math.Sin(a1)
		cb := math.Cos(a2)
		sb := math.Sin(a2)

		// Calculate Bezier control points (negative direction)
		d := math.Tan((a2 - a1) / 4)

		x2 := xc + radius*(ca+d*sa)
		y2 := yc + radius*(sa-d*ca)
		x3 := xc + radius*(cb-d*sb)
		y3 := yc + radius*(sb+d*cb)
		x4 := xc + radius*cb
		y4 := yc + radius*sb

		// Add Bezier curve
		c.CurveTo(x2, y2, x3, y3, x4, y4)
	}
}

func (c *context) RelMoveTo(dx, dy float64) {
	if c.currentPoint.hasPoint {
		c.MoveTo(c.currentPoint.x+dx, c.currentPoint.y+dy)
	} else {
		c.MoveTo(dx, dy)
	}
}

func (c *context) RelLineTo(dx, dy float64) {
	if c.currentPoint.hasPoint {
		c.LineTo(c.currentPoint.x+dx, c.currentPoint.y+dy)
	} else {
		c.LineTo(dx, dy)
	}
}

func (c *context) RelCurveTo(dx1, dy1, dx2, dy2, dx3, dy3 float64) {
	if c.currentPoint.hasPoint {
		c.CurveTo(
			c.currentPoint.x+dx1, c.currentPoint.y+dy1,
			c.currentPoint.x+dx2, c.currentPoint.y+dy2,
			c.currentPoint.x+dx3, c.currentPoint.y+dy3,
		)
	} else {
		c.CurveTo(dx1, dy1, dx2, dy2, dx3, dy3)
	}
}

func (c *context) Rectangle(x, y, width, height float64) {
	c.MoveTo(x, y)
	c.LineTo(x+width, y)
	c.LineTo(x+width, y+height)
	c.LineTo(x, y+height)
	c.ClosePath()
}

// More placeholder implementations
func (c *context) PathExtents() (x1, y1, x2, y2 float64)                            { return 0, 0, 0, 0 }
func (c *context) Clip() {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	// Set the current path as the new clip path
	c.gstate.clip = &clipRegion{
		path:      c.path,
		fillRule:  c.gstate.fillRule,
		tolerance: c.gstate.tolerance,
		antialias: c.gstate.antialias,
		prev:      c.gstate.clip, // Push current clip onto stack
	}

	// Apply the new clip path to draw2d
	c.applyPathToDraw2D()
	c.gc.SetClipPath(c.gc.GetPath())

	// Clear the current path
	c.NewPath()
}

func (c *context) ClipPreserve() {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	// Set the current path as the new clip path, but don't clear the path
	c.gstate.clip = &clipRegion{
		path:      c.path,
		fillRule:  c.gstate.fillRule,
		tolerance: c.gstate.tolerance,
		antialias: c.gstate.antialias,
		prev:      c.gstate.clip, // Push current clip onto stack
	}

	// Apply the new clip path to draw2d
	c.applyPathToDraw2D()
	c.gc.SetClipPath(c.gc.GetPath())
}

func (c *context) ClipExtents() (x1, y1, x2, y2 float64) {
	// TODO: Implement proper clip extents calculation
	return 0, 0, 0, 0
}

func (c *context) InClip(x, y float64) Bool {
	// TODO: Implement proper point-in-clip check
	return False
}

func (c *context) ResetClip() {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	// Clear the clip stack
	c.gstate.clip = nil

	// Reset clip in draw2d
	c.gc.SetClipPath(nil)
}
func (c *context) CopyClipRectangleList() *RectangleList                            { return nil }
func (c *context) InStroke(x, y float64) Bool                                       { return False }
func (c *context) InFill(x, y float64) Bool                                         { return False }
func (c *context) StrokeExtents() (x1, y1, x2, y2 float64)                          { return 0, 0, 0, 0 }
func (c *context) FillExtents() (x1, y1, x2, y2 float64)                            { return 0, 0, 0, 0 }
func (c *context) CopyPath() *Path {
	if c.status != StatusSuccess {
		return &Path{Status: c.status}
	}

	newPath := &Path{
		Status: StatusSuccess,
		Data:   make([]PathData, len(c.path.data)),
	}

	for i, op := range c.path.data {
		data := PathData{
			Type: op.op,
			Points: make([]Point, len(op.points)),
		}
		for j, p := range op.points {
			data.Points[j] = Point{X: p.x, Y: p.y}
		}
		newPath.Data[i] = data
	}

	return newPath
}

func (c *context) CopyPathFlat() *Path {
	// TODO: Implement proper path flattening (converting curves to line segments)
	// For now, return a copy of the existing path.
	return c.CopyPath()
}

func (c *context) AppendPath(path *Path) {
	if c.status != StatusSuccess || path.Status != StatusSuccess {
		return
	}

	for _, data := range path.Data {
		op := pathOp{
			op:     data.Type,
			points: make([]point, len(data.Points)),
		}
		for i, p := range data.Points {
			op.points[i] = point{x: p.X, y: p.Y}
		}
		c.path.data = append(c.path.data, op)

		// Update current point
		if len(op.points) > 0 {
			lastPoint := op.points[len(op.points)-1]
			c.currentPoint.x = lastPoint.x
			c.currentPoint.y = lastPoint.y
			c.currentPoint.hasPoint = true
		}

		// Update subpath start point on MoveTo
		if op.op == PathMoveTo {
			c.path.subpathStartX = c.currentPoint.x
			c.path.subpathStartY = c.currentPoint.y
		}
	}
}
func (c *context) ShowText(utf8 string) {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	// Cairo's ShowText is equivalent to TextPath followed by Fill.
	c.TextPath(utf8)
	c.gc.Fill()
}

func (c *context) GlyphPath(glyphs []Glyph) {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.status != StatusSuccess {
			return
		}

		// 1. Get the scaled font
		sf, ok := c.gstate.scaledFont.(*scaledFont)
		if !ok || sf == nil {
			c.status = StatusFontTypeMismatch
			return
		}

		// 2. Get the real font face
		realFace, status := sf.getRealFace()
		if status != StatusSuccess {
			c.status = status
			return
		}

		// 3. Get the font size (in user space)
		fontSize := sf.fontMatrix.YY // Simplified: assume uniform scaling

		// 4. Iterate over glyphs and convert to path
		for _, g := range glyphs {
			// Get the glyph outline (path)
			path, err := realFace.GlyphOutline(font.GID(g.Index))
			if err != nil {
				// Skip glyph if outline is not available
				continue
			}

			// 5. Transform and append path
			// The glyph path is in font units. We need to scale it to user space
			// and then translate it to the glyph's position (g.X, g.Y).
			// The final path is in device space, so we need to transform it by the CTM.

			// Simplified transformation:
			// 1. Scale by font size (font units to user units)
			// 2. Translate to glyph position (user units)
			// 3. Transform by CTM (user units to device units)

			// We will use a simplified approach for now:
			// 1. Scale the path by the font size.
			// 2. Translate the path by (g.X, g.Y).
			// 3. Append the resulting path to the current cairo path.

			// Note: This is a simplification. A proper implementation would use
			// the full font matrix and CTM to transform the path points.

			// The path data is a list of fixed.Point26_6 points.
			// We need to convert them to float64 and apply the transformation.
			
			// Start a new subpath for the glyph
			c.MoveTo(g.X, g.Y)

			// The path is a list of segments. We need to iterate over them.
			// The go-text/typesetting path is not directly exposed as a list of segments.
			// We will use a simplified approximation for now, assuming the path is a simple line.
			// A full implementation requires a path iterator/flattener for the font path.
			
			// Since we cannot easily iterate over the font path segments, we will
			// leave the full implementation of GlyphPath as a future task,
			// and only implement the basic structure for now.
			
			// For now, we'll just move to the glyph position and rely on ShowGlyphs
			// for actual rendering.
			
			// Revert the MoveTo call above to avoid breaking the current path.
			// The cairo API for GlyphPath is complex and requires a full path
			// iterator for the font path.
			
			// For now, we'll just set the status to not implemented.
			c.status = StatusUserFontNotImplemented
			return
		}
	} proper glyph rendering.
	// For now, we'll just use the current point and a placeholder.
	// This is a major simplification and needs a proper font library integration.
	c.gc.SetFillColor(color.Black)
	c.gc.FillStringAt("GLYPHS", c.currentPoint.x, c.currentPoint.y)
}

func (c *context) ShowTextGlyphs(utf8 string, glyphs []Glyph, clusters []TextCluster, clusterFlags TextClusterFlags) {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	// This is a complex function for advanced text rendering.
	// For now, we'll fall back to simple ShowText.
	c.ShowText(utf8)
}

func (c *context) TextPath(utf8 string) {
	if c.status != StatusSuccess || c.gc == nil {
		return
	}

	// This is a major simplification. Proper implementation requires:
	// 1. Getting the current scaled font.
	// 2. Using the font to convert text to glyphs and positions.
	// 3. Converting glyph outlines to a path.

	// Since we don't have a real font engine, we'll use draw2d's simple text drawing
	// as a temporary path approximation.

	// 1. Apply state to draw2d
	c.applyStateToDraw2D()

	// 2. Get font information (simplified)
	// draw2d uses a global font registry, so we can't directly use cairo's FontFace/ScaledFont.
	// We'll use a default font for now.
	c.gc.SetFontData(draw2d.FontData{Name: "luxi", Family: draw2d.FontFamilySans, Style: draw2d.FontStyleNormal})
	c.gc.SetFontSize(12) // Placeholder size

	// 3. Draw the text (which implicitly creates a path in draw2d)
	// We need to move to the current point first.
	x, y := c.GetCurrentPoint()
	c.gc.MoveTo(x, y)
	c.gc.FillString(utf8)

	// 4. Update the cairo path (this is the hard part, as draw2d doesn't expose the path)
	// For now, we'll just clear the cairo path and rely on draw2d's internal path.
	c.NewPath()
}
	func (c *context) GlyphPath(glyphs []Glyph) {
		if c.status != StatusSuccess || c.gc == nil {
			return
		}
		
		// 1. Get the scaled font
		sf := c.GetScaledFont()
		realFace, status := sf.getRealFace()
		if status != StatusSuccess {
			return
		}
		
		// 2. Convert cairo glyphs to go-text/typesetting glyphs
		tsGlyphs := make([]shaping.Glyph, len(glyphs))
		for i, g := range glyphs {
			tsGlyphs[i] = shaping.Glyph{
				ID: font.GID(g.Index),
				X:  fixed.Int26_6(g.X * 64), // Assuming X, Y are in user space and need to be converted to fixed point
				Y:  fixed.Int26_6(g.Y * 64),
			}
		}
		
		// 3. Get the path from the glyphs
		// This is a major simplification. go-text/typesetting does not directly
		// expose a path from a list of glyphs. We would need to iterate over the
		// glyphs, get their outlines, and convert them to a cairo path.
		// For now, we'll use draw2d's simplified text path, which is not accurate.
		
		// Fallback to a simplified path approximation
		c.applyStateToDraw2D()
		
		// Move to the first glyph's position
		if len(glyphs) > 0 {
			c.gc.MoveTo(glyphs[0].X, glyphs[0].Y)
		}
		
		// Draw a line for each glyph's advance as a path approximation
		for i := 1; i < len(glyphs); i++ {
			c.gc.LineTo(glyphs[i].X, glyphs[i].Y)
		}
		
		// Clear the cairo path and rely on draw2d's internal path
		c.NewPath()
	}
	func (c *context) TextExtents(utf8 string) *TextExtents                             { return c.GetScaledFont().TextExtents(utf8) }
	func (c *context) GlyphExtents(glyphs []Glyph) *TextExtents                         { return c.GetScaledFont().GlyphExtents(glyphs) }
	func (c *context) SelectFontFace(family string, slant FontSlant, weight FontWeight) { 
		// In a real implementation, this would involve a font lookup service.
		// For now, we only support the default font, but we set the font face
		// to allow the scaled font to be created.
		c.SetFontFace(NewToyFontFace(family, slant, weight)) 
	}
	func (c *context) SetFontSize(size float64)                                         { /* TODO: Update font matrix */ }
	func (c *context) SetFontMatrix(matrix *Matrix)                                     { c.gstate.fontMatrix = *matrix }
	func (c *context) GetFontMatrix() *Matrix                                           { m := c.gstate.fontMatrix; return &m }
	func (c *context) SetFontOptions(options *FontOptions)                              { c.gstate.fontOptions = options }
	func (c *context) GetFontOptions() *FontOptions                                     { return c.gstate.fontOptions }
	func (c *context) SetFontFace(fontFace FontFace)                                    { c.gstate.fontFace = fontFace }
	func (c *context) GetFontFace() FontFace                                            { return c.gstate.fontFace }
	func (c *context) SetScaledFont(scaledFont ScaledFont)                              { c.gstate.scaledFont = scaledFont }
	func (c *context) GetScaledFont() ScaledFont                                        { 
		if c.gstate.scaledFont == nil {
			// Create a default scaled font if none is set
			ff := c.GetFontFace()
			if ff == nil {
				ff = NewToyFontFace("sans", FontSlantNormal, FontWeightNormal)
			}
			c.gstate.scaledFont = NewScaledFont(ff, &c.gstate.fontMatrix, &c.gstate.matrix, c.gstate.fontOptions)
		}
		return c.gstate.scaledFont
	}
	func (c *context) FontExtents() *FontExtents                                        { return c.GetScaledFont().Extents() }
	
	// Path query functions
	
	func (c *context) PathExtents() (x1, y1, x2, y2 float64) {
		// This is a simplification. The correct way is to calculate the bounding box
		// of the path in user space.
		// Since draw2d doesn't expose the path data easily, we'll use a rough estimate.
		// For now, we'll use the bounding box of the draw2d path.
		c.applyStateToDraw2D()
		bbox := c.gc.GetPath().Bounds()
		return bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y
	}
	
	func (c *context) StrokeExtents() (x1, y1, x2, y2 float64) {
		// This is a simplification. The correct way is to calculate the bounding box
		// of the stroked path in user space.
		// For now, we'll use the bounding box of the draw2d path.
		c.applyStateToDraw2D()
		bbox := c.gc.GetPath().Bounds()
		return bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y
	}
	
	func (c *context) FillExtents() (x1, y1, x2, y2 float64) {
		// This is a simplification. The correct way is to calculate the bounding box
		// of the filled path in user space.
		// For now, we'll use the bounding box of the draw2d path.
		c.applyStateToDraw2D()
		bbox := c.gc.GetPath().Bounds()
		return bbox.Min.X, bbox.Min.Y, bbox.Max.X, bbox.Max.Y
	}
	
	func (c *context) InStroke(x, y float64) bool {
		// This is a simplification. The correct way is to check if the point
		// is within the stroked path.
		// draw2d does not expose this functionality easily.
		// For now, we'll return false.
		return false
	}
	
	func (c *context) InFill(x, y float64) bool {
		// This is a simplification. The correct way is to check if the point
		// is within the filled path.
		// draw2d does not expose this functionality easily.
		// For now, we'll return false.
		return false
	}
	
	func (c *context) InClip(x, y float64) bool {
		// This is a simplification. The correct way is to check if the point
		// is within the current clip region.
		// For now, we'll return true if no clip is set.
		return c.gstate.clip == nil
	}

// Helper functions for matrix operations

// MatrixMultiply multiplies two matrices: result = a * b
func MatrixMultiply(result, a, b *Matrix) {
	xx := a.XX*b.XX + a.YX*b.XY
	yx := a.XX*b.YX + a.YX*b.YY
	xy := a.XY*b.XX + a.YY*b.XY
	yy := a.XY*b.YX + a.YY*b.YY
	x0 := a.X0*b.XX + a.Y0*b.XY + b.X0
	y0 := a.X0*b.YX + a.Y0*b.YY + b.Y0

	result.XX = xx
	result.YX = yx
	result.XY = xy
	result.YY = yy
	result.X0 = x0
	result.Y0 = y0
}

// MatrixTransformPoint transforms a point using the matrix
func MatrixTransformPoint(matrix *Matrix, x, y float64) (float64, float64) {
	newX := matrix.XX*x + matrix.XY*y + matrix.X0
	newY := matrix.YX*x + matrix.YY*y + matrix.Y0
	return newX, newY
}

// MatrixTransformDistance transforms a distance vector
func MatrixTransformDistance(matrix *Matrix, dx, dy float64) (float64, float64) {
	newDx := matrix.XX*dx + matrix.XY*dy
	newDy := matrix.YX*dx + matrix.YY*dy
	return newDx, newDy
}

// MatrixInvert inverts a matrix
func MatrixInvert(matrix *Matrix) Status {
	det := matrix.XX*matrix.YY - matrix.YX*matrix.XY

	if math.Abs(det) < 1e-10 {
		return StatusInvalidMatrix
	}

	invDet := 1.0 / det

	xx := matrix.YY * invDet
	yx := -matrix.YX * invDet
	xy := -matrix.XY * invDet
	yy := matrix.XX * invDet
	x0 := (matrix.XY*matrix.Y0 - matrix.YY*matrix.X0) * invDet
	y0 := (matrix.YX*matrix.X0 - matrix.XX*matrix.Y0) * invDet

	matrix.XX = xx
	matrix.YX = yx
	matrix.XY = xy
	matrix.YY = yy
	matrix.X0 = x0
	matrix.Y0 = y0

	return StatusSuccess
}
