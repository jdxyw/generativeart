package main

import (
	"image"
	"image/color"
	"image/draw"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func renderFractalFlame(w, h int) *generativeart.Canva {
	rand.Seed(101)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 4, G: 4, B: 10, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Magma)
	c.Draw(arts.NewFractalFlame(1800000, 5))
	return c
}

func renderJoyDivision(w, h int) *generativeart.Canva {
	rand.Seed(202)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 8, G: 10, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewJoyDivision(72, 1.0, 1.0))
	return c
}

func renderBoidsFlock(w, h int) *generativeart.Canva {
	rand.Seed(303)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 10, G: 14, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewBoidsFlock(240, 140, 52.0))
	return c
}

func renderIslamicStar(w, h int) *generativeart.Canva {
	rand.Seed(404)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 18, G: 22, B: 32, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewIslamicStar(4, 67.5, 4.5))
	return c
}

func renderDelaunayStained(w, h int) *generativeart.Canva {
	rand.Seed(505)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 12, G: 14, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Plasma)
	c.Draw(arts.NewDelaunayStained(220, 0.65))
	return c
}

func renderPhyllotaxis(w, h int) *generativeart.Canva {
	rand.Seed(606)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 12, G: 18, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Botanical)
	c.Draw(arts.NewPhyllotaxis(420, 137.50776405, 11.5))
	return c
}

func main() {
	panelW, panelH := 600, 600
	c := generativeart.NewCanva(panelW*3, panelH*2)

	p1 := renderFractalFlame(panelW, panelH)
	if err := p1.ToPNG("images/fractalflame.png"); err != nil {
		panic(err)
	}

	p2 := renderJoyDivision(panelW, panelH)
	if err := p2.ToPNG("images/joydivision.png"); err != nil {
		panic(err)
	}

	p3 := renderBoidsFlock(panelW, panelH)
	if err := p3.ToPNG("images/boidsflock.png"); err != nil {
		panic(err)
	}

	p4 := renderIslamicStar(panelW, panelH)
	if err := p4.ToPNG("images/islamicstar.png"); err != nil {
		panic(err)
	}

	p5 := renderDelaunayStained(panelW, panelH)
	if err := p5.ToPNG("images/delaunaystained.png"); err != nil {
		panic(err)
	}

	p6 := renderPhyllotaxis(panelW, panelH)
	if err := p6.ToPNG("images/phyllotaxis.png"); err != nil {
		panic(err)
	}

	draw.Draw(c.Img(), image.Rect(0, 0, panelW, panelH), p1.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, 0, panelW*2, panelH), p2.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*2, 0, panelW*3, panelH), p3.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(0, panelH, panelW, panelH*2), p4.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, panelH, panelW*2, panelH*2), p5.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*2, panelH, panelW*3, panelH*2), p6.Img(), image.Point{}, draw.Src)

	if err := c.ToPNG("images/exploration_gallery_3.png"); err != nil {
		panic(err)
	}
}
