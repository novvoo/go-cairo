// main.go - 终极修复版：移除 Hinting（可选参数），确保100%兼容
package main

import (
	"fmt"
	"os"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

func main() {
	const W, H = 400, 400
	dc := gg.NewContext(W, H)

	// 白色背景
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// 四角彩色方块
	dc.SetRGB(1, 0, 0)
	dc.DrawRectangle(50, 50, 30, 30)
	dc.Fill()

	dc.SetRGB(0, 1, 0)
	dc.DrawRectangle(W-80, 50, 30, 30)
	dc.Fill()

	dc.SetRGB(0, 0, 1)
	dc.DrawRectangle(50, H-80, 30, 30)
	dc.Fill()

	dc.SetRGB(1, 1, 0)
	dc.DrawRectangle(W-80, H-80, 30, 30)
	dc.Fill()

	// 对角线
	dc.SetRGB(0, 0, 0)
	dc.SetLineWidth(2)
	dc.DrawLine(0, 0, W, H)
	dc.DrawLine(0, H, W, 0)
	dc.Stroke()

	// 中心紫色圆
	dc.SetRGB(1, 0, 1)
	dc.SetLineWidth(3)
	dc.DrawCircle(W/2, H/2, 50)
	dc.Stroke()

	// 使用 Go 官方内置字体（goregular）—— 100% 无文件依赖
	ttf, _ := truetype.Parse(goregular.TTF)
	const dpi = 72

	// 小字体 18pt（移除 Hinting 参数，避免版本问题）
	face18 := truetype.NewFace(ttf, &truetype.Options{
		Size: 18,
		DPI:  dpi,
	})
	dc.SetFontFace(face18)
	dc.SetRGB(0, 0, 0)

	// 四角文字（完美对齐）
	dc.DrawStringAnchored("Top Left", 10, 20, 0, 1)
	dc.DrawStringAnchored("Top Right", W-10, 20, 1, 1)
	dc.DrawStringAnchored("Bottom Left", 10, H-10, 0, 0)
	dc.DrawStringAnchored("Bottom Right", W-10, H-10, 1, 0)

	// 中心大文字 42pt（移除 Hinting 参数）
	face42 := truetype.NewFace(ttf, &truetype.Options{
		Size: 42,
		DPI:  dpi,
	})
	dc.SetFontFace(face42)
	dc.SetRGB(0.15, 0.15, 0.15)
	dc.DrawStringAnchored("Center", W/2, H/2, 0.5, 0.5)

	// 青色贝塞尔曲线
	dc.SetRGB(0, 1, 1)
	dc.SetLineWidth(4)
	dc.MoveTo(100, 100)
	dc.CubicTo(150, 20, 250, 380, 300, 300)
	dc.Stroke()

	// 保存
	_ = os.Mkdir("images", 0755)
	err := dc.SavePNG("comprehensive_test_gg.png")
	if err != nil {
		panic(err)
	}

	fmt.Println("成功！图片已保存到 comprehensive_test_gg.png")
}
