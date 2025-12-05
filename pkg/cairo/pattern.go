package cairo

import (
	"image"
	"image/color"
	"math"
	"sync/atomic"
	"unsafe"
)

// solidPattern implements solid color patterns
type solidPattern struct {
	basePattern
	red, green, blue, alpha float64
}

// surfacePattern implements surface-based patterns  
type surfacePattern struct {
	basePattern
	surface Surface
}

// cairoSurfacePatternImage implements image.Image and draw2d.Pattern,
// handling the transformation, extend, and filter logic.
type cairoSurfacePatternImage struct {
	sourceImg image.Image
	pattern   *surfacePattern
	ctm       Matrix // Current Transformation Matrix from cairo.Context
}

// ColorModel implements image.Image.
func (p *cairoSurfacePatternImage) ColorModel() color.Model {
	return p.sourceImg.ColorModel()
}

// Bounds implements image.Image.
func (p *cairoSurfacePatternImage) Bounds() image.Rectangle {
	// The bounds of the pattern are effectively infinite, but for draw2d
	// we can return the bounds of the target surface, or just the source image.
	// Since draw2d will use the At() method, the bounds are less critical
	// than the At() implementation. Let's use the source image bounds.
	return p.sourceImg.Bounds()
}

// At implements image.Image. This is the core logic.
func (p *cairoSurfacePatternImage) At(x, y int) color.Color {
	// 1. Convert device coordinate (x, y) to user space (ux, uy)
	// This step is implicitly handled by draw2d when it calls At(x, y)
	// on the pattern image, as draw2d's GraphicContext already applies the CTM
	// to the fill/stroke operation before sampling the pattern.
	// However, draw2d's pattern sampling is typically done in device space
	// and then transformed by the pattern's matrix.
	
	// Let's assume (x, y) are coordinates in the pattern's user space,
	// which is what draw2d's SetFillPattern expects after applying the CTM.
	
	// 2. Convert pattern user space (x, y) to pattern source space (sx, sy)
	// Pattern source space = Pattern Matrix Inverse * Pattern User Space
	
	// Copy pattern matrix and invert it
	patMatrix := p.pattern.matrix
	if status := MatrixInvert(&patMatrix); status != StatusSuccess {
		// Fallback to solid black on error
		return color.NRGBA{A: 0xFF}
	}
	
	sx, sy := MatrixTransformPoint(&patMatrix, float64(x), float64(y))
	
	// 3. Apply Extend logic
	srcBounds := p.sourceImg.Bounds()
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())
	
	// Normalize coordinates to [0, srcW) and [0, srcH)
	var finalX, finalY float64
	
	switch p.pattern.extend {
	case ExtendNone:
		if sx < 0 || sx >= srcW || sy < 0 || sy >= srcH {
			return color.NRGBA{A: 0x00} // Transparent
		}
		finalX, finalY = sx, sy
	case ExtendRepeat:
		finalX = math.Mod(sx, srcW)
		if finalX < 0 {
			finalX += srcW
		}
		finalY = math.Mod(sy, srcH)
		if finalY < 0 {
			finalY += srcH
		}
	case ExtendReflect:
		// Reflect logic: 0..W, W..0, 0..W, ...
		finalX = math.Mod(sx, 2*srcW)
		if finalX < 0 {
			finalX += 2 * srcW
		}
		if finalX >= srcW {
			finalX = 2*srcW - finalX
		}
		
		finalY = math.Mod(sy, 2*srcH)
		if finalY < 0 {
			finalY += 2 * srcH
		}
		if finalY >= srcH {
			finalY = 2*srcH - finalY
		}
	case ExtendPad:
		finalX = math.Max(0, math.Min(sx, srcW-1))
		finalY = math.Max(0, math.Min(sy, srcH-1))
	default:
		finalX, finalY = sx, sy // Fallback to no extend
	}
	
		// 4. Apply Filter logic (simplification: nearest neighbor for Fast, bilinear for Good/Best)
		// Since draw2d's At() method is called with integer coordinates, we'll use nearest neighbor.
		// For better filtering, we would need to implement a custom image sampler.
		
		// Convert back to integer coordinates relative to the source image's Min point
		srcX := int(finalX) + srcBounds.Min.X
		srcY := int(finalY) + srcBounds.Min.Y
		
		// 5. Sample color
		// TODO: Implement proper filtering based on p.pattern.filter
		return p.sourceImg.At(srcX, srcY)
}

// gradientPattern is the base for gradient patterns
type gradientPattern struct {
	basePattern
	stops []gradientStop
}

type gradientStop struct {
	offset float64
	red, green, blue, alpha float64
}

// linearGradient implements linear gradient patterns
type linearGradient struct {
	gradientPattern
	x0, y0, x1, y1 float64
}

// radialGradient implements radial gradient patterns  
type radialGradient struct {
	gradientPattern
	cx0, cy0, radius0 float64
	cx1, cy1, radius1 float64
}

// basePattern provides common pattern functionality
type basePattern struct {
	refCount int32
	status   Status
	patternType PatternType
	matrix   Matrix
	extend   Extend
	filter   Filter
	userData map[*UserDataKey]interface{}
}

// NewPatternRGB creates a solid color pattern with RGB values
func NewPatternRGB(red, green, blue float64) Pattern {
	return NewPatternRGBA(red, green, blue, 1.0)
}

// NewPatternRGBA creates a solid color pattern with RGBA values
func NewPatternRGBA(red, green, blue, alpha float64) Pattern {
	pattern := &solidPattern{
		basePattern: basePattern{
			refCount: 1,
			status: StatusSuccess,
			patternType: PatternTypeSolid,
			extend: ExtendNone,
			filter: FilterFast,
			userData: make(map[*UserDataKey]interface{}),
		},
		red: red,
		green: green, 
		blue: blue,
		alpha: alpha,
	}
	pattern.matrix.InitIdentity()
	return pattern
}

// NewPatternForSurface creates a pattern from a surface
func NewPatternForSurface(surface Surface) Pattern {
	if surface == nil {
		return newPatternInError(StatusNullPointer)
	}
	
	pattern := &surfacePattern{
		basePattern: basePattern{
			refCount: 1,
			status: StatusSuccess,
			patternType: PatternTypeSurface,
			extend: ExtendNone,
			filter: FilterFast,
			userData: make(map[*UserDataKey]interface{}),
		},
		surface: surface.Reference(),
	}
	pattern.matrix.InitIdentity()
	return pattern
}

// NewPatternLinear creates a linear gradient pattern
func NewPatternLinear(x0, y0, x1, y1 float64) Pattern {
	pattern := &linearGradient{
		gradientPattern: gradientPattern{
			basePattern: basePattern{
				refCount: 1,
				status: StatusSuccess,
				patternType: PatternTypeLinear,
				extend: ExtendNone,
				filter: FilterFast,
				userData: make(map[*UserDataKey]interface{}),
			},
			stops: make([]gradientStop, 0),
		},
		x0: x0, y0: y0,
		x1: x1, y1: y1,
	}
	pattern.matrix.InitIdentity()
	return pattern
}

// NewPatternRadial creates a radial gradient pattern  
func NewPatternRadial(cx0, cy0, radius0, cx1, cy1, radius1 float64) Pattern {
	pattern := &radialGradient{
		gradientPattern: gradientPattern{
			basePattern: basePattern{
				refCount: 1,
				status: StatusSuccess,
				patternType: PatternTypeRadial,
				extend: ExtendNone,
				filter: FilterFast,
				userData: make(map[*UserDataKey]interface{}),
			},
			stops: make([]gradientStop, 0),
		},
		cx0: cx0, cy0: cy0, radius0: radius0,
		cx1: cx1, cy1: cy1, radius1: radius1,
	}
	pattern.matrix.InitIdentity()
	return pattern
}

func newPatternInError(status Status) Pattern {
	pattern := &solidPattern{
		basePattern: basePattern{
			refCount: 1,
			status: status,
			patternType: PatternTypeSolid,
			userData: make(map[*UserDataKey]interface{}),
		},
	}
	return pattern
}

// Base pattern interface implementation

func (p *basePattern) Reference() Pattern {
	atomic.AddInt32(&p.refCount, 1)
	// Return the actual pattern type, not basePattern
	return p.getPattern()
}

func (p *basePattern) getPattern() Pattern {
	// This is a bit of a hack - in a real implementation we'd need
	// to store a reference to the concrete type
	return nil // This will be overridden in concrete types
}

func (p *basePattern) Destroy() {
	if atomic.AddInt32(&p.refCount, -1) == 0 {
		// Clean up resources specific to pattern type
		p.cleanup()
	}
}

func (p *basePattern) cleanup() {
	// Base cleanup - overridden in concrete types
}

func (p *basePattern) GetReferenceCount() int {
	return int(atomic.LoadInt32(&p.refCount))
}

func (p *basePattern) Status() Status {
	return p.status
}

func (p *basePattern) GetType() PatternType {
	return p.patternType
}

func (p *basePattern) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	if p.status != StatusSuccess {
		return p.status
	}
	
	p.userData[key] = userData
	// TODO: Store destroy function and call it when appropriate
	return StatusSuccess
}

func (p *basePattern) GetUserData(key *UserDataKey) unsafe.Pointer {
	if data, exists := p.userData[key]; exists {
		return data.(unsafe.Pointer)
	}
	return nil
}

func (p *basePattern) SetMatrix(matrix *Matrix) {
	if p.status != StatusSuccess {
		return
	}
	p.matrix = *matrix
}

func (p *basePattern) GetMatrix() *Matrix {
	matrix := &Matrix{}
	*matrix = p.matrix
	return matrix
}

func (p *basePattern) SetExtend(extend Extend) {
	if p.status != StatusSuccess {
		return
	}
	p.extend = extend
}

func (p *basePattern) GetExtend() Extend {
	return p.extend
}

func (p *basePattern) SetFilter(filter Filter) {
	if p.status != StatusSuccess {
		return
	}
	p.filter = filter
}

func (p *basePattern) GetFilter() Filter {
	return p.filter
}

// Solid pattern implementation

func (p *solidPattern) getPattern() Pattern {
	return p
}

func (p *solidPattern) Reference() Pattern {
	atomic.AddInt32(&p.refCount, 1)
	return p
}

func (p *solidPattern) GetRGBA() (red, green, blue, alpha float64) {
	return p.red, p.green, p.blue, p.alpha
}

// Surface pattern implementation

func (p *surfacePattern) getPattern() Pattern {
	return p
}

func (p *surfacePattern) Reference() Pattern {
	atomic.AddInt32(&p.refCount, 1)
	return p
}

func (p *surfacePattern) cleanup() {
	if p.surface != nil {
		p.surface.Destroy()
	}
}

func (p *surfacePattern) GetSurface() Surface {
	return p.surface
}

// Gradient pattern implementation

func (p *gradientPattern) AddColorStopRGB(offset, red, green, blue float64) {
	p.AddColorStopRGBA(offset, red, green, blue, 1.0)
}

func (p *gradientPattern) AddColorStopRGBA(offset, red, green, blue, alpha float64) {
	if p.status != StatusSuccess {
		return
	}
	
	if offset < 0.0 || offset > 1.0 {
		p.status = StatusInvalidIndex
		return
	}
	
	stop := gradientStop{
		offset: offset,
		red: red,
		green: green,
		blue: blue,
		alpha: alpha,
	}
	
	// Insert in sorted order by offset
	inserted := false
	for i, existingStop := range p.stops {
		if offset <= existingStop.offset {
			// Insert at position i
			p.stops = append(p.stops[:i], append([]gradientStop{stop}, p.stops[i:]...)...)
			inserted = true
			break
		}
	}
	
	if !inserted {
		p.stops = append(p.stops, stop)
	}
}

func (p *gradientPattern) GetColorStopCount() int {
	return len(p.stops)
}

func (p *gradientPattern) GetColorStop(index int) (offset, red, green, blue, alpha float64, status Status) {
	if index < 0 || index >= len(p.stops) {
		return 0, 0, 0, 0, 0, StatusInvalidIndex
	}
	
	stop := p.stops[index]
	return stop.offset, stop.red, stop.green, stop.blue, stop.alpha, StatusSuccess
}

// Linear gradient implementation

func (p *linearGradient) getPattern() Pattern {
	return p
}

func (p *linearGradient) Reference() Pattern {
	atomic.AddInt32(&p.refCount, 1)
	return p
}

func (p *linearGradient) GetLinearPoints() (x0, y0, x1, y1 float64) {
	return p.x0, p.y0, p.x1, p.y1
}

// Radial gradient implementation

func (p *radialGradient) getPattern() Pattern {
	return p
}

func (p *radialGradient) Reference() Pattern {
	atomic.AddInt32(&p.refCount, 1)  
	return p
}

func (p *radialGradient) GetRadialCircles() (cx0, cy0, radius0, cx1, cy1, radius1 float64) {
	return p.cx0, p.cy0, p.radius0, p.cx1, p.cy1, p.radius1
}

// Pattern-specific interfaces for type assertions

type SolidPattern interface {
	Pattern
	GetRGBA() (red, green, blue, alpha float64)
}

type SurfacePattern interface {
	Pattern  
	GetSurface() Surface
}

type GradientPattern interface {
	Pattern
	AddColorStopRGB(offset, red, green, blue float64)
	AddColorStopRGBA(offset, red, green, blue, alpha float64)
	GetColorStopCount() int
	GetColorStop(index int) (offset, red, green, blue, alpha float64, status Status)
}

type LinearGradientPattern interface {
	GradientPattern
	GetLinearPoints() (x0, y0, x1, y1 float64)
}

type RadialGradientPattern interface {
	GradientPattern
	GetRadialCircles() (cx0, cy0, radius0, cx1, cy1, radius1 float64)
}