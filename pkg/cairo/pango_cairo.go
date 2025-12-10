package cairo

import (
	"fmt"
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

// PangoCairoFontMap represents a Pango font map integrated with Cairo
type PangoCairoFontMap struct {
	refCount int32
	status   Status
	userData map[*UserDataKey]interface{}
}

// PangoCairoFont represents a Pango font integrated with Cairo
type PangoCairoFont struct {
	baseFontFace
	family   string
	slant    FontSlant
	weight   FontWeight
	realFace font.Face
	fontData []byte
}

// PangoCairoFontMetrics represents font metrics in PangoCairo
type PangoCairoFontMetrics struct {
	refCount       int32
	status         Status
	ascent         float64
	descent        float64
	height         float64
	lineGap        float64
	underlinePos   float64
	underlineThick float64
	// strikethroughPos and strikethroughThick are reserved for future use
	_ float64 // strikethroughPos
	_ float64 // strikethroughThick
}

// PangoCairoLayout represents a Pango layout for text arrangement
type PangoCairoLayout struct {
	refCount int32
	status   Status
	context  *PangoCairoContext
	text     string
	fontDesc *PangoFontDescription
	// attributes is reserved for future attribute support
	_           []PangoAttribute // attributes
	width       int
	height      int
	wrap        PangoWrapMode
	ellipsize   PangoEllipsizeMode
	align       PangoAlignment
	spacing     float64
	lineSpacing float64
	userData    map[*UserDataKey]interface{}
}

// PangoCairoContext represents a Pango context integrated with Cairo
type PangoCairoContext struct {
	refCount int32
	status   Status
	fontMap  *PangoCairoFontMap
	// fontDescription is stored but accessed via getter/setter methods
	fontDescription *PangoFontDescription
	baseDir         PangoDirection
	userData        map[*UserDataKey]interface{}
}

// PangoFontDescription describes a font in Pango
type PangoFontDescription struct {
	family  string
	style   PangoStyle
	variant PangoVariant
	weight  PangoWeight
	stretch PangoStretch
	size    float64
}

// PangoAttribute represents text attributes in Pango
type PangoAttribute struct {
	// These fields are reserved for future attribute support
	_ int           // startIndex
	_ int           // endIndex
	_ PangoAttrType // attrType
	_ interface{}   // value
}

// Enumerations for PangoCairo

type PangoDirection int
type PangoStyle int
type PangoVariant int
type PangoWeight int
type PangoStretch int
type PangoWrapMode int
type PangoEllipsizeMode int
type PangoAlignment int
type PangoAttrType int

const (
	PangoDirectionLTR PangoDirection = iota
	PangoDirectionRTL
	PangoDirectionTTB
	PangoDirectionBTT
)

const (
	PangoStyleNormal PangoStyle = iota
	PangoStyleOblique
	PangoStyleItalic
)

const (
	PangoVariantNormal PangoVariant = iota
	PangoVariantSmallCaps
)

const (
	PangoWeightThin PangoWeight = 100 + iota*100
	PangoWeightUltraLight
	PangoWeightLight
	PangoWeightSemiLight
	PangoWeightBook
	PangoWeightNormal
	PangoWeightMedium
	PangoWeightSemiBold
	PangoWeightBold
	PangoWeightUltraBold
	PangoWeightHeavy
	PangoWeightUltraHeavy
)

const (
	PangoStretchUltraCondensed PangoStretch = iota
	PangoStretchExtraCondensed
	PangoStretchCondensed
	PangoStretchSemiCondensed
	PangoStretchNormal
	PangoStretchSemiExpanded
	PangoStretchExpanded
	PangoStretchExtraExpanded
	PangoStretchUltraExpanded
)

const (
	PangoWrapWord PangoWrapMode = iota
	PangoWrapChar
	PangoWrapWordChar
)

const (
	PangoEllipsizeNone PangoEllipsizeMode = iota
	PangoEllipsizeStart
	PangoEllipsizeMiddle
	PangoEllipsizeEnd
)

const (
	PangoAlignLeft PangoAlignment = iota
	PangoAlignCenter
	PangoAlignRight
)

const (
	PangoAttrInvalid PangoAttrType = iota
	PangoAttrLanguage
	PangoAttrFamily
	PangoAttrStyle
	PangoAttrWeight
	PangoAttrVariant
	PangoAttrStretch
	PangoAttrSize
	PangoAttrFontDesc
	PangoAttrForeground
	PangoAttrBackground
	PangoAttrUnderline
	PangoAttrStrikethrough
	PangoAttrRise
	PangoAttrShape
	PangoAttrScale
	PangoAttrFallback
	PangoAttrLetterSpacing
	PangoAttrFontFeatures
	PangoAttrForegroundAlpha
	PangoAttrBackgroundAlpha
	PangoAttrAllowBreaks
	PangoAttrShow
	PangoAttrInsertHyphens
	PangoAttrOverline
)

// PangoCairoScaledFont represents a scaled font in PangoCairo
type PangoCairoScaledFont struct {
	refCount    int32
	status      Status
	fontType    FontType
	fontFace    FontFace
	fontMatrix  Matrix
	ctm         Matrix
	scaleMatrix Matrix
	options     *FontOptions
	pangoFont   *PangoCairoFont
}

// NewPangoCairoFontMap creates a new Pango font map integrated with Cairo
func NewPangoCairoFontMap() *PangoCairoFontMap {
	return &PangoCairoFontMap{
		refCount: 1,
		status:   StatusSuccess,
		userData: make(map[*UserDataKey]interface{}),
	}
}

// Reference management for PangoCairoFontMap
func (fm *PangoCairoFontMap) Reference() *PangoCairoFontMap {
	atomic.AddInt32(&fm.refCount, 1)
	return fm
}

func (fm *PangoCairoFontMap) Destroy() {
	if atomic.AddInt32(&fm.refCount, -1) == 0 {
		// Cleanup resources if needed
	}
}

func (fm *PangoCairoFontMap) GetReferenceCount() int {
	return int(atomic.LoadInt32(&fm.refCount))
}

func (fm *PangoCairoFontMap) Status() Status {
	return fm.status
}

// UserData management for PangoCairoFontMap
func (fm *PangoCairoFontMap) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	if fm.status != StatusSuccess {
		return fm.status
	}
	if fm.userData == nil {
		fm.userData = make(map[*UserDataKey]interface{})
	}
	fm.userData[key] = userData
	_ = destroy // destroy func is currently ignored
	return StatusSuccess
}

func (fm *PangoCairoFontMap) GetUserData(key *UserDataKey) unsafe.Pointer {
	if fm.userData == nil {
		return nil
	}
	if data, ok := fm.userData[key]; ok {
		return data.(unsafe.Pointer)
	}
	return nil
}

// NewPangoCairoFont creates a new Pango font integrated with Cairo
func NewPangoCairoFont(family string, slant FontSlant, weight FontWeight) *PangoCairoFont {
	pf := &PangoCairoFont{
		baseFontFace: baseFontFace{
			refCount: 1,
			status:   StatusSuccess,
			fontType: FontTypeUser,
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

	pf.realFace = face
	pf.fontData = data

	if pf.realFace == nil {
		pf.status = StatusFontTypeMismatch
	}
	return pf
}

// FontFace interface implementation for PangoCairoFont
func (f *PangoCairoFont) Reference() FontFace {
	atomic.AddInt32(&f.refCount, 1)
	return f
}

func (f *PangoCairoFont) Destroy() {
	if atomic.AddInt32(&f.refCount, -1) == 0 {
		// nothing to free at the moment
	}
}

func (f *PangoCairoFont) GetReferenceCount() int {
	return int(atomic.LoadInt32(&f.refCount))
}

func (f *PangoCairoFont) Status() Status {
	return f.status
}

func (f *PangoCairoFont) GetType() FontType {
	return f.fontType
}

func (f *PangoCairoFont) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	if f.status != StatusSuccess {
		return f.status
	}
	if f.userData == nil {
		f.userData = make(map[*UserDataKey]interface{})
	}
	f.userData[key] = userData
	_ = destroy // destroy func is currently ignored
	return StatusSuccess
}

func (f *PangoCairoFont) GetUserData(key *UserDataKey) unsafe.Pointer {
	if f.userData == nil {
		return nil
	}
	if data, ok := f.userData[key]; ok {
		return data.(unsafe.Pointer)
	}
	return nil
}

// NewPangoCairoFontMetrics creates new font metrics
func NewPangoCairoFontMetrics(ascent, descent, height, lineGap float64) *PangoCairoFontMetrics {
	return &PangoCairoFontMetrics{
		refCount:       1,
		status:         StatusSuccess,
		ascent:         ascent,
		descent:        descent,
		height:         height,
		lineGap:        lineGap,
		underlinePos:   -descent * 0.5,
		underlineThick: (ascent + descent) * 0.05,
	}
}

// Reference management for PangoCairoFontMetrics
func (fm *PangoCairoFontMetrics) Reference() *PangoCairoFontMetrics {
	atomic.AddInt32(&fm.refCount, 1)
	return fm
}

func (fm *PangoCairoFontMetrics) Destroy() {
	if atomic.AddInt32(&fm.refCount, -1) == 0 {
		// Cleanup resources if needed
	}
}

func (fm *PangoCairoFontMetrics) GetReferenceCount() int {
	return int(atomic.LoadInt32(&fm.refCount))
}

func (fm *PangoCairoFontMetrics) Status() Status {
	return fm.status
}

// Metric getters
func (fm *PangoCairoFontMetrics) GetAscent() float64 {
	return fm.ascent
}

func (fm *PangoCairoFontMetrics) GetDescent() float64 {
	return fm.descent
}

func (fm *PangoCairoFontMetrics) GetHeight() float64 {
	return fm.height
}

func (fm *PangoCairoFontMetrics) GetLineGap() float64 {
	return fm.lineGap
}

func (fm *PangoCairoFontMetrics) GetUnderlinePosition() float64 {
	return fm.underlinePos
}

func (fm *PangoCairoFontMetrics) GetUnderlineThickness() float64 {
	return fm.underlineThick
}

// NewPangoCairoLayout creates a new Pango layout
func NewPangoCairoLayout(context *PangoCairoContext) *PangoCairoLayout {
	return &PangoCairoLayout{
		refCount: 1,
		status:   StatusSuccess,
		context:  context,
		width:    -1, // Unset
		height:   -1, // Unset
		wrap:     PangoWrapWord,
		align:    PangoAlignLeft,
		userData: make(map[*UserDataKey]interface{}),
	}
}

// Reference management for PangoCairoLayout
func (l *PangoCairoLayout) Reference() *PangoCairoLayout {
	atomic.AddInt32(&l.refCount, 1)
	return l
}

func (l *PangoCairoLayout) Destroy() {
	if atomic.AddInt32(&l.refCount, -1) == 0 {
		if l.context != nil {
			l.context.Destroy()
		}
		if l.fontDesc != nil {
			// Destroy font description if needed
		}
	}
}

func (l *PangoCairoLayout) GetReferenceCount() int {
	return int(atomic.LoadInt32(&l.refCount))
}

func (l *PangoCairoLayout) Status() Status {
	return l.status
}

// Layout property setters and getters
func (l *PangoCairoLayout) SetText(text string) {
	l.text = text
}

func (l *PangoCairoLayout) GetText() string {
	return l.text
}

func (l *PangoCairoLayout) SetFontDescription(desc *PangoFontDescription) {
	l.fontDesc = desc
}

func (l *PangoCairoLayout) GetFontDescription() *PangoFontDescription {
	return l.fontDesc
}

func (l *PangoCairoLayout) SetWidth(width int) {
	l.width = width
}

func (l *PangoCairoLayout) GetWidth() int {
	return l.width
}

func (l *PangoCairoLayout) SetHeight(height int) {
	l.height = height
}

func (l *PangoCairoLayout) GetHeight() int {
	return l.height
}

func (l *PangoCairoLayout) SetWrap(wrap PangoWrapMode) {
	l.wrap = wrap
}

func (l *PangoCairoLayout) GetWrap() PangoWrapMode {
	return l.wrap
}

func (l *PangoCairoLayout) SetEllipsize(ellipsize PangoEllipsizeMode) {
	l.ellipsize = ellipsize
}

func (l *PangoCairoLayout) GetEllipsize() PangoEllipsizeMode {
	return l.ellipsize
}

func (l *PangoCairoLayout) SetAlignment(align PangoAlignment) {
	l.align = align
}

func (l *PangoCairoLayout) GetAlignment() PangoAlignment {
	return l.align
}

func (l *PangoCairoLayout) SetSpacing(spacing float64) {
	l.spacing = spacing
}

func (l *PangoCairoLayout) GetSpacing() float64 {
	return l.spacing
}

func (l *PangoCairoLayout) SetLineSpacing(lineSpacing float64) {
	l.lineSpacing = lineSpacing
}

func (l *PangoCairoLayout) GetLineSpacing() float64 {
	return l.lineSpacing
}

// UserData management for PangoCairoLayout
func (l *PangoCairoLayout) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	if l.status != StatusSuccess {
		return l.status
	}
	if l.userData == nil {
		l.userData = make(map[*UserDataKey]interface{})
	}
	l.userData[key] = userData
	_ = destroy // destroy func is currently ignored
	return StatusSuccess
}

func (l *PangoCairoLayout) GetUserData(key *UserDataKey) unsafe.Pointer {
	if l.userData == nil {
		return nil
	}
	if data, ok := l.userData[key]; ok {
		return data.(unsafe.Pointer)
	}
	return nil
}

// NewPangoCairoContext creates a new Pango context integrated with Cairo
func NewPangoCairoContext(fontMap *PangoCairoFontMap) *PangoCairoContext {
	return &PangoCairoContext{
		refCount: 1,
		status:   StatusSuccess,
		fontMap:  fontMap,
		baseDir:  PangoDirectionLTR,
		userData: make(map[*UserDataKey]interface{}),
	}
}

// Reference management for PangoCairoContext
func (c *PangoCairoContext) Reference() *PangoCairoContext {
	atomic.AddInt32(&c.refCount, 1)
	return c
}

func (c *PangoCairoContext) Destroy() {
	if atomic.AddInt32(&c.refCount, -1) == 0 {
		if c.fontMap != nil {
			c.fontMap.Destroy()
		}
	}
}

func (c *PangoCairoContext) GetReferenceCount() int {
	return int(atomic.LoadInt32(&c.refCount))
}

func (c *PangoCairoContext) Status() Status {
	return c.status
}

// Context property setters and getters
func (c *PangoCairoContext) SetFontMap(fontMap *PangoCairoFontMap) {
	if c.fontMap != nil {
		c.fontMap.Destroy()
	}
	c.fontMap = fontMap.Reference()
}

func (c *PangoCairoContext) GetFontMap() *PangoCairoFontMap {
	return c.fontMap.Reference()
}

func (c *PangoCairoContext) SetBaseDir(direction PangoDirection) {
	c.baseDir = direction
}

func (c *PangoCairoContext) GetBaseDir() PangoDirection {
	return c.baseDir
}

// UserData management for PangoCairoContext
func (c *PangoCairoContext) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	if c.status != StatusSuccess {
		return c.status
	}
	if c.userData == nil {
		c.userData = make(map[*UserDataKey]interface{})
	}
	c.userData[key] = userData
	_ = destroy // destroy func is currently ignored
	return StatusSuccess
}

func (c *PangoCairoContext) GetUserData(key *UserDataKey) unsafe.Pointer {
	if c.userData == nil {
		return nil
	}
	if data, ok := c.userData[key]; ok {
		return data.(unsafe.Pointer)
	}
	return nil
}

// NewPangoFontDescription creates a new font description
func NewPangoFontDescription() *PangoFontDescription {
	return &PangoFontDescription{
		family:  "sans",
		style:   PangoStyleNormal,
		variant: PangoVariantNormal,
		weight:  PangoWeightNormal,
		stretch: PangoStretchNormal,
		size:    12.0, // Default size in points
	}
}

// FontDescription property setters and getters
func (fd *PangoFontDescription) SetFamily(family string) {
	fd.family = family
}

func (fd *PangoFontDescription) GetFamily() string {
	return fd.family
}

func (fd *PangoFontDescription) SetStyle(style PangoStyle) {
	fd.style = style
}

func (fd *PangoFontDescription) GetStyle() PangoStyle {
	return fd.style
}

func (fd *PangoFontDescription) SetWeight(weight PangoWeight) {
	fd.weight = weight
}

func (fd *PangoFontDescription) GetWeight() PangoWeight {
	return fd.weight
}

func (fd *PangoFontDescription) SetStretch(stretch PangoStretch) {
	fd.stretch = stretch
}

func (fd *PangoFontDescription) GetStretch() PangoStretch {
	return fd.stretch
}

func (fd *PangoFontDescription) SetSize(size float64) {
	fd.size = size
}

func (fd *PangoFontDescription) GetSize() float64 {
	return fd.size
}

// NewPangoCairoScaledFont creates a new scaled font for PangoCairo
func NewPangoCairoScaledFont(fontFace FontFace, fontMatrix, ctm *Matrix, options *FontOptions) *PangoCairoScaledFont {
	sf := &PangoCairoScaledFont{
		refCount: 1,
		status:   StatusSuccess,
		fontType: FontTypeUser,
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
	// For our implementation we just copy fontMatrix into scaleMatrix.
	sf.scaleMatrix = sf.fontMatrix

	// If the font face is a PangoCairoFont, keep a reference to it
	if pcFont, ok := fontFace.(*PangoCairoFont); ok {
		sf.pangoFont = pcFont
	}

	return sf
}

// ScaledFont interface implementation for PangoCairoScaledFont
func (s *PangoCairoScaledFont) Reference() ScaledFont {
	atomic.AddInt32(&s.refCount, 1)
	return s
}

func (s *PangoCairoScaledFont) Destroy() {
	if atomic.AddInt32(&s.refCount, -1) == 0 {
		if s.fontFace != nil {
			s.fontFace.Destroy()
		}
	}
}

func (s *PangoCairoScaledFont) GetReferenceCount() int {
	return int(atomic.LoadInt32(&s.refCount))
}

func (s *PangoCairoScaledFont) Status() Status {
	return s.status
}

func (s *PangoCairoScaledFont) GetType() FontType {
	return s.fontType
}

func (s *PangoCairoScaledFont) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	// For now we store user data in the associated FontFace to keep things simple.
	if s.fontFace == nil {
		return StatusNullPointer
	}
	return s.fontFace.SetUserData(key, userData, destroy)
}

func (s *PangoCairoScaledFont) GetUserData(key *UserDataKey) unsafe.Pointer {
	if s.fontFace == nil {
		return nil
	}
	return s.fontFace.GetUserData(key)
}

func (s *PangoCairoScaledFont) GetFontFace() FontFace {
	if s.fontFace == nil {
		return nil
	}
	return s.fontFace.Reference()
}

func (s *PangoCairoScaledFont) GetFontMatrix() *Matrix {
	m := s.fontMatrix
	return &m
}

func (s *PangoCairoScaledFont) GetCTM() *Matrix {
	m := s.ctm
	return &m
}

func (s *PangoCairoScaledFont) GetScaleMatrix() *Matrix {
	m := s.scaleMatrix
	return &m
}

func (s *PangoCairoScaledFont) GetFontOptions() *FontOptions {
	if s.options == nil {
		return NewFontOptions()
	}
	return s.options.Copy()
}

// getRealFace returns the underlying font.Face and checks for errors.
func (s *PangoCairoScaledFont) getRealFace() (font.Face, Status) {
	if s.fontFace == nil {
		return nil, StatusNullPointer
	}

	// Try to get as PangoCairoFont first
	if pcFont, ok := s.fontFace.(*PangoCairoFont); ok && pcFont.realFace != nil {
		return pcFont.realFace, StatusSuccess
	}

	// Fall back to toy font
	toy, ok := s.fontFace.(*toyFontFace)
	if !ok || toy.realFace == nil {
		return nil, StatusFontTypeMismatch
	}
	return toy.realFace, StatusSuccess
}

// Extents returns font extents using the real font face.
func (s *PangoCairoScaledFont) Extents() *FontExtents {
	fe := &FontExtents{}

	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		// Fallback to toy extents if real face is not available
		return s.toyExtentsFallback()
	}

	// Get font metrics from go-text/typesetting
	// Ascent, Descent, Height in FUnits
	metrics, _ := realFace.FontHExtents()
	ascentFUnits := float64(metrics.Ascender)
	descentFUnits := float64(metrics.Descender)
	lineGapFUnits := float64(metrics.LineGap)

	// Convert to user space units
	fe.Ascent = ascentFUnits / 64.0
	fe.Descent = -descentFUnits / 64.0 // Descent is negative in FUnits, cairo expects positive
	fe.Height = fe.Ascent + fe.Descent + lineGapFUnits/64.0
	fe.LineGap = lineGapFUnits / 64.0

	// Max advance is a guess without shaping a string
	fe.MaxXAdvance = fe.Ascent + fe.Descent
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
func (s *PangoCairoScaledFont) toyExtentsFallback() *FontExtents {
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
func (s *PangoCairoScaledFont) TextExtents(utf8 string) *TextExtents {
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

	// Calculate total advance and bounds
	var totalAdvance fixed.Int26_6
	var curX float64 // Current X position for glyph placement
	var minX, minY, maxX, maxY float64
	firstGlyph := true

	for _, g := range output.Glyphs {
		// Get glyph outline for bounds calculation
		glyphData := realFace.GlyphData(api.GID(g.GlyphID))
		if outline, ok := glyphData.(api.GlyphOutline); ok {
			// Convert outline points to user space - harfbuzz already provides user space coordinates
			for _, seg := range outline.Segments {
				for _, arg := range seg.Args {
					x := float64(arg.X) / 64.0
					y := float64(arg.Y) / 64.0

					// Add glyph position (current X + offset)
					x += curX + float64(g.XOffset)/64.0
					y -= float64(g.YOffset) / 64.0 // Subtract because glyph offsets are in font coordinate system

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

		// Advance to next glyph position
		curX += float64(g.XAdvance) / 64.0
		totalAdvance += g.XAdvance
	}

	// Convert to user space units
	ext.XAdvance = float64(totalAdvance) / 64.0
	ext.YAdvance = 0

	// Set proper width and height based on actual bounds
	ext.Width = maxX - minX
	ext.Height = maxY - minY
	ext.XBearing = minX
	ext.YBearing = -maxY // Negative because Y axis is inverted in Cairo

	return ext
}

// toyTextExtentsFallback computes naive text extents assuming fixed advance width.
func (s *PangoCairoScaledFont) toyTextExtentsFallback(utf8 string) *TextExtents {
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
func (s *PangoCairoScaledFont) GlyphExtents(glyphs []Glyph) *TextExtents {
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
func (s *PangoCairoScaledFont) GlyphPath(glyphID uint64) (*Path, error) {
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

	// Check if we need to flip the Y axis based on the font matrix
	// Font glyphs are designed for Y growing upward, but our coordinate system has Y growing downward.
	// Since we now use positive Y scale in font matrix, we always need to flip.
	flipY := true

	// Get font units per em and scale factor for coordinate transformation
	unitsPerEm := float64(realFace.Upem())
	scaleX := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	scaleY := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	if scaleX == 0 {
		scaleX = 1.0
	}
	if scaleY == 0 {
		scaleY = 1.0
	}

	// Iterate over the path segments
	// Note: The outline coordinates from go-text/typesetting are in font units (float32)
	// We need to scale them to user space and preserve the segment types
	for _, seg := range outline.Segments {
		var pd PathData

		switch seg.Op {
		case api.SegmentOpMoveTo:
			// Convert from font units to user space
			x := (float64(seg.Args[0].X) / unitsPerEm) * scaleX
			y := (float64(seg.Args[0].Y) / unitsPerEm) * scaleY
			// Apply Y flip if needed
			if flipY {
				y = -y
			}
			pd.Type = PathMoveTo
			pd.Points = []Point{{X: x, Y: y}}

		case api.SegmentOpLineTo:
			x := (float64(seg.Args[0].X) / unitsPerEm) * scaleX
			y := (float64(seg.Args[0].Y) / unitsPerEm) * scaleY
			// Apply Y flip if needed
			if flipY {
				y = -y
			}
			pd.Type = PathLineTo
			pd.Points = []Point{{X: x, Y: y}}

		case api.SegmentOpQuadTo:
			// Convert quadratic Bezier to cubic Bezier
			// For a quadratic curve with control point Q and end point P2,
			// the cubic equivalent has control points:
			// C1 = current_point + 2/3 * (Q - current_point)
			// C2 = P2 + 2/3 * (Q - P2)
			// However, since we don't track current point here, we'll use a simpler conversion
			x1 := (float64(seg.Args[0].X) / unitsPerEm) * scaleX
			y1 := (float64(seg.Args[0].Y) / unitsPerEm) * scaleY
			x2 := (float64(seg.Args[1].X) / unitsPerEm) * scaleX
			y2 := (float64(seg.Args[1].Y) / unitsPerEm) * scaleY
			// Apply Y flip if needed
			if flipY {
				y1 = -y1
				y2 = -y2
			}
			// Simplified: use the control point twice for cubic conversion
			pd.Type = PathCurveTo
			pd.Points = []Point{
				{X: x1, Y: y1},
				{X: x1, Y: y1},
				{X: x2, Y: y2},
			}

		case api.SegmentOpCubeTo:
			x1 := (float64(seg.Args[0].X) / unitsPerEm) * scaleX
			y1 := (float64(seg.Args[0].Y) / unitsPerEm) * scaleY
			x2 := (float64(seg.Args[1].X) / unitsPerEm) * scaleX
			y2 := (float64(seg.Args[1].Y) / unitsPerEm) * scaleY
			x3 := (float64(seg.Args[2].X) / unitsPerEm) * scaleX
			y3 := (float64(seg.Args[2].Y) / unitsPerEm) * scaleY
			// Apply Y flip if needed
			if flipY {
				y1 = -y1
				y2 = -y2
				y3 = -y3
			}
			pd.Type = PathCurveTo
			pd.Points = []Point{
				{X: x1, Y: y1},
				{X: x2, Y: y2},
				{X: x3, Y: y3},
			}
		}

		cairoPath.Data = append(cairoPath.Data, pd)
	}

	return cairoPath, nil
}

// GetTextBearingMetrics returns the bearing metrics for a text string
func (s *PangoCairoScaledFont) GetTextBearingMetrics(text string) (xBearing, yBearing float64, status Status) {
	metrics := s.TextExtents(text)
	if metrics == nil {
		return 0, 0, StatusFontTypeMismatch
	}
	return metrics.XBearing, metrics.YBearing, StatusSuccess
}

// GetTextAlignmentOffset calculates the Y offset for text alignment
func (s *PangoCairoScaledFont) GetTextAlignmentOffset(alignment TextAlignment) (float64, Status) {
	fontExtents := s.Extents()
	if fontExtents == nil {
		return 0, StatusFontTypeMismatch
	}
	return GetAlignmentOffset(alignment, fontExtents), StatusSuccess
}

// GetKerning returns the kerning adjustment between two runes
func (s *PangoCairoScaledFont) GetKerning(r1, r2 rune) (float64, Status) {
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
func (s *PangoCairoScaledFont) applyHinting(points []Point) []Point {
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
func (s *PangoCairoScaledFont) GetGlyphBearingMetrics(r rune) (xBearing, yBearing float64, status Status) {
	metrics, status := s.GetGlyphMetrics(r)
	if status != StatusSuccess {
		return 0, 0, status
	}
	return metrics.XBearing, metrics.YBearing, StatusSuccess
}

// GetGlyphMetrics returns detailed metrics for a specific glyph
func (s *PangoCairoScaledFont) GetGlyphMetrics(r rune) (*GlyphMetrics, Status) {
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

	// Get font units per em and scale factor first
	unitsPerEm := float64(realFace.Upem())
	scaleX := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	scaleY := math.Hypot(s.fontMatrix.XY, s.fontMatrix.YY)
	if scaleX == 0 {
		scaleX = 1.0
	}
	if scaleY == 0 {
		scaleY = 1.0
	}

	// Calculate bounding box from outline
	// Note: Outline coordinates from go-text/typesetting are in font units (float32)
	var xmin, xmax, ymin, ymax float64
	firstPoint := true

	pointCount := 0
	for _, seg := range outline.Segments {
		for _, arg := range seg.Args {
			// Coordinates are already in font units (float32), just convert to float64
			xInFontUnits := float64(arg.X)
			yInFontUnits := float64(arg.Y)

			pointCount++

			if firstPoint {
				xmin, xmax = xInFontUnits, xInFontUnits
				ymin, ymax = yInFontUnits, yInFontUnits
				firstPoint = false
			} else {
				if xInFontUnits < xmin {
					xmin = xInFontUnits
				}
				if xInFontUnits > xmax {
					xmax = xInFontUnits
				}
				if yInFontUnits < ymin {
					ymin = yInFontUnits
				}
				if yInFontUnits > ymax {
					ymax = yInFontUnits
				}
			}
		}
	}

	// Scale bounding box to user space
	xmin = (xmin / unitsPerEm) * scaleX
	xmax = (xmax / unitsPerEm) * scaleX
	ymin = (ymin / unitsPerEm) * scaleY
	ymax = (ymax / unitsPerEm) * scaleY

	// Get horizontal metrics from the font's hmtx table
	// HorizontalAdvance returns the advance width in font units (not 26.6 format)
	rawAdvance := realFace.HorizontalAdvance(gid)

	// Convert from font units to user space units
	// Formula: (font_units / units_per_em) * font_size
	advanceInFontUnits := float64(rawAdvance)
	advanceWidth := (advanceInFontUnits / unitsPerEm) * scaleX

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
func (s *PangoCairoScaledFont) GetGlyphs(utf8 string) (glyphs []Glyph, status Status) {
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

	return glyphs, StatusSuccess
}

// TextToGlyphs performs text shaping to get accurate glyphs and clusters.
func (s *PangoCairoScaledFont) TextToGlyphs(x, y float64, utf8 string) (glyphs []Glyph, clusters []TextCluster, clusterFlags TextClusterFlags, status Status) {
	realFace, status := s.getRealFace()
	if status != StatusSuccess {
		return s.toyTextToGlyphsFallback(x, y, utf8)
	}

	// Get the font size from the font matrix
	// The font size is typically the YY component of the font matrix
	fontSize := math.Hypot(s.fontMatrix.XX, s.fontMatrix.YX)
	if fontSize == 0 {
		fontSize = 12.0 // Default fallback
	}

	// 1. Shape the text with the correct font size
	// fixed.I() converts an integer to 26.6 fixed point format
	runes := []rune(utf8)
	input := shaping.Input{
		Text:      runes,
		RunStart:  0,
		RunEnd:    len(runes),
		Direction: di.DirectionLTR,
		Face:      realFace,
		Size:      fixed.I(int(fontSize)), // Convert to 26.6 fixed point
	}
	output := (&shaping.HarfbuzzShaper{}).Shape(input)

	// 2. Convert shaped output to cairo's Glyph and TextCluster structures
	glyphs = make([]Glyph, len(output.Glyphs))
	var curX, curY float64

	// Process each glyph with proper spacing
	for i, g := range output.Glyphs {
		// Position is in user space, relative to the start point (x, y)
		glyphs[i] = Glyph{
			Index: uint64(g.GlyphID),
			X:     x + curX + float64(g.XOffset)/64.0,
			Y:     y + curY - float64(g.YOffset)/64.0, // Subtract because glyph offsets are in font coordinate system
		}

		// Add the advance width for the next glyph
		// The shaper returns advances in 26.6 fixed point format
		curX += float64(g.XAdvance) / 64.0
		curY += float64(g.YAdvance) / 64.0
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
func (s *PangoCairoScaledFont) toyTextToGlyphsFallback(x, y float64, utf8 string) (glyphs []Glyph, clusters []TextCluster, clusterFlags TextClusterFlags, status Status) {
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

// PangoCairoShowText renders text using PangoCairo directly to the surface
func PangoCairoShowText(ctx Context, layout *PangoCairoLayout) {
	if ctx.Status() != StatusSuccess {
		return
	}

	// Get current point or use (0, 0)
	x, y := ctx.GetCurrentPoint()
	if x == 0 && y == 0 && ctx.HasCurrentPoint() == False {
		x, y = 0, 0
	}

	// Create scaled font from layout's font description
	if layout.fontDesc == nil {
		ctx.(*context).status = StatusFontTypeMismatch
		return
	}

	fontFace := NewPangoCairoFont(layout.fontDesc.family, FontSlantNormal, FontWeightNormal)
	defer fontFace.Destroy()

	fontMatrix := NewMatrix()
	// Use positive Y scale - our coordinate system has Y growing downward,
	// and we'll handle the glyph flip in the rendering code
	fontMatrix.InitScale(layout.fontDesc.size, layout.fontDesc.size)

	ctm := NewMatrix()
	ctm.InitIdentity()

	sf := NewPangoCairoScaledFont(fontFace, fontMatrix, ctm, nil)
	defer sf.Destroy()

	// Perform text shaping to get glyphs
	glyphs, _, _, status := sf.TextToGlyphs(x, y, layout.GetText())
	if status != StatusSuccess {
		ctx.(*context).status = status
		return
	}

	// Render glyphs directly to surface using PangoCairo
	c := ctx.(*context)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get the current source pattern for text color
	source := c.gstate.source
	if source == nil {
		return
	}

	// Render each glyph directly to the surface
	for _, glyph := range glyphs {
		// Get the glyph path
		glyphPath, err := sf.GlyphPath(glyph.Index)
		if err != nil || glyphPath == nil {
			continue
		}

		// Clear current path and create a new one for this glyph
		c.NewPath()

		// Translate the glyph path to the correct position and add to context
		// The glyph path is in font space, we need to translate it to the glyph position
		for _, pathData := range glyphPath.Data {
			switch pathData.Type {
			case PathMoveTo:
				if len(pathData.Points) > 0 {
					c.MoveTo(pathData.Points[0].X+glyph.X, pathData.Points[0].Y+glyph.Y)
				}
			case PathLineTo:
				if len(pathData.Points) > 0 {
					c.LineTo(pathData.Points[0].X+glyph.X, pathData.Points[0].Y+glyph.Y)
				}
			case PathCurveTo:
				if len(pathData.Points) >= 3 {
					c.CurveTo(
						pathData.Points[0].X+glyph.X, pathData.Points[0].Y+glyph.Y,
						pathData.Points[1].X+glyph.X, pathData.Points[1].Y+glyph.Y,
						pathData.Points[2].X+glyph.X, pathData.Points[2].Y+glyph.Y,
					)
				}
			case PathClosePath:
				c.ClosePath()
			}
		}

		// Fill the glyph
		c.Fill()
	}

	// Update current point to the position after the last glyph
	if len(glyphs) > 0 {
		extents := sf.TextExtents(layout.GetText())
		c.currentPoint.x = x + extents.XAdvance
		c.currentPoint.y = y + extents.YAdvance
		c.currentPoint.hasPoint = true
	}
}

// PangoCairoUpdateLayout updates a layout to match the current transformation matrix of a Cairo context
func PangoCairoUpdateLayout(ctx Context, layout *PangoCairoLayout) {
	// Implementation would synchronize the layout with the Cairo context transformation
	// For now, this is a placeholder
	_ = ctx
	_ = layout
}

// PangoCairoCreateLayout creates a new Pango layout for a Cairo context
func PangoCairoCreateLayout(ctx Context) *PangoCairoLayout {
	// Create a default font map and context
	fontMap := NewPangoCairoFontMap()
	pangoCtx := NewPangoCairoContext(fontMap)
	layout := NewPangoCairoLayout(pangoCtx)
	return layout
}

// GlyphCornerCoordinates represents the four corners of a glyph's bounding box
type GlyphCornerCoordinates struct {
	TopLeftX, TopLeftY         float64
	TopRightX, TopRightY       float64
	BottomLeftX, BottomLeftY   float64
	BottomRightX, BottomRightY float64
}

// GetGlyphCornerCoordinates calculates the four corner coordinates of a glyph
func (s *PangoCairoScaledFont) GetGlyphCornerCoordinates(glyph Glyph) (*GlyphCornerCoordinates, Status) {
	// Get glyph metrics
	metrics, status := s.GetGlyphMetrics(rune(glyph.Index))
	if status != StatusSuccess {
		return nil, status
	}

	// Calculate the four corners based on glyph position and advance width
	// The bounding box represents the visual bounds of the glyph
	topRightX := glyph.X + metrics.BoundingBox.XMax

	if glyph.Index == uint64('H') {
		fmt.Printf("[DEBUG GetGlyphCornerCoordinates] 'H': glyph.X=%.2f, BBox.XMax=%.2f, TopRightX=%.2f\n",
			glyph.X, metrics.BoundingBox.XMax, topRightX)
	}

	coords := &GlyphCornerCoordinates{
		TopLeftX:     glyph.X + metrics.BoundingBox.XMin,
		TopLeftY:     glyph.Y + metrics.BoundingBox.YMin,
		TopRightX:    topRightX,
		TopRightY:    glyph.Y + metrics.BoundingBox.YMin,
		BottomLeftX:  glyph.X + metrics.BoundingBox.XMin,
		BottomLeftY:  glyph.Y + metrics.BoundingBox.YMax,
		BottomRightX: topRightX,
		BottomRightY: glyph.Y + metrics.BoundingBox.YMax,
	}

	return coords, StatusSuccess
}

// CheckGlyphCollision checks if two glyphs' bounding boxes overlap
// char1 and char2 are the actual characters (runes) corresponding to the glyphs
func (s *PangoCairoScaledFont) CheckGlyphCollision(glyph1, glyph2 Glyph, char1, char2 rune) (bool, Status) {
	// Get metrics for both characters
	metrics1, status := s.GetGlyphMetrics(char1)
	if status != StatusSuccess {
		return false, status
	}

	metrics2, status := s.GetGlyphMetrics(char2)
	if status != StatusSuccess {
		return false, status
	}

	// Calculate bounding boxes in absolute coordinates
	box1MinX := glyph1.X + metrics1.BoundingBox.XMin
	box1MaxX := glyph1.X + metrics1.BoundingBox.XMax
	box1MinY := glyph1.Y + metrics1.BoundingBox.YMin
	box1MaxY := glyph1.Y + metrics1.BoundingBox.YMax

	box2MinX := glyph2.X + metrics2.BoundingBox.XMin
	box2MaxX := glyph2.X + metrics2.BoundingBox.XMax
	box2MinY := glyph2.Y + metrics2.BoundingBox.YMin
	box2MaxY := glyph2.Y + metrics2.BoundingBox.YMax

	// Check for overlap
	// Two rectangles overlap if:
	// 1. The left edge of rect1 is to the left of the right edge of rect2
	// 2. The right edge of rect1 is to the right of the left edge of rect2
	// 3. The top edge of rect1 is above the bottom edge of rect2
	// 4. The bottom edge of rect1 is below the top edge of rect2
	overlap := box1MinX < box2MaxX &&
		box1MaxX > box2MinX &&
		box1MinY < box2MaxY &&
		box1MaxY > box2MinY

	return overlap, StatusSuccess
}

// PrintGlyphInfo prints detailed information about a glyph including its corner coordinates
func (s *PangoCairoScaledFont) PrintGlyphInfo(glyph Glyph, char rune) {
	// Get metrics using the correct character, not the glyph index
	metrics, status := s.GetGlyphMetrics(char)
	if status != StatusSuccess {
		fmt.Printf("无法获取字符 '%c' 的度量信息: %v\n", char, status)
		return
	}

	// Calculate corners manually using the correct metrics
	visualWidth := metrics.BoundingBox.XMax - metrics.BoundingBox.XMin

	coords := &GlyphCornerCoordinates{
		TopLeftX:     glyph.X + metrics.BoundingBox.XMin,
		TopLeftY:     glyph.Y + metrics.BoundingBox.YMin,
		TopRightX:    glyph.X + metrics.BoundingBox.XMax,
		TopRightY:    glyph.Y + metrics.BoundingBox.YMin,
		BottomLeftX:  glyph.X + metrics.BoundingBox.XMin,
		BottomLeftY:  glyph.Y + metrics.BoundingBox.YMax,
		BottomRightX: glyph.X + metrics.BoundingBox.XMax,
		BottomRightY: glyph.Y + metrics.BoundingBox.YMax,
	}

	fmt.Printf("字符 '%c' 位置信息:\n", char)
	fmt.Printf("  位置: (%.2f, %.2f)\n", glyph.X, glyph.Y)
	fmt.Printf("  边界框: minX=%.2f, minY=%.2f, maxX=%.2f, maxY=%.2f\n",
		metrics.BoundingBox.XMin, metrics.BoundingBox.YMin,
		metrics.BoundingBox.XMax, metrics.BoundingBox.YMax)
	fmt.Printf("  视觉宽度: %.2f, Advance: %.2f\n", visualWidth, metrics.XAdvance)
	fmt.Printf("  左上角: (%.2f, %.2f)\n", coords.TopLeftX, coords.TopLeftY)
	fmt.Printf("  右上角: (%.2f, %.2f)\n", coords.TopRightX, coords.TopRightY)
	fmt.Printf("  左下角: (%.2f, %.2f)\n", coords.BottomLeftX, coords.BottomLeftY)
	fmt.Printf("  右下角: (%.2f, %.2f)\n", coords.BottomRightX, coords.BottomRightY)
	fmt.Println()
}

// PrintTextGlyphsInfo prints information for all glyphs in a text string
func (s *PangoCairoScaledFont) PrintTextGlyphsInfo(utf8 string, glyphs []Glyph) {
	runes := []rune(utf8)

	// Print info for each glyph
	for i, glyph := range glyphs {
		var char rune
		if i < len(runes) {
			char = runes[i]
		} else {
			char = rune(glyph.Index)
		}

		s.PrintGlyphInfo(glyph, char)

		// Check for collisions with subsequent glyphs
		for j := i + 1; j < len(glyphs); j++ {
			var nextChar rune
			if j < len(runes) {
				nextChar = runes[j]
			} else {
				nextChar = rune(glyphs[j].Index)
			}
			collides, status := s.CheckGlyphCollision(glyph, glyphs[j], char, nextChar)
			if status == StatusSuccess && collides {
				fmt.Printf("警告: 字符 '%c' 和 '%c' 之间存在重叠!\n\n", char, nextChar)
			}
		}
	}
}

// PangoRectangle represents a rectangle in Pango coordinates
type PangoRectangle struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// GetPixelExtents returns the pixel extents of the layout
func (l *PangoCairoLayout) GetPixelExtents() *PangoRectangle {
	if l.text == "" || l.fontDesc == nil {
		return &PangoRectangle{}
	}

	// Create a temporary scaled font to get text extents
	fontFace := NewPangoCairoFont(l.fontDesc.family, FontSlantNormal, FontWeightNormal)
	defer fontFace.Destroy()

	fontMatrix := NewMatrix()
	// Use positive Y scale - our coordinate system has Y growing downward, and we'll handle the glyph flip in the rendering code
	fontMatrix.InitScale(l.fontDesc.size, l.fontDesc.size)

	ctm := NewMatrix()
	ctm.InitIdentity()

	scaledFont := NewPangoCairoScaledFont(fontFace, fontMatrix, ctm, nil)
	defer scaledFont.Destroy()

	extents := scaledFont.TextExtents(l.text)

	return &PangoRectangle{
		X:      extents.XBearing,
		Y:      extents.YBearing,
		Width:  extents.Width,
		Height: extents.Height,
	}
}

// GetFontExtents returns the font extents for the layout
func (l *PangoCairoLayout) GetFontExtents() *FontExtents {
	if l.fontDesc == nil {
		return &FontExtents{}
	}

	// Create a temporary scaled font to get font extents
	fontFace := NewPangoCairoFont(l.fontDesc.family, FontSlantNormal, FontWeightNormal)
	defer fontFace.Destroy()

	fontMatrix := NewMatrix()
	// Use positive Y scale - our coordinate system has Y growing downward, and we'll handle the glyph flip in the rendering code
	fontMatrix.InitScale(l.fontDesc.size, l.fontDesc.size)

	ctm := NewMatrix()
	ctm.InitIdentity()

	scaledFont := NewPangoCairoScaledFont(fontFace, fontMatrix, ctm, nil)
	defer scaledFont.Destroy()

	return scaledFont.Extents()
}
