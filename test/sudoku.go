//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"math"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

var puzzle = [9][9]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 3},
	{0, 0, 0, 0, 6, 3, 0, 4, 0},
	{0, 0, 4, 0, 0, 2, 6, 9, 7},
	{0, 9, 0, 7, 0, 0, 3, 1, 0},
	{3, 0, 0, 0, 0, 0, 0, 6, 4},
	{8, 0, 0, 0, 5, 0, 0, 0, 0},
	{0, 1, 0, 0, 0, 8, 2, 0, 0},
	{0, 7, 8, 0, 0, 0, 0, 0, 0},
	{4, 0, 2, 0, 0, 0, 0, 0, 0},
}

func main() {
	fmt.Println("🧩 Starting Sudoku rendering with Cairo...")

	// 创建 600x600 图像（方便网格计算）
	const width, height = 600, 600
	surface := cairo.NewImageSurface(cairo.FormatARGB32, width, height)
	defer surface.Destroy()

	status := surface.Status() // 👈 提前声明，作用域拉到整个函数
	if status != cairo.StatusSuccess {
		panic(fmt.Sprintf("Surface creation failed: %v", status))
	}
	fmt.Printf("✅ Surface created: %dx%d, status=%v\n", width, height, status)

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()
	fmt.Printf("✅ Context created, status=%v\n", ctx.Status())

	// 白底
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// 画网格：每格 60x60（留 30px 边距，600 - 2*30 = 540；540/9 = 60）
	const margin = 30.0
	const cellSize = 60.0
	startX, startY := margin, margin

	// 设置线宽
	ctx.SetLineWidth(1.0)
	ctx.SetSourceRGB(0.7, 0.7, 0.7) // 灰色细线

	// 绘制 10 条横线 + 10 条竖线
	for i := 0; i <= 9; i++ {
		y := startY + float64(i)*cellSize
		ctx.MoveTo(startX, y)
		ctx.LineTo(startX+9*cellSize, y)
		ctx.Stroke()

		x := startX + float64(i)*cellSize
		ctx.MoveTo(x, startY)
		ctx.LineTo(x, startY+9*cellSize)
		ctx.Stroke()
	}

	// 重绘粗线（每 3 格加粗）
	ctx.SetLineWidth(3.0)
	ctx.SetSourceRGB(0.2, 0.2, 0.2) // 深灰粗线

	for i := 0; i <= 3; i++ {
		y := startY + float64(i*3)*cellSize
		ctx.MoveTo(startX, y)
		ctx.LineTo(startX+9*cellSize, y)
		ctx.Stroke()

		x := startX + float64(i*3)*cellSize
		ctx.MoveTo(x, startY)
		ctx.LineTo(x, startY+9*cellSize)
		ctx.Stroke()
	}

	// 创建 PangoLayout 用于数字显示
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("Sans") // 使用通用无衬线字体（兼容性好）
	fontDesc.SetSize(24)       // 24 * PANGO_SCALE = 24pt ≈ 合适大小
	fontDesc.SetWeight(cairo.PangoWeightBold)
	layout.SetFontDescription(fontDesc)

	ctx.SetSourceRGB(0.2, 0.2, 0.2) // 深灰色数字

	// 绘制数字
	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			num := puzzle[row][col]
			if num == 0 {
				continue
			}

			// 单元格左上角
			x0 := startX + float64(col)*cellSize
			y0 := startY + float64(row)*cellSize

			// 文字内容
			text := fmt.Sprintf("%d", num)
			layout.SetText(text)
			extents := layout.GetPixelExtents()

			// 居中：使用文字边界框的实际高度来居中
			centerX := x0 + cellSize/2
			centerY := y0 + cellSize/2

			// 计算基线位置：
			// extents.Y 是负值（从基线向上的距离）
			// extents.Height 是边界框的实际高度
			// 单元格中心 - 文字视觉高度的一半 - extents.Y（调整到基线）
			textVisualHeight := float64(extents.Height)
			baselineY := centerY - textVisualHeight/2 - float64(extents.Y)

			drawX := centerX - float64(extents.Width)/2 - float64(extents.X)
			drawY := baselineY

			// 👇 为调试可开启（模仿你原风格）
			// fmt.Printf("Cell(%d,%d): num=%d, draw@(%5.1f,%5.1f), center=(%5.1f,%5.1f), extents(w=%d,h=%d)\n",
			// 	row, col, num, drawX, drawY, centerX, centerY, extents.Width, extents.Height)

			ctx.MoveTo(drawX, drawY)
			ctx.PangoCairoShowText(layout)
		}
	}

	// 保存 PNG
	fmt.Println("💾 Saving to sudoku.png...")
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("sudoku.png")
		if status != cairo.StatusSuccess {
			panic(fmt.Sprintf("WriteToPNG failed: %v", status))
		}
		fmt.Println("✅ sudoku.png saved successfully (600×600)")
	} else {
		panic("Unexpected surface type")
	}

	// 额外调试信息（按你风格加的）
	// 检查矩阵 & 坐标映射一致性
	fmt.Println("\n🔍 Final context state:")
	m := ctx.GetMatrix()
	fmt.Printf("   CTM — [XX=%.3f, YY=%.3f, X0=%.1f, Y0=%.1f]\n", m.XX, m.YY, m.X0, m.Y0)

	// 测试中心点映射
	devX, devY := 300.0, 300.0
	uX, uY := ctx.DeviceToUser(devX, devY)
	fmt.Printf("   Device(300,300) → User(%.2f, %.2f) [identity expected]\n", uX, uY)

	// 数字"5"在中心格 (4,4) 的绘制偏移分析（若存在）
	row, col := 4, 4
	if puzzle[row][col] != 0 {
		x0 := startX + float64(col)*cellSize
		y0 := startY + float64(row)*cellSize
		cx, cy := x0+cellSize/2, y0+cellSize/2
		layout.SetText("5")
		ext := layout.GetPixelExtents()
		textVisualHeight := float64(ext.Height)
		baseline := cy - textVisualHeight/2 - float64(ext.Y)
		drawX := cx - float64(ext.Width)/2 - float64(ext.X)
		drawY := baseline

		textCenterX := drawX + float64(ext.Width)/2
		textTop := drawY + float64(ext.Y)
		textBottom := textTop + float64(ext.Height)
		textCenterY := (textTop + textBottom) / 2

		dx, dy := math.Abs(cx-textCenterX), math.Abs(cy-textCenterY)
		fmt.Printf("   Cell(4,4) center=(%.1f,%.1f), text center=(%.1f,%.1f), Δ=(%.2f,%.2f)\n",
			cx, cy, textCenterX, textCenterY, dx, dy)
		if dx < 0.5 && dy < 0.5 {
			fmt.Println("   ✅ Perfect centering!")
		}
	}

	fmt.Println("🎉 Sudoku rendering complete!")
}
