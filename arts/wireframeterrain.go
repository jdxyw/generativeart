package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type wireframeTerrain struct {
	gridRes     int
	reliefScale float64
}

// NewWireframeTerrain returns a wireframeTerrain engine that renders a 3D
// perspective-projected triangulated alpine glacier & fjord mesh (3D 三角网冰川雪山透视地形)
// with analytical 3D facet normal shading, glowing wireframe edges, and 3D contour rings.
func NewWireframeTerrain(gridRes int, reliefScale float64) *wireframeTerrain {
	if gridRes <= 0 {
		gridRes = 76
	}
	if reliefScale <= 0 {
		reliefScale = 1.0
	}
	return &wireframeTerrain{
		gridRes:     gridRes,
		reliefScale: reliefScale,
	}
}

// Generative renders the 3D perspective triangulated alpine terrain mesh.
func (wt *wireframeTerrain) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Alpine Glacier, Fjord Turquoise, Sunset Gold & Wireframe Palette
	nightSky := color.RGBA{R: 6, G: 9, B: 22, A: 255}
	auroraTeal := color.RGBA{R: 28, G: 138, B: 156, A: 255}
	fjordCobalt := color.RGBA{R: 18, G: 54, B: 118, A: 255}
	glacierCyan := color.RGBA{R: 72, G: 214, B: 232, A: 255}
	sunsetCoral := color.RGBA{R: 228, G: 76, B: 58, A: 255}
	solarGold := color.RGBA{R: 252, G: 198, B: 68, A: 255}
	snowIvory := color.RGBA{R: 255, G: 252, B: 238, A: 255}
	shadowSlate := color.RGBA{R: 12, G: 18, B: 38, A: 255}

	rc.FillBackground(nightSky)
	rc.SetBlendMode(render.BlendOver)

	// Starry night sky + subtle aurora glow in the upper 28% of the canvas
	for y := 0; y < int(0.36*h); y++ {
		fy := float64(y) / (0.36 * h)
		for x := 0; x < int(w); x++ {
			fx := float64(x) / w
			wave := math.Sin(fx*math.Pi*2.5+0.4) * math.Sin(fy*math.Pi)
			col := render.LerpOklab(nightSky, auroraTeal, common.Clamp01(wave*0.55))
			rc.BlendPixel(x, y, col, 1.0)
		}
	}

	ne := common.NewNoiseEngine(c.Seed() + 707)
	nx := wt.gridRes
	if nx < 40 || nx > 110 {
		nx = 76
	}
	ny := nx

	// Evaluate Alpine Twin-Peak + Central Glacial Fjord Heightfield Z(u, v) for u, v in [-1, 1]
	evalHeight := func(u, v float64) float64 {
		// Twin Matterhorn-like glacial peaks on left (u=-0.46) and right (u=+0.44), separated by a fjord valley
		dLeft := math.Hypot((u+0.46)*1.15, (v+0.12)*0.95)
		dRight := math.Hypot((u-0.44)*1.15, (v-0.08)*0.95)
		dBack := math.Hypot(u*0.9, (v+0.58)*1.2)

		peakL := 0.86 * math.Exp(-math.Pow(dLeft/0.44, 1.55))
		peakR := 0.82 * math.Exp(-math.Pow(dRight/0.42, 1.55))
		peakB := 0.64 * math.Exp(-math.Pow(dBack/0.48, 1.6))

		ridged := ne.Ridged2D(u*2.8+1.3, v*2.8-1.7, 4, 2.1, 0.52)
		elev := (math.Max(math.Max(peakL, peakR), peakB)*0.68 + ridged*0.36) * wt.reliefScale

		// Central glacial fjord water plane at z = 0.14
		fjordDist := math.Abs(u - 0.14*math.Sin(v*2.4))
		fjordCarve := common.Smoothstep(0.06, 0.34, fjordDist)
		elev *= (0.12 + 0.88*fjordCarve)
		if elev < 0.135 {
			elev = 0.135 // Flat reflective fjord water level
		}
		return elev
	}

	type meshVert struct {
		p3   vec3
		pt2D render.Point
		elev float64
	}

	verts := make([][]meshVert, ny+1)
	for iy := 0; iy <= ny; iy++ {
		verts[iy] = make([]meshVert, nx+1)
		tz := float64(iy) / float64(ny) // 0 = far horizon, 1 = foreground
		v := tz*2.0 - 1.0

		persp := 0.94 + 0.28*tz
		syBase := (0.20 + tz*0.84) * h
		zScale := 255.0 * scale * (0.76 + 0.32*tz)

		for ix := 0; ix <= nx; ix++ {
			tx := float64(ix)/float64(nx) - 0.5
			u := tx * 2.15
			elev := evalHeight(u, v)

			sx := 0.5*w + tx*w*1.16*persp
			sy := syBase - (elev-0.135)*zScale

			verts[iy][ix] = meshVert{
				p3:   vec3{u, v, elev * 0.65},
				pt2D: render.Pt(sx, sy),
				elev: elev,
			}
		}
	}

	keyLight := (vec3{-0.62, -0.45, 0.64}).norm()
	fillLight := (vec3{0.58, 0.42, 0.70}).norm()
	halfVec := (vec3{keyLight.x, keyLight.y, keyLight.z + 1.0}).norm()

	wireCyan := &canvas.Path{}
	wireGold := &canvas.Path{}

	// Render triangulated terrain rows from back (iy = 0) to front (iy = ny-1)
	for iy := 0; iy < ny; iy++ {
		for ix := 0; ix < nx; ix++ {
			v00 := verts[iy][ix]
			v10 := verts[iy][ix+1]
			v11 := verts[iy+1][ix+1]
			v01 := verts[iy+1][ix]

			tris := [2][3]meshVert{
				{v00, v10, v11},
				{v00, v11, v01},
			}

			for _, tri := range tris {
				a, b, cPt := tri[0], tri[1], tri[2]
				avgElev := (a.elev + b.elev + cPt.elev) / 3.0

				// 3D surface normal in world space
				n := b.p3.sub(a.p3).cross(cPt.p3.sub(a.p3)).norm()
				if n.z < 0 {
					n = vec3{-n.x, -n.y, -n.z}
				}

				ndotl := common.Constrain(n.dot(keyLight), 0, 1)
				ndotFill := common.Constrain(n.dot(fillLight), 0, 1)
				spec := math.Pow(common.Constrain(n.dot(halfVec), 0, 1), 18.0)

				isWater := a.elev <= 0.136 && b.elev <= 0.136 && cPt.elev <= 0.136
				var col color.RGBA
				if isWater {
					ripple := 0.5 + 0.5*math.Sin(float64(ix+iy)*0.8)
					col = render.LerpOklab(fjordCobalt, auroraTeal, ripple*0.65)
				} else {
					uE := common.Clamp01((avgElev - 0.14) / 0.72)
					var baseCol color.RGBA
					switch {
					case uE < 0.35:
						baseCol = render.LerpOklab(fjordCobalt, auroraTeal, uE/0.35)
					case uE < 0.68:
						baseCol = render.LerpOklab(auroraTeal, sunsetCoral, (uE-0.35)/0.33)
					default:
						baseCol = render.LerpOklab(solarGold, snowIvory, (uE-0.68)/0.32)
					}

					illum := common.Constrain(0.18+0.74*ndotl+0.22*ndotFill, 0.08, 1.0)
					col = render.LerpOklab(shadowSlate, baseCol, illum)
					if spec > 0.05 {
						col = render.LerpOklab(col, snowIvory, common.Constrain(spec*0.78, 0, 1))
					}
				}

				rc.FillPolygon([]render.Point{a.pt2D, b.pt2D, cPt.pt2D}, col, 1.0)
			}

			// Record wireframe edges on every second grid cell for a crisp, un-cluttered architectural TIN look
			if (ix+iy)%2 == 0 {
				avgE := 0.5 * (v00.elev + v11.elev)
				targetPath := wireCyan
				if avgE > 0.52 {
					targetPath = wireGold
				}
				targetPath.MoveTo(v00.pt2D.X, v00.pt2D.Y)
				targetPath.LineTo(v10.pt2D.X, v10.pt2D.Y)
				targetPath.LineTo(v11.pt2D.X, v11.pt2D.Y)
				targetPath.LineTo(v00.pt2D.X, v00.pt2D.Y)
			}
		}
	}

	rc.StrokePath(wireCyan, 0.65*scale, canvas.RoundCap, canvas.RoundJoin, glacierCyan, 0.34)
	rc.StrokePath(wireGold, 0.75*scale, canvas.RoundCap, canvas.RoundJoin, solarGold, 0.52)

	rc.ApplyBloom(0.80, 6, 0.24)
	c.MergeRenderCanvas(rc)
}
