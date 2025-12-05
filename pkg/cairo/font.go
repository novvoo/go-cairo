package cairo

import (
	"math"
	"sync/atomic"
		"unsafe"
		"strings"
	
	"github.com/go-text/typesetting/font"
		"github.com/go-text/typesetting/shaping"
		"golang.org/x/image/font/gofont/goregular"
		"golang.org/x/image/font/gofont/gobold"
		"golang.org/x/image/font/gofont/goitalic"
		"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/math/fixed"
)

	var defaultFont font.Face
	
	func init() {
		// Load Go Regular as a fallback
		f, err := font.ParseTTF(goregular.TTF)
		if err == nil {
			defaultFont = f
		}
	}

	// fontLookupTable is a simple map for toy font face lookup
	var fontLookupTable = map[string]map[FontSlant]map[FontWeight]font.Face{
		"sans": {
			FontSlantNormal: {
				FontWeightNormal: loadGoFont(goregular.TTF),
				FontWeightBold:   loadGoFont(gobold.TTF),
			},
			FontSlantItalic: {
				FontWeightNormal: loadGoFont(goitalic.TTF),
				FontWeightBold:   loadGoFont(gobolditalic.TTF),
			},
		},
		"serif": {
			FontSlantNormal: {
				FontWeightNormal: loadGoFont(goregular.TTF),
				FontWeightBold:   loadGoFont(gobold.TTF),
			},
			FontSlantItalic: {
				FontWeightNormal: loadGoFont(goitalic.TTF),
				FontWeightBold:   loadGoFont(gobolditalic.TTF),
			},
		},
		"monospace": {
			FontSlantNormal: {
				FontWeightNormal: loadGoFont(goregular.TTF),
				FontWeightBold:   loadGoFont(gobold.TTF),
			},
			FontSlantItalic: {
				FontWeightNormal: loadGoFont(goitalic.TTF),
				FontWeightBold:   loadGoFont(gobolditalic.TTF),
			},
		},
	}

	func loadGoFont(ttf []byte) font.Face {
		f, err := font.ParseTTF(ttf)
		if err != nil {
			return nil
		}
		return f
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

		// Simple font lookup based on family, slant, and weight
		familyKey := strings.ToLower(ff.family)
		if familyKey == "sans-serif" || familyKey == "sans" {
			familyKey = "sans"
		} else if familyKey == "serif" {
			familyKey = "serif"
		} else if familyKey == "monospace" {
			familyKey = "monospace"
		} else {
			familyKey = "sans" // Fallback to sans
		}

		if slants, ok := fontLookupTable[familyKey]; ok {
			if weights, ok := slants[ff.slant]; ok {
				if face, ok := weights[ff.weight]; ok && face != nil {
					ff.realFace = face
				}
			}
		}

		// Final fallback
		if ff.realFace == nil {
			ff.realFace = defaultFont
		}

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
	unitsPerEm := float64(realFace.UnitsPerEm())
	
	// Ascent, Descent, Height in FUnits
	ascentFUnits := float64(realFace.Ascender())
	descentFUnits := float64(realFace.Descender())
	lineGapFUnits := float64(realFace.LineGap())
	
	// Convert to user space units
	fe.Ascent = ascentFUnits * sx / unitsPerEm
	fe.Descent = -descentFUnits * sy / unitsPerEm // Descent is negative in FUnits, cairo expects positive
	fe.Height = fe.Ascent + fe.Descent + lineGapFUnits * sy / unitsPerEm
	
	// Max advance is a guess without shaping a string
	fe.MaxXAdvance = sx
	fe.MaxYAdvance = 0
	
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
	fe.MaxXAdvance = size
	fe.MaxYAdvance = 0
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
	shaper := shaping.NewShaper(realFace)
	output := shaper.Shape(utf8)
	
	// 2. Calculate extents from shaped output
	// We need to convert fixed.Int26_6 to float64 (divide by 64)
	
	// Bounding box in FUnits
	bbox := output.Bounds()
	
	// Scale factor from font matrix
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	unitsPerEm := float64(realFace.UnitsPerEm())
	
	// Convert FUnits to user space units
	funitToUser := func(f fixed.Int26_6, scale float64) float64 {
		return float64(f.Round()) * scale / unitsPerEm
	}
	
	// Extents
	ext.XBearing = funitToUser(bbox.Min.X, sx)
	ext.YBearing = funitToUser(bbox.Min.Y, sy)
	ext.Width = funitToUser(bbox.Dx(), sx)
	ext.Height = funitToUser(bbox.Dy(), sy)
	
	// Advance
	lastGlyph := output.Glyphs[len(output.Glyphs)-1]
	ext.XAdvance = float64(lastGlyph.XAdvance) * sx / unitsPerEm
	ext.YAdvance = float64(lastGlyph.YAdvance) * sy / unitsPerEm
	
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

	// Load the glyph from the font face
	glyph, err := realFace.LoadGlyph(font.GID(glyphID))
	if err != nil {
		return nil, newError(StatusFontTypeMismatch, err.Error())
	}

	// Convert the font.Path to cairo.Path
	cairoPath := &Path{
		Status: StatusSuccess,
		Data:   make([]PathData, 0),
	}

	// The font.Path is a sequence of draw commands (MoveTo, LineTo, CurveTo, ClosePath)
	// We need to convert these to cairo's PathData structure.
	// The path coordinates are in FUnits. We need to scale them by the font matrix.

	// Scale factor from font matrix
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	unitsPerEm := float64(realFace.UnitsPerEm())

	// FUnits to user space: FUnits * (scale / unitsPerEm)
	funitToUser := func(f float64, scale float64) float64 {
		return f * scale / unitsPerEm
	}

	// Iterate over the path segments
	var currentX, currentY float64
	for _, seg := range glyph.Path {
		switch seg.Op {
		case font.MoveTo:
			p := seg.Points[0]
			currentX = funitToUser(float64(p.X), sx)
			currentY = funitToUser(float64(p.Y), sy)
			cairoPath.Data = append(cairoPath.Data, PathData{
				Type: PathMoveTo,
				Points: []Point{{X: currentX, Y: currentY}},
			})
		case font.LineTo:
			p := seg.Points[0]
			currentX = funitToUser(float64(p.X), sx)
			currentY = funitToUser(float64(p.Y), sy)
			cairoPath.Data = append(cairoPath.Data, PathData{
				Type: PathLineTo,
				Points: []Point{{X: currentX, Y: currentY}},
			})
		case font.CurveTo:
			p1 := seg.Points[0]
			p2 := seg.Points[1]
			p3 := seg.Points[2]
			currentX = funitToUser(float64(p3.X), sx)
			currentY = funitToUser(float64(p3.Y), sy)
			cairoPath.Data = append(cairoPath.Data, PathData{
				Type: PathCurveTo,
				Points: []Point{
					{X: funitToUser(float64(p1.X), sx), Y: funitToUser(float64(p1.Y), sy)},
					{X: funitToUser(float64(p2.X), sx), Y: funitToUser(float64(p2.Y), sy)},
					{X: currentX, Y: currentY},
				},
			})
		case font.ClosePath:
			cairoPath.Data = append(cairoPath.Data, PathData{
				Type: PathClosePath,
				Points: []Point{},
			})
		}
	}

	return cairoPath, nil
}

// TextToGlyphs performs text shaping to get accurate glyphs and clusters.
func (s *scaledFont) TextToGlyphs(x, y float64, utf8 string) (glyphs []Glyph, clusters []TextCluster, clusterFlags TextClusterFlags, status Status) {
	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		return s.toyTextToGlyphsFallback(x, y, utf8)
	}
	
	// 1. Shape the text
	shaper := shaping.NewShaper(realFace)
	output := shaper.Shape(utf8)
	
	// 2. Convert shaped output to cairo's Glyph and TextCluster structures
	
	// Scale factor from font matrix
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	unitsPerEm := float64(realFace.UnitsPerEm())
	
	// FUnits to user space: FUnits * (scale / unitsPerEm)
	funitToUser := func(f float64, scale float64) float64 {
		return f * scale / unitsPerEm
	}
	
	// Glyphs
	glyphs = make([]Glyph, len(output.Glyphs))
	for i, g := range output.Glyphs {
		// Position is in user space, relative to the start point (x, y)
		glyphs[i] = Glyph{
			Index: uint64(g.ID),
			X:     x + funitToUser(float64(g.XOffset), sx) + funitToUser(float64(g.X), sx),
			Y:     y - funitToUser(float64(g.YOffset), sy) - funitToUser(float64(g.Y), sy), // Y is inverted in cairo
		}
	}
	
	// Clusters
	clusters = make([]TextCluster, len(output.Clusters))
	for i, c := range output.Clusters {
		clusters[i] = TextCluster{
			NumBytes:  int(c.NumBytes),
			NumGlyphs: int(c.NumGlyphs),
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
