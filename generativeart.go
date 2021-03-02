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
	opts          Options
}

type Options struct {
	background       color.RGBA
	foreground       color.RGBA
	lineColor        color.RGBA
	lineWidth        float64
	colorSchema      []color.RGBA
	isPolarCoodinate bool
	step             int
	nIters           int
	alpha            int
	rectLenSide      int
	radius           float64
	decay            float64
}

// NewCanva returns a canva.
func NewCanva(h, w int, x, y float64) *canva {
	return &canva{
		height: h,
		width:  w,
		xaixs:  x,
		yaixs:  y,
		img:    image.NewRGBA(image.Rect(0, 0, h, w)),
		// Set some defaults value
		opts: Options{
			background:       Azure,
			foreground:       MistyRose,
			lineColor:        Tomato,
			lineWidth:        3,
			colorSchema:      Plasma,
			isPolarCoodinate: false,
			step:             24,
			nIters:           20,
			alpha:            30,
			rectLenSide:      10,
			radius:           1.0,
			decay:            0.2,
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

func (c *canva) SetPolarCoodinate() {
	c.opts.isPolarCoodinate = true
}

func (c *canva) SetStep(step int) {
	c.opts.step = step
}

func (c *canva) SetIterations(nIters int) {
	c.opts.nIters = nIters
}

func (c *canva) SetAlpha(alpha int) {
	c.opts.alpha = alpha
}

func (c *canva) SetRectLenSide(l int) {
	c.opts.rectLenSide = l
}

func (c *canva) SetRadius(r float64) {
	c.opts.radius = r
}

func (c *canva) SetDecay(d float64) {
	c.opts.decay = d
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
