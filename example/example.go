package main

import (
	"generativeart"
	"math"
)

func formula1(x, y float64) (float64, float64) {
	return 0.18*x*x - math.Sin(y*y),
		0.059*y*y*y - math.Cos(x*x)
}

func main() {
	c := generativeart.NewCanva(1600, 1600, 2, 2)
	c.FillBackgroud(generativeart.White)
	g := generativeart.NewWave(formula1)
	g.GenerativePolar(c, generativeart.DarkSalmon)
	c.ToPng("g1.png")
}
