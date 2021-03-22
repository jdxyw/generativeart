package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math/rand"
)

type colorCanva struct {
	seg float64
}

func NewColorCanve(seg float64) *colorCanva {
	return &colorCanva{seg: seg}
}

// Generative returns a color canva image.
func (cc *colorCanva) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	ctex.SetLineWidth(c.Opts().LineWidth())

	rects := make([]*common.Rect, 0)
	w := float64(c.Width()) / cc.seg
	for i := 0.0; i < cc.seg; i += 1.0 {
		for j := 0.0; j < cc.seg; j += 1.0 {
			x := i*w + w/2
			y := j*w + w/2
			rects = append(rects, common.NewRect(x, y, w, w))
		}
	}

	rand.Shuffle(len(rects), func(i, j int) {
		rects[i], rects[j] = rects[j], rects[i]
	})

	ctex.Translate(float64(c.Width())/2, float64(c.Height())/2)
	ctex.Scale(0.6, 0.6)
	ctex.Translate(-float64(c.Width())/2, -float64(c.Height())/2)

	for i := 0; i < len(rects); i++ {
		cc.draw(c, ctex, rects[i])
		cc.draw(c, ctex, rects[i])
	}
}

func (cc *colorCanva) draw(c *generativeart.Canva, ctex *gg.Context, rect *common.Rect) {
	rnd := rand.Intn(4)
	ww := (rect.W() / 5) * float64(rand.Intn(int(cc.seg)*2)+1)
	hh := (rect.H() / 5) * float64(rand.Intn(int(cc.seg)*2)+1)

	switch rnd {
	case 0:
		ctex.DrawRectangle(rect.X()-ww/2+rect.W()/2, rect.Y()+hh/2+rect.H()/2, ww, hh)
	case 1:
		ctex.DrawRectangle(rect.X()-ww/2-rect.W()/2, rect.Y()+hh/2+rect.H()/2, ww, hh)
	case 2:
		ctex.DrawRectangle(rect.X()-ww/2+rect.W()/2, rect.Y()+hh/2-rect.H()/2, ww, hh)
	case 3:
		ctex.DrawRectangle(rect.X()-ww/2-rect.W()/2, rect.Y()+hh/2-rect.H()/2, ww, hh)
	}
	ctex.SetColor(common.Black)
	ctex.StrokePreserve()
	ctex.SetColor(c.Opts().ColorSchema()[rand.Intn(len(c.Opts().ColorSchema()))])
	ctex.Fill()
}
