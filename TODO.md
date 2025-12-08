# go-cairo TODO List

This file tracks the remaining tasks for improving the go-cairo library and achieving better compatibility with the official Cairo API.

## High Priority

- [x] **CGO Removal:** Remove all CGO dependencies and implement pure Go alternatives ✅ COMPLETED
- [x] **Font API Update:** Update font.go to use the new go-text/typesetting v0.1.1 API ✅ COMPLETED
- [x] **Error Handling Uniformity:** Finalize mapping of all `Status` codes to the Go `error` interface, supporting new 1.18+ enumerations. ✅ COMPLETED
- [x] **Font Subsetting:** Implement logic for font subsetting (e.g., `cairo_scaled_font_get_glyphs`) for PDF/SVG output, including support for 1.18 color fonts. ✅ COMPLETED (GetGlyphs implemented with Harfbuzz shaping; color font support via go-text/typesetting)

- [ ] **Surface Backend Extension:** Implement Xlib/XCB/Win32 surfaces using Go syscalls or third-party libraries (e.g., `golang.org/x/sys`). This is a major undertaking and requires platform-specific code.
- [ ] **Pixman-like Operations:** Implement pixman-like pixel operations using the `image` package to match native Cairo 1.18+ behavior.

## Medium Priority

- [ ] **Build/Dependency:** Update `go.mod` to reflect `pixman >= 0.40` requirement (if CGo is used) and add cross-platform build tags.
- [ ] **Antialiasing/Precision:** Synchronize with 1.18's `ft-font-accuracy-new` by adding precision hints to `SetAntialias` (e.g., `AntialiasBest`).
- [ ] **Testing Coverage:** Add fuzz testing (`go test -fuzz`) and implement PNG output validation.
- [ ] **Performance Optimization:** Implement Go mutex wrappers to simulate Cairo's spinlock optimization.
- [ ] **Documentation/Compatibility Matrix:** Add a compatibility table to `README.md` listing supported Cairo versions.

- [ ] **Filters/Blur:** Implement `CAIRO_FILTER_GAUSSIAN` using a library like `gonum/mat` for convolution.
- [ ] **Advanced Text Layout:** Integrate `golang.org/x/image/font` more deeply to simulate FreeType and handle BiDi/RTL text.
- [ ] **Full Blend Implementation:** Implement the full set of `CAIRO_OPERATOR_*` blend modes with pixel-level shaders.

## Low Priority

- [ ] **Testing:** Port the official Cairo test suite (`test/`) to Go and use PNG diffing for validation.
- [ ] **Performance Benchmarking:** Add `go test -bench` benchmarks and compare against native Cairo performance.
- [ ] **Build System:** Create a `mage`-based build system to automate tasks and API diffing.
- [ ] **Documentation:** Expand godoc with an API mapping table and more examples.
