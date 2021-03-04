package main

import (
	"github.com/jdxyw/generativeart"
	"math/rand"
	"time"
)

func julia1(z complex128) complex128 {
	c := complex(-0.1, 0.651)

	z = z*z + c

	return z
}

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500, 1.5, 1.5)
	c.SetIterations(800)
	c.SetColorSchema(generativeart.Viridis)
	c.FillBackground()
	c.Draw(generativeart.NewJulia(julia1, 40))
	c.ToPNG("julia.png")
}
