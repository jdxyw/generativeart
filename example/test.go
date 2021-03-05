package main

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"math/rand"
	"time"
)

func main() {
	const S = 400
	rand.Seed(time.Now().Unix())
	dest := image.NewRGBA(image.Rect(0, 0, 500, 500))

	gc := draw2dimg.NewGraphicContext(dest)
	gc.SetStrokeColor(generativeart.Orange)
	gc.BeginPath()
	gc.MoveTo(100, 100)
	gc.QuadCurveTo(150, 50, 200, 100)
	//fmt.Println(gc.LastPoint())
	//gc.QuadCurveTo(200, 100, 300, 200)
	gc.QuadCurveTo(250, 250, 200, 400)
	gc.QuadCurveTo(100, 250, 100, 100)
	//gc.QuadCurveTo(200, 400, 100, 10)
	//gc.Close()
	gc.Stroke()

	ctex := gg.NewContextForRGBA(dest)
	ctex.SetColor(generativeart.White)
	//ctex.DrawCircle(100, 100, 5)
	//ctex.DrawCircle(200, 400, 5)
	//ctex.DrawCircle(200, 100, 5)
	//ctex.DrawCircle(300, 200, 5)
	//ctex.DrawCircle(200, 400, 5)
	ctex.Fill()
	//gc.FillStroke()

	draw2dimg.SaveToPngFile("test.png", dest)
}
