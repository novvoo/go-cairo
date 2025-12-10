//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// verifyTextOrientation 验证文本方向并记录详细的翻转信息
func verifyTextOrientation(ctx cairo.Context, text string, x, y float64) {
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Printf("=== 文本方向验证: \"%s\" ===\n", text)
	fmt.Printf(strings.Repeat("=", 70) + "\n\n")

	// 1. 获取当前变换矩阵
	matrix := ctx.GetMatrix()
	fmt.Printf("【变换矩阵信息】\n")
	fmt.Printf("  XX (X轴缩放): %.6f\n", matrix.XX)
	fmt.Printf("  YX (X轴倾斜): %.6f\n", matrix.YX)
	fmt.Printf("  XY (Y轴倾斜): %.6f\n", matrix.XY)
	fmt.Printf("  YY (Y轴缩放): %.6f\n", matrix.YY)
	fmt.Printf("  X0 (X平移):   %.6f\n", matrix.X0)
	fmt.Printf("  Y0 (Y平移):   %.6f\n\n", matrix.Y0)

	// 2. 分析坐标系翻转状态
	isFlippedX := matrix.XX < 0
	isFlippedY := matrix.YY < 0
	isRotated := math.Abs(matrix.YX) > 0.001 || math.Abs(matrix.XY) > 0.001

	fmt.Printf("【坐标系状态分析】\n")
	if isFlippedX {
		fmt.Printf("  ❌ X轴翻转: 是 (XX=%.6f < 0)\n", matrix.XX)
	} else {
		fmt.Printf("  ✅ X轴翻转: 否 (XX=%.6f >= 0)\n", matrix.XX)
	}

	if isFlippedY {
		fmt.Printf("  ❌ Y轴翻转: 是 (YY=%.6f < 0)\n", matrix.YY)
	} else {
		fmt.Printf("  ✅ Y轴翻转: 否 (YY=%.6f >= 0)\n", matrix.YY)
	}

	if isRotated {
		angle := math.Atan2(matrix.YX, matrix.XX) * 180 / math.Pi
		fmt.Printf("  🔄 旋转角度: %.2f度\n", angle)
	} else {
		fmt.Printf("  ✅ 旋转角度: 0度 (无旋转)\n")
	}
	fmt.Println()

	// 3. 创建PangoCairo布局
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("sans")
	fontDesc.SetSize(24.0)
	layout.SetFontDescription(fontDesc)
	layout.SetText(text)

	// 4. 获取字体度量信息
	fontExtents := layout.GetFontExtents()
	fmt.Printf("【字体度量信息】\n")
	fmt.Printf("  Ascent (上升高度):  %.2f\n", fontExtents.Ascent)
	fmt.Printf("  Descent (下降高度): %.2f\n", fontExtents.Descent)
	fmt.Printf("  Height (总高度):    %.2f\n", fontExtents.Height)
	fmt.Printf("  LineGap (行间距):   %.2f\n", fontExtents.LineGap)
	fmt.Println()

	// 5. 获取文本范围
	textExtents := layout.GetPixelExtents()
	fmt.Printf("【文本范围信息】\n")
	fmt.Printf("  X偏移:  %.2f\n", textExtents.X)
	fmt.Printf("  Y偏移:  %.2f\n", textExtents.Y)
	fmt.Printf("  宽度:   %.2f\n", textExtents.Width)
	fmt.Printf("  高度:   %.2f\n", textExtents.Height)
	fmt.Println()

	// 6. 计算实际边界框
	actualLeft := x + textExtents.X
	actualRight := x + textExtents.X + textExtents.Width
	actualTop := y + textExtents.Y
	actualBottom := y + textExtents.Y + textExtents.Height

	fmt.Printf("【文本边界框】\n")
	fmt.Printf("  渲染位置: (%.2f, %.2f)\n", x, y)
	fmt.Printf("  左边界:   %.2f\n", actualLeft)
	fmt.Printf("  右边界:   %.2f\n", actualRight)
	fmt.Printf("  上边界:   %.2f\n", actualTop)
	fmt.Printf("  下边界:   %.2f\n", actualBottom)
	fmt.Printf("  中心点:   (%.2f, %.2f)\n", (actualLeft+actualRight)/2, (actualTop+actualBottom)/2)
	fmt.Println()

	// 7. 检测文本方向问题
	fmt.Printf("【文本方向诊断】\n")
	hasIssue := false

	if isFlippedY {
		fmt.Printf("  ⚠️  检测到Y轴翻转\n")
		fmt.Printf("      - 这会导致文本上下颠倒\n")
		fmt.Printf("      - 原因: 字体矩阵的YY分量为负值\n")
		fmt.Printf("      - 解决: 使用负的Y缩放 (fontMatrix.InitScale(size, -size))\n")
		hasIssue = true
	}

	if isFlippedX {
		fmt.Printf("  ⚠️  检测到X轴翻转\n")
		fmt.Printf("      - 这会导致文本左右镜像\n")
		fmt.Printf("      - 原因: 字体矩阵的XX分量为负值\n")
		hasIssue = true
	}

	if textExtents.Y > 0 {
		fmt.Printf("  ⚠️  文本Y偏移为正值 (%.2f)\n", textExtents.Y)
		fmt.Printf("      - 这可能表示文本基线位置不正确\n")
		fmt.Printf("      - 正常情况下Y偏移应该为负值（文本在基线上方）\n")
		hasIssue = true
	}

	if !hasIssue {
		fmt.Printf("  ✅ 文本方向正常，无翻转问题\n")
	}
	fmt.Println()

	// 8. 提供修复建议
	if hasIssue {
		fmt.Printf("【修复建议】\n")
		if isFlippedY {
			fmt.Printf("  1. 在创建ScaledFont时使用负的Y缩放:\n")
			fmt.Printf("     fontMatrix.InitScale(fontSize, -fontSize)\n\n")
			fmt.Printf("  2. 在GlyphPath函数中正确处理Y轴翻转:\n")
			fmt.Printf("     flipY := s.fontMatrix.YY < 0\n\n")
		}
		if isFlippedX {
			fmt.Printf("  3. 检查是否错误地应用了X轴镜像变换\n\n")
		}
	}

	// 9. 渲染文本用于视觉验证
	ctx.MoveTo(x, y)
	ctx.PangoCairoShowText(layout)

	fmt.Printf("【渲染完成】\n")
	fmt.Printf("  文本 \"%s\" 已渲染到位置 (%.2f, %.2f)\n", text, x, y)
	fmt.Printf(strings.Repeat("=", 70) + "\n\n")
}

func main() {
	fmt.Println("🔍 文本方向验证工具")
	fmt.Println("=" + strings.Repeat("=", 69))
	fmt.Println()

	// 创建测试表面
	width, height := 800, 600
	surface := cairo.NewImageSurface(cairo.FormatARGB32, width, height)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 设置白色背景
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// 设置文本颜色为黑色
	ctx.SetSourceRGB(0, 0, 0)

	// 测试1: 标准文本（无变换）
	fmt.Println("【测试 1】标准文本渲染（无变换）")
	verifyTextOrientation(ctx, "Hello, Cairo!", 50, 100)

	// 测试2: 带Y轴翻转的文本
	fmt.Println("【测试 2】Y轴翻转测试")
	ctx.Save()
	ctx.Scale(1, -1)
	ctx.Translate(0, -300)
	verifyTextOrientation(ctx, "Flipped Y", 50, 200)
	ctx.Restore()

	// 测试3: 带旋转的文本
	fmt.Println("【测试 3】旋转文本测试")
	ctx.Save()
	ctx.Translate(400, 300)
	ctx.Rotate(math.Pi / 6) // 30度
	verifyTextOrientation(ctx, "Rotated", 0, 0)
	ctx.Restore()

	// 测试4: 缩放文本
	fmt.Println("【测试 4】缩放文本测试")
	ctx.Save()
	ctx.Scale(1.5, 1.5)
	verifyTextOrientation(ctx, "Scaled", 50, 300)
	ctx.Restore()

	// 保存图像
	if imgSurf, ok := surface.(cairo.ImageSurface); ok {
		filename := "text_orientation_test.png"
		status := imgSurf.WriteToPNG(filename)
		if status != cairo.StatusSuccess {
			fmt.Printf("❌ 保存PNG失败: %v\n", status)
			os.Exit(1)
		}
		fmt.Printf("✅ 测试图像已保存到: %s\n", filename)
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("✅ 文本方向验证完成")
}
