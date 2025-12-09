package cairo

import (
	"image"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"image/color"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

// GroupSurface is a temporary surface used for group operations.
type GroupSurface struct {
	Surface
	originalTarget Surface
	originalGC     draw2d.GraphicContext
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
		imgSurf := target.(ImageSurface)
		goImage := imgSurf.GetGoImage()
		if goImage != nil {
			ctx.gc = draw2dimg.NewGraphicContext(goImage.(*image.RGBA))
		} else {
			dummyImage := image.NewRGBA(image.Rect(0, 0, imgSurf.GetWidth(), imgSurf.GetHeight()))
			ctx.gc = draw2dimg.NewGraphicContext(dummyImage)
		}

		// Apply Y-axis flip compensation for ImageSurface to match Cairo's coordinate system
		// In Cairo, the default coordinate system has Y growing upward, but image formats
		// have Y growing downward. We need to flip the Y axis and translate to match.
		ctx.gstate.matrix.InitIdentity()
		ctx.gstate.matrix.YY = -1.0
		ctx.gstate.matrix.Y0 = float64(imgSurf.GetHeight())

		// Also apply the transformation to the draw2d context to ensure consistency
		ctx.gc.SetMatrixTransform(draw2d.Matrix{
			1.0, 0.0,
			0.0, -1.0,
			0.0, float64(imgSurf.GetHeight()),
		})
	case *pdfSurface:
		// Create a draw2d PDF context
		dummyImage := image.NewRGBA(image.Rect(0, 0, int(s.width), int(s.height)))
		ctx.gc = draw2dimg.NewGraphicContext(dummyImage)
		// Store a reference in the surface for Finish()
	case *svgSurface:
		// Create a draw2d SVG context
		dummyImage := image.NewRGBA(image.Rect(0, 0, int(s.width), int(s.height)))
		ctx.gc = draw2dimg.NewGraphicContext(dummyImage)
		// Store a reference in the surface for Finish()
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
	// Matrix is already initialized for ImageSurface above
	if ctx.gstate.matrix.XX == 0 && ctx.gstate.matrix.YY == 0 && ctx.gstate.matrix.XY == 0 && ctx.gstate.matrix.YX == 0 {
		ctx.gstate.matrix.InitIdentity()
	}

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

// Status returns the current status of the context.
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
func (c *context) Save() error {
	if c.status != StatusSuccess {
		return newError(c.status, "")
	}

	// Create a copy of current state
	newState := &graphicsState{
		source:       c.gstate.source.Reference(),
		operator:     c.gstate.operator,
		tolerance:    c.gstate.tolerance,
		antialias:    c.gstate.antialias,
		fillRule:     c.gstate.fillRule,
		lineWidth:    c.gstate.lineWidth,
		lineCap:      c.gstate.lineCap,
		lineJoin:     c.gstate.lineJoin,
		miterLimit:   c.gstate.miterLimit,
		matrix:       c.gstate.matrix,
		fontMatrix:   c.gstate.fontMatrix,
		fontOptions:  c.gstate.fontOptions, // TODO: Copy font options
		clip:         c.gstate.clip,        // Clip is part of the graphics state
		next:         c.gstate,
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
	return nil
}

func (c *context) Restore() error {
	if c.status != StatusSuccess {
		return newError(c.status, "")
	}

	if c.gstate.next == nil {
		c.status = StatusInvalidRestore
		return newError(StatusInvalidRestore, "")
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
	// This is a simplification; a proper implementation would need to store the draw2d path
	// or re-create it from the cairo path structure.
	// For now, we'll just reset the clip.
	// Note: draw2d doesn't have SetClipPath method, so we skip this for now

	return nil
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
	// TODO: Implement full Porter-Duff compositing logic in the drawing pipeline
	// (e.g., in applyStateToDraw2D or a custom draw2d implementation)
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

	// Sync 1.18's ft-font-accuracy-new: AntialiasBest implies higher precision
	if antialias == AntialiasBest {
		// This is a placeholder for setting a higher precision flag in the underlying font system
		// For draw2d, we can set a lower tolerance for path flattening
		c.gstate.tolerance = 0.01 // A smaller tolerance for better path accuracy
	} else if antialias == AntialiasDefault {
		c.gstate.tolerance = 0.1 // Default tolerance
	}
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
			c.gc.CubicCurveTo(p1.x, p1.y, p2.x, p2.y, p3.x, p3.y)
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
	// Note: draw2d doesn't have SetMiterLimit method
	c.gc.SetLineDash(c.gstate.dash, c.gstate.dashOffset)

	// Transformation matrix
	m := c.gstate.matrix
	c.gc.SetMatrixTransform(draw2d.Matrix{
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

		fontSize := math.Hypot(c.gstate.fontMatrix.XX, c.gstate.fontMatrix.YX)
		c.gc.SetFontSize(fontSize)
	case LinearGradientPattern:
		// Gradient patterns are complex and not fully supported in draw2d
		// For now, use a solid color approximation
		c.gc.SetFillColor(color.NRGBA{R: 128, G: 128, B: 128, A: 255})
		c.gc.SetStrokeColor(color.NRGBA{R: 128, G: 128, B: 128, A: 255})
	case RadialGradientPattern:
		// Gradient patterns are complex and not fully supported in draw2d
		// For now, use a solid color approximation
		c.gc.SetFillColor(color.NRGBA{R: 128, G: 128, B: 128, A: 255})
		c.gc.SetStrokeColor(color.NRGBA{R: 128, G: 128, B: 128, A: 255})
	case SurfacePattern:
		// Surface patterns are complex and not fully supported in draw2d
		// For now, use a solid color approximation
		c.gc.SetFillColor(color.NRGBA{R: 128, G: 128, B: 128, A: 255})
		c.gc.SetStrokeColor(color.NRGBA{R: 128, G: 128, B: 128, A: 255})
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
	newCtx := NewContext(newSurface)

	// 4. Replace current context's target and gc with the new one
	c.target = newSurface
	if ctxImpl, ok := newCtx.(*context); ok {
		c.gc = ctxImpl.gc
	}

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

func (c *context) Paint() error {
	if c.status != StatusSuccess || c.gc == nil {
		return newError(c.status, "")
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
	return nil
}

func (c *context) PaintWithAlpha(alpha float64) error {
	if c.status != StatusSuccess || c.gc == nil {
		return newError(c.status, "")
	}

	// 1. Save current state
	if err := c.Save(); err != nil {
		return err
	}

	// 2. Modify the source pattern's alpha (if possible)
	// This is a simplification. Cairo creates a new pattern with the alpha applied.
	// Note: draw2d doesn't have SetGlobalAlpha method, so we skip this for now

	// 3. Perform the paint operation
	if err := c.Paint(); err != nil {
		return err
	}

	// 4. Restore the state (which restores the original alpha)
	return c.Restore()
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
func (c *context) Stroke() error {
	if c.status != StatusSuccess || c.gc == nil {
		return newError(c.status, "")
	}

	c.applyStateToDraw2D()
	c.applyPathToDraw2D()
	c.gc.Stroke()
	c.NewPath() // Clear path after stroke
	return nil
}

func (c *context) StrokePreserve() error {
	if c.status != StatusSuccess || c.gc == nil {
		return newError(c.status, "")
	}

	c.applyStateToDraw2D()
	c.applyPathToDraw2D()
	c.gc.Stroke()
	return nil
}

func (c *context) Fill() error {
	if c.status != StatusSuccess || c.gc == nil {
		return newError(c.status, "")
	}

	c.applyStateToDraw2D()
	c.applyPathToDraw2D()
	c.gc.Fill()
	c.NewPath() // Clear path after fill
	return nil
}

func (c *context) FillPreserve() error {
	if c.status != StatusSuccess || c.gc == nil {
		return newError(c.status, "")
	}

	c.applyStateToDraw2D()
	c.applyPathToDraw2D()
	c.gc.Fill()
	return nil
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
func (c *context) PathExtents() (x1, y1, x2, y2 float64) { return 0, 0, 0, 0 }
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
	// Note: draw2d doesn't have SetClipPath method, so we skip this for now

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
	// Note: draw2d doesn't have SetClipPath method, so we skip this for now
}

func (c *context) ClipExtents() (x1, y1, x2, y2 float64) {
	if c.status != StatusSuccess || c.gstate.clip == nil {
		return 0, 0, 0, 0
	}

	// For now, we'll return the extents of the clipping path.
	// A proper implementation would consider the intersection of the path and the surface bounds.
	// Note: path.extents() method doesn't exist, so we return default values
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
	// Note: draw2d doesn't have SetClipPath method, so we skip this for now
}
func (c *context) CopyClipRectangleList() *RectangleList   { return nil }
func (c *context) InStroke(x, y float64) Bool              { return False }
func (c *context) InFill(x, y float64) Bool                { return False }
func (c *context) StrokeExtents() (x1, y1, x2, y2 float64) { return 0, 0, 0, 0 }
func (c *context) FillExtents() (x1, y1, x2, y2 float64)   { return 0, 0, 0, 0 }
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
			Type:   op.op,
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
	if c.status != StatusSuccess {
		return &Path{Status: c.status}
	}

	// Flattening converts curves to line segments
	// For now, we'll just return a copy of the path
	// A proper implementation would flatten all curves
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

// ShowText is a simplified version of ShowTextGlyphs that performs shaping internally.
func (c *context) ShowText(utf8 string) {
	if c.status != StatusSuccess {
		return
	}

	// Ensure we have a scaled font
	sf := c.GetScaledFont()
	if sf == nil {
		c.status = StatusFontTypeMismatch
		return
	}
	defer sf.Destroy()

	// Get current point or use (0, 0)
	x, y := c.GetCurrentPoint()
	if !c.currentPoint.hasPoint {
		x, y = 0, 0
	}

	// Perform text shaping to get glyphs
	glyphs, clusters, clusterFlags, status := sf.TextToGlyphs(x, y, utf8)
	if status != StatusSuccess {
		c.status = status
		return
	}

	// Draw the glyphs
	c.ShowTextGlyphs(utf8, glyphs, clusters, clusterFlags)
}

// ShowTextGlyphs draws the given glyphs by converting them to paths and filling.
func (c *context) ShowTextGlyphs(utf8 string, glyphs []Glyph, clusters []TextCluster, flags TextClusterFlags) {
	if c.status != StatusSuccess {
		return
	}

	// Ensure we have a scaled font
	sf := c.GetScaledFont()
	if sf == nil {
		c.status = StatusFontTypeMismatch
		return
	}
	defer sf.Destroy()

	// Method 1: Use glyph paths (proper Cairo way)
	// This converts each glyph to its outline path and fills it
	for _, glyph := range glyphs {
		// Get the glyph path
		sfImpl, ok := sf.(*scaledFont)
		if !ok {
			continue
		}

		glyphPath, err := sfImpl.GlyphPath(glyph.Index)
		if err != nil || glyphPath == nil {
			continue
		}

		// Save current state
		c.Save()

		// Translate to glyph position
		c.Translate(glyph.X, glyph.Y)

		// Apply proper Y axis scaling to prevent text flipping in downward Y coordinate systems
		// In Cairo, the default coordinate system has Y growing downward, but font glyphs
		// are designed for Y growing upward. We need to flip the Y axis for proper text orientation.
		// Check if we need to flip the Y axis based on the font matrix
		fontMatrix := c.GetFontMatrix()
		if fontMatrix.YY > 0 {
			// In a coordinate system where Y grows downward and font matrix has positive YY,
			// we need to flip the Y axis for proper text orientation
			c.Scale(1, -1)
		}

		// Append the glyph path to current path
		c.AppendPath(glyphPath)

		// Fill the glyph
		c.Fill()

		// Restore state
		c.Restore()
	}

	// Update current point to the position after the last glyph
	if len(glyphs) > 0 {
		lastGlyph := glyphs[len(glyphs)-1]

		// Get text extents to calculate the final position
		extents := c.TextExtents(utf8)

		// Correctly update the current point after drawing text
		c.currentPoint.x = lastGlyph.X + extents.XAdvance
		c.currentPoint.y = lastGlyph.Y
		c.currentPoint.hasPoint = true
	}
}

// GlyphPath adds the given glyphs to the current path.
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

	// 2. Iterate over glyphs and convert to path
	for _, g := range glyphs {
		// Get the glyph path from the scaled font
		glyphPath, err := sf.GlyphPath(g.Index)
		if err != nil || glyphPath == nil {
			// Skip glyphs without outlines
			continue
		}

		// Save current transformation
		savedMatrix := c.gstate.matrix

		// Translate to glyph position
		c.Translate(g.X, g.Y)

		// Apply proper Y axis scaling to prevent text flipping in downward Y coordinate systems
		// In Cairo, the default coordinate system has Y growing downward, but font glyphs
		// are designed for Y growing upward. We need to flip the Y axis for proper text orientation.
		// Check if we need to flip the Y axis based on the font matrix
		if c.gstate.fontMatrix.YY > 0 {
			// In a coordinate system where Y grows downward and font matrix has positive YY,
			// we need to flip the Y axis for proper text orientation
			c.Scale(1, -1)
		}

		// Append the glyph path to current path
		c.AppendPath(glyphPath)

		// Restore transformation
		c.gstate.matrix = savedMatrix
	}
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

// Font operations
func (c *context) SelectFontFace(family string, slant FontSlant, weight FontWeight) {
	if c.status != StatusSuccess {
		return
	}
	fontFace := NewToyFontFace(family, slant, weight)
	c.SetFontFace(fontFace)
	fontFace.Destroy()
}

func (c *context) SetFontSize(size float64) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.fontMatrix.InitScale(size, size)
}

func (c *context) SetFontMatrix(matrix *Matrix) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.fontMatrix = *matrix
}

func (c *context) GetFontMatrix() *Matrix {
	m := c.gstate.fontMatrix
	return &m
}

func (c *context) SetFontOptions(options *FontOptions) {
	if c.status != StatusSuccess {
		return
	}
	c.gstate.fontOptions = options.Copy()
}

func (c *context) GetFontOptions() *FontOptions {
	if c.gstate.fontOptions == nil {
		return NewFontOptions()
	}
	return c.gstate.fontOptions.Copy()
}

func (c *context) SetFontFace(fontFace FontFace) {
	if c.status != StatusSuccess {
		return
	}
	if c.gstate.fontFace != nil {
		c.gstate.fontFace.Destroy()
	}
	c.gstate.fontFace = fontFace.Reference()
}

func (c *context) GetFontFace() FontFace {
	if c.gstate.fontFace == nil {
		return nil
	}
	return c.gstate.fontFace.Reference()
}

func (c *context) SetScaledFont(scaledFont ScaledFont) {
	if c.status != StatusSuccess {
		return
	}
	if c.gstate.scaledFont != nil {
		c.gstate.scaledFont.Destroy()
	}
	c.gstate.scaledFont = scaledFont.Reference()
}

func (c *context) GetScaledFont() ScaledFont {
	if c.gstate.scaledFont == nil {
		// Create a scaled font from current font face and matrices
		if c.gstate.fontFace == nil {
			c.gstate.fontFace = NewToyFontFace("sans", FontSlantNormal, FontWeightNormal)
		}
		c.gstate.scaledFont = NewScaledFont(
			c.gstate.fontFace,
			&c.gstate.fontMatrix,
			&c.gstate.matrix,
			c.gstate.fontOptions,
		)
	}
	return c.gstate.scaledFont.Reference()
}

func (c *context) FontExtents() *FontExtents {
	sf := c.GetScaledFont()
	if sf == nil {
		return &FontExtents{}
	}
	defer sf.Destroy()
	return sf.Extents()
}

func (c *context) TextExtents(utf8 string) *TextExtents {
	sf := c.GetScaledFont()
	if sf == nil {
		return &TextExtents{}
	}
	defer sf.Destroy()
	return sf.TextExtents(utf8)
}

func (c *context) GlyphExtents(glyphs []Glyph) *TextExtents {
	sf := c.GetScaledFont()
	if sf == nil {
		return &TextExtents{}
	}
	defer sf.Destroy()
	return sf.GlyphExtents(glyphs)
}

func (c *context) ShowGlyphs(glyphs []Glyph) {
	c.ShowTextGlyphs("", glyphs, nil, 0)
}

// TextPath adds the text to the current path as glyph outlines.
func (c *context) TextPath(utf8 string) {
	if c.status != StatusSuccess {
		return
	}

	// Ensure we have a scaled font
	sf := c.GetScaledFont()
	if sf == nil {
		c.status = StatusFontTypeMismatch
		return
	}
	defer sf.Destroy()

	// Get current point or use (0, 0)
	x, y := c.GetCurrentPoint()
	if !c.currentPoint.hasPoint {
		x, y = 0, 0
	}

	// Perform text shaping to get glyphs
	glyphs, _, _, status := sf.TextToGlyphs(x, y, utf8)
	if status != StatusSuccess {
		c.status = status
		return
	}

	// Add glyphs to path
	c.GlyphPath(glyphs)

	// Update current point
	if len(glyphs) > 0 {
		extents := c.TextExtents(utf8)
		c.currentPoint.x = x + extents.XAdvance
		c.currentPoint.y = y + extents.YAdvance
		c.currentPoint.hasPoint = true
	}
}
