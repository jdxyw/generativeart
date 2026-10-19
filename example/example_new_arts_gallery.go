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

func renderCliffordNebula(w, h int) *generativeart.Canva {
	rand.Seed(101)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 6, G: 6, B: 14, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Inferno)
	c.Draw(arts.NewCliffordNebula(1500000, -1.7, 1.8, -1.9, -0.4))
	return c
}

func renderPhysarum(w, h int) *generativeart.Canva {
	rand.Seed(202)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 6, G: 10, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewPhysarum(45000, 110, 0, 0))
	return c
}

func renderDifferentialGrowth(w, h int) *generativeart.Canva {
	rand.Seed(303)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 10, G: 12, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewDifferentialGrowth(220, 1600, 3.5, 7.5, 18.0))
	return c
}

func renderChladni(w, h int) *generativeart.Canva {
	rand.Seed(404)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 8, G: 12, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewChladni(90000, 7, 3, 1.0, -1.0))
	return c
}

func renderVoronoiCrystal(w, h int) *generativeart.Canva {
	rand.Seed(505)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 24, G: 22, B: 30, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewVoronoiCrystal(85, 4))
	return c
}

func renderInkLandscape(w, h int) *generativeart.Canva {
	rand.Seed(606)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 244, G: 239, B: 228, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewInkLandscape(8, 0.45))
	return c
}

func main() {
	panelW, panelH := 600, 600
	c := generativeart.NewCanva(panelW*3, panelH*2)

	p1 := renderCliffordNebula(panelW, panelH)
	if err := p1.ToPNG("images/cliffordnebula.png"); err != nil {
		panic(err)
	}

	p2 := renderPhysarum(panelW, panelH)
	if err := p2.ToPNG("images/physarum.png"); err != nil {
		panic(err)
	}

	p3 := renderDifferentialGrowth(panelW, panelH)
	if err := p3.ToPNG("images/differentialgrowth.png"); err != nil {
		panic(err)
	}

	p4 := renderChladni(panelW, panelH)
	if err := p4.ToPNG("images/chladni.png"); err != nil {
		panic(err)
	}

	p5 := renderVoronoiCrystal(panelW, panelH)
	if err := p5.ToPNG("images/voronoicrystal.png"); err != nil {
		panic(err)
	}

	p6 := renderInkLandscape(panelW, panelH)
	if err := p6.ToPNG("images/inklandscape.png"); err != nil {
		panic(err)
	}

	draw.Draw(c.Img(), image.Rect(0, 0, panelW, panelH), p1.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, 0, panelW*2, panelH), p2.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*2, 0, panelW*3, panelH), p3.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(0, panelH, panelW, panelH*2), p4.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, panelH, panelW*2, panelH*2), p5.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*2, panelH, panelW*3, panelH*2), p6.Img(), image.Point{}, draw.Src)

	if err := c.ToPNG("images/new_arts_gallery.png"); err != nil {
		panic(err)
	}
}
