package generativeart

import (
	"github.com/fogleman/gg"
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

func (r *randCircle) newCircleSlice(cn, w, h int) []circle {
	var circles []circle

	for i := 0; i < cn; i++ {
		x := rand.Intn(w) + 1
		y := rand.Intn(h) + 1
		radius := float64(rand.Intn(int(r.minRadius))) + r.maxRadius - r.minRadius
		angle := rand.Float64() * math.Pi * 2.0
		step := r.minSteps + rand.Float64()*(r.maxSteps-r.minSteps)
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

func (r *randCircle) circleSliceUpdate(cs []circle, w, h int) []circle {
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
func (r *randCircle) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)

	for j := 0; j < c.opts.nIters; j++ {
		cn := rand.Intn(r.maxCircle) + int(r.maxCircle/3)
		circles := r.newCircleSlice(cn, c.width, c.height)

		for i := 0; i < r.maxStepsPerCircle; i++ {
			for _, c1 := range circles {
				for _, c2 := range circles {

					cl := c.opts.lineColor
					if r.isRandColor {
						cl = c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))]
					}

					if c1 == c2 {
						continue
					}

					distance := Distance(c1.x, c1.y, c2.x, c2.y)

					if distance <= c1.radius+c2.radius {
						cx := (c1.x + c2.x) / 2
						cy := (c1.y + c2.y) / 2

						ctex.SetRGBA255(int(cl.R), int(cl.G), int(cl.B), 30)
						ctex.SetLineWidth(c.opts.lineWidth)
						ctex.DrawEllipse(cx, cy, distance/2, distance/2)
						ctex.Stroke()
					}
				}
			}

			circles = r.circleSliceUpdate(circles, c.width, c.height)
		}
	}
}
