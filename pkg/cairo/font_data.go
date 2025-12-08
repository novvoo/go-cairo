package cairo

import (
	"bytes"
	_ "embed"
	"sync"

	"github.com/go-text/typesetting/font"
)

// Embedded font files
// Note: go:embed paths are relative to the source file
// Since we're in pkg/cairo/, we need to reference ../../assets/
// However, go:embed doesn't support ../ paths, so we'll load fonts at runtime instead

// Font cache to avoid re-parsing fonts
var (
	fontCache       = make(map[string]font.Face)
	fontCacheMu     sync.RWMutex
	defaultFontData = goRegularData
)

// LoadEmbeddedFont loads an embedded font by name
func LoadEmbeddedFont(name string) (font.Face, error) {
	fontCacheMu.RLock()
	if face, ok := fontCache[name]; ok {
		fontCacheMu.RUnlock()
		return face, nil
	}
	fontCacheMu.RUnlock()

	var data []byte
	switch name {
	case "DejaVuSans":
		data = dejaVuSansData
	case "Go-Regular":
		data = goRegularData
	default:
		data = defaultFontData
	}

	face, err := font.ParseTTF(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	fontCacheMu.Lock()
	fontCache[name] = face
	fontCacheMu.Unlock()

	return face, nil
}

// GetDefaultFont returns the default embedded font
func GetDefaultFont() font.Face {
	face, err := LoadEmbeddedFont("Go-Regular")
	if err != nil {
		return nil
	}
	return face
}

// GetDejaVuSans returns the DejaVu Sans font
func GetDejaVuSans() font.Face {
	face, err := LoadEmbeddedFont("DejaVuSans")
	if err != nil {
		return GetDefaultFont()
	}
	return face
}
