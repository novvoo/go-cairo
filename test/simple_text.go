//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("=== 简单文字渲染测试 ===\n")

	// Create a new image surface
	width, height := 400, 200
	surface := cairo.NewImageSurface(cairo.FormatARGB32, width, height)
	defer surface.Destroy()

	// Create a context
	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// Set background color to white
	ctx.SetSourceRGB(1, 1, 1)
	ctx.Paint()
	fmt.Println("✓ 白色背景已绘制")

	// Set text color to black
	ctx.SetSourceRGB(0, 0, 0)
	fmt.Println("✓ 设置文字颜色为黑色")

	// Use PangoCairo for text rendering
	layout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)

	// Set font description
	fontDesc := cairo.NewPangoFontDescription()
	fontDesc.SetFamily("sans")
	fontDesc.SetWeight(cairo.PangoWeightNormal)
	fontDesc.SetSize(24)
	layout.SetFontDescription(fontDesc)
	fmt.Println("✓ 使用 PangoCairo 设置字体")

	// Move to position
	ctx.MoveTo(50, 100)
	fmt.Println("✓ 移动到位置 (50, 100)")

	// Show text using PangoCairo
	text := "Hello, Cairo!"
	layout.SetText(text)
	ctx.PangoCairoShowText(layout)
	fmt.Printf("✓ 使用 PangoCairo 渲染: \"%s\"\n\n", text)

	// Save to PNG
	wd, _ := os.Getwd()
	filename := filepath.Join(wd, "simple_text_test.png")

	if imageSurface, ok := surface.(cairo.ImageSurface); ok {
		status := imageSurface.WriteToPNG(filename)
		if status != cairo.StatusSuccess {
			log.Fatal("保存 PNG 失败:", status)
		}
		fmt.Printf("✓ 图像已保存: %s\n", filename)
	}

	fmt.Println("\n=== 测试完成 ===")
}
