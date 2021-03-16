package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
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
	noise := common.NewPerlinNoise()
	for i := 0; i < d.dotsN; i++ {
		x := common.RandomRangeFloat64(-0.1, 1.1) * float64(c.Width())
		y := common.RandomRangeFloat64(-0.1, 1.1) * float64(c.Height())

		num := common.RandomRangeFloat64(100, 1000)
		r := rand.Float64() * float64(c.Width()) * 0.15 * rand.Float64()
		ind := common.RandomRangeFloat64(1, 8)

		ctex.Push()
		ctex.Translate(x, y)
		ctex.Rotate(float64(rand.Intn(8)) * math.Pi / 4)
		rand.Shuffle(len(c.Opts().ColorSchema()), func(i, j int) {
			c.Opts().ColorSchema()[i], c.Opts().ColorSchema()[j] = c.Opts().ColorSchema()[j], c.Opts().ColorSchema()[i]
		})
		for j := 0.0; j < num; j += ind {
			s := float64(c.Width()) * 0.15 * common.RandomRangeFloat64(0, common.RandomRangeFloat64(0, common.RandomRangeFloat64(0, common.RandomRangeFloat64(0, common.RandomRangeFloat64(0, common.RandomRangeFloat64(0, rand.Float64()))))))
			ci := int(float64(len(c.Opts().ColorSchema())) * noise.Noise3D(j*0.01, x, y))
			ctex.SetColor(c.Opts().ColorSchema()[ci])
			ctex.DrawCircle(j, r*math.Sin(j*0.05), s*2/3)
			ctex.Fill()
		}
		ctex.Pop()
	}
}
