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

func renderGrayScott(w, h int) *generativeart.Canva {
	rand.Seed(101)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 8, G: 12, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Viridis)
	c.Draw(arts.NewGrayScott(3200, 0.037, 0.060))
	return c
}

func renderLissajousRibbon(w, h int) *generativeart.Canva {
	rand.Seed(202)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 6, G: 8, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Plasma)
	c.Draw(arts.NewLissajousRibbon(140, 3.0, 4.0, 5.0))
	return c
}

func renderApollonianGasket(w, h int) *generativeart.Canva {
	rand.Seed(303)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 10, G: 14, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewApollonianGasket(7, 2.5))
	return c
}

func renderSpaceColonization(w, h int) *generativeart.Canva {
	rand.Seed(404)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 8, G: 14, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewSpaceColonization(1800, 160, 6.0, 55.0))
	return c
}

func renderTruchetFlow(w, h int) *generativeart.Canva {
	rand.Seed(505)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 245, G: 240, B: 230, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewTruchetFlow(6, 3))
	return c
}

func renderAuroraWave(w, h int) *generativeart.Canva {
	rand.Seed(606)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 5, G: 8, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewAuroraWave(4, 220, 1.0))
	return c
}

func main() {
	panelW, panelH := 600, 600
	c := generativeart.NewCanva(panelW*3, panelH*2)

	p1 := renderGrayScott(panelW, panelH)
	if err := p1.ToPNG("images/grayscott.png"); err != nil {
		panic(err)
	}

	p2 := renderLissajousRibbon(panelW, panelH)
	if err := p2.ToPNG("images/lissajousribbon.png"); err != nil {
		panic(err)
	}

	p3 := renderApollonianGasket(panelW, panelH)
	if err := p3.ToPNG("images/apolloniangasket.png"); err != nil {
		panic(err)
	}

	p4 := renderSpaceColonization(panelW, panelH)
	if err := p4.ToPNG("images/spacecolonization.png"); err != nil {
		panic(err)
	}

	p5 := renderTruchetFlow(panelW, panelH)
	if err := p5.ToPNG("images/truchetflow.png"); err != nil {
		panic(err)
	}

	p6 := renderAuroraWave(panelW, panelH)
	if err := p6.ToPNG("images/aurorawave.png"); err != nil {
		panic(err)
	}

	draw.Draw(c.Img(), image.Rect(0, 0, panelW, panelH), p1.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, 0, panelW*2, panelH), p2.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*2, 0, panelW*3, panelH), p3.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(0, panelH, panelW, panelH*2), p4.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, panelH, panelW*2, panelH*2), p5.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*2, panelH, panelW*3, panelH*2), p6.Img(), image.Point{}, draw.Src)

	if err := c.ToPNG("images/exploration_gallery_2.png"); err != nil {
		panic(err)
	}
}
