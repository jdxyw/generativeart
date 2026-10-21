package generativeart

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"

	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type Engine interface {
	Generative(c *Canva)
}

type Canva struct {
	height, width int
	img           *image.RGBA
	opts          Options
	rng           *rand.Rand
	noise         *common.NoiseEngine
}

func (c *Canva) Opts() Options {
	return c.opts
}

func (c *Canva) Img() *image.RGBA {
	return c.img
}

func (c *Canva) Width() int {
	return c.width
}

func (c *Canva) Height() int {
	return c.height
}

type Options struct {
	background  color.RGBA
	foreground  color.RGBA
	lineColor   color.RGBA
	lineWidth   float64
	colorSchema []color.RGBA
	nIters      int
	alpha       int
	seed        int64
}

func (o Options) Seed() int64 {
	return o.seed
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
	defaultSeed := int64(20261021)
	return &Canva{
		height: h,
		width:  w,
		img:    image.NewRGBA(image.Rect(0, 0, w, h)),
		rng:    rand.New(rand.NewSource(defaultSeed)),
		// Set some defaults value
		opts: Options{
			background:  common.Azure,
			foreground:  common.MistyRose,
			lineColor:   common.Tomato,
			lineWidth:   3,
			colorSchema: common.Youthful,
			nIters:      20,
			alpha:       255,
			seed:        defaultSeed,
		},
	}
}

// SetSeed sets a deterministic random seed for this canvas, updating its local
// *rand.Rand, NoiseEngine, and the package-level math/rand seed for legacy engines.
func (c *Canva) SetSeed(seed int64) {
	c.opts.seed = seed
	c.rng = rand.New(rand.NewSource(seed))
	c.noise = common.NewNoiseEngine(seed)
	rand.Seed(seed)
}

// Seed returns the deterministic seed configured on this canvas.
func (c *Canva) Seed() int64 {
	return c.opts.seed
}

// Rand returns the canvas-scoped deterministic *rand.Rand generator.
func (c *Canva) Rand() *rand.Rand {
	if c.rng == nil {
		c.rng = rand.New(rand.NewSource(c.opts.seed))
	}
	return c.rng
}

// Noise returns the canvas-scoped deterministic *common.NoiseEngine.
func (c *Canva) Noise() *common.NoiseEngine {
	if c.noise == nil {
		c.noise = common.NewNoiseEngine(c.opts.seed)
	}
	return c.noise
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

func (c *Canva) DrawTimelapse(e Engine) {

}

// FillBackground fills the background of the Canva.
func (c *Canva) FillBackground() {
	draw.Draw(c.Img(), c.Img().Bounds(), &image.Uniform{c.Opts().Background()}, image.ZP, draw.Src)
}

// NewRenderCanvas creates a render.Canvas initialized from c.Img().
func (c *Canva) NewRenderCanvas() *render.Canvas {
	return render.NewCanvasFromRGBA(c.Img())
}

// MergeRenderCanvas resolves rc back into c.Img().
func (c *Canva) MergeRenderCanvas(rc *render.Canvas) {
	if rc == nil {
		return
	}
	c.img = rc.ResolveToRGBA(c.Img())
}

// MergeRenderCanvasToneMapped resolves rc back into c.Img() using ACES filmic tone mapping.
func (c *Canva) MergeRenderCanvasToneMapped(rc *render.Canvas, exposure, gamma float64) {
	if rc == nil {
		return
	}
	c.img = rc.ResolveToneMapped(c.Img(), exposure, gamma)
}

// Blend applies common.Blend onto c.Img().
func (c *Canva) Blend(src *image.RGBA, mode common.BlendMode) {
	if src == nil || c.img == nil {
		return
	}
	blended := common.Blend(src, c.img, mode)
	draw.Draw(c.img, c.img.Bounds(), blended, blended.Bounds().Min, draw.Src)
}

// ToPng saves the image to local with PNG format.
func (c *Canva) ToPNG(fpath string) error {
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	if err = png.Encode(f, c.Img()); err != nil {
		f.Close()
		return err
	}

	if err = f.Close(); err != nil {
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

// ToBytes returns the image as a jpeg-encoded []byte
func (c *Canva) ToBytes() ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, c.Img(), nil); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
