package generativeart

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
)

type Engine interface {
	Generative(c *canva)
}

type canva struct {
	height, width int
	xaixs, yaixs  float64
	img           *image.RGBA
}

// NewCanva returns a canva.
func NewCanva(h, w int, x, y float64) *canva {
	return &canva{
		height: h,
		width:  w,
		xaixs:  x,
		yaixs:  y,
		img:    image.NewRGBA(image.Rect(0, 0, h, w)),
	}
}

// FillBackgroud fills the backgroud of the canva.
func (c *canva) FillBackgroud(rgba color.RGBA) {
	draw.Draw(c.img, c.img.Bounds(), &image.Uniform{rgba}, image.ZP, draw.Src)
}

// ToPng saves the image to local with PNG format.
func (c *canva) ToPng(fpath string) error {
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	if err := png.Encode(f, c.img); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// ToJpeg saves the image to local with Jpeg format.
func (c *canva) ToJpeg(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := jpeg.Encode(f, c.img, nil); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
