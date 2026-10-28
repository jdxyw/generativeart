package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type hexMandala struct {
	cols       int
	rows       int
	colorScale float64
}

// NewHexMandala returns a hexMandala engine that renders a dense flat-topped hexagonal
// honeycomb tessellation of intricate 6-fold and 12-fold procedural mandalas modulated
// by a luminous cyan-turquoise nebula versus burnished gold-bronze-umber macro field.
func NewHexMandala(cols, rows int, colorScale float64) *hexMandala {
	if cols <= 0 {
		cols = 13
	}
	if rows <= 0 {
		rows = 16
	}
	if colorScale <= 0 {
		colorScale = 1.0
	}
	return &hexMandala{
		cols:       cols,
		rows:       rows,
		colorScale: colorScale,
	}
}

// Generative renders the hexagonal mandala honeycomb tessellation.
func (hm *hexMandala) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	fw, fh := float64(c.Width()), float64(c.Height())
	scale := math.Min(fw, fh) / 600.0

	obsidianBg := color.RGBA{R: 7, G: 6, B: 5, A: 255}
	hexVoid := color.RGBA{R: 11, G: 9, B: 8, A: 255}
	darkInk := color.RGBA{R: 9, G: 8, B: 7, A: 255}

	rc.FillBackground(obsidianBg)
	rc.SetBlendMode(render.BlendOver)

	// Cyan / Turquoise / Cerulean Nebula Palette Stops
	deepTeal := color.RGBA{R: 12, G: 60, B: 76, A: 255}       // #0c3c4c
	cerulean := color.RGBA{R: 24, G: 120, B: 152, A: 255}     // #187898
	electricTurq := color.RGBA{R: 56, G: 189, B: 214, A: 255} // #38bdd6
	paleIceCyan := color.RGBA{R: 140, G: 232, B: 248, A: 255} // #8ce8f8
	amberCore := color.RGBA{R: 234, G: 162, B: 40, A: 255}    // #eaa228

	// Gold / Bronze / Umber Palette Stops
	darkEspresso := color.RGBA{R: 30, G: 16, B: 8, A: 255}     // #1e1008
	deepBronze := color.RGBA{R: 94, G: 44, B: 16, A: 255}      // #5e2c10
	copperRust := color.RGBA{R: 168, G: 82, B: 28, A: 255}     // #a8521c
	burnishedGold := color.RGBA{R: 226, G: 152, B: 40, A: 255} // #e29828
	champagneGold := color.RGBA{R: 250, G: 212, B: 107, A: 255} // #fad46b

	cyanGrad := render.NewMultiStopGradient(deepTeal, cerulean, electricTurq, paleIceCyan)
	goldGrad := render.NewMultiStopGradient(darkEspresso, deepBronze, copperRust, burnishedGold, champagneGold)

	noise := common.NewNoiseEngine(4219)
	cs := hm.colorScale

	hexR := fw / (float64(hm.cols) * 1.5)
	dx := 1.5 * hexR
	dy := math.Sqrt(3.0) * hexR
	maxRow := common.MaxInt(hm.rows, int(math.Ceil(fh/dy)))

	scaleColorLum := func(col color.RGBA, lum float64) color.RGBA {
		lab := render.RGBAToOklab(col)
		lab.L = common.Constrain(lab.L*lum, 0.04, 0.96)
		if lum > 1.0 {
			lab.A *= 0.94
			lab.B *= 0.94
		}
		return render.OklabToRGBA(lab, 255)
	}

	buildHexPts := func(cx, cy, r, rot float64) []render.Point {
		pts := make([]render.Point, 6)
		for k := 0; k < 6; k++ {
			ang := float64(k)*(math.Pi/3.0) + rot
			pts[k] = render.Pt(cx+r*math.Cos(ang), cy+r*math.Sin(ang))
		}
		return pts
	}

	strokeClosedPoly := func(pts []render.Point, width float64, col color.RGBA, alpha float64) {
		if len(pts) < 3 {
			return
		}
		p := &canvas.Path{}
		p.MoveTo(pts[0].X, pts[0].Y)
		for i := 1; i < len(pts); i++ {
			p.LineTo(pts[i].X, pts[i].Y)
		}
		p.Close()
		rc.StrokePath(p, width, canvas.RoundCap, canvas.RoundJoin, col, alpha)
	}

	fillAndStrokeCircle := func(x, y, r float64, fillCol, strokeCol color.RGBA, strokeW, alpha float64) {
		if r <= 0 {
			return
		}
		rc.DrawGlowDot(x, y, r, fillCol, alpha, 0.96)
		if strokeW > 0 && strokeCol.A > 0 {
			rc.DrawSoftRing(x, y, r, strokeW, strokeCol, alpha)
		}
	}

	for col := -1; col <= hm.cols+1; col++ {
		for row := -1; row <= maxRow+1; row++ {
			oddCol := 0.0
			if col&1 != 0 {
				oddCol = 0.5 * dy
			}
			cx := float64(col) * dx
			cy := float64(row)*dy + oddCol

			// Skip cells completely outside the canvas bounds
			if cx+hexR < -2 || cx-hexR > fw+2 || cy+hexR < -2 || cy-hexR > fh+2 {
				continue
			}

			nx := cx / fw
			ny := cy / fh

			// Deterministic per-cell PRNG so each hexagon has a unique sacred geometry mandala
			cellSeed := int64(col*73856093 ^ row*19349663 ^ 0x5A3C9E17)
			rng := rand.New(rand.NewSource(cellSeed))

			// Macro Color & Luminance Field matching Reference Image 4
			fbm1 := noise.FBM2D(nx*2.5*cs, ny*2.5*cs, 3, 2.0, 0.5)
			fbm2 := noise.FBM2D((nx+3.7)*3.2*cs, (ny+1.4)*3.2*cs, 3, 2.0, 0.5)

			// Primary left-center Cyan/Turquoise nebula around (0.36, 0.46) extending to (0.28, 0.78)
			distCyanMain := math.Hypot((nx-0.36)*1.25, (ny-0.46)*0.95) - 0.18*fbm1
			distCyanLower := math.Hypot((nx-0.29)*1.42, (ny-0.75)*1.12) - 0.15*fbm2
			distCyanUpper := math.Hypot((nx-0.25)*1.55, (ny-0.22)*1.30) - 0.12*fbm1
			distCyan := math.Min(distCyanMain, math.Min(distCyanLower, distCyanUpper))
			cyanWeight := common.Smoothstep(0.46, 0.10, distCyan)

			// Secondary cyan pockets in the bottom-right & mid-right
			if nx > 0.55 && ny > 0.58 {
				distPocket := math.Hypot((nx-0.77)*1.48, (ny-0.80)*1.38) - 0.16*fbm2
				cyanWeight = math.Max(cyanWeight, 0.86*common.Smoothstep(0.32, 0.08, distPocket))
			}

			// Per-cell stochastic mix jitter so adjacent hexagons have jewel-like tapestry variation
			cellMix := common.Clamp01(cyanWeight + (rng.Float64()-0.48)*0.26)

			// Per-cell brightness modulation in [0.38, 1.08] with darker right-edge vignette
			baseLum := 0.42 + 0.64*math.Pow(rng.Float64(), 0.88)
			// Boost luminance in the glowing cyan core and upper-center gold crown
			nebulaGlow := common.Smoothstep(0.38, 0.05, distCyanMain)*0.22 +
				common.Smoothstep(0.42, 0.08, math.Hypot((nx-0.54)*1.1, (ny-0.24)*1.1))*0.16
			rightVignette := 1.0 - 0.36*common.Smoothstep(0.80, 1.02, nx)
			edgeVignette := 1.0 - 0.18*common.Smoothstep(0.88, 1.05, math.Hypot(nx-0.5, ny-0.5)*1.45)
			cellLum := common.Constrain((baseLum+nebulaGlow)*rightVignette*edgeVignette, 0.36, 1.12)

			// Sample a 5-stop cell palette blending Gold/Bronze and Cyan/Turquoise
			sampleCellColor := func(t float64, accentShift float64) color.RGBA {
				t = common.Clamp01(t)
				gCol := goldGrad.SampleOklab(t)
				cCol := cyanGrad.SampleOklab(t)
				mix := common.Clamp01(cellMix + accentShift)
				blended := render.LerpOklab(gCol, cCol, mix)
				return scaleColorLum(blended, cellLum)
			}

			// 1. Outer Flat-Topped Hexagon Base & Frame
			hexVerts := buildHexPts(cx, cy, hexR*0.985, 0.0)
			rc.FillPolygon(hexVerts, hexVoid, 1.0)

			rimCol := sampleCellColor(0.32, 0.0)
			strokeClosedPoly(hexVerts, 1.05*scale, rimCol, 0.88)

			innerRimVerts := buildHexPts(cx, cy, hexR*0.935, 0.0)
			strokeClosedPoly(innerRimVerts, 0.7*scale, sampleCellColor(0.52, 0.0), 0.65)

			// Clip all mandala geometry strictly to this hexagon
			rc.PushClipPolygon(hexVerts)

			// Corner spandrel micro-dots inside the 6 hexagon vertices (outside the circular mandala rings)
			if rng.Float64() < 0.78 {
				spCol := sampleCellColor(0.62, 0.0)
				for k := 0; k < 6; k++ {
					ang := float64(k) * (math.Pi / 3.0)
					px := cx + hexR*0.84*math.Cos(ang)
					py := cy + hexR*0.84*math.Sin(ang)
					fillAndStrokeCircle(px, py, hexR*0.042, spCol, darkInk, 0.6*scale, 0.92)
				}
			}

			// 2. Five Concentric Radial Mandala Bands from r = 0.90*hexR down to r = 0.14*hexR
			bandRadii := [5]float64{
				0.89 * hexR,
				0.71 * hexR,
				0.54 * hexR,
				0.38 * hexR,
				0.23 * hexR,
			}

			for b := 0; b < 5; b++ {
				r := bandRadii[b]
				rNext := 0.11 * hexR
				if b < 4 {
					rNext = bandRadii[b+1]
				}

				// Contrast-alternating palette stops across concentric bands
				tPrimary := common.Clamp01(0.28 + 0.16*float64((b*2+rng.Intn(2))%5))
				tSecondary := common.Clamp01(0.45 + 0.14*float64((b+2)%4))
				// Occasional complementary gold-in-cyan or cyan-in-gold accent band
				accentShift := 0.0
				if b >= 2 && rng.Float64() < 0.28 {
					if cellMix > 0.5 {
						accentShift = -0.65
					} else {
						accentShift = 0.65
					}
				}
				colMain := sampleCellColor(tPrimary, accentShift)
				colBright := sampleCellColor( math.Min(1.0, tPrimary+0.26), accentShift)
				colDeep := sampleCellColor(math.Max(0.08, tSecondary-0.22), 0.0)

				// Dark separator disc behind each band for crisp engraving contrast
				rc.DrawGlowDot(cx, cy, r*1.02, hexVoid, 0.96, 0.98)

				motif := (b + rng.Intn(5)) % 5
				switch motif {
				case 0:
					// Motif 1: Nested Hexagonal / Star-of-David {6/2} Hexagram Frame
					hexA := buildHexPts(cx, cy, r, 0.0)
					hexB := buildHexPts(cx, cy, r, math.Pi/6.0)
					rc.FillPolygon(hexA, colDeep, 0.95)
					strokeClosedPoly(hexA, 1.15*scale, colBright, 0.95)
					rc.FillPolygon(hexB, colMain, 0.88)
					strokeClosedPoly(hexB, 1.05*scale, colBright, 0.95)

					// 12-point star vertex ticks connecting the two rotated triangles/hexagons
					for k := 0; k < 12; k++ {
						ang := float64(k) * (math.Pi / 6.0)
						p0 := render.Pt(cx+r*0.72*math.Cos(ang), cy+r*0.72*math.Sin(ang))
						p1 := render.Pt(cx+r*0.98*math.Cos(ang), cy+r*0.98*math.Sin(ang))
						rc.StrokePolyline([]render.Point{p0, p1}, 0.75*scale, darkInk, 0.9)
					}

					// Inner dark cutout hexagon at 0.76*r
					cutHex := buildHexPts(cx, cy, r*0.76, float64(b%2)*(math.Pi/6.0))
					rc.FillPolygon(cutHex, hexVoid, 0.98)
					strokeClosedPoly(cutHex, 0.95*scale, colMain, 0.92)

				case 1:
					// Motif 2: 12-Petal or 6-Petal Scalloped Lotus Rosette
					nPetals := 12
					if rng.Float64() < 0.35 {
						nPetals = 6
					}
					phase := float64(b%2) * (math.Pi / float64(nPetals))
					const samples = 72
					rosettePts := make([]render.Point, samples)
					innerRosette := make([]render.Point, samples)
					for i := 0; i < samples; i++ {
						th := (float64(i) / float64(samples)) * 2.0 * math.Pi
						w := 0.5 * (1.0 + math.Cos(float64(nPetals)*(th-phase)))
						rad := r * (0.68 + 0.32*math.Pow(w, 0.75))
						rosettePts[i] = render.Pt(cx+rad*math.Cos(th), cy+rad*math.Sin(th))

						wIn := 0.5 * (1.0 + math.Cos(float64(nPetals)*(th-phase+math.Pi/float64(nPetals))))
						radIn := r * (0.52 + 0.24*math.Pow(wIn, 0.8))
						innerRosette[i] = render.Pt(cx+radIn*math.Cos(th), cy+radIn*math.Sin(th))
					}
					rc.FillPolygon(rosettePts, colMain, 0.96)
					strokeClosedPoly(rosettePts, 1.1*scale, darkInk, 0.95)
					strokeClosedPoly(rosettePts, 0.65*scale, colBright, 0.85)

					rc.FillPolygon(innerRosette, colDeep, 0.95)
					strokeClosedPoly(innerRosette, 0.9*scale, colBright, 0.92)

				case 2:
					// Motif 3: Radial Sunburst Spokes & Beaded Rim
					rc.DrawGlowDot(cx, cy, r, colDeep, 0.95, 0.98)
					rc.DrawSoftRing(cx, cy, r*0.96, 1.2*scale, colMain, 0.95)
					nSpokes := 12
					if r > 0.45*hexR {
						nSpokes = 24
					}
					for s := 0; s < nSpokes; s++ {
						ang := (float64(s) / float64(nSpokes)) * 2.0 * math.Pi
						r0 := math.Max(rNext*0.92, 0.35*r)
						r1 := r * 0.90
						p0 := render.Pt(cx+r0*math.Cos(ang), cy+r0*math.Sin(ang))
						p1 := render.Pt(cx+r1*math.Cos(ang), cy+r1*math.Sin(ang))
						rc.StrokePolyline([]render.Point{p0, p1}, 0.78*scale, colBright, 0.92)

						// Tiny circular pearl bead at each spoke tip
						bx := cx + r*0.88*math.Cos(ang)
						by := cy + r*0.88*math.Sin(ang)
						rc.DrawGlowDot(bx, by, math.Max(0.8*scale, r*0.045), colBright, 0.95, 0.95)
					}
					rc.DrawSoftRing(cx, cy, r*0.74, 0.85*scale, darkInk, 0.92)

				case 3:
					// Motif 4: Ring of 6 or 12 Micro-Medallions / Pearl Beads with Star Cores
					rc.DrawGlowDot(cx, cy, r, colDeep, 0.95, 0.98)
					rc.DrawSoftRing(cx, cy, r*0.96, 1.15*scale, colBright, 0.92)
					nMedallions := 12
					medR := 0.135 * r
					ringR := 0.80 * r
					if r < 0.42*hexR || rng.Float64() < 0.38 {
						nMedallions = 6
						medR = 0.175 * r
						ringR = 0.76 * r
					}
					phase := float64(b%2) * (math.Pi / float64(nMedallions))
					for m := 0; m < nMedallions; m++ {
						ang := float64(m)*(2.0*math.Pi/float64(nMedallions)) + phase
						mx := cx + ringR*math.Cos(ang)
						my := cy + ringR*math.Sin(ang)
						fillAndStrokeCircle(mx, my, medR, colBright, darkInk, 0.85*scale, 0.96)
						fillAndStrokeCircle(mx, my, medR*0.48, darkInk, colMain, 0.55*scale, 0.95)
					}

				case 4:
					// Motif 5: Notched Cog / Segmented Annulus
					rc.DrawGlowDot(cx, cy, r, darkInk, 0.98, 0.98)
					rc.DrawSoftRing(cx, cy, r*0.98, 1.0*scale, colBright, 0.90)
					nSectors := 12
					if r > 0.50*hexR && rng.Float64() < 0.65 {
						nSectors = 18
					}
					dTheta := (2.0 * math.Pi) / float64(nSectors)
					gap := dTheta * 0.18
					rIn := math.Max(rNext*1.02, 0.72*r)
					rOut := 0.93 * r
					for s := 0; s < nSectors; s++ {
						th1 := float64(s)*dTheta + gap
						th2 := float64(s+1)*dTheta - gap
						secCol := colMain
						if s%2 == 1 {
							secCol = colBright
						}
						rc.DrawAnnularSector(cx, cy, rIn, rOut, th1, th2, secCol, 0.95)
					}
					rc.DrawSoftRing(cx, cy, rIn, 0.95*scale, colMain, 0.92)
				}
			}

			// 3. Central Jewel Eye at (cx, cy)
			eyeR := 0.145 * hexR
			rc.DrawGlowDot(cx, cy, eyeR*1.12, darkInk, 0.98, 0.98)

			// Contrasting jewel core: warm amber-gold in cyan cells, or glowing ice-cyan/champagne in gold cells
			var eyeOuterCol, eyeCoreCol color.RGBA
			if cellMix > 0.55 && rng.Float64() < 0.45 {
				eyeOuterCol = scaleColorLum(amberCore, cellLum*1.05)
				eyeCoreCol = scaleColorLum(champagneGold, cellLum*1.10)
			} else if cellMix <= 0.55 && rng.Float64() < 0.25 {
				eyeOuterCol = scaleColorLum(electricTurq, cellLum*1.05)
				eyeCoreCol = scaleColorLum(paleIceCyan, cellLum*1.10)
			} else {
				eyeOuterCol = sampleCellColor(0.68, 0.0)
				eyeCoreCol = sampleCellColor(0.95, 0.0)
			}

			// 6-pointed star + concentric jewel rings in the center
			starPts := make([]render.Point, 12)
			for k := 0; k < 12; k++ {
				ang := float64(k) * (math.Pi / 6.0)
				sr := eyeR
				if k%2 == 1 {
					sr = eyeR * 0.52
				}
				starPts[k] = render.Pt(cx+sr*math.Cos(ang), cy+sr*math.Sin(ang))
			}
			rc.FillPolygon(starPts, eyeOuterCol, 0.96)
			strokeClosedPoly(starPts, 0.75*scale, darkInk, 0.92)

			fillAndStrokeCircle(cx, cy, eyeR*0.58, darkInk, eyeCoreCol, 0.75*scale, 0.98)
			rc.DrawGlowDot(cx, cy, eyeR*0.32, eyeCoreCol, 0.98, 0.85)

			rc.PopClip()

			// Crisp final dark obsidian mortar + metallic inner highlight along the hexagon boundary
			strokeClosedPoly(hexVerts, 1.35*scale, obsidianBg, 0.96)
			innerBevel := buildHexPts(cx, cy, hexR*0.965, 0.0)
			strokeClosedPoly(innerBevel, 0.75*scale, rimCol, 0.78)
		}
	}

	rc.ApplyBloom(0.76, 3.5*scale, 0.14)
	rc.ApplyVignette(0.12)
	c.MergeRenderCanvas(rc)
}
