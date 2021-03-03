package generativeart

import (
	"github.com/fogleman/gg"
	"math/rand"
)

type silkSky struct {
	circleNum int
	sunRadius float64
}

// NewSilkSky returns a silkSky object.
func NewSilkSky(circleNum int, sunRadius float64) *silkSky {
	return &silkSky{
		circleNum: circleNum,
		sunRadius: sunRadius,
	}
}

// Generative draws a silk sky image.
func (s *silkSky) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)
	xm := float64(rand.Intn(c.width/5)) + float64(c.width*4/5-c.width/5)
	ym := float64(rand.Intn(c.height/5)) + float64(c.height*4/5-c.height/5)

	mh := s.circleNum*2 + 2
	ms := s.circleNum*2 + 50
	mv := 100

	for i := 0; i < s.circleNum; i++ {
		for j := 0; j < s.circleNum; j++ {
			hsv := HSV{
				H: s.circleNum + j,
				S: i + 50,
				V: 70,
			}
			rgba := hsv.ToRGB(mh, ms, mv)
			xn := (float64(i) + 0.5) * float64(c.width) / float64(s.circleNum)
			yn := (float64(j) + 0.5) * float64(c.height) / float64(s.circleNum)
			ctex.SetRGBA255(int(rgba.R), int(rgba.G), int(rgba.B), c.opts.alpha)
			r := Distance(xn, yn, xm, ym)
			ctex.DrawEllipse(xn, yn, r-s.sunRadius/2, r-s.sunRadius/2)
			ctex.Fill()
		}
	}
}
