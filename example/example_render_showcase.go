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

func renderPanel1CurlFlow(w, h int) *generativeart.Canva {
	rand.Seed(101)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 14, G: 18, B: 28, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewCurlFlow(110, 36, 0.0045))
	return c
}

func renderPanel2SandRing(w, h int) *generativeart.Canva {
	rand.Seed(202)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 8, G: 10, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Plasma)
	c.Draw(arts.NewSandRing(22, 22, 1.3))
	return c
}

func renderPanel3Watercolor(w, h int) *generativeart.Canva {
	rand.Seed(303)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 247, G: 243, B: 234, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.SetLineColor(color.RGBA{R: 32, G: 28, B: 38, A: 255})
	c.Draw(arts.NewWatercolor(9, 22, 5, 0.32))
	return c
}

func renderPanel4Guilloche(w, h int) *generativeart.Canva {
	rand.Seed(404)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 12, G: 15, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewGuilloche(28, 175, 65, 115))
	return c
}

func main() {
	panelW, panelH := 600, 600
	c := generativeart.NewCanva(panelW*2, panelH*2)

	p1 := renderPanel1CurlFlow(panelW, panelH)
	if err := p1.ToPNG("images/curlflow.png"); err != nil {
		panic(err)
	}

	p2 := renderPanel2SandRing(panelW, panelH)
	if err := p2.ToPNG("images/sandring.png"); err != nil {
		panic(err)
	}

	p3 := renderPanel3Watercolor(panelW, panelH)
	if err := p3.ToPNG("images/watercolor.png"); err != nil {
		panic(err)
	}

	p4 := renderPanel4Guilloche(panelW, panelH)
	if err := p4.ToPNG("images/guilloche.png"); err != nil {
		panic(err)
	}

	draw.Draw(c.Img(), image.Rect(0, 0, panelW, panelH), p1.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, 0, panelW*2, panelH), p2.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(0, panelH, panelW, panelH*2), p3.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, panelH, panelW*2, panelH*2), p4.Img(), image.Point{}, draw.Src)

	if err := c.ToPNG("images/render_upgrade_showcase.png"); err != nil {
		panic(err)
	}
}
