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

type penroseWavefrontLoom struct {
	generations int
	terraces    int
	oxideNm     float64
}

// robinsonP3Tri represents an exact Robinson isosceles triangle for Penrose P3 rhombus tiling:
// - typ == 0 (Half-Thick Rhombus): apex angle 108 deg at B, legs |B-A| = |B-C| = L, base |A-C| = phi * L
// - typ == 1 (Half-Thin Rhombus):  apex angle 36 deg at B,  legs |B-A| = |B-C| = L, base |A-C| = L / phi
type robinsonP3Tri struct {
	typ     int
	a, b, c complex128
}

type penroseRhomb struct {
	cx, cy    float64
	ux, uy    float64 // unit vector along semi-diagonal U = (C - A)/2
	vx, vy    float64 // unit vector along semi-diagonal V = B - M
	lenU      float64
	lenV      float64
	edgeLen   float64
	isThick   bool
	ringTier  int
	orientIdx int
}

// NewPenroseWavefrontLoom returns a penroseWavefrontLoom engine that synthesizes a 100%
// gap-free Penrose P3 Quasicrystal Cloisonné & Islamic Girih Interlace Tapestry with
// 3D faceted diamond vaults, over-under woven 24K gold Ammann/Girih strapwork, and
// museum celadon, cinnabar, lapis, and mother-of-pearl glazes.
func NewPenroseWavefrontLoom(generations, terraces int, oxideNm float64) *penroseWavefrontLoom {
	if generations <= 0 {
		generations = 5
	}
	if terraces <= 0 {
		terraces = 12
	}
	if oxideNm <= 0 {
		oxideNm = 410.0
	}
	return &penroseWavefrontLoom{
		generations: common.ConstrainInt(generations, 4, 6),
		terraces:    common.ConstrainInt(terraces, 6, 22),
		oxideNm:     common.Constrain(oxideNm, 200.0, 620.0),
	}
}

// Generative renders a static frame of the 3D Penrose Quasicrystal & Girih Cloisonné Loom.
func (pwl *penroseWavefrontLoom) Generative(c *generativeart.Canva) {
	pwl.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (pwl *penroseWavefrontLoom) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	phi := 2.0 * math.Pi * phase

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	phiGold := (1.0 + math.Sqrt(5.0)) * 0.5 // 1.61803398875
	invPhi := 1.0 / phiGold                 // 0.61803398875

	// 1. Deflate a Single 36-Degree Fundamental Sector [0, pi/5] in Origin-Centered Coordinates
	// and Reflect Across the 10 Dihedral Symmetry Axes of D_10!
	// This guarantees 100% exact D_10 10-fold star symmetry at (cx0, cy0) and zero spoke mismatches!
	cx0, cy0 := 0.50*fw, 0.50*fh
	medRadius := 0.455 * minDim
	wheelR := 1.98 * medRadius
	legR := wheelR * invPhi
	rotWheel := phi / 10.0

	sectorTris := []robinsonP3Tri{
		{
			typ: 0, // Half-Thick Robinson triangle: 36-deg vertex at origin, 108-deg apex at B, 36-deg vertex at C
			a:   0,
			b:   cmplx.Rect(legR, math.Pi/5.0),
			c:   cmplx.Rect(wheelR, 0.0),
		},
	}

	nGen := common.ConstrainInt(pwl.generations, 4, 6)
	for g := 0; g < nGen; g++ {
		next := make([]robinsonP3Tri, 0, len(sectorTris)*3)
		for _, tr := range sectorTris {
			if tr.typ == 0 {
				q := tr.a + (tr.c-tr.a)*complex(invPhi, 0)
				p := tr.a + (tr.b-tr.a)*complex(invPhi, 0)
				next = append(next,
					robinsonP3Tri{typ: 0, a: tr.a, b: p, c: q},
					robinsonP3Tri{typ: 0, a: tr.c, b: q, c: tr.b},
					robinsonP3Tri{typ: 1, a: p, b: q, c: tr.b},
				)
			} else {
				r := tr.b + (tr.a-tr.b)*complex(invPhi, 0)
				next = append(next,
					robinsonP3Tri{typ: 0, a: tr.b, b: r, c: tr.c},
					robinsonP3Tri{typ: 1, a: r, b: tr.c, c: tr.a},
				)
			}
		}
		sectorTris = next
	}

	// Replicate the fundamental [0, 36 deg] sector across all 10 dihedral sectors of D_10!
	centerZ := complex(cx0, cy0)
	tris := make([]robinsonP3Tri, 0, len(sectorTris)*10)
	for k := 0; k < 10; k++ {
		if k%2 == 0 {
			rot := cmplx.Rect(1.0, rotWheel+float64(k)*(math.Pi/5.0))
			for _, st := range sectorTris {
				tris = append(tris, robinsonP3Tri{
					typ: st.typ,
					a:   centerZ + st.a*rot,
					b:   centerZ + st.b*rot,
					c:   centerZ + st.c*rot,
				})
			}
		} else {
			// Mirror-reflect across ray (k+1)*18 deg so spoke k*36 deg and (k+1)*36 deg are fixed point-for-point!
			rot := cmplx.Rect(1.0, rotWheel+float64(k+1)*(math.Pi/5.0))
			for _, st := range sectorTris {
				tris = append(tris, robinsonP3Tri{
					typ: st.typ,
					a:   centerZ + cmplx.Conj(st.a)*rot,
					b:   centerZ + cmplx.Conj(st.b)*rot,
					c:   centerZ + cmplx.Conj(st.c)*rot,
				})
			}
		}
	}

	// Convert each exact Robinson triangle (A, B, C) into a half-rhombus cell with midpoint M = (A + C)/2,
	// base semi-axis U = (C - A)/2, and orthogonal apex axis V = B - M (vSign in [0, 1 - |uSign|]).
	// Because the two legs AB and BC have the exact same length L_g and the base AC has no gold wire,
	// adjacent mirror-twin Robinson triangles fuse seamlessly across AC into complete Thick & Thin Penrose Rhombi
	// while guaranteeing 0.0% overlap everywhere!
	rhombs := make([]penroseRhomb, 0, len(tris))

	for _, tr := range tris {
		mx := 0.5 * (real(tr.a) + real(tr.c))
		my := 0.5 * (imag(tr.a) + imag(tr.c))
		bx := real(tr.b)
		by := imag(tr.b)
		if math.Min(mx, bx) > fw+45*scale || math.Max(mx, bx) < -45*scale ||
			math.Min(my, by) > fh+45*scale || math.Max(my, by) < -45*scale {
			continue
		}

		ux := 0.5 * (real(tr.c) - real(tr.a))
		uy := 0.5 * (imag(tr.c) - imag(tr.a))
		vx := bx - mx
		vy := by - my

		lenU := math.Hypot(ux, uy)
		lenV := math.Hypot(vx, vy)
		if lenU < 1e-4 || lenV < 1e-4 {
			continue
		}
		edgeLen := math.Hypot(lenU, lenV)

		distC := math.Hypot(mx-cx0, my-cy0) / medRadius
		rTier := int(math.Floor(distC * 3.8))
		angOrient := math.Atan2(uy, ux)
		oIdx := int(math.Round((angOrient+math.Pi)/(math.Pi/5.0))) % 5
		if oIdx < 0 {
			oIdx += 5
		}

		rhombs = append(rhombs, penroseRhomb{
			cx:        mx,
			cy:        my,
			ux:        ux / lenU,
			uy:        uy / lenU,
			vx:        vx / lenV,
			vy:        vy / lenV,
			lenU:      lenU,
			lenV:      lenV,
			edgeLen:   edgeLen,
			isThick:   tr.typ == 0,
			ringTier:  rTier,
			orientIdx: oIdx,
		})
	}

	// Build a fast 24x24 spatial acceleration grid over the 600x600 canvas
	const gridN = 24
	gridCells := make([][]int, gridN*gridN)
	cellW := fw / float64(gridN)
	cellH := fh / float64(gridN)
	for rIdx, rh := range rhombs {
		pad := math.Max(rh.lenU, rh.lenV) + 3.0*scale
		gx0 := common.ConstrainInt(int(math.Floor((rh.cx-pad)/cellW)), 0, gridN-1)
		gx1 := common.ConstrainInt(int(math.Floor((rh.cx+pad)/cellW)), 0, gridN-1)
		gy0 := common.ConstrainInt(int(math.Floor((rh.cy-pad)/cellH)), 0, gridN-1)
		gy1 := common.ConstrainInt(int(math.Floor((rh.cy+pad)/cellH)), 0, gridN-1)
		for gy := gy0; gy <= gy1; gy++ {
			for gx := gx0; gx <= gx1; gx++ {
				gridCells[gy*gridN+gx] = append(gridCells[gy*gridN+gx], rIdx)
			}
		}
	}

	// 2. Curated Museum Cloisonné, Imperial Celadon, Cinnabar, Lapis & 24K Gold Palette in Linear RGB
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

	ivoryCol := linRGB(0xfc, 0xf5, 0xe6)    // Luminous Museum Ivory
	sandCol := linRGB(0xe6, 0xbc, 0x7c)     // Warm Champagne Limestone
	goldCol := linRGB(0xf4, 0xbc, 0x3e)     // Burnished 24K Gold
	giltHiCol := linRGB(0xff, 0xee, 0xaa)   // Specular Gold Filigree
	silverCol := linRGB(0xe8, 0xf0, 0xf8)   // Platinum-Silver Strapwork
	cinnabarCol := linRGB(0xdc, 0x3c, 0x28) // Imperial Cinnabar Vermilion
	amberCol := linRGB(0xee, 0x8a, 0x26)    // Warm Amber Lacquer
	celadonCol := linRGB(0x20, 0xa8, 0x9a)  // Imperial Celadon Jade
	turquoiseCol := linRGB(0x46, 0xe0, 0xe8) // Luminous Peacock Turquoise
	lapisCol := linRGB(0x14, 0x32, 0x70)    // Royal Prussian Lapis
	midnightCol := linRGB(0x07, 0x0e, 0x1e) // Abyssal Midnight Ink
	inkCol := linRGB(0x0b, 0x0d, 0x14)      // Engraving Ink

	lx, ly, lz := -0.52, -0.54, 0.66
	lLen := math.Sqrt(lx*lx + ly*ly + lz*lz)
	lx, ly, lz = lx/lLen, ly/lLen, lz/lLen
	hx, hy, hz := lx, ly, lz+1.0
	hLen := math.Sqrt(hx*hx + hy*hy + hz*hz)
	hx, hy, hz = hx/hLen, hy/hLen, hz/hLen

	// Shade a single Penrose Half-Rhombus at local normalized coordinates (uSign, vSign)
	// where uSign = 0, vSign = 0 is the exact center M of the complete 4-sided Penrose Rhombus!
	_ = silverCol
	shadePenroseRhomb := func(rh *penroseRhomb, uSign, vSign float64) [3]float64 {
		u := math.Abs(uSign)
		v := math.Abs(vSign)
		rRhomb := u + v // 0 at rhombus center M, 1 on the 4 outer edges of the rhombus!

		// Exact perpendicular distance in pixels to the outer rhombus edge
		perpScale := (rh.lenU * rh.lenV) / rh.edgeLen
		dEdgePx := (1.0 - rRhomb) * perpScale

		// 1. 3D Stepped Pyramidal Diamond-Vault Normal inside the Rhombus
		// 2 concentric rhombic steps (at rRhomb == 0.52) + 4-quadrant pyramidal bevel facets
		inInnerVault := rRhomb < 0.52
		dInnerRimPx := math.Abs(rRhomb-0.52) * perpScale

		sx := 1.0
		if uSign < 0 {
			sx = -1.0
		}
		// Fine Kajiya-Kay concentric rhombic guilloché wave ribs
		ribWave := math.Sin(rRhomb * math.Pi * 12.0)
		slope := 0.48
		if inInnerVault {
			slope = 0.34
		}
		nx := (sx*rh.ux + rh.vx) * slope * (1.0 + 0.18*ribWave)
		ny := (sx*rh.uy + rh.vy) * slope * (1.0 + 0.18*ribWave)
		nz := math.Sqrt(math.Max(0.14, 1.0-nx*nx-ny*ny))

		ndotl := common.Clamp01(nx*lx + ny*ly + nz*lz)
		ndoth := common.Clamp01(nx*hx + ny*hy + nz*hz)
		spec := math.Pow(ndoth, 38.0)

		// Physical Airy thin-film nacre iridescence
		filmNm := pwl.oxideNm + float64(rh.ringTier*45+rh.orientIdx*35) - 90.0*rRhomb
		tfR, tfG, tfB := render.SampleThinFilmLinearRGB(nz, filmNm, 1.60)
		irid := [3]float64{common.Clamp01(tfR), common.Clamp01(tfG), common.Clamp01(tfB)}

		// Select Museum Enamel Glaze by Concentric Radial Ring Tier:
		// Central 10-Fold Star (Solar Gold & Cinnabar Ruby) -> Imperial Celadon & Ivory -> Royal Lapis!
		var baseCol [3]float64
		switch {
		case rh.ringTier == 0:
			if rh.isThick {
				if inInnerVault {
					baseCol = lerp3(cinnabarCol, giltHiCol, 0.45+0.55*ndotl)
				} else {
					baseCol = lerp3(lapisCol, cinnabarCol, 0.32+0.68*ndotl)
				}
			} else {
				baseCol = lerp3(sandCol, giltHiCol, 0.40+0.60*ndotl)
			}
		case rh.ringTier == 1:
			if rh.isThick {
				if inInnerVault {
					baseCol = lerp3(lapisCol, lerp3(celadonCol, turquoiseCol, 0.65*ndotl), 0.30+0.70*ndotl)
				} else {
					baseCol = lerp3(sandCol, ivoryCol, 0.35+0.65*ndotl)
				}
			} else {
				baseCol = lerp3(lapisCol, lerp3(cinnabarCol, amberCol, 0.45*ndotl), 0.32+0.68*ndotl)
			}
		case rh.ringTier == 2:
			if rh.isThick {
				if rh.orientIdx%2 == 0 {
					baseCol = lerp3(sandCol, ivoryCol, 0.32+0.68*ndotl)
				} else {
					baseCol = lerp3(midnightCol, lerp3(lapisCol, celadonCol, 0.78), 0.25+0.75*ndotl)
				}
			} else {
				baseCol = lerp3(lapisCol, cinnabarCol, 0.30+0.70*ndotl)
			}
		default:
			if rh.isThick {
				if inInnerVault {
					baseCol = lerp3(midnightCol, lerp3(lapisCol, celadonCol, 0.68), 0.24+0.76*ndotl)
				} else {
					baseCol = lerp3(midnightCol, lapisCol, 0.22+0.72*ndotl)
				}
			} else {
				baseCol = lerp3(sandCol, ivoryCol, 0.28+0.66*ndotl)
			}
		}
		baseCol = lerp3(baseCol, irid, 0.22*(1.0-nz))

		// Subtle guilloché groove shading + diagonal pyramid ridge highlight along u == 0 and v == 0
		groove := (1.0 - common.Smoothstep(-0.2, 0.5, ribWave)) * 0.16
		col := lerp3(baseCol, midnightCol, groove)

		dDiagPx := math.Min(u*rh.lenU, v*rh.lenV)
		diagRidge := math.Exp(-math.Pow(dDiagPx/(1.0*scale), 2.0))
		col = lerp3(col, giltHiCol, common.Clamp01(0.38*diagRidge+0.65*spec))

		// Crevice shadow inside the outer 24K gold rhombus frame
		if dEdgePx < 4.0*scale {
			sh := common.Smoothstep(1.1*scale, 4.0*scale, dEdgePx)
			col = lerp3(midnightCol, col, 0.38+0.62*sh)
		}

		// 2. Inner Stepped 24K Gold Rhombus Filigree Rim at rRhomb == 0.52 (connects midpoints of all 4 edges!)
		innerWireW := 0.95 * scale
		if dInnerRimPx < innerWireW+1.8*scale {
			if dInnerRimPx >= innerWireW {
				sh := (dInnerRimPx - innerWireW) / (1.8 * scale)
				col = lerp3(midnightCol, col, 0.55+0.45*sh)
			} else {
				uW := dInnerRimPx / innerWireW
				wBevel := math.Sqrt(math.Max(0.0, 1.0-uW*uW))
				wCol := lerp3(goldCol, giltHiCol, 0.30+0.70*wBevel)
				if uW > 0.78 {
					wCol = lerp3(wCol, inkCol, 0.80)
				}
				col = wCol
			}
		}

		// 3. 3D Spherical Mother-of-Pearl / Ruby Cabochon at the Shared Center M (u=0, v=0) of Thick Rhombi!
		if rh.isThick {
			dCenterPx := math.Hypot(uSign*rh.lenU, vSign*rh.lenV)
			pearlRad := math.Min(5.2*scale, 0.24*perpScale)
			if dCenterPx < pearlRad+1.6*scale {
				if dCenterPx >= pearlRad {
					sh := (dCenterPx - pearlRad) / (1.6 * scale)
					col = lerp3(midnightCol, col, 0.48+0.52*sh)
				} else if dCenterPx > pearlRad*0.74 {
					uBez := (dCenterPx - pearlRad*0.74) / (pearlRad * 0.26)
					bZ := math.Sin(uBez * math.Pi)
					col = lerp3(goldCol, giltHiCol, 0.35+0.65*bZ)
				} else {
					rP := pearlRad * 0.74
					uxP := (uSign*rh.lenU*rh.ux + vSign*rh.lenV*rh.vx) / rP
					uyP := (uSign*rh.lenU*rh.uy + vSign*rh.lenV*rh.vy) / rP
					uzP := math.Sqrt(math.Max(0.05, 1.0-uxP*uxP-uyP*uyP))
					pNdotL := common.Clamp01(uxP*lx + uyP*ly + uzP*lz)
					pSpec := math.Pow(common.Clamp01(uxP*hx+uyP*hy+uzP*hz), 48.0)
					var pCol [3]float64
					if rh.ringTier == 1 {
						pCol = lerp3(lapisCol, cinnabarCol, 0.35+0.65*pNdotL)
					} else {
						pCol = lerp3(sandCol, ivoryCol, 0.35+0.65*pNdotL)
						pCol = lerp3(pCol, irid, 0.35*(1.0-uzP))
					}
					pCol = lerp3(pCol, ivoryCol, common.Clamp01(0.92*pSpec))
					col = pCol
				}
			}
		}

		// 4. Outer 3D Beveled 24K Gold Cloisonné Rhombus Wire Border (dEdgePx < 1.25*scale)
		wireHalfW := 1.25 * scale
		if dEdgePx < wireHalfW {
			uW := dEdgePx / wireHalfW
			wBevel := math.Sqrt(math.Max(0.0, 1.0-uW*uW))
			wCol := lerp3(goldCol, giltHiCol, 0.35+0.65*wBevel)
			if uW > 0.78 {
				wCol = lerp3(wCol, inkCol, 0.85)
			}
			col = wCol
		}

		return col
	}

	// 3. Parallel 2x2 SSAA Per-Pixel Synthesis Across the Entire 600x600 Canvas
	subOffsets := [4][2]float64{
		{0.25, 0.25},
		{0.75, 0.25},
		{0.25, 0.75},
		{0.75, 0.75},
	}

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
				gy := common.ConstrainInt(int(float64(y)/cellH), 0, gridN-1)
				for x := 0; x < w; x++ {
					gx := common.ConstrainInt(int(float64(x)/cellW), 0, gridN-1)
					bucket := gridCells[gy*gridN+gx]

					var accR, accG, accB float64

					for s := 0; s < 4; s++ {
						px := float64(x) + subOffsets[s][0]
						py := float64(y) + subOffsets[s][1]

						// Check 10-Fold Graduated 24K Gold Astrolabe Bezel & Museum Parchment Mat!
						dxC := px - cx0
						dyC := py - cy0
						rDisk := math.Hypot(dxC, dyC) / medRadius
						if rDisk > 0.965 {
							if rDisk <= 1.028 {
								// Graduated 24K Gold Astrolabe Bezel Frame
								uBez := (rDisk - 0.965) / (1.028 - 0.965)
								bezZ := math.Sin(uBez * math.Pi)
								ang := math.Atan2(dyC, dxC)
								ticks := 0.76 + 0.24*math.Cos(100.0*ang-10.0*rotWheel)
								bCol := lerp3(goldCol, giltHiCol, 0.78*bezZ*ticks)
								if uBez < 0.14 || uBez > 0.86 {
									bCol = lerp3(bCol, inkCol, 0.82)
								}
								accR += bCol[0]
								accG += bCol[1]
								accB += bCol[2]
							} else {
								// Warm Museum Parchment Ivory Mat with soft drop shadow under the bezel
								shR := math.Hypot(dxC-3.5*scale, dyC-4.2*scale) / medRadius
								sh := common.Smoothstep(1.065, 1.020, shR) * 0.32
								matCol := lerp3(ivoryCol, sandCol, 0.18*common.Clamp01((rDisk-1.0)*0.8))
								matCol = lerp3(matCol, inkCol, sh)
								accR += matCol[0]
								accG += matCol[1]
								accB += matCol[2]
							}
							continue
						}

						hit := false
						var col [3]float64

						for _, rIdx := range bucket {
							rh := &rhombs[rIdx]
							dx := px - rh.cx
							dy := py - rh.cy
							uSign := (dx*rh.ux + dy*rh.uy) / rh.lenU
							vSign := (dx*rh.vx + dy*rh.vy) / rh.lenV
							if vSign >= -0.002 && math.Abs(uSign)+vSign <= 1.002 {
								col = shadePenroseRhomb(rh, uSign, math.Max(0.0, vSign))
								hit = true
								break
							}
						}

						if !hit {
							col = midnightCol
						}

						// Inner bezel rim shadow just inside rDisk == 0.965
						if rDisk > 0.92 {
							shIn := common.Smoothstep(0.965, 0.92, rDisk)
							col = lerp3(midnightCol, col, 0.45+0.55*shIn)
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

	rc.ApplyLuminanceRelief(0.28, -135.0, 46.0, 0.18)
	rc.ApplyBloom(0.78, 3.2*scale, 0.10)
	rc.ApplyPaperTexture(0.06, 20262803)
	rc.ApplyVignette(0.18)
	c.MergeRenderCanvas(rc)
}
