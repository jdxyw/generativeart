package main

import "generativeart"

func main() {
	c := generativeart.NewCanva(800, 800, 2, 2)
	c.FillBackgroud(generativeart.Azure)
	g := generativeart.NewMaze()
	g.Generative(c, generativeart.Tomato)
	c.ToPng("maze.png")
}
