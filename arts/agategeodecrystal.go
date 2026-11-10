package arts

import (
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type agateGeodeCrystal struct {
	bands        int
	druzyRings   int
	crystalScale float64
}

type spireCrystal3D struct {
	baseX, baseY float64
	cosA, sinA   float64
	length       float64
	halfWidth    float64
	zDepth       float64
	asymL, asymR float64
	crystalTyp   int // 0: Royal Amethyst, 1: Aquamarine Beryl, 2: Imperial Citrine-Gold, 3: Rose-Magenta Tourmaline
	glintPhase   float64
}

// NewAgateGeodeCrystal returns an agateGeodeCrystal engine that synthesizes a 3D dielectric
// cathedral cluster of hexagonal amethyst, aquamarine-beryl, and citrine quartz prisms
// emerging from a dark lapis & turquoise Brazilian fortification-agate geode cradle over
// a reflective black obsidian plinth.
func NewAgateGeodeCrystal(bands, druzyRings int, crystalScale float64) *agateGeodeCrystal {
	if bands <= 0 {
		bands = 20
	}
	if druzyRings <= 0 {
		druzyRings = 3
	}
	if crystalScale <= 0 {
		crystalScale = 1.0
	}
	return &agateGeodeCrystal{
		bands:        common.ConstrainInt(bands, 10, 36),
		druzyRings:   common.ConstrainInt(druzyRings, 2, 5),
		crystalScale: common.Constrain(crystalScale, 0.5, 2.0),
	}
}

// Generative renders a static frame of the 3D Amethyst & Quartz Cathedral Cluster.
func (agc *agateGeodeCrystal) Generative(c *generativeart.Canva) {
	agc.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (agc *agateGeodeCrystal) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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

	seed := c.Seed()
	if seed == 0 {
		seed = 20262701
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed + 2701)

	hubX := 0.49 * fw
	hubY := 0.75 * fh
	plinthY := 0.84 * fh

	// 1. Construct 34 Asymmetric 3D Hexagonal Quartz, Amethyst & Aquamarine Prisms in 4 Depth Tiers
	// All crystal bases are rooted deep inside the geode cradle (hubY + 15..55px), and foreground
	// basal druzy crystals + the front fortification-agate lip naturally overlap their roots!
	spires := make([]spireCrystal3D, 0, 36)

	// Tier A: 9 Monarch Cathedral Columns (asymmetric natural cluster with a dominant left-center scepter)
	monarchSpec := []struct {
		ang, lenPx, widPx, dx, dy, z float64
		typ                          int
	}{
		{-1.63, 415.0, 38.0, -14.0, 18.0, 0.58, 0}, // Dominant Royal Amethyst monarch spire
		{-1.36, 365.0, 32.0, 26.0, 22.0, 0.52, 1},  // Tall Aquamarine-Beryl right-center column
		{-1.88, 340.0, 30.0, -48.0, 24.0, 0.54, 0}, // Left-leaning Royal Amethyst spire
		{-1.12, 295.0, 26.5, 62.0, 28.0, 0.46, 2},  // Golden Citrine-Ametrine right spire
		{-2.14, 280.0, 25.5, -82.0, 30.0, 0.48, 3}, // Crimson-Rose Tourmaline/Amethyst left spire
		{-1.50, 270.0, 28.0, 6.0, 32.0, 0.68, 0},   // Mid-front Amethyst scepter
		{-0.88, 235.0, 22.0, 96.0, 34.0, 0.42, 1},  // Far-right Aquamarine prism
		{-2.38, 225.0, 21.5, -112.0, 36.0, 0.40, 2}, // Far-left Citrine prism
		{-1.74, 245.0, 24.5, -32.0, 35.0, 0.65, 1}, // Mid-front Aquamarine column
	}
	for _, m := range monarchSpec {
		spires = append(spires, spireCrystal3D{
			baseX:      hubX + m.dx*scale + (rng.Float64()-0.5)*6.0*scale,
			baseY:      hubY + m.dy*scale + (rng.Float64()-0.5)*5.0*scale,
			cosA:       math.Cos(m.ang + (rng.Float64()-0.5)*0.03),
			sinA:       math.Sin(m.ang + (rng.Float64()-0.5)*0.03),
			length:     m.lenPx * scale * agc.crystalScale,
			halfWidth:  m.widPx * scale * agc.crystalScale,
			zDepth:     m.z,
			asymL:      0.90 + 0.18*rng.Float64(),
			asymR:      0.90 + 0.18*rng.Float64(),
			crystalTyp: m.typ,
			glintPhase: rng.Float64() * 2.0 * math.Pi,
		})
	}

	// Tier B: 13 Background Deep-Violet & Sapphire-Quartz Needles (z in [0.12, 0.34])
	for i := 0; i < 13; i++ {
		u := (float64(i) + 0.5) / 13.0
		ang := -math.Pi*0.5 + (u-0.5)*2.05 + (rng.Float64()-0.5)*0.08
		bx := hubX + (u-0.5)*170.0*scale
		by := hubY + 14.0*scale + math.Abs(u-0.5)*24.0*scale
		cLen := (210.0 + 145.0*math.Cos((u-0.46)*1.9) + 35.0*rng.Float64()) * scale * agc.crystalScale
		cWid := (14.0 + 7.5*rng.Float64()) * scale * agc.crystalScale
		spires = append(spires, spireCrystal3D{
			baseX:      bx,
			baseY:      by,
			cosA:       math.Cos(ang),
			sinA:       math.Sin(ang),
			length:     cLen,
			halfWidth:  cWid,
			zDepth:     0.12 + 0.22*rng.Float64(),
			asymL:      0.88 + 0.22*rng.Float64(),
			asymR:      0.88 + 0.22*rng.Float64(),
			crystalTyp: i % 3,
			glintPhase: rng.Float64() * 2.0 * math.Pi,
		})
	}

	// Tier C: 12 Foreground Basal Druzy Points (z in [0.74, 0.94]) that cluster across the root zone
	// so no flat crystal bottoms are ever exposed!
	for i := 0; i < 12; i++ {
		u := (float64(i) + 0.5) / 12.0
		ang := -math.Pi*0.5 + (u-0.5)*2.28 + (rng.Float64()-0.5)*0.12
		bx := hubX + (u-0.5)*195.0*scale + (rng.Float64()-0.5)*12.0*scale
		by := hubY + (28.0 + 18.0*math.Sin(u*math.Pi) + 10.0*rng.Float64()) * scale
		cLen := (105.0 + 70.0*math.Sin(u*math.Pi) + 30.0*rng.Float64()) * scale * agc.crystalScale
		cWid := (18.0 + 9.0*rng.Float64()) * scale * agc.crystalScale
		spires = append(spires, spireCrystal3D{
			baseX:      bx,
			baseY:      by,
			cosA:       math.Cos(ang),
			sinA:       math.Sin(ang),
			length:     cLen,
			halfWidth:  cWid,
			zDepth:     0.74 + 0.20*rng.Float64(),
			asymL:      0.90 + 0.18*rng.Float64(),
			asymR:      0.90 + 0.18*rng.Float64(),
			crystalTyp: (i + 1) % 4,
			glintPhase: rng.Float64() * 2.0 * math.Pi,
		})
	}

	sort.Slice(spires, func(i, j int) bool {
		return spires[i].zDepth > spires[j].zDepth
	})

	// 2. Sculpted Dark-Luxury Brazilian Fortification-Agate Geode Cradle Around the Crystal Roots
	// Instead of a flat yellow bullseye pancake, the crystal cluster sits inside a fractured,
	// crescent-shaped dark lapis, peacock-turquoise, crimson-ruby & 24K gold fortification-agate
	// geode rim (y in [0.66*fh, plinthY]) whose front lip rises in front of the crystal bases!
	evalAgateCradle := func(px, py float64) (backWeight, frontWeight, uBand, bastionCrease, w1, w2 float64) {
		if py < hubY-0.12*fh || py > plinthY+4.0*scale {
			return 0, 0, 0, 0, 0, 0
		}
		ux := (px - hubX) / (0.43 * fw)
		// Crescent bowl geometry: outer boundary & inner cavity boundary
		bowlFloorY := hubY + 0.035*fh - 0.050*fh*(1.0-math.Min(1.2, ux*ux))
		dy := (py - bowlFloorY) / (0.11 * fh)

		w1 = ne.Perlin2D(ux*2.6+1.3, dy*2.2-1.7)
		w2 = ne.Perlin2D(ux*5.8-2.1, dy*4.8+0.9)

		// Polygonal fortification bastion modulation along the geode rim
		ang := math.Atan2(dy, ux)
		bastionMod := 0.045 * math.Abs(math.Sin(4.5*ang+0.5*w1))
		bastionCrease = math.Exp(-math.Pow(math.Sin(4.5*ang+0.5*w1)/0.12, 2.0))

		rBowl := math.Hypot(ux*0.92, dy*0.78) + bastionMod + 0.04*w1 + 0.015*w2
		if rBowl > 1.02 {
			return 0, 0, 0, 0, w1, w2
		}

		uBand = common.Clamp01((rBowl - 0.32) / 0.70)
		outerFade := common.Smoothstep(1.02, 0.90, uBand)
		// Smoothly transition from back shell to front lip wrapping around the crystal bases
		frontSplit := common.Smoothstep(bowlFloorY-4.0*scale, bowlFloorY+14.0*scale, py) *
			common.Smoothstep(0.32, 0.48, rBowl)
		frontWeight = outerFade * frontSplit
		backWeight = outerFade * common.Smoothstep(0.22, 0.36, rBowl) * (1.0 - frontSplit)
		return backWeight, frontWeight, uBand, bastionCrease, w1, w2
	}

	// 3. Curated High-Contrast Gemstone & Fortification-Agate Linear-RGB Palette
	velvetBg := linRGBAgate(4, 6, 15)
	nebulaViolet := linRGBAgate(34, 10, 72)
	nebulaCyan := linRGBAgate(10, 52, 86)

	amethystAbyss := linRGBAgate(16, 4, 42)
	amethystRoyal := linRGBAgate(118, 22, 214)
	amethystOrchid := linRGBAgate(228, 92, 255)

	aquaDeep := linRGBAgate(8, 38, 96)
	aquaBeryl := linRGBAgate(24, 188, 222)
	aquaIce := linRGBAgate(214, 248, 255)

	citrineDeep := linRGBAgate(96, 28, 18)
	citrineAmber := linRGBAgate(246, 148, 32)
	imperialGold := linRGBAgate(255, 214, 84)

	rubyDeep := linRGBAgate(78, 8, 36)
	rubyCrimson := linRGBAgate(218, 24, 78)
	rubyRose := linRGBAgate(255, 118, 176)

	lapisDark := linRGBAgate(10, 22, 58)
	lapisRoyal := linRGBAgate(22, 64, 148)
	diamondWhite := linRGBAgate(252, 248, 255)

	lightX := -0.48 + 0.20*math.Cos(phi)
	lightY := -0.50 + 0.16*math.Sin(phi)
	lightZ := math.Sqrt(math.Max(0.1, 1.0-lightX*lightX-lightY*lightY))

	// Pure, glassy 3D planar facet shader for a single hexagonal crystal spire
	shadeSpire := func(sp *spireCrystal3D, uC, vC, ridgeBevel, edgeDist float64, faceID int) (r, g, b, nxWorld, nyWorld float64) {
		var localNX, localNY, localNZ float64
		switch faceID {
		case 0: // Left hexagonal prism face (-60 deg transverse normal — strongly lit by upper-left key light!)
			localNX, localNY, localNZ = -0.05, -0.84, 0.54
		case 1: // Center hexagonal prism table face
			localNX, localNY, localNZ = -0.06, 0.0, 0.998
		case 2: // Right hexagonal prism face (+60 deg transverse normal — deep Beer-Lambert shadow side!)
			localNX, localNY, localNZ = -0.05, 0.84, 0.54
		case 3: // Center rhombohedral pyramid termination face
			localNX, localNY, localNZ = 0.66, 0.0, 0.75
		case 4: // Left rhombohedral pyramid termination face (catches intense key specular!)
			localNX, localNY, localNZ = 0.52, -0.62, 0.59
		default: // Right rhombohedral pyramid termination face
			localNX, localNY, localNZ = 0.52, 0.62, 0.59
		}

		nxWorld = localNX*sp.cosA - localNY*sp.sinA
		nyWorld = localNX*sp.sinA + localNY*sp.cosA
		nzWorld := localNZ

		ndotl := math.Max(0.0, nxWorld*lightX+nyWorld*lightY+nzWorld*lightZ)
		// Secondary rim light from the right-below so shadowed right facets glow with internal Beer-Lambert color
		rimBack := math.Max(0.0, 0.62*nxWorld-0.52*nyWorld+0.58*nzWorld)

		halfX := lightX * 0.5
		halfY := lightY * 0.5
		halfZ := (lightZ + 1.0) * 0.5
		hLen := math.Hypot(math.Hypot(halfX, halfY), halfZ)
		ndoth := math.Max(0.0, (nxWorld*halfX+nyWorld*halfY+nzWorld*halfZ)/hLen)
		spec := math.Pow(ndoth, 56.0)
		if faceID == 4 || faceID == 0 {
			spec *= 1.35
		}

		tLen := common.Clamp01(uC / sp.length)

		// Single delicate phantom-quartz growth chevron near the upper termination (uC ~ 0.64*L), NOT a repeating fishbone!
		tipStart := 0.70 * sp.length
		phantomDist1 := ((uC - 0.58*sp.length) + 0.58*math.Abs(vC)) / (7.5 * scale)
		phantomDist2 := ((uC - 0.44*sp.length) + 0.58*math.Abs(vC)) / (9.0 * scale)
		phantomGhost := (0.65*math.Exp(-phantomDist1*phantomDist1) + 0.35*math.Exp(-phantomDist2*phantomDist2)) *
			common.Smoothstep(0.15, 0.50, tLen) * common.Smoothstep(0.95, 0.65, tLen)
		_ = tipStart

		var deepCol, midCol, tipCol [3]float64
		switch sp.crystalTyp {
		case 0: // Royal Brazilian Amethyst
			deepCol, midCol, tipCol = amethystAbyss, amethystRoyal, amethystOrchid
		case 1: // Aquamarine Beryl & Ice Quartz
			deepCol, midCol, tipCol = aquaDeep, aquaBeryl, aquaIce
		case 2: // Imperial Citrine & Ametrine
			deepCol, midCol, tipCol = citrineDeep, citrineAmber, imperialGold
		default: // Rubellite Rose-Tourmaline & Amethyst
			deepCol, midCol, tipCol = rubyDeep, rubyCrimson, rubyRose
		}

		// Smooth Beer-Lambert axial color zoning (deep saturated root -> luminous crystalline tip)
		var body [3]float64
		if tLen < 0.52 {
			body = lerpLin3(deepCol, midCol, math.Pow(tLen/0.52, 0.9))
		} else {
			body = lerpLin3(midCol, tipCol, math.Pow((tLen-0.52)/0.48, 0.85))
		}

		// Distinct planar facet contrast between the 6 faces of the crystal!
		facetContrast := 1.0
		switch faceID {
		case 0:
			facetContrast = 1.24 // sunlit left prism face
		case 1:
			facetContrast = 0.96 // center prism table
		case 2:
			facetContrast = 0.56 // deep shadowed right prism face
		case 3:
			facetContrast = 1.18 // center pyramid termination
		case 4:
			facetContrast = 1.42 // blazing sunlit left pyramid termination
		case 5:
			facetContrast = 0.72 // shadowed right pyramid termination
		}

		fresnel := math.Pow(common.Clamp01(1.0-edgeDist), 3.2)
		illum := (0.20 + 0.86*ndotl + 0.25*math.Pow(rimBack, 2.2)) * facetContrast
		bevelGlow := (0.78*ridgeBevel + 0.38*fresnel + 0.42*phantomGhost) * (0.30 + 0.70*tLen)
		glintMod := 0.74 + 0.26*math.Sin(phi+sp.glintPhase+float64(faceID)*1.1)

		r = body[0]*illum + tipCol[0]*bevelGlow + diamondWhite[0]*spec*1.55*glintMod
		g = body[1]*illum + tipCol[1]*bevelGlow + diamondWhite[1]*spec*1.50*glintMod
		b = body[2]*illum + tipCol[2]*bevelGlow + diamondWhite[2]*spec*1.65*glintMod
		return r, g, b, nxWorld, nyWorld
	}

	evalSpireHit := func(sp *spireCrystal3D, px, py float64) (hit bool, zSurf, uC, vC, ridgeBevel, edgeDist float64, faceID int) {
		dx := px - sp.baseX
		dy := py - sp.baseY
		uC = dx*sp.cosA + dy*sp.sinA
		if uC < 0 || uC > sp.length {
			return false, 0, 0, 0, 0, 0, 0
		}
		vC = -dx*sp.sinA + dy*sp.cosA
		R := sp.halfWidth
		if math.Abs(vC) > R {
			return false, 0, 0, 0, 0, 0, 0
		}

		prismTop := 0.24
		p0 := prismTop * (R + vC) / (0.52 * R)
		p1 := prismTop
		p2 := prismTop * (R - vC) / (0.52 * R)

		tipStart := 0.72 * sp.length
		tipSpan := 0.28 * sp.length
		uTip := uC - tipStart
		p3 := prismTop * (1.0 - (uTip/tipSpan)*0.96)
		p4 := prismTop * (1.0 - (uTip/tipSpan)*sp.asymL + (vC+0.22*R)/(1.35*R))
		p5 := prismTop * (1.0 - (uTip/tipSpan)*sp.asymR - (vC-0.22*R)/(1.35*R))

		planes := [6]float64{p0, p1, p2, p3, p4, p5}
		minP := planes[0]
		secP := 1e9
		faceID = 0
		for f := 1; f < 6; f++ {
			if planes[f] < minP {
				secP = minP
				minP = planes[f]
				faceID = f
			} else if planes[f] < secP {
				secP = planes[f]
			}
		}

		if minP <= 0.0 {
			return false, 0, 0, 0, 0, 0, 0
		}

		zSurf = sp.zDepth + minP
		ridgeBevel = math.Exp(-math.Pow((secP-minP)/0.020, 2.0))
		edgeDist = common.Clamp01(minP / prismTop)
		return true, zSurf, uC, vC, ridgeBevel, edgeDist, faceID
	}

	// Shade the Dark-Luxury Lapis, Turquoise & Gold Brazilian Fortification-Agate Geode Cradle
	shadeAgateCradle := func(px, py, uBand, bastionCrease, w1, w2 float64) (r, g, b float64) {
		freq := float64(agc.bands) * 1.15
		wave := math.Sin(uBand*freq*math.Pi + 0.45*w1 - phi*0.5)
		fine := math.Sin(uBand*freq*math.Pi*3.2 - 0.7*w2 + 0.5*wave)

		// Dark-luxury mineral progression: inner amethyst/opal druzy margin -> peacock turquoise ->
		// royal lapis -> crimson carnelian seam -> dark basalt/gold outer rind
		var col [3]float64
		switch {
		case uBand < 0.20:
			col = lerpLin3(amethystOrchid, diamondWhite, uBand/0.20)
		case uBand < 0.44:
			col = lerpLin3(diamondWhite, aquaBeryl, (uBand-0.20)/0.24)
		case uBand < 0.68:
			col = lerpLin3(aquaBeryl, lapisRoyal, (uBand-0.44)/0.24)
		case uBand < 0.84:
			col = lerpLin3(lapisRoyal, rubyCrimson, (uBand-0.68)/0.16)
		default:
			col = lerpLin3(rubyCrimson, lapisDark, (uBand-0.84)/0.16)
		}

		opalSeam := math.Pow(0.5+0.5*math.Sin(uBand*freq*math.Pi*2.0+wave*1.8), 12.0)
		goldSeam := math.Pow(0.5+0.5*math.Cos(uBand*freq*math.Pi-wave*1.5), 18.0)
		mod := 0.68 + 0.38*(0.5+0.5*fine)

		r = col[0]*mod*(1.0-0.60*opalSeam) + diamondWhite[0]*0.60*opalSeam
		g = col[1]*mod*(1.0-0.60*opalSeam) + diamondWhite[1]*0.60*opalSeam
		b = col[2]*mod*(1.0-0.60*opalSeam) + diamondWhite[2]*0.60*opalSeam

		gMix := math.Min(1.0, 0.88*goldSeam+0.28*bastionCrease*opalSeam)
		r = r*(1.0-gMix) + imperialGold[0]*1.18*gMix
		g = g*(1.0-gMix) + imperialGold[1]*1.10*gMix
		b = b*(1.0-gMix) + imperialGold[2]*0.90*gMix

		// 3D sculptural relief shading across the fortification cradle
		relief := 0.55 + 0.45*math.Cos(uBand*freq*math.Pi*2.0-0.5)
		return r * relief, g * relief, b * relief
	}

	// Evaluate complete upper scene (sky, crystallographic gold astrolabe rings, 3D crystals & agate cradle) at (px, py)
	evalUpperScene := func(px, py float64) (rOut, gOut, bOut float64) {
		// 1. Midnight-obsidian background + soft radial amethyst/cyan nebula + engraved 24K gold crystallographic halo
		dxHalo := px - hubX
		dyHalo := py - 0.44*fh
		rHalo := math.Hypot(dxHalo, dyHalo) / (0.42 * minDim)
		radGlow := math.Exp(-rHalo * rHalo * 1.35)

		bgR := velvetBg[0] + nebulaViolet[0]*0.58*radGlow + nebulaCyan[0]*0.32*radGlow
		bgG := velvetBg[1] + nebulaViolet[1]*0.58*radGlow + nebulaCyan[1]*0.32*radGlow
		bgB := velvetBg[2] + nebulaViolet[2]*0.58*radGlow + nebulaCyan[2]*0.32*radGlow

		// Delicate engraved gold crystallographic stereographic rings behind the cathedral cluster
		ring1 := math.Exp(-math.Pow((rHalo-0.82)/0.0045, 2.0))
		ring2 := math.Exp(-math.Pow((rHalo-0.96)/0.0035, 2.0))
		angHalo := math.Atan2(dyHalo, dxHalo)
		ticks := math.Pow(0.5+0.5*math.Cos(angHalo*36.0+phi), 16.0) * math.Exp(-math.Pow((rHalo-0.89)/0.05, 2.0))
		astroGold := (0.28*ring1 + 0.22*ring2 + 0.18*ticks) * common.Smoothstep(0.85*fh, 0.55*fh, py)
		bgR += imperialGold[0] * astroGold
		bgG += imperialGold[1] * astroGold
		bgB += imperialGold[2] * astroGold

		backWeight, frontWeight, uBand, bastionCrease, w1, w2 := evalAgateCradle(px, py)
		var agR, agG, agB float64
		if backWeight > 0 || frontWeight > 0 {
			agR, agG, agB = shadeAgateCradle(px, py, uBand, bastionCrease, w1, w2)
		}
		if backWeight > 0 {
			bgR = bgR*(1.0-backWeight) + agR*backWeight
			bgG = bgG*(1.0-backWeight) + agG*backWeight
			bgB = bgB*(1.0-backWeight) + agB*backWeight
		}

		var lipR, lipG, lipB, lipAlpha float64
		if frontWeight > 0 {
			lipR, lipG, lipB = agR, agG, agB
			lipAlpha = frontWeight
		}

		// Find top-2 highest 3D crystal spires at (px, py)
		bestZ := -1e9
		bestIdx := -1
		var bestU, bestV, bestRidge, bestEdge float64
		bestFace := 0

		for i := range spires {
			hit, zSurf, uC, vC, ridge, edgeD, fID := evalSpireHit(&spires[i], px, py)
			if !hit {
				continue
			}
			if zSurf > bestZ {
				bestZ, bestIdx, bestU, bestV, bestRidge, bestEdge, bestFace = zSurf, i, uC, vC, ridge, edgeD, fID
			}
		}

		rScene, gScene, bScene := bgR, bgG, bgB
		if bestIdx >= 0 {
			r1, g1, b1, nxW, nyW := shadeSpire(&spires[bestIdx], bestU, bestV, bestRidge, bestEdge, bestFace)

			// True Snell lateral dielectric refraction: sample background crystal at refracted offset!
			pxRefr := px - nxW*9.5*scale
			pyRefr := py - nyW*9.5*scale
			secZ := -1e9
			secIdx := -1
			var secU, secV, secRidge, secEdge float64
			secFace := 0
			for i := range spires {
				if i == bestIdx {
					continue
				}
				hit, zSurf, uC, vC, ridge, edgeD, fID := evalSpireHit(&spires[i], pxRefr, pyRefr)
				if hit && zSurf < bestZ && zSurf > secZ {
					secZ, secIdx, secU, secV, secRidge, secEdge, secFace = zSurf, i, uC, vC, ridge, edgeD, fID
				}
			}

			rBehind, gBehind, bBehind := bgR, bgG, bgB
			if secIdx >= 0 {
				rBehind, gBehind, bBehind, _, _ = shadeSpire(&spires[secIdx], secU, secV, secRidge, secEdge, secFace)
				ao := common.Smoothstep(0.0, 0.09, bestZ-secZ)
				r1 *= 0.58 + 0.42*ao
				g1 *= 0.58 + 0.42*ao
				b1 *= 0.58 + 0.42*ao
			}

			// Root ambient occlusion where crystals emerge from the geode cradle
			rootAO := common.Smoothstep(0.0, 26.0*scale, bestU)
			r1 *= 0.45 + 0.55*rootAO
			g1 *= 0.45 + 0.55*rootAO
			b1 *= 0.45 + 0.55*rootAO

			alphaEdge := common.Smoothstep(0.0, 0.08, bestEdge)
			rC := r1*0.74 + rBehind*0.26
			gC := g1*0.74 + gBehind*0.26
			bC := b1*0.74 + bBehind*0.26

			rScene = bgR*(1.0-alphaEdge) + rC*alphaEdge
			gScene = bgG*(1.0-alphaEdge) + gC*alphaEdge
			bScene = bgB*(1.0-alphaEdge) + bC*alphaEdge
		}

		// Composite the front fortification-agate geode lip over the deepest crystal roots!
		if lipAlpha > 0 {
			rScene = rScene*(1.0-lipAlpha) + lipR*lipAlpha
			gScene = gScene*(1.0-lipAlpha) + lipG*lipAlpha
			bScene = bScene*(1.0-lipAlpha) + lipB*lipAlpha
		}
		return rScene, gScene, bScene
	}

	// 4. Parallel 2x2 SSAA Per-Pixel Synthesis (Upper Cathedral + Polished Obsidian Mirror Plinth)
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	chunk := (h + workers - 1) / workers
	subOffsets := [2]float64{0.25, 0.75}

	for wk := 0; wk < workers; wk++ {
		y0 := wk * chunk
		y1 := y0 + chunk
		if y0 >= h {
			break
		}
		if y1 > h {
			y1 = h
		}
		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for y := startY; y < endY; y++ {
				fyBase := float64(y)
				for x := 0; x < w; x++ {
					fxBase := float64(x)
					var accR, accG, accB float64

					for syIdx := 0; syIdx < 2; syIdx++ {
						py := fyBase + subOffsets[syIdx]
						for sxIdx := 0; sxIdx < 2; sxIdx++ {
							px := fxBase + subOffsets[sxIdx]

							if py <= plinthY {
								rS, gS, bS := evalUpperScene(px, py)
								accR += rS
								accG += gS
								accB += bS
							} else {
								// Polished black obsidian altar plinth reflecting the crystal cathedral & agate geode!
								pDepth := (py - plinthY) / (fh - plinthY)
								ripple := math.Sin((py-plinthY)*0.45/scale+px*0.02/scale-phi) * (1.2 + 4.5*pDepth) * scale
								reflY := plinthY - (py-plinthY)*1.35
								rRefl, gRefl, bRefl := evalUpperScene(px+ripple, math.Max(0, reflY))

								// Schlick Fresnel obsidian mirror reflectance + golden horizon bevel line
								fresnel := (0.26 + 0.34*math.Pow(1.0-pDepth, 1.4)) * common.Smoothstep(1.0, 0.05, pDepth)
								horizonLine := math.Exp(-math.Pow((py-plinthY)/(1.4*scale), 2.0)) *
									math.Exp(-math.Pow((px-hubX)/(0.38*fw), 2.0))

								accR += velvetBg[0]*0.7 + rRefl*fresnel + imperialGold[0]*0.55*horizonLine
								accG += velvetBg[1]*0.7 + gRefl*fresnel + imperialGold[1]*0.50*horizonLine
								accB += velvetBg[2]*0.7 + bRefl*fresnel + imperialGold[2]*0.40*horizonLine
							}
						}
					}

					rc.SetPixelLinear(x, y, accR*0.25, accG*0.25, accB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 5. Crisp 6-Ray Optical Star-Burst Flares on the Terminations of the Cathedral Crystal Spires
	rc.SetBlendMode(render.BlendAdd)
	flareWhite := color.RGBA{R: 255, G: 250, B: 255, A: 255}
	flareOrchid := color.RGBA{R: 228, G: 112, B: 255, A: 255}
	flareCyan := color.RGBA{R: 88, G: 238, B: 255, A: 255}

	for i := range spires {
		sp := &spires[i]
		tx := sp.baseX + sp.cosA*(0.96*sp.length)
		ty := sp.baseY + sp.sinA*(0.96*sp.length)
		if tx < 15*scale || tx > fw-15*scale || ty < 15*scale || ty > plinthY-10*scale {
			continue
		}
		glint := 0.5 + 0.5*math.Sin(phi+sp.glintPhase)
		if glint > 0.60 {
			gStr := (glint - 0.60) / 0.40
			haloCol := flareOrchid
			if sp.crystalTyp == 1 {
				haloCol = flareCyan
			}
			rc.DrawGlowDot(tx, ty, (5.2+4.2*gStr)*scale, haloCol, 0.50*gStr, 0.32)
			rc.DrawGlowDot(tx, ty, 1.9*scale, flareWhite, 0.94*gStr, 0.85)

			if sp.halfWidth > 22.0*scale && gStr > 0.42 {
				rayLen := (12.0 + 14.0*gStr) * scale
				rc.DrawTaperedSpline([]render.Point{
					{X: tx - rayLen, Y: ty},
					{X: tx, Y: ty},
					{X: tx + rayLen, Y: ty},
				}, 1.15*scale, render.TaperBothEnds, func(float64) color.RGBA { return flareWhite }, 0.62*gStr)
				rc.DrawTaperedSpline([]render.Point{
					{X: tx, Y: ty - rayLen},
					{X: tx, Y: ty},
					{X: tx, Y: ty + rayLen},
				}, 1.15*scale, render.TaperBothEnds, func(float64) color.RGBA { return flareWhite }, 0.62*gStr)
			}
		}
	}
	rc.SetBlendMode(render.BlendOver)

	rc.ApplyBloom(0.76, 3.8*scale, 0.15)
	rc.ApplyVignette(0.20)
	c.MergeRenderCanvas(rc)
}

func linRGBAgate(r, g, b uint8) [3]float64 {
	lr, lg, lb, _ := render.ColorToLinearRGB(color.RGBA{R: r, G: g, B: b, A: 255})
	return [3]float64{lr, lg, lb}
}

func lerpLin3(a, b [3]float64, t float64) [3]float64 {
	t = common.Clamp01(t)
	return [3]float64{
		a[0] + (b[0]-a[0])*t,
		a[1] + (b[1]-a[1])*t,
		a[2] + (b[2]-a[2])*t,
	}
}
