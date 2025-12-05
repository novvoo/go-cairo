package cairo

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// UserFontFace implements a custom font face using user-provided data.
type UserFontFace interface {
	FontFace
	// Add other user font specific methods here, e.g.,
	// SetInitFunc, SetRenderGlyphFunc, etc.
}

// userFontFace implements the UserFontFace interface.
type userFontFace struct {
	baseFontFace
	
	// User-defined functions (placeholders)
	initFunc func(face FontFace) Status
	renderGlyphFunc func(scaledFont ScaledFont, glyphID uint64, context Context) Status
}

// NewUserFontFace creates a new user font face.
func NewUserFontFace() UserFontFace {
	face := &userFontFace{
		baseFontFace: baseFontFace{
			refCount: 1,
			status: StatusSuccess,
			fontType: FontTypeUser,
			userData: make(map[*UserDataKey]interface{}),
		},
	}
	
	runtime.SetFinalizer(face, (*userFontFace).Destroy)
	return face
}

func (f *userFontFace) getFontFace() FontFace {
	return f
}

// SetInitFunc sets the initialization function for the user font face.
func (f *userFontFace) SetInitFunc(initFunc func(face FontFace) Status) {
	f.initFunc = initFunc
}

// SetRenderGlyphFunc sets the function to render a single glyph.
func (f *userFontFace) SetRenderGlyphFunc(renderGlyphFunc func(scaledFont ScaledFont, glyphID uint64, context Context) Status) {
	f.renderGlyphFunc = renderGlyphFunc
}

// The ScaledFont implementation needs to be updated to call these functions.
// This is a complex task and requires significant changes to the font rendering pipeline.
// For now, this file defines the surface structure.
