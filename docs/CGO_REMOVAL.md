# CGO 移除说明

## 概述

本项目已完全移除 CGO 依赖，实现了纯 Go 的 Cairo 2D 图形库。

## 移除的 CGO 组件

### 1. PostScript Surface (ps_surface_cgo.go)
- **原实现**: 使用 Cairo C 库的 `cairo_ps_surface_create` 等函数
- **新实现**: `ps_surface_pure.go` - 纯 Go 实现，直接生成 PostScript 代码
- **优势**: 
  - 无需 Cairo C 库依赖
  - 跨平台编译更简单
  - 更容易调试和维护

### 2. HarfBuzz 文本整形 (harfbuzz_cgo.go)
- **原实现**: 使用 HarfBuzz C 库进行文本整形
- **新实现**: 使用 `github.com/go-text/typesetting/shaping`
- **优势**:
  - 纯 Go 实现，无需外部依赖
  - 支持复杂文本布局（BiDi、连字等）
  - 性能优秀

### 3. 平台特定 Surface (cgo_platform_surface.go)
- **原实现**: 使用 CGO 调用平台 API
- **新实现**: `platform_surface_pure.go` - 预留接口，可使用 `golang.org/x/sys`
- **未来计划**:
  - Win32: 使用 `golang.org/x/sys/windows`
  - Quartz: 使用 `golang.org/x/sys/darwin`
  - X11: 使用 `golang.org/x/sys/unix`

## 依赖变化

### 移除的依赖
```
- cairo (C library)
- cairo-ps (C library)
- harfbuzz (C library)
- freetype (C library)
```

### 新增的纯 Go 依赖
```
+ github.com/go-text/typesetting - 文本整形和字体处理
+ github.com/llgcode/draw2d - PDF/SVG 输出支持
+ golang.org/x/image - 图像处理
```

## 编译优势

### 之前（使用 CGO）
```bash
# 需要安装 C 库
apt-get install libcairo2-dev libharfbuzz-dev

# 交叉编译困难
CGO_ENABLED=1 GOOS=windows go build  # 需要交叉编译工具链
```

### 现在（纯 Go）
```bash
# 无需外部依赖
go build

# 交叉编译简单
GOOS=windows go build
GOOS=darwin go build
GOOS=linux go build
```

## 性能对比

纯 Go 实现在大多数场景下性能与 CGO 版本相当：

- **文本整形**: `go-text/typesetting` 性能优秀，支持现代 OpenType 特性
- **图像处理**: Go 的 `image` 包经过高度优化
- **PostScript 生成**: 纯文本输出，性能不是瓶颈

## 功能完整性

| 功能 | CGO 版本 | 纯 Go 版本 | 状态 |
|------|---------|-----------|------|
| 图像 Surface | ✅ | ✅ | 完全支持 |
| PDF Surface | ✅ | ✅ | 完全支持 |
| SVG Surface | ✅ | ✅ | 完全支持 |
| PS Surface | ✅ | ✅ | 完全支持 |
| 文本整形 | ✅ | ✅ | 完全支持 |
| 字体渲染 | ✅ | ✅ | 完全支持 |
| Win32 Surface | ⚠️ | 🚧 | 计划中 |
| Quartz Surface | ⚠️ | 🚧 | 计划中 |
| X11 Surface | ⚠️ | 🚧 | 计划中 |

## 迁移指南

如果你之前使用了 CGO 版本，迁移非常简单：

### PostScript Surface
```go
// 之前和之后的 API 完全相同
surface := cairo.NewPSSurface("output.ps", 612, 792)
ctx := cairo.NewContext(surface)
// ... 绘图操作
surface.Destroy()
```

### 文本处理
```go
// API 保持不变，底层自动使用纯 Go 实现
ctx.SelectFontFace("sans-serif", cairo.FontSlantNormal, cairo.FontWeightNormal)
ctx.SetFontSize(12)
ctx.ShowText("Hello, World!")
```

## 未来计划

1. **平台 Surface 实现**: 使用 `golang.org/x/sys` 实现 Win32/Quartz/X11 支持
2. **性能优化**: 继续优化关键路径的性能
3. **测试覆盖**: 移植 Cairo 官方测试套件

## 总结

移除 CGO 后，go-cairo 成为了一个真正的纯 Go 库，具有以下优势：

✅ **简化部署**: 无需安装 C 库依赖  
✅ **跨平台编译**: 支持所有 Go 支持的平台  
✅ **更好的调试**: 纯 Go 代码更容易调试  
✅ **更快的编译**: 无需编译 C 代码  
✅ **更好的可移植性**: 单一二进制文件，无动态链接  

同时保持了与 Cairo C API 的兼容性和功能完整性。
