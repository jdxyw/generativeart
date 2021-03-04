package generativeart

import "math/cmplx"

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
func (j *julia) Generative(c *canva) {
	n := len(c.opts.colorSchema)
	if n > 255 {
		n = 255
	}

	for i := 0; i <= c.width; i++ {
		for k := 0; k <= c.height; k++ {
			nit := 0
			z := complex(float64(i)/float64(c.width)*2.0*j.yaixs-j.yaixs, float64(k)/float64(c.height)*2.0*j.yaixs-j.yaixs)

			for cmplx.Abs(z) <= j.maxz && nit < c.opts.nIters {
				z = j.fn(z)
				nit += 1
			}
			idx := uint8(nit*255/c.opts.nIters) % uint8(n)
			c.img.Set(i, k, c.opts.colorSchema[idx])
		}
	}
}
