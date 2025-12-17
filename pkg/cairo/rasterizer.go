package cairo

import (
	"image"
	"image/color"
	"math"
	"sort"
)

// AdvancedRasterizer 高级光栅化器
// 实现高质量的路径光栅化，支持抗锯齿和子像素精度
type AdvancedRasterizer struct {
	width  int
	height int

	// 边缘表
	edges []Edge

	// 扫描线缓冲
	scanBuffer []float64

	// 抗锯齿级别 (1 = 无抗锯齿, 4 = 4x, 8 = 8x)
	aaLevel int
}

// Edge 表示一条边
type Edge struct {
	x0, y0 float64 // 起点
	x1, y1 float64 // 终点
	dir    int     // 方向 (1 = 向下, -1 = 向上)
}

// NewAdvancedRasterizer 创建新的光栅化器
func NewAdvancedRasterizer(width, height int) *AdvancedRasterizer {
	return &AdvancedRasterizer{
		width:      width,
		height:     height,
		edges:      make([]Edge, 0, 1024),
		scanBuffer: make([]float64, width*8),
		aaLevel:    8, // 默认 8x 抗锯齿
	}
}

// Reset 重置光栅化器
func (r *AdvancedRasterizer) Reset() {
	r.edges = r.edges[:0]
}

// AddEdge 添加一条边
func (r *AdvancedRasterizer) AddEdge(x0, y0, x1, y1 float64) {
	if y0 == y1 {
		return // 水平边不参与扫描
	}

	// 确保 y0 < y1
	if y0 > y1 {
		x0, y0, x1, y1 = x1, y1, x0, y0
		r.edges = append(r.edges, Edge{x0, y0, x1, y1, -1})
	} else {
		r.edges = append(r.edges, Edge{x0, y0, x1, y1, 1})
	}
}

// AddLine 添加直线
func (r *AdvancedRasterizer) AddLine(x0, y0, x1, y1 float64) {
	r.AddEdge(x0, y0, x1, y1)
}

// AddQuadraticBezier 添加二次贝塞尔曲线
func (r *AdvancedRasterizer) AddQuadraticBezier(x0, y0, x1, y1, x2, y2 float64) {
	// 自适应细分
	r.subdivideQuadratic(x0, y0, x1, y1, x2, y2, 0)
}

// AddCubicBezier 添加三次贝塞尔曲线
func (r *AdvancedRasterizer) AddCubicBezier(x0, y0, x1, y1, x2, y2, x3, y3 float64) {
	// 自适应细分
	r.subdivideCubic(x0, y0, x1, y1, x2, y2, x3, y3, 0)
}

// subdivideQuadratic 递归细分二次贝塞尔曲线
func (r *AdvancedRasterizer) subdivideQuadratic(x0, y0, x1, y1, x2, y2 float64, depth int) {
	if depth > 12 {
		r.AddEdge(x0, y0, x2, y2)
		return
	}

	// 检查平坦度
	dx := x2 - x0
	dy := y2 - y0
	d := math.Abs((x1-x2)*dy - (y1-y2)*dx)

	if d*d < 0.25*(dx*dx+dy*dy) {
		r.AddEdge(x0, y0, x2, y2)
		return
	}

	// De Casteljau 细分
	x01 := (x0 + x1) / 2
	y01 := (y0 + y1) / 2
	x12 := (x1 + x2) / 2
	y12 := (y1 + y2) / 2
	x012 := (x01 + x12) / 2
	y012 := (y01 + y12) / 2

	r.subdivideQuadratic(x0, y0, x01, y01, x012, y012, depth+1)
	r.subdivideQuadratic(x012, y012, x12, y12, x2, y2, depth+1)
}

// subdivideCubic 递归细分三次贝塞尔曲线
func (r *AdvancedRasterizer) subdivideCubic(x0, y0, x1, y1, x2, y2, x3, y3 float64, depth int) {
	if depth > 12 {
		r.AddEdge(x0, y0, x3, y3)
		return
	}

	// 检查平坦度
	dx := x3 - x0
	dy := y3 - y0
	d2 := math.Abs((x1-x3)*dy - (y1-y3)*dx)
	d3 := math.Abs((x2-x3)*dy - (y2-y3)*dx)

	if (d2+d3)*(d2+d3) < 0.25*(dx*dx+dy*dy) {
		r.AddEdge(x0, y0, x3, y3)
		return
	}

	// De Casteljau 细分
	x01 := (x0 + x1) / 2
	y01 := (y0 + y1) / 2
	x12 := (x1 + x2) / 2
	y12 := (y1 + y2) / 2
	x23 := (x2 + x3) / 2
	y23 := (y2 + y3) / 2
	x012 := (x01 + x12) / 2
	y012 := (y01 + y12) / 2
	x123 := (x12 + x23) / 2
	y123 := (y12 + y23) / 2
	x0123 := (x012 + x123) / 2
	y0123 := (y012 + y123) / 2

	r.subdivideCubic(x0, y0, x01, y01, x012, y012, x0123, y0123, depth+1)
	r.subdivideCubic(x0123, y0123, x123, y123, x23, y23, x3, y3, depth+1)
}

// Rasterize 光栅化到图像
func (r *AdvancedRasterizer) Rasterize(img *image.RGBA, c color.Color, fillRule FillRule) {
	if len(r.edges) == 0 {
		return
	}

	// 按 y 坐标排序边
	sort.Slice(r.edges, func(i, j int) bool {
		return r.edges[i].y0 < r.edges[j].y0
	})

	// 扫描线算法
	for y := 0; y < r.height; y++ {
		r.scanLine(img, y, c, fillRule)
	}
}

// scanLine 扫描一行
func (r *AdvancedRasterizer) scanLine(img *image.RGBA, y int, c color.Color, fillRule FillRule) {
	// 清空扫描缓冲
	for i := range r.scanBuffer {
		r.scanBuffer[i] = 0
	}

	// 对每个子像素行进行扫描
	for subY := 0; subY < r.aaLevel; subY++ {
		yf := float64(y) + float64(subY)/float64(r.aaLevel)

		// 收集与当前扫描线相交的边
		intersections := make([]float64, 0, 32)

		for i := range r.edges {
			edge := &r.edges[i]
			if edge.y0 <= yf && edge.y1 > yf {
				// 计算交点 x 坐标
				t := (yf - edge.y0) / (edge.y1 - edge.y0)
				x := edge.x0 + t*(edge.x1-edge.x0)
				intersections = append(intersections, x)
			}
		}

		// 排序交点
		sort.Float64s(intersections)

		// 填充像素
		for i := 0; i+1 < len(intersections); i += 2 {
			x0 := intersections[i]
			x1 := intersections[i+1]

			// 转换为像素坐标
			px0 := int(math.Floor(x0))
			px1 := int(math.Ceil(x1))

			for px := px0; px <= px1 && px < r.width; px++ {
				if px < 0 {
					continue
				}

				// 计算覆盖率
				coverage := 0.0
				pxf := float64(px)

				if pxf >= x0 && pxf+1 <= x1 {
					coverage = 1.0
				} else if pxf < x0 && pxf+1 > x0 {
					coverage = pxf + 1 - x0
				} else if pxf < x1 && pxf+1 > x1 {
					coverage = x1 - pxf
				} else if pxf >= x0 && pxf+1 <= x1 {
					coverage = 1.0
				}

				r.scanBuffer[px] += coverage / float64(r.aaLevel)
			}
		}
	}

	// 应用颜色
	for x := 0; x < r.width; x++ {
		coverage := r.scanBuffer[x]
		if coverage > 0 {
			coverage = math.Min(coverage, 1.0)
			r.blendPixel(img, x, y, c, coverage)
		}
	}
}

// blendPixel 混合像素
func (r *AdvancedRasterizer) blendPixel(img *image.RGBA, x, y int, c color.Color, alpha float64) {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return
	}

	sr, sg, sb, sa := c.RGBA()
	srcR := float64(sr>>8) / 255.0
	srcG := float64(sg>>8) / 255.0
	srcB := float64(sb>>8) / 255.0
	srcA := float64(sa>>8) / 255.0 * alpha

	dst := img.At(x, y)
	dr, dg, db, da := dst.RGBA()
	dstR := float64(dr>>8) / 255.0
	dstG := float64(dg>>8) / 255.0
	dstB := float64(db>>8) / 255.0
	dstA := float64(da>>8) / 255.0

	// Porter-Duff Over 混合
	outA := srcA + dstA*(1-srcA)
	var outR, outG, outB float64
	if outA > 0 {
		outR = (srcR*srcA + dstR*dstA*(1-srcA)) / outA
		outG = (srcG*srcA + dstG*dstA*(1-srcA)) / outA
		outB = (srcB*srcA + dstB*dstA*(1-srcA)) / outA
	}

	result := color.NRGBA{
		R: uint8(math.Min(math.Max(outR*255, 0), 255)),
		G: uint8(math.Min(math.Max(outG*255, 0), 255)),
		B: uint8(math.Min(math.Max(outB*255, 0), 255)),
		A: uint8(math.Min(math.Max(outA*255, 0), 255)),
	}

	img.Set(x, y, result)
}
