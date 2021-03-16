package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
)

type swirl struct {
	// These are some math parameters.
	// http://paulbourke.net/fractals/peterdejong/
	a, b, c, d, xaixs, yaixs float64
}

func NewSwirl(a, b, c, d, xaixs, yaixs float64) *swirl {
	return &swirl{
		a:     a,
		b:     b,
		c:     c,
		d:     d,
		xaixs: xaixs,
		yaixs: yaixs,
	}
}

// Generative draws a swirl image.
func (s *swirl) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	ctex.SetLineWidth(c.Opts().LineWidth())
	start := point{
		x: 1.0,
		y: 1.0,
	}
	cl := c.Opts().Foreground()

	for i := 0; i < c.Opts().NIters(); i++ {
		next := s.swirlTransform(start)
		x, y := common.ConvertCartesianToPixel(next.x, next.y, s.xaixs, s.yaixs, c.Height(), c.Width())
		c.Img().Set(x, y, cl)
		start = next
	}

	s.removeNoisy(c)
	s.removeNoisy(c)
}

func (s *swirl) swirlTransform(p point) point {
	return point{
		x: math.Sin(s.a*p.y) - math.Cos(s.b*p.x),
		y: math.Sin(s.c*p.x) - math.Cos(s.d*p.y),
	}
}

func (s *swirl) removeNoisy(c *generativeart.Canva) {
	for i := 1; i < c.Width()-1; i++ {
		for j := 1; j < c.Height()-1; j++ {
			if c.Img().At(i, j) == c.Opts().Background() {
				continue
			}

			var n int
			if c.Img().At(i+1, j) == c.Opts().Background() {
				n += 1
			}
			if c.Img().At(i+1, j+1) == c.Opts().Background() {
				n += 1
			}
			if c.Img().At(i, j+1) == c.Opts().Background() {
				n += 1
			}

			if c.Img().At(i-1, j) == c.Opts().Background() {
				n += 1
			}

			if c.Img().At(i-1, j+1) == c.Opts().Background() {
				n += 1
			}

			if c.Img().At(i-1, j-1) == c.Opts().Background() {
				n += 1
			}

			if c.Img().At(i+1, j-1) == c.Opts().Background() {
				n += 1
			}

			if c.Img().At(i, j-1) == c.Opts().Background() {
				n += 1
			}

			if n > 5 {
				c.Img().Set(i, j, c.Opts().Background())
			}

		}
	}
}
