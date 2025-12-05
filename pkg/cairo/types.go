package cairo

import (
	"math"
	"unsafe"
)

// Bool represents cairo_bool_t
type Bool int

const (
	False Bool = 0
	True  Bool = 1
)

// Status represents cairo_status_t - error status codes
type Status int

const (
	StatusSuccess Status = iota
	StatusNoMemory
	StatusInvalidRestore
	StatusInvalidPopGroup
	StatusNoCurrentPoint
	StatusInvalidMatrix
	StatusInvalidStatus
	StatusNullPointer
	StatusInvalidString
	StatusInvalidPathData
	StatusReadError
	StatusWriteError
	StatusSurfaceFinished
	StatusSurfaceTypeMismatch
	StatusPatternTypeMismatch
	StatusInvalidContent
	StatusInvalidFormat
	StatusInvalidVisual
	StatusFileNotFound
	StatusInvalidDash
	StatusInvalidDscComment
	StatusInvalidIndex
	StatusClipNotRepresentable
	StatusTempFileError
	StatusInvalidStride
	StatusFontTypeMismatch
	StatusUserFontImmutable
	StatusUserFontError
	StatusNegativeCount
	StatusInvalidClusters
	StatusInvalidSlant
	StatusInvalidWeight
	StatusInvalidSize
	StatusUserFontNotImplemented
	StatusDeviceTypeMismatch
	StatusDeviceError
	StatusInvalidMeshConstruction
	StatusDeviceFinished
	StatusJbig2GlobalMissing
	StatusPngError
	StatusFreetypeError
	StatusWin32GdiError
	StatusTagError
	StatusDwriteError
	StatusSvgFontError
	StatusLastStatus
)

// Content represents cairo_content_t - surface content types
type Content int

const (
	ContentColor      Content = 0x1000
	ContentAlpha      Content = 0x2000
	ContentColorAlpha Content = 0x3000
)

// Format represents cairo_format_t - pixel formats for image surfaces
type Format int

const (
	FormatInvalid   Format = -1
	FormatARGB32    Format = 0
	FormatRGB24     Format = 1
	FormatA8        Format = 2
	FormatA1        Format = 3
	FormatRGB16565  Format = 4
	FormatRGB30     Format = 5
	FormatRGB96F    Format = 6
	FormatRGBA128F  Format = 7
)

// Dither represents cairo_dither_t - dithering modes
type Dither int

const (
	DitherNone    Dither = iota
	DitherDefault
	DitherFast
	DitherGood
	DitherBest
)

// Operator represents cairo_operator_t - compositing operators
type Operator int

const (
	OperatorClear Operator = iota
	OperatorSource
	OperatorOver
	OperatorIn
	OperatorOut
	OperatorAtop
	OperatorDest
	OperatorDestOver
	OperatorDestIn
	OperatorDestOut
	OperatorDestAtop
	OperatorXor
	OperatorAdd
	OperatorSaturate
	OperatorMultiply
	OperatorScreen
	OperatorOverlay
	OperatorDarken
	OperatorLighten
	OperatorColorDodge
	OperatorColorBurn
	OperatorHardLight
	OperatorSoftLight
	OperatorDifference
	OperatorExclusion
	OperatorHslHue
	OperatorHslSaturation
	OperatorHslColor
	OperatorHslLuminosity
)

// Antialias represents cairo_antialias_t - antialiasing modes
type Antialias int

const (
	AntialiasDefault Antialias = iota
	AntialiasNone
	AntialiasGray
	AntialiasSubpixel
	AntialiasFast
	AntialiasGood
	AntialiasBest
)

// FillRule represents cairo_fill_rule_t - fill rule for paths
type FillRule int

const (
	FillRuleWinding  FillRule = iota
	FillRuleEvenOdd
)

// LineCap represents cairo_line_cap_t - line cap styles
type LineCap int

const (
	LineCapButt   LineCap = iota
	LineCapRound
	LineCapSquare
)

// LineJoin represents cairo_line_join_t - line join styles  
type LineJoin int

const (
	LineJoinMiter LineJoin = iota
	LineJoinRound
	LineJoinBevel
)

// Matrix represents cairo_matrix_t - 2D affine transformation matrix
type Matrix struct {
	XX, YX float64
	XY, YY float64
	X0, Y0 float64
}

// NewMatrix creates an identity matrix
func NewMatrix() *Matrix {
	return &Matrix{
		XX: 1.0, YX: 0.0,
		XY: 0.0, YY: 1.0,
		X0: 0.0, Y0: 0.0,
	}
}

// InitIdentity initializes matrix to identity
func (m *Matrix) InitIdentity() {
	m.XX = 1.0; m.YX = 0.0
	m.XY = 0.0; m.YY = 1.0
	m.X0 = 0.0; m.Y0 = 0.0
}

// InitTranslate initializes matrix to translation
func (m *Matrix) InitTranslate(tx, ty float64) {
	m.InitIdentity()
	m.X0 = tx
	m.Y0 = ty
}

// InitScale initializes matrix to scaling
func (m *Matrix) InitScale(sx, sy float64) {
	m.InitIdentity()
	m.XX = sx
	m.YY = sy
}

// InitRotate initializes matrix to rotation
func (m *Matrix) InitRotate(radians float64) {
	s := Sin(radians)
	c := Cos(radians)
	
	m.XX = c; m.YX = s
	m.XY = -s; m.YY = c
	m.X0 = 0.0; m.Y0 = 0.0
}

// Rectangle represents cairo_rectangle_t - floating point rectangle
type Rectangle struct {
	X, Y          float64
	Width, Height float64
}

// RectangleInt represents cairo_rectangle_int_t - integer rectangle
type RectangleInt struct {
	X, Y          int
	Width, Height int
}

// UserDataKey represents cairo_user_data_key_t - key for user data
type UserDataKey struct {
	Unused int
}

// DestroyFunc represents cairo_destroy_func_t - cleanup callback
type DestroyFunc func(data unsafe.Pointer)

// WriteFunc represents cairo_write_func_t - write callback for surfaces
type WriteFunc func(closure interface{}, data []byte) Status

// ReadFunc represents cairo_read_func_t - read callback for surfaces  
type ReadFunc func(closure interface{}, data []byte) Status

// TextExtents represents cairo_text_extents_t - text measurement
type TextExtents struct {
	XBearing, YBearing float64
	Width, Height      float64
	XAdvance, YAdvance float64
}

// FontExtents represents cairo_font_extents_t - font metrics
type FontExtents struct {
	Ascent, Descent    float64
	Height             float64
	MaxXAdvance        float64
	MaxYAdvance        float64
}

// Glyph represents cairo_glyph_t - positioned glyph
type Glyph struct {
	Index uint64
	X, Y  float64
}

// TextCluster represents cairo_text_cluster_t - text cluster mapping
type TextCluster struct {
	NumBytes  int
	NumGlyphs int
}

// TextClusterFlags represents cairo_text_cluster_flags_t - cluster flags
type TextClusterFlags int

const (
	TextClusterFlagBackward TextClusterFlags = 0x00000001
)

// FontSlant represents cairo_font_slant_t - font slant styles
type FontSlant int

const (
	FontSlantNormal  FontSlant = iota
	FontSlantItalic
	FontSlantOblique
)

// FontWeight represents cairo_font_weight_t - font weight styles
type FontWeight int

const (
	FontWeightNormal FontWeight = iota
	FontWeightBold
)

// Helper math functions
func Sin(x float64) float64 {
	return math.Sin(x)
}

func Cos(x float64) float64 {
	return math.Cos(x)
}
