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

	for i := 0; i <src.Bounds().Max.X; i++ {
		for j :=0; j<src.Bounds().Max.Y; j++ {
			switch mode {
			case Add:
				if compareColor(src.At(i, j), Black) {
					img.Set(i, j, dest.At(i, j))
				} else {
					img.SetRGBA(i, j, add(src.At(i, j), dest.At(i, j)))
				}
			}
		}
	}
	return img
}

func compareColor(src, dst color.Color) bool {
	sr, sg, sb ,sa := src.RGBA()
	dr, dg, db, da := dst.RGBA()

	if sr == dr && sg == dg && sb == db && sa == da {
		return true
	}
	return false
}

func add(srcC, dstC color.Color) color.RGBA {
	c := color.RGBA{}
	sr, sg, sb ,sa := srcC.RGBA()
	dr, dg, db, da := dstC.RGBA()

	//aSrc := float64(sa)/255.0
	//aDst := 1.0 - aSrc

	//c.R = uint8(float64(sr) * aSrc + float64(dr)*aDst)
	//c.G = uint8(float64(sg) * aSrc + float64(dg)*aDst)
	//c.B = uint8(float64(sb) * aSrc + float64(db)*aDst)
	//c.A = uint8(float64(sa) * aSrc + float64(da)*aDst)

	c.R = uint8(ConstrainInt(int(uint32(float64(sr)*0.05)+dr), 0, 255))
	c.G = uint8(ConstrainInt(int(uint32(float64(sg)*0.05)+dg), 0, 255))
	c.B = uint8(ConstrainInt(int(uint32(float64(sb)*0.05)+db), 0, 255))
	c.A = uint8(ConstrainInt(int(uint32(float64(sa)*0.9)+da), 0, 255))
	//c.A = uint8(float64(sa) * aSrc + float64(da)*aDst)

	return c
}