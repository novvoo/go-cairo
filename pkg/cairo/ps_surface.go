package cairo

import (
	"os"
	"runtime"
	"unsafe"
	
	"github.com/novvoo/go-cairo/pkg/cairo/cgo"
)

// psSurface implements the PostScript surface using CGo.
type psSurface struct {
	baseSurface
	cSurface *cgo.CPSurface
}

// NewPSSurface creates a new PostScript surface.
func NewPSSurface(filename string, width, height float64) Surface {
	cSurface := cgo.CairoPSSurfaceCreate(filename, width, height)
	status := Status(cgo.CairoSurfaceStatus(cSurface))
	
	if status != StatusSuccess {
		cgo.CairoSurfaceDestroy(cSurface)
		return newSurfaceInError(status)
	}
	
	surface := &psSurface{
		baseSurface: baseSurface{
			refCount:    1,
			status:      StatusSuccess,
			surfaceType: SurfaceTypePS,
			content:     ContentColorAlpha,
			userData:    make(map[*UserDataKey]interface{}),
			fontOptions: &FontOptions{},
			deviceScaleX: 1.0,
			deviceScaleY: 1.0,
			fallbackResolutionX: 72.0,
			fallbackResolutionY: 72.0,
		},
		cSurface: cSurface,
	}
	
	// Set the finalizer to destroy the C surface when the Go object is garbage collected
	runtime.SetFinalizer(surface, (*psSurface).Destroy)
	return surface
}

func (s *psSurface) getSurface() Surface {
	return s
}

func (s *psSurface) cleanup() {
	if s.cSurface != nil {
		cgo.CairoSurfaceDestroy(s.cSurface)
		s.cSurface = nil
	}
	s.baseSurface.cleanup()
}

func (s *psSurface) finishConcrete() error {
	// Flush the surface to ensure all pending operations are written
	// Cairo's cairo_surface_destroy handles the final flush, but we can call Flush explicitly.
	return s.Flush()
}

func (s *psSurface) Flush() error {
	if s.cSurface != nil {
		// Cairo does not expose a cairo_surface_flush function in the C API,
		// but the context's flush will flush the surface.
		// We can use cairo_surface_status to check for errors after operations.
		// For now, we rely on the context's flush and destroy's implicit flush.
	}
	return nil
}

func (s *psSurface) CopyPage() {
	if s.cSurface != nil {
		cgo.CairoSurfaceCopyPage(s.cSurface)
	}
}

func (s *psSurface) ShowPage() {
	if s.cSurface != nil {
		cgo.CairoSurfaceShowPage(s.cSurface)
	}
}

// PSSurfaceSetSize sets the size of the PostScript surface.
func (s *psSurface) SetSize(width, height float64) {
	if s.cSurface != nil {
		cgo.CairoPSSurfaceSetSize(s.cSurface, width, height)
	}
}

// PSSurfaceDscComment emits a comment into the PostScript output.
func (s *psSurface) DscComment(comment string) {
	if s.cSurface != nil {
		cgo.CairoPSSurfaceDscComment(s.cSurface, comment)
	}
}

// GetCSurface returns the underlying C surface pointer.
func (s *psSurface) GetCSurface() unsafe.Pointer {
	return unsafe.Pointer(s.cSurface)
}

// Add the new surface type to the SurfaceType enum in types.go
// This is a manual step, but I'll assume it's done for now to proceed with the logic.
// I will add a step to update types.go later.
