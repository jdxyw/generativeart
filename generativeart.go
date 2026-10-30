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
	height, width    int
	img              *image.RGBA
	opts             Options
	rng              *rand.Rand
	noise            *common.NoiseEngine
	lastRenderCanvas *render.Canvas
	hdrDirty         bool
	rgbaTouched      bool
	frames           []*image.RGBA
	lastRecipe       *RecipeMetadata
	postEffects      []render.PostEffect
}

func (c *Canva) Opts() Options {
	return c.opts
}

func (c *Canva) Img() *image.RGBA {
	if c.hdrDirty && c.lastRenderCanvas != nil {
		c.img = c.lastRenderCanvas.ResolveToRGBA(c.img)
		c.hdrDirty = false
	}
	c.rgbaTouched = true
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

// AddPostEffect appends one or more composable HDR PostEffect passes to this Canva.
// Registered post-effects run automatically after c.Draw(e) and after each frame in c.DrawTimelapseFrames.
func (c *Canva) AddPostEffect(effects ...render.PostEffect) *Canva {
	if c == nil {
		return nil
	}
	for _, eff := range effects {
		if eff != nil {
			c.postEffects = append(c.postEffects, eff)
		}
	}
	return c
}

// ClearPostEffects removes all registered HDR PostEffect passes from this Canva.
func (c *Canva) ClearPostEffects() {
	if c != nil {
		c.postEffects = nil
	}
}

// PostEffects returns a copy of the registered HDR PostEffect pipeline on this Canva.
func (c *Canva) PostEffects() []render.PostEffect {
	if c == nil || len(c.postEffects) == 0 {
		return nil
	}
	out := make([]render.PostEffect, len(c.postEffects))
	copy(out, c.postEffects)
	return out
}

// ApplyPostEffects executes the provided PostEffect passes (or c.PostEffects() if none are
// explicitly passed) directly on the Canva's 64-bit linear HDR buffer.
func (c *Canva) ApplyPostEffects(effects ...render.PostEffect) {
	if c == nil {
		return
	}
	list := effects
	if len(list) == 0 {
		list = c.postEffects
	}
	if len(list) == 0 {
		return
	}
	rc := c.NewRenderCanvas()
	rc.ApplyPostEffects(list...)
	c.MergeRenderCanvas(rc)
	if c.lastRecipe != nil {
		names := make([]string, 0, len(list))
		for _, eff := range list {
			if eff != nil {
				names = append(names, eff.Name())
			}
		}
		c.lastRecipe.PostEffects = names
	}
}

func (c *Canva) Draw(e Engine) {
	c.lastRecipe = captureEngineRecipe(c, e)
	e.Generative(c)
	if len(c.postEffects) > 0 {
		c.ApplyPostEffects(c.postEffects...)
	}
}

// DrawTimelapse renders a multi-frame timelapse or seamless loop sequence (using
// c.Opts().NIters() frames, defaulting to 12) into c.Frames(), which can then be
// exported as an animated GIF via c.ToGIF(fpath, delayCs).
func (c *Canva) DrawTimelapse(e Engine) {
	frames := c.opts.nIters
	if frames <= 0 {
		frames = 12
	}
	c.DrawTimelapseFrames(e, frames)
}

// FillBackground fills the background of the Canva in both RGBA and persistent HDR buffers.
func (c *Canva) FillBackground() {
	draw.Draw(c.img, c.img.Bounds(), &image.Uniform{c.Opts().Background()}, image.ZP, draw.Src)
	if c.lastRenderCanvas == nil || c.lastRenderCanvas.Width != c.width || c.lastRenderCanvas.Height != c.height {
		c.lastRenderCanvas = render.NewCanvas(c.width, c.height)
	}
	c.lastRenderCanvas.FillBackground(c.Opts().Background())
	c.hdrDirty = false
	c.rgbaTouched = false
}

// NewRenderCanvas returns the persistent 64-bit linear HDR render.Canvas for this Canva.
// If no legacy RGBA operations touched c.Img() since the last FillBackground or MergeRenderCanvas,
// the persistent HDR buffer is returned directly without re-decoding 8-bit RGBA pixels.
func (c *Canva) NewRenderCanvas() *render.Canvas {
	if c.lastRenderCanvas == nil || c.lastRenderCanvas.Width != c.width || c.lastRenderCanvas.Height != c.height {
		c.lastRenderCanvas = render.NewCanvasFromRGBA(c.img)
		c.rgbaTouched = false
		return c.lastRenderCanvas
	}
	if c.rgbaTouched {
		c.lastRenderCanvas = render.NewCanvasFromRGBA(c.img)
		c.rgbaTouched = false
	}
	return c.lastRenderCanvas
}

// MergeRenderCanvas records rc as the active HDR buffer and marks RGBA for lazy resolution
// when c.Img(), ToPNG(), ToJPEG(), or ToGIF() is invoked.
func (c *Canva) MergeRenderCanvas(rc *render.Canvas) {
	if rc == nil {
		return
	}
	c.lastRenderCanvas = rc
	c.hdrDirty = true
	c.rgbaTouched = false
}

// MergeRenderCanvasToneMapped resolves rc back into c.Img() using ACES filmic tone mapping.
func (c *Canva) MergeRenderCanvasToneMapped(rc *render.Canvas, exposure, gamma float64) {
	if rc == nil {
		return
	}
	c.lastRenderCanvas = rc
	c.img = rc.ResolveToneMapped(c.img, exposure, gamma)
	c.hdrDirty = false
	c.rgbaTouched = false
}

// LastRenderCanvas returns the most recent *render.Canvas used by this Canva.
func (c *Canva) LastRenderCanvas() *render.Canvas {
	return c.lastRenderCanvas
}

// Blend applies common.Blend onto c.Img().
func (c *Canva) Blend(src *image.RGBA, mode common.BlendMode) {
	dst := c.Img()
	if src == nil || dst == nil {
		return
	}
	blended := common.Blend(src, dst, mode)
	draw.Draw(dst, dst.Bounds(), blended, blended.Bounds().Min, draw.Src)
}

// ToPNG saves the image to local with PNG format and embeds the deterministic
// GenerativeArt-Recipe JSON metadata inside a standard PNG tEXt chunk.
func (c *Canva) ToPNG(fpath string) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, c.Img()); err != nil {
		return err
	}
	rec := c.lastRecipe
	if rec == nil {
		rec = captureEngineRecipe(c, nil)
	}
	pngBytes := buf.Bytes()
	if withMeta, err := InjectPNGRecipe(pngBytes, rec); err == nil {
		pngBytes = withMeta
	}
	return os.WriteFile(fpath, pngBytes, 0644)
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
