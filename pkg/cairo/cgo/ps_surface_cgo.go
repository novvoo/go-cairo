package cgo

/*
#cgo pkg-config: cairo cairo-ps
#include <cairo.h>
#include <stdlib.h>
#include <string.h>

// Helper function to create a PS surface for a filename
cairo_surface_t* cairo_ps_surface_create_wrapper(const char *filename, double width_in_points, double height_in_points) {
    return cairo_ps_surface_create(filename, width_in_points, height_in_points);
}

// Helper function to set the size of a PS surface
void cairo_ps_surface_set_size_wrapper(cairo_surface_t *surface, double width_in_points, double height_in_points) {
    cairo_ps_surface_set_size(surface, width_in_points, height_in_points);
}

// Helper function to add a DSC comment
void cairo_ps_surface_dsc_comment_wrapper(cairo_surface_t *surface, const char *comment) {
    cairo_ps_surface_dsc_comment(surface, comment);
}

// Helper function to get the status of a surface
cairo_status_t cairo_surface_status_wrapper(cairo_surface_t *surface) {
    return cairo_surface_status(surface);
}

// Helper function to reference a surface
cairo_surface_t* cairo_surface_reference_wrapper(cairo_surface_t *surface) {
    return cairo_surface_reference(surface);
}

// Helper function to destroy a surface
void cairo_surface_destroy_wrapper(cairo_surface_t *surface) {
    cairo_surface_destroy(surface);
}

// Helper function for show_page
void cairo_surface_show_page_wrapper(cairo_surface_t *surface) {
    cairo_surface_show_page(surface);
}

// Helper function for copy_page
void cairo_surface_copy_page_wrapper(cairo_surface_t *surface) {
    cairo_surface_copy_page(surface);
}

*/
import "C"
import (
	"unsafe"
)

// CPSurface is a wrapper for cairo_surface_t for PostScript surfaces.
type CPSurface C.cairo_surface_t

// CairoPSSurfaceCreate creates a new PostScript surface.
func CairoPSSurfaceCreate(filename string, widthInPoints, heightInPoints float64) *CPSurface {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	
	surface := C.cairo_ps_surface_create_wrapper(cFilename, C.double(widthInPoints), C.double(heightInPoints))
	return (*CPSurface)(surface)
}

// CairoPSSurfaceSetSize sets the size of a PostScript surface.
func CairoPSSurfaceSetSize(surface *CPSurface, widthInPoints, heightInPoints float64) {
	C.cairo_ps_surface_set_size_wrapper((*C.cairo_surface_t)(surface), C.double(widthInPoints), C.double(heightInPoints))
}

// CairoPSSurfaceDscComment emits a comment into the PostScript output.
func CairoPSSurfaceDscComment(surface *CPSurface, comment string) {
	cComment := C.CString(comment)
	defer C.free(unsafe.Pointer(cComment))
	C.cairo_ps_surface_dsc_comment_wrapper((*C.cairo_surface_t)(surface), cComment)
}

// CairoSurfaceStatus gets the status of a surface.
func CairoSurfaceStatus(surface *CPSurface) C.cairo_status_t {
	return C.cairo_surface_status_wrapper((*C.cairo_surface_t)(surface))
}

// CairoSurfaceReference references a surface.
func CairoSurfaceReference(surface *CPSurface) *CPSurface {
	return (*CPSurface)(C.cairo_surface_reference_wrapper((*C.cairo_surface_t)(surface)))
}

// CairoSurfaceDestroy destroys a surface.
func CairoSurfaceDestroy(surface *CPSurface) {
	C.cairo_surface_destroy_wrapper((*C.cairo_surface_t)(surface))
}

// CairoSurfaceShowPage emits the current page and starts a new one.
func CairoSurfaceShowPage(surface *CPSurface) {
	C.cairo_surface_show_page_wrapper((*C.cairo_surface_t)(surface))
}

// CairoSurfaceCopyPage copies the current page to the surface.
func CairoSurfaceCopyPage(surface *CPSurface) {
	C.cairo_surface_copy_page_wrapper((*C.cairo_surface_t)(surface))
}
