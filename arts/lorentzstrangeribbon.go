package arts

import (
	"image/color"
	"math"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type lorentzStrangeRibbon struct {
	strands int
	rho     float64
	beta    float64
}

// NewLorentzStrangeRibbon returns a lorentzStrangeRibbon generator that sculpts a 3D depth-sorted
// Lorenz Butterfly Strange-Attractor & Figure-Eight Metallic Silk Ribbon Sculpture over deep obsidian space.
func NewLorentzStrangeRibbon(strands int, rho, beta float64) *lorentzStrangeRibbon {
	if strands <= 0 {
		strands = 26
	}
	if rho <= 0 {
		rho = 28.0
	}
	if beta <= 0 {
		beta = 2.667
	}
	return &lorentzStrangeRibbon{
		strands: common.ConstrainInt(strands, 10, 42),
		rho:     common.Constrain(rho, 20.0, 36.0),
		beta:    common.Constrain(beta, 1.5, 4.0),
	}
}

// Generative renders the static LorentzStrangeRibbon 3D sculpture.
func (lsr *lorentzStrangeRibbon) Generative(c *generativeart.Canva) {
	lsr.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping LorentzStrangeRibbon 3D sculpture.
func (lsr *lorentzStrangeRibbon) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := c.Seed()
	if seed == 0 {
		seed = 20262603
	}
	ne := common.NewNoiseEngine(seed + 2603)

	phi := 2.0 * math.Pi * phase
	loopX := 0.15 * math.Cos(phi)
	loopY := 0.15 * math.Sin(phi)

	cornerObsidian := color.RGBA{R: 6, G: 9, B: 19, A: 255}
	centerObsidian := color.RGBA{R: 15, G: 26, B: 52, A: 255}
	deepNebulaTint := color.RGBA{R: 22, G: 38, B: 74, A: 255}

	cx, cy := 0.5*fw, 0.5*fh
	maxRad := math.Hypot(0.5*fw, 0.5*fh)
	if maxRad < 1.0 {
		maxRad = 1.0
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		dy := fy - cy
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			dx := fx - cx
			rNorm := common.Constrain(math.Hypot(dx, dy)/maxRad, 0.0, 1.0)
			centerWeight := 1.0 - common.Smoothstep(0.0, 1.0, rNorm)
			neb := 0.5 + 0.5*ne.Perlin3D(fx/minDim*1.6+loopX, fy/minDim*1.6+loopY, 0.7)
			bgCol := render.LerpOklab(cornerObsidian, centerObsidian, math.Pow(centerWeight, 0.85))
			bgCol = render.LerpOklab(bgCol, deepNebulaTint, centerWeight*0.24*neb)

			rLin, gLin, bLin, _ := render.ColorToLinearRGB(bgCol)
			rc.SetPixelLinear(x, y, rLin, gLin, bLin, 1.0)
		}
	}

	// 3 Interwoven 3D Lorenz-Butterfly & Lemniscate Strange-Attractor Ribbon Bundles:
	//   Bundle 0: Primary Double-Wing Lorenz Butterfly Lemniscate (closed figure-eight weaving left & right lobes)
	//   Bundle 1: Counter-Rotating Tilted Attractor Orbit weaving OVER and UNDER Bundle 0 in 3D depth Z
	//   Bundle 2: Sweeping Diagonal Separatrix Ribbon connecting the upper-left and lower-right saddle manifolds
	const (
		numBundles = 3
		numSteps   = 220
		numPts     = numSteps + 1
	)
	numStrands := lsr.strands
	rhoNorm := lsr.rho / 28.0
	betaNorm := lsr.beta / 2.667

	evalBundleSpine := func(bundleID int, u float64) (spine silkVec3, halfWidth, twist, env float64) {
		switch bundleID {
		case 0:
			// Bundle 0: Closed 3D Lorenz Butterfly Figure-Eight (Gerono/Lissajous Attractor Manifold)
			theta := 2.0 * math.Pi * u
			rawX := 0.36 * rhoNorm * math.Sin(theta)
			rawY := 0.28 * math.Sin(2.0*theta+0.35*betaNorm) + 0.05*math.Cos(theta-phi)
			rawZ := 0.54*math.Cos(theta+0.25) + 0.26*math.Sin(3.0*theta-phi)

			rot := -0.32
			cr, sr := math.Cos(rot), math.Sin(rot)
			bx := rawX*cr - rawY*sr
			by := rawX*sr + rawY*cr
			env = 1.0
			halfWidth = 0.082 * (0.84 + 0.16*math.Cos(2.0*theta-phi))
			twist = 0.85*math.Sin(theta+phi) + 0.45*math.Cos(2.0*theta)
			return silkVec3{bx, by, rawZ}, halfWidth, twist, env

		case 1:
			// Bundle 1: Conjugate Transverse Butterfly Wing Loop (orthogonally tilted in 3D Z)
			theta := 2.0 * math.Pi * u
			rawX := 0.34 * math.Sin(theta+0.4)
			rawY := 0.27 * betaNorm * math.Sin(2.0*theta-0.4)
			rawZ := -0.56*math.Sin(theta+0.25) + 0.28*math.Cos(3.0*theta+phi)

			rot := 0.44
			cr, sr := math.Cos(rot), math.Sin(rot)
			bx := rawX*cr - rawY*sr
			by := rawX*sr + rawY*cr
			env = 1.0
			halfWidth = 0.076 * (0.84 + 0.16*math.Sin(2.0*theta+phi))
			twist = -0.82*math.Cos(theta-phi) - 0.42*math.Sin(2.0*theta)
			return silkVec3{bx, by, rawZ}, halfWidth, twist, env

		default:
			// Bundle 2: Sweeping Unstable-Manifold Separatrix Ribbon from Top-Left to Bottom-Right
			s := (u - 0.5) * 2.0
			arg := 2.0*math.Pi*u - 0.25 + phi
			bx := 0.41*s + 0.14*math.Sin(arg)
			by := 0.38*s - 0.15*math.Cos(arg-0.6)
			bz := 0.52*math.Cos(2.0*math.Pi*u+phi) - 0.24*math.Sin(4.0*math.Pi*u)

			taper := math.Sin(math.Pi * u)
			env = 0.28 + 0.72*math.Pow(taper, 0.65)
			halfWidth = 0.092 * env * (0.88 + 0.12*math.Cos(2.0*math.Pi*u))
			twist = 1.15*math.Sin(2.0*math.Pi*u+phi) + 0.55*s
			return silkVec3{bx, by, bz}, halfWidth, twist, env
		}
	}

	evalManifoldPoint := func(bundleID int, u, v float64) silkVec3 {
		spine, halfWidth, twist, env := evalBundleSpine(bundleID, u)

		const epsU = 1e-3
		var uPrev, uNext float64
		if bundleID < 2 {
			uPrev = math.Mod(u-epsU+1.0, 1.0)
			uNext = math.Mod(u+epsU, 1.0)
		} else {
			uPrev = math.Max(0.0, u-epsU)
			uNext = math.Min(1.0, u+epsU)
		}
		spPrev, _, _, _ := evalBundleSpine(bundleID, uPrev)
		spNext, _, _, _ := evalBundleSpine(bundleID, uNext)
		tang := spNext.sub(spPrev).normalize()

		upRef := silkVec3{0, 0, 1}
		inPlane := tang.cross(upRef)
		if inPlane.dot(inPlane) < 1e-6 {
			inPlane = silkVec3{1, 0, 0}
		} else {
			inPlane = inPlane.normalize()
		}
		binorm := inPlane.cross(tang).normalize()

		cosT, sinT := math.Cos(twist), math.Sin(twist)
		latAxis := inPlane.scale(cosT).add(binorm.scale(sinT)).normalize()
		normAxis := tang.cross(latAxis).normalize()

		angU := 2.0 * math.Pi * u
		p3 := ne.Perlin3D(
			math.Cos(angU)*0.9+float64(bundleID)*1.9+loopX,
			math.Sin(angU)*0.9+v*0.45+loopY,
			spine.z*2.0,
		)
		pleat := math.Sin(v*math.Pi*3.5 + 3.0*angU - phi)
		latSpread := v * halfWidth * (1.0 + 0.14*p3)
		sheetWave := (0.022*p3 + 0.009*pleat) * env

		return spine.add(latAxis.scale(latSpread)).add(normAxis.scale(sheetWave))
	}

	grid3D := make([][][]silkVec3, numBundles)
	screen2D := make([][][]render.Point, numBundles)
	pScale := make([][][]float64, numBundles)
	const camDist = 2.60

	for b := 0; b < numBundles; b++ {
		grid3D[b] = make([][]silkVec3, numStrands)
		screen2D[b] = make([][]render.Point, numStrands)
		pScale[b] = make([][]float64, numStrands)
		for j := 0; j < numStrands; j++ {
			grid3D[b][j] = make([]silkVec3, numPts)
			screen2D[b][j] = make([]render.Point, numPts)
			pScale[b][j] = make([]float64, numPts)
			v := -1.0 + 2.0*float64(j)/float64(numStrands-1)
			for i := 0; i < numPts; i++ {
				if b < 2 && i == numSteps {
					grid3D[b][j][i] = grid3D[b][j][0]
					screen2D[b][j][i] = screen2D[b][j][0]
					pScale[b][j][i] = pScale[b][j][0]
					continue
				}
				u := float64(i) / float64(numSteps)
				p := evalManifoldPoint(b, u, v)
				grid3D[b][j][i] = p

				denom := math.Max(0.85, camDist-p.z*0.58)
				persp := camDist / denom
				pScale[b][j][i] = persp
				sx := cx + p.x*minDim*persp
				sy := cy + p.y*minDim*persp
				screen2D[b][j][i] = render.Pt(sx, sy)
			}
		}
	}

	segments := make([]silkMicroSegment, 0, numBundles*numStrands*numSteps)
	baseCordWidth := 1.80 * scale

	for b := 0; b < numBundles; b++ {
		for j := 0; j < numStrands; j++ {
			v := -1.0 + 2.0*float64(j)/float64(numStrands-1)
			ribBoost := 1.0 + 0.14*math.Cos(float64(j)*math.Pi*0.5)
			if j == 0 || j == numStrands-1 {
				ribBoost = 1.34
			}
			jPrev := common.MaxInt(0, j-1)
			jNext := common.MinInt(numStrands-1, j+1)

			for i := 0; i < numSteps; i++ {
				p3d0 := grid3D[b][j][i]
				p3d1 := grid3D[b][j][i+1]

				tang3D := p3d1.sub(p3d0).normalize()
				lat0 := grid3D[b][jNext][i].sub(grid3D[b][jPrev][i])
				lat1 := grid3D[b][jNext][i+1].sub(grid3D[b][jPrev][i+1])
				lat3D := lat0.add(lat1).scale(0.5).normalize()

				sheetNorm := tang3D.cross(lat3D).normalize()
				if sheetNorm.z < 0 {
					sheetNorm = sheetNorm.scale(-1.0)
				}
				flutePhase := float64(j)*1.57 + v*math.Pi
				norm3D := sheetNorm.add(lat3D.scale(0.28 * math.Sin(flutePhase))).normalize()
				if norm3D.z < 0.05 {
					norm3D.z = 0.05
					norm3D = norm3D.normalize()
				}

				pt0 := screen2D[b][j][i]
				pt1 := screen2D[b][j][i+1]
				dx := pt1.X - pt0.X
				dy := pt1.Y - pt0.Y
				ext := 0.58
				pExt0 := render.Pt(pt0.X-dx*ext, pt0.Y-dy*ext)
				pExt1 := render.Pt(pt1.X+dx*ext, pt1.Y+dy*ext)
				if b == 2 {
					if i == 0 {
						pExt0 = pt0
					}
					if i == numSteps-1 {
						pExt1 = pt1
					}
				}

				uMid := (float64(i) + 0.5) / float64(numSteps)
				endTaper := 1.0
				if b == 2 {
					endTaper = 0.42 + 0.58*math.Sin(math.Pi*uMid)
				}
				persp := 0.5 * (pScale[b][j][i] + pScale[b][j][i+1])
				segW := baseCordWidth * ribBoost * endTaper * (0.72 + 0.34*persp)

				segments = append(segments, silkMicroSegment{
					p0:        pt0,
					p1:        pt1,
					pExt0:     pExt0,
					pExt1:     pExt1,
					zDepth:    0.5 * (p3d0.z + p3d1.z),
					normal3D:  norm3D,
					tangent3D: tang3D,
					strandIdx: j,
					bundleID:  b,
					uParam:    uMid,
					width:     segW,
				})
			}
		}
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].zDepth < segments[j].zDepth
	})

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Dunhuang
	}
	palGrad := render.NewMultiStopGradient(schema...)

	gold24k := color.RGBA{R: 248, G: 194, B: 70, A: 255}
	coralVermilion := color.RGBA{R: 232, G: 72, B: 56, A: 255}
	champagneCream := color.RGBA{R: 255, G: 246, B: 223, A: 255}
	bronzeViolet := color.RGBA{R: 46, G: 22, B: 58, A: 255}

	peacockCyan := color.RGBA{R: 32, G: 216, B: 224, A: 255}
	royalAzure := color.RGBA{R: 36, G: 104, B: 216, A: 255}
	platinumWhite := color.RGBA{R: 244, G: 250, B: 255, A: 255}
	abyssalIndigo := color.RGBA{R: 14, G: 26, B: 68, A: 255}

	roseGold := color.RGBA{R: 242, G: 132, B: 148, A: 255}
	warmAmberGold := color.RGBA{R: 246, G: 178, B: 112, A: 255}
	turquoisePearl := color.RGBA{R: 68, G: 222, B: 212, A: 255}
	plumShadow := color.RGBA{R: 32, G: 20, B: 62, A: 255}

	shadowCol := color.RGBA{R: 4, G: 6, B: 15, A: 255}

	keyLight := (silkVec3{-0.42, -0.54, 0.73}).normalize()
	fillLight := (silkVec3{0.56, 0.40, 0.72}).normalize()
	viewDir := silkVec3{0.0, 0.0, 1.0}
	halfVec := keyLight.add(viewDir).normalize()

	var segPts [2]render.Point
	var shPts [2]render.Point
	shOffX := 3.0 * scale
	shOffY := 4.0 * scale

	for idx := range segments {
		seg := &segments[idx]
		vNorm := float64(seg.strandIdx) / float64(numStrands-1)
		vSym := math.Abs(vNorm*2.0 - 1.0)

		shPts[0] = render.Pt(seg.pExt0.X+shOffX, seg.pExt0.Y+shOffY)
		shPts[1] = render.Pt(seg.pExt1.X+shOffX, seg.pExt1.Y+shOffY)
		depthNorm := common.Constrain((seg.zDepth+0.80)/1.60, 0.0, 1.0)
		shAlpha := 0.11 + 0.14*depthNorm
		rc.DrawTaperedSpline(
			shPts[:],
			seg.width*1.95,
			render.TaperBothEnds,
			func(t float64) color.RGBA { return shadowCol },
			shAlpha,
		)

		n := seg.normal3D
		tg := seg.tangent3D

		ndotl1 := math.Max(0.0, n.dot(keyLight))
		ndotl2 := math.Max(0.0, n.dot(fillLight))
		diffuse := common.Constrain(0.18+0.68*ndotl1+0.26*ndotl2, 0.0, 1.0)

		tdoth := tg.dot(halfVec)
		sinTH := math.Sqrt(math.Max(0.0, 1.0-tdoth*tdoth))
		ndoth := math.Max(0.0, n.dot(halfVec))
		anisoSpec := math.Pow(sinTH, 24.0) * math.Pow(ndoth, 5.0)
		sharpSpec := math.Pow(ndoth, 36.0)
		spec := common.Constrain(0.52*anisoSpec+0.68*sharpSpec, 0.0, 1.0)

		cosTheta := common.Constrain(n.z, 0.08, 0.99)
		filmNm := 360.0 + 140.0*math.Sin(seg.uParam*math.Pi*4.0+vNorm*math.Pi*2.0+float64(seg.bundleID)*1.4-phi)
		iridR, iridG, iridB := render.SampleThinFilmLinearRGB(cosTheta, filmNm, 1.50)
		iridCol := color.RGBA{
			R: uint8(common.Constrain(math.Sqrt(common.Constrain(iridR, 0.0, 1.0))*255.0, 0.0, 255.0)),
			G: uint8(common.Constrain(math.Sqrt(common.Constrain(iridG, 0.0, 1.0))*255.0, 0.0, 255.0)),
			B: uint8(common.Constrain(math.Sqrt(common.Constrain(iridB, 0.0, 1.0))*255.0, 0.0, 255.0)),
			A: 255,
		}

		palSample := palGrad.SampleOklab(math.Mod(seg.uParam+float64(seg.bundleID)*0.25, 1.0))
		waveMix := 0.5 + 0.5*math.Sin(seg.uParam*math.Pi*5.0+vNorm*math.Pi*2.5-phi)

		var baseCol, crestCol color.RGBA
		switch seg.bundleID {
		case 0:
			midTone := render.LerpOklab(coralVermilion, gold24k, 0.28+0.64*waveMix+0.08*vSym)
			midTone = render.LerpOklab(midTone, palSample, 0.22)
			midTone = render.LerpOklab(midTone, iridCol, 0.18*(1.0-cosTheta))
			foldFactor := common.Smoothstep(0.12, 0.88, diffuse*(0.45+0.55*depthNorm))
			baseCol = render.LerpOklab(bronzeViolet, midTone, foldFactor)
			crestCol = champagneCream
		case 1:
			midTone := render.LerpOklab(royalAzure, peacockCyan, 0.25+0.65*waveMix+0.10*(1.0-vSym))
			midTone = render.LerpOklab(midTone, palSample, 0.22)
			midTone = render.LerpOklab(midTone, iridCol, 0.24*(1.0-cosTheta))
			foldFactor := common.Smoothstep(0.12, 0.88, diffuse*(0.45+0.55*depthNorm))
			baseCol = render.LerpOklab(abyssalIndigo, midTone, foldFactor)
			crestCol = platinumWhite
		default:
			midTone := render.LerpOklab(roseGold, turquoisePearl, 0.5+0.48*math.Sin(seg.uParam*math.Pi*6.0-vNorm*math.Pi*2.0-phi))
			midTone = render.LerpOklab(midTone, warmAmberGold, 0.28*vSym)
			midTone = render.LerpOklab(midTone, palSample, 0.20)
			midTone = render.LerpOklab(midTone, iridCol, 0.30)
			foldFactor := common.Smoothstep(0.10, 0.86, diffuse*(0.48+0.52*depthNorm))
			baseCol = render.LerpOklab(plumShadow, midTone, foldFactor)
			crestCol = champagneCream
		}

		ribLight := 0.5 + 0.5*math.Cos(float64(seg.strandIdx)*math.Pi)
		litCol := render.LerpOklab(baseCol, crestCol, common.Constrain(spec*0.78+0.08*ribLight, 0.0, 0.96))

		segPts[0] = seg.pExt0
		segPts[1] = seg.pExt1
		segAlpha := 0.84 + 0.14*depthNorm

		rc.DrawTaperedSpline(
			segPts[:],
			seg.width,
			render.TaperBothEnds,
			func(t float64) color.RGBA { return litCol },
			segAlpha,
		)

		isSelvedge := seg.strandIdx == 0 || seg.strandIdx == numStrands-1
		if spec > 0.34 || (isSelvedge && diffuse > 0.58) {
			specAlpha := common.Constrain((spec-0.28)*0.75, 0.18, 0.68)
			if isSelvedge && specAlpha < 0.35 {
				specAlpha = 0.35
			}
			rc.DrawTaperedSpline(
				segPts[:],
				seg.width*0.42,
				render.TaperBothEnds,
				func(t float64) color.RGBA { return crestCol },
				specAlpha,
			)
		}
	}

	rc.ApplyBloom(0.80, 4.0*scale, 0.16)
	rc.ApplyVignette(0.22)
	c.MergeRenderCanvas(rc)
}
