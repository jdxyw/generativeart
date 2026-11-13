package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type lorenzWatercolorSilk struct {
	vortices   int
	fiberCount int
	filmNm     float64
}

// NewLorenzWatercolorSilk returns a lorenzWatercolorSilk engine that synthesizes a 3D
// Alhambra & Isfahan Muqarnas Honeycomb Stalactite Vault Dome with 14 corbelled tiers of
// 3D concave fan-ribbed niches, stalactite pendants, clerestory stained-glass lancets,
// dual-sun horizon-marched soft shadows, 12-tap vault AO, and 24K gold-leaf & Persian
// lapis/turquoise/cinnabar glazed ceramic tilework.
func NewLorenzWatercolorSilk(vortices, fiberCount int, filmNm float64) *lorenzWatercolorSilk {
	if vortices <= 0 {
		vortices = 8
	}
	if fiberCount <= 0 {
		fiberCount = 20000
	}
	if filmNm <= 0 {
		filmNm = 395.0
	}
	return &lorenzWatercolorSilk{
		vortices:   common.ConstrainInt(vortices, 6, 16),
		fiberCount: common.ConstrainInt(fiberCount, 3000, 30000),
		filmNm:     common.Constrain(filmNm, 220.0, 620.0),
	}
}

// Generative renders a static frame of the 3D Alhambra Muqarnas Stalactite Vault Dome.
func (lws *lorenzWatercolorSilk) Generative(c *generativeart.Canva) {
	lws.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (lws *lorenzWatercolorSilk) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	phi := 2.0 * math.Pi * phase

	rc := c.NewRenderCanvas()
	rc.FillBackground(color.RGBA{R: 8, G: 11, B: 22, A: 255})

	cx0, cy0 := 0.50*fw, 0.50*fh
	domeScale := 0.48 * minDim

	// Exact 16-fold rotational symmetry period: advancing by phi/16 is 100% seamless!
	rotDome := phi / 16.0

	// 1. Evaluate the 3D Corbelled Muqarnas Stalactite Vault Geometry at (px, py)
	// A real Islamic Muqarnas dome is organized into concentric corbelled tiers k = 0..13 from
	// the central 16-pointed shamsa cupola (r = 0) out to the octagonal squinch walls (r > 1):
	// Each tier k has N_k angular stalactite niches (16, 16, 24, 24, 32, 32...) where odd tiers
	// shift angularly by half a cell (pi / N_k) so each upper niche corbels directly over the
	// stalactite pendant vertex between two lower niches!
	evalMuqarnasVault := func(px, py float64) (zElev, ribRidge, goldWire, tilePattern float64, tierIdx, cellParity int) {
		dx := (px - cx0) / domeScale
		dy := (py - cy0) / domeScale
		rEuc := math.Hypot(dx, dy)
		theta := math.Atan2(dy, dx) + rotDome

		// Octagonal-to-hexadecagonal squinch star modulation of the radial coordinate
		octStar := 0.032*math.Cos(8.0*theta) + 0.018*math.Cos(16.0*theta)
		r := rEuc * (1.0 + octStar)

		// Central 16-Pointed Shamsa Sunburst Lantern Cupola (r < 0.13)
		if r < 0.13 {
			uCup := r / 0.13
			spoke16 := math.Abs(math.Sin(16.0 * theta))
			star8 := 0.52 + 0.24*math.Cos(8.0*theta)
			dStar := math.Abs(uCup - star8)

			// Deeply fluted 16-ribbed hemispherical lantern dome
			domeZ := math.Sqrt(math.Max(0.0, 1.0-uCup*uCup))*14.0 + 3.5*math.Sqrt(spoke16)*(1.0-0.5*uCup)
			zElev = (26.0 + domeZ) * scale
			ribRidge = math.Exp(-math.Pow((1.0-spoke16)/0.12, 2.0)) + math.Exp(-math.Pow(dStar/0.055, 2.0))
			goldWire = math.Max(
				math.Exp(-math.Pow(spoke16/0.09, 2.0)),
				math.Exp(-math.Pow(dStar/0.045, 2.0)),
			)
			goldWire = math.Max(goldWire, math.Exp(-math.Pow((1.0-uCup)/0.06, 2.0)))
			return zElev, common.Clamp01(ribRidge), common.Clamp01(goldWire), uCup, 0, int(math.Floor((theta+math.Pi*4.0)/(math.Pi/8.0))) & 1
		}

		// Concentric Corbelled Muqarnas Tiers for r >= 0.13
		// Use non-linear logarithmic-perspective radial compression so inner tiers are delicate and outer tiers are bold!
		uRad := math.Pow((r-0.13)/0.92, 0.86)
		const numTiers = 11.0
		tierCoord := uRad * numTiers
		tFl := math.Floor(tierCoord)
		tierIdx = int(tFl) + 1
		vTier := tierCoord - tFl // [0 (inner/upper rim of niche) .. 1 (outer/lower corbel lip of niche)]

		// Number of stalactite niches around the ring increases in harmonic multiples of 8:
		// Tiers 1..3: 16 cells, Tiers 4..6: 24 cells, Tiers 7..9: 32 cells, Tiers 10+: 40 cells
		var nCells float64
		switch {
		case tierIdx <= 3:
			nCells = 16.0
		case tierIdx <= 6:
			nCells = 24.0
		case tierIdx <= 9:
			nCells = 32.0
		default:
			nCells = 40.0
		}

		// Alternate half-cell angular offset on odd tiers (true corbelled honeycomb staggering!)
		angShift := 0.0
		if tierIdx%2 == 1 {
			angShift = 0.5
		}
		cellCoord := (theta/(2.0*math.Pi))*nCells + angShift + 100.0
		cFl := math.Floor(cellCoord)
		cellIdx := int(cFl)
		uCell := (cellCoord - cFl)*2.0 - 1.0 // [-1, 1] across the width of a single Muqarnas niche!
		absU := math.Abs(uCell)
		cellParity = (cellIdx + tierIdx) & 1

		// Pointed Islamic Arch (Muqarnas Conch / Squat Arch) boundary inside each cell:
		// At vTier = 0 (top of niche), the pointed arch apex meets at uCell = 0;
		// As vTier increases to 1 (base of niche), the arch widens to |uCell| = 1, forming a hanging
		// stalactite pendant column (tas) at |uCell| ~ 1!
		archCurve := math.Pow(vTier, 0.68)
		dArchWall := absU - archCurve // < 0 inside the concave pointed-arch niche, > 0 on the hanging stalactite spandrel!

		// Global dome step elevation (inner cupola is highest in 3D relief, stepping down toward outer squinches,
		// while each overhanging stalactite pendant juts sharply upward toward the viewer!)
		baseTierZ := (25.0 - float64(tierIdx)*1.55) * scale

		var localZ float64
		if dArchWall <= 0.0 {
			// INSIDE THE CONCAVE POINTED-ARCH MUQARNAS NICHE (Conch / Bovedilla):
			//Normalized width inside the arch at height vTier: wNorm in [0 (center spine) .. 1 (arch wall)]
			wNorm := absU / math.Max(0.08, archCurve)
			// Concave vaulted ceiling that curves deeper into the dome toward the back of the niche,
			// sculpted with 5 radiating scallop fan-ribs (fluted conch shell vaulting)!
			fanAngle := math.Atan2(vTier, uCell)
			fanRibs := 0.5 + 0.5*math.Cos(7.0*fanAngle)
			concaveDepth := (1.0 - wNorm*wNorm) * (1.0 - 0.35*vTier)
			localZ = (-4.2*concaveDepth + 0.85*fanRibs*(1.0-wNorm)) * scale

			// Central Spine Ridge (uCell == 0) & Pointed-Arch Vault Ribs
			spineProx := math.Exp(-math.Pow(absU/0.065, 2.0))
			archWallProx := math.Exp(-math.Pow(dArchWall/0.065, 2.0))
			fanRidgeProx := math.Pow(fanRibs, 6.0) * (1.0 - wNorm*0.5)

			ribRidge = math.Max(spineProx, math.Max(archWallProx, 0.55*fanRidgeProx))
			goldWire = math.Max(
				math.Exp(-math.Pow(dArchWall/0.042, 2.0)),
				0.75*math.Exp(-math.Pow(absU/0.032, 2.0)),
			)
			tilePattern = 0.5 + 0.5*math.Cos(5.0*wNorm*math.Pi+vTier*math.Pi*2.0)
		} else {
			// ON THE OVERHANGING STALACTITE PENDANT SPANDREL (between two adjacent pointed arches):
			// Forms a sharp 3D prism ridge at |uCell| = 1.0 that juts out in 3D relief and terminates
			// in a carved stalactite drop (mocárabe) at vTier -> 0!
			spandrelW := (1.0 - absU) / math.Max(0.08, 1.0-archCurve) // 0 at pendant ridge |u|=1, 1 at arch wall
			// Stalactite prism roof peaks along |u| = 1 (spandrelW = 0)
			pendantRoof := (1.0 - spandrelW) * (1.0 - 0.25*vTier)
			// Carved geometric muqarnas step facets along the stalactite drop
			stepFacet := 0.5 + 0.5*math.Cos(vTier*math.Pi*3.0)
			localZ = (3.6*pendantRoof + 0.9*stepFacet) * scale

			pendantCrest := math.Exp(-math.Pow((1.0-absU)/0.060, 2.0))
			archWallProx := math.Exp(-math.Pow(dArchWall/0.065, 2.0))
			ribRidge = math.Max(pendantCrest, archWallProx)
			goldWire = math.Max(
				math.Exp(-math.Pow((1.0-absU)/0.038, 2.0)),
				math.Exp(-math.Pow(dArchWall/0.042, 2.0)),
			)
			tilePattern = stepFacet
		}

		// Corbelled tier step lip at vTier == 0 and vTier == 1
		tierLip := math.Exp(-math.Pow(vTier/0.055, 2.0)) + math.Exp(-math.Pow((1.0-vTier)/0.055, 2.0))
		localZ += 1.35 * scale * tierLip
		ribRidge = math.Max(ribRidge, tierLip)
		goldWire = math.Max(goldWire, math.Exp(-math.Pow(vTier/0.032, 2.0))+math.Exp(-math.Pow((1.0-vTier)/0.032, 2.0)))

		return baseTierZ + localZ, common.Clamp01(ribRidge), common.Clamp01(goldWire), tilePattern, tierIdx, cellParity
	}

	// 2. Populate Full-Canvas 3D G-Buffer Heightfield Z[w*h] & Auxiliary Channels
	gb := render.NewFiberGBuffer(w, h, color.RGBA{R: 8, G: 11, B: 22, A: 255})
	ribBuf := make([]float64, w*h)
	wireBuf := make([]float64, w*h)
	patBuf := make([]float64, w*h)
	tierBuf := make([]uint8, w*h)
	parBuf := make([]uint8, w*h)

	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		rowOff := y * w
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			idx := rowOff + x

			z, rib, wire, pat, tIdx, cPar := evalMuqarnasVault(fx, fy)
			gb.Z[idx] = z
			ribBuf[idx] = rib
			wireBuf[idx] = wire
			patBuf[idx] = pat
			tierBuf[idx] = uint8(common.ConstrainInt(tIdx, 0, 32))
			parBuf[idx] = uint8(cPar)
		}
	}

	// 3. Multi-Directional Ray-Marched Horizon Soft Shadows (Key Sun + Opposite Fill Sun)
	keyLight := [3]float64{-0.50, -0.56, 0.66}
	fillLight := [3]float64{0.54, 0.44, 0.72}
	horizonRad := common.MaxInt(10, int(math.Round(22.0*scale)))
	keyVis := gb.ComputeHorizonShadows(keyLight, horizonRad, 1.35*scale)
	fillVis := gb.ComputeHorizonShadows(fillLight, common.MaxInt(8, horizonRad/2), 1.75*scale)

	normVec := func(v [3]float64) (float64, float64, float64) {
		l := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
		if l < 1e-9 {
			return 0, 0, 1
		}
		return v[0] / l, v[1] / l, v[2] / l
	}
	klx, kly, klz := normVec(keyLight)
	khx, khy, khz := normVec([3]float64{klx, kly, klz + 1.0})
	flx, fly, flz := normVec(fillLight)
	fhx, fhy, fhz := normVec([3]float64{flx, fly, flz + 1.0})

	r1 := common.MaxInt(1, int(math.Round(2.0*scale)))
	r2 := common.MaxInt(2, int(math.Round(5.0*scale)))
	r3 := common.MaxInt(4, int(math.Round(9.0*scale)))
	aoOffsets := [12][2]int{
		{-r1, 0}, {r1, 0}, {0, -r1}, {0, r1},
		{-r2, -r1}, {r2, -r1}, {-r2, r1}, {r2, r1},
		{-r3, -r2}, {r3, -r2}, {-r2, r3}, {r2, r3},
	}

	// Curated Alhambra & Isfahan Haft-Rangi Ceramic, Alabaster & 24K Gold Palette in Linear RGB
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

	alabasterIvory := linRGB(0xfc, 0xf4, 0xe4)  // Carved Stucco & Alabaster Ivory
	warmSandstone := linRGB(0xe4, 0xb8, 0x78)   // Alhambra Warm Honey Limestone
	gold24k := linRGB(0xf6, 0xbc, 0x3c)         // 24K Muqarnas Gold Leaf
	giltSpec := linRGB(0xff, 0xf0, 0xb4)        // Specular Burnished Gold
	isfahanLapis := linRGB(0x10, 0x2c, 0x74)    // Deep Isfahan Cobalt Lapis Glaze
	persianTurquoise := linRGB(0x1e, 0xb4, 0xba)// Persian Turquoise Ceramic Glaze
	cinnabarTerracotta := linRGB(0xd8, 0x3c, 0x28) // Nasrid Cinnabar Vermilion
	amberSaffron := linRGB(0xf2, 0x92, 0x24)    // Saffron Clerestory Glow
	vaultAbyss := linRGB(0x06, 0x0b, 0x1a)      // Deep Recessed Vault Shadow

	invScale := 1.0 / math.Max(0.25, scale)

	// 4. Deferred 3D Normal, Horizon Shadow, 12-Tap Vault AO & Ceramic/Gold-Leaf Shader
	for y := 0; y < h; y++ {
		ym1 := common.MaxInt(0, y-1)
		yp1 := common.MinInt(h-1, y+1)
		rowOff := y * w
		for x := 0; x < w; x++ {
			xm1 := common.MaxInt(0, x-1)
			xp1 := common.MinInt(w-1, x+1)
			idx := rowOff + x
			z0 := gb.Z[idx]

			dzdx := (gb.Z[y*w+xp1] - gb.Z[y*w+xm1]) * 0.5 * invScale
			dzdy := (gb.Z[yp1*w+x] - gb.Z[ym1*w+x]) * 0.5 * invScale
			nx := -dzdx * 1.45
			ny := -dzdy * 1.45
			nz := 1.0
			nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
			nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

			// 12-tap Ambient Occlusion inside the recessed Muqarnas honeycombs
			occ := 0.0
			ridgeLift := 0.0
			for _, off := range aoOffsets {
				sx := common.ConstrainInt(x+off[0], 0, w-1)
				sy := common.ConstrainInt(y+off[1], 0, h-1)
				dz := (gb.Z[sy*w+sx] - z0) * invScale
				if dz > 0.10 {
					occ += common.Constrain(dz/3.8, 0.0, 1.0)
				} else if dz < -0.10 {
					ridgeLift += common.Constrain(-dz/3.2, 0.0, 1.0)
				}
			}
			cavityAO := common.Constrain(1.0-(occ/12.0)*0.84, 0.12, 1.0)

			kShadow := 1.0
			if keyVis != nil {
				kShadow = 0.16 + 0.84*keyVis[idx]
			}
			fShadow := 1.0
			if fillVis != nil {
				fShadow = 0.22 + 0.78*fillVis[idx]
			}

			tIdx := int(tierBuf[idx])
			cPar := int(parBuf[idx])
			pat := patBuf[idx]
			wire := wireBuf[idx]
			rib := ribBuf[idx]

			// Physical Airy thin-film luster glaze on the ceramic tiles
			localFilm := lws.filmNm + float64(tIdx*38+cPar*55) - 65.0*(1.0-nz)
			tfR, tfG, tfB := render.SampleThinFilmLinearRGB(nz, localFilm, 1.56)
			irid := [3]float64{common.Clamp01(tfR), common.Clamp01(tfG), common.Clamp01(tfB)}

			// Select Alhambra & Isfahan Ceramic / Stucco Material by Corbelled Tier!
			var baseCol [3]float64
			isBacklitWindow := false
			switch {
			case tIdx == 0:
				// Central 16-Pointed Shamsa Sunburst Cupola: 24K Gold Leaf, Cinnabar & Alabaster!
				if cPar == 0 {
					baseCol = lerp3(cinnabarTerracotta, gold24k, 0.45+0.55*pat)
				} else {
					baseCol = lerp3(warmSandstone, alabasterIvory, 0.40+0.60*pat)
				}
			case tIdx == 3 || tIdx == 7:
				// Backlit Clerestory Stained-Glass & Saffron-Amber Muqarnas Drum Tiers!
				if cPar == 0 {
					baseCol = lerp3(cinnabarTerracotta, amberSaffron, 0.35+0.65*pat)
					isBacklitWindow = true
				} else {
					baseCol = lerp3(isfahanLapis, persianTurquoise, 0.40+0.60*pat)
				}
			case tIdx%3 == 1:
				// Carved Alabaster Stucco & Persian Turquoise Honeycombs
				if cPar == 0 {
					baseCol = lerp3(warmSandstone, alabasterIvory, 0.35+0.65*pat)
				} else {
					baseCol = lerp3(isfahanLapis, persianTurquoise, 0.55+0.45*pat)
				}
			case tIdx%3 == 2:
				// Deep Isfahan Cobalt Lapis & 24K Gold-Leaf Arabesque Niches
				if cPar == 0 {
					baseCol = lerp3(vaultAbyss, isfahanLapis, 0.35+0.65*pat)
				} else {
					baseCol = lerp3(warmSandstone, alabasterIvory, 0.30+0.70*pat)
				}
			default:
				// Cinnabar Vermilion, Turquoise & Cobalt Squinch Tiers
				if cPar == 0 {
					baseCol = lerp3(isfahanLapis, persianTurquoise, 0.45+0.55*pat)
				} else {
					baseCol = lerp3(isfahanLapis, cinnabarTerracotta, 0.35+0.65*pat)
				}
			}

			// Blend subtle ceramic luster iridescence at grazing vault angles
			baseCol = lerp3(baseCol, irid, 0.18*(1.0-nz))

			// Directional key + fill illumination, vault ambient occlusion & central oculus sunlight spotlight!
			dxC := (float64(x) + 0.5 - cx0) / domeScale
			dyC := (float64(y) + 0.5 - cy0) / domeScale
			rDome := math.Hypot(dxC, dyC)
			oculusSun := 0.24 + 1.08*math.Exp(-rDome*rDome*1.25)

			ndotlKey := math.Max(0.0, nx*klx+ny*kly+nz*klz)
			ndotlFill := math.Max(0.0, nx*flx+ny*fly+nz*flz)
			keyDiff := (0.14 + 0.86*ndotlKey) * kShadow
			fillDiff := (0.08 + 0.62*ndotlFill) * fShadow
			illum := (0.72*keyDiff + 0.28*fillDiff + 0.14*nz) * cavityAO * oculusSun
			if isBacklitWindow {
				// Self-luminous stained-glass clerestory glow!
				illum = math.Max(illum, 0.78*cavityAO*oculusSun)
			}

			col := [3]float64{baseCol[0] * illum, baseCol[1] * illum, baseCol[2] * illum}

			// 24K Gold-Leaf Muqarnas Ribs & Stalactite Pendant Ridges
			ndothKey := math.Max(0.0, nx*khx+ny*khy+nz*khz)
			ndothFill := math.Max(0.0, nx*fhx+ny*fhy+nz*fhz)
			specTile := (math.Pow(ndothKey, 48.0)*0.52*kShadow + math.Pow(ndothFill, 32.0)*0.22*fShadow) * oculusSun
			ribGlint := rib * (0.28 + 0.92*math.Pow(ndothKey, 14.0)*kShadow + 0.38*math.Pow(ndothFill, 10.0)*fShadow) *
				(0.45 + 0.55*(ridgeLift/12.0)) * (0.35 + 0.65*oculusSun)

			if wire > 0.05 {
				gCol := lerp3(gold24k, giltSpec, common.Clamp01(0.35+0.65*ndotlKey*kShadow+0.55*ribGlint))
				gCol[0] *= 0.38 + 0.68*oculusSun
				gCol[1] *= 0.38 + 0.68*oculusSun
				gCol[2] *= 0.38 + 0.68*oculusSun
				col = lerp3(col, gCol, common.Clamp01(wire*0.92))
			}

			col[0] += (specTile + ribGlint*0.65) * giltSpec[0]
			col[1] += (specTile + ribGlint*0.65) * giltSpec[1]
			col[2] += (specTile + ribGlint*0.65) * giltSpec[2]

			rc.SetPixelLinear(x, y, col[0], col[1], col[2], 1.0)
		}
	}

	rc.ApplyLuminanceRelief(0.34, -135.0, 44.0, 0.24)
	rc.ApplyBloom(0.80, 3.5*scale, 0.14)
	rc.ApplyColorGrade(1.18, 1.18, 232.0, 42.0, 0.14)
	rc.ApplyVignette(0.22)
	c.MergeRenderCanvas(rc)
}
