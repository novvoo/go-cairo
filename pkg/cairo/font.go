package cairo

import (
	"math"
	"sync/atomic"
	"unsafe"
)

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

// toyFontSize derives an approximate font size from the font matrix.
func (s *scaledFont) toyFontSize() float64 {
	// Use average of xx and yy scale as size; fall back to 12 if zero.
	sx := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	sy := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	size := (sx + sy) * 0.5
	if size == 0 {
		size = 12
	}
	return size
}

// Extents returns toy font extents based on the derived font size.
func (s *scaledFont) Extents() *FontExtents {
	size := s.toyFontSize()
	fe := &FontExtents{}
	fe.Ascent = size * 0.8
	fe.Descent = size * 0.2
	fe.Height = fe.Ascent + fe.Descent
	fe.MaxXAdvance = size
	fe.MaxYAdvance = 0
	return fe
}

// TextExtents computes naive text extents assuming fixed advance width.
func (s *scaledFont) TextExtents(utf8 string) *TextExtents {
	size := s.toyFontSize()
	advancePerRune := size * 0.6

	var runeCount int
	for range utf8 {
		runeCount++
	}

	ext := &TextExtents{}
	ext.Width = float64(runeCount) * advancePerRune
	ext.Height = s.Extents().Height
	ext.XAdvance = ext.Width
	ext.YAdvance = 0
	ext.XBearing = 0
	ext.YBearing = -s.Extents().Ascent
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

// TextToGlyphs performs a trivial Unicode->glyph mapping similar to
// cairo_scaled_font_text_to_glyphs but without complex shaping.
func (s *scaledFont) TextToGlyphs(x, y float64, utf8 string) (glyphs []Glyph, clusters []TextCluster, clusterFlags TextClusterFlags, status Status) {
	// Simple left-to-right mapping: one glyph per rune.
	size := s.toyFontSize()
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
