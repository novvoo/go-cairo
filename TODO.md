# go-cairo TODO List

This file tracks the remaining tasks for improving the go-cairo library and achieving better compatibility with the official Cairo API.

## High Priority

- [ ] **Surface Backend Extension:** Implement Xlib/XCB/Win32 surfaces using Go syscalls or third-party libraries (e.g., `golang.org/x/sys`). This is a major undertaking and requires platform-specific code.
- [ ] **Script Surface:** Implement `cairo_script_surface_create` to serialize drawing operations to JSON for debugging and replay.
- [ ] **Pixman-like Operations:** Implement pixman-like pixel operations using the `image` package to match native Cairo 1.18+ behavior.

## Medium Priority

- [ ] **Filters/Blur:** Implement `CAIRO_FILTER_GAUSSIAN` using a library like `gonum/mat` for convolution.
- [ ] **Advanced Text Layout:** Integrate `golang.org/x/image/font` more deeply to simulate FreeType and handle BiDi/RTL text.
- [ ] **Full Blend Implementation:** Implement the full set of `CAIRO_OPERATOR_*` blend modes with pixel-level shaders.

## Low Priority

- [ ] **Testing:** Port the official Cairo test suite (`test/`) to Go and use PNG diffing for validation.
- [ ] **Performance Benchmarking:** Add `go test -bench` benchmarks and compare against native Cairo performance.
- [ ] **Build System:** Create a `mage`-based build system to automate tasks and API diffing.
- [ ] **Documentation:** Expand godoc with an API mapping table and more examples.
