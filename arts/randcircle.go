package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
)

type circle struct {
	x, y   float64
	radius float64
	dx, dy float64
}

type randCircle struct {
	maxCircle         int
	maxStepsPerCircle int
	minSteps          float64
	maxSteps          float64
	minRadius         float64
	maxRadius         float64
	isRandColor       bool
}

func NewRandCicle(mc, msp int, minStep, maxStep, minr, maxr float64, isRandColor bool) *randCircle {
	return &randCircle{
		maxCircle:         mc,
		maxStepsPerCircle: msp,
		minSteps:          minStep,
		maxSteps:          maxStep,
		minRadius:         minr,
		maxRadius:         maxr,
		isRandColor:       isRandColor,
	}
}

func newCircleSlice(cn, w, h int, minStep, maxStep, minRadius, maxRadius float64) []circle {
	var circles []circle

	for i := 0; i < cn; i++ {
		x := rand.Intn(w) + 1
		y := rand.Intn(h) + 1
		radius := float64(rand.Intn(int(minRadius))) + maxRadius - minRadius
		angle := rand.Float64() * math.Pi * 2.0
		step := minStep + rand.Float64()*(maxStep-minStep)
		circles = append(circles, circle{
			x:      float64(x),
			y:      float64(y),
			radius: radius,
			dx:     step * math.Cos(angle),
			dy:     step * math.Sin(angle),
		})
	}

	return circles
}

func circleSliceUpdate(cs []circle, w, h int) []circle {
	var circles []circle

	for _, c := range cs {
		c.x += c.dx
		c.y += c.dy

		if c.x <= 0 {
			c.x = 0
			c.dx *= -1
		}

		if c.y <= 0 {
			c.y = 0
			c.dy *= -1
		}

		if c.x > float64(w) {
			c.x = float64(w)
			c.dx *= -1
		}

		if c.y > float64(h) {
			c.y = float64(h)
			c.dy *= -1
		}

		circles = append(circles, c)
	}

	return circles
}

// Generative draws a random circles image.
func (r *randCircle) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())

	for j := 0; j < c.Opts().NIters(); j++ {
		cn := rand.Intn(r.maxCircle) + int(r.maxCircle/3)
		circles := newCircleSlice(cn, c.Width(), c.Height(), r.minSteps, r.maxSteps, r.minRadius, r.maxRadius)

		for i := 0; i < r.maxStepsPerCircle; i++ {
			for _, c1 := range circles {
				for _, c2 := range circles {

					cl := c.Opts().LineColor()
					if r.isRandColor {
						cl = c.Opts().ColorSchema()[rand.Intn(len(c.Opts().ColorSchema()))]
					}

					if c1 == c2 {
						continue
					}

					distance := common.Distance(c1.x, c1.y, c2.x, c2.y)

					if distance <= c1.radius+c2.radius {
						cx := (c1.x + c2.x) / 2
						cy := (c1.y + c2.y) / 2

						ctex.SetRGBA255(int(cl.R), int(cl.G), int(cl.B), 30)
						ctex.SetLineWidth(c.Opts().LineWidth())
						ctex.DrawEllipse(cx, cy, distance/2, distance/2)
						ctex.Stroke()
					}
				}
			}

			circles = circleSliceUpdate(circles, c.Width(), c.Height())
		}
	}
}
