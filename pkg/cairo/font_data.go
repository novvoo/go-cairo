package cairo

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-text/typesetting/font"
	"github.com/llgcode/draw2d"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"
)

// Font cache to avoid re-parsing fonts
var (
	fontCache         = make(map[string]font.Face)
	fontDataCache     = make(map[string][]byte)
	fontCacheMu       sync.RWMutex
	draw2dFontCache   = make(map[draw2d.FontData]font.Face)
	draw2dFontCacheMu sync.RWMutex
)

// Internal font data storage
var embeddedFonts = map[string][]byte{
	"Go-Regular":       goregular.TTF,
	"Go-Bold":          gobold.TTF,
	"Go-Italic":        goitalic.TTF,
	"Go-BoldItalic":    gobolditalic.TTF,
	"sans-regular":     goregular.TTF,
	"sans-bold":        gobold.TTF,
	"sans-italic":      goitalic.TTF,
	"sans-bolditalic":  gobolditalic.TTF,
	"serif-regular":    goregular.TTF,
	"serif-bold":       gobold.TTF,
	"serif-italic":     goitalic.TTF,
	"serif-bolditalic": gobolditalic.TTF,
	"mono-regular":     goregular.TTF,
	"mono-bold":        gobold.TTF,
	"mono-italic":      goitalic.TTF,
	"mono-bolditalic":  gobolditalic.TTF,
}

// LoadFontFromFile loads a font from a file path
func LoadFontFromFile(path string) (font.Face, []byte, error) {
	// Check cache first
	fontCacheMu.RLock()
	if face, ok := fontCache[path]; ok {
		data := fontDataCache[path]
		fontCacheMu.RUnlock()
		return face, data, nil
	}
	fontCacheMu.RUnlock()

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	// Parse font
	face, err := font.ParseTTF(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}

	// Cache it
	fontCacheMu.Lock()
	fontCache[path] = face
	fontDataCache[path] = data
	fontCacheMu.Unlock()

	return face, data, nil
}

// LoadEmbeddedFont loads an embedded font by name
func LoadEmbeddedFont(name string) (font.Face, []byte, error) {
	fontCacheMu.RLock()
	if face, ok := fontCache[name]; ok {
		data := fontDataCache[name]
		fontCacheMu.RUnlock()
		return face, data, nil
	}
	fontCacheMu.RUnlock()

	// Try loading from embedded fonts
	data, ok := embeddedFonts[name]
	if !ok {
		// Try loading from assets directory
		assetsPath := filepath.Join("assets", name+".ttf")
		if face, fontData, err := LoadFontFromFile(assetsPath); err == nil {
			return face, fontData, nil
		}
		// Fallback to Go-Regular
		data = goregular.TTF
	}

	face, err := font.ParseTTF(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}

	fontCacheMu.Lock()
	fontCache[name] = face
	fontDataCache[name] = data
	fontCacheMu.Unlock()

	return face, data, nil
}

// GetDefaultFont returns the default embedded font
func GetDefaultFont() (font.Face, []byte) {
	face, data, err := LoadEmbeddedFont("Go-Regular")
	if err != nil {
		// This should never happen as Go-Regular is embedded
		panic("failed to load default font")
	}
	return face, data
}

// GetDejaVuSans returns the DejaVu Sans font
func GetDejaVuSans() (font.Face, []byte) {
	face, data, err := LoadEmbeddedFont("DejaVuSans")
	if err != nil {
		return GetDefaultFont()
	}
	return face, data
}

// RegisterFontWithDraw2D registers a font with the draw2d font system
func RegisterFontWithDraw2D(fontData draw2d.FontData, face font.Face) {
	draw2dFontCacheMu.Lock()
	defer draw2dFontCacheMu.Unlock()
	draw2dFontCache[fontData] = face
}

// GetDraw2DFont retrieves a font from the draw2d cache
func GetDraw2DFont(fontData draw2d.FontData) (font.Face, bool) {
	draw2dFontCacheMu.RLock()
	defer draw2dFontCacheMu.RUnlock()
	face, ok := draw2dFontCache[fontData]
	return face, ok
}

// InitDraw2DFonts initializes the draw2d font system with our fonts
func InitDraw2DFonts() {
	// Register all embedded fonts with draw2d
	fontMappings := []struct {
		name   string
		family draw2d.FontFamily
		style  draw2d.FontStyle
	}{
		{"sans-regular", draw2d.FontFamilySans, draw2d.FontStyleNormal},
		{"sans-bold", draw2d.FontFamilySans, draw2d.FontStyleBold},
		{"sans-italic", draw2d.FontFamilySans, draw2d.FontStyleItalic},
		{"sans-bolditalic", draw2d.FontFamilySans, draw2d.FontStyleBold | draw2d.FontStyleItalic},
		{"serif-regular", draw2d.FontFamilySerif, draw2d.FontStyleNormal},
		{"serif-bold", draw2d.FontFamilySerif, draw2d.FontStyleBold},
		{"serif-italic", draw2d.FontFamilySerif, draw2d.FontStyleItalic},
		{"serif-bolditalic", draw2d.FontFamilySerif, draw2d.FontStyleBold | draw2d.FontStyleItalic},
		{"mono-regular", draw2d.FontFamilyMono, draw2d.FontStyleNormal},
		{"mono-bold", draw2d.FontFamilyMono, draw2d.FontStyleBold},
		{"mono-italic", draw2d.FontFamilyMono, draw2d.FontStyleItalic},
		{"mono-bolditalic", draw2d.FontFamilyMono, draw2d.FontStyleBold | draw2d.FontStyleItalic},
	}

	for _, mapping := range fontMappings {
		face, _, err := LoadEmbeddedFont(mapping.name)
		if err == nil {
			fontData := draw2d.FontData{
				Family: mapping.family,
				Style:  mapping.style,
			}
			RegisterFontWithDraw2D(fontData, face)
		}
	}
}
