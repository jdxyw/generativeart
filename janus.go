package generativeart

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart/common"
	"math"
)

type janus struct {
	n     int
	decay float64
}

// NewJanus returns a janus object
func NewJanus(n int, decay float64) *janus {
	return &janus{
		n:     n,
		decay: decay,
	}
}

// Generative draws a janus image
// TODO not finished.
func (j *janus) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)
	//ctex.SetColor(c.opts.foreground)
	s := 220.0
	r := 0.3

	for i := 0; i < j.n; i++ {
		//k := rand.Intn(len(c.opts.colorSchema))
		k := i
		ctex.Push()
		ctex.Translate(float64(c.width/2), float64(c.height/2))

		//theta += rand.Float64()*math.Pi/2
		theta := common.RandomRangeFloat64(math.Pi/4, 3*math.Pi/4)
		x1, y1 := math.Cos(theta)*r, math.Sin(theta)*r
		x2, y2 := -x1, -y1

		noise := common.RandomRangeFloat64(-math.Abs(y1), math.Abs(y1))
		y1 += noise
		y2 += noise

		//r = r - r*j.decay
		s = s * 0.836
		ctex.Scale(s, s)
		//r = r * 0.836
		ctex.DrawArc(x1, y1, 1.0, math.Pi*3/2+theta, math.Pi*5/2+theta)
		ctex.SetColor(c.opts.colorSchema[k])
		ctex.Fill()
		ctex.DrawArc(x2, y2, 1.0, math.Pi/2+theta, math.Pi*3/2+theta)
		ctex.SetColor(c.opts.colorSchema[k])
		ctex.Fill()
		ctex.Pop()
	}
}
