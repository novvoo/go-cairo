//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"math"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 绘制圆角矩形路径
func drawRoundedRect(ctx cairo.Context, x, y, width, height, radius float64) {
	ctx.NewPath()
	ctx.Arc(x+radius, y+radius, radius, math.Pi, 1.5*math.Pi)         // 左上
	ctx.Arc(x+width-radius, y+radius, radius, 1.5*math.Pi, 2*math.Pi) // 右上
	ctx.Arc(x+width-radius, y+height-radius, radius, 0, 0.5*math.Pi)  // 右下
	ctx.Arc(x+radius, y+height-radius, radius, 0.5*math.Pi, math.Pi)  // 左下
	ctx.ClosePath()
}

func main() {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 500, 300)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// === 1. 渐变背景 (淡灰色到白色)
	bgGradient := cairo.NewPatternLinear(0, 0, 0, 300)
	if bg, ok := bgGradient.(cairo.LinearGradientPattern); ok {
		bg.AddColorStopRGB(0, 0.95, 0.95, 0.95)
		bg.AddColorStopRGB(1, 1, 1, 1)
		ctx.SetSource(bg)
	}
	ctx.Paint()

	// === 2. 半透明黑色蒙版
	ctx.SetSourceRGBA(0, 0, 0, 0.2)
	ctx.Rectangle(0, 0, 500, 300)
	ctx.Fill()

	// === 3. 圆角矩形按钮
	rectX, rectY := 150.0, 100.0
	rectWidth, rectHeight := 200.0, 100.0
	rectRadius := 15.0

	// 绘制阴影效果
	ctx.SetSourceRGBA(0, 0, 0, 0.3)
	drawRoundedRect(ctx, rectX+3, rectY+3, rectWidth, rectHeight, rectRadius)
	ctx.Fill()

	// 绘制主按钮
	drawRoundedRect(ctx, rectX, rectY, rectWidth, rectHeight, rectRadius)
	// 渐变填充
	buttonGradient := cairo.NewPatternLinear(rectX, rectY, rectX, rectY+rectHeight)
	if grad, ok := buttonGradient.(cairo.LinearGradientPattern); ok {
		grad.AddColorStopRGB(0, 0.2, 0.6, 0.8) // 蓝色顶部
		grad.AddColorStopRGB(1, 0.1, 0.4, 0.6) // 深蓝色底部
		ctx.SetSource(grad)
	}
	ctx.FillPreserve()

	// 边框
	ctx.SetSourceRGBA(1, 1, 1, 0.8)
	ctx.SetLineWidth(2)
	ctx.Stroke()

	// === 4. 使用PangoCairo渲染"MI"文字
	// 使用PangoCairo创建布局
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)

	// 设置字体描述
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("Go Regular")
	fontDesc.SetWeight(cairo.PangoWeightBold)
	fontDesc.SetSize(48.0)
	layout.SetFontDescription(fontDesc)

	// 获取文本尺寸
	layout.SetText("M")
	mExtents := layout.GetPixelExtents()
	layout.SetText("I")
	iExtents := layout.GetPixelExtents()
	totalWidth := mExtents.Width + iExtents.Width + 2.0 // 加上间距

	// 计算居中位置
	fontExtents := layout.GetFontExtents()
	midx := rectX + rectWidth/2 - totalWidth/2
	midy := rectY + rectHeight/2 + fontExtents.Ascent/2 - fontExtents.Descent/2

	// === 坐标系统诊断 ===
	matrix := ctx.GetMatrix()
	fmt.Printf("\n=== 坐标系统诊断 ===\n")
	fmt.Printf("当前变换矩阵: xx=%.2f, yx=%.2f, xy=%.2f, yy=%.2f, x0=%.2f, y0=%.2f\n",
		matrix.XX, matrix.YX, matrix.XY, matrix.YY, matrix.X0, matrix.Y0)

	isFlippedY := matrix.YY < 0
	if isFlippedY {
		fmt.Printf("✅ Y轴翻转: 是 (YY=%.2f < 0) - 这是Cairo的标准配置\n", matrix.YY)
		fmt.Printf("   → 库会自动处理文字方向，无需手动修正\n")
	} else {
		fmt.Printf("⚠️  Y轴翻转: 否 (YY=%.2f >= 0) - 非标准配置\n", matrix.YY)
	}

	// === 计算字母间距，防止重叠 ===
	baseSpacing := 2.0
	dynamicSpacing := fontExtents.Height * 0.05 // 字体高度的5%
	letterSpacing := math.Max(baseSpacing, dynamicSpacing)

	// === 渲染文字（让库自动处理坐标系）===
	fmt.Printf("\n=== 渲染文字 ===\n")

	// 渲染 M
	layout.SetText("M")
	ctx.SetSourceRGB(1.0, 1.0, 1.0)
	ctx.MoveTo(midx, midy)
	ctx.PangoCairoShowText(layout)
	fmt.Printf("✓ 渲染字母 'M' 在位置 (%.2f, %.2f)\n", midx, midy)

	// 渲染 I（使用动态间距）
	iStartX := midx + mExtents.Width + letterSpacing
	layout.SetText("I")
	ctx.MoveTo(iStartX, midy)
	ctx.PangoCairoShowText(layout)
	fmt.Printf("✓ 渲染字母 'I' 在位置 (%.2f, %.2f)\n", iStartX, midy)
	fmt.Printf("✓ 字母间距: %.2f 像素\n", letterSpacing)

	// === 检测字母重叠 ===
	// 计算M的实际边界框
	mLeft := midx + mExtents.X
	mRight := midx + mExtents.X + mExtents.Width
	mTop := midy + mExtents.Y
	mBottom := midy + mExtents.Y + mExtents.Height

	// 计算I的实际边界框
	iLeft := iStartX + iExtents.X
	iRight := iStartX + iExtents.X + iExtents.Width
	iTop := midy + iExtents.Y
	iBottom := midy + iExtents.Y + iExtents.Height

	// 检测水平重叠
	hasOverlap := false
	overlapAmount := 0.0
	if mRight > iLeft {
		hasOverlap = true
		overlapAmount = mRight - iLeft
	}

	fmt.Printf("\n=== 字母重叠检测 ===\n")
	fmt.Printf("M边界: 左=%.2f, 右=%.2f, 上=%.2f, 下=%.2f\n", mLeft, mRight, mTop, mBottom)
	fmt.Printf("I边界: 左=%.2f, 右=%.2f, 上=%.2f, 下=%.2f\n", iLeft, iRight, iTop, iBottom)
	fmt.Printf("字母间距: %.2f 像素\n", letterSpacing)
	fmt.Printf("实际间隙: %.2f 像素\n", iLeft-mRight)

	if hasOverlap {
		fmt.Printf("❌ 检测到字母重叠！重叠量: %.2f 像素\n", overlapAmount)
		fmt.Printf("重叠原因分析:\n")
		fmt.Printf("  - M的右边界(%.2f) > I的左边界(%.2f)\n", mRight, iLeft)
		fmt.Printf("  - 可能原因: 字体度量不准确或间距设置过小\n")
		fmt.Printf("  - 建议间距: %.2f 像素（当前: %.2f）\n", overlapAmount+5.0, letterSpacing)
	} else {
		fmt.Printf("✅ 字母无重叠，间隙正常\n")
	}

	// === 详细位置信息 ===
	fmt.Printf("\n=== MI 整体位置信息 ===\n")
	fmt.Printf("MI起始位置: (%.2f, %.2f)\n", midx, midy)
	fmt.Printf("MI文本总宽度: %.2f\n", totalWidth)
	fmt.Printf("实际渲染宽度: %.2f\n", iRight-mLeft)

	fmt.Printf("\n=== 字母 M 详细信息 ===\n")
	fmt.Printf("M起始位置: (%.2f, %.2f)\n", midx, midy)
	fmt.Printf("M文本范围: 宽度=%.2f, 高度=%.2f\n", mExtents.Width, mExtents.Height)
	fmt.Printf("M边界偏移: X=%.2f, Y=%.2f\n", mExtents.X, mExtents.Y)
	fmt.Printf("M实际边界框:\n")
	fmt.Printf("  左上角: (%.2f, %.2f)\n", mLeft, mTop)
	fmt.Printf("  右上角: (%.2f, %.2f)\n", mRight, mTop)
	fmt.Printf("  左下角: (%.2f, %.2f)\n", mLeft, mBottom)
	fmt.Printf("  右下角: (%.2f, %.2f)\n", mRight, mBottom)

	fmt.Printf("\n=== 字母 I 详细信息 ===\n")
	fmt.Printf("I起始位置: (%.2f, %.2f)\n", iStartX, midy)
	fmt.Printf("I文本范围: 宽度=%.2f, 高度=%.2f\n", iExtents.Width, iExtents.Height)
	fmt.Printf("I边界偏移: X=%.2f, Y=%.2f\n", iExtents.X, iExtents.Y)
	fmt.Printf("I实际边界框:\n")
	fmt.Printf("  左上角: (%.2f, %.2f)\n", iLeft, iTop)
	fmt.Printf("  右上角: (%.2f, %.2f)\n", iRight, iTop)
	fmt.Printf("  左下角: (%.2f, %.2f)\n", iLeft, iBottom)
	fmt.Printf("  右下角: (%.2f, %.2f)\n", iRight, iBottom)

	// === 5. 保存图片
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		status := imgSurf.WriteToPNG("mi_pango.png")
		if status != cairo.StatusSuccess {
			fmt.Printf("保存PNG失败: %v\n", status)
		} else {
			fmt.Println("\n✅ MI示例已保存到 mi_pango.png")
		}
	}
}
