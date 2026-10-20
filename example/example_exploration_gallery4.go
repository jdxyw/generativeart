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

func renderPoincareDisk(w, h int) *generativeart.Canva {
	rand.Seed(101)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 8, G: 10, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewPoincareDisk(7, 3, 5))
	return c
}

func renderSuminagashi(w, h int) *generativeart.Canva {
	rand.Seed(202)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 244, G: 239, B: 229, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewSuminagashi(48, 7, 1.0))
	return c
}

func renderAizawaAttractor(w, h int) *generativeart.Canva {
	rand.Seed(303)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 5, G: 7, B: 15, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewAizawaAttractor(900000, 1.0))
	return c
}

func renderPenroseTiling(w, h int) *generativeart.Canva {
	rand.Seed(404)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 10, G: 14, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewPenroseTiling(6))
	return c
}

func renderIsometricCity(w, h int) *generativeart.Canva {
	rand.Seed(505)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 18, G: 20, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewIsometricCity(11, 1.0))
	return c
}

func renderDLAFrost(w, h int) *generativeart.Canva {
	rand.Seed(606)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 4, G: 8, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewDLAFrost(18000, 0.55))
	return c
}

func main() {
	panelW, panelH := 600, 600
	c := generativeart.NewCanva(panelW*3, panelH*2)

	p1 := renderPoincareDisk(panelW, panelH)
	if err := p1.ToPNG("images/poincaredisk.png"); err != nil {
		panic(err)
	}

	p2 := renderSuminagashi(panelW, panelH)
	if err := p2.ToPNG("images/suminagashi.png"); err != nil {
		panic(err)
	}

	p3 := renderAizawaAttractor(panelW, panelH)
	if err := p3.ToPNG("images/aizawaattractor.png"); err != nil {
		panic(err)
	}

	p4 := renderPenroseTiling(panelW, panelH)
	if err := p4.ToPNG("images/penrosetiling.png"); err != nil {
		panic(err)
	}

	p5 := renderIsometricCity(panelW, panelH)
	if err := p5.ToPNG("images/isometriccity.png"); err != nil {
		panic(err)
	}

	p6 := renderDLAFrost(panelW, panelH)
	if err := p6.ToPNG("images/dlafrost.png"); err != nil {
		panic(err)
	}

	draw.Draw(c.Img(), image.Rect(0, 0, panelW, panelH), p1.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, 0, panelW*2, panelH), p2.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*2, 0, panelW*3, panelH), p3.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(0, panelH, panelW, panelH*2), p4.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, panelH, panelW*2, panelH*2), p5.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*2, panelH, panelW*3, panelH*2), p6.Img(), image.Point{}, draw.Src)

	if err := c.ToPNG("images/exploration_gallery_4.png"); err != nil {
		panic(err)
	}
}
