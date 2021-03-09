package main

import "github.com/fogleman/gg"

func main() {
	//const S = 400
	//rand.Seed(time.Now().Unix())
	//dest := image.NewRGBA(image.Rect(0, 0, 500, 500))
	//
	//ctex := gg.NewContextForRGBA(dest)
	//ctex.Push()
	//ctex.Translate(500/2, 500/2)
	//ctex.Rotate(40)
	//ctex.SetColor(color.RGBA{0xFF, 0x00, 0x00, 255})
	//for theta :=0.0; theta<361.0; theta+=1.0 {
	//	x := 100*math.Cos(gg.Radians(theta)) - 100*math.Pow(math.Sin(gg.Radians(theta)), 2) / math.Sqrt(2)
	//	y := 100*math.Cos(gg.Radians(theta))*math.Sin(gg.Radians(theta))
	//
	//	x1 := 100*math.Cos(gg.Radians(theta+1)) - 100*math.Pow(math.Sin(gg.Radians(theta+1)), 2) / math.Sqrt(2)
	//	y1 := 100*math.Cos(gg.Radians(theta+1))*math.Sin(gg.Radians(theta+1))
	//
	//	ctex.DrawLine(x, y, x1, y1)
	//	ctex.Stroke()
	//}
	//ctex.Pop()
	//
	//
	//f, _ := os.Create("test.png")
	//
	//if err := png.Encode(f, dest); err != nil {
	//	f.Close()
	//}

	dc := gg.NewContext(1000, 1000)
	dc.DrawCircle(350, 500, 300)
	dc.Clip()
	dc.DrawCircle(650, 500, 300)
	dc.Clip()
	dc.DrawRectangle(0, 0, 1000, 1000)
	dc.SetRGB(0, 0, 0)
	dc.Fill()
	dc.SavePNG("out.png")
}
