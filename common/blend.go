package common

import (
	"image"
	"image/color"
)

type BlendMode int

const (
	Add BlendMode = iota
)

func Blend(src, dest *image.RGBA, mode BlendMode) *image.RGBA {
	img := image.NewRGBA(src.Bounds())

	for i := 0; i < src.Bounds().Max.X; i++ {
		for j := 0; j < src.Bounds().Max.Y; j++ {
			switch mode {
			case Add:
				if compareColor(src.RGBAAt(i, j), Black) {
					img.Set(i, j, dest.RGBAAt(i, j))
				} else {
					img.SetRGBA(i, j, add(src.RGBAAt(i, j), dest.RGBAAt(i, j)))
				}
			}
		}
	}
	return img
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
