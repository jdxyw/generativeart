package arts

import (
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type nautilusChamberScroll struct {
	chambersPerTurn int
	terraces        int
	spiralTightness float64
}

// NewNautilusChamberScroll returns a NautilusChamberScroll generator that synthesizes
// a full-bleed 2x2-supersampled 3D Conformal Logarithmic-Spiral Nautilus Chamber &
// Gilded Nacre Septa Relief: a golden-ratio logarithmic spiral shell partitioned
// into concave 3D stepped mother-of-pearl & silk-brocade chambers separated by
// gleaming 24K gold septa walls, internal nacre growth striations, and deep
// shadow-box drop shadows.
func NewNautilusChamberScroll(chambersPerTurn, terraces int, spiralTightness float64) *nautilusChamberScroll {
	if chambersPerTurn <= 0 {
		chambersPerTurn = 14
	}
	if terraces <= 0 {
		terraces = 4
	}
	if spiralTightness <= 0 {
		spiralTightness = 1.0
	}
	return &nautilusChamberScroll{
		chambersPerTurn: common.ConstrainInt(chambersPerTurn, 8, 24),
		terraces:        common.ConstrainInt(terraces, 3, 7),
		spiralTightness: common.Constrain(spiralTightness, 0.7, 1.4),
	}
}

// Generative renders the static NautilusChamberScroll composition.
func (ncs *nautilusChamberScroll) Generative(c *generativeart.Canva) {
	ncs.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping NautilusChamberScroll animation.
func (ncs *nautilusChamberScroll) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	phi := 2.0 * math.Pi * phase

	seed := c.Seed()
	if seed == 0 {
		seed = 20263303
	}
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// =========================================================================
	// 1. IRIDESCENT NACRE, KINGFISHER, SUNBURST GOLD, CORAL & SAPPHIRE PALETTE
	// =========================================================================
	nacreGrad := render.ResolveCyclicGradient(c.Opts().ColorSchema(), common.Dunhuang,
		color.RGBA{R: 252, G: 214, B: 88, A: 255},  // Blazing Solar Gold (Inner Spire)
		color.RGBA{R: 244, G: 124, B: 52, A: 255},  // Sunset Tangerine Coral
		color.RGBA{R: 208, G: 44, B: 62, A: 255},   // Alizarin Cinnabar Ruby
		color.RGBA{R: 118, G: 28, B: 98, A: 255},   // Imperial Amethyst Plum
		color.RGBA{R: 20, G: 58, B: 136, A: 255},   // Nocturne Lapis Sapphire
		color.RGBA{R: 20, G: 138, B: 162, A: 255},  // Kingfisher Cyan-Turquoise
		color.RGBA{R: 46, G: 174, B: 132, A: 255},  // Celadon Emerald Jade
		color.RGBA{R: 248, G: 232, B: 184, A: 255}, // Mother-of-Pearl Champagne Ivory
		color.RGBA{R: 252, G: 214, B: 88, A: 255},  // Loop back to Solar Gold
	)

	goldRim := color.RGBA{R: 255, G: 238, B: 174, A: 255}
	goldBody := color.RGBA{R: 238, G: 182, B: 52, A: 255}
	goldShadow := color.RGBA{R: 88, G: 54, B: 16, A: 255}

	// Logarithmic spiral growth parameter b: r = r0 * exp(b * theta)
	// Choose b so 2*pi*b ≈ 1.08 (each full revolution expands radius by e^1.08 ≈ 2.94x)
	bSpiral := 0.172 * ncs.spiralTightness
	twoPiB := 2.0 * math.Pi * bSpiral
	nChambers := float64(ncs.chambersPerTurn)
	nTiers := ncs.terraces

	loopX := 0.24 * (math.Cos(phi) - 1.0)
	loopY := 0.24 * math.Sin(phi)

	// =========================================================================
	// 2. RASTERIZE 2X SUPERSAMPLED CONFORMAL NAUTILUS CHAMBERS & 24K GOLD SEPTA
	// =========================================================================
	gw, gh := w*2+1, h*2+1
	zGrid := make([]float64, gw*gh)
	idGrid := make([]int, gw*gh)
	rimGrid := make([]float64, gw*gh)
	albRGrid := make([]float64, gw*gh)
	albGGrid := make([]float64, gw*gh)
	albBGrid := make([]float64, gw*gh)

	// Slightly golden-ratio offset spiral center at (0.48, 0.52)
	const cxNorm, cyNorm = 0.48, 0.52

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}

	var wgRaster sync.WaitGroup
	for wk := 0; wk < workers; wk++ {
		gy0 := (wk * gh) / workers
		gy1 := ((wk + 1) * gh) / workers
		wgRaster.Add(1)
		go func(r0, r1 int) {
			defer wgRaster.Done()
			for gy := r0; gy < r1; gy++ {
				uy := (float64(gy) * 0.5) / fh
				dy := uy - cyNorm
				rowOff := gy * gw
				for gx := 0; gx < gw; gx++ {
					ux := (float64(gx) * 0.5) / fw
					dx := ux - cxNorm
					idx := rowOff + gx

					r := math.Hypot(dx, dy)
					theta := math.Atan2(dy, dx)

					// Central 3-Tier Recessed Solar Cabochon Eye (r < 0.042)
					if r < 0.042 {
						dPx := (0.042 - r) * fw
						curCol := color.RGBA{R: 255, G: 244, B: 196, A: 255}
						if r > 0.026 {
							curCol = color.RGBA{R: 246, G: 178, B: 52, A: 255}
						}
						rL, gL, bL, _ := render.ColorToLinearRGB(curCol)
						zGrid[idx] = 24.0
						idGrid[idx] = 9999
						if dPx < 1.8 {
							rimGrid[idx] = common.Smoothstep(1.8, 0.3, dPx)
						}
						albRGrid[idx], albGGrid[idx], albBGrid[idx] = rL, gL, bL
						continue
					}

					// Conformal log-polar coordinates: u = ln(r)
					// Animated zoom-rotation invariance: advancing theta by +phi (or u by -twoPiB*phase)
					// makes the entire Nautilus spiral smoothlyunfurl in an infinite Droste loop!
					u := math.Log(r)
					// Continuous spiral whorl index psi: integer part nWhorl, fractional part fWhorl in [0, 1)
					psi := (u - bSpiral*(theta+phi)) / twoPiB
					nWhorl := int(math.Floor(psi))
					fWhorl := psi - float64(nWhorl) // [0 = inner spiral wall, 1 = outer spiral wall]

					// Continuous chamber coordinate along the spiral whorl:
					// Unwrapped total angle thetaTotal = theta + 2*pi*nWhorl + phi (continuous across branch cut!)
					thetaUnwrapped := (u - twoPiB*fWhorl) / bSpiral
					// Concave Nautilus septum arch: curves gracefully backward at the center of the whorl (fWhorl ≈ 0.5)
					septumArch := 0.42 * math.Sin(fWhorl*math.Pi)
					pSeptum := 0.06 * ne.Perlin2D(math.Cos(theta)*1.5+loopX*0.3, math.Sin(theta)*1.5+loopY*0.3)

					chamberCoord := (thetaUnwrapped/(2.0*math.Pi))*nChambers + septumArch + pSeptum
					kChamber := int(math.Floor(chamberCoord))
					fChamber := chamberCoord - float64(kChamber) // [0 = back septum, 1 = front septum]

					// Convert fractional coordinates (fWhorl, fChamber) to exact Euclidean pixel distances!
					// Whorl radial width at radius r is approximately r * twoPiB
					whorlWidthPx := math.Max(14.0, r*fw*twoPiB)
					// Chamber arc length at radius r is approximately r * 2*pi / nChambers
					chamberLenPx := math.Max(12.0, r*fw*(2.0*math.Pi/nChambers))

					distWhorlWallPx := math.Min(fWhorl, 1.0-fWhorl) * whorlWidthPx
					distSeptumWallPx := math.Min(fChamber, 1.0-fChamber) * chamberLenPx
					distBorderPx := math.Min(distWhorlWallPx, distSeptumWallPx)

					// 1) 3D Cylindrical 24K Gold Spiral Whorl Wall (distWhorlWallPx < 2.2) & Septum Wall (distSeptumWallPx < 1.65)
					whorlWallHalfW := 2.15
					septumWallHalfW := 1.60
					if distWhorlWallPx < whorlWallHalfW || distSeptumWallPx < septumWallHalfW {
						uWall := math.Min(distWhorlWallPx/whorlWallHalfW, distSeptumWallPx/septumWallHalfW)
						var wCol color.RGBA
						if uWall < 0.44 {
							wCol = render.LerpOklab(goldRim, goldBody, uWall/0.44)
						} else {
							wCol = render.LerpOklab(goldBody, goldShadow, (uWall-0.44)/0.56)
						}
						rL, gL, bL, _ := render.ColorToLinearRGB(wCol)
						zGrid[idx] = 22.0 + 1.8*(1.0-uWall*uWall)
						idGrid[idx] = 8888
						rimGrid[idx] = 1.0 - 0.5*uWall
						albRGrid[idx], albGGrid[idx], albBGrid[idx] = rL, gL, bL
						continue
					}

					// 2) Inside Nautilus Chamber (nWhorl, kChamber):
					// Step downward through nTiers recessed 3D shadow-box nacre & silk-brocade alcoves!
					inwardPx := distBorderPx - 1.60
					tierStepPx := math.Max(4.2, math.Min(whorlWidthPx, chamberLenPx)*0.115)
					qTier := inwardPx / tierStepPx
					if qTier < 0 {
						qTier = 0
					}
					kTier := int(math.Floor(qTier))
					if kTier >= nTiers {
						kTier = nTiers - 1
					}
					fTier := common.Clamp01(qTier - float64(kTier))
					distStepPx := (qTier - float64(kTier)) * tierStepPx

					// Smooth color progression along the unwrapped spiral chamber index kChamber (ZERO branch-cut seam at theta = +-pi!)
					hue := math.Mod(float64(kChamber)*0.038+float64(kTier)*0.08+100.0, 1.0)
					colOuter := nacreGrad.SampleOklab(hue)
					colInner := nacreGrad.SampleOklab(math.Mod(hue+0.11, 1.0))
					pCol := render.LerpOklab(colOuter, colInner, fTier*0.75)

					// Blend subtle Airy thin-film mother-of-pearl iridescence inside the deepest chamber alcove!
					if kTier >= nTiers-2 {
						tfCol := render.SampleThinFilmIridescence(0.78, 380.0+180.0*hue+45.0*fChamber, 1.48)
						pCol = render.LerpOklab(pCol, tfCol, 0.26)
					}

					// Fine curved mother-of-pearl & silk growth striations inside each chamber tier!
					ribWave := math.Sin(distStepPx*0.85 + fWhorl*8.0)
					if ribWave > 0 {
						pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 248, B: 226, A: 255}, ribWave*0.14)
					} else {
						pCol = render.LerpOklab(pCol, color.RGBA{R: 10, G: 8, B: 22, A: 255}, -ribWave*0.13)
					}

					// Outer chamber tiers (kTier = 0) are highest (z = 19.0) and step down into the recessed chamber well!
					curZ := 19.0 - float64(kTier)*2.8 + 0.35*math.Exp(-distStepPx/5.0)
					// Use ONLY kChamber and kTier (never nWhorl!) so ID is 100% branch-cut-free across theta = +-pi!
					curID := (kChamber+5000)*10 + kTier
					curRim := 0.0
					if distStepPx < 1.55 {
						curRim = common.Smoothstep(1.55, 0.30, distStepPx)
					}

					grain := ne.Perlin2D(float64(gx)*0.35, float64(gy)*0.35) * 0.020
					rL, gL, bL, _ := render.ColorToLinearRGB(pCol)
					if curRim > 0 {
						rr, rg, rb, _ := render.ColorToLinearRGB(goldRim)
						rL = rL*(1.0-curRim*0.82) + rr*curRim*0.82
						gL = gL*(1.0-curRim*0.82) + rg*curRim*0.82
						bL = bL*(1.0-curRim*0.82) + rb*curRim*0.82
					}

					zGrid[idx] = curZ
					idGrid[idx] = curID
					rimGrid[idx] = curRim
					albRGrid[idx] = common.Clamp01(rL + grain)
					albGGrid[idx] = common.Clamp01(gL + grain)
					albBGrid[idx] = common.Clamp01(bL + grain)
				}
			}
		}(gy0, gy1)
	}
	wgRaster.Wait()

	// =========================================================================
	// 3. PARALLEL 2X2 SUPERSAMPLED SHADOW-BOX DROP SHADOWS & AO
	// =========================================================================
	subOffs := [4][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
	aoOffsets := [12][2]int{
		{-3, -3}, {0, -4}, {3, -3}, {-4, 0},
		{4, 0}, {-3, 3}, {0, 4}, {3, 3},
		{-8, -6}, {8, -6}, {-6, 8}, {6, 8},
	}

	var wgShade sync.WaitGroup
	chunk := (h + workers - 1) / workers
	for wk := 0; wk < workers; wk++ {
		y0 := wk * chunk
		y1 := y0 + chunk
		if y0 >= h {
			break
		}
		if y1 > h {
			y1 = h
		}
		wgShade.Add(1)
		go func(startY, endY int) {
			defer wgShade.Done()
			for y := startY; y < endY; y++ {
				for x := 0; x < w; x++ {
					var sumR, sumG, sumB float64

					for _, so := range subOffs {
						gx := x*2 + so[0]
						gy := y*2 + so[1]
						idx := gy*gw + gx
						z0 := zGrid[idx]
						id0 := idGrid[idx]

						aoSum := 0.0
						for _, off := range aoOffsets {
							sx := common.ConstrainInt(gx+off[0], 0, gw-1)
							sy := common.ConstrainInt(gy+off[1], 0, gh-1)
							sIdx := sy*gw + sx
							if idGrid[sIdx] != id0 && zGrid[sIdx] > z0 {
								dz := math.Min(zGrid[sIdx]-z0, 4.5)
								aoSum += common.Constrain(dz/3.0, 0.0, 1.0)
							}
						}
						aoShade := 1.0 - (aoSum/12.0)*0.60

						castOcc := 0.0
						for step := 1; step <= 12; step++ {
							dist := float64(step) * 1.55
							sx := common.ConstrainInt(int(math.Round(float64(gx)-0.62*dist)), 0, gw-1)
							sy := common.ConstrainInt(int(math.Round(float64(gy)-0.78*dist)), 0, gh-1)
							sIdx := sy*gw + sx
							if idGrid[sIdx] != id0 && zGrid[sIdx] > z0 {
								falloff := 1.0 - float64(step-1)/12.0
								castOcc = math.Max(castOcc, falloff*0.52)
							}
						}
						castShade := 1.0 - castOcc

						ux := (float64(gx) * 0.5) / fw
						uy := (float64(gy) * 0.5) / fh
						dEye := math.Hypot(ux-cxNorm, uy-cyNorm)
						eyeGlow := 0.14 * math.Exp(-(dEye*dEye)/0.05)

						shade := aoShade * castShade
						rOut := albRGrid[idx]*shade + eyeGlow*1.05 + rimGrid[idx]*0.16
						gOut := albGGrid[idx]*shade + eyeGlow*0.85 + rimGrid[idx]*0.13
						bOut := albBGrid[idx]*shade + eyeGlow*0.45 + rimGrid[idx]*0.06

						sumR += common.Clamp01(rOut)
						sumG += common.Clamp01(gOut)
						sumB += common.Clamp01(bOut)
					}

					rc.SetPixelLinear(x, y, sumR*0.25, sumG*0.25, sumB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wgShade.Wait()

	c.MergeRenderCanvas(rc)
}
