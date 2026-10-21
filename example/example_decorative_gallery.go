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
	c.SetBackground(color.RGBA{R: 12, G: 12, B: 18, A: 255})
	c.FillBackground()

	rand.Seed(901)
	p1 := generativeart.NewCanva(panelW, panelH)
	p1.SetBackground(color.RGBA{R: 18, G: 14, B: 22, A: 255})
	p1.FillBackground()
	p1.SetColorSchema(common.Dunhuang)
	p1.Draw(arts.NewCloisonneEnamel(6, 12))
	if err := p1.ToPNG("images/cloisonneenamel.png"); err != nil {
		panic(err)
	}

	rand.Seed(902)
	p2 := generativeart.NewCanva(panelW, panelH)
	p2.SetBackground(color.RGBA{R: 242, G: 232, B: 206, A: 255})
	p2.FillBackground()
	p2.SetColorSchema(common.UkiyoE)
	p2.Draw(arts.NewWoodblockWave(5, 42))
	if err := p2.ToPNG("images/woodblockwave.png"); err != nil {
		panic(err)
	}

	rand.Seed(903)
	p3 := generativeart.NewCanva(panelW, panelH)
	p3.SetBackground(color.RGBA{R: 14, G: 16, B: 24, A: 255})
	p3.FillBackground()
	p3.SetColorSchema(common.Sunset)
	p3.Draw(arts.NewArtDecoSunburst(36, 7))
	if err := p3.ToPNG("images/artdecosunburst.png"); err != nil {
		panic(err)
	}

	rand.Seed(904)
	p4 := generativeart.NewCanva(panelW, panelH)
	p4.SetBackground(color.RGBA{R: 12, G: 11, B: 18, A: 255})
	p4.FillBackground()
	p4.SetColorSchema(common.Cyberpunk)
	p4.Draw(arts.NewStainedRoseWindow(12, 5))
	if err := p4.ToPNG("images/stainedrosewindow.png"); err != nil {
		panic(err)
	}

	draw.Draw(c.Img(), image.Rect(0, 0, panelW, panelH), p1.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, 0, panelW*2, panelH), p2.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(0, panelH, panelW, panelH*2), p3.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, panelH, panelW*2, panelH*2), p4.Img(), image.Point{}, draw.Src)

	if err := c.ToPNG("images/exploration_gallery_decorative.png"); err != nil {
		panic(err)
	}
}
