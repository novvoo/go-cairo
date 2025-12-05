module go-cairo

go 1.21

require (
	// CGo dependencies (for future use, e.g., pixman, FreeType)
// //go:build cgo
// // #cgo pkg-config: cairo pixman-1 >= 0.40
// // #cgo CFLAGS: -DCAIRO_HAS_C11_ATOMIC
// // 图像处理相关依赖
	golang.org/x/image v0.18.0
)