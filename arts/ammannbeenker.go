package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type ammannBeenker struct {
	depth int
}

// NewAmmannBeenker returns an ammannBeenker engine that renders an 8-fold
// octagonal Ammann-Beenker aperiodic quasicrystal tessellation of squares and
// 45-degree rhombi using De Bruijn's 4-grid dual projection, cloisonné enamel
// shading, and golden Ammann bars.
func NewAmmannBeenker(depth int) *ammannBeenker {
	if depth <= 0 {
		depth = 5
	}
	return &ammannBeenker{
		depth: depth,
	}
}

// Generative renders the 8-fold Ammann-Beenker quasicrystal cloisonné tessellation.
func (ab *ammannBeenker) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h

	// Curated Dunhuang / Imperial Cloisonné Quasicrystal Palette
	obsidianGrout := color.RGBA{R: 14, G: 14, B: 22, A: 255}
	wireShadow := color.RGBA{R: 58, G: 38, B: 18, A: 255}
	wireGold := color.RGBA{R: 232, G: 184, B: 74, A: 255}
	wireSpec := color.RGBA{R: 255, G: 242, B: 172, A: 255}

	type tileSwatch struct {
		base, rim, spec color.RGBA
	}
	cinnabarCoral := tileSwatch{
		base: color.RGBA{R: 214, G: 62, B: 46, A: 255},
		rim:  color.RGBA{R: 108, G: 22, B: 22, A: 255},
		spec: color.RGBA{R: 255, G: 168, B: 142, A: 255},
	}
	imperialAmber := tileSwatch{
		base: color.RGBA{R: 242, G: 174, B: 48, A: 255},
		rim:  color.RGBA{R: 138, G: 78, B: 16, A: 255},
		spec: color.RGBA{R: 255, G: 240, B: 156, A: 255},
	}
	jingtaiTurquoise := tileSwatch{
		base: color.RGBA{R: 34, G: 162, B: 178, A: 255},
		rim:  color.RGBA{R: 12, G: 76, B: 98, A: 255},
		spec: color.RGBA{R: 168, G: 242, B: 246, A: 255},
	}
	lapisCobalt := tileSwatch{
		base: color.RGBA{R: 36, G: 82, B: 168, A: 255},
		rim:  color.RGBA{R: 14, G: 34, B: 88, A: 255},
		spec: color.RGBA{R: 132, G: 188, B: 250, A: 255},
	}
	celadonJade := tileSwatch{
		base: color.RGBA{R: 52, G: 158, B: 118, A: 255},
		rim:  color.RGBA{R: 18, G: 76, B: 56, A: 255},
		spec: color.RGBA{R: 176, G: 242, B: 208, A: 255},
	}
	royalPlum := tileSwatch{
		base: color.RGBA{R: 148, G: 38, B: 96, A: 255},
		rim:  color.RGBA{R: 72, G: 14, B: 48, A: 255},
		spec: color.RGBA{R: 246, G: 158, B: 204, A: 255},
	}

	rc.FillBackground(obsidianGrout)
	rc.SetBlendMode(render.BlendOver)

	// De Bruijn 4-Grid Basis Vectors for 8-Fold Ammann-Beenker Quasicrystal:
	// e_r = (cos(r*pi/4), sin(r*pi/4)) for r = 0, 1, 2, 3
	var eVec [4]render.Point
	for r := 0; r < 4; r++ {
		ang := float64(r) * math.Pi / 4.0
		eVec[r] = render.Pt(math.Cos(ang), math.Sin(ang))
	}

	// Non-singular symmetric grid offsets that preserve 8-fold star centers while avoiding 3-line degeneracy
	gamma := [4]float64{0.50013, 0.50029, 0.50013, 0.49971}

	d := ab.depth
	if d < 3 {
		d = 5
	}
	if d > 8 {
		d = 8
	}
	// Tile edge length in pixels scaled to canvas size and depth
	tileEdge := (34.0 / (1.0 + 0.18*float64(d-5))) * scale
	maxGridK := int(math.Ceil((math.Hypot(w, h)*0.58)/(tileEdge*1.8))) + 2

	wireW := 1.45 * scale
	ammannBarPath := &canvas.Path{}

	// Enumerate all grid line intersections between family r and family s (0 <= r < s < 4)
	for r := 0; r < 4; r++ {
		for s := r + 1; s < 4; s++ {
			det := eVec[r].X*eVec[s].Y - eVec[r].Y*eVec[s].X
			if math.Abs(det) < 1e-9 {
				continue
			}
			invDet := 1.0 / det
			isSquare := (s - r) == 2

			for kr := -maxGridK; kr <= maxGridK; kr++ {
				rhsR := float64(kr) - gamma[r]
				for ks := -maxGridK; ks <= maxGridK; ks++ {
					rhsS := float64(ks) - gamma[s]

					// Intersection point in grid-space coordinates
					ix := (rhsR*eVec[s].Y - rhsS*eVec[r].Y) * invDet
					iy := (eVec[r].X*rhsS - eVec[s].X*rhsR) * invDet

					// Compute integer grid indices K[0..3] at (ix, iy)
					var K [4]int
					for m := 0; m < 4; m++ {
						if m == r {
							K[m] = kr
						} else if m == s {
							K[m] = ks
						} else {
							val := ix*eVec[m].X + iy*eVec[m].Y + gamma[m]
							K[m] = int(math.Ceil(val))
						}
					}

					// Project the 4 dual vertices V00, V10, V11, V01 into 2D screen space
					projectDual := func(dr, ds int) render.Point {
						var vx, vy float64
						for m := 0; m < 4; m++ {
							km := K[m]
							if m == r {
								km += dr
							} else if m == s {
								km += ds
							}
							vx += float64(km) * eVec[m].X
							vy += float64(km) * eVec[m].Y
						}
						return render.Pt(cx+vx*tileEdge, cy+vy*tileEdge)
					}

					p00 := projectDual(0, 0)
					p10 := projectDual(1, 0)
					p11 := projectDual(1, 1)
					p01 := projectDual(0, 1)

					centroid := render.Pt(
						0.25*(p00.X+p10.X+p11.X+p01.X),
						0.25*(p00.Y+p10.Y+p11.Y+p01.Y),
					)
					if centroid.X < -tileEdge*1.5 || centroid.X > w+tileEdge*1.5 ||
						centroid.Y < -tileEdge*1.5 || centroid.Y > h+tileEdge*1.5 {
						continue
					}

					// Color assignment by radial ring distance & tile topology (Square vs 45° Star Rhombus)
					rDist := math.Hypot(centroid.X-cx, centroid.Y-cy) / (math.Min(w, h) * 0.5)
					ringBand := int(math.Floor(rDist * 4.5))

					var sw tileSwatch
					if isSquare {
						switch (ringBand + r) % 3 {
						case 0:
							sw = lapisCobalt
						case 1:
							sw = jingtaiTurquoise
						default:
							sw = celadonJade
						}
					} else {
						// 45° Rhombi form the iconic 8-pointed quasicrystal stars!
						switch ringBand % 4 {
						case 0:
							sw = imperialAmber
						case 1:
							sw = cinnabarCoral
						case 2:
							sw = royalPlum
						default:
							sw = jingtaiTurquoise
						}
					}

					pts := []render.Point{p00, p10, p11, p01}
					rc.DrawCloisonneCell(pts, sw.base, sw.rim, sw.spec, wireShadow, wireGold, wireSpec, wireW)

					// Record continuous golden Ammann matching bars across each tile
					e0 := p00.Lerp(p10, 0.38)
					e1 := p10.Lerp(p11, 0.62)
					e2 := p11.Lerp(p01, 0.38)
					e3 := p01.Lerp(p00, 0.62)
					ammannBarPath.MoveTo(e0.X, e0.Y)
					ammannBarPath.LineTo(e1.X, e1.Y)
					ammannBarPath.MoveTo(e2.X, e2.Y)
					ammannBarPath.LineTo(e3.X, e3.Y)
				}
			}
		}
	}

	// Stroke delicate golden Ammann bars across the quasicrystal lattice
	rc.StrokePath(ammannBarPath, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, wireSpec, 0.55)

	c.MergeRenderCanvas(rc)
}
