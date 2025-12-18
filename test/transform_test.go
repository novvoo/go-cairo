package cairo

import (
	"math"
	"testing"

	"github.com/novvoo/go-cairo/pkg/cairo"
)

// 测试坐标变换
func TestUserToDevice(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 应用平移
	ctx.Translate(10, 20)

	// 测试点变换
	x, y := ctx.UserToDevice(5, 5)
	if x != 15 || y != 25 {
		t.Errorf("UserToDevice failed: expected (15, 25), got (%f, %f)", x, y)
	}
}

// 测试距离变换
func TestUserToDeviceDistance(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 应用缩放
	ctx.Scale(2, 3)

	// 测试距离变换
	dx, dy := ctx.UserToDeviceDistance(10, 10)
	if dx != 20 || dy != 30 {
		t.Errorf("UserToDeviceDistance failed: expected (20, 30), got (%f, %f)", dx, dy)
	}
}

// 测试逆变换
func TestDeviceToUser(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 应用平移
	ctx.Translate(10, 20)

	// 测试逆变换
	x, y := ctx.DeviceToUser(15, 25)
	if math.Abs(x-5) > 0.001 || math.Abs(y-5) > 0.001 {
		t.Errorf("DeviceToUser failed: expected (5, 5), got (%f, %f)", x, y)
	}
}

// 测试组合变换
func TestCombinedTransforms(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 组合变换：平移 + 缩放 + 旋转
	ctx.Translate(50, 50)
	ctx.Scale(2, 2)
	ctx.Rotate(math.Pi / 4)

	// 绘制矩形
	ctx.Rectangle(-10, -10, 20, 20)
	ctx.SetSourceRGB(1, 0, 0)
	err := ctx.Fill()
	if err != nil {
		t.Errorf("Fill with combined transforms failed: %v", err)
	}
}

// 测试 IdentityMatrix
func TestIdentityMatrix(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 应用一些变换
	ctx.Translate(10, 20)
	ctx.Scale(2, 3)

	// 重置为单位矩阵
	ctx.IdentityMatrix()

	matrix := ctx.GetMatrix()
	if matrix.XX != 1 || matrix.YY != 1 || matrix.X0 != 0 || matrix.Y0 != 0 {
		t.Error("IdentityMatrix failed to reset transformation")
	}
}

// 测试 SetMatrix 和 GetMatrix
func TestSetGetMatrix(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 创建自定义矩阵
	matrix := cairo.NewMatrix()
	matrix.InitScale(2, 3)
	matrix.X0 = 10
	matrix.Y0 = 20

	// 设置矩阵
	ctx.SetMatrix(matrix)

	// 获取并验证
	result := ctx.GetMatrix()
	if result.XX != 2 || result.YY != 3 || result.X0 != 10 || result.Y0 != 20 {
		t.Error("SetMatrix/GetMatrix failed")
	}
}

// 测试变换对路径的影响
func TestTransformOnPath(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 200, 200)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 在不同变换下绘制相同的形状
	for i := 0; i < 4; i++ {
		ctx.Save()
		ctx.Translate(100, 100)
		ctx.Rotate(float64(i) * math.Pi / 2)
		ctx.Translate(-100, -100)

		ctx.Rectangle(90, 90, 20, 20)
		ctx.SetSourceRGB(float64(i)/4, 0, 1-float64(i)/4)
		ctx.Fill()

		ctx.Restore()
	}
}

// 测试变换栈
func TestTransformStack(t *testing.T) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	// 保存初始状态
	ctx.Save()
	ctx.Translate(10, 10)

	// 再次保存
	ctx.Save()
	ctx.Scale(2, 2)

	matrix := ctx.GetMatrix()
	if matrix.XX != 2 || matrix.X0 != 10 {
		t.Error("Transform stack level 2 incorrect")
	}

	// 恢复一次
	ctx.Restore()
	matrix = ctx.GetMatrix()
	if matrix.XX != 1 || matrix.X0 != 10 {
		t.Error("Transform stack level 1 incorrect after restore")
	}

	// 再恢复一次
	ctx.Restore()
	matrix = ctx.GetMatrix()
	if matrix.X0 != 0 {
		t.Error("Transform stack level 0 incorrect after restore")
	}
}

// 基准测试：变换操作
func BenchmarkTransform(b *testing.B) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.Translate(1, 1)
		ctx.Scale(1.1, 1.1)
		ctx.Rotate(0.1)
	}
}

// 基准测试：坐标转换
func BenchmarkCoordinateTransform(b *testing.B) {
	surface := cairo.NewImageSurface(cairo.FormatARGB32, 100, 100)
	defer surface.Destroy()

	ctx := cairo.NewContext(surface)
	defer ctx.Destroy()

	ctx.Translate(10, 20)
	ctx.Scale(2, 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.UserToDevice(50, 50)
		ctx.DeviceToUser(100, 150)
	}
}
