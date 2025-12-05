package cairo

import (
	"unsafe"
)

// Surface represents cairo_surface_t - drawing target interface
type Surface interface {
	// Reference management
	Reference() Surface
	Destroy()
	GetReferenceCount() int
	
	// Status and properties
	Status() Status
	GetType() SurfaceType
	GetContent() Content
	
	// Device management  
	GetDevice() Device
	
	// User data management
	SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status
	GetUserData(key *UserDataKey) unsafe.Pointer
	
	// Surface operations
	Flush()
	MarkDirty()
	MarkDirtyRectangle(x, y, width, height int)
	
	// Font options
	GetFontOptions() *FontOptions
	
	// Finish operations
	Finish()
	
	// Similar surface creation
	CreateSimilar(content Content, width, height int) Surface
	CreateSimilarImage(format Format, width, height int) Surface
	CreateForRectangle(x, y, width, height float64) Surface
	
	// Transformations
	SetDeviceScale(xScale, yScale float64)
	GetDeviceScale() (xScale, yScale float64)
	SetDeviceOffset(xOffset, yOffset float64) 
	GetDeviceOffset() (xOffset, yOffset float64)
	
	// Fallback resolution
	SetFallbackResolution(xPixelsPerInch, yPixelsPerInch float64)
	GetFallbackResolution() (xPixelsPerInch, yPixelsPerInch float64)
	
	// Copy operations
	CopyPage()
	ShowPage()
}

// Context represents cairo_t - drawing context interface  
type Context interface {
	// Reference management
	Reference() Context
	Destroy()
	GetReferenceCount() int
	
	// Status
	Status() Status
	
	// Target surface
	GetTarget() Surface
	GetGroupTarget() Surface
	
	// User data
	SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status
	GetUserData(key *UserDataKey) unsafe.Pointer
	
	// State management
	Save()
	Restore()
	
	// Group operations
	PushGroup()
	PushGroupWithContent(content Content)
	PopGroup() Pattern
	PopGroupToSource()
	
	// Drawing operations
	Paint()
	PaintWithAlpha(alpha float64)
	Mask(pattern Pattern)
	MaskSurface(surface Surface, surfaceX, surfaceY float64)
	
	// Path operations
	Stroke()
	StrokePreserve()
	Fill()
	FillPreserve()
	
	// Source pattern
	SetSource(source Pattern)
	SetSourceRGB(red, green, blue float64)
	SetSourceRGBA(red, green, blue, alpha float64)
	SetSourceSurface(surface Surface, x, y float64)
	GetSource() Pattern
	
	// Drawing properties
	SetOperator(op Operator)
	GetOperator() Operator
	
	SetTolerance(tolerance float64)
	GetTolerance() float64
	
	SetAntialias(antialias Antialias)
	GetAntialias() Antialias
	
	// Fill properties
	SetFillRule(fillRule FillRule)
	GetFillRule() FillRule
	
	// Line properties
	SetLineWidth(width float64)
	GetLineWidth() float64
	
	SetLineCap(lineCap LineCap)
	GetLineCap() LineCap
	
	SetLineJoin(lineJoin LineJoin)
	GetLineJoin() LineJoin
	
	SetDash(dashes []float64, offset float64)
	GetDashCount() int
	GetDash() (dashes []float64, offset float64)
	
	SetMiterLimit(limit float64)
	GetMiterLimit() float64
	
	// Transformations
	Translate(tx, ty float64)
	Scale(sx, sy float64)
	Rotate(angle float64)
	Transform(matrix *Matrix)
	SetMatrix(matrix *Matrix)
	GetMatrix() *Matrix
	IdentityMatrix()
	
	// Coordinate transformations
	UserToDevice(x, y float64) (float64, float64)
	UserToDeviceDistance(dx, dy float64) (float64, float64)
	DeviceToUser(x, y float64) (float64, float64)  
	DeviceToUserDistance(dx, dy float64) (float64, float64)
	
	// Path creation
	NewPath()
	MoveTo(x, y float64)
	NewSubPath()
	LineTo(x, y float64)
	CurveTo(x1, y1, x2, y2, x3, y3 float64)
	Arc(xc, yc, radius, angle1, angle2 float64)
	ArcNegative(xc, yc, radius, angle1, angle2 float64)
	RelMoveTo(dx, dy float64)
	RelLineTo(dx, dy float64)
	RelCurveTo(dx1, dy1, dx2, dy2, dx3, dy3 float64)
	Rectangle(x, y, width, height float64)
	ClosePath()
	PathExtents() (x1, y1, x2, y2 float64)
	
	// Clipping
	Clip()
	ClipPreserve()
	ClipExtents() (x1, y1, x2, y2 float64)
	InClip(x, y float64) Bool
	ResetClip()
	CopyClipRectangleList() *RectangleList
	
	// Point tests
	InStroke(x, y float64) Bool
	InFill(x, y float64) Bool
	
	// Extents
	StrokeExtents() (x1, y1, x2, y2 float64)
	FillExtents() (x1, y1, x2, y2 float64)
	
	// Current point
	HasCurrentPoint() Bool
	GetCurrentPoint() (x, y float64)
	
	// Path access
	CopyPath() *Path
	CopyPathFlat() *Path
	AppendPath(path *Path)
	
	// Text operations
	ShowText(utf8 string)
	ShowGlyphs(glyphs []Glyph)
	ShowTextGlyphs(utf8 string, glyphs []Glyph, clusters []TextCluster, clusterFlags TextClusterFlags)
	TextPath(utf8 string)
	GlyphPath(glyphs []Glyph)
	TextExtents(utf8 string) *TextExtents
	GlyphExtents(glyphs []Glyph) *TextExtents
	
	// Font operations
	SelectFontFace(family string, slant FontSlant, weight FontWeight)
	SetFontSize(size float64)
	SetFontMatrix(matrix *Matrix)
	GetFontMatrix() *Matrix
	SetFontOptions(options *FontOptions)
	GetFontOptions() *FontOptions
	SetFontFace(fontFace FontFace)
	GetFontFace() FontFace
	SetScaledFont(scaledFont ScaledFont)
	GetScaledFont() ScaledFont
	FontExtents() *FontExtents
}

// Pattern represents cairo_pattern_t - paint source interface
type Pattern interface {
	// Reference management
	Reference() Pattern
	Destroy()
	GetReferenceCount() int
	
	// Status and properties
	Status() Status
	GetType() PatternType
	
	// User data
	SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status
	GetUserData(key *UserDataKey) unsafe.Pointer
	
	// Pattern matrix
	SetMatrix(matrix *Matrix)
	GetMatrix() *Matrix
	
	// Extend mode
	SetExtend(extend Extend)
	GetExtend() Extend
	
	// Filter mode
	SetFilter(filter Filter)
	GetFilter() Filter
}

// Device represents cairo_device_t - rendering backend interface
type Device interface {
	// Reference management
	Reference() Device
	Destroy()
	GetReferenceCount() int
	
	// Status and properties
	Status() Status
	GetType() DeviceType
	
	// User data
	SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status
	GetUserData(key *UserDataKey) unsafe.Pointer
	
	// Device operations
	Acquire() Status
	Release()
	Flush()
	Finish()
}

// FontFace represents cairo_font_face_t - font face interface
type FontFace interface {
	// Reference management
	Reference() FontFace
	Destroy()
	GetReferenceCount() int
	
	// Status and properties
	Status() Status
	GetType() FontType
	
	// User data
	SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status
	GetUserData(key *UserDataKey) unsafe.Pointer
}

// ScaledFont represents cairo_scaled_font_t - scaled font interface
type ScaledFont interface {
	// Reference management
	Reference() ScaledFont
	Destroy()
	GetReferenceCount() int
	
	// Status and properties
	Status() Status
	GetType() FontType
	
	// User data
	SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status
	GetUserData(key *UserDataKey) unsafe.Pointer
	
	// Font properties
	GetFontFace() FontFace
	GetFontMatrix() *Matrix
	GetCTM() *Matrix
	GetScaleMatrix() *Matrix
	GetFontOptions() *FontOptions
	
	// Text measurement
	Extents() *FontExtents
	TextExtents(utf8 string) *TextExtents
	GlyphExtents(glyphs []Glyph) *TextExtents
	TextToGlyphs(x, y float64, utf8 string) (glyphs []Glyph, clusters []TextCluster, clusterFlags TextClusterFlags, status Status)
}

// Additional enum types for interfaces

// SurfaceType represents cairo_surface_type_t
type SurfaceType int

const (
	SurfaceTypeImage SurfaceType = iota
	SurfaceTypePDF
	SurfaceTypePS
	SurfaceTypeXlib
	SurfaceTypeXcb
	SurfaceTypeGlitz
	SurfaceTypeQuartz
	SurfaceTypeWin32
	SurfaceTypeBeOS
	SurfaceTypeDirectFB
	SurfaceTypeSVG
	SurfaceTypeOS2
	SurfaceTypeWin32Printing
	SurfaceTypeQtWin32
	SurfaceTypeRecording
	SurfaceTypeVG
	SurfaceTypeGL
	SurfaceTypeDRM
	SurfaceTypeTee
	SurfaceTypeXML
	SurfaceTypeSkia
	SurfaceTypeSubsurface
	SurfaceTypeCogl
)

// PatternType represents cairo_pattern_type_t
type PatternType int

const (
	PatternTypeSolid PatternType = iota
	PatternTypeSurface
	PatternTypeLinear
	PatternTypeRadial
	PatternTypeMesh
	PatternTypeRasterSource
)

// DeviceType represents cairo_device_type_t
type DeviceType int

const (
	DeviceTypeDRM DeviceType = iota
	DeviceTypeGL
	DeviceTypeScript
	DeviceTypeXcb
	DeviceTypeXlib
	DeviceTypeXML
	DeviceTypeCogl
	DeviceTypeWin32
	DeviceTypeInvalid
)

// FontType represents cairo_font_type_t
type FontType int

const (
	FontTypeToy FontType = iota
	FontTypeFt
	FontTypeWin32
	FontTypeQuartz
	FontTypeUser
	FontTypeDwrite
)

// Extend represents cairo_extend_t - pattern extend modes
type Extend int

const (
	ExtendNone Extend = iota
	ExtendRepeat
	ExtendReflect
	ExtendPad
)

// Filter represents cairo_filter_t - pattern filter modes
type Filter int

const (
	FilterFast Filter = iota
	FilterGood
	FilterBest
	FilterNearest
	FilterBilinear
	FilterGaussian
)

// Additional data structures

// Path represents cairo_path_t
type Path struct {
	Status   Status
	Data     []*PathData
	NumData  int
}

// PathDataType represents cairo_path_data_type_t
type PathDataType int

const (
	PathMoveTo PathDataType = iota
	PathLineTo
	PathCurveTo
	PathClosePath
)

// PathData represents cairo_path_data_t
type PathData struct {
	Header struct {
		Type   PathDataType
		Length int
	}
	Point struct {
		X, Y float64
	}
}

// RectangleList represents cairo_rectangle_list_t
type RectangleList struct {
	Status         Status
	Rectangles     []*Rectangle
	NumRectangles  int
}

// FontOptions represents cairo_font_options_t
type FontOptions struct {
	// This will be implemented as a concrete type with methods
	// to match the interface requirements
}