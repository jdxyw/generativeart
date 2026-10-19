package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type phyllotaxis struct {
	count       int
	goldenAngle float64
	scale       float64
}

// NewPhyllotaxis returns a phyllotaxis engine that renders a sculptural 3D golden-angle
// Vogel spiral botanical lotus / succulent rosette with ambient occlusion shadows and Oklab blush petals.
func NewPhyllotaxis(count int, goldenAngle, scale float64) *phyllotaxis {
	if count <= 0 {
		count = 420
	}
	if goldenAngle <= 0 {
		goldenAngle = 137.50776405
	}
	if scale <= 0 {
		scale = 11.5
	}
	return &phyllotaxis{
		count:       count,
		goldenAngle: goldenAngle,
		scale:       scale,
	}
}

// buildLotusPetalGeometry constructs a curved teardrop/lotus succulent petal starting near basePos,
// oriented along radial angle theta with organic tip curl.
// It returns:
//   - poly: closed petal polygon contour
//   - spine: central midrib curve from base (u=0) to acuminate mucro tip (u=1)
//   - leftEdge, rightEdge: left and right petal rim polylines
//   - lateralVeins: 2 pairs of branching lateral veins from the midrib toward the margins
func buildLotusPetalGeometry(
	basePos render.Point,
	theta, length, halfWidth, curl float64,
	scaleFactor, tipPull float64,
) (
	poly []render.Point,
	spine []render.Point,
	leftEdge []render.Point,
	rightEdge []render.Point,
	lateralVeins [][]render.Point,
) {
	const steps = 16
	spine = make([]render.Point, steps+1)
	leftEdge = make([]render.Point, steps+1)
	rightEdge = make([]render.Point, steps+1)

	for s := 0; s <= steps; s++ {
		u := float64(s) / float64(steps)
		// Tip pull shifts inner concentric layers toward the upper/outer tip of the petal
		uEff := tipPull + (1.0-tipPull)*u

		localAng := theta + curl*uEff*uEff
		dir := render.Pt(math.Cos(localAng), math.Sin(localAng))
		perp := dir.Perp()

		centerPt := basePos.Add(dir.Mul(uEff * length))
		spine[s] = centerPt

		// Plump echeveria/lotus width profile: narrow base, full belly at u=0.58, acuminate mucro tip at u=1
		var wEnv float64
		if u < 0.999 {
			belly := math.Sin(math.Pi * math.Pow(u, 0.72))
			shoulder := 1.0 + 0.24*math.Sin(math.Pi*u)*math.Sin(0.5*math.Pi*u)
			baseTab := 0.16 * (1.0 - u) * (1.0 - u)
			wEnv = (belly*shoulder + baseTab) * scaleFactor
		}
		hw := halfWidth * wEnv

		leftEdge[s] = centerPt.Add(perp.Mul(hw))
		rightEdge[s] = centerPt.Sub(perp.Mul(hw))
	}

	poly = make([]render.Point, 0, 2*(steps+1))
	for s := 0; s <= steps; s++ {
		poly = append(poly, leftEdge[s])
	}
	for s := steps - 1; s >= 0; s-- {
		poly = append(poly, rightEdge[s])
	}

	// Construct 2 pairs of curved lateral veins branching from the midrib toward the margins
	veinAnchors := [2][2]int{{4, 10}, {7, 13}}
	lateralVeins = make([][]render.Point, 0, 4)
	for _, va := range veinAnchors {
		s0, s1 := va[0], va[1]
		if s1 <= steps {
			midS := (s0 + s1) / 2
			lateralVeins = append(lateralVeins,
				[]render.Point{
					spine[s0],
					spine[midS].Lerp(leftEdge[midS], 0.52),
					spine[s1].Lerp(leftEdge[s1], 0.85),
				},
				[]render.Point{
					spine[s0],
					spine[midS].Lerp(rightEdge[midS], 0.52),
					spine[s1].Lerp(rightEdge[s1], 0.85),
				},
			)
		}
	}

	return poly, spine, leftEdge, rightEdge, lateralVeins
}

// offsetPolygon translates all vertices of poly by delta.
func offsetPolygon(poly []render.Point, delta render.Point) []render.Point {
	out := make([]render.Point, len(poly))
	for i, p := range poly {
		out[i] = p.Add(delta)
	}
	return out
}

// Generative renders the Sculptural 3D Golden-Angle Phyllotaxis Botanical Lotus / Succulent Rosette.
func (ph *phyllotaxis) Generative(c *generativeart.Canva) {
	count := ph.count
	goldenAngle := ph.goldenAngle
	baseScale := ph.scale
	if count <= 0 {
		count = 420
	}
	if goldenAngle <= 0 {
		goldenAngle = 137.50776405
	}
	if baseScale <= 0 {
		baseScale = 11.5
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	canvasScale := fw / 600.0
	if canvasScale <= 0 {
		canvasScale = 1.0
	}
	effScale := baseScale * canvasScale

	seed := int64(9991618)
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Botanical
	}
	grad := render.NewMultiStopGradient(schema...)
	sunsetGrad := render.NewMultiStopGradient(common.Sunset...)

	rc := c.NewRenderCanvas()

	// Deep botanical forest-moss & midnight pond backdrop
	bgCol := render.LerpOklab(color.RGBA{R: 12, G: 20, B: 18, A: 255}, grad.SampleOklab(0.08), 0.35)
	rc.FillBackground(bgCol)

	cx, cy := fw*0.5, fh*0.5
	center := render.Pt(cx, cy)

	// Soft ambient pond halo behind the rosette
	rc.SetBlendMode(render.BlendAdd)
	maxRadius := effScale * math.Sqrt(float64(count))
	rc.DrawGlowDot(cx, cy, maxRadius*1.08, grad.SampleOklab(0.35), 0.14, 0.0)
	rc.SetBlendMode(render.BlendOver)

	goldenRad := goldenAngle * math.Pi / 180.0
	lightDir := render.Pt(-0.55, -0.83).Norm()

	shadowDark := color.RGBA{R: 6, G: 10, B: 12, A: 255}
	ivoryHighlight := color.RGBA{R: 255, G: 250, B: 238, A: 255}
	pollenGold := color.RGBA{R: 255, G: 218, B: 132, A: 255}

	// Render primordia from outer/bottom petals (i = count-1) down to inner/top core (i = 0)
	// so inner petals naturally overlap outer ones in true 3D botanical order.
	for i := count - 1; i >= 0; i-- {
		fi := float64(i)
		sqrtI := math.Sqrt(fi)
		uNorm := common.Clamp01(sqrtI / math.Sqrt(float64(common.MaxInt(1, count-1))))

		theta := fi * goldenRad
		r := effScale * sqrtI

		// Subtle FBM organic perturbation to radial angle & tip curl
		pxNorm := math.Cos(theta) * uNorm
		pyNorm := math.Sin(theta) * uNorm
		fbm := noise.FBM2D(pxNorm*2.4+3.1, pyNorm*2.4+7.3, 3, 2.0, 0.5)
		curl := 0.16 * fbm * (0.35 + 0.65*uNorm)

		// Base attachment point of petal i
		radDir := render.Pt(math.Cos(theta), math.Sin(theta))
		basePos := center.Add(radDir.Mul(r * 0.84))

		// Petal length and half-width scale smoothly with sqrt(i) so concentric whorls remain visible
		petalLen := (10.5 + 2.95*sqrtI) * canvasScale
		petalHalfW := (4.2 + 1.18*sqrtI) * canvasScale

		poly, spine, leftEdge, rightEdge, latVeins := buildLotusPetalGeometry(
			basePos, theta, petalLen, petalHalfW, curl, 1.0, 0.0,
		)
		if len(poly) < 6 {
			continue
		}

		// 1. Soft dark ambient-occlusion drop shadow cast outward & away from upper-left light source
		shadowShift := radDir.Mul(3.8*canvasScale*(0.35+0.65*uNorm)).Sub(lightDir.Mul(2.6 * canvasScale))
		rc.SetBlendMode(render.BlendMultiply)
		shadowPolyOuter := offsetPolygon(poly, shadowShift.Mul(1.35))
		shadowPolyInner := offsetPolygon(poly, shadowShift.Mul(0.75))
		rc.FillPolygon(shadowPolyOuter, shadowDark, 0.28)
		rc.FillPolygon(shadowPolyInner, shadowDark, 0.42)

		// 2. Compute a rich multi-hue Oklab botanical palette for petal i:
		// Radial gradient across the rosette (uNorm: 0 at glowing lotus heart -> 1 at deep emerald/teal outer leaves)
		// combined with per-petal base -> jade/turquoise belly -> ruby/coral/gold mucro tip blush.
		illum := -(radDir.X*lightDir.X + radDir.Y*lightDir.Y)
		deepForest := color.RGBA{R: 12, G: 44, B: 38, A: 255}
		tealJade := color.RGBA{R: 34, G: 118, B: 102, A: 255}
		mintSage := color.RGBA{R: 106, G: 178, B: 128, A: 255}
		crimsonRose := color.RGBA{R: 218, G: 48, B: 84, A: 255}
		coralGold := color.RGBA{R: 248, G: 152, B: 58, A: 255}

		// Inner core warms toward coral-gold & lotus rose; outer whorls are rich teal-jade & mint-sage
		innerWarmth := math.Pow(1.0-uNorm, 2.1)
		bellyJade := render.LerpOklab(tealJade, mintSage, common.Clamp01(1.0-0.75*uNorm+0.15*fbm))
		midBodyCol := render.LerpOklab(
			render.LerpOklab(bellyJade, grad.SampleOklab(common.Clamp01(0.18+0.42*(1.0-uNorm))), 0.35),
			coralGold,
			innerWarmth*0.72,
		)
		deepBaseCol := render.LerpOklab(
			deepForest,
			midBodyCol,
			0.28+0.25*innerWarmth,
		)
		tipBlushCol := render.LerpOklab(
			crimsonRose,
			sunsetGrad.SampleOklab(common.Clamp01(0.42+0.52*(1.0-uNorm)+0.08*fbm)),
			0.38+0.32*innerWarmth,
		)

		// Modulate lightness slightly by directional lighting while keeping rich saturation
		midLab := render.RGBAToOklab(midBodyCol)
		midLab.L = common.Constrain(midLab.L+illum*0.055, 0.18, 0.80)
		midBodyCol = render.OklabToRGBA(midLab, 255)

		// 3. Fill the petal body with concentric layered polygons transitioning in Oklab space:
		// Since inner whorls overlap u = 0..0.55 of outer petals, keep u = 0.50..0.82 rich emerald/teal-jade,
		// and focus the ruby/coral/gold blush specifically on the mucro tip (u = 0.76..1.0)!
		rc.SetBlendMode(render.BlendOver)
		rc.FillPolygon(poly, deepBaseCol, 0.96)

		const layers = 5
		for l := 1; l <= layers; l++ {
			lt := float64(l) / float64(layers)
			var scaleW, tipPull float64
			var layerCol color.RGBA
			if l <= 3 {
				// Layers 1..3 sculpt the exposed emerald/jade/gold belly of the petal
				bt := float64(l) / 3.0
				scaleW = 1.0 - 0.28*bt
				tipPull = 0.18 * bt
				layerCol = render.LerpOklab(deepBaseCol, midBodyCol, bt)
			} else if l == 4 {
				// Layer 4 transitions from jade belly into warm coral-rose near the upper third
				scaleW = 0.56
				tipPull = 0.64
				layerCol = render.LerpOklab(midBodyCol, tipBlushCol, 0.52)
			} else {
				// Layer 5 blushes the acuminate mucro tip with ruby/coral/gold
				scaleW = 0.38
				tipPull = 0.76
				layerCol = tipBlushCol
			}
			lPoly, _, _, _, _ := buildLotusPetalGeometry(
				basePos, theta, petalLen*(1.0-0.01*lt), petalHalfW, curl, scaleW, tipPull,
			)
			rc.FillPolygon(lPoly, layerCol, 0.86)
		}

		// Subtle wet-on-wet watercolor blush near the upper petal body
		if i%2 == 0 && petalLen > 18.0*canvasScale {
			rc.SetBlendMode(render.BlendMultiply)
			washPoly, _, _, _, _ := buildLotusPetalGeometry(
				basePos, theta, petalLen*0.95, petalHalfW, curl, 0.78, 0.18,
			)
			rc.DrawWatercolorPolygon(washPoly, midBodyCol, 3, 2, 0.12, 0.04, rng)
			rc.SetBlendMode(render.BlendOver)
		}

		// 4. Draw fine central & lateral petal veins and a crisp crimson-gold rim highlight using rc.DrawTaperedSpline
		veinCol := render.LerpOklab(midBodyCol, pollenGold, 0.42)
		tipRimCol := render.LerpOklab(tipBlushCol, ivoryHighlight, 0.28)

		// Central midrib vein
		rc.DrawTaperedSpline(
			spine,
			1.45*canvasScale*(0.55+0.45*uNorm),
			render.TaperTail,
			func(t float64) color.RGBA {
				return render.LerpOklab(deepBaseCol, veinCol, common.Clamp01(t*1.15))
			},
			0.48,
		)

		// Lateral branching veins
		for _, lv := range latVeins {
			rc.DrawTaperedSpline(
				lv,
				0.78*canvasScale*(0.5+0.5*uNorm),
				render.TaperBothEnds,
				func(t float64) color.RGBA {
					return render.LerpOklab(midBodyCol, veinCol, t)
				},
				0.30,
			)
		}

		// Crisp crimson-blush & gold rim highlights along left and right margins
		rimWidth := (0.95 + 0.85*uNorm) * canvasScale
		rc.DrawTaperedSpline(
			leftEdge,
			rimWidth,
			render.TaperBothEnds,
			func(t float64) color.RGBA {
				return render.LerpOklab(midBodyCol, tipRimCol, math.Pow(t, 0.55))
			},
			0.74,
		)
		rc.DrawTaperedSpline(
			rightEdge,
			rimWidth*0.85,
			render.TaperBothEnds,
			func(t float64) color.RGBA {
				return render.LerpOklab(deepBaseCol, tipRimCol, math.Pow(t, 0.65))
			},
			0.62,
		)

		// 5. Subtle glowing dewdrop / mucro pearl at the petal tip
		tipPt := spine[len(spine)-1]
		subTipPt := spine[len(spine)-2].Lerp(tipPt, 0.55)
		rc.SetBlendMode(render.BlendAdd)
		pearlRadius := (1.4 + 1.4*(1.0-0.45*uNorm)) * canvasScale
		pearlCol := render.LerpOklab(tipBlushCol, pollenGold, 0.42+0.35*(1.0-uNorm))
		rc.DrawGlowDot(subTipPt.X, subTipPt.Y, pearlRadius*1.35, pearlCol, 0.22, 0.16)
		rc.DrawGlowDot(tipPt.X, tipPt.Y, pearlRadius*0.62, ivoryHighlight, 0.38, 0.45)
		rc.SetBlendMode(render.BlendOver)
	}

	// Luminous golden stamen/nectar core glow at the exact rosette center
	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(cx, cy, 16.0*canvasScale, pollenGold, 0.32, 0.15)
	rc.DrawGlowDot(cx, cy, 5.0*canvasScale, ivoryHighlight, 0.65, 0.50)
	rc.SetBlendMode(render.BlendOver)

	rc.ApplyBloom(0.68, 6.0, 0.20)
	rc.ApplyPaperTexture(0.22, 999)
	rc.ApplyVignette(0.28)
	c.MergeRenderCanvasToneMapped(rc, 1.04, 0)
}
