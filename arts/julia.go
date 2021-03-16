package arts

import (
	"github.com/jdxyw/generativeart"
	"math/cmplx"
)

// GenFunc defines a func type used by julia set.
type GenFunc func(complex128) complex128

type julia struct {
	fn           GenFunc
	maxz         float64
	xaixs, yaixs float64
}

func NewJulia(formula GenFunc, maxz, xaixs, yaixs float64) *julia {
	return &julia{
		fn:    formula,
		maxz:  maxz,
		xaixs: xaixs,
		yaixs: yaixs,
	}
}

// Generative draws a julia set.
func (j *julia) Generative(c *generativeart.Canva) {
	n := len(c.Opts().ColorSchema())
	if n > 255 {
		n = 255
	}

	for i := 0; i <= c.Width(); i++ {
		for k := 0; k <= c.Height(); k++ {
			nit := 0
			z := complex(float64(i)/float64(c.Width())*2.0*j.yaixs-j.yaixs, float64(k)/float64(c.Height())*2.0*j.yaixs-j.yaixs)

			for cmplx.Abs(z) <= j.maxz && nit < c.Opts().NIters() {
				z = j.fn(z)
				nit += 1
			}
			idx := uint8(nit*255/c.Opts().NIters()) % uint8(n)
			c.Img().Set(i, k, c.Opts().ColorSchema()[idx])
		}
	}
}
