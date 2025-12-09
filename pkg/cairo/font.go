package cairo

import (
	"math"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/go-text/typesetting/di"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/opentype/api"
	apifont "github.com/go-text/typesetting/opentype/api/font"
	"github.com/go-text/typesetting/shaping"
	"golang.org/x/image/math/fixed"
)

var defaultFont font.Face

func init() {
	// Load default font and initialize draw2d fonts
	f, _ := GetDefaultFont()
	defaultFont = f
	InitDraw2DFonts()
}

// getFontKey creates a lookup key for font cache
func getFontKey(family string, slant FontSlant, weight FontWeight) string {
	familyKey := family
	if family == "sans-serif" || family == "sans" {
		familyKey = "sans"
	} else if family == "serif" {
		familyKey = "serif"
	} else if family == "monospace" || family == "mono" {
		familyKey = "mono"
	} else {
		familyKey = "sans" // Fallback
	}

	slantKey := "regular"
	if slant == FontSlantItalic || slant == FontSlantOblique {
		slantKey = "italic"
	}

	weightKey := ""
	if weight == FontWeightBold {
		weightKey = "bold"
	}

	if weightKey != "" && slantKey == "italic" {
		return familyKey + "-bolditalic"
	} else if weightKey != "" {
		return familyKey + "-bold"
	} else if slantKey == "italic" {
		return familyKey + "-italic"
	}
	return familyKey + "-regular"
}

// ---------------- Font options (cairo_font_options_t) ----------------

// FontOptions represents cairo_font_options_t - font rendering options
// inspired by the C API in cplus/src/cairo.h around cairo_font_options_t.
type FontOptions struct {
	Status        Status
	Antialias     Antialias
	SubpixelOrder SubpixelOrder
	HintStyle     HintStyle
	HintMetrics   HintMetrics
	ColorMode     ColorMode
	ColorPalette  uint

	// CustomPalette stores optional per-index RGBA colors in user-space 0..1
	CustomPalette map[uint]Color
}

// Color represents an RGBA color with float components in [0,1].
type Color struct {
	R, G, B, A float64
}

// NewFontOptions creates a new FontOptions with default values.
func NewFontOptions() *FontOptions {
	return &FontOptions{
		Status:        StatusSuccess,
		Antialias:     AntialiasDefault,
		SubpixelOrder: SubpixelOrderDefault,
		HintStyle:     HintStyleDefault,
		HintMetrics:   HintMetricsDefault,
		ColorMode:     ColorModeDefault,
		ColorPalette:  0,
		CustomPalette: make(map[uint]Color),
	}
}

// Copy returns a deep copy of the font options.
func (o *FontOptions) Copy() *FontOptions {
	if o == nil {
		return nil
	}
	copy := *o
	if o.CustomPalette != nil {
		copy.CustomPalette = make(map[uint]Color, len(o.CustomPalette))
		for k, v := range o.CustomPalette {
			copy.CustomPalette[k] = v
		}
	}
	return &copy
}

// Merge merges values from other into o, following cairo_font_options_merge
// semantics: "default" values in o are replaced by concrete values in other.
func (o *FontOptions) Merge(other *FontOptions) {
	if o == nil || other == nil {
		return
	}
	if other.Antialias != AntialiasDefault {
		o.Antialias = other.Antialias
	}
	if other.SubpixelOrder != SubpixelOrderDefault {
		o.SubpixelOrder = other.SubpixelOrder
	}
	if other.HintStyle != HintStyleDefault {
		o.HintStyle = other.HintStyle
	}
	if other.HintMetrics != HintMetricsDefault {
		o.HintMetrics = other.HintMetrics
	}
	if other.ColorMode != ColorModeDefault {
		o.ColorMode = other.ColorMode
	}
	if other.ColorPalette != 0 {
		o.ColorPalette = other.ColorPalette
	}
	for k, v := range other.CustomPalette {
		o.SetCustomPaletteColor(k, v.R, v.G, v.B, v.A)
	}
}

// Equal reports whether two FontOptions are equal.
func (o *FontOptions) Equal(other *FontOptions) bool {
	if o == nil || other == nil {
		return o == other
	}
	if o.Antialias != other.Antialias ||
		o.SubpixelOrder != other.SubpixelOrder ||
		o.HintStyle != other.HintStyle ||
		o.HintMetrics != other.HintMetrics ||
		o.ColorMode != other.ColorMode ||
		o.ColorPalette != other.ColorPalette {
		return false
	}
	if len(o.CustomPalette) != len(other.CustomPalette) {
		return false
	}
	for k, v := range o.CustomPalette {
		ov, ok := other.CustomPalette[k]
		if !ok || v != ov {
			return false
		}
	}
	return true
}

// Hash returns a stable hash value for the font options.
func (o *FontOptions) Hash() uint64 {
	if o == nil {
		return 0
	}
	// Simple FNV-1a style hash over the fields.
	var h uint64 = 1469598103934665603
	add := func(v uint64) {
		const prime = 1099511628211
		h ^= v
		h *= prime
	}
	add(uint64(o.Antialias))
	add(uint64(o.SubpixelOrder))
	add(uint64(o.HintStyle))
	add(uint64(o.HintMetrics))
	add(uint64(o.ColorMode))
	add(uint64(o.ColorPalette))
	for k, v := range o.CustomPalette {
		add(uint64(k))
		add(math.Float64bits(v.R))
		add(math.Float64bits(v.G))
		add(math.Float64bits(v.B))
		add(math.Float64bits(v.A))
	}
	return h
}

// Status returns the current status of the FontOptions.
func (o *FontOptions) StatusCode() Status {
	if o == nil {
		return StatusNullPointer
	}
	return o.Status
}

// SetAntialias sets the antialiasing mode.
func (o *FontOptions) SetAntialias(a Antialias) {
	if o == nil {
		return
	}
	o.Antialias = a
}

// GetAntialias returns the antialiasing mode.
func (o *FontOptions) GetAntialias() Antialias {
	if o == nil {
		return AntialiasDefault
	}
	return o.Antialias
}

// SetSubpixelOrder sets subpixel order for subpixel AA.
func (o *FontOptions) SetSubpixelOrder(order SubpixelOrder) {
	if o == nil {
		return
	}
	o.SubpixelOrder = order
}

// GetSubpixelOrder gets subpixel order.
func (o *FontOptions) GetSubpixelOrder() SubpixelOrder {
	if o == nil {
		return SubpixelOrderDefault
	}
	return o.SubpixelOrder
}

// SetHintStyle sets outline hinting style.
func (o *FontOptions) SetHintStyle(style HintStyle) {
	if o == nil {
		return
	}
	o.HintStyle = style
}

// GetHintStyle gets outline hinting style.
func (o *FontOptions) GetHintStyle() HintStyle {
	if o == nil {
		return HintStyleDefault
	}
	return o.HintStyle
}

// SetHintMetrics sets metrics hinting behavior.
func (o *FontOptions) SetHintMetrics(m HintMetrics) {
	if o == nil {
		return
	}
	o.HintMetrics = m
}

// GetHintMetrics gets metrics hinting behavior.
func (o *FontOptions) GetHintMetrics() HintMetrics {
	if o == nil {
		return HintMetricsDefault
	}
	return o.HintMetrics
}

// SetColorMode selects whether color fonts are rendered in color.
func (o *FontOptions) SetColorMode(mode ColorMode) {
	if o == nil {
		return
	}
	o.ColorMode = mode
}

// GetColorMode gets font color mode.
func (o *FontOptions) GetColorMode() ColorMode {
	if o == nil {
		return ColorModeDefault
	}
	return o.ColorMode
}

// GetColorPalette returns the current palette index.
func (o *FontOptions) GetColorPalette() uint {
	if o == nil {
		return 0
	}
	return o.ColorPalette
}

// SetColorPalette sets the active palette index.
func (o *FontOptions) SetColorPalette(idx uint) {
	if o == nil {
		return
	}
	o.ColorPalette = idx
}

// SetCustomPaletteColor sets RGBA for a custom palette index.
func (o *FontOptions) SetCustomPaletteColor(index uint, r, g, b, a float64) {
	if o == nil {
		return
	}
	if o.CustomPalette == nil {
		o.CustomPalette = make(map[uint]Color)
	}
	o.CustomPalette[index] = Color{R: r, G: g, B: b, A: a}
}

// GetCustomPaletteColor retrieves RGBA for a custom palette index.
func (o *FontOptions) GetCustomPaletteColor(index uint) (r, g, b, a float64, status Status) {
	if o == nil {
		return 0, 0, 0, 0, StatusNullPointer
	}
	c, ok := o.CustomPalette[index]
	if !ok {
		return 0, 0, 0, 0, StatusInvalidIndex
	}
	return c.R, c.G, c.B, c.A, StatusSuccess
}

// ---------------- FontFace implementation (cairo_font_face_t) ----------------

// baseFontFace provides common functionality shared by concrete font faces.
type baseFontFace struct {
	refCount int32
	status   Status
	fontType FontType
	userData map[*UserDataKey]interface{}
}

// toyFontFace is a simple implementation mimicking cairo_toy_font_face.
type toyFontFace struct {
	baseFontFace
	family string
	slant  FontSlant
	weight FontWeight

	// Real font face from go-text/typesetting
	realFace font.Face
	fontData []byte
}

// NewToyFontFace creates a toy font face similar to cairo_toy_font_face_create.
func NewToyFontFace(family string, slant FontSlant, weight FontWeight) FontFace {
	ff := &toyFontFace{
		baseFontFace: baseFontFace{
			refCount: 1,
			status:   StatusSuccess,
			fontType: FontTypeToy,
			userData: make(map[*UserDataKey]interface{}),
		},
		family: family,
		slant:  slant,
		weight: weight,
	}

	// Get font key and load font
	fontKey := getFontKey(family, slant, weight)
	face, data, err := LoadEmbeddedFont(fontKey)
	if err != nil {
		// Try loading from assets if family looks like a file
		if strings.Contains(family, "/") || strings.Contains(family, "\\") {
			face, data, err = LoadFontFromFile(family)
		}
		if err != nil {
			// Final fallback to default font
			face, data = GetDefaultFont()
		}
	}

	ff.realFace = face
	ff.fontData = data

	if ff.realFace == nil {
		ff.status = StatusFontTypeMismatch
	}
	return ff
}

// FontFace interface implementation for toyFontFace.

func (f *toyFontFace) Reference() FontFace {
	atomic.AddInt32(&f.refCount, 1)
	return f
}

func (f *toyFontFace) Destroy() {
	if atomic.AddInt32(&f.refCount, -1) == 0 {
		// nothing to free at the moment
	}
}

func (f *toyFontFace) GetReferenceCount() int {
	return int(atomic.LoadInt32(&f.refCount))
}

func (f *toyFontFace) Status() Status {
	return f.status
}

func (f *toyFontFace) GetType() FontType {
	return f.fontType
}

func (f *toyFontFace) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	if f.status != StatusSuccess {
		return f.status
	}
	if f.userData == nil {
		f.userData = make(map[*UserDataKey]interface{})
	}
	f.userData[key] = userData
	// destroy func is currently ignored, consistent with other parts of this package
	_ = destroy
	return StatusSuccess
}

func (f *toyFontFace) GetUserData(key *UserDataKey) unsafe.Pointer {
	if f.userData == nil {
		return nil
	}
	if data, ok := f.userData[key]; ok {
		return data.(unsafe.Pointer)
	}
	return nil
}

// ---------------- ScaledFont implementation (cairo_scaled_font_t) ----------------

type scaledFont struct {
	refCount int32
	status   Status
	fontType FontType

	fontFace FontFace

	fontMatrix Matrix
	ctm        Matrix
	// scaleMatrix is derived from fontMatrix and ctm (for now we keep
	// a copy of fontMatrix as a reasonable approximation for toy fonts).
	scaleMatrix Matrix

	options *FontOptions
}

// NewScaledFont creates a new scaled font similar to cairo_scaled_font_create.
func NewScaledFont(fontFace FontFace, fontMatrix, ctm *Matrix, options *FontOptions) ScaledFont {
	if fontFace == nil {
		return nil
	}
	sf := &scaledFont{
		refCount: 1,
		status:   StatusSuccess,
		fontType: fontFace.GetType(),
		fontFace: fontFace.Reference(),
		options:  options,
	}
	if fontMatrix != nil {
		sf.fontMatrix = *fontMatrix
	} else {
		sf.fontMatrix = *NewMatrix()
	}
	if ctm != nil {
		sf.ctm = *ctm
	} else {
		sf.ctm = *NewMatrix()
	}
	// For our toy implementation we just copy fontMatrix into scaleMatrix.
	sf.scaleMatrix = sf.fontMatrix
	return sf
}

// ScaledFont interface implementation

func (s *scaledFont) Reference() ScaledFont {
	atomic.AddInt32(&s.refCount, 1)
	return s
}

func (s *scaledFont) Destroy() {
	if atomic.AddInt32(&s.refCount, -1) == 0 {
		if s.fontFace != nil {
			s.fontFace.Destroy()
		}
	}
}

func (s *scaledFont) GetReferenceCount() int {
	return int(atomic.LoadInt32(&s.refCount))
}

func (s *scaledFont) Status() Status {
	return s.status
}

func (s *scaledFont) GetType() FontType {
	return s.fontType
}

func (s *scaledFont) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	// For now we store user data in the associated FontFace to keep things simple.
	if s.fontFace == nil {
		return StatusNullPointer
	}
	return s.fontFace.SetUserData(key, userData, destroy)
}

func (s *scaledFont) GetUserData(key *UserDataKey) unsafe.Pointer {
	if s.fontFace == nil {
		return nil
	}
	return s.fontFace.GetUserData(key)
}

func (s *scaledFont) GetFontFace() FontFace {
	if s.fontFace == nil {
		return nil
	}
	return s.fontFace.Reference()
}

func (s *scaledFont) GetFontMatrix() *Matrix {
	m := s.fontMatrix
	return &m
}

func (s *scaledFont) GetCTM() *Matrix {
	m := s.ctm
	return &m
}

func (s *scaledFont) GetScaleMatrix() *Matrix {
	m := s.scaleMatrix
	return &m
}

func (s *scaledFont) GetFontOptions() *FontOptions {
	if s.options == nil {
		return NewFontOptions()
	}
	return s.options.Copy()
}

// getRealFace returns the underlying font.Face and checks for errors.
func (s *scaledFont) getRealFace() (font.Face, Status) {
	if s.fontFace == nil {
		return nil, StatusNullPointer
	}
	toy, ok := s.fontFace.(*toyFontFace)
	if !ok || toy.realFace == nil {
		return nil, StatusFontTypeMismatch
	}
	return toy.realFace, StatusSuccess
}

// Extents returns font extents using the real font face.
func (s *scaledFont) Extents() *FontExtents {
	fe := &FontExtents{}

	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		// Fallback to toy extents if real face is not available
		return s.toyExtentsFallback()
	}

	// Get font metrics from go-text/typesetting
	// The font matrix defines the scale and transformation.
	// We need to calculate the point size from the font matrix.
	// Cairo's font matrix is typically a scale matrix (size in user space units).
	// We'll use the average of the scale factors as the nominal size.

	// Scale factor from font matrix
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)

	// Font metrics are in font units (FUnits). We need to convert them to user space units.
	// FUnits to user space: FUnits * (scale / unitsPerEm)
	unitsPerEm := float64(realFace.Upem())

	// Ascent, Descent, Height in FUnits
	metrics, _ := realFace.FontHExtents()
	ascentFUnits := float64(metrics.Ascender)
	descentFUnits := float64(metrics.Descender)
	lineGapFUnits := float64(metrics.LineGap)

	// Convert to user space units
	fe.Ascent = ascentFUnits * sx / unitsPerEm
	fe.Descent = -descentFUnits * sy / unitsPerEm // Descent is negative in FUnits, cairo expects positive
	fe.Height = fe.Ascent + fe.Descent + lineGapFUnits*sy/unitsPerEm
	fe.LineGap = lineGapFUnits * sy / unitsPerEm

	// Max advance is a guess without shaping a string
	fe.MaxXAdvance = sx
	fe.MaxYAdvance = 0

	// Calculate underline metrics
	fe.UnderlinePosition = -fe.Descent * 0.5
	fe.UnderlineThickness = (fe.Ascent + fe.Descent) * 0.05

	// Approximate cap height and x-height
	fe.CapHeight = fe.Ascent * 0.7 // Typical ratio
	fe.XHeight = fe.Ascent * 0.5   // Typical ratio

	return fe
}

// toyExtentsFallback returns toy font extents based on the derived font size.
func (s *scaledFont) toyExtentsFallback() *FontExtents {
	// Use average of xx and yy scale as size; fall back to 12 if zero.
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	size := (sx + sy) * 0.5
	if size == 0 {
		size = 12
	}
	fe := &FontExtents{}
	fe.Ascent = size * 0.8
	fe.Descent = size * 0.2
	fe.Height = fe.Ascent + fe.Descent
	fe.LineGap = size * 0.2 // Typical line gap
	fe.MaxXAdvance = size
	fe.MaxYAdvance = 0
	fe.UnderlinePosition = -fe.Descent * 0.5
	fe.UnderlineThickness = size * 0.05
	fe.CapHeight = fe.Ascent * 0.7 // Typical ratio
	fe.XHeight = fe.Ascent * 0.5   // Typical ratio
	return fe
}

// TextExtents computes text extents using the real font face and shaping.
func (s *scaledFont) TextExtents(utf8 string) *TextExtents {
	ext := &TextExtents{}

	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		return s.toyTextExtentsFallback(utf8)
	}

	// 1. Shape the text
	runes := []rune(utf8)
	input := shaping.Input{
		Text:      runes,
		RunStart:  0,
		RunEnd:    len(runes),
		Direction: di.DirectionLTR,
		Face:      realFace,
		Size:      fixed.I(12), // Default size, will be scaled by font matrix
	}
	output := (&shaping.HarfbuzzShaper{}).Shape(input)

	// 2. Calculate extents from shaped output
	// Scale factor from font matrix
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	unitsPerEm := float64(realFace.Upem())

	// Calculate total advance and bounds
	var totalAdvance fixed.Int26_6
	var minX, minY, maxX, maxY float64
	firstGlyph := true

	for _, g := range output.Glyphs {
		totalAdvance += g.XAdvance

		// Get glyph outline for bounds calculation
		glyphData := realFace.GlyphData(api.GID(g.GlyphID))
		if outline, ok := glyphData.(api.GlyphOutline); ok {
			// Convert outline points to user space
			for _, seg := range outline.Segments {
				for _, arg := range seg.Args {
					x := float64(arg.X) * sx / unitsPerEm
					y := float64(arg.Y) * sy / unitsPerEm

					// Add glyph position
					x += float64(g.XOffset) / 64.0 * sx / unitsPerEm
					y -= float64(g.YOffset) / 64.0 * sy / unitsPerEm

					// For the first glyph, initialize bounds
					if firstGlyph {
						minX, maxX = x, x
						minY, maxY = y, y
						firstGlyph = false
					} else {
						if x < minX {
							minX = x
						}
						if x > maxX {
							maxX = x
						}
						if y < minY {
							minY = y
						}
						if y > maxY {
							maxY = y
						}
					}
				}
			}
		}
	}

	// Convert to user space units
	ext.XAdvance = float64(totalAdvance) / 64.0 * sx / unitsPerEm
	ext.YAdvance = 0

	// Set proper width and height based on actual bounds
	ext.Width = maxX - minX
	ext.Height = maxY - minY
	ext.XBearing = minX
	ext.YBearing = -maxY // Negative because Y axis is inverted in Cairo

	return ext
}

// toyTextExtentsFallback computes naive text extents assuming fixed advance width.
func (s *scaledFont) toyTextExtentsFallback(utf8 string) *TextExtents {
	size := s.toyExtentsFallback().Ascent + s.toyExtentsFallback().Descent
	advancePerRune := size * 0.6

	var runeCount int
	for range utf8 {
		runeCount++
	}

	ext := &TextExtents{}
	ext.Width = float64(runeCount) * advancePerRune
	ext.Height = s.toyExtentsFallback().Height
	ext.XAdvance = ext.Width
	ext.YAdvance = 0
	ext.XBearing = 0
	ext.YBearing = -s.toyExtentsFallback().Ascent
	return ext
}

// GlyphExtents computes extents based on glyph positions.
func (s *scaledFont) GlyphExtents(glyphs []Glyph) *TextExtents {
	if len(glyphs) == 0 {
		return &TextExtents{}
	}
	// Assume glyph positions are advances from origin.
	last := glyphs[len(glyphs)-1]
	ext := &TextExtents{}
	ext.XAdvance = last.X
	ext.YAdvance = last.Y
	ext.Width = last.X
	ext.Height = s.Extents().Height
	ext.XBearing = 0
	ext.YBearing = -s.Extents().Ascent
	return ext
}

// GlyphPath returns the path for a single glyph ID.
func (s *scaledFont) GlyphPath(glyphID uint64) (*Path, error) {
	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		return nil, newError(status, "failed to get real font face")
	}

	// Load the glyph outline from the font face
	gid := api.GID(glyphID)
	glyphData := realFace.GlyphData(gid)

	// Extract outline from glyph data
	outline, ok := glyphData.(api.GlyphOutline)
	if !ok {
		return nil, newError(StatusFontTypeMismatch, "glyph has no outline")
	}

	// Convert the outline to cairo.Path
	cairoPath := &Path{
		Status: StatusSuccess,
		Data:   make([]PathData, 0),
	}

	// Scale factor from font matrix
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	unitsPerEm := float64(realFace.Upem())

	// FUnits to user space: FUnits * (scale / unitsPerEm)
	funitToUser := func(f float32, scale float64) float64 {
		return float64(f) * scale / unitsPerEm
	}

	// Iterate over the path segments
	var pathPoints []Point
	for _, seg := range outline.Segments {
		switch seg.Op {
		case api.SegmentOpMoveTo:
			point := Point{
				X: funitToUser(seg.Args[0].X, sx),
				Y: funitToUser(seg.Args[0].Y, sy),
			}
			pathPoints = append(pathPoints, point)
		case api.SegmentOpLineTo:
			point := Point{
				X: funitToUser(seg.Args[0].X, sx),
				Y: funitToUser(seg.Args[0].Y, sy),
			}
			pathPoints = append(pathPoints, point)
		case api.SegmentOpQuadTo:
			// Convert quadratic to cubic Bezier
			// For simplicity, we'll add the control point and end point
			p1 := Point{
				X: funitToUser(seg.Args[0].X, sx),
				Y: funitToUser(seg.Args[0].Y, sy),
			}
			p2 := Point{
				X: funitToUser(seg.Args[1].X, sx),
				Y: funitToUser(seg.Args[1].Y, sy),
			}
			pathPoints = append(pathPoints, p1, p1, p2)
		case api.SegmentOpCubeTo:
			p1 := Point{
				X: funitToUser(seg.Args[0].X, sx),
				Y: funitToUser(seg.Args[0].Y, sy),
			}
			p2 := Point{
				X: funitToUser(seg.Args[1].X, sx),
				Y: funitToUser(seg.Args[1].Y, sy),
			}
			p3 := Point{
				X: funitToUser(seg.Args[2].X, sx),
				Y: funitToUser(seg.Args[2].Y, sy),
			}
			pathPoints = append(pathPoints, p1, p2, p3)
		}
	}

	// Apply hinting to the path points
	hintedPoints := s.applyHinting(pathPoints)

	// Convert hinted points back to path data
	// This is a simplified approach - in reality, we'd need to preserve
	// the segment structure while applying hinting
	for i, point := range hintedPoints {
		var pd PathData
		if i == 0 {
			pd.Type = PathMoveTo
		} else {
			pd.Type = PathLineTo
		}
		pd.Points = []Point{point}
		cairoPath.Data = append(cairoPath.Data, pd)
	}

	return cairoPath, nil
}

// GetTextBearingMetrics returns the bearing metrics for a text string
func (s *scaledFont) GetTextBearingMetrics(text string) (xBearing, yBearing float64, status Status) {
	metrics := s.TextExtents(text)
	if metrics == nil {
		return 0, 0, StatusFontTypeMismatch
	}
	return metrics.XBearing, metrics.YBearing, StatusSuccess
}

// GetTextAlignmentOffset calculates the Y offset for text alignment
func (s *scaledFont) GetTextAlignmentOffset(alignment TextAlignment) (float64, Status) {
	fontExtents := s.Extents()
	if fontExtents == nil {
		return 0, StatusFontTypeMismatch
	}
	return GetAlignmentOffset(alignment, fontExtents), StatusSuccess
}

// GetKerning returns the kerning adjustment between two runes
func (s *scaledFont) GetKerning(r1, r2 rune) (float64, Status) {
	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		return 0, status
	}

	// Get the glyph indices for the runes
	gid1, ok1 := realFace.NominalGlyph(r1)
	gid2, ok2 := realFace.NominalGlyph(r2)
	if !ok1 || !ok2 {
		return 0, StatusInvalidGlyph
	}

	// Check if we have kerning data
	var kernValue int16
	if len(realFace.Kern) > 0 {
		// Try Kern tables first
		for _, kernSubtable := range realFace.Kern {
			if kd, ok := kernSubtable.Data.(apifont.Kern0); ok {
				kernValue = kd.KernPair(gid1, gid2)
				break
			}
		}
	} else if len(realFace.Kerx) > 0 {
		// Try Kerx tables if no Kern tables
		for _, kerxSubtable := range realFace.Kerx {
			if kd, ok := kerxSubtable.Data.(apifont.Kern0); ok {
				kernValue = kd.KernPair(gid1, gid2)
				break
			}
		}
	}

	// Scale factor from font matrix
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	unitsPerEm := float64(realFace.Upem())

	// Convert kerning value to user space units
	kerning := float64(kernValue) * sx / unitsPerEm

	return kerning, StatusSuccess
}

// applyHinting applies font hinting based on the font options
func (s *scaledFont) applyHinting(points []Point) []Point {
	// If no options or hinting is disabled, return points as-is
	if s.options == nil || s.options.HintStyle == HintStyleNone {
		return points
	}

	// For now, we'll just return the points as-is since go-text/typesetting
	// doesn't directly support hinting. In a more complete implementation,
	// this would adjust the points based on the hinting style.
	// TODO: Implement actual hinting algorithms
	return points
}

// GetGlyphBearingMetrics returns the bearing metrics for a specific glyph
func (s *scaledFont) GetGlyphBearingMetrics(r rune) (xBearing, yBearing float64, status Status) {
	metrics, status := s.GetGlyphMetrics(r)
	if status != StatusSuccess {
		return 0, 0, status
	}
	return metrics.XBearing, metrics.YBearing, StatusSuccess
}

// GetGlyphMetrics returns detailed metrics for a specific glyph
func (s *scaledFont) GetGlyphMetrics(r rune) (*GlyphMetrics, Status) {
	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		return nil, status
	}

	// Get the glyph index for the rune
	gid, ok := realFace.NominalGlyph(r)
	if !ok || gid == 0 {
		return nil, StatusInvalidGlyph
	}

	// Load glyph outline
	glyphData := realFace.GlyphData(gid)
	outline, ok := glyphData.(api.GlyphOutline)
	if !ok {
		return nil, StatusFontTypeMismatch
	}

	// Scale factor from font matrix
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	unitsPerEm := float64(realFace.Upem())

	// FUnits to user space conversion function
	funitToUser := func(f float32, scale float64) float64 {
		return float64(f) * scale / unitsPerEm
	}

	// Calculate bounding box from outline
	var xmin, xmax, ymin, ymax float64
	firstPoint := true

	for _, seg := range outline.Segments {
		for _, arg := range seg.Args {
			x := funitToUser(arg.X, sx)
			y := funitToUser(arg.Y, sy)

			if firstPoint {
				xmin, xmax = x, x
				ymin, ymax = y, y
				firstPoint = false
			} else {
				if x < xmin {
					xmin = x
				}
				if x > xmax {
					xmax = x
				}
				if y < ymin {
					ymin = y
				}
				if y > ymax {
					ymax = y
				}
			}
		}
	}

	// Get horizontal metrics from the font's hmtx table
	advanceWidth := float64(realFace.HorizontalAdvance(gid)) * sx / unitsPerEm

	// Create metrics
	metrics := &GlyphMetrics{
		Width:    advanceWidth,
		Height:   0, // For horizontal text
		XAdvance: advanceWidth,
		YAdvance: 0, // For horizontal text
		XBearing: xmin,
		YBearing: -ymax, // Negative because Y axis is inverted in Cairo
	}

	// Set bounding box
	metrics.BoundingBox.XMin = xmin
	metrics.BoundingBox.YMin = ymin
	metrics.BoundingBox.XMax = xmax
	metrics.BoundingBox.YMax = ymax

	// Calculate side bearings
	metrics.LSB = xmin
	metrics.RSB = advanceWidth - xmax

	return metrics, StatusSuccess
}

// GetGlyphs returns the glyphs for a given text string.
// This is a simplified version of cairo_scaled_font_get_glyphs, primarily for font subsetting.
func (s *scaledFont) GetGlyphs(utf8 string) (glyphs []Glyph, status Status) {
	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		return nil, status
	}

	// 1. Shape the text
	runes := []rune(utf8)
	input := shaping.Input{
		Text:      runes,
		RunStart:  0,
		RunEnd:    len(runes),
		Direction: di.DirectionLTR,
		Face:      realFace,
		Size:      fixed.I(12),
	}
	output := (&shaping.HarfbuzzShaper{}).Shape(input)

	// 2. Convert shaped output to cairo's Glyph structures
	glyphs = make([]Glyph, len(output.Glyphs))
	for i, g := range output.Glyphs {
		glyphs[i] = Glyph{
			Index: uint64(g.GlyphID),
			X:     0, // Position is not relevant for subsetting
			Y:     0,
		}
	}

	// TODO: Integrate with color font (COLRv0/1) handling for 1.18+ compatibility.

	return glyphs, StatusSuccess
}

// TextToGlyphs performs text shaping to get accurate glyphs and clusters.
func (s *scaledFont) TextToGlyphs(x, y float64, utf8 string) (glyphs []Glyph, clusters []TextCluster, clusterFlags TextClusterFlags, status Status) {
	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		return s.toyTextToGlyphsFallback(x, y, utf8)
	}

	// 1. Shape the text
	runes := []rune(utf8)
	input := shaping.Input{
		Text:      runes,
		RunStart:  0,
		RunEnd:    len(runes),
		Direction: di.DirectionLTR,
		Face:      realFace,
		Size:      fixed.I(12),
	}
	output := (&shaping.HarfbuzzShaper{}).Shape(input)

	// 2. Convert shaped output to cairo's Glyph and TextCluster structures

	// Scale factor from font matrix
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	unitsPerEm := float64(realFace.Upem())

	// Glyphs
	glyphs = make([]Glyph, len(output.Glyphs))
	var curX, curY float64

	// Process each glyph with proper spacing
	for i, g := range output.Glyphs {
		// Position is in user space, relative to the start point (x, y)
		glyphs[i] = Glyph{
			Index: uint64(g.GlyphID),
			X:     x + curX + float64(g.XOffset)/64.0*sx/unitsPerEm,
			Y:     y + curY - float64(g.YOffset)/64.0*sy/unitsPerEm,
		}

		// Add the advance width for the next glyph
		advance := float64(g.XAdvance) / 64.0 * sx / unitsPerEm
		curX += advance

		// Add kerning between characters if this is not the last glyph
		if i < len(output.Glyphs)-1 {
			// Get kerning adjustment between current and next glyph
			kerning, _ := s.GetKerning(runes[i], runes[i+1])
			curX += kerning
		}

		// Add vertical advance
		curY += float64(g.YAdvance) / 64.0 * sy / unitsPerEm
	}

	// Clusters - simplified mapping (one cluster per glyph)
	clusters = make([]TextCluster, len(output.Glyphs))
	for i := range output.Glyphs {
		clusters[i] = TextCluster{
			NumBytes:  1, // Simplified: assume 1 byte per glyph
			NumGlyphs: 1,
		}
	}

	// Cluster flags (simplified)
	clusterFlags = 0

	return glyphs, clusters, clusterFlags, StatusSuccess
}

// toyTextToGlyphsFallback performs a trivial Unicode->glyph mapping similar to
// cairo_scaled_font_text_to_glyphs but without complex shaping.
func (s *scaledFont) toyTextToGlyphsFallback(x, y float64, utf8 string) (glyphs []Glyph, clusters []TextCluster, clusterFlags TextClusterFlags, status Status) {
	// Simple left-to-right mapping: one glyph per rune.
	size := s.toyExtentsFallback().Ascent + s.toyExtentsFallback().Descent
	advancePerRune := size * 0.6

	glyphs = make([]Glyph, 0, len(utf8))
	clusters = make([]TextCluster, 0, len(utf8))

	var curX = x
	// We need byte offsets for clusters.
	for i, r := range utf8 {
		g := Glyph{
			Index: uint64(r),
			X:     curX,
			Y:     y,
		}
		glyphs = append(glyphs, g)

		// Each rune maps to one cluster: num_bytes is number of bytes for this rune.
		var nextByte int
		if i == len(utf8)-1 {
			nextByte = len(utf8)
		} else {
			// This loop body is over runes, but range on string gives byte offsets
			nextByte = i + len(string(r))
		}
		cluster := TextCluster{
			NumBytes:  nextByte - i,
			NumGlyphs: 1,
		}
		clusters = append(clusters, cluster)

		curX += advancePerRune
	}

	clusterFlags = 0
	return glyphs, clusters, clusterFlags, StatusSuccess
}
