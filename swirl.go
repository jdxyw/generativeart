package generativeart

import (
	"github.com/fogleman/gg"
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
func (s *swirl) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)
	ctex.SetLineWidth(c.opts.lineWidth)
	start := point{
		x: 1.0,
		y: 1.0,
	}
	cl := c.opts.foreground

	for i := 0; i < c.opts.nIters; i++ {
		next := s.swirlTransform(start)
		x, y := ConvertCartesianToPixel(next.x, next.y, s.xaixs, s.yaixs, c.height, c.width)
		c.img.Set(x, y, cl)
		start = next
	}

	s.removeNoisy(c)
}

func (s *swirl) swirlTransform(p point) point {
	return point{
		x: math.Sin(s.a*p.y) - math.Cos(s.b*p.x),
		y: math.Sin(s.c*p.x) - math.Cos(s.d*p.y),
	}
}

func (s *swirl) removeNoisy(c *canva) {
	for i := 1; i < c.width-1; i++ {
		for j := 1; j < c.height-1; j++ {
			if c.img.At(i, j) == c.opts.background {
				continue
			}

			var n int
			if c.img.At(i+1, j) == c.opts.background {
				n += 1
			}
			if c.img.At(i+1, j+1) == c.opts.background {
				n += 1
			}
			if c.img.At(i, j+1) == c.opts.background {
				n += 1
			}

			if c.img.At(i-1, j) == c.opts.background {
				n += 1
			}

			if c.img.At(i-1, j+1) == c.opts.background {
				n += 1
			}

			if c.img.At(i-1, j-1) == c.opts.background {
				n += 1
			}

			if c.img.At(i+1, j-1) == c.opts.background {
				n += 1
			}

			if c.img.At(i, j-1) == c.opts.background {
				n += 1
			}

			if n > 5 {
				c.img.Set(i, j, c.opts.background)
			}

		}
	}
}
