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
	Generative(c *Canva)
}

type Canva struct {
	height, width int
	img           *image.RGBA
	opts          Options
}

func (c *Canva) Opts() Options {
	return c.opts
}

func (c *Canva) Img() *image.RGBA {
	return c.Img()
}

func (c *Canva) Width() int {
	return c.Width()
}

func (c *Canva) Height() int {
	return c.Height()
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

func (o Options) Alpha() int {
	return o.alpha
}

func (o Options) NIters() int {
	return o.nIters
}

func (o Options) ColorSchema() []color.RGBA {
	return o.colorSchema
}

func (o Options) LineWidth() float64 {
	return o.lineWidth
}

func (o Options) LineColor() color.RGBA {
	return o.lineColor
}

func (o Options) Foreground() color.RGBA {
	return o.foreground
}

func (o Options) Background() color.RGBA {
	return o.background
}

// NewCanva returns a Canva.
func NewCanva(w, h int) *Canva {
	return &Canva{
		height: h,
		width:  w,
		img:    image.NewRGBA(image.Rect(0, 0, w, h)),
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

func (c *Canva) SetOptions(opts Options) {
	c.opts = opts
}

func (c *Canva) SetBackground(rgba color.RGBA) {
	c.opts.background = rgba
}

func (c *Canva) SetForeground(rgba color.RGBA) {
	c.opts.foreground = rgba
}

func (c *Canva) SetColorSchema(rgbas []color.RGBA) {
	c.opts.colorSchema = rgbas
}

func (c *Canva) SetLineColor(rgba color.RGBA) {
	c.opts.lineColor = rgba
}

func (c *Canva) SetLineWidth(lw float64) {
	c.opts.lineWidth = lw
}

func (c *Canva) SetIterations(nIters int) {
	c.opts.nIters = nIters
}

func (c *Canva) SetAlpha(alpha int) {
	c.opts.alpha = alpha
}

func (c *Canva) Draw(e Engine) {
	e.Generative(c)
}

// FillBackground fills the background of the Canva.
func (c *Canva) FillBackground() {
	draw.Draw(c.Img(), c.Img().Bounds(), &image.Uniform{c.Opts().Background()}, image.ZP, draw.Src)
}

// ToPng saves the image to local with PNG format.
func (c *Canva) ToPNG(fpath string) error {
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	if err := png.Encode(f, c.Img()); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// ToJpeg saves the image to local with Jpeg format.
func (c *Canva) ToJPEG(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := jpeg.Encode(f, c.Img(), nil); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}
