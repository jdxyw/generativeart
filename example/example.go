package main

import (
	"generativeart"
	"math"
)

func formula1(x, y float64) (float64, float64) {
	return -0.13*x*x - math.Sin(y*y) + math.Cos(x*y),
		-0.96*y*y*y - math.Cos(x*x)
}

func main() {
	c := generativeart.NewCanva(800, 800, 4, 4)
	c.FillBackgroud(generativeart.AntiqueWhite)
	g := generativeart.NewWave(formula1)
	g.GenerativePolar(c, generativeart.DarkSalmon)
	c.ToPng("g1.png")
}
