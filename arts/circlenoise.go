package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
)

type circleNoise struct {
	dotsN              int
	colorMin, colorMax int
}

type dot struct {
	x, y         float64
	prevx, prevy float64
	theta        float64
	count        int
	cx, cy       float64
}

func NewCircleNoise(dotsN, colorMin, colorMax int) *circleNoise {
	return &circleNoise{
		dotsN:    dotsN,
		colorMin: colorMin,
		colorMax: colorMax,
	}
}

// Generative draws a circle with perlin noise.
func (cn *circleNoise) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	ctex.SetLineWidth(2.0)
	ctex.SetColor(common.Black)
	radius := float64(c.Width()) * 0.8 / 2
	ctex.DrawCircle(float64(c.Width()/2), float64(c.Height()/2), radius)
	ctex.Stroke()
	//ctex.Clip()

	var factor = 0.008
	noise := common.NewPerlinNoise()

	dots := make([]dot, 0)
	for i := 0; i < cn.dotsN; i++ {
		theta := rand.Float64() * math.Pi * 2
		x, y := float64(c.Width())/2+math.Sin(theta)*radius, float64(c.Height())/2+math.Cos(theta)*radius
		dots = append(dots, dot{
			theta: theta,
			x:     x,
			y:     y,
			prevx: x,
			prevy: y,
			count: 0,
		})
	}

	for j := 0; j < c.Opts().NIters(); j++ {
		for i, _ := range dots {
			n := noise.Noise2D(dots[i].x*factor, dots[i].y*factor)
			nx, ny := math.Cos(n*math.Pi*2+float64(dots[i].count)*math.Pi)*2, math.Sin(n*math.Pi*2+float64(dots[i].count)*math.Pi)*2
			dots[i].prevx, dots[i].prevy = dots[i].x, dots[i].y
			dots[i].x, dots[i].y = dots[i].x+nx, dots[i].y+ny
			hsv := common.HSV{
				H: int(common.Remap(n, 0, 1, float64(cn.colorMin), float64(cn.colorMax))),
				S: 100,
				V: 20,
			}
			if common.Distance(float64(c.Width())/2, float64(c.Height())/2, dots[i].x, dots[i].y) > radius+2 {
				dots[i].count += 1
			}

			if common.Distance(float64(c.Width())/2, float64(c.Height())/2, dots[i].x, dots[i].y) < radius &&
				common.Distance(float64(c.Width())/2, float64(c.Height())/2, dots[i].prevx, dots[i].prevy) < radius {
				ctex.SetLineWidth(c.Opts().LineWidth())
				rgb := hsv.ToRGB(100, 100, 100)
				rgb.A = uint8(c.Opts().Alpha())
				ctex.SetColor(rgb)
				ctex.DrawLine(dots[i].prevx, dots[i].prevy, dots[i].x, dots[i].y)
				ctex.Stroke()
			}

		}
	}

}
