package generativeart

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
)

type pixelHole struct {
	dotN  int
	noise *common.PerlinNoise
}

// NewPixelHole returns a pixelHole object.
func NewPixelHole(dotN int) *pixelHole {
	return &pixelHole{
		dotN:  dotN,
		noise: common.NewPerlinNoise(),
	}
}

// Generative draws a pixel hole images.
func (p *pixelHole) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)
	for i := 0.0; i < float64(c.opts.nIters); i += 1.0 {
		ctex.Push()
		ctex.Translate(float64(c.width)/2, float64(c.height)/2)
		p.draw(ctex, c, i)
		ctex.Pop()
	}
}

func (p *pixelHole) draw(ctex *gg.Context, c *canva, frameCount float64) {
	ctex.SetLineWidth(2.0)
	c1 := int(frameCount/100.0) % len(c.opts.colorSchema)
	c2 := (int(frameCount/100.0) + 1) % len(c.opts.colorSchema)
	ratio := frameCount/100 - math.Floor(frameCount/100)
	cl := common.LerpColor(c.opts.colorSchema[c1], c.opts.colorSchema[c2], ratio)
	for i := 0.0; i < float64(p.dotN); i += 1.0 {
		ctex.Push()
		ctex.SetColor(cl)
		ctex.Rotate(frameCount/(50+10*math.Log(frameCount+1)) + i/20)
		dd := frameCount/(5+i) + frameCount/5 + math.Sin(i)*50
		ctex.Translate(common.RandomRangeFloat64(dd/2, dd), 0)
		x := p.noise.Noise2D(frameCount/50+i/50, 5000)*float64(c.width)/10 + rand.Float64()*float64(c.width)/20
		y := p.noise.Noise2D(frameCount/50+i/50, 10000)*float64(c.height)/10 + rand.Float64()*float64(c.height)/20

		rr := common.RandomRangeFloat64(1.0, 6-math.Log(frameCount+1)/10)
		ctex.DrawEllipse(x, y, rr, rr)
		ctex.Fill()
		ctex.Pop()
	}
}
