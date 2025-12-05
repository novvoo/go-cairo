package cairo

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"sync/atomic"
	"unsafe"
)

// imageSurface implements image-based surfaces
type imageSurface struct {
	baseSurface
	
	// Image data  
	data   []byte
	width  int
	height int
	stride int
	format Format
	
	// Go image for interoperability
	goImage *image.NRGBA
}

// baseSurface provides common surface functionality
type baseSurface struct {
	refCount    int32
	status      Status
	surfaceType SurfaceType
	content     Content
	
	// Device properties
	device Device
	
	// User data
	userData map[*UserDataKey]interface{}
	
	// Font options
	fontOptions *FontOptions
	
	// Transform properties
	deviceTransform        Matrix
	deviceTransformInverse Matrix
	deviceOffsetX          float64
	deviceOffsetY          float64
	deviceScaleX           float64
	deviceScaleY           float64
	
	// Fallback resolution
	fallbackResolutionX float64
	fallbackResolutionY float64
	
	// Surface state
	finished bool
	
	// Snapshots
	snapshotOf Surface
	snapshots  []Surface
}

// NewImageSurface creates a new image surface
func NewImageSurface(format Format, width, height int) Surface {
	if width <= 0 || height <= 0 {
		return newSurfaceInError(StatusInvalidSize)
	}
	
	stride := formatStrideForWidth(format, width)
	if stride < 0 {
		return newSurfaceInError(StatusInvalidStride)
	}
	
	data := make([]byte, stride*height)
	
	surface := &imageSurface{
		baseSurface: baseSurface{
			refCount:    1,
			status:      StatusSuccess,
			surfaceType: SurfaceTypeImage,
			content:     formatToContent(format),
			userData:    make(map[*UserDataKey]interface{}),
			fontOptions: &FontOptions{},
			deviceScaleX: 1.0,
			deviceScaleY: 1.0,
			fallbackResolutionX: 72.0,
			fallbackResolutionY: 72.0,
		},
		data:   data,
		width:  width,
		height: height,  
		stride: stride,
		format: format,
	}
	
	// Initialize transforms
	surface.deviceTransform.InitIdentity()
	surface.deviceTransformInverse.InitIdentity()
	
	// Create Go image for interoperability
	surface.createGoImage()
	
	return surface
}

// NewImageSurfaceForData creates a surface using existing data
func NewImageSurfaceForData(data []byte, format Format, width, height, stride int) Surface {
	if width <= 0 || height <= 0 {
		return newSurfaceInError(StatusInvalidSize)
	}
	
	if stride < formatStrideForWidth(format, width) {
		return newSurfaceInError(StatusInvalidStride)
	}
	
	if len(data) < stride*height {
		return newSurfaceInError(StatusInvalidSize)
	}
	
	surface := &imageSurface{
		baseSurface: baseSurface{
			refCount:    1,
			status:      StatusSuccess,
			surfaceType: SurfaceTypeImage,
			content:     formatToContent(format),
			userData:    make(map[*UserDataKey]interface{}),
			fontOptions: &FontOptions{},
			deviceScaleX: 1.0,
			deviceScaleY: 1.0,
			fallbackResolutionX: 72.0,
			fallbackResolutionY: 72.0,
		},
		data:   data,
		width:  width,
		height: height,
		stride: stride,
		format: format,
	}
	
	surface.deviceTransform.InitIdentity()
	surface.deviceTransformInverse.InitIdentity()
	surface.createGoImage()
	
	return surface
}

func newSurfaceInError(status Status) Surface {
	surface := &imageSurface{
		baseSurface: baseSurface{
			refCount: 1,
			status:   status,
			userData: make(map[*UserDataKey]interface{}),
		},
	}
	return surface
}

// Helper functions

func formatStrideForWidth(format Format, width int) int {
	switch format {
	case FormatARGB32, FormatRGB24:
		return width * 4
	case FormatA8:
		return width
	case FormatA1:
		return (width + 31) / 32 * 4  // Round up to 32-bit boundary
	case FormatRGB16565:
		return width * 2
	case FormatRGB30:
		return width * 4
	case FormatRGB96F:
		return width * 12  // 3 * 4 bytes per pixel
	case FormatRGBA128F:
		return width * 16  // 4 * 4 bytes per pixel
	default:
		return -1
	}
}

func formatToContent(format Format) Content {
	switch format {
	case FormatARGB32, FormatRGBA128F:
		return ContentColorAlpha
	case FormatRGB24, FormatRGB16565, FormatRGB30, FormatRGB96F:
		return ContentColor
	case FormatA8, FormatA1:
		return ContentAlpha
	default:
		return ContentColorAlpha
	}
}

func (s *imageSurface) createGoImage() {
	if s.format != FormatARGB32 {
		return // Only support ARGB32 for now
	}
	
	s.goImage = &image.NRGBA{
		Pix:    s.data,
		Stride: s.stride,
		Rect:   image.Rect(0, 0, s.width, s.height),
	}
}

// baseSurface implementation

func (s *baseSurface) Reference() Surface {
	atomic.AddInt32(&s.refCount, 1)
	return s.getSurface()
}

func (s *baseSurface) getSurface() Surface {
	// This will be overridden in concrete types
	return nil
}

func (s *baseSurface) Destroy() {
	if atomic.AddInt32(&s.refCount, -1) == 0 {
		s.cleanup()
	}
}

func (s *baseSurface) cleanup() {
	// Base cleanup - overridden in concrete types
	if s.device != nil {
		s.device.Destroy()
	}
}

func (s *baseSurface) GetReferenceCount() int {
	return int(atomic.LoadInt32(&s.refCount))
}

func (s *baseSurface) Status() Status {
	return s.status
}

func (s *baseSurface) GetType() SurfaceType {
	return s.surfaceType
}

func (s *baseSurface) GetContent() Content {
	return s.content
}

func (s *baseSurface) GetDevice() Device {
	return s.device
}

func (s *baseSurface) SetUserData(key *UserDataKey, userData unsafe.Pointer, destroy DestroyFunc) Status {
	if s.status != StatusSuccess {
		return s.status
	}
	
	s.userData[key] = userData
	// TODO: Store destroy function and call it when appropriate
	return StatusSuccess
}

func (s *baseSurface) GetUserData(key *UserDataKey) unsafe.Pointer {
	if data, exists := s.userData[key]; exists {
		return data.(unsafe.Pointer)
	}
	return nil
}

func (s *baseSurface) Flush() {
	// Default implementation does nothing
}

func (s *baseSurface) MarkDirty() {
	// Default implementation does nothing
}

func (s *baseSurface) MarkDirtyRectangle(x, y, width, height int) {
	// Default implementation does nothing
}

func (s *baseSurface) GetFontOptions() *FontOptions {
	return s.fontOptions
}

func (s *baseSurface) Finish() {
	if s.finished {
		return
	}
	s.finished = true
	
	// Clean up snapshots
	for _, snapshot := range s.snapshots {
		snapshot.Destroy()
	}
	s.snapshots = nil
}

func (s *baseSurface) CreateSimilar(content Content, width, height int) Surface {
	// Default implementation creates an image surface
	var format Format
	switch content {
	case ContentColor:
		format = FormatRGB24
	case ContentAlpha:
		format = FormatA8
	case ContentColorAlpha:
		format = FormatARGB32
	default:
		return newSurfaceInError(StatusInvalidContent)
	}
	
	return NewImageSurface(format, width, height)
}

func (s *baseSurface) CreateSimilarImage(format Format, width, height int) Surface {
	return NewImageSurface(format, width, height)
}

func (s *baseSurface) CreateForRectangle(x, y, width, height float64) Surface {
	// TODO: Implement subsurface creation
	return s.CreateSimilar(s.content, int(width), int(height))
}

func (s *baseSurface) SetDeviceScale(xScale, yScale float64) {
	s.deviceScaleX = xScale
	s.deviceScaleY = yScale
	
	// Update transform matrices
	s.deviceTransform.InitScale(xScale, yScale)
	s.deviceTransformInverse.InitScale(1.0/xScale, 1.0/yScale)
}

func (s *baseSurface) GetDeviceScale() (xScale, yScale float64) {
	return s.deviceScaleX, s.deviceScaleY
}

func (s *baseSurface) SetDeviceOffset(xOffset, yOffset float64) {
	s.deviceOffsetX = xOffset
	s.deviceOffsetY = yOffset
	
	// Update transform matrices  
	s.deviceTransform.InitTranslate(xOffset, yOffset)
	s.deviceTransformInverse.InitTranslate(-xOffset, -yOffset)
}

func (s *baseSurface) GetDeviceOffset() (xOffset, yOffset float64) {
	return s.deviceOffsetX, s.deviceOffsetY
}

func (s *baseSurface) SetFallbackResolution(xPixelsPerInch, yPixelsPerInch float64) {
	s.fallbackResolutionX = xPixelsPerInch
	s.fallbackResolutionY = yPixelsPerInch
}

func (s *baseSurface) GetFallbackResolution() (xPixelsPerInch, yPixelsPerInch float64) {
	return s.fallbackResolutionX, s.fallbackResolutionY
}

func (s *baseSurface) CopyPage() {
	// Default implementation does nothing (only meaningful for paginated surfaces)
}

func (s *baseSurface) ShowPage() {
	// Default implementation does nothing (only meaningful for paginated surfaces)
}

// imageSurface specific implementation

func (s *imageSurface) getSurface() Surface {
	return s
}

func (s *imageSurface) Reference() Surface {
	atomic.AddInt32(&s.refCount, 1)
	return s
}

// Image surface specific methods

func (s *imageSurface) GetData() []byte {
	return s.data
}

func (s *imageSurface) GetWidth() int {
	return s.width
}

func (s *imageSurface) GetHeight() int {
	return s.height
}

func (s *imageSurface) GetStride() int {
	return s.stride
}

func (s *imageSurface) GetFormat() Format {
	return s.format
}

func (s *imageSurface) GetGoImage() image.Image {
	return s.goImage
}

// WriteToPNG writes the surface to a PNG file
func (s *imageSurface) WriteToPNG(filename string) Status {
	if s.status != StatusSuccess {
		return s.status
	}
	
	if s.goImage == nil {
		return StatusSurfaceTypeMismatch
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return StatusWriteError
	}
	defer file.Close()
	
	// Convert NRGBA to RGBA if needed
	var img image.Image = s.goImage
	if s.format == FormatARGB32 {
		// Convert from premultiplied ARGB to non-premultiplied RGBA
		img = s.convertToRGBA()
	}
	
	err = png.Encode(file, img)
	if err != nil {
		return StatusWriteError
	}
	
	return StatusSuccess
}

func (s *imageSurface) convertToRGBA() *image.RGBA {
	bounds := s.goImage.Bounds()
	rgba := image.NewRGBA(bounds)
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := s.goImage.At(x, y)
			rgba.Set(x, y, c)
		}
	}
	
	return rgba
}

// Format utilities

func FormatStrideForWidth(format Format, width int) int {
	return formatStrideForWidth(format, width)
}

// LoadPNGSurface creates an image surface from a PNG file
func LoadPNGSurface(filename string) (Surface, error) {
	file, err := os.Open(filename)
	if err != nil {
		return newSurfaceInError(StatusFileNotFound), err
	}
	defer file.Close()
	
	img, err := png.Decode(file)
	if err != nil {
		return newSurfaceInError(StatusReadError), err
	}
	
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	
	surface := NewImageSurface(FormatARGB32, width, height).(*imageSurface)
	
	// Copy image data
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.NRGBAModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.NRGBA)
			surface.goImage.SetNRGBA(x, y, c)
		}
	}
	
	return surface, nil
}

// Surface-specific interfaces for type assertions

type ImageSurface interface {
	Surface
	GetData() []byte
	GetWidth() int
	GetHeight() int
	GetStride() int
	GetFormat() Format
	GetGoImage() image.Image
	WriteToPNG(filename string) Status
}