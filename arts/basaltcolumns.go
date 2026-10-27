package arts

import (
	"image/color"
	"math"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type basaltColumns struct {
	hexRings     int
	pillarRelief float64
}

// NewBasaltColumns returns a basaltColumns engine that renders an isometric 3D
// hexagonal basalt columnar jointing cliff and volcanic magma fissure canyon
// (巨人堤道六棱玄武岩柱状节理与熔岩海蚀峡谷) with painter-sorted 3D hexagonal prisms.
func NewBasaltColumns(hexRings int, pillarRelief float64) *basaltColumns {
	if hexRings <= 0 {
		hexRings = 16
	}
	if pillarRelief <= 0 {
		pillarRelief = 1.0
	}
	return &basaltColumns{
		hexRings:     hexRings,
		pillarRelief: pillarRelief,
	}
}

// Generative renders the 3D isometric hexagonal basalt column terrain.
func (bc *basaltColumns) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.58*h

	// Volcanic Magma Fissure, Seafoam Turquoise & Columnar Basalt Palette
	abyssBg := color.RGBA{R: 8, G: 11, B: 20, A: 255}
	magmaWhite := color.RGBA{R: 255, G: 244, B: 192, A: 255}
	magmaGold := color.RGBA{R: 252, G: 184, B: 46, A: 255}
	magmaCrimson := color.RGBA{R: 218, G: 52, B: 38, A: 255}

	seafoamMint := color.RGBA{R: 138, G: 236, B: 218, A: 255}
	tealBasalt := color.RGBA{R: 42, G: 152, B: 164, A: 255}
	slateBlue := color.RGBA{R: 44, G: 78, B: 122, A: 255}
	basaltMid := color.RGBA{R: 28, G: 42, B: 68, A: 255}
	basaltDark := color.RGBA{R: 14, G: 18, B: 32, A: 255}
	goldRim := color.RGBA{R: 244, G: 198, B: 92, A: 255}

	rc.FillBackground(abyssBg)
	rc.SetBlendMode(render.BlendOver)

	ne := common.NewNoiseEngine(c.Seed() + 303)
	rings := bc.hexRings
	if rings < 10 || rings > 22 {
		rings = 16
	}

	hexR := (15.5 * scale) * (16.0 / float64(rings))
	isoAspect := 0.58 // Isometric vertical foreshortening

	type hexPillar struct {
		q, r         int
		wx, wy       float64
		sx, sy       float64
		height       float64
		elevNorm     float64
		isMagmaTrough bool
	}

	pillars := make([]hexPillar, 0, (2*rings+1)*(2*rings+1))

	for q := -rings; q <= rings; q++ {
		r1 := common.MaxInt(-rings, -q-rings)
		r2 := common.MinInt(rings, -q+rings)
		for r := r1; r <= r2; r++ {
			// Axial hex (q, r) to 2D Cartesian world coordinates
			wx := hexR * math.Sqrt(3.0) * (float64(q) + 0.5*float64(r))
			wy := hexR * 1.5 * float64(r)

			// Rotate world coordinates by 15 degrees for dynamic isometric perspective
			cosA, sinA := math.Cos(0.26), math.Sin(0.26)
			rx := wx*cosA - wy*sinA
			ry := wx*sinA + wy*cosA

			u := rx / (0.46 * w)
			v := ry / (0.46 * h)

			// S-curved volcanic fissure canyon cutting diagonally across the basalt plateau
			fissureDist := math.Abs(u - 0.32*math.Sin(v*2.6))
			canyonWall := common.Smoothstep(0.08, 0.52, fissureDist)

			// Stepped terraced basalt column heights
			nVal := 0.5 + 0.5*ne.FBM2D(u*2.8+1.5, v*2.8-1.2, 3, 2.0, 0.5)
			ridged := ne.Ridged2D(u*3.4-0.8, v*3.4+2.1, 3, 2.0, 0.5)
			rawElev := (0.15 + 0.85*canyonWall*(0.55*nVal+0.45*ridged)) * bc.pillarRelief

			// Quantize into crisp geological basalt tiers
			quantElev := math.Floor(rawElev*14.0)/14.0*0.65 + rawElev*0.35
			if fissureDist < 0.14 {
				quantElev = 0.04 + 0.06*nVal
			}

			pillars = append(pillars, hexPillar{
				q:             q,
				r:             r,
				wx:            rx,
				wy:            ry,
				sx:            cx + rx,
				sy:            cy + ry*isoAspect,
				height:        quantElev * 168.0 * scale,
				elevNorm:      common.Clamp01(quantElev),
				isMagmaTrough: fissureDist < 0.16,
			})
		}
	}

	// Sort hex pillars strictly from back (smallest sy) to front (largest sy) for Painter's occlusion
	sort.Slice(pillars, func(i, j int) bool {
		if math.Abs(pillars[i].sy-pillars[j].sy) > 0.5 {
			return pillars[i].sy < pillars[j].sy
		}
		return pillars[i].sx < pillars[j].sx
	})

	// 6 isometric vertices of a flat-topped/pointy-topped hexagon
	hexCorner := func(centerX, centerY, rad float64, idx int) render.Point {
		ang := (float64(idx)*60.0 + 30.0) * math.Pi / 180.0
		return render.Pt(centerX+rad*math.Cos(ang), centerY+rad*isoAspect*math.Sin(ang))
	}

	for _, p := range pillars {
		topY := p.sy - p.height
		baseY := p.sy + 28.0*scale

		topPts := make([]render.Point, 6)
		botPts := make([]render.Point, 6)
		for k := 0; k < 6; k++ {
			topPts[k] = hexCorner(p.sx, topY, hexR*0.97, k)
			botPts[k] = hexCorner(p.sx, baseY, hexR*0.97, k)
		}

		// Determine cap & cliff colors based on elevation and magma proximity
		var capCol, leftWallCol, frontWallCol, rightWallCol, rimStroke color.RGBA
		if p.isMagmaTrough {
			capCol = render.LerpOklab(magmaCrimson, magmaGold, p.elevNorm*5.0)
			leftWallCol = magmaCrimson
			frontWallCol = render.LerpOklab(magmaCrimson, basaltDark, 0.45)
			rightWallCol = basaltDark
			rimStroke = magmaWhite
		} else {
			u := p.elevNorm
			switch {
			case u < 0.35:
				// Warm magma-lit lower canyon ledges
				capCol = render.LerpOklab(magmaCrimson, tealBasalt, u/0.35)
			case u < 0.72:
				capCol = render.LerpOklab(tealBasalt, seafoamMint, (u-0.35)/0.37)
			default:
				capCol = render.LerpOklab(seafoamMint, goldRim, (u-0.72)/0.28)
			}
			leftWallCol = render.LerpOklab(slateBlue, tealBasalt, u*0.65)
			frontWallCol = render.LerpOklab(basaltMid, slateBlue, u*0.55)
			rightWallCol = basaltDark
			rimStroke = render.LerpOklab(capCol, magmaWhite, 0.55)
		}

		// Front-facing 3 vertical side faces of the pointy-topped hexagonal prism:
		// Vertices 0 (30 deg, front-right), 1 (90 deg, bottom-front), 2 (150 deg, front-left), 5 (-30 deg, right)
		// Face Left-Front: topPts[1] -> topPts[2] -> botPts[2] -> botPts[1]
		// Face Right-Front: topPts[0] -> topPts[1] -> botPts[1] -> botPts[0]
		// Face Far-Right: topPts[5] -> topPts[0] -> botPts[0] -> botPts[5]
		faceLeft := []render.Point{topPts[1], topPts[2], botPts[2], botPts[1]}
		faceFront := []render.Point{topPts[0], topPts[1], botPts[1], botPts[0]}
		faceRight := []render.Point{topPts[5], topPts[0], botPts[0], botPts[5]}

		rc.FillPolygon(faceRight, rightWallCol, 1.0)
		rc.FillPolygon(faceFront, frontWallCol, 1.0)
		rc.FillPolygon(faceLeft, leftWallCol, 1.0)

		// Vertical basalt jointing striation lines along the prism walls
		joints := &canvas.Path{}
		for _, vIdx := range []int{0, 1, 2, 5} {
			joints.MoveTo(topPts[vIdx].X, topPts[vIdx].Y)
			joints.LineTo(botPts[vIdx].X, botPts[vIdx].Y)
		}
		rc.StrokePath(joints, 0.95*scale, canvas.RoundCap, canvas.RoundJoin, abyssBg, 0.78)

		// Top Hexagonal Cap with Recessed Beveled Inner Facet
		rc.FillPolygon(topPts, basaltDark, 1.0)
		inCap := render.InsetPolygon(topPts, 1.1*scale)
		rc.FillPolygon(inCap, capCol, 1.0)
		innerCore := render.InsetPolygon(topPts, 3.2*scale)
		if len(innerCore) >= 3 {
			rc.FillPolygon(innerCore, render.LerpOklab(capCol, magmaWhite, 0.24), 0.85)
		}

		capOutline := &canvas.Path{}
		capOutline.MoveTo(inCap[0].X, inCap[0].Y)
		for i := 1; i < len(inCap); i++ {
			capOutline.LineTo(inCap[i].X, inCap[i].Y)
		}
		capOutline.Close()
		rc.StrokePath(capOutline, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, rimStroke, 0.72)
	}

	rc.ApplyBloom(0.80, 6, 0.25)
	c.MergeRenderCanvas(rc)
}
