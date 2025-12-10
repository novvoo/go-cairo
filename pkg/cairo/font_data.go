package cairo

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-text/typesetting/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"
)

// Font cache to avoid re-parsing fonts
var (
	fontCache     = make(map[string]font.Face)
	fontDataCache = make(map[string][]byte)
	fontCacheMu   sync.RWMutex
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
