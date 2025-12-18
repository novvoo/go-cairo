//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

func main() {
	fmt.Println("=== 换行符测试 ===")
	fmt.Println("测试不同操作系统的换行符处理")

	// 创建画布
	width, height := 600, 400
	surface := cairo.NewImageSurface(cairo.FormatARGB32, width, height)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 白色背景
	ctx.SetSourceRGB(1.0, 1.0, 1.0)
	ctx.Paint()

	// 黑色文字
	ctx.SetSourceRGB(0.0, 0.0, 0.0)

	// 创建标题字体
	titleLayout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
	titleFontDesc := cairo.NewPangoFontDescription()
	titleFontDesc.SetFamily("sans-serif")
	titleFontDesc.SetSize(16.0)
	titleLayout.SetFontDescription(titleFontDesc)

	// 创建内容字体
	contentFontDesc := cairo.NewPangoFontDescription()
	contentFontDesc.SetFamily("sans-serif")
	contentFontDesc.SetSize(20.0)

	// 测试不同的换行符
	testCases := []struct {
		name string
		text string
		y    float64
	}{
		{
			name: "Unix/Linux (\\n)",
			text: "第一行\n第二行\n第三行",
			y:    50.0,
		},
		{
			name: "Windows (\\r\\n)",
			text: "第一行\r\n第二行\r\n第三行",
			y:    150.0,
		},
		{
			name: "Old Mac (\\r)",
			text: "第一行\r第二行\r第三行",
			y:    250.0,
		},
	}

	for _, tc := range testCases {
		fmt.Printf("测试: %s\n", tc.name)
		fmt.Printf("文本: %q\n", tc.text)

		// 渲染标题
		ctx.MoveTo(50.0, tc.y-25.0)
		titleLayout.SetText(tc.name)
		ctx.PangoCairoShowText(titleLayout)

		// 为每个测试用例创建新的 layout
		contentLayout := ctx.PangoCairoCreateLayout().(*cairo.PangoCairoLayout)
		contentLayout.SetFontDescription(contentFontDesc)

		// 标准化换行符：将 \r\n 和 \r 都转换为 \n
		normalizedText := strings.ReplaceAll(tc.text, "\r\n", "\n")
		normalizedText = strings.ReplaceAll(normalizedText, "\r", "\n")

		// 设置位置并渲染内容
		ctx.MoveTo(50.0, tc.y)
		contentLayout.SetText(normalizedText)
		ctx.PangoCairoShowText(contentLayout)

		fmt.Printf("渲染完成\n\n")
	}

	// 保存图片
	wd, _ := os.Getwd()
	filename := filepath.Join(wd, "newline_test.png")
	fmt.Printf("保存路径: %s\n", filename)

	if imageSurface, ok := surface.(cairo.ImageSurface); ok {
		status := imageSurface.WriteToPNG(filename)
		if status != cairo.StatusSuccess {
			log.Fatal("保存 PNG 失败:", status)
		}
		fmt.Printf("✓ 图像已成功保存\n")
	} else {
		log.Fatal("Surface 不是 ImageSurface 类型")
	}

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("请检查生成的图像，确认:")
	fmt.Println("1. 每组文本都有三行")
	fmt.Println("2. 每行的 X 坐标都从左边开始（X=50）")
	fmt.Println("3. 行与行之间有适当的间距")
}
