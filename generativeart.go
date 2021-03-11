package generativeart

import (
	"github.com/jdxyw/generativeart/common"
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
	img           *image.RGBA
	opts          Options
}

type Options struct {
	background  color.RGBA
	foreground  color.RGBA
	lineColor   color.RGBA
	lineWidth   float64
	colorSchema []color.RGBA
	nIters      int
	alpha       int
}

// NewCanva returns a canva.
func NewCanva(h, w int) *canva {
	return &canva{
		height: h,
		width:  w,
		img:    image.NewRGBA(image.Rect(0, 0, h, w)),
		// Set some defaults value
		opts: Options{
			background:  common.Azure,
			foreground:  common.MistyRose,
			lineColor:   common.Tomato,
			lineWidth:   3,
			colorSchema: common.Plasma,
			nIters:      20,
			alpha:       255,
		},
	}
}

func (c *canva) SetOptions(opts Options) {
	c.opts = opts
}

func (c *canva) SetBackground(rgba color.RGBA) {
	c.opts.background = rgba
}

func (c *canva) SetForeground(rgba color.RGBA) {
	c.opts.foreground = rgba
}

func (c *canva) SetColorSchema(rgbas []color.RGBA) {
	c.opts.colorSchema = rgbas
}

func (c *canva) SetLineColor(rgba color.RGBA) {
	c.opts.lineColor = rgba
}

func (c *canva) SetLineWidth(lw float64) {
	c.opts.lineWidth = lw
}

func (c *canva) SetIterations(nIters int) {
	c.opts.nIters = nIters
}

func (c *canva) SetAlpha(alpha int) {
	c.opts.alpha = alpha
}

func (c *canva) Draw(e Engine) {
	e.Generative(c)
}

// FillBackground fills the background of the canva.
func (c *canva) FillBackground() {
	draw.Draw(c.img, c.img.Bounds(), &image.Uniform{c.opts.background}, image.ZP, draw.Src)
}

// ToPng saves the image to local with PNG format.
func (c *canva) ToPNG(fpath string) error {
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
func (c *canva) ToJPEG(path string) error {
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
