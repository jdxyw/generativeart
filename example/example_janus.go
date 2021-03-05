package main

import (
	"github.com/jdxyw/generativeart"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(generativeart.Black)
	c.FillBackground()
	c.SetColorSchema(generativeart.DarkRed)
	c.SetForeground(generativeart.LightPink)
	c.Draw(generativeart.NewJanus(10, 0.2))
	c.ToPNG("janus.png")
}
