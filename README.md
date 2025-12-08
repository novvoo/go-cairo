# Go-Cairo

A pure Go implementation of the Cairo 2D graphics library, fully compatible with the original Cairo C++ library API and behavior.

## Features

This implementation provides a complete port of Cairo's functionality to Go, including:

- **2D Vector Graphics**: Full support for vector graphics operations
- **Multiple Surface Types**: Image surfaces, PDF, SVG, and more
- **Pattern System**: Solid colors, gradients, and image patterns  
- **Path Operations**: Lines, curves, rectangles, arcs, and complex paths
- **Text Rendering**: Font selection and text drawing capabilities
- **Transformations**: Matrix operations and coordinate transformations
- **Clipping**: Path-based clipping regions

## Installation

```bash
go get github.com/novvoo/go-cairo
```

## Quick Start

```go
package main

import (
    "github.com/novvoo/go-cairo/pkg/cairo"
    "math"
)

func main() {
    // Create an image surface
    surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 200)
    defer surface.Destroy()
    
    // Create a context for drawing
    ctx := cairo.NewContext(surface)
    defer ctx.Destroy()
    
    // Set source color to red
    ctx.SetSourceRGB(1.0, 0.0, 0.0)
    
    // Draw a filled circle
    ctx.Arc(100, 100, 50, 0, 2*math.Pi)
    ctx.Fill()
    
    // Save to PNG
    surface.WriteToPNG("circle.png")
}
```

## API Compatibility

This library maintains API compatibility with the original Cairo library. Function names and parameters follow the same patterns, adapted for Go conventions:

- C function `cairo_move_to(cr, x, y)` becomes `ctx.MoveTo(x, y)`
- C function `cairo_set_source_rgb(cr, r, g, b)` becomes `ctx.SetSourceRGB(r, g, b)`
- Enums like `CAIRO_FORMAT_ARGB32` become constants like `cairo.FormatARGB32`

## Architecture

The library is organized into several packages:

- `pkg/cairo`: Main public API
- `internal/surface`: Surface implementations
- `internal/pattern`: Pattern implementations  
- `internal/path`: Path operations
- `internal/font`: Font and text handling
- `internal/image`: Image format support

## License

This project is dual-licensed under LGPL 2.1 and MPL 1.1, same as the original Cairo library.
