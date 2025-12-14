# Go-Cairo

A pure Go implementation of the Cairo 2D graphics library, fully compatible with the original Cairo C++ library API and behavior.

## Features

This implementation provides a complete port of Cairo's functionality to Go, including:

- **2D Vector Graphics**: Full support for vector graphics operations
- **Multiple Surface Types**: Image surfaces, PDF, SVG, and more
- **Pattern System**: Solid colors, gradients, and image patterns  
- **Path Operations**: Lines, curves, rectangles, arcs, and complex paths
- **Text Rendering**: Font selection and text drawing capabilities with full OpenType support
- **Advanced Typography**: 
  - Automatic text direction detection (LTR/RTL)
  - Language and script detection
  - OpenType features (ligatures, small caps, kerning, etc.)
  - Complex script shaping (Arabic, Hebrew, Indic, Thai, etc.)
  - Bidirectional text support
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

## OpenType 特性支持

go-cairo 现在支持完整的 OpenType 高级排版特性：

- ✅ **自动文本方向检测**：自动识别 LTR/RTL 文本
- ✅ **语言和文字系统检测**：支持 50+ 种语言和文字系统
- ✅ **OpenType 特性控制**：连字、小型大写、旧式数字等
- ✅ **复杂文字系统**：阿拉伯文、希伯来文、印地语、泰文等
- ✅ **双向文本处理**：正确处理混合方向文本
- ✅ **垂直排版**：支持中日韩文字的垂直书写

详细文档请参考：[OpenType 特性文档](docs/OPENTYPE_FEATURES.md)

### 快速示例

```go
// 自动检测并渲染多语言文本
texts := []string{
    "Hello World",      // 英文 (LTR)
    "مرحبا بالعالم",    // 阿拉伯文 (RTL)
    "你好世界",          // 中文 (LTR)
}

for _, text := range texts {
    // 自动检测方向、语言和文字系统
    options := cairo.NewShapingOptions()
    options.Direction = cairo.DetectTextDirection(text)
    options.Language = cairo.DetectLanguage(text)
    options.Script = cairo.DetectScript(text)
    
    ctx.ShowText(text) // 自动应用正确的文本塑形
}
```

## 测试示例

`test` 目录包含了多个示例程序，展示了 go-cairo 的各种功能：

### 综合功能测试

**sudoku.go** - 数独
- 基础图形绘制

![数独效果](test/sudoku.png)

**comprehensive.go** - 完整的功能演示，包括：
- 基础图形绘制（矩形、圆形、线条）
- 文本渲染和对齐
- 贝塞尔曲线
- 坐标变换

![综合测试效果](test/comprehensive_test.png)

运行方式：
```bash
cd test
go run comprehensive.go
```

### 圆形绘制对比

**circle_comparison.go** - 对比 `Arc` 和 `DrawCircle` 两种绘制圆形的方法

![圆形对比](test/circle_comparison.png)

### PangoCairo 文本渲染

**pangocairo.go** - 展示 PangoCairo 文本渲染功能：
- 字体加载和配置
- 文本度量和定位
- 字形分析

![PangoCairo 示例](test/pangocairo.png)

### 文本边界框分析

**mi_with_bounds.go** - 可视化文本边界框和字符间距

![文本边界框](test/mi_with_bounds.png)

### 渐变效果

**gradient.go** - 基础渐变效果演示：
- 线性渐变
- 径向渐变
- 多色渐变

![渐变效果](test/gradient_test.png)

**gradient_advanced.go** - 高级渐变效果：
- 复杂渐变模式
- 渐变变换
- 多重渐变组合

![高级渐变](test/gradient_advanced_test.png)

**chinese_gradient.go** - 中文文本渐变效果

![中文渐变](test/chinese_gradient_test.png)

运行方式：
```bash
cd test
go run gradient.go
go run gradient_advanced.go
go run chinese_gradient.go
```

详细文档请参考：[渐变效果文档](test/README_GRADIENT.md)

### 中文文本渲染

**chinese_text.go** - 中文文本渲染演示

![中文文本](test/chinese_text_test.png)

### 换行处理

**newline.go** - 文本换行和多行文本处理

![换行测试](test/newline_test.png)

### 字形分析工具

**glyph_analysis.go** - 字形渲染和碰撞检测分析

![字形分析](test/glyph_simple.png)
![字形碰撞](test/glyph_collision.png)

**glyph_outline_debug.go** - 字形轮廓调试工具，输出字形的详细信息

### OpenType 特性测试

**opentype_features_test.go** - OpenType 特性完整测试：
- 自动文本方向检测
- RTL 文本渲染
- 混合方向文本
- 连字特性
- 小型大写字母
- 复杂文字系统检测
- 语言和文字系统检测

![OpenType 特性](test/opentype_features_test.png)

**multilingual_text.go** - 多语言文本渲染演示：
- 10+ 种语言的自动检测和渲染
- LTR/RTL 自动处理
- 复杂文字系统支持

![多语言文本](test/multilingual_text.png)

运行方式：
```bash
cd test
go run opentype_features_test.go
go run multilingual_text.go
```

## License

This project is dual-licensed under LGPL 2.1 and MPL 1.1, same as the original Cairo library.
