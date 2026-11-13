package arts

import (
	"image/color"
	"math"
	"math/cmplx"
	"runtime"
	"sort"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type apollonianGasketFoam struct {
	depth    int
	symmetry int
	filmNm   float64
}

type soddyCircle struct {
	k      float64    // signed bend (curvature = +-1/r)
	kz     complex128 // curvature * center (k * z)
	cx     float64
	cy     float64
	r      float64
	gen    int // recursion generation
	level  int // nesting level (0: primary, 1: nested inner, 2: deep inner)
	branch int // 0, 1, 2 for the 3 primary Soddy triad branches, 3 for central/interstitial
}

// NewApollonianGasketFoam returns an apollonianGasketFoam engine that synthesizes a full-canvas
// Descartes-Soddy Apollonian sphere packing (extended to 100% of the canvas via Kelvin circle
// inversion and nested sub-gaskets), where every tangent circle is a 3D beveled 24K gold-rimmed
// dielectric glass & thin-film opal sphere with Snell lens refraction and cardioid caustics.
func NewApollonianGasketFoam(depth, symmetry int, filmNm float64) *apollonianGasketFoam {
	if depth <= 0 {
		depth = 6
	}
	if symmetry <= 0 {
		symmetry = 3
	}
	if filmNm <= 0 {
		filmNm = 380.0
	}
	return &apollonianGasketFoam{
		depth:    common.ConstrainInt(depth, 3, 7),
		symmetry: common.ConstrainInt(symmetry, 3, 6),
		filmNm:   common.Constrain(filmNm, 220.0, 620.0),
	}
}

// Generative renders a static frame of the 3D Refractive Apollonian Gasket Foam.
func (agf *apollonianGasketFoam) Generative(c *generativeart.Canva) {
	agf.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree Möbius rotation loop (phase in [0, 1)).
func (agf *apollonianGasketFoam) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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

	// 1. Generate Exact Canonical Apollonian Gasket Inside the Unit Disk |z| <= 1
	// Using Vieta's linear reflection formula for Descartes' Circle Theorem:
	//   k4' = 2*(k1 + k2 + k3) - k4
	//   (k4'*z4') = 2*(k1*z1 + k2*z2 + k3*z3) - k4*z4
	maxGen := agf.depth
	minRadWorld := 0.0045

	makeSoddy := func(k float64, kz complex128, gen, level, branch int) soddyCircle {
		var z complex128
		if math.Abs(k) > 1e-12 {
			z = kz / complex(k, 0)
		}
		return soddyCircle{
			k:      k,
			kz:     kz,
			cx:     real(z),
			cy:     imag(z),
			r:      1.0 / math.Abs(k),
			gen:    gen,
			level:  level,
			branch: branch,
		}
	}

	reflectVieta := func(c1, c2, c3, c4 soddyCircle, gen, level, branch int) soddyCircle {
		kNew := 2.0*(c1.k+c2.k+c3.k) - c4.k
		kzNew := 2.0*(c1.kz+c2.kz+c3.kz) - c4.kz
		return makeSoddy(kNew, kzNew, gen, level, branch)
	}

	// Build canonical unit-disk gasket rotated by angle rot0
	buildUnitGasket := func(rot0 float64, maxDepth, level int, minR float64) []soddyCircle {
		var list []soddyCircle
		c0 := makeSoddy(-1.0, 0, 0, level, 3)

		rSoddy := 2.0*math.Sqrt(3.0) - 3.0
		dSoddy := 1.0 - rSoddy
		kSoddy := 1.0 / rSoddy

		var cTri [3]soddyCircle
		for m := 0; m < 3; m++ {
			ang := rot0 + float64(m)*(2.0*math.Pi/3.0)
			zm := cmplx.Rect(dSoddy, ang)
			cTri[m] = makeSoddy(kSoddy, complex(kSoddy, 0)*zm, 1, level, m)
			list = append(list, cTri[m])
		}

		r4 := 7.0 - 4.0*math.Sqrt(3.0)
		k4 := 1.0 / r4
		c4 := makeSoddy(k4, 0, 1, level, 3)
		list = append(list, c4)

		type quadState struct {
			a, b, c, d soddyCircle
			depth      int
		}

		queue := []quadState{
			{cTri[0], cTri[1], cTri[2], c4, 1},
			{c0, cTri[0], cTri[1], cTri[2], 1},
		}

		for len(queue) > 0 {
			q := queue[0]
			queue = queue[1:]
			if q.depth >= maxDepth {
				continue
			}

			candidates := [4][4]soddyCircle{
				{q.a, q.b, q.c, q.d},
				{q.a, q.b, q.d, q.c},
				{q.a, q.c, q.d, q.b},
				{q.b, q.c, q.d, q.a},
			}
			for cIdx, cand := range candidates {
				cNew := reflectVieta(cand[0], cand[1], cand[2], cand[3], q.depth+1, level, (q.depth+cIdx)%4)
				if cNew.k <= cand[3].k+1e-6 || cNew.r < minR {
					continue
				}
				dup := false
				for i := len(list) - 1; i >= 0; i-- {
					if math.Abs(list[i].k-cNew.k) < 1e-4 &&
						math.Hypot(list[i].cx-cNew.cx, list[i].cy-cNew.cy) < 1e-4 {
						dup = true
						break
					}
				}
				if dup {
					continue
				}
				list = append(list, cNew)
				queue = append(queue, quadState{cand[0], cand[1], cand[2], cNew, q.depth + 1})
			}
		}
		return list
	}

	rotBase := phi/3.0 + 0.26
	baseCircles := buildUnitGasket(rotBase, maxGen, 0, minRadWorld)

	mainScale := 0.445 * minDim
	cx0, cy0 := 0.50*fw, 0.50*fh

	var allCircles []soddyCircle
	allCircles = append(allCircles, soddyCircle{
		cx: cx0, cy: cy0, r: mainScale, gen: 0, level: -1, branch: 3,
	})

	subTemplate1 := buildUnitGasket(-rotBase*2.0+0.52, common.MaxInt(4, maxGen-1), 1, 0.014)
	subTemplate2 := buildUnitGasket(rotBase*3.0-0.35, common.MaxInt(3, maxGen-2), 2, 0.026)

	for _, bc := range baseCircles {
		sc := soddyCircle{
			cx:     cx0 + bc.cx*mainScale,
			cy:     cy0 + bc.cy*mainScale,
			r:      bc.r * mainScale,
			gen:    bc.gen,
			level:  0,
			branch: bc.branch,
		}
		allCircles = append(allCircles, sc)

		// Level-1 & Level-2 Recursive Nested Sub-Gaskets inside larger spheres!
		if bc.r > 0.075 {
			innerScale1 := sc.r * 0.89
			for _, sub1 := range subTemplate1 {
				r1Px := sub1.r * innerScale1
				if r1Px < 1.9*scale {
					continue
				}
				c1 := soddyCircle{
					cx:     sc.cx + sub1.cx*innerScale1,
					cy:     sc.cy + sub1.cy*innerScale1,
					r:      r1Px,
					gen:    bc.gen + sub1.gen,
					level:  1,
					branch: (bc.branch + sub1.branch) % 4,
				}
				allCircles = append(allCircles, c1)

				if r1Px > 16.0*scale && sub1.r > 0.12 {
					innerScale2 := r1Px * 0.88
					for _, sub2 := range subTemplate2 {
						r2Px := sub2.r * innerScale2
						if r2Px < 1.7*scale {
							continue
						}
						allCircles = append(allCircles, soddyCircle{
							cx:     c1.cx + sub2.cx*innerScale2,
							cy:     c1.cy + sub2.cy*innerScale2,
							r:      r2Px,
							gen:    c1.gen + sub2.gen,
							level:  2,
							branch: (c1.branch + sub2.branch) % 4,
						})
					}
				}
			}
		}
	}

	// 4 Symmetric Corner Spandrel Sub-Gasket Medallions tangent to the outer astrolabe bezel in the 4 corners!
	cornerDist := 0.576 * minDim
	cornerRad := 0.114 * minDim
	for cIdx := 0; cIdx < 4; cIdx++ {
		cAng := float64(cIdx)*(math.Pi*0.5) + math.Pi*0.25
		ccx := cx0 + math.Cos(cAng)*cornerDist
		ccy := cy0 + math.Sin(cAng)*cornerDist
		allCircles = append(allCircles, soddyCircle{
			cx: ccx, cy: ccy, r: cornerRad, gen: 1, level: 0, branch: cIdx % 4,
		})
		cSubScale := cornerRad * 0.88
		for _, sub1 := range subTemplate1 {
			if sub1.r*cSubScale < 1.8*scale {
				continue
			}
			allCircles = append(allCircles, soddyCircle{
				cx:     ccx + sub1.cx*cSubScale,
				cy:     ccy + sub1.cy*cSubScale,
				r:      sub1.r * cSubScale,
				gen:    1 + sub1.gen,
				level:  1,
				branch: (cIdx + sub1.branch) % 4,
			})
		}
	}

	// Sort circles from smallest radius to largest so the innermost/deepest containing sphere is tested first!
	sort.Slice(allCircles, func(i, j int) bool {
		return allCircles[i].r < allCircles[j].r
	})

	// Build a fast 24x24 spatial acceleration grid over the canvas so per-pixel lookup is O(1)!
	const gridN = 24
	gridCells := make([][]int, gridN*gridN)
	cellW := fw / float64(gridN)
	cellH := fh / float64(gridN)

	for idx, cObj := range allCircles {
		if cObj.level == -1 {
			continue
		}
		pad := cObj.r + 5.0*scale
		gx0 := common.ConstrainInt(int(math.Floor((cObj.cx-pad)/cellW)), 0, gridN-1)
		gx1 := common.ConstrainInt(int(math.Floor((cObj.cx+pad)/cellW)), 0, gridN-1)
		gy0 := common.ConstrainInt(int(math.Floor((cObj.cy-pad)/cellH)), 0, gridN-1)
		gy1 := common.ConstrainInt(int(math.Floor((cObj.cy+pad)/cellH)), 0, gridN-1)
		for gy := gy0; gy <= gy1; gy++ {
			for gx := gx0; gx <= gx1; gx++ {
				gridCells[gy*gridN+gx] = append(gridCells[gy*gridN+gx], idx)
			}
		}
	}

	// 2. Curated Museum Cloisonné, Opal, Celadon, Cinnabar & 24K Gold Palette in Linear RGB
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

	ivoryCol := linRGB(0xfc, 0xf5, 0xe6)
	sandCol := linRGB(0xe6, 0xbc, 0x7c)
	goldCol := linRGB(0xf4, 0xbc, 0x40)
	giltHiCol := linRGB(0xff, 0xee, 0xb0)
	cinnabarCol := linRGB(0xdc, 0x3c, 0x2a)
	celadonCol := linRGB(0x20, 0xa6, 0x98)
	turquoiseCol := linRGB(0x46, 0xe0, 0xe8)
	lapisCol := linRGB(0x14, 0x32, 0x70)
	midnightCol := linRGB(0x08, 0x10, 0x22)
	inkCol := linRGB(0x0b, 0x0d, 0x14)

	lx, ly, lz := -0.50, -0.54, 0.68
	lLen := math.Sqrt(lx*lx + ly*ly + lz*lz)
	lx, ly, lz = lx/lLen, ly/lLen, lz/lLen
	hx, hy, hz := lx, ly, lz+1.0
	hLen := math.Sqrt(hx*hx + hy*hy + hz*hz)
	hx, hy, hz = hx/hLen, hy/hLen, hz/hLen

	// Evaluate the engraved 24K gold & lapis astrolabe substrate at (px, py) (seen in interstices & refracted through spheres!)
	evalSubstratePattern := func(px, py float64) [3]float64 {
		dx := (px - cx0) / mainScale
		dy := (py - cy0) / mainScale
		r := math.Hypot(dx, dy)
		ang := math.Atan2(dy, dx)

		// Crisp engraved celestial planisphere concentric rings & radial meridian lines
		ringWave := math.Pow(0.5+0.5*math.Cos(r*math.Pi*28.0), 14.0)
		spokeWave := math.Pow(0.5+0.5*math.Cos(36.0*ang-3.0*phi), 16.0)
		grid := math.Max(ringWave, spokeWave) * common.Smoothstep(1.42, 0.85, r)

		base := lerp3(midnightCol, lapisCol, (0.28+0.18*math.Cos(6.0*ang))*common.Smoothstep(1.48, 0.75, r))
		return lerp3(base, goldCol, 0.24*grid)
	}

	// Shade a single 3D Dielectric Glass / Opal Apollonian Sphere cObj at pixel (px, py)
	shadeSphere := func(cObj *soddyCircle, px, py float64) [3]float64 {
		dx := px - cObj.cx
		dy := py - cObj.cy
		rNorm := math.Hypot(dx, dy) / cObj.r
		ux := dx / cObj.r
		uy := dy / cObj.r

		// Outer 3D Beveled 24K Gold Collet Bezel Rim (rNorm in [0.89, 1.0])
		bezelStart := 0.89
		if cObj.r < 12.0*scale {
			bezelStart = 0.82
		}
		if rNorm >= bezelStart {
			uBez := (rNorm - bezelStart) / (1.0 - bezelStart)
			bezZ := math.Sin(common.Clamp01(uBez) * math.Pi)
			ang := math.Atan2(dy, dx)
			bead := 0.78 + 0.22*math.Cos(24.0*ang)
			gCol := lerp3(goldCol, giltHiCol, 0.75*bezZ*bead)
			if uBez > 0.82 || uBez < 0.14 {
				gCol = lerp3(gCol, inkCol, 0.72)
			}
			return gCol
		}

		// Inside the 3D Dielectric Glass & Precious Opal Sphere (rInner in [0, 1])
		rIn := rNorm / bezelStart
		uxIn := ux / bezelStart
		uyIn := uy / bezelStart
		uzIn := math.Sqrt(math.Max(0.04, 1.0-uxIn*uxIn-uyIn*uyIn))

		ndotl := common.Clamp01(uxIn*lx + uyIn*ly + uzIn*lz)
		ndoth := common.Clamp01(uxIn*hx + uyIn*hy + uzIn*hz)
		spec := math.Pow(ndoth, 56.0)
		secSpec := math.Pow(common.Clamp01(-uxIn*lx-uyIn*ly+uzIn*lz), 24.0) * 0.35

		// 1. 3D Snell Dielectric Lens Refraction of the underlying engraved astrolabe substrate!
		refrX := px - uxIn*cObj.r*0.42*(1.0-0.35*uzIn)
		refrY := py - uyIn*cObj.r*0.42*(1.0-0.35*uzIn)
		refrCol := evalSubstratePattern(refrX, refrY)

		// 2. Physical Airy Thin-Film Interference across the 3D sphere curvature
		localFilm := agf.filmNm + float64((cObj.gen*65+cObj.level*110)%240) - 90.0*(1.0-uzIn)
		tfR, tfG, tfB := render.SampleThinFilmLinearRGB(uzIn, localFilm, 1.58)
		irid := [3]float64{common.Clamp01(tfR), common.Clamp01(tfG), common.Clamp01(tfB)}

		// 3. Internal Cardioid Reflective Caustic Crescent (focuses opposite the key light!)
		angIn := math.Atan2(uyIn, uxIn)
		lightAng := math.Atan2(ly, lx)
		causticRadius := 0.56 - 0.18*math.Cos(angIn-lightAng)
		causticArc := math.Exp(-math.Pow((rIn-causticRadius)/0.075, 2.0)) *
			math.Pow(0.5+0.5*math.Cos(angIn-(lightAng+math.Pi)), 2.0)

		// Select jewel body tint by branch & nesting level so the 3 main branches form a balanced
		// Triad of Imperial Celadon Jade, Cinnabar Vermilion Ruby & Royal Prussian Sapphire!
		palMod := cObj.branch % 4
		var jewelBase [3]float64
		switch palMod {
		case 0:
			// Imperial Celadon Jade & Peacock Turquoise Glass
			jewelBase = lerp3(midnightCol, lerp3(lapisCol, celadonCol, 0.78), 0.20+0.80*ndotl)
		case 1:
			// Imperial Cinnabar Ruby & Burnished Gold Glass
			jewelBase = lerp3(midnightCol, lerp3(cinnabarCol, goldCol, 0.38*ndotl), 0.24+0.76*ndotl)
		case 2:
			// Royal Sapphire & Turquoise Chalcedony Glass
			jewelBase = lerp3(midnightCol, lerp3(lapisCol, turquoiseCol, 0.56*ndotl), 0.22+0.78*ndotl)
		default:
			// Luminous Museum Ivory & Fire-Opal Nacre
			jewelBase = lerp3(sandCol, ivoryCol, 0.30+0.70*ndotl)
		}

		// Combine jewel body, refracted lens image, Fresnel thin-film opalescence, and internal caustic fire!
		fresnel := math.Pow(1.0-uzIn, 2.5)
		col := lerp3(jewelBase, refrCol, 0.26*uzIn)
		col = lerp3(col, irid, 0.16+0.46*fresnel)
		col = lerp3(col, giltHiCol, common.Clamp01(0.75*causticArc+0.35*secSpec))

		// Crevice shadow just inside the raised 24K gold bezel rim
		rimShadow := common.Smoothstep(1.0, 0.84, rIn)
		col = lerp3(midnightCol, col, 0.32+0.68*rimShadow)

		// Razor-sharp studio window specular highlight
		col = lerp3(col, ivoryCol, common.Clamp01(0.96*spec))
		return col
	}

	// 3. Parallel 2x2 SSAA Per-Pixel Rendering Across the Full 600x600 Canvas
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
				for x := 0; x < w; x++ {
					gx := common.ConstrainInt(int(float64(x)/cellW), 0, gridN-1)
					gy := common.ConstrainInt(int(float64(y)/cellH), 0, gridN-1)
					bucket := gridCells[gy*gridN+gx]

					var accR, accG, accB float64

					for s := 0; s < 4; s++ {
						px := float64(x) + subOffsets[s][0]
						py := float64(y) + subOffsets[s][1]

						// Check the main Kelvin-inversion astrolabe ring at r == 1.018 * mainScale
						distCenter := math.Hypot(px-cx0, py-cy0)
						ringDist := math.Abs(distCenter - 1.018*mainScale)
						if ringDist < 5.5*scale {
							uRing := ringDist / (5.5 * scale)
							ringZ := math.Sqrt(math.Max(0.0, 1.0-uRing*uRing))
							ang := math.Atan2(py-cy0, px-cx0)
							ticks := 0.72 + 0.28*math.Cos(72.0*ang-6.0*phi)
							rCol := lerp3(goldCol, giltHiCol, 0.80*ringZ*ticks)
							if uRing > 0.80 {
								rCol = lerp3(rCol, inkCol, 0.82)
							}
							accR += rCol[0]
							accG += rCol[1]
							accB += rCol[2]
							continue
						}

						// Find smallest containing Apollonian sphere (since allCircles is sorted smallest-first,
						// nested level-1 spheres are automatically found before their level-0 parent sphere!)
						hitIdx := -1
						minShadow := 1.0

						for _, cIdx := range bucket {
							cObj := &allCircles[cIdx]
							dx := px - cObj.cx
							dy := py - cObj.cy
							d := math.Hypot(dx, dy)
							if d <= cObj.r {
								if hitIdx < 0 {
									hitIdx = cIdx
								}
							} else if d < cObj.r+4.2*scale {
								// Soft drop shadow cast by adjacent spheres
								sh := (d - cObj.r) / (4.2 * scale)
								if sh < minShadow {
									minShadow = sh
								}
							}
						}

						var col [3]float64
						if hitIdx >= 0 {
							col = shadeSphere(&allCircles[hitIdx], px, py)
						} else {
							col = evalSubstratePattern(px, py)
							col = lerp3(midnightCol, col, 0.35+0.65*minShadow)
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

	rc.ApplyLuminanceRelief(0.24, -135.0, 46.0, 0.16)
	rc.ApplyBloom(0.76, 3.4*scale, 0.12)
	rc.ApplyPaperTexture(0.06, 20262802)
	rc.ApplyVignette(0.18)
	c.MergeRenderCanvas(rc)
}
