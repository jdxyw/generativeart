package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type origamiTessellation struct {
	cols      int
	rows      int
	foldDepth float64
}

// NewOrigamiTessellation returns an origamiTessellation engine that renders a 3D
// Miura-ori and starburst diamond folded paper sculpture with analytical 3D normal
// Lambertian/Blinn-Phong shading, valley crease ambient occlusion, and gilded ridges.
func NewOrigamiTessellation(cols, rows int, foldDepth float64) *origamiTessellation {
	if cols <= 0 {
		cols = 14
	}
	if rows <= 0 {
		rows = 14
	}
	if foldDepth <= 0 {
		foldDepth = 1.0
	}
	return &origamiTessellation{
		cols:      cols,
		rows:      rows,
		foldDepth: foldDepth,
	}
}

type vec3 struct {
	x, y, z float64
}

func (v vec3) sub(o vec3) vec3 {
	return vec3{v.x - o.x, v.y - o.y, v.z - o.z}
}

func (v vec3) cross(o vec3) vec3 {
	return vec3{
		v.y*o.z - v.z*o.y,
		v.z*o.x - v.x*o.z,
		v.x*o.y - v.y*o.x,
	}
}

func (v vec3) dot(o vec3) float64 {
	return v.x*o.x + v.y*o.y + v.z*o.z
}

func (v vec3) norm() vec3 {
	l := math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
	if l < 1e-12 {
		return vec3{0, 0, 1}
	}
	return vec3{v.x / l, v.y / l, v.z / l}
}

// Generative renders the 3D Miura-ori & starburst pyramid origami tessellation.
func (ot *origamiTessellation) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Sculptural Origami Palette (High-Contrast Dunhuang Vermilion, Imperial Gold, Turquoise & Deep Indigo Shadows)
	deepIndigo := color.RGBA{R: 12, G: 18, B: 38, A: 255}
	valleyShadow := color.RGBA{R: 20, G: 44, B: 92, A: 255}
	slateTeal := color.RGBA{R: 34, G: 138, B: 158, A: 255}
	jadeAqua := color.RGBA{R: 72, G: 198, B: 184, A: 255}

	terracottaCoral := color.RGBA{R: 224, G: 68, B: 46, A: 255}
	goldenAmber := color.RGBA{R: 248, G: 176, B: 44, A: 255}
	warmIvory := color.RGBA{R: 255, G: 244, B: 208, A: 255}
	champagneGold := color.RGBA{R: 255, G: 226, B: 124, A: 255}

	rc.FillBackground(deepIndigo)
	rc.SetBlendMode(render.BlendOver)

	nx := ot.cols
	ny := ot.rows
	if nx < 6 {
		nx = 14
	}
	if ny < 6 {
		ny = 14
	}

	// Cover the canvas with a slight overscan margin so folded edges bleed cleanly off-frame
	margin := -28.0 * scale
	spanW := w - 2.0*margin
	spanH := h - 2.0*margin
	dx := spanW / float64(nx)
	dy := spanH / float64(ny)

	// 3D Lighting Setup: Warm Sunlight from Upper-Left-Front + Cool Sky Fill from Lower-Right
	keyLight := (vec3{-0.58, -0.62, 0.52}).norm()
	fillLight := (vec3{0.65, 0.55, 0.52}).norm()
	halfVec := (vec3{keyLight.x, keyLight.y, keyLight.z + 1.0}).norm()

	// Construct the 3D Miura-ori vertex lattice V[iy][ix]
	grid := make([][]vec3, ny+1)
	for iy := 0; iy <= ny; iy++ {
		grid[iy] = make([]vec3, nx+1)
		vFrac := float64(iy) / float64(ny)
		for ix := 0; ix <= nx; ix++ {
			uFrac := float64(ix) / float64(nx)

			baseX := margin + float64(ix)*dx
			baseY := margin + float64(iy)*dy

			// Radial distance from center for concentric wave modulation
			rx := uFrac - 0.5
			ry := vFrac - 0.5
			rDist := math.Hypot(rx, ry)

			// Miura-ori herringbone lateral chevron offset along X and Y
			zigY := 1.0
			if iy%2 == 1 {
				zigY = -1.0
			}
			zigX := 1.0
			if ix%2 == 1 {
				zigX = -1.0
			}

			chevronShiftX := zigY * dx * 0.24 * (1.0 + 0.25*math.Cos(rDist*math.Pi*3.0))
			chevronShiftY := zigX * dy * 0.14 * math.Sin((uFrac+vFrac)*math.Pi*2.0)

			// 3D elevation Z combining alternating crease pleats + global dome wave
			pleatZ := (zigX*0.55 + zigY*0.45) * 24.0 * scale * ot.foldDepth
			waveZ := 28.0 * scale * math.Cos(rDist*math.Pi*3.2)
			px := baseX + chevronShiftX
			py := baseY + chevronShiftY - waveZ*0.22
			pz := pleatZ + waveZ

			grid[iy][ix] = vec3{px, py, pz}
		}
	}

	shadeFacet := func(p0, p1, p2 vec3, baseTintT float64, isValley bool) color.RGBA {
		n := p1.sub(p0).cross(p2.sub(p0)).norm()
		if n.z < 0 {
			n = vec3{-n.x, -n.y, -n.z}
		}

		ndotl := common.Constrain(n.dot(keyLight), 0, 1)
		ndotFill := common.Constrain(n.dot(fillLight), 0, 1)
		spec := math.Pow(common.Constrain(n.dot(halfVec), 0, 1), 16.0)

		// Base paper color shifts across canvas from cool jade/teal through golden amber to warm coral
		var surfaceBase color.RGBA
		tClamped := common.Constrain(baseTintT, 0, 1)
		if tClamped < 0.5 {
			surfaceBase = render.LerpOklab(jadeAqua, goldenAmber, tClamped*2.0)
		} else {
			surfaceBase = render.LerpOklab(goldenAmber, terracottaCoral, (tClamped-0.5)*2.0)
		}
		litCol := render.LerpOklab(surfaceBase, warmIvory, 0.28)
		coolCol := render.LerpOklab(valleyShadow, slateTeal, common.Constrain(ndotFill*0.92, 0, 1))

		// Strong 3D facet separation between sunlit slopes and shadowed slopes
		lightCurve := common.Constrain((ndotl-0.12)/0.78, 0, 1)
		lightCurve = lightCurve * lightCurve * (3.0 - 2.0*lightCurve)
		col := render.LerpOklab(coolCol, litCol, lightCurve)
		if isValley {
			// Valley ambient occlusion darkening
			col = render.LerpOklab(col, deepIndigo, 0.38)
		}
		if spec > 0.04 {
			col = render.LerpOklab(col, champagneGold, common.Constrain(spec*0.82, 0, 1))
		}
		return col
	}

	mountainCreases := &canvas.Path{}
	valleyCreases := &canvas.Path{}

	// Subdivide each quad cell into 4 starburst origami facets around an elevated/recessed apex
	for iy := 0; iy < ny; iy++ {
		for ix := 0; ix < nx; ix++ {
			v00 := grid[iy][ix]
			v10 := grid[iy][ix+1]
			v11 := grid[iy+1][ix+1]
			v01 := grid[iy+1][ix]

			uMid := (float64(ix) + 0.5) / float64(nx)
			vMid := (float64(iy) + 0.5) / float64(ny)
			diagT := common.Constrain(uMid*0.58+(1.0-vMid)*0.42+0.15*math.Sin((uMid-vMid)*math.Pi*2.0), 0, 1)

			// Alternate raised mountain pyramids and recessed valley Hopp-folds
			isRaised := (ix+iy)%2 == 0
			apexOffsetZ := 22.0 * scale * ot.foldDepth
			if !isRaised {
				apexOffsetZ = -18.0 * scale * ot.foldDepth
			}

			center := vec3{
				x: 0.25*(v00.x+v10.x+v11.x+v01.x) - apexOffsetZ*0.12,
				y: 0.25*(v00.y+v10.y+v11.y+v01.y) - apexOffsetZ*0.18,
				z: 0.25*(v00.z+v10.z+v11.z+v01.z) + apexOffsetZ,
			}

			tris := [4][3]vec3{
				{v00, v10, center}, // Top facet
				{v10, v11, center}, // Right facet
				{v11, v01, center}, // Bottom facet
				{v01, v00, center}, // Left facet
			}

			for tIdx, tri := range tris {
				a := render.Pt(tri[0].x, tri[0].y)
				b := render.Pt(tri[1].x, tri[1].y)
				cPt := render.Pt(tri[2].x, tri[2].y)

				isValleyFacet := (!isRaised && (tIdx == 1 || tIdx == 2))
				col := shadeFacet(tri[0], tri[1], tri[2], diagT, isValleyFacet)
				rc.FillPolygon([]render.Point{a, b, cPt}, col, 1.0)

				// Subtle folded paper grain lines inside each triangle facet
				grainCol := render.LerpOklab(col, warmIvory, 0.16)
				if isValleyFacet {
					grainCol = render.LerpOklab(col, deepIndigo, 0.20)
				}
				rc.DrawHatchPatch(
					[]render.Point{a, b},
					[]render.Point{a.Lerp(cPt, 0.92), b.Lerp(cPt, 0.92)},
					5, 0.72*scale, 0.0,
					func(s, t float64) color.RGBA { return grainCol },
					0.45,
				)
			}

			// Record mountain and valley crease lines
			p00 := render.Pt(v00.x, v00.y)
			p10 := render.Pt(v10.x, v10.y)
			p11 := render.Pt(v11.x, v11.y)
			p01 := render.Pt(v01.x, v01.y)
			pC := render.Pt(center.x, center.y)

			if isRaised {
				for _, corner := range []render.Point{p00, p10, p11, p01} {
					mountainCreases.MoveTo(pC.X, pC.Y)
					mountainCreases.LineTo(corner.X, corner.Y)
				}
				valleyCreases.MoveTo(p00.X, p00.Y)
				valleyCreases.LineTo(p10.X, p10.Y)
				valleyCreases.LineTo(p11.X, p11.Y)
			} else {
				for _, corner := range []render.Point{p00, p10, p11, p01} {
					valleyCreases.MoveTo(pC.X, pC.Y)
					valleyCreases.LineTo(corner.X, corner.Y)
				}
				mountainCreases.MoveTo(p00.X, p00.Y)
				mountainCreases.LineTo(p01.X, p01.Y)
				mountainCreases.LineTo(p11.X, p11.Y)
			}
		}
	}

	// Stroke Valley Creases (dark indigo AO shadow) and Mountain Creases (crisp champagne-gold ridge)
	rc.StrokePath(valleyCreases, 1.45*scale, canvas.RoundCap, canvas.RoundJoin, deepIndigo, 0.86)
	rc.StrokePath(mountainCreases, 1.15*scale, canvas.RoundCap, canvas.RoundJoin, champagneGold, 0.88)

	rc.ApplyPaperTexture(0.06, 202614)

	c.MergeRenderCanvas(rc)
}
