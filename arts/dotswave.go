package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"image/color"
	"math"
)

type dotsWave struct {
	dotsN int
}

// NewDotsWave returns a dotsWave object.
func NewDotsWave(dotsN int) *dotsWave {
	return &dotsWave{
		dotsN: dotsN,
	}
}

// Generative draws a dots wave images.
func (d *dotsWave) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	noise := common.NewPerlinNoiseWithSeed(c.Seed())
	rng := c.Rand()
	colors := append([]color.RGBA(nil), c.Opts().ColorSchema()...)
	for i := 0; i < d.dotsN; i++ {
		x := common.RandomRangeFloat64From(rng, -0.1, 1.1) * float64(c.Width())
		y := common.RandomRangeFloat64From(rng, -0.1, 1.1) * float64(c.Height())

		num := common.RandomRangeFloat64From(rng, 100, 1000)
		r := rng.Float64() * float64(c.Width()) * 0.15 * rng.Float64()
		ind := common.RandomRangeFloat64From(rng, 1, 8)

		ctex.Push()
		ctex.Translate(x, y)
		ctex.Rotate(float64(rng.Intn(8)) * math.Pi / 4)
		rng.Shuffle(len(colors), func(i, j int) {
			colors[i], colors[j] = colors[j], colors[i]
		})
		for j := 0.0; j < num; j += ind {
			s := float64(c.Width()) * 0.15 * common.RandomRangeFloat64From(rng, 0, common.RandomRangeFloat64From(rng, 0, common.RandomRangeFloat64From(rng, 0, common.RandomRangeFloat64From(rng, 0, common.RandomRangeFloat64From(rng, 0, common.RandomRangeFloat64From(rng, 0, rng.Float64()))))))
			ci := int(float64(len(colors)) * noise.Noise3D(j*0.01, x, y))
			if ci < 0 {
				ci = 0
			} else if ci >= len(colors) {
				ci = len(colors) - 1
			}
			ctex.SetColor(colors[ci])
			ctex.DrawCircle(j, r*math.Sin(j*0.05), s*2/3)
			ctex.Fill()
		}
		ctex.Pop()
	}
}
