package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type mosaicLotus struct {
	petals     int
	brickScale float64
}

// NewMosaicLotus returns a mosaicLotus engine that renders a monumental low-poly
// stained-glass lotus mandala over concentric jade-and-turquoise pond ripples,
// sinuous golden koi flow ribbons, and faceted emerald lily pads.
func NewMosaicLotus(petals int, brickScale float64) *mosaicLotus {
	if petals <= 0 {
		petals = 12
	}
	if brickScale <= 0 {
		brickScale = 1.0
	}
	return &mosaicLotus{
		petals:     petals,
		brickScale: brickScale,
	}
}

// Generative renders the stained-glass lotus mandala, rippling pond courses, and faceted lily pads.
func (ml *mosaicLotus) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := (math.Min(w, h) / 600.0) * ml.brickScale
	if scale <= 0 {
		scale = 1.0
	}

	// Curated Stained-Glass Lotus & Jade Pond Palette
	deepPondNavy := color.RGBA{R: 16, G: 24, B: 42, A: 255}
	espressoGrout := color.RGBA{R: 38, G: 22, B: 28, A: 255}

	cobaltBlue := color.RGBA{R: 38, G: 106, B: 164, A: 255}
	jadeTeal := color.RGBA{R: 52, G: 156, B: 148, A: 255}
	mintAqua := color.RGBA{R: 144, G: 214, B: 188, A: 255}
	paleIce := color.RGBA{R: 216, G: 242, B: 230, A: 255}

	emeraldPad := color.RGBA{R: 36, G: 128, B: 92, A: 255}
	chartreuseLeaf := color.RGBA{R: 132, G: 192, B: 78, A: 255}
	deepForest := color.RGBA{R: 22, G: 76, B: 62, A: 255}

	rubyPlum := color.RGBA{R: 138, G: 24, B: 78, A: 255}
	vibrantMagenta := color.RGBA{R: 214, G: 38, B: 118, A: 255}
	coralRose := color.RGBA{R: 236, G: 92, B: 104, A: 255}
	blushPink := color.RGBA{R: 248, G: 172, B: 182, A: 255}
	cadmiumOrange := color.RGBA{R: 236, G: 122, B: 44, A: 255}
	goldenAmber := color.RGBA{R: 244, G: 182, B: 54, A: 255}
	sunYellow := color.RGBA{R: 252, G: 220, B: 88, A: 255}

	rc.FillBackground(deepPondNavy)
	rc.SetBlendMode(render.BlendOver)

	rng := rand.New(rand.NewSource(202610))

	lotusCX, lotusCY := 0.50*w, 0.48*h
	rLotusOuter := 0.29 * w
	brickThick := 9.2 * scale

	// Lotus star-envelope radius as a function of polar angle theta
	numPetals := ml.petals
	if numPetals < 6 {
		numPetals = 12
	}
	lotusBoundaryR := func(th float64) float64 {
		petalWave := math.Abs(math.Cos(float64(numPetals) * 0.5 * th))
		return rLotusOuter * (0.72 + 0.28*math.Pow(petalWave, 0.75))
	}

	// Two foreground lily-pad anchors (bottom-left and upper-right)
	pad1CX, pad1CY, rPad1 := 0.16*w, 0.84*h, 0.135*w
	pad2CX, pad2CY, rPad2 := 0.85*w, 0.16*h, 0.115*w

	// 1. Concentric Rippling Pond Mosaic Courses + Sinuous Golden Koi Flow Streams
	maxPondR := math.Hypot(w, h) * 0.82
	for ring := 0; ring < 34; ring++ {
		rBase := rLotusOuter*0.76 + (float64(ring)+0.5)*brickThick
		if rBase > maxPondR {
			break
		}
		circum := 2.0 * math.Pi * rBase
		brickLen := common.Constrain((20.0+4.0*math.Sin(float64(ring)*0.6))*scale, 15.0*scale, 31.0*scale)
		numBricks := int(math.Max(20, math.Round(circum/brickLen)))
		dTheta := (2.0 * math.Pi) / float64(numBricks)
		ringPhase := float64(ring) * 0.35

		for k := 0; k < numBricks; k++ {
			th0 := float64(k)*dTheta + ringPhase
			th1 := th0 + dTheta*0.96
			thMid := 0.5 * (th0 + th1)

			// Ripple modulation synced with the lotus petals near the center, smoothing outward
			petalSync := math.Exp(-float64(ring)*0.12) * 14.0 * scale * math.Abs(math.Cos(float64(numPetals)*0.5*thMid))
			wave0 := petalSync + 6.5*scale*math.Sin(4.0*th0-float64(ring)*0.32)
			wave1 := petalSync + 6.5*scale*math.Sin(4.0*th1-float64(ring)*0.32)

			p0 := render.Pt(lotusCX+(rBase+wave0)*math.Cos(th0), lotusCY+(rBase+wave0)*math.Sin(th0))
			p1 := render.Pt(lotusCX+(rBase+wave1)*math.Cos(th1), lotusCY+(rBase+wave1)*math.Sin(th1))
			mid := p0.Lerp(p1, 0.5)

			if mid.X < -30 || mid.X > w+30 || mid.Y < -30 || mid.Y > h+30 {
				continue
			}
			// Keep outside the lotus petal silhouette and lily pads
			if math.Hypot(mid.X-lotusCX, mid.Y-lotusCY) < lotusBoundaryR(thMid)-2.0*scale {
				continue
			}
			if math.Hypot(mid.X-pad1CX, mid.Y-pad1CY) < rPad1*0.98 || math.Hypot(mid.X-pad2CX, mid.Y-pad2CY) < rPad2*0.98 {
				continue
			}

			jitter := rng.Float64()
			// Sinuous golden-orange koi current ribbons sweeping in opposing arcs around the pond
			koiSpiral1 := math.Abs(math.Sin(thMid - float64(ring)*0.16 + 0.4))
			koiSpiral2 := math.Abs(math.Sin(thMid + float64(ring)*0.14 - 1.2))

			var col color.RGBA
			switch {
			case ring >= 2 && ring <= 14 && (koiSpiral1 < 0.16 || (ring >= 5 && koiSpiral2 < 0.14)):
				if jitter > 0.68 {
					col = sunYellow
				} else if jitter > 0.30 {
					col = cadmiumOrange
				} else {
					col = vibrantMagenta
				}
			case jitter < 0.09:
				col = deepPondNavy
			default:
				rippleVal := math.Sin(float64(ring)*0.75+thMid*3.0) + (jitter-0.5)*0.38
				switch {
				case rippleVal > 0.46:
					col = paleIce
				case rippleVal > 0.04:
					col = mintAqua
				case rippleVal > -0.40:
					col = jadeTeal
				default:
					col = cobaltBlue
				}
			}

			shear := 0.30 * math.Cos(thMid*3.0+float64(ring)*0.25)
			rc.DrawMosaicBrick(p0, p1, brickThick*0.48, shear, scale, col, 1.0)
		}
	}

	// 2. Helper to draw a Low-Poly Faceted Lily Pad with a V-Notch
	drawFacetedLilyPad := func(cx, cy, radius, notchAng float64) {
		const sectors = 14
		const rings = 3
		// Dark backing for crisp grout lines
		padBack := make([]render.Point, 0, sectors+2)
		padBack = append(padBack, render.Pt(cx, cy))
		for s := 0; s <= sectors; s++ {
			ang := notchAng + 0.28 + (float64(s)/float64(sectors))*(2.0*math.Pi-0.56)
			padBack = append(padBack, render.Pt(cx+radius*1.02*math.Cos(ang), cy+radius*1.02*math.Sin(ang)))
		}
		rc.FillPolygon(padBack, deepPondNavy, 1.0)

		pGrid := make([][]render.Point, rings+1)
		for ir := 0; ir <= rings; ir++ {
			pGrid[ir] = make([]render.Point, sectors+1)
			rCur := (float64(ir) / float64(rings)) * radius * 0.96
			for is := 0; is <= sectors; is++ {
				ang := notchAng + 0.28 + (float64(is)/float64(sectors))*(2.0*math.Pi-0.56)
				pGrid[ir][is] = render.Pt(cx+rCur*math.Cos(ang), cy+rCur*math.Sin(ang))
			}
		}
		padCols := []color.RGBA{emeraldPad, chartreuseLeaf, jadeTeal, deepForest, mintAqua}
		for ir := 0; ir < rings; ir++ {
			for is := 0; is < sectors; is++ {
				if ir == 0 {
					tri := render.InsetPolygon([]render.Point{
						render.Pt(cx, cy), pGrid[1][is], pGrid[1][is+1],
					}, 0.5*scale)
					rc.FillPolygon(tri, padCols[(is)%len(padCols)], 1.0)
				} else {
					q00, q10, q11, q01 := pGrid[ir][is], pGrid[ir+1][is], pGrid[ir+1][is+1], pGrid[ir][is+1]
					tri1 := render.InsetPolygon([]render.Point{q00, q10, q11}, 0.5*scale)
					tri2 := render.InsetPolygon([]render.Point{q00, q11, q01}, 0.5*scale)
					rc.FillPolygon(tri1, padCols[(ir+is)%len(padCols)], 1.0)
					rc.FillPolygon(tri2, padCols[(ir+is+1)%len(padCols)], 1.0)
				}
			}
		}
	}
	drawFacetedLilyPad(pad1CX, pad1CY, rPad1, -0.65)
	drawFacetedLilyPad(pad2CX, pad2CY, rPad2, 2.35)

	// 3. Monumental Low-Poly Stained-Glass Lotus Bloom at (lotusCX, lotusCY)
	// First lay down an espresso grout star backing behind all petals
	starSamples := numPetals * 8
	lotusBacking := make([]render.Point, starSamples)
	for i := 0; i < starSamples; i++ {
		th := (float64(i) / float64(starSamples)) * 2.0 * math.Pi
		r := lotusBoundaryR(th) * 1.01
		lotusBacking[i] = render.Pt(lotusCX+r*math.Cos(th), lotusCY+r*math.Sin(th))
	}
	rc.FillPolygon(lotusBacking, espressoGrout, 1.0)

	// Render Outer Petal Ring, Middle Petal Ring, and Inner Seed Core using low-poly diamond/triangle facets
	drawPetalRing := func(rInner, rOuter, phaseOffset float64, palette []color.RGBA) {
		dAng := (2.0 * math.Pi) / float64(numPetals)
		for p := 0; p < numPetals; p++ {
			thCenter := float64(p)*dAng + phaseOffset
			thLeft := thCenter - dAng*0.48
			thRight := thCenter + dAng*0.48

			rMid := rInner + (rOuter-rInner)*0.56
			rShoulder := rInner + (rOuter-rInner)*0.62

			pBase := render.Pt(lotusCX+rInner*math.Cos(thCenter), lotusCY+rInner*math.Sin(thCenter))
			pBaseL := render.Pt(lotusCX+rInner*math.Cos(thLeft), lotusCY+rInner*math.Sin(thLeft))
			pBaseR := render.Pt(lotusCX+rInner*math.Cos(thRight), lotusCY+rInner*math.Sin(thRight))
			pMid := render.Pt(lotusCX+rMid*math.Cos(thCenter), lotusCY+rMid*math.Sin(thCenter))
			pShL := render.Pt(lotusCX+rShoulder*math.Cos(thCenter-dAng*0.34), lotusCY+rShoulder*math.Sin(thCenter-dAng*0.34))
			pShR := render.Pt(lotusCX+rShoulder*math.Cos(thCenter+dAng*0.34), lotusCY+rShoulder*math.Sin(thCenter+dAng*0.34))
			pTip := render.Pt(lotusCX+rOuter*math.Cos(thCenter), lotusCY+rOuter*math.Sin(thCenter))

			// Subdivide each sculptured lotus petal into 6 crisp low-poly crystal shards
			facets := [][]render.Point{
				{pBase, pBaseL, pShL},
				{pBase, pShL, pMid},
				{pBase, pMid, pShR},
				{pBase, pShR, pBaseR},
				{pMid, pShL, pTip},
				{pMid, pTip, pShR},
			}
			for fIdx, fc := range facets {
				insetFc := render.InsetPolygon(fc, 0.58*scale)
				col := palette[(fIdx+p%2)%len(palette)]
				rc.FillPolygon(insetFc, col, 1.0)
			}
		}
	}

	// Outer petals: ruby plum, vibrant magenta, coral rose, blush pink, golden tips
	drawPetalRing(
		0.15*w, rLotusOuter*0.98, 0.0,
		[]color.RGBA{rubyPlum, vibrantMagenta, coralRose, rubyPlum, blushPink, goldenAmber},
	)
	// Inner offset petals: warm coral, cadmium orange, vibrant magenta, golden amber, sun yellow
	drawPetalRing(
		0.075*w, 0.195*w, math.Pi/float64(numPetals),
		[]color.RGBA{vibrantMagenta, cadmiumOrange, coralRose, vibrantMagenta, goldenAmber, sunYellow},
	)

	// Central Golden Lotus Seed Pod (r < 0.082w)
	rCore := 0.082 * w
	coreBack := make([]render.Point, 24)
	for i := 0; i < 24; i++ {
		ang := (float64(i) / 24.0) * 2.0 * math.Pi
		coreBack[i] = render.Pt(lotusCX+rCore*1.04*math.Cos(ang), lotusCY+rCore*1.04*math.Sin(ang))
	}
	rc.FillPolygon(coreBack, espressoGrout, 1.0)

	coreCols := []color.RGBA{sunYellow, goldenAmber, cadmiumOrange, paleIce}
	for p := 0; p < numPetals; p++ {
		th0 := (float64(p) / float64(numPetals)) * 2.0 * math.Pi
		th1 := (float64(p+1) / float64(numPetals)) * 2.0 * math.Pi
		thMid := 0.5 * (th0 + th1)

		pCenter := render.Pt(lotusCX, lotusCY)
		pInner0 := render.Pt(lotusCX+rCore*0.52*math.Cos(th0), lotusCY+rCore*0.52*math.Sin(th0))
		pInner1 := render.Pt(lotusCX+rCore*0.52*math.Cos(th1), lotusCY+rCore*0.52*math.Sin(th1))
		pOuter := render.Pt(lotusCX+rCore*0.95*math.Cos(thMid), lotusCY+rCore*0.95*math.Sin(thMid))

		t1 := render.InsetPolygon([]render.Point{pCenter, pInner0, pInner1}, 0.48*scale)
		t2 := render.InsetPolygon([]render.Point{pInner0, pOuter, pInner1}, 0.48*scale)
		rc.FillPolygon(t1, coreCols[p%2], 1.0)
		rc.FillPolygon(t2, coreCols[(p+1)%len(coreCols)], 1.0)
	}

	c.MergeRenderCanvas(rc)
}
