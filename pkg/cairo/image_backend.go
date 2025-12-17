package cairo

import (
	"image"
	"image/color"
	"math"
)

// ImageBackend 图像后端
// 提供高性能的像素级操作
type ImageBackend struct {
	img    *image.RGBA
	width  int
	height int
}

// NewImageBackend 创建新的图像后端
func NewImageBackend(width, height int) *ImageBackend {
	return &ImageBackend{
		img:    image.NewRGBA(image.Rect(0, 0, width, height)),
		width:  width,
		height: height,
	}
}

// GetImage 获取图像
func (b *ImageBackend) GetImage() *image.RGBA {
	return b.img
}

// Clear 清空图像
func (b *ImageBackend) Clear(c color.Color) {
	r, g, bl, a := c.RGBA()
	fillColor := color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(bl >> 8),
		A: uint8(a >> 8),
	}

	for y := 0; y < b.height; y++ {
		for x := 0; x < b.width; x++ {
			b.img.Set(x, y, fillColor)
		}
	}
}

// FillRect 填充矩形
func (b *ImageBackend) FillRect(x, y, width, height int, c color.Color) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && py >= 0 && px < b.width && py < b.height {
				b.img.Set(px, py, c)
			}
		}
	}
}

// BlendPixel 混合单个像素
func (b *ImageBackend) BlendPixel(x, y int, c color.Color, op Operator) {
	if x < 0 || y < 0 || x >= b.width || y >= b.height {
		return
	}

	src := colorToNRGBA(c)
	dst := colorToNRGBA(b.img.At(x, y))
	result := PorterDuffBlend(src, dst, op)
	b.img.Set(x, y, result)
}

// DrawLine 绘制直线
func (b *ImageBackend) DrawLine(x0, y0, x1, y1 int, c color.Color, width float64) {
	// Bresenham 算法
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		b.drawThickPixel(x0, y0, c, width)

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// drawThickPixel 绘制粗像素
func (b *ImageBackend) drawThickPixel(x, y int, c color.Color, width float64) {
	halfWidth := int(math.Ceil(width / 2))
	for dy := -halfWidth; dy <= halfWidth; dy++ {
		for dx := -halfWidth; dx <= halfWidth; dx++ {
			px := x + dx
			py := y + dy
			if px >= 0 && py >= 0 && px < b.width && py < b.height {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist <= width/2 {
					b.img.Set(px, py, c)
				}
			}
		}
	}
}

// colorToNRGBA 转换颜色为 NRGBA
func colorToNRGBA(c color.Color) color.NRGBA {
	if nrgba, ok := c.(color.NRGBA); ok {
		return nrgba
	}
	r, g, b, a := c.RGBA()
	return color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}
