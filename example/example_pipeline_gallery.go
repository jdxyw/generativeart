package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	panelW, panelH := 600, 600
	c := generativeart.NewCanva(panelW*2, panelH*2)
	c.SetBackground(color.RGBA{R: 12, G: 14, B: 22, A: 255})
	c.FillBackground()

	p1 := generativeart.NewCanva(panelW, panelH)
	p1.SetSeed(1001)
	p1.SetBackground(color.RGBA{R: 12, G: 16, B: 28, A: 255})
	p1.FillBackground()
	p1.SetColorSchema(common.Oceanic)
	p1.Draw(arts.NewEvenStreamlines(7.5, 3.9))
	if err := p1.ToPNG("images/evenstreamlines.png"); err != nil {
		panic(err)
	}

	p2 := generativeart.NewCanva(panelW, panelH)
	p2.SetSeed(1002)
	p2.SetBackground(color.RGBA{R: 12, G: 22, B: 42, A: 255})
	p2.FillBackground()
	p2.SetColorSchema(common.Dunhuang)
	p2.Draw(arts.NewMarchingIsolines(16, 1.0))
	if err := p2.ToPNG("images/marchingisolines.png"); err != nil {
		panic(err)
	}

	p3 := generativeart.NewCanva(panelW, panelH)
	p3.SetSeed(1003)
	p3.SetBackground(color.RGBA{R: 16, G: 22, B: 44, A: 255})
	p3.FillBackground()
	p3.SetColorSchema(common.Dunhuang)
	p3.Draw(arts.NewOrigamiTessellation(14, 14, 1.0))
	if err := p3.ToPNG("images/origamitessellation.png"); err != nil {
		panic(err)
	}

	p4 := generativeart.NewCanva(panelW, panelH)
	p4.SetSeed(1004)
	p4.SetBackground(color.RGBA{R: 18, G: 18, B: 24, A: 255})
	p4.FillBackground()
	p4.SetColorSchema(common.Sunset)
	p4.Draw(arts.NewMondrianSubdivision(4, 6.0))
	if err := p4.ToPNG("images/mondriansubdivision.png"); err != nil {
		panic(err)
	}

	draw.Draw(c.Img(), image.Rect(0, 0, panelW, panelH), p1.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, 0, panelW*2, panelH), p2.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(0, panelH, panelW, panelH*2), p3.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, panelH, panelW*2, panelH*2), p4.Img(), image.Point{}, draw.Src)

	if err := c.ToPNG("images/exploration_gallery_pipeline.png"); err != nil {
		panic(err)
	}
}
