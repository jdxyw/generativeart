package arts

import (
	"image/color"
	"math"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type hopfFibration struct {
	toriN          int
	fibersPerTorus int
}

// NewHopfFibration returns a hopfFibration engine that renders the 3D stereographic
// projection of the S^3 -> S^2 Hopf fibration into linked Villarceau circles sweeping
// nested Clifford tori with true 3D depth-sorted occlusion and tubular shading.
func NewHopfFibration(toriN, fibersPerTorus int) *hopfFibration {
	if toriN <= 0 {
		toriN = 4
	}
	if fibersPerTorus <= 0 {
		fibersPerTorus = 24
	}
	return &hopfFibration{
		toriN:          toriN,
		fibersPerTorus: fibersPerTorus,
	}
}

// Generative renders the 3D stereographic Hopf fibration of linked Villarceau circles.
func (hf *hopfFibration) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	cx, cy := 0.5*fw, 0.5*fh

	abyssSpace := color.RGBA{R: 6, G: 8, B: 18, A: 255}
	ivorySpec := color.RGBA{R: 255, G: 250, B: 230, A: 255}
	shadowCol := color.RGBA{R: 2, G: 4, B: 10, A: 255}

	// Curated Torus Shell Palettes (Inner core -> Middle shells -> Outer veil)
	torusPals := [][2]color.RGBA{
		{color.RGBA{252, 198, 62, 255}, color.RGBA{218, 68, 42, 255}},  // Shell 0 (Inner): Imperial Gold & Cinnabar
		{color.RGBA{226, 58, 96, 255}, color.RGBA{148, 34, 156, 255}},  // Shell 1: Coral & Byzantium Magenta
		{color.RGBA{44, 196, 188, 255}, color.RGBA{26, 98, 186, 255}},  // Shell 2: Jingtai Turquoise & Cerulean
		{color.RGBA{68, 146, 238, 255}, color.RGBA{24, 48, 128, 255}},  // Shell 3 (Outer): Sapphire & Lapis
	}

	rc.FillBackground(abyssSpace)
	rc.SetBlendMode(render.BlendOver)

	// Subtle radial cosmic background glow
	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			r := math.Hypot(fx-cx, fy-cy) / (0.52 * fw)
			if r < 1.0 {
				bg := render.LerpOklab(color.RGBA{R: 16, G: 26, B: 56, A: 255}, abyssSpace, math.Pow(r, 1.35))
				rc.BlendPixel(x, y, bg, 1.0)
			}
		}
	}

	// 3D Camera Rotation (tilted to look into the nested Villarceau torus vortex)
	rotX, rotY, rotZ := 0.68, -0.44, 0.25
	cxR, sxR := math.Cos(rotX), math.Sin(rotX)
	cyR, syR := math.Cos(rotY), math.Sin(rotY)
	czR, szR := math.Cos(rotZ), math.Sin(rotZ)

	rotate3D := func(x, y, z float64) (rx, ry, rz float64) {
		// Rotate around Z
		x1 := x*czR - y*szR
		y1 := x*szR + y*czR
		// Rotate around X
		y2 := y1*cxR - z*sxR
		z2 := y1*sxR + z*cxR
		// Rotate around Y
		rx = x1*cyR + z2*syR
		ry = y2
		rz = -x1*syR + z2*cyR
		return
	}

	// Exact Stereographic Hopf Fibration Map:
	// A point on S^2 with polar angle theta in (0, pi) and azimuth phi in [0, 2*pi)
	// lifts to the fiber circle in S^3:
	//   z1 = cos(theta/2) * exp(i * t)
	//   z2 = sin(theta/2) * exp(i * (t + phi))
	// Stereographic projection from S^3 \ {(0,0,0,1)} to R^3:
	//   (X, Y, Z) = (Re(z1), Im(z1), Re(z2)) / (1 - Im(z2))
	hopfPoint3D := func(theta, phi, t float64) (rx, ry, rz float64) {
		a := math.Cos(0.5 * theta)
		b := math.Sin(0.5 * theta)
		x4 := a * math.Cos(t)
		y4 := a * math.Sin(t)
		z4 := b * math.Cos(t+phi)
		w4 := b * math.Sin(t+phi)

		denom := 1.0 - w4*0.76 // bounded conformal stereographic factor so outer fibers stay compact
		if denom < 0.15 {
			denom = 0.15
		}
		return rotate3D(x4/denom, y4/denom, z4/denom)
	}

	type fiberSegment struct {
		p0, p1, p2 render.Point
		zDepth     float64
		width      float64
		baseCol    color.RGBA
		specCol    color.RGBA
		alpha      float64
	}

	numTori := hf.toriN
	if numTori < 2 || numTori > 6 {
		numTori = 4
	}
	baseFibers := hf.fibersPerTorus
	if baseFibers < 10 {
		baseFibers = 24
	}

	segments := make([]fiberSegment, 0, numTori*baseFibers*48)
	projScale := 138.0 * scale
	const fov = 4.2

	// Generate linked Villarceau fiber circles across nested Clifford tori
	for tIdx := 0; tIdx < numTori; tIdx++ {
		uTorus := float64(tIdx) / float64(common.MaxInt(1, numTori-1))
		// Polar angle theta on S^2 from inner core (0.28*pi) to outer shell (0.64*pi)
		theta := (0.26 + 0.38*uTorus) * math.Pi
		nFibers := baseFibers + tIdx*4
		pal := torusPals[tIdx%len(torusPals)]

		const curveSteps = 64
		for f := 0; f < nFibers; f++ {
			phi := (float64(f) / float64(nFibers)) * 2.0 * math.Pi
			fColorMix := 0.5 + 0.5*math.Sin(phi*2.0+float64(tIdx)*0.8)
			fiberCol := render.LerpOklab(pal[0], pal[1], fColorMix)

			pts2D := make([]render.Point, curveSteps+1)
			zCoords := make([]float64, curveSteps+1)

			for s := 0; s <= curveSteps; s++ {
				t := (float64(s) / float64(curveSteps)) * 2.0 * math.Pi
				x3, y3, z3 := hopfPoint3D(theta, phi, t)
				persp := fov / math.Max(1.2, fov-z3*0.55)
				pts2D[s] = render.Pt(cx+x3*projScale*persp, cy+y3*projScale*persp)
				zCoords[s] = z3
			}

			// Break each closed Villarceau circle into overlapping 2-step sub-segments for exact 3D depth sorting
			for s := 0; s < curveSteps; s += 2 {
				s1 := (s + 1) % curveSteps
				s2 := (s + 2) % curveSteps
				zMid := zCoords[s1]
				depthNorm := common.Constrain((zMid+1.6)/3.2, 0, 1)

				segCol := fiberCol
				if depthNorm < 0.45 {
					segCol = render.LerpOklab(abyssSpace, fiberCol, 0.32+0.68*(depthNorm/0.45))
				} else {
					segCol = render.LerpOklab(fiberCol, ivorySpec, (depthNorm-0.45)*0.32)
				}

				baseW := (4.8 - 0.55*float64(tIdx)) * (0.68 + 0.52*depthNorm) * scale
				segments = append(segments, fiberSegment{
					p0:      pts2D[s],
					p1:      pts2D[s1],
					p2:      pts2D[s2],
					zDepth:  zMid,
					width:   baseW,
					baseCol: segCol,
					specCol: render.LerpOklab(segCol, ivorySpec, 0.68),
					alpha:   0.86 + 0.12*depthNorm,
				})
			}
		}
	}

	// Sort all 3D fiber segments from back (smallest zDepth) to front (largest zDepth)
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].zDepth < segments[j].zDepth
	})

	lightDir := render.Pt(-0.62, -0.74)
	for _, seg := range segments {
		subPts := []render.Point{seg.p0, seg.p1, seg.p2}
		// Subtle dark occlusion casing so crossing fibers cast authentic 3D over/under shadows
		rc.StrokePolyline([]render.Point{
			render.Pt(seg.p0.X+1.2*scale, seg.p0.Y+1.6*scale),
			render.Pt(seg.p1.X+1.2*scale, seg.p1.Y+1.6*scale),
			render.Pt(seg.p2.X+1.2*scale, seg.p2.Y+1.6*scale),
		}, seg.width*1.32, shadowCol, 0.52)

		rc.DrawImpastoStroke(subPts, seg.width, seg.baseCol, lightDir, seg.alpha)
		rc.StrokePolyline(subPts, seg.width*0.28, seg.specCol, seg.alpha*0.85)
	}

	rc.ApplyBloom(0.82, 6, 0.24)
	c.MergeRenderCanvas(rc)
}
