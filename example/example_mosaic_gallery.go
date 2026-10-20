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

func main() {
	panelW, panelH := 600, 600
	c := generativeart.NewCanva(panelW*2, panelH*2)
	c.SetBackground(color.RGBA{R: 12, G: 14, B: 24, A: 255})
	c.FillBackground()

	rand.Seed(801)
	p1 := generativeart.NewCanva(panelW, panelH)
	p1.SetBackground(color.RGBA{R: 16, G: 20, B: 38, A: 255})
	p1.FillBackground()
	p1.SetColorSchema(common.Oceanic)
	p1.Draw(arts.NewMosaicVortex(34, 1.0))
	if err := p1.ToPNG("images/mosaicvortex.png"); err != nil {
		panic(err)
	}

	rand.Seed(802)
	p2 := generativeart.NewCanva(panelW, panelH)
	p2.SetBackground(color.RGBA{R: 18, G: 24, B: 44, A: 255})
	p2.FillBackground()
	p2.SetColorSchema(common.UkiyoE)
	p2.Draw(arts.NewMosaicWave(32, 1.0))
	if err := p2.ToPNG("images/mosaicwave.png"); err != nil {
		panic(err)
	}

	rand.Seed(803)
	p3 := generativeart.NewCanva(panelW, panelH)
	p3.SetBackground(color.RGBA{R: 22, G: 18, B: 32, A: 255})
	p3.FillBackground()
	p3.SetColorSchema(common.Dunhuang)
	p3.Draw(arts.NewMosaicCanyon(32, 1.0))
	if err := p3.ToPNG("images/mosaiccanyon.png"); err != nil {
		panic(err)
	}

	rand.Seed(804)
	p4 := generativeart.NewCanva(panelW, panelH)
	p4.SetBackground(color.RGBA{R: 16, G: 24, B: 42, A: 255})
	p4.FillBackground()
	p4.SetColorSchema(common.Botanical)
	p4.Draw(arts.NewMosaicLotus(12, 1.0))
	if err := p4.ToPNG("images/mosaiclotus.png"); err != nil {
		panic(err)
	}

	draw.Draw(c.Img(), image.Rect(0, 0, panelW, panelH), p1.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, 0, panelW*2, panelH), p2.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(0, panelH, panelW, panelH*2), p3.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, panelH, panelW*2, panelH*2), p4.Img(), image.Point{}, draw.Src)

	if err := c.ToPNG("images/exploration_gallery_mosaic.png"); err != nil {
		panic(err)
	}
}
