package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"image"
	"image/color"
	"image/draw"
	"math"
)

type solarFlare struct {
}

func NewSolarFlare() *solarFlare {
	return &solarFlare{}
}

// Generative draws a solar flare images.
func (o *solarFlare) Generative(c *generativeart.Canva) {
	var xOffset, yOffset float64
	var offsetInc = 0.006
	var inc = 1.0
	var r = 1.0
	var m = 1.005
	noise := common.NewPerlinNoise()

	for r < 200 {
		for i := 0; i < 10; i++ {
			nPoints := int(2 * math.Pi * r)
			nPoints = common.MinInt(nPoints, 500)

			img := image.NewRGBA(image.Rect(0, 0, c.Width(), c.Height()))
			draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)
			ctex := gg.NewContextForRGBA(img)

			ctex.Push()
			ctex.Translate(float64(c.Width()/2), float64(c.Height()/2))
			ctex.SetLineWidth(1.0)
			ctex.SetColor(c.Opts().LineColor())

			for j := 0.0; j < float64(nPoints+1); j += 1.0 {
				a := j / float64(nPoints) * math.Pi * 2
				px := math.Cos(a)
				py := math.Sin(a)
				n := noise.Noise2D(xOffset+px*inc, yOffset+py*inc) * r
				px *= n
				py *= n
				ctex.LineTo(px, py)
			}
			ctex.Stroke()
			ctex.Pop()

			//c.Img() = common.Blend(img, c.Img(), common.Add)
			xOffset += offsetInc
			yOffset += offsetInc
			r *= m
		}
	}
}
