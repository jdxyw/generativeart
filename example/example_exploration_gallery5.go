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

func renderImpastoFlow(w, h int) *generativeart.Canva {
	rand.Seed(101)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 8, G: 7, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewImpastoFlow(3200, 520, 0.0032))
	return c
}

func renderVoxelTerrain(w, h int) *generativeart.Canva {
	rand.Seed(202)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 10, G: 12, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewVoxelTerrain(340, 2.6, 1.0))
	return c
}

func renderPackedCapsules(w, h int) *generativeart.Canva {
	rand.Seed(303)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 238, G: 236, B: 229, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewPackedCapsules(2200, 7.2, 3.6))
	return c
}

func renderOrphicDiscs(w, h int) *generativeart.Canva {
	rand.Seed(404)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 72, G: 178, B: 102, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Youthful)
	c.Draw(arts.NewOrphicDiscs(34))
	return c
}

func renderNautilusSpiral(w, h int) *generativeart.Canva {
	rand.Seed(505)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 18, G: 24, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewNautilusSpiral(5, 68))
	return c
}

func renderMosaicSun(w, h int) *generativeart.Canva {
	rand.Seed(606)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 22, G: 26, B: 48, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewMosaicSun(32, 1.0))
	return c
}

func renderTangentCones(w, h int) *generativeart.Canva {
	rand.Seed(707)
	c := generativeart.NewCanva(w, h)
	c.SetBackground(color.RGBA{R: 234, G: 88, B: 56, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewTangentCones(26))
	return c
}

func main() {
	panelW, panelH := 600, 600
	c := generativeart.NewCanva(panelW*4, panelH*2)
	c.SetBackground(color.RGBA{R: 10, G: 12, B: 20, A: 255})
	c.FillBackground()

	p1 := renderImpastoFlow(panelW, panelH)
	if err := p1.ToPNG("images/impastoflow.png"); err != nil {
		panic(err)
	}

	p2 := renderVoxelTerrain(panelW, panelH)
	if err := p2.ToPNG("images/voxelterrain.png"); err != nil {
		panic(err)
	}

	p3 := renderPackedCapsules(panelW, panelH)
	if err := p3.ToPNG("images/packedcapsules.png"); err != nil {
		panic(err)
	}

	p4 := renderOrphicDiscs(panelW, panelH)
	if err := p4.ToPNG("images/orphicdiscs.png"); err != nil {
		panic(err)
	}

	p5 := renderNautilusSpiral(panelW, panelH)
	if err := p5.ToPNG("images/nautilusspiral.png"); err != nil {
		panic(err)
	}

	p6 := renderMosaicSun(panelW, panelH)
	if err := p6.ToPNG("images/mosaicsun.png"); err != nil {
		panic(err)
	}

	p7 := renderTangentCones(panelW, panelH)
	if err := p7.ToPNG("images/tangentcones.png"); err != nil {
		panic(err)
	}

	// Row 1: 4 panels
	draw.Draw(c.Img(), image.Rect(0, 0, panelW, panelH), p1.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW, 0, panelW*2, panelH), p2.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*2, 0, panelW*3, panelH), p3.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(panelW*3, 0, panelW*4, panelH), p4.Img(), image.Point{}, draw.Src)

	// Row 2: 3 panels centered horizontally
	offsetX := panelW / 2
	draw.Draw(c.Img(), image.Rect(offsetX, panelH, offsetX+panelW, panelH*2), p5.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(offsetX+panelW, panelH, offsetX+panelW*2, panelH*2), p6.Img(), image.Point{}, draw.Src)
	draw.Draw(c.Img(), image.Rect(offsetX+panelW*2, panelH, offsetX+panelW*3, panelH*2), p7.Img(), image.Point{}, draw.Src)

	if err := c.ToPNG("images/exploration_gallery_5.png"); err != nil {
		panic(err)
	}
}
