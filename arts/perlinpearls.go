package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
)

type perlinPearls struct {
	circleN            int
	dotsN              int
	colorMin, colorMax int
}

type circles struct {
	x, y               float64
	radius             float64
	colorMin, colorMax int
}

func NewPerlinPerls(circleN, dotsN, colorMin, colorMax int) *perlinPearls {
	return &perlinPearls{
		circleN:  circleN,
		dotsN:    dotsN,
		colorMin: colorMin,
		colorMax: colorMax,
	}
}

// Generative draws a circle with perlin noise.
func (pp *perlinPearls) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	ctex.SetLineWidth(0.5)
	ctex.SetColor(common.Black)

	cs := make([]circles, 0)

	for len(cs) < pp.circleN {
		c := circles{
			x:        common.RandomRangeFloat64(100, float64(c.Width())-50),
			y:        common.RandomRangeFloat64(100, float64(c.Height())-50),
			radius:   common.RandomRangeFloat64(20, 100),
			colorMin: pp.colorMin,
			colorMax: pp.colorMax,
		}
		var overlapping bool
		for _, cl := range cs {
			d := common.Distance(c.x, c.y, cl.x, cl.y)
			if d < c.radius+cl.radius {
				overlapping = true
				break
			}
		}

		if overlapping == false {
			cs = append(cs, c)
		}

	}

	ds := make([][]dot, 0)

	for i := 0; i < pp.circleN; i++ {
		dots := make([]dot, 0)
		for j := 0; j < pp.dotsN; j++ {
			theta := rand.Float64() * math.Pi * 2
			//x, y := float64(c.Width())/2+math.Sin(theta)*radius, float64(c.Height())/2+math.Cos(theta)*radius
			dots = append(dots, dot{
				theta: theta,
				cx:    cs[i].x,
				cy:    cs[i].y,
				x:     cs[i].x + math.Sin(theta)*cs[i].radius,
				y:     cs[i].y + math.Cos(theta)*cs[i].radius,
				prevx: cs[i].x + math.Sin(theta)*cs[i].radius,
				prevy: cs[i].y + math.Cos(theta)*cs[i].radius,
				count: 0,
			})
		}

		ds = append(ds, dots)
	}

	noise := common.NewPerlinNoise()

	for i := 0; i < pp.circleN; i++ {
		ctex.SetLineWidth(0.5)
		ctex.SetColor(common.Black)
		ctex.DrawCircle(cs[i].x, cs[i].y, cs[i].radius)
		ctex.Stroke()

		var factor = 0.008
		for j := 0; j < c.Opts().NIters(); j++ {
			for k := range ds[i] {
				n := noise.Noise2D(ds[i][k].x*factor, ds[i][k].y*factor)
				nx, ny := math.Cos(n*math.Pi*2+float64(ds[i][k].count)*math.Pi)*2, math.Sin(n*math.Pi*2+float64(ds[i][k].count)*math.Pi)*2
				ds[i][k].prevx, ds[i][k].prevy = ds[i][k].x, ds[i][k].y
				ds[i][k].x, ds[i][k].y = ds[i][k].x+nx, ds[i][k].y+ny

				hsv := common.HSV{
					H: int(common.Remap(n, 0, 1, float64(cs[i].colorMin), float64(cs[i].colorMax))),
					S: 100,
					V: 20,
				}

				if common.Distance(cs[i].x, cs[i].y, ds[i][k].x, ds[i][k].y) > cs[i].radius+1 {
					ds[i][k].count += 1
				}

				if common.Distance(cs[i].x, cs[i].y, ds[i][k].x, ds[i][k].y) < cs[i].radius &&
					common.Distance(cs[i].x, cs[i].y, ds[i][k].prevx, ds[i][k].prevy) < cs[i].radius {
					ctex.SetLineWidth(c.Opts().LineWidth())
					rgb := hsv.ToRGB(100, 100, 100)
					rgb.A = uint8(c.Opts().Alpha())
					ctex.SetColor(rgb)
					ctex.DrawLine(ds[i][k].prevx, ds[i][k].prevy, ds[i][k].x, ds[i][k].y)
					ctex.Stroke()
				}
			}
		}
	}
}
