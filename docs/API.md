# Go-Cairo API Documentation

This document provides comprehensive documentation for the Go-Cairo library, a pure Go implementation of the Cairo 2D graphics library.

## Overview

Go-Cairo provides the following core components:

- **Context**: The main drawing context (`cairo_t` equivalent)
- **Surface**: Drawing targets like image surfaces (`cairo_surface_t` equivalent)  
- **Pattern**: Paint sources including colors and gradients (`cairo_pattern_t` equivalent)
- **Matrix**: 2D affine transformation matrices (`cairo_matrix_t` equivalent)

## Core Types

### Status

Error status enumeration:

```go
type Status int

const (
    StatusSuccess Status = iota
    StatusNoMemory
    StatusInvalidRestore
    StatusInvalidPopGroup
    // ... more status codes
)
```

### Format

Pixel formats for image surfaces:

```go
type Format int

const (
    FormatInvalid   Format = -1
    FormatARGB32    Format = 0  // 32-bit ARGB  
    FormatRGB24     Format = 1  // 24-bit RGB
    FormatA8        Format = 2  // 8-bit alpha
    FormatA1        Format = 3  // 1-bit alpha
    // ... more formats
)
```

### Matrix

2D affine transformation matrix:

```go
type Matrix struct {
    XX, YX float64  // x-axis transformation
    XY, YY float64  // y-axis transformation  
    X0, Y0 float64  // translation
}

// Methods
func NewMatrix() *Matrix
func (m *Matrix) InitIdentity()
func (m *Matrix) InitTranslate(tx, ty float64)
func (m *Matrix) InitScale(sx, sy float64)  
func (m *Matrix) InitRotate(radians float64)
```

## Surface Interface

### Creating Surfaces

```go
// Create image surface
surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 200)
defer surface.Destroy()

// Create surface from existing data
data := make([]byte, stride*height)
surface := cairo.NewImageSurfaceForData(data, format, width, height, stride)

// Load PNG image
surface, err := cairo.LoadPNGSurface("image.png")
```

### Surface Methods

```go
type Surface interface {
    // Reference management
    Reference() Surface
    Destroy()
    GetReferenceCount() int
    
    // Properties  
    Status() Status
    GetType() SurfaceType
    GetContent() Content
    
    // Operations
    Flush()
    MarkDirty()
    Finish()
    
    // Transformations
    SetDeviceScale(xScale, yScale float64)
    GetDeviceScale() (float64, float64)
    SetDeviceOffset(xOffset, yOffset float64)
    GetDeviceOffset() (float64, float64)
    
    // Similar surface creation
    CreateSimilar(content Content, width, height int) Surface
    CreateSimilarImage(format Format, width, height int) Surface
}
```

### ImageSurface Specific

```go
type ImageSurface interface {
    Surface
    
    // Image data access
    GetData() []byte
    GetWidth() int
    GetHeight() int
    GetStride() int
    GetFormat() Format
    GetGoImage() image.Image
    
    // Save to PNG
    WriteToPNG(filename string) Status
}
```

## Context Interface

### Creating Contexts

```go
surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 200)
ctx := cairo.NewContext(surface)
defer ctx.Destroy()
```

### State Management

```go
// Save/restore graphics state
ctx.Save()
// ... modify state
ctx.Restore()

// Group operations
ctx.PushGroup()
// ... draw into group
pattern := ctx.PopGroup()
```

### Drawing Properties

```go
// Set drawing operator
ctx.SetOperator(cairo.OperatorOver)

// Set source color/pattern
ctx.SetSourceRGB(1.0, 0.0, 0.0)  // Red
ctx.SetSourceRGBA(1.0, 0.0, 0.0, 0.5)  // Semi-transparent red
ctx.SetSource(pattern)

// Line properties
ctx.SetLineWidth(2.0)
ctx.SetLineCap(cairo.LineCapRound)  
ctx.SetLineJoin(cairo.LineJoinRound)
ctx.SetDash([]float64{5, 5}, 0)  // Dashed line

// Fill properties
ctx.SetFillRule(cairo.FillRuleWinding)
ctx.SetAntialias(cairo.AntialiasGood)
```

### Transformations

```go
// Basic transformations
ctx.Translate(10, 20)
ctx.Scale(2.0, 2.0)  
ctx.Rotate(math.Pi / 4)  // 45 degrees

// Matrix operations
matrix := cairo.NewMatrix()
matrix.InitScale(2.0, 2.0)
ctx.Transform(matrix)
ctx.SetMatrix(matrix)

// Coordinate conversion
deviceX, deviceY := ctx.UserToDevice(userX, userY)
userX, userY = ctx.DeviceToUser(deviceX, deviceY)
```

### Path Operations

```go
// Path creation
ctx.NewPath()
ctx.MoveTo(10, 10)
ctx.LineTo(100, 10) 
ctx.LineTo(100, 100)
ctx.LineTo(10, 100)
ctx.ClosePath()

// Geometric shapes
ctx.Rectangle(10, 10, 80, 80)
ctx.Arc(50, 50, 30, 0, 2*math.Pi)  // Circle

// Bezier curves
ctx.CurveTo(x1, y1, x2, y2, x3, y3)

// Relative operations  
ctx.RelMoveTo(dx, dy)
ctx.RelLineTo(dx, dy)
```

### Drawing Operations

```go
// Fill and stroke
ctx.Fill()          // Fill path and clear it
ctx.FillPreserve()  // Fill path and keep it
ctx.Stroke()        // Stroke path and clear it 
ctx.StrokePreserve() // Stroke path and keep it

// Paint entire surface
ctx.Paint()
ctx.PaintWithAlpha(0.5)

// Masking
ctx.Mask(maskPattern)
ctx.MaskSurface(maskSurface, x, y)
```

### Clipping

```go
// Set clipping region from current path
ctx.Clip()
ctx.ClipPreserve()

// Reset clipping
ctx.ResetClip()

// Query clipping
inClip := ctx.InClip(x, y)
x1, y1, x2, y2 := ctx.ClipExtents()
```

## Pattern Interface

### Creating Patterns

```go
// Solid colors
pattern := cairo.NewPatternRGB(1.0, 0.0, 0.0)     // Red
pattern := cairo.NewPatternRGBA(1.0, 0.0, 0.0, 0.5) // Semi-transparent red

// Surface patterns
surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
pattern := cairo.NewPatternForSurface(surface)

// Linear gradients
gradient := cairo.NewPatternLinear(0, 0, 100, 0)
if grad, ok := gradient.(cairo.LinearGradientPattern); ok {
    grad.AddColorStopRGB(0.0, 1.0, 0.0, 0.0)    // Red at start
    grad.AddColorStopRGB(1.0, 0.0, 0.0, 1.0)    // Blue at end
}

// Radial gradients  
radial := cairo.NewPatternRadial(50, 50, 10, 50, 50, 50)
if grad, ok := radial.(cairo.RadialGradientPattern); ok {
    grad.AddColorStopRGB(0.0, 1.0, 1.0, 1.0)    // White at center
    grad.AddColorStopRGB(1.0, 0.0, 0.0, 0.0)    // Black at edge
}
```

### Pattern Properties

```go
// Transformation matrix
matrix := cairo.NewMatrix()
matrix.InitScale(2.0, 2.0)
pattern.SetMatrix(matrix)

// Extend mode (for gradients and surface patterns)
pattern.SetExtend(cairo.ExtendRepeat)
pattern.SetExtend(cairo.ExtendReflect)  
pattern.SetExtend(cairo.ExtendPad)

// Filter mode
pattern.SetFilter(cairo.FilterBest)
```

## Usage Examples

### Basic Drawing

```go
surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 200)
defer surface.Destroy()

ctx := cairo.NewContext(surface)
defer ctx.Destroy()

// White background
ctx.SetSourceRGB(1, 1, 1)
ctx.Paint()

// Red circle
ctx.SetSourceRGB(1, 0, 0)
ctx.Arc(100, 100, 50, 0, 2*math.Pi)
ctx.Fill()

// Save to PNG
if img, ok := surface.(cairo.ImageSurface); ok {
    img.WriteToPNG("output.png")
}
```

### Transformations

```go
ctx.Save()
ctx.Translate(100, 100)  // Move origin to center
ctx.Rotate(math.Pi / 4)  // Rotate 45 degrees
ctx.Scale(2, 1)          // Stretch horizontally

// Draw transformed shape
ctx.Rectangle(-25, -10, 50, 20)
ctx.Fill()

ctx.Restore()  // Restore original coordinate system
```

### Gradients

```go
// Create linear gradient
gradient := cairo.NewPatternLinear(0, 0, 200, 0)
if grad, ok := gradient.(cairo.LinearGradientPattern); ok {
    grad.AddColorStopRGB(0.0, 1, 0, 0)    // Red
    grad.AddColorStopRGB(0.5, 0, 1, 0)    // Green  
    grad.AddColorStopRGB(1.0, 0, 0, 1)    // Blue
}

ctx.SetSource(gradient)
ctx.Rectangle(10, 10, 180, 50)
ctx.Fill()

gradient.Destroy()
```

## Error Handling

All operations that can fail return a Status code or store it in the object:

```go
ctx := cairo.NewContext(surface)
if ctx.Status() != cairo.StatusSuccess {
    log.Fatalf("Failed to create context: %v", ctx.Status())
}

status := surface.WriteToPNG("output.png")
if status != cairo.StatusSuccess {
    log.Fatalf("Failed to write PNG: %v", status)  
}
```

## Memory Management

Go-Cairo uses reference counting similar to the original Cairo:

```go
// Objects start with reference count 1
surface := cairo.NewImageSurface(format, width, height)

// Increase reference count
surface2 := surface.Reference()

// Decrease reference count  
surface.Destroy()   // Count becomes 1
surface2.Destroy()  // Count becomes 0, object freed
```

## Thread Safety

Go-Cairo objects are not thread-safe. Each thread should use its own context and objects, or external synchronization must be used.

## Performance Notes

- Image surfaces are backed by Go image types for easy interoperability
- Matrix operations use standard floating-point arithmetic
- Pattern creation and destruction are lightweight operations
- Context state management uses copy-on-write for efficient save/restore