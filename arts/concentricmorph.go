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

type concentricMorph struct {
	gridCols int
	gridRows int
	disorder float64
}

// NewConcentricMorph creates a Vera Molnár-inspired order-to-disorder grid engine
// where terracotta concentric squares morph into ivory concentric circles and
// scatter into cerulean/teal geometric clouds along a diagonal entropy wave.
func NewConcentricMorph(gridCols, gridRows int, disorder float64) *concentricMorph {
	if gridCols <= 0 {
		gridCols = 18
	}
	if gridRows <= 0 {
		gridRows = 18
	}
	if disorder <= 0 {
		disorder = 1.0
	}
	return &concentricMorph{
		gridCols: gridCols,
		gridRows: gridRows,
		disorder: disorder,
	}
}

func strokeAxisSquare(rc *render.Canvas, cx, cy, halfSide, strokeW float64, col color.RGBA, alpha float64) {
	if halfSide <= 0.2 || strokeW <= 0 || alpha <= 0 {
		return
	}
	p := &canvas.Path{}
	p.MoveTo(cx-halfSide, cy-halfSide)
	p.LineTo(cx+halfSide, cy-halfSide)
	p.LineTo(cx+halfSide, cy+halfSide)
	p.LineTo(cx-halfSide, cy+halfSide)
	p.Close()
	rc.StrokePath(p, strokeW, canvas.SquareCap, canvas.MiterJoin, col, alpha)
}

func strokeCircleOutline(rc *render.Canvas, cx, cy, radius, strokeW float64, col color.RGBA, alpha float64) {
	if radius <= 0.25 || strokeW <= 0 || alpha <= 0 {
		return
	}
	const segs = 56
	p := &canvas.Path{}
	for i := 0; i <= segs; i++ {
		ang := (float64(i) / float64(segs)) * 2.0 * math.Pi
		x := cx + radius*math.Cos(ang)
		y := cy + radius*math.Sin(ang)
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.Close()
	rc.StrokePath(p, strokeW, canvas.RoundCap, canvas.RoundJoin, col, alpha)
}

// Generative renders the ConcentricMorph composition.
func (cm *concentricMorph) Generative(c *generativeart.Canva) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := c.Seed()
	if seed == 0 {
		seed = 2026102801
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	bgNavy := color.RGBA{R: 22, G: 27, B: 34, A: 255} // #161b22
	rc.FillBackground(bgNavy)
	rc.SetBlendMode(render.BlendOver)

	// Curated palette stops matching Reference Image 1
	coralRed := color.RGBA{R: 214, G: 102, B: 82, A: 255}     // #d66652
	coralDeep := color.RGBA{R: 178, G: 72, B: 56, A: 255}     // #b24838
	coralLight := color.RGBA{R: 228, G: 136, B: 110, A: 255}  // #e4886e
	warmPeach := color.RGBA{R: 230, G: 178, B: 136, A: 255}   // #e6b288
	ivoryCream := color.RGBA{R: 236, G: 230, B: 196, A: 255}  // #ece6c4
	sandMuted := color.RGBA{R: 206, G: 198, B: 164, A: 255}   // #cec6a4
	paleSage := color.RGBA{R: 164, G: 202, B: 194, A: 255}    // #a4cac2
	tealBright := color.RGBA{R: 68, G: 162, B: 184, A: 255}   // #44a2b8
	tealCerulean := color.RGBA{R: 42, G: 124, B: 152, A: 255} // #2a7c98
	tealDeep := color.RGBA{R: 28, G: 88, B: 112, A: 255}      // #1c5870

	cols := cm.gridCols
	rows := cm.gridRows
	marginX := 14.0 * scale
	marginY := 14.0 * scale
	cellW := (fw - 2.0*marginX) / float64(cols)
	cellH := (fh - 2.0*marginY) / float64(rows)
	cellMin := math.Min(cellW, cellH)

	// Pass 1: Render ordered and moderately-eccentric cells first, then high-disorder cells
	// so scattered teal shapes naturally float across neighboring cells.
	type cellSpec struct {
		col, row    int
		u, v        float64
		cx, cy      float64
		disorderAmt float64
		cornerWarm  float64
		isSquare    bool
	}

	specs := make([]cellSpec, 0, cols*rows)
	for r := 0; r < rows; r++ {
		v := 0.5
		if rows > 1 {
			v = float64(r) / float64(rows-1)
		}
		for col := 0; col < cols; col++ {
			u := 0.5
			if cols > 1 {
				u = float64(col) / float64(cols-1)
			}
			cx := marginX + (float64(col)+0.5)*cellW
			cy := marginY + (float64(r)+0.5)*cellH

			nWarp := 0.06 * noise.Simplex2D(u*3.2+1.1, v*3.2-0.7)

			// Distance from Top-Left corner (0,0) and Bottom-Right corner (1,1)
			dTL := math.Hypot(u*1.15, v*0.92) + nWarp*0.4
			dBR := math.Hypot((1.0-u)*1.05, (1.0-v)*1.10) - nWarp*0.4

			// Ordered mask is 1 in the Top-Left / Mid-Left ordered region and Bottom-Right ordered region
			// Look at Image 1:
			// Ordered Zone A covers: top-left down to (u=0, v=0.75) and across to (u=0.56, v=0)
			// Ordered Zone B covers: bottom-right from (u=0.48, v=1.0) up to (u=0.92, v=0.56)
			distOrderA := u*1.08 + v*0.72 + 0.10*math.Sin(v*math.Pi*1.8) + nWarp
			orderA := 1.0 - common.Smoothstep(0.56, 0.76, distOrderA)

			distOrderB := (1.0-u)*1.05 + (1.0-v)*0.95 + 0.08*math.Cos(u*math.Pi*1.5) - nWarp
			orderB := 1.0 - common.Smoothstep(0.42, 0.62, distOrderB)

			orderMask := math.Max(orderA, orderB)
			disAmt := common.Constrain((1.0-orderMask)*cm.disorder, 0, 1)

			// Boost disorder in the Upper-Right (u > 0.62, v < 0.52) and Bottom-Left (u < 0.45, v > 0.74)
			urBoost := common.Smoothstep(0.54, 0.78, u) * common.Smoothstep(0.58, 0.22, v)
			blBoost := common.Smoothstep(0.48, 0.18, u) * common.Smoothstep(0.68, 0.92, v)
			disAmt = common.Constrain(math.Max(disAmt, math.Max(urBoost, blBoost)*cm.disorder), 0, 1)

			// Warmth factor: 1.0 at TL and BR terracotta square corners, 0.0 near the ivory/sage boundary
			warmTL := 1.0 - common.Smoothstep(0.14, 0.56, dTL)
			warmBR := 1.0 - common.Smoothstep(0.10, 0.46, dBR)
			cornerWarm := math.Max(warmTL, warmBR)

			// Determine Square vs Circle:
			// 1. Ordered Top-Left corner: u < 0.22 && v < 0.38 && (u*1.45 + v < 0.48)
			// 2. Ordered Bottom-Right corner: u > 0.76 && v > 0.82 && ((1-u)*1.2 + (1-v) < 0.34)
			// 3. Disordered Upper-Right square cloud: u > 0.66 && v < 0.52 && disAmt > 0.52
			// 4. Disordered extreme Bottom-Left corner: u < 0.20 && v > 0.88 && disAmt > 0.65
			isSq := false
			if u < 0.23 && v < 0.39 && (u*1.42+v+nWarp*0.35 < 0.49) {
				isSq = true
				disAmt = 0 // Keep TL terracotta squares strictly ordered
			} else if u > 0.76 && v > 0.82 && ((1.0-u)*1.15+(1.0-v)-nWarp*0.35 < 0.35) {
				isSq = true
				disAmt = 0 // Keep BR terracotta squares strictly ordered
			} else if u > 0.65 && v < 0.52 && disAmt > 0.52 {
				if u+0.35*(0.52-v)+nWarp > 0.72 {
					isSq = true
				}
			} else if u < 0.22 && v > 0.88 && disAmt > 0.65 {
				isSq = true
			}

			specs = append(specs, cellSpec{
				col:         col,
				row:         r,
				u:           u,
				v:           v,
				cx:          cx,
				cy:          cy,
				disorderAmt: disAmt,
				cornerWarm:  cornerWarm,
				isSquare:    isSq,
			})
		}
	}

	// Render cells
	for _, sp := range specs {
		baseR := 0.435 * cellMin
		if sp.isSquare && sp.disorderAmt < 0.1 {
			baseR = 0.445 * cellMin
		}

		// Determine cell base color
		var cellBaseCol color.RGBA
		if sp.disorderAmt < 0.12 {
			// Ordered zone: Coral Red -> Warm Peach -> Ivory Cream -> Pale Sage
			if sp.cornerWarm > 0.55 {
				t := (sp.cornerWarm - 0.55) / 0.45
				cellBaseCol = render.LerpOklab(warmPeach, coralRed, common.Constrain(t, 0, 1))
			} else if sp.cornerWarm > 0.18 {
				t := (sp.cornerWarm - 0.18) / 0.37
				cellBaseCol = render.LerpOklab(ivoryCream, warmPeach, common.Constrain(t, 0, 1))
			} else {
				t := sp.cornerWarm / 0.18
				cellBaseCol = render.LerpOklab(paleSage, ivoryCream, common.Constrain(t*0.85+0.15, 0, 1))
			}
		} else if sp.disorderAmt < 0.52 {
			// Transition eccentric zone: Ivory/Pale Sage -> Bright Teal
			t := (sp.disorderAmt - 0.12) / 0.40
			cellBaseCol = render.LerpOklab(ivoryCream, tealBright, common.Constrain(t*0.75, 0, 1))
		} else {
			// Deep disorder zone: Cerulean & Deep Teal (with occasional pale sage/cream accent rings)
			t := (sp.disorderAmt - 0.52) / 0.48
			cellBaseCol = render.LerpOklab(tealBright, tealCerulean, common.Constrain(t, 0, 1))
		}

		numRings := 6
		if !sp.isSquare && sp.disorderAmt < 0.35 {
			numRings = 7
		}

		// Eccentric drift direction for transition cells (points toward the disordered trough)
		driftAngle := math.Atan2(0.55-sp.v, 0.72-sp.u) + (rng.Float64()-0.5)*0.95

		// Choose 1 or 2 rings in ordered cells to have a wider dark gap (like vinyl grooves in Image 1)
		skipRing := -1
		if sp.disorderAmt < 0.25 && rng.Float64() < 0.42 {
			skipRing = 3 + rng.Intn(3)
		}

		for k := 0; k < numRings; k++ {
			if k == skipRing {
				continue
			}
			tRing := float64(k) / float64(numRings-1) // 0 = outermost, 1 = innermost

			var rk, rx, ry, strokeW, alpha float64
			ringCol := cellBaseCol

			if sp.disorderAmt < 0.06 {
				// Strictly ordered concentric square or circle
				rk = baseR * (1.0 - 0.82*tRing)
				rx, ry = sp.cx, sp.cy
				strokeW = (1.25 + 0.65*math.Sin(float64(k)*2.1+float64(sp.col*3+sp.row)*0.7)) * scale

				// Subtle per-ring tonal variation matching Image 1
				ringVar := math.Sin(float64(k)*1.8 + float64(sp.col*5+sp.row*7)*0.9)
				if sp.cornerWarm > 0.45 {
					if ringVar > 0.35 {
						ringCol = render.LerpOklab(ringCol, coralLight, 0.45)
					} else if ringVar < -0.35 {
						ringCol = render.LerpOklab(ringCol, coralDeep, 0.50)
					}
				} else {
					if ringVar > 0.35 {
						ringCol = render.LerpOklab(ringCol, ivoryCream, 0.55)
					} else if ringVar < -0.35 {
						ringCol = render.LerpOklab(ringCol, sandMuted, 0.55)
					}
				}
				alpha = 0.94
			} else if sp.disorderAmt < 0.52 {
				// Eccentric transition zone: outer ring stays near cell center while inner rings
				// slide progressively toward one flank and spill slightly out!
				ecc := (sp.disorderAmt - 0.06) / 0.46
				rk = baseR * (1.0 - 0.80*tRing) * (1.0 - 0.12*ecc*rng.Float64())
				shift := ecc * baseR * 1.15 * math.Pow(tRing, 0.85)
				jitX := (rng.Float64() - 0.5) * ecc * baseR * 0.45
				jitY := (rng.Float64() - 0.5) * ecc * baseR * 0.45
				rx = sp.cx + math.Cos(driftAngle)*shift + jitX
				ry = sp.cy + math.Sin(driftAngle)*shift + jitY
				strokeW = (1.45 - 0.42*ecc + 0.35*rng.Float64()) * scale

				if rng.Float64() < 0.35*ecc {
					ringCol = render.LerpOklab(paleSage, tealBright, 0.65)
				} else if rng.Float64() < 0.45 {
					ringCol = ivoryCream
				}
				alpha = 0.90
			} else {
				// High-disorder scattered zone (Upper-Right squares & Center/Bottom-Left circles)
				scatter := (sp.disorderAmt - 0.45) / 0.55
				spread := cellMin * (0.42 + 0.42*scatter)
				rx = sp.cx + (rng.Float64()-0.5)*2.0*spread
				ry = sp.cy + (rng.Float64()-0.5)*2.0*spread

				// Keep scattered shapes inside the outer frame margin
				rx = common.Constrain(rx, marginX+4*scale, fw-marginX-4*scale)
				ry = common.Constrain(ry, marginY+4*scale, fh-marginY-4*scale)

				if sp.isSquare {
					// Scattered open squares in Upper-Right & Bottom-Left (varying sizes, crisp thin teal lines)
					rk = baseR * (0.16 + 0.76*math.Pow(1.0-tRing*0.75, 1.1)*(0.55+0.60*rng.Float64()))
					strokeW = (0.95 + 0.35*rng.Float64()) * scale
					if rng.Float64() < 0.45 {
						ringCol = tealCerulean
					} else if rng.Float64() < 0.75 {
						ringCol = tealBright
					} else {
						ringCol = tealDeep
					}
					alpha = 0.78 + 0.18*rng.Float64()
				} else {
					// Scattered overlapping circles in the diagonal trough (teal-blue + occasional cream/sage rings)
					rk = baseR * (0.14 + 0.78*(1.0-0.72*tRing)*(0.60+0.50*rng.Float64()))
					strokeW = (1.05 + 0.45*rng.Float64()) * scale
					rPick := rng.Float64()
					if rPick < 0.16*(1.0-scatter*0.6) {
						ringCol = ivoryCream
						strokeW *= 1.18
					} else if rPick < 0.36 {
						ringCol = paleSage
					} else if rPick < 0.72 {
						ringCol = tealBright
					} else {
						ringCol = tealCerulean
					}
					alpha = 0.84 + 0.14*rng.Float64()
				}
			}

			if sp.isSquare {
				strokeAxisSquare(rc, rx, ry, rk, strokeW, ringCol, alpha)
			} else {
				strokeCircleOutline(rc, rx, ry, rk, strokeW, ringCol, alpha)
			}
		}
	}

	c.MergeRenderCanvas(rc)
}
