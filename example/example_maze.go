package main

import (
	"generativeart"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(800, 800, 2, 2)
	c.SetBackground(generativeart.Azure)
	c.SetForeground(generativeart.Tomato)
	c.FillBackground()
	c.Draw(generativeart.NewMaze())
	c.ToPNG("maze.png")
}
