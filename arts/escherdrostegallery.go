package arts

import (
	"image/color"
	"math"
	"math/cmplx"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type escherDrosteGallery struct {
	branches      int
	archesPerRing int
	scaleRatio    float64
}

// NewEscherDrosteGallery creates an EscherDrosteGallery generator that renders
// a 100% holomorphic conformal M.C. Escher logarithmic-spiral composition
// W = (kR + i*kI)*ln(z). Supports two curated conformal pattern modes:
//   - Royal Cloisonné & Iridescent Nautilus Chambers (default: branches == 1 and even seed)
//   - Escher Glide-Reflection Interlocking Wave-Birds (branches >= 2 or odd seed)
func NewEscherDrosteGallery(branches, archesPerRing int, scaleRatio float64) *escherDrosteGallery {
	if branches <= 0 {
		branches = 1
	}
	if archesPerRing <= 0 {
		archesPerRing = 8
	}
	if scaleRatio <= 1.1 {
		scaleRatio = 16.0
	}
	return &escherDrosteGallery{
		branches:      branches,
		archesPerRing: archesPerRing,
		scaleRatio:    scaleRatio,
	}
}

// Generative renders the static EscherDrosteGallery composition.
func (edg *escherDrosteGallery) Generative(c *generativeart.Canva) {
	edg.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless infinite Droste zoom loop (phase in [0, 1)).
func (edg *escherDrosteGallery) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	invNorm := 1.0 / (0.48 * minDim)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Select between the two curated conformal Droste spiral styles:
	//   useWaveBirds == false: Royal Cloisonné & Iridescent Nautilus Shell Chambers (Style 3)
	//   useWaveBirds == true:  Escher Glide-Reflection Interlocking Wave-Birds (Style 1)
	useWaveBirds := edg.branches >= 2 || (c.Seed()%2 != 0)

	linRGB := func(r, g, b uint8) [3]float64 {
		lr, lg, lb, _ := render.ColorToLinearRGB(color.RGBA{R: r, G: g, B: b, A: 255})
		return [3]float64{lr, lg, lb}
	}
	lerp3 := func(a, b [3]float64, t float64) [3]float64 {
		t = common.Clamp01(t)
		return [3]float64{
			a[0] + (b[0]-a[0])*t,
			a[1] + (b[1]-a[1])*t,
			a[2] + (b[2]-a[2])*t,
		}
	}

	// Curated museum woodcut, cloisonné & printmaking palette in linear RGB
	ivoryCol := linRGB(0xfc, 0xf5, 0xe6)    // #fcf5e6 Luminous Museum Ivory
	sandCol := linRGB(0xea, 0xc2, 0x82)     // #eac282 Warm Golden Limestone
	goldCol := linRGB(0xf4, 0xbe, 0x42)     // #f4be42 Burnished 24k Gold
	giltHiCol := linRGB(0xff, 0xe8, 0x9c)   // #ffe89c Specular Gold Filigree
	cinnabarCol := linRGB(0xdc, 0x44, 0x2e) // #dc442e Cinnabar Vermilion
	celadonCol := linRGB(0x24, 0xa2, 0x96)  // #24a296 Imperial Celadon Jade
	cyanCol := linRGB(0x48, 0xd8, 0xe2)     // #48d8e2 Luminous Turquoise
	indigoCol := linRGB(0x10, 0x24, 0x4a)   // #10244a Deep Prussian Lapis
	midnightCol := linRGB(0x07, 0x0f, 0x20) // #070f20 Abyssal Midnight Ink
	inkCol := linRGB(0x0c, 0x0e, 0x14)      // #0c0e14 Woodcut Engraving Ink

	// Exact holomorphic conformal spiral lattice parameters:
	// Across theta = +-pi, uCont shifts by -2 (even integer) and vCont shifts by
	// +numCellsAngular (even integer), guaranteeing 100% branch-cut-free C^infinity
	// continuity for every 2x2-periodic function on (uCont, vCont).
	numCellsAngular := edg.archesPerRing + 4
	if numCellsAngular < 12 {
		numCellsAngular = 12
	}
	if numCellsAngular%2 != 0 {
		numCellsAngular++
	}
	effBranches := edg.branches
	if effBranches == 2 {
		effBranches = 1
	}
	spiralShiftB := 2 * effBranches

	kR := float64(numCellsAngular) / (2.0 * math.Pi)
	kI := float64(spiralShiftB) / (2.0 * math.Pi)
	phi := 2.0 * math.Pi * phase

	subOffsets := [4][2]float64{
		{0.25, 0.25},
		{0.75, 0.25},
		{0.25, 0.75},
		{0.75, 0.75},
	}

	// =========================================================================
	// MODE B: Escher Glide-Reflection Interlocking Wave-Birds (100% Seam-Free)
	// =========================================================================
	// Uses the exact glide-reflection anti-symmetric implicit field:
	//   Phi(u + 1, -v) == -Phi(u, v)
	// and shades every figure purely via contour level-sets Phi(u,v) and its analytical
	// gradient (dPhi/du, dPhi/dv) so there are ZERO internal grid/Voronoi seams.
	evalWaveBirds := func(uCont, vCont, rScreen, confScale float64) [3]float64 {
		detailVis := common.Smoothstep(0.008, 0.042, confScale)

		piU := math.Pi * uCont
		piV := math.Pi * vCont

		sinU, cosU := math.Sin(piU), math.Cos(piU)
		sinV, cosV := math.Sin(piV), math.Cos(piV)
		sin2U, cos2U := math.Sin(2.0*piU), math.Cos(2.0*piU)
		sin2V, cos2V := math.Sin(2.0*piV), math.Cos(2.0*piV)
		sin3U, cos3U := math.Sin(3.0*piU), math.Cos(3.0*piU)

		phiField := cosU*cosV +
			0.50*sinU*cos2V +
			0.44*cos2U*sinV +
			0.24*sin2U*sinV -
			0.16*sin3U*cos2V

		dPhidu := -sinU*cosV +
			0.50*cosU*cos2V -
			0.88*sin2U*sinV +
			0.48*cos2U*sinV -
			0.48*cos3U*cos2V
		dPhidv := -cosU*sinV -
			1.00*sinU*sin2V +
			0.44*cos2U*cosV +
			0.24*sin2U*cosV +
			0.32*sin3U*sin2V

		gradMag := math.Hypot(dPhidu, dPhidv) + 1e-6
		distIso := math.Abs(phiField) / (0.65 + 0.35*gradMag)
		isPos := phiField >= 0

		sign := 1.0
		if !isPos {
			sign = -1.0
		}
		nx := sign * dPhidu / gradMag
		ny := sign * dPhidv / gradMag
		domeZ := common.Clamp01(distIso * 1.45)
		light := common.Clamp01(0.58 + 0.28*(-0.6*nx-0.5*ny) + 0.22*domeZ)

		flowWave := 0.5 + 0.5*math.Sin(piU-sign*piV+0.5*phi)
		radialT := common.Clamp01(rScreen / 0.85)

		var col [3]float64
		if isPos {
			wingTone := lerp3(cinnabarCol, goldCol, 0.35+0.55*flowWave)
			coreTone := lerp3(ivoryCol, giltHiCol, 0.40*(1.0-radialT))
			col = lerp3(wingTone, coreTone, common.Smoothstep(0.05, 0.58, distIso))
			col = lerp3(midnightCol, col, 0.25+0.75*light)
		} else {
			wingTone := lerp3(indigoCol, celadonCol, 0.30+0.60*flowWave)
			coreTone := lerp3(celadonCol, cyanCol, 0.45+0.40*(1.0-radialT))
			col = lerp3(indigoCol, lerp3(wingTone, coreTone, common.Smoothstep(0.06, 0.60, distIso)), 0.35+0.65*light)
		}

		contourRibs := 0.5 + 0.5*math.Cos(22.0*distIso-3.0*flowWave-phi)
		groove := (1.0 - common.Smoothstep(0.22, 0.52, contourRibs)) *
			detailVis * common.Smoothstep(0.04, 0.14, distIso) * (1.0 - common.Smoothstep(0.55, 0.78, distIso))
		ridge := common.Smoothstep(0.75, 0.96, contourRibs) *
			detailVis * common.Smoothstep(0.05, 0.16, distIso)

		if isPos {
			col = lerp3(col, inkCol, 0.34*groove)
			col = lerp3(col, giltHiCol, 0.32*ridge)
		} else {
			col = lerp3(col, midnightCol, 0.42*groove)
			col = lerp3(col, giltHiCol, 0.34*ridge)
		}

		if distIso > 0.60 {
			eyeT := (distIso - 0.60) / 0.28
			if eyeT < 0.45 {
				ring := math.Sin(eyeT / 0.45 * math.Pi)
				col = lerp3(col, giltHiCol, 0.85*detailVis*ring)
			} else {
				pupil := common.Smoothstep(0.45, 0.68, eyeT)
				if isPos {
					col = lerp3(col, indigoCol, 0.90*detailVis*pupil)
				} else {
					col = lerp3(col, cinnabarCol, 0.90*detailVis*pupil)
				}
			}
		}

		goldBand := common.Smoothstep(0.065, 0.024, distIso) * common.Smoothstep(0.007, 0.022, distIso)
		inkLine := common.Smoothstep(0.024, 0.005, distIso)
		col = lerp3(col, giltHiCol, 0.80*goldBand)
		col = lerp3(col, inkCol, 0.92*inkLine)

		return col
	}

	// =========================================================================
	// MODE A: Royal Cloisonné & Iridescent Nautilus Shell Chambers (*Spirals*)
	// =========================================================================
	evalNautilusChambers := func(uCont, vCont, confScale float64) [3]float64 {
		detailVis := common.Smoothstep(0.008, 0.045, confScale)

		// 12 fluted logarithmic spiral tubes separated by braided 24k gold cables
		cableCoord := vCont + 0.14*math.Sin(math.Pi*uCont)
		flV := math.Floor(cableCoord)
		tubeIdx := int(flV)
		tubeV := (cableCoord - flV)*2.0 - 1.0 // in [-1, 1] across each spiral tube

		// Concave crescent Nautilus septa chambers along uCont
		septaU := uCont - 0.32*tubeV*tubeV
		flU := math.Floor(septaU)
		chamberIdx := int(flU)
		fu := septaU - flU // in [0, 1) within each chamber
		chamberX := 2.0*fu - 1.0

		// 3D curved nacre chamber normal (nx, ny, nz)
		nx := -0.65 * chamberX
		ny := -0.75 * tubeV
		nz := math.Sqrt(math.Max(0.08, 1.0-nx*nx-ny*ny))

		ndotl := math.Max(0.0, -0.42*nx-0.54*ny+0.73*nz)
		spec := math.Pow(math.Max(0.0, -0.20*nx-0.26*ny+0.94*nz), 32.0)

		// Single-order mother-of-pearl sheen
		filmNm := 340.0 + 75.0*(0.5+0.5*math.Sin(math.Pi*chamberX+0.8*tubeV-phi))
		tfR, tfG, tfB := render.SampleThinFilmLinearRGB(nz, filmNm, 1.55)
		irid := [3]float64{
			common.Clamp01(tfR),
			common.Clamp01(tfG),
			common.Clamp01(tfB),
		}

		// High-contrast 2x2-periodic museum palette across spiral chambers (100% invariant under Δu=-2, Δv=+12)
		uParity := chamberIdx & 1
		vParity := tubeIdx & 1
		var baseCol [3]float64
		switch {
		case uParity == 0 && vParity == 0:
			// Deep Prussian Lapis & Imperial Celadon with turquoise iridescence
			baseCol = lerp3(midnightCol, lerp3(indigoCol, celadonCol, 0.72), 0.25+0.75*ndotl)
			baseCol = lerp3(baseCol, irid, 0.22*(1.0-nz))
		case uParity == 1 && vParity == 1:
			// Cinnabar Vermilion & Imperial Amber-Gold lacquer
			baseCol = lerp3(indigoCol, lerp3(cinnabarCol, goldCol, 0.45*ndotl), 0.35+0.65*ndotl)
			baseCol = lerp3(baseCol, giltHiCol, 0.20*ndotl)
		default:
			// Luminous Museum Ivory & Warm Champagne Pearl with golden sheen
			baseCol = lerp3(sandCol, ivoryCol, 0.30+0.70*ndotl)
			baseCol = lerp3(baseCol, giltHiCol, 0.25*ndotl)
			baseCol = lerp3(baseCol, irid, 0.16*(1.0-nz))
		}

		// Cast shadow from the concave golden septum onto the chamber below it
		if chamberX > 0.60 {
			sh := common.Smoothstep(0.60, 0.88, chamberX)
			baseCol = lerp3(baseCol, midnightCol, 0.45*sh)
		}

		// Fine longitudinal growth ribs along the shell
		ribWave := 0.5 + 0.5*math.Cos(12.0*math.Pi*tubeV)
		groove := (1.0 - common.Smoothstep(0.25, 0.55, ribWave)) * 0.22 * detailVis
		col := lerp3(baseCol, midnightCol, groove)
		col = lerp3(col, ivoryCol, common.Clamp01(0.72*spec))

		// Concave 24k Gold Septum Suture Wall (|chamberX| > 0.84)
		septaDist := 1.0 - math.Abs(chamberX)
		if septaDist < 0.16 {
			sCov := 1.0 - common.Smoothstep(0.11, 0.16, septaDist)
			sNorm := septaDist / 0.13
			sBevel := math.Sqrt(math.Max(0.0, 1.0-sNorm*sNorm))
			sCol := lerp3(goldCol, giltHiCol, 0.68*sBevel)
			if sNorm > 0.76 {
				sCol = lerp3(sCol, inkCol, 0.82)
			}
			col = lerp3(col, sCol, sCov)
		}

		// Helical Braided 24k Gold Guilloché Cables along |tubeV| > 0.80
		cableDist := 1.0 - math.Abs(tubeV)
		if cableDist < 0.20 {
			if cableDist >= 0.13 {
				sh := 1.0 - common.Smoothstep(0.13, 0.20, cableDist)
				col = lerp3(col, midnightCol, 0.52*sh)
			} else {
				cNorm := cableDist / 0.13
				braid := 0.5 + 0.5*math.Sin(14.0*math.Pi*uCont+3.5*tubeV-phi)
				cLight := math.Sqrt(math.Max(0.0, 1.0-cNorm*cNorm)) * (0.62 + 0.38*braid)
				cCol := lerp3(goldCol, giltHiCol, cLight)
				if cNorm > 0.78 {
					cCol = lerp3(cCol, inkCol, 0.85)
				}
				col = lerp3(col, cCol, 1.0-common.Smoothstep(0.10, 0.13, cableDist))
			}
		}

		return col
	}

	// Parallel row rendering with 2x2 SSAA
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	var wg sync.WaitGroup
	rowsPerWorker := (h + numWorkers - 1) / numWorkers

	for worker := 0; worker < numWorkers; worker++ {
		y0 := worker * rowsPerWorker
		y1 := y0 + rowsPerWorker
		if y1 > h {
			y1 = h
		}
		if y0 >= y1 {
			continue
		}

		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for y := startY; y < endY; y++ {
				for x := 0; x < w; x++ {
					var accR, accG, accB float64

					for s := 0; s < 4; s++ {
						px := float64(x) + subOffsets[s][0]
						py := float64(y) + subOffsets[s][1]

						zScreen := complex((px-0.5*fw)*invNorm, (py-0.5*fh)*invNorm)
						rScreen := cmplx.Abs(zScreen)
						theta := cmplx.Phase(zScreen)

						_, _, _, confScale := common.DrosteSpiralPullback(zScreen, edg.scaleRatio, effBranches, 1.0)

						// 100% Holomorphic Conformal Multiplication:
						// W = (kR + i * kI) * (ln(r) + i * theta) + 2.0 * phase
						lnR := math.Log(math.Max(rScreen, 1e-9))
						uCont := kR*lnR - kI*theta + 2.0*phase
						vCont := kI*lnR + kR*theta

						var col [3]float64
						if useWaveBirds {
							col = evalWaveBirds(uCont, vCont, rScreen, confScale)
						} else {
							col = evalNautilusChambers(uCont, vCont, confScale)
						}

						// Central 16-Pointed Compass-Rose Vortex Seal at rScreen < 0.030
						if rScreen < 0.030 {
							medCov := 1.0 - common.Smoothstep(0.022, 0.030, rScreen)
							nr := rScreen / 0.027
							thetaCompass := theta - phi/8.0

							var medCol [3]float64
							if nr > 0.82 {
								ticks := 0.5 + 0.5*math.Cos(24.0*thetaCompass)
								medCol = lerp3(goldCol, giltHiCol, ticks)
							} else {
								star := 0.46 + 0.22*math.Cos(8.0*thetaCompass)
								if nr < 0.24 {
									medCol = cinnabarCol
								} else if nr < star {
									medCol = giltHiCol
								} else {
									medCol = indigoCol
								}
							}
							col = lerp3(col, medCol, medCov)
						}

						accR += col[0]
						accG += col[1]
						accB += col[2]
					}

					rc.SetPixelLinear(x, y, accR*0.25, accG*0.25, accB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// Post-processing pass: 2.5D luminance relief, museum cotton paper texture, and radial vignette
	const seed int64 = 20261101
	rc.ApplyLuminanceRelief(0.24, -135.0, 48.0, 0.16)
	rc.ApplyPaperTexture(0.08, seed)
	rc.ApplyVignette(0.16)
	c.MergeRenderCanvas(rc)
}
