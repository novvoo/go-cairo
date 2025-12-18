//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"image"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 本示例演示 Go-Cairo 的四种边缘保持平滑算法：
// 1. 原始图像（带锯齿）- 用于对比
// 2. 边缘检测 + 高斯模糊 - 显式边缘检测，选择性平滑
// 3. 各向异性扩散 (Anisotropic Diffusion) - ⭐沿边缘方向扩散，经典算法
// 4. 双边滤波 (Bilateral Filter) - 同时考虑空间和颜色相似度
//
// 注意：所有图像都添加了锯齿效果，方便对比不同算法的平滑效果
//
// 重要说明：为什么需要添加噪点？
// 平滑算法作用于已经光栅化的像素数据，而不是矢量图形。
// Cairo 绘制的矢量图形本身已经经过抗锯齿处理，边缘相对平滑。
// 为了展示平滑算法的效果，我们：
// 1. 添加随机噪点（模拟图像噪声）
// 2. 多次应用平滑算法（累积效果）
// 3. 使用更大的平滑半径（增强效果）
//
// 本示例使用 SetSourceSurface + Paint() 来组合图像，并通过以下方式解决常见问题：
// 1. 使用 Save()/Restore() 保存和恢复绘图状态，避免状态污染
// 2. 使用 Rectangle() 设置裁剪区域，确保只在指定区域绘制
// 3. SetSourceSurface 的偏移参数控制源表面的绘制位置
// 4. 裁剪区域和源表面偏移配合使用，实现精确的图像定位

func main() {
	fmt.Println("=== Go-Cairo 边缘保持平滑算法演示 ===")
	fmt.Println("展示三种不同的边缘保持平滑算法")

	// 使用更大的图像尺寸以展示更明显的效果
	width, height := 300, 300

	// 1. 原始图像（带锯齿）
	fmt.Println("1. 绘制原始图像（添加锯齿）...")
	surface1 := createTestSurface(width, height)
	defer surface1.Destroy()
	addJaggies(surface1) // 添加锯齿效果
	// 注意：原始图像不需要平滑处理，直接添加标签
	addLabel(surface1, "1. 原始图像", 1)

	// 2. 边缘检测 + 高斯模糊
	fmt.Println("2. 应用边缘检测 + 高斯模糊...")
	surface2 := createTestSurface(width, height)
	defer surface2.Destroy()
	addJaggies(surface2)
	applySmoothToSurface(surface2, "edge_gaussian")
	// 在平滑处理后添加标签，避免文字被模糊
	addLabel(surface2, "2. 边缘检测", 2)

	// 3. 各向异性扩散
	fmt.Println("3. 应用各向异性扩散（Perona-Malik）...")
	surface3 := createTestSurface(width, height)
	defer surface3.Destroy()
	addJaggies(surface3)
	applySmoothToSurface(surface3, "anisotropic")
	// 在平滑处理后添加标签，避免文字被模糊
	addLabel(surface3, "3. 各向异性", 3)

	// 4. 双边滤波
	fmt.Println("4. 应用双边滤波...")
	surface4 := createTestSurface(width, height)
	defer surface4.Destroy()
	addJaggies(surface4)
	applySmoothToSurface(surface4, "bilateral")
	// 在平滑处理后添加标签，避免文字被模糊
	addLabel(surface4, "4. 双边滤波", 4)

	// 创建最终的组合图像 - 使用 SetSourceSurface + Paint() + 裁剪区域
	// 布局：[原始图像 | 高斯模糊 | 中值滤波 | 双边滤波]
	finalWidth := width * 4
	finalSurface := cairo.NewImageSurface(cairo.FormatARGB32, finalWidth, height)
	defer finalSurface.Destroy()

	finalCtx := cairo.NewContext(finalSurface)
	defer finalCtx.Destroy()

	// 绘制白色背景
	finalCtx.SetSourceRGB(1, 1, 1)
	finalCtx.Paint()

	// 使用 Save/Restore + Clip + SetSourceSurface 的正确方法
	//
	// 关键点：
	// 1. Save()/Restore() - 保存和恢复绘图状态（包括裁剪区域、变换矩阵等）
	// 2. Rectangle() + Clip() - 设置裁剪区域，限制绘制范围
	// 3. SetSourceSurface(surface, x, y) - 设置源表面和偏移量
	//    - surface: 源表面
	//    - x, y: 源表面左上角在目标表面中的位置
	// 4. Paint() - 在裁剪区域内绘制源表面
	//
	// 为什么需要裁剪区域？
	// 如果不设置裁剪区域，Paint() 会将整个源表面绘制到目标表面，
	// 可能覆盖其他区域或超出边界。裁剪区域确保只在指定矩形内绘制。

	// 绘制第一个图像（原始）到左侧 [0, 0] - [width, height]
	finalCtx.Save()
	// 设置裁剪区域：只允许在左侧区域绘制
	finalCtx.Rectangle(0, 0, float64(width), float64(height))
	finalCtx.Clip()
	// 设置源表面，偏移量 (0, 0) 表示源表面左上角对齐到目标的 (0, 0)
	finalCtx.SetSourceSurface(surface1, 0, 0)
	// Paint() 会在裁剪区域内绘制源表面
	finalCtx.Paint()
	finalCtx.Restore()

	// 绘制第二个图像（双线性插值）到中间 [width, 0] - [width*2, height]
	finalCtx.Save()
	// 设置裁剪区域：只允许在中间区域绘制
	finalCtx.Rectangle(float64(width), 0, float64(width), float64(height))
	finalCtx.Clip()
	// 设置源表面，偏移量 (width, 0) 表示源表面左上角对齐到目标的 (width, 0)
	finalCtx.SetSourceSurface(surface2, float64(width), 0)
	// Paint() 会在裁剪区域内绘制源表面
	finalCtx.Paint()
	finalCtx.Restore()

	// 绘制第三个图像（中值滤波）[width*2, 0] - [width*3, height]
	finalCtx.Save()
	finalCtx.Rectangle(float64(width*2), 0, float64(width), float64(height))
	finalCtx.Clip()
	finalCtx.SetSourceSurface(surface3, float64(width*2), 0)
	finalCtx.Paint()
	finalCtx.Restore()

	// 绘制第四个图像（双边滤波）[width*3, 0] - [width*4, height]
	finalCtx.Save()
	finalCtx.Rectangle(float64(width*3), 0, float64(width), float64(height))
	finalCtx.Clip()
	finalCtx.SetSourceSurface(surface4, float64(width*3), 0)
	finalCtx.Paint()
	finalCtx.Restore()

	// 添加分隔线以区分四个区域
	finalCtx.SetSourceRGB(0.5, 0.5, 0.5) // 灰色分隔线
	finalCtx.SetLineWidth(2)

	// 第一条分隔线
	finalCtx.MoveTo(float64(width), 0)
	finalCtx.LineTo(float64(width), float64(height))
	finalCtx.Stroke()

	// 第二条分隔线
	finalCtx.MoveTo(float64(width*2), 0)
	finalCtx.LineTo(float64(width*2), float64(height))
	finalCtx.Stroke()

	// 第三条分隔线
	finalCtx.MoveTo(float64(width*3), 0)
	finalCtx.LineTo(float64(width*3), float64(height))
	finalCtx.Stroke()

	// 保存结果
	imgSurface := finalSurface.(cairo.ImageSurface)
	status := imgSurface.WriteToPNG("smooth_demo.png")
	if status == cairo.StatusSuccess {
		fmt.Println("\n✓ 图像已保存到 smooth_demo.png")
	} else {
		fmt.Printf("\n✗ 保存失败: %v\n", status)
	}

	fmt.Println("\n=== 边缘保持平滑算法对比 ===")
	fmt.Println("🔴 第1格：原始图像（带锯齿）")
	fmt.Println("🟢 第2格：边缘检测 + 高斯模糊")
	fmt.Println("       (显式边缘检测 → 掩码羽化 → 选择性平滑)")
	fmt.Println("🔵 第3格：各向异性扩散 ⭐")
	fmt.Println("       (Perona-Malik算法，沿边缘方向扩散)")
	fmt.Println("🟠 第4格：双边滤波")
	fmt.Println("       (空间距离 + 颜色相似度双重权重)")
	fmt.Println("\n提示：所有图像都添加了锯齿效果，方便对比平滑效果")
	fmt.Println("      左上角彩色圆圈标识不同算法")
	fmt.Println("      观察边缘的平滑程度和锐利度")
	fmt.Println("==================")

	fmt.Println("\n=== SetSourceSurface 使用要点 ===")
	fmt.Println("✓ 使用 Save()/Restore() 保护绘图状态")
	fmt.Println("✓ 使用 Rectangle() + Clip() 设置裁剪区域")
	fmt.Println("✓ SetSourceSurface(surface, x, y) 设置源和偏移")
	fmt.Println("✓ Paint() 在裁剪区域内绘制")
	fmt.Println("==================")
}

// createTestSurface 创建包含多种图形的测试图案
// 包含圆形、矩形、三角形和线条，用于展示平滑效果
func createTestSurface(width, height int) cairo.Surface {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, width, height)
	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 绘制白色背景
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()

	// 绘制红色圆形 - 测试曲线平滑效果
	ctx.SetSourceRGB(1, 0, 0)
	ctx.Arc(100, 100, 60, 0, 2*3.14159)
	ctx.Fill()

	// 绘制绿色矩形 - 测试直角边缘平滑效果
	ctx.SetSourceRGB(0, 0.7, 0)
	ctx.Rectangle(150, 150, 80, 80)
	ctx.Fill()

	// 绘制蓝色三角形 - 测试锐角平滑效果
	ctx.SetSourceRGB(0, 0, 1)
	ctx.MoveTo(50, 250)
	ctx.LineTo(150, 250)
	ctx.LineTo(100, 180)
	ctx.ClosePath()
	ctx.Fill()

	// 绘制紫色线条 - 测试细线平滑效果
	ctx.SetSourceRGB(0.5, 0, 0.5)
	ctx.SetLineWidth(3)
	ctx.MoveTo(200, 50)
	ctx.LineTo(280, 150)
	ctx.Stroke()

	return surface
}

// addJaggies 向图像添加锯齿效果
// 通过在边缘添加像素噪点来模拟锯齿
func addJaggies(surface cairo.Surface) {
	imgSurface := surface.(cairo.ImageSurface)
	goImg := imgSurface.GetGoImage()

	if rgba, ok := goImg.(*image.RGBA); ok {
		bounds := rgba.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()

		// 检测边缘并添加锯齿
		for y := 1; y < height-1; y++ {
			for x := 1; x < width-1; x++ {
				center := rgba.At(x, y)
				_, _, _, ca := center.RGBA()

				// 检查是否是边缘（与邻居颜色不同）
				isEdge := false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}
						neighbor := rgba.At(x+dx, y+dy)
						_, _, _, na := neighbor.RGBA()
						if ca != na || (ca>>8) != (na>>8) {
							isEdge = true
							break
						}
					}
					if isEdge {
						break
					}
				}

				// 在边缘添加锯齿（随机偏移像素）
				if isEdge && (x*y)%3 == 0 {
					// 随机选择邻居的颜色
					offset := (x + y) % 4
					dx, dy := 0, 0
					switch offset {
					case 0:
						dx, dy = 1, 0
					case 1:
						dx, dy = -1, 0
					case 2:
						dx, dy = 0, 1
					case 3:
						dx, dy = 0, -1
					}
					if x+dx >= 0 && x+dx < width && y+dy >= 0 && y+dy < height {
						neighborColor := rgba.At(x+dx, y+dy)
						rgba.Set(x, y, neighborColor)
					}
				}
			}
		}
	}
}

// addLabel 在图像左上角添加文字标注
func addLabel(surface cairo.Surface, text string, labelNum int) {
	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 根据编号选择颜色
	var r, g, b float64
	switch labelNum {
	case 1:
		r, g, b = 1.0, 0.3, 0.3 // 红色
	case 2:
		r, g, b = 0.3, 0.8, 0.3 // 绿色
	case 3:
		r, g, b = 0.3, 0.5, 1.0 // 蓝色
	case 4:
		r, g, b = 1.0, 0.7, 0.2 // 橙色
	}

	// 使用 PangoCairo 绘制文字
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)

	fontDesc := cairo.NewPangoFontDescription()
	// 使用 sans 字体，系统会自动选择支持中文的字体
	fontDesc.SetFamily("sans")
	fontDesc.SetSize(14)
	fontDesc.SetWeight(cairo.PangoWeightBold)
	layout.SetFontDescription(fontDesc)
	layout.SetText(text)

	// 获取文字尺寸
	extents := layout.GetPixelExtents()
	fontExtents := layout.GetFontExtents()

	// 计算背景框尺寸（留出边距）
	padding := 5.0
	bgWidth := extents.Width + padding*2
	bgHeight := fontExtents.Height + padding*2

	// 绘制半透明背景
	ctx.SetSourceRGBA(0, 0, 0, 0.7)
	ctx.Rectangle(5, 5, bgWidth, bgHeight)
	ctx.Fill()

	// 绘制彩色文字
	// X: 左边距 + padding
	// Y: 上边距 + padding + Ascent（基线位置）
	ctx.SetSourceRGB(r, g, b)
	ctx.MoveTo(5+padding, 5+padding+fontExtents.Ascent)
	ctx.PangoCairoShowText(layout)
}

// applySmoothToSurface 对表面应用平滑处理
//
// 工作流程：
// 1. 从 Cairo 表面获取 Go image.RGBA 数据
// 2. 复制到 ImageBackend（提供平滑算法）
// 3. 应用指定的平滑算法
// 4. 将处理后的数据复制回 Cairo 表面
//
// 参数：
//   - surface: 要处理的 Cairo 表面
//   - method: 平滑方法 ("bilinear", "gaussian", "median")
func applySmoothToSurface(surface cairo.Surface, method string) {
	imgSurface := surface.(cairo.ImageSurface)
	goImg := imgSurface.GetGoImage()

	// 获取图像尺寸
	bounds := goImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建 ImageBackend - 提供高性能的像素级操作和平滑算法
	backend := cairo.NewImageBackend(width, height)
	backendImg := backend.GetImage()

	// 复制像素数据到后端
	// 使用 copy() 直接复制底层字节数组，比逐像素复制快得多
	if rgba, ok := goImg.(*image.RGBA); ok {
		copy(backendImg.Pix, rgba.Pix)
	}

	// 应用平滑算法
	switch method {
	case "edge_gaussian":
		// 边缘检测 + 高斯模糊：
		// 1. Sobel 算子检测边缘
		// 2. 创建边缘掩码并羽化
		// 3. 只对非边缘区域应用高斯模糊
		// smoothRadius=3: 高斯模糊半径
		// edgeThreshold=0.15: 边缘检测阈值（0-1，越小越敏感）
		backend.SmoothWithEdgeDetection(3, 0.15)
	case "anisotropic":
		// 各向异性扩散（Perona-Malik 算法）：
		// 通过控制扩散方向来保护边缘
		// iterations=10: 迭代次数
		// kappa=20: 扩散系数阈值（控制边缘敏感度）
		// lambda=0.2: 扩散速率
		backend.SmoothAnisotropicDiffusion(10, 20, 0.2)
	case "bilateral":
		// 双边滤波：同时考虑空间距离和颜色相似度
		// spatialSigma=3: 空间域标准差
		// colorSigma=30: 颜色域标准差
		backend.SmoothBilateral(3, 30)
	}

	// 重要：平滑算法会创建新的图像，需要重新获取
	backendImg = backend.GetImage()

	// 将平滑后的数据复制回 Cairo 表面
	// 这样原始表面就包含了平滑后的图像数据
	if rgba, ok := goImg.(*image.RGBA); ok {
		copy(rgba.Pix, backendImg.Pix)
	}

	// 注意：不要调用 MarkDirty()！
	// MarkDirty() 会从 ARGB 数据读取并覆盖 RGBA 数据，导致修改丢失
	// 因为我们直接修改了 RGBA 数据，所以不需要调用 MarkDirty()
}
