package common

import (
	"image"
	"image/color"
	"math"
)

type BlendMode int

const (
	Add BlendMode = iota
	BlendOver
	BlendMultiply
	BlendScreen
	BlendOverlay
	BlendSoftLight
	BlendHardLight
	BlendDarken
	BlendLighten
	BlendDifference
	BlendGlow
)

func Blend(src, dest *image.RGBA, mode BlendMode) *image.RGBA {
	bounds := src.Bounds()
	img := image.NewRGBA(bounds)

	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			sc := src.RGBAAt(i, j)
			dc := dest.RGBAAt(i, j)
			switch mode {
			case Add:
				if compareColor(sc, Black) {
					img.SetRGBA(i, j, dc)
				} else {
					img.SetRGBA(i, j, add(sc, dc))
				}
			default:
				alpha := float64(sc.A) / 255.0
				img.SetRGBA(i, j, BlendColors(sc, dc, alpha, mode))
			}
		}
	}
	return img
}

// BlendColors blends src onto dst with the given alpha in [0, 1] and BlendMode.
func BlendColors(src, dst color.RGBA, alpha float64, mode BlendMode) color.RGBA {
	alpha = Clamp01(alpha)
	if alpha == 0 {
		return dst
	}
	sr, sg, sb := float64(src.R)/255.0, float64(src.G)/255.0, float64(src.B)/255.0
	dr, dg, db := float64(dst.R)/255.0, float64(dst.G)/255.0, float64(dst.B)/255.0

	var br, bg, bb float64
	switch mode {
	case Add:
		br, bg, bb = sr, sg, sb
	case BlendOver:
		br, bg, bb = sr, sg, sb
	case BlendMultiply:
		br, bg, bb = sr*dr, sg*dg, sb*db
	case BlendScreen:
		br = 1.0 - (1.0-sr)*(1.0-dr)
		bg = 1.0 - (1.0-sg)*(1.0-dg)
		bb = 1.0 - (1.0-sb)*(1.0-db)
	case BlendOverlay:
		br = blendOverlayChannel(sr, dr)
		bg = blendOverlayChannel(sg, dg)
		bb = blendOverlayChannel(sb, db)
	case BlendSoftLight:
		br = blendSoftLightChannel(sr, dr)
		bg = blendSoftLightChannel(sg, dg)
		bb = blendSoftLightChannel(sb, db)
	case BlendHardLight:
		br = blendOverlayChannel(dr, sr)
		bg = blendOverlayChannel(dg, sg)
		bb = blendOverlayChannel(db, sb)
	case BlendDarken:
		br = math.Min(sr, dr)
		bg = math.Min(sg, dg)
		bb = math.Min(sb, db)
	case BlendLighten:
		br = math.Max(sr, dr)
		bg = math.Max(sg, dg)
		bb = math.Max(sb, db)
	case BlendDifference:
		br = math.Abs(dr - sr)
		bg = math.Abs(dg - sg)
		bb = math.Abs(db - sb)
	case BlendGlow:
		outR := Clamp01(dr + sr*alpha)
		outG := Clamp01(dg + sg*alpha)
		outB := Clamp01(db + sb*alpha)
		outA := Clamp01(float64(dst.A)/255.0 + alpha*(1.0-float64(dst.A)/255.0))
		return color.RGBA{
			R: uint8(math.Round(outR * 255.0)),
			G: uint8(math.Round(outG * 255.0)),
			B: uint8(math.Round(outB * 255.0)),
			A: uint8(math.Round(outA * 255.0)),
		}
	default:
		br, bg, bb = sr, sg, sb
	}

	outR := Clamp01(dr*(1.0-alpha) + br*alpha)
	outG := Clamp01(dg*(1.0-alpha) + bg*alpha)
	outB := Clamp01(db*(1.0-alpha) + bb*alpha)
	da := float64(dst.A) / 255.0
	outA := Clamp01(alpha + da*(1.0-alpha))

	return color.RGBA{
		R: uint8(math.Round(outR * 255.0)),
		G: uint8(math.Round(outG * 255.0)),
		B: uint8(math.Round(outB * 255.0)),
		A: uint8(math.Round(outA * 255.0)),
	}
}

func blendOverlayChannel(s, d float64) float64 {
	if d < 0.5 {
		return 2.0 * s * d
	}
	return 1.0 - 2.0*(1.0-s)*(1.0-d)
}

func blendSoftLightChannel(s, d float64) float64 {
	if s <= 0.5 {
		return d - (1.0-2.0*s)*d*(1.0-d)
	}
	var g float64
	if d <= 0.25 {
		g = ((16.0*d-12.0)*d + 4.0) * d
	} else {
		g = math.Sqrt(d)
	}
	return d + (2.0*s-1.0)*(g-d)
}

func compareColor(src, dst color.RGBA) bool {
	sr, sg, sb, sa := src.R, src.G, src.B, src.A
	dr, dg, db, da := dst.R, dst.G, dst.B, dst.A

	if sr == dr && sg == dg && sb == db && sa == da {
		return true
	}
	return false
}

func add(srcC, dstC color.RGBA) color.RGBA {
	c := color.RGBA{}
	sr, sg, sb, sa := srcC.R, srcC.G, srcC.B, srcC.A
	dr, dg, db, da := dstC.R, dstC.G, dstC.B, dstC.A

	aSrc := float64(sa) / 255.0
	aDst := 1.0 - aSrc

	c.R = uint8(ConstrainInt(int(float64(sr)*aSrc+float64(dr)*aDst), 0, 255))
	c.G = uint8(ConstrainInt(int(float64(sg)*aSrc+float64(dg)*aDst), 0, 255))
	c.B = uint8(ConstrainInt(int(float64(sb)*aSrc+float64(db)*aDst), 0, 255))
	c.A = uint8(ConstrainInt(int(float64(sa)*aSrc+float64(da)*aDst), 0, 255))

	return c
}
