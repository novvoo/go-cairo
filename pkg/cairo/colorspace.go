package cairo

import "math"

// 颜色空间转换模块
// 支持 RGB, HSL, HSV, LAB, XYZ 等颜色空间

// RgbToHSL RGB 到 HSL 转换
func RgbToHSL(r, g, b float64) (h, s, l float64) {
	return rgbToHSL(r, g, b)
}

// rgb到HSL转换（内部）
func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	l = (max + min) / 2

	if max == min {
		h, s = 0, 0 // 灰色
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h /= 6
	}
	return
}

// HslToRGB HSL 到 RGB 转换
func HslToRGB(h, s, l float64) (r, g, b float64) {
	return hslToRGB(h, s, l)
}

// hsl到RGB转换（内部）
func hslToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		r, g, b = l, l, l // 灰色
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}
	return
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// RGB 到 HSV 转换
func rgbToHSV(r, g, b float64) (h, s, v float64) {
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	v = max

	d := max - min
	if max == 0 {
		s = 0
	} else {
		s = d / max
	}

	if max == min {
		h = 0 // 灰色
	} else {
		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h /= 6
	}
	return
}

// HSV 到 RGB 转换
func hsvToRGB(h, s, v float64) (r, g, b float64) {
	if s == 0 {
		r, g, b = v, v, v
		return
	}

	h *= 6
	i := math.Floor(h)
	f := h - i
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	switch int(i) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	case 5:
		r, g, b = v, p, q
	}
	return
}

// RGB 到 XYZ 转换 (使用 sRGB 色彩空间)
func rgbToXYZ(r, g, b float64) (x, y, z float64) {
	// sRGB 到线性 RGB
	r = srgbToLinear(r)
	g = srgbToLinear(g)
	b = srgbToLinear(b)

	// 使用 D65 白点的转换矩阵
	x = r*0.4124564 + g*0.3575761 + b*0.1804375
	y = r*0.2126729 + g*0.7151522 + b*0.0721750
	z = r*0.0193339 + g*0.1191920 + b*0.9503041
	return
}

// XYZ 到 RGB 转换
func xyzToRGB(x, y, z float64) (r, g, b float64) {
	// 使用 D65 白点的逆转换矩阵
	r = x*3.2404542 + y*-1.5371385 + z*-0.4985314
	g = x*-0.9692660 + y*1.8760108 + z*0.0415560
	b = x*0.0556434 + y*-0.2040259 + z*1.0572252

	// 线性 RGB 到 sRGB
	r = linearToSRGB(r)
	g = linearToSRGB(g)
	b = linearToSRGB(b)

	// 限制范围
	r = math.Max(0, math.Min(1, r))
	g = math.Max(0, math.Min(1, g))
	b = math.Max(0, math.Min(1, b))
	return
}

// sRGB 到线性 RGB
func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// 线性 RGB 到 sRGB
func linearToSRGB(c float64) float64 {
	if c <= 0.0031308 {
		return c * 12.92
	}
	return 1.055*math.Pow(c, 1/2.4) - 0.055
}

// XYZ 到 LAB 转换 (使用 D65 白点)
func xyzToLAB(x, y, z float64) (l, a, b float64) {
	// D65 白点
	const xn, yn, zn = 0.95047, 1.00000, 1.08883

	x = labF(x / xn)
	y = labF(y / yn)
	z = labF(z / zn)

	l = 116*y - 16
	a = 500 * (x - y)
	b = 200 * (y - z)
	return
}

// LAB 到 XYZ 转换
func labToXYZ(l, a, b float64) (x, y, z float64) {
	// D65 白点
	const xn, yn, zn = 0.95047, 1.00000, 1.08883

	fy := (l + 16) / 116
	fx := a/500 + fy
	fz := fy - b/200

	x = xn * labFInv(fx)
	y = yn * labFInv(fy)
	z = zn * labFInv(fz)
	return
}

func labF(t float64) float64 {
	const delta = 6.0 / 29.0
	if t > delta*delta*delta {
		return math.Pow(t, 1.0/3.0)
	}
	return t/(3*delta*delta) + 4.0/29.0
}

func labFInv(t float64) float64 {
	const delta = 6.0 / 29.0
	if t > delta {
		return t * t * t
	}
	return 3 * delta * delta * (t - 4.0/29.0)
}

// RGB 到 LAB 转换
func rgbToLAB(r, g, b float64) (l, a, bVal float64) {
	x, y, z := rgbToXYZ(r, g, b)
	return xyzToLAB(x, y, z)
}

// LAB 到 RGB 转换
func labToRGB(l, a, b float64) (r, g, bVal float64) {
	x, y, z := labToXYZ(l, a, b)
	return xyzToRGB(x, y, z)
}

// 颜色差异计算 (Delta E 2000)
func ColorDeltaE2000(l1, a1, b1, l2, a2, b2 float64) float64 {
	// 简化的 Delta E 2000 实现
	const kL, kC, kH = 1.0, 1.0, 1.0

	c1 := math.Sqrt(a1*a1 + b1*b1)
	c2 := math.Sqrt(a2*a2 + b2*b2)
	cBar := (c1 + c2) / 2

	g := 0.5 * (1 - math.Sqrt(math.Pow(cBar, 7)/(math.Pow(cBar, 7)+math.Pow(25, 7))))

	a1p := (1 + g) * a1
	a2p := (1 + g) * a2

	c1p := math.Sqrt(a1p*a1p + b1*b1)
	c2p := math.Sqrt(a2p*a2p + b2*b2)

	h1p := math.Atan2(b1, a1p)
	h2p := math.Atan2(b2, a2p)

	if h1p < 0 {
		h1p += 2 * math.Pi
	}
	if h2p < 0 {
		h2p += 2 * math.Pi
	}

	dL := l2 - l1
	dC := c2p - c1p
	dH := h2p - h1p

	if dH > math.Pi {
		dH -= 2 * math.Pi
	} else if dH < -math.Pi {
		dH += 2 * math.Pi
	}

	dH = 2 * math.Sqrt(c1p*c2p) * math.Sin(dH/2)

	lBar := (l1 + l2) / 2
	cBar = (c1p + c2p) / 2

	sL := 1 + (0.015*(lBar-50)*(lBar-50))/math.Sqrt(20+(lBar-50)*(lBar-50))
	sC := 1 + 0.045*cBar
	sH := 1 + 0.015*cBar

	deltaE := math.Sqrt(
		math.Pow(dL/(kL*sL), 2) +
			math.Pow(dC/(kC*sC), 2) +
			math.Pow(dH/(kH*sH), 2))

	return deltaE
}
