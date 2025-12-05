package cairo

import (
	"math"
	"sync/atomic"
	"unsafe"
)

// context implements the Context interface
type context struct {
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
		next:        c.gstate,
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
	return MatrixTransformPoint(&c.gstate.matrix, x, y)
}

func (c *context) UserToDeviceDistance(dx, dy float64) (float64, float64) {
	return MatrixTransformDistance(&c.gstate.matrix, dx, dy)
}

func (c *context) DeviceToUser(x, y float64) (float64, float64) {
	matrix := c.gstate.matrix
	if MatrixInvert(&matrix) != StatusSuccess {
		return x, y
	}
	return MatrixTransformPoint(&matrix, x, y)
}

func (c *context) DeviceToUserDistance(dx, dy float64) (float64, float64) {
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

// Group operations
func (c *context) PushGroup() {
	if c.status != StatusSuccess {
		return
	}
	// TODO: Implement proper group operations
}

func (c *context) PushGroupWithContent(content Content) {
	if c.status != StatusSuccess {
		return
	}
	// TODO: Implement proper group operations with content
}

func (c *context) PopGroup() Pattern {
	if c.status != StatusSuccess {
		return nil
	}
	// TODO: Implement proper group operations
	return nil
}

func (c *context) PopGroupToSource() {
	if c.status != StatusSuccess {
		return
	}
	// TODO: Implement proper group operations
}

// Drawing operations
func (c *context) Paint() {
	if c.status != StatusSuccess {
		return
	}
	// TODO: Implement paint operation
}

func (c *context) PaintWithAlpha(alpha float64) {
	if c.status != StatusSuccess {
		return
	}
	// Save current source
	oldSource := c.gstate.source

	// Create new source with alpha
	// TODO: Implement proper alpha blending

	// Restore source
	c.gstate.source = oldSource
	if oldSource != nil {
		oldSource.Destroy()
	}
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
	if c.status != StatusSuccess {
		return
	}
	// TODO: Implement stroke operation
	c.NewPath() // Clear path after stroke
}

func (c *context) StrokePreserve() {
	if c.status != StatusSuccess {
		return
	}
	// TODO: Implement stroke operation without clearing path
}

func (c *context) Fill() {
	if c.status != StatusSuccess {
		return
	}
	// TODO: Implement fill operation
	c.NewPath() // Clear path after fill
}

func (c *context) FillPreserve() {
	if c.status != StatusSuccess {
		return
	}
	// TODO: Implement fill operation without clearing path
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
func (c *context) Clip()                                                            { /* TODO */ }
func (c *context) ClipPreserve()                                                    { /* TODO */ }
func (c *context) ClipExtents() (x1, y1, x2, y2 float64)                            { return 0, 0, 0, 0 }
func (c *context) InClip(x, y float64) Bool                                         { return False }
func (c *context) ResetClip()                                                       { /* TODO */ }
func (c *context) CopyClipRectangleList() *RectangleList                            { return nil }
func (c *context) InStroke(x, y float64) Bool                                       { return False }
func (c *context) InFill(x, y float64) Bool                                         { return False }
func (c *context) StrokeExtents() (x1, y1, x2, y2 float64)                          { return 0, 0, 0, 0 }
func (c *context) FillExtents() (x1, y1, x2, y2 float64)                            { return 0, 0, 0, 0 }
func (c *context) CopyPath() *Path                                                  { return nil }
func (c *context) CopyPathFlat() *Path                                              { return nil }
func (c *context) AppendPath(path *Path)                                            { /* TODO */ }
func (c *context) ShowText(utf8 string)                                             { /* TODO */ }
func (c *context) ShowGlyphs(glyphs []Glyph)                                        { /* TODO */ }
func (c *context) ShowTextGlyphs(string, []Glyph, []TextCluster, TextClusterFlags)  { /* TODO */ }
func (c *context) TextPath(utf8 string)                                             { /* TODO */ }
func (c *context) GlyphPath(glyphs []Glyph)                                         { /* TODO */ }
func (c *context) TextExtents(utf8 string) *TextExtents                             { return nil }
func (c *context) GlyphExtents(glyphs []Glyph) *TextExtents                         { return nil }
func (c *context) SelectFontFace(family string, slant FontSlant, weight FontWeight) { /* TODO */ }
func (c *context) SetFontSize(size float64)                                         { /* TODO */ }
func (c *context) SetFontMatrix(matrix *Matrix)                                     { c.gstate.fontMatrix = *matrix }
func (c *context) GetFontMatrix() *Matrix                                           { m := c.gstate.fontMatrix; return &m }
func (c *context) SetFontOptions(options *FontOptions)                              { c.gstate.fontOptions = options }
func (c *context) GetFontOptions() *FontOptions                                     { return c.gstate.fontOptions }
func (c *context) SetFontFace(fontFace FontFace)                                    { c.gstate.fontFace = fontFace }
func (c *context) GetFontFace() FontFace                                            { return c.gstate.fontFace }
func (c *context) SetScaledFont(scaledFont ScaledFont)                              { c.gstate.scaledFont = scaledFont }
func (c *context) GetScaledFont() ScaledFont                                        { return c.gstate.scaledFont }
func (c *context) FontExtents() *FontExtents                                        { return nil }

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
