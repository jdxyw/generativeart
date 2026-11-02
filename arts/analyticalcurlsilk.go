package arts

import (
	"image/color"
	"math"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type analyticalCurlSilk struct {
	streamlines int
	octaves     int
	curlScale   float64
}

// NewAnalyticalCurlSilk returns an AnalyticalCurlSilk generator that sculpts a 3D depth-sorted
// Perlin silk and metallic ribbon wave manifold over deep obsidian space.
func NewAnalyticalCurlSilk(streamlines, octaves int, curlScale float64) *analyticalCurlSilk {
	if streamlines <= 0 {
		streamlines = 420
	}
	if octaves <= 0 {
		octaves = 4
	}
	if curlScale <= 0 {
		curlScale = 2.2
	}
	return &analyticalCurlSilk{
		streamlines: streamlines,
		octaves:     octaves,
		curlScale:   curlScale,
	}
}

type silkVec3 struct {
	x, y, z float64
}

func (a silkVec3) add(b silkVec3) silkVec3 {
	return silkVec3{a.x + b.x, a.y + b.y, a.z + b.z}
}

func (a silkVec3) sub(b silkVec3) silkVec3 {
	return silkVec3{a.x - b.x, a.y - b.y, a.z - b.z}
}

func (a silkVec3) scale(s float64) silkVec3 {
	return silkVec3{a.x * s, a.y * s, a.z * s}
}

func (a silkVec3) dot(b silkVec3) float64 {
	return a.x*b.x + a.y*b.y + a.z*b.z
}

func (a silkVec3) cross(b silkVec3) silkVec3 {
	return silkVec3{
		x: a.y*b.z - a.z*b.y,
		y: a.z*b.x - a.x*b.z,
		z: a.x*b.y - a.y*b.x,
	}
}

func (a silkVec3) normalize() silkVec3 {
	l := math.Sqrt(a.x*a.x + a.y*a.y + a.z*a.z)
	if l < 1e-9 {
		return silkVec3{0, 0, 1}
	}
	inv := 1.0 / l
	return silkVec3{a.x * inv, a.y * inv, a.z * inv}
}

type silkMicroSegment struct {
	p0, p1       render.Point
	pExt0, pExt1 render.Point
	zDepth       float64
	normal3D     silkVec3
	tangent3D    silkVec3
	strandIdx    int
	bundleID     int
	uParam       float64
	width        float64
}

// Generative renders the static AnalyticalCurlSilk 3D sculpture.
func (acs *analyticalCurlSilk) Generative(c *generativeart.Canva) {
	acs.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping AnalyticalCurlSilk 3D sculpture.
func (acs *analyticalCurlSilk) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20261903
	}
	ne := common.NewNoiseEngine(seed)

	phi := 2.0 * math.Pi * phase
	loopX := 0.16 * math.Cos(phi)
	loopY := 0.16 * math.Sin(phi)
	cScale := common.Constrain(acs.curlScale, 1.4, 3.2)

	// 1. Background & Negative Space:
	// Deep, velvety midnight-obsidian cosmos gradient (#060913 at corners to #0f1a34 at center)
	// so the glowing 3D ribbons float in dramatic negative space (~35% dark background visible).
	cornerObsidian := color.RGBA{R: 6, G: 9, B: 19, A: 255}   // #060913
	centerObsidian := color.RGBA{R: 15, G: 26, B: 52, A: 255} // #0f1a34
	deepNebulaTint := color.RGBA{R: 20, G: 34, B: 68, A: 255} // subtle central cosmos glow

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

			// Subtle large-scale cosmic silk nebula modulation in the deep background
			neb := 0.5 + 0.5*ne.Perlin3D(fx/minDim*1.6+loopX, fy/minDim*1.6+loopY, 0.7)
			bgCol := render.LerpOklab(cornerObsidian, centerObsidian, math.Pow(centerWeight, 0.85))
			bgCol = render.LerpOklab(bgCol, deepNebulaTint, centerWeight*0.22*neb)

			rLin, gLin, bLin, _ := render.ColorToLinearRGB(bgCol)
			rc.SetPixelLinear(x, y, rLin, gLin, bLin, 1.0)
		}
	}

	// 2. 3D Perlin-Warped Ribbon Manifold:
	// 3 interwoven, sweeping 3D ribbon wave bundles, each with 32 parallel cords (96 cords total)
	// and 220 sample steps along u in [0, 1].
	const (
		numBundles = 3
		numStrands = 32
		numSteps   = 220
		numPts     = numSteps + 1
	)

	// Evaluate base 3D spine position, half-width, and twist angle for bundle b at parameter u in [0, 1].
	evalBundleSpine := func(bundleID int, u float64) (spine silkVec3, halfWidth, twist, env float64) {
		switch bundleID {
		case 0:
			// Bundle A (Primary Golden-Coral S-Wave):
			// Sweeps diagonally from bottom-left to top-right in a grand 3D folding S-curve ribbon sheet.
			s := (u - 0.5) * 2.0
			arg := 2.0*math.Pi*u - 0.30 + phi
			bx := 0.42*s + 0.13*math.Sin(arg) + 0.04*math.Cos(2.0*arg)
			by := -0.37*s + 0.16*math.Sin(arg+1.15) + 0.05*math.Sin(2.0*arg-0.4)
			bz := 0.52*math.Sin(2.0*math.Pi*u+0.22+phi) + 0.24*math.Cos(4.0*math.Pi*u-0.35)

			taper := math.Sin(math.Pi * u)
			env = 0.28 + 0.72*math.Pow(taper, 0.65)
			halfWidth = 0.106 * env * (0.88 + 0.12*math.Cos(2.0*math.Pi*u))
			twist = 1.42*math.Sin(2.0*math.Pi*u+0.35+phi) + 0.65*s
			return silkVec3{bx, by, bz}, halfWidth, twist, env

		case 1:
			// Bundle B (Secondary Turquoise-Sapphire Counter-Wave):
			// Sweeps from top-left to bottom-right, weaving OVER and UNDER Bundle A in true 3D depth Z.
			s := (u - 0.5) * 2.0
			arg := 2.0*math.Pi*u + 0.40 - phi
			bx := 0.42*s - 0.13*math.Sin(arg) + 0.04*math.Sin(2.0*arg)
			by := 0.36*s + 0.15*math.Sin(arg-1.05) - 0.05*math.Cos(2.0*arg+0.3)
			// Phase-inverted Z relative to Bundle A so the two sheets weave over and under each other
			bz := -0.54*math.Sin(2.0*math.Pi*u+0.22+phi) + 0.26*math.Sin(4.0*math.Pi*u+0.65)

			taper := math.Sin(math.Pi * u)
			env = 0.28 + 0.72*math.Pow(taper, 0.65)
			halfWidth = 0.098 * env * (0.88 + 0.12*math.Sin(2.0*math.Pi*u))
			twist = -1.38*math.Cos(2.0*math.Pi*u-0.25-phi) - 0.60*s
			return silkVec3{bx, by, bz}, halfWidth, twist, env

		default:
			// Bundle C (Central Orbital Vortex Ring):
			// A graceful 3D tilted torus/helix ribbon loop in the center weaving through Bundles A & B.
			theta := 2.0 * math.Pi * u
			ringR := 0.235 + 0.044*math.Cos(3.0*theta-phi) + 0.018*math.Sin(2.0*theta+phi)
			rawX := ringR * math.Cos(theta)
			rawY := ringR * math.Sin(theta) * 0.72
			// Rotate ring orientation in 2D plane for dynamic diagonal harmony
			rot := -0.26
			cosR, sinR := math.Cos(rot), math.Sin(rot)
			bx := rawX*cosR - rawY*sinR
			by := rawX*sinR + rawY*cosR
			bz := 0.56*math.Sin(theta+0.45) + 0.28*math.Sin(3.0*theta-phi)

			env = 1.0
			halfWidth = 0.070 * (0.82 + 0.18*math.Cos(3.0*theta-phi))
			twist = 1.0*theta + 0.48*math.Sin(3.0*theta-phi)
			return silkVec3{bx, by, bz}, halfWidth, twist, env
		}
	}

	// Evaluate 3D Perlin-warped point (X(u,v), Y(u,v), Z(u,v)) on the ribbon manifold
	evalManifoldPoint := func(bundleID int, u, v float64) silkVec3 {
		spine, halfWidth, twist, env := evalBundleSpine(bundleID, u)

		// Numerical tangent along spine
		const epsU = 1e-3
		var uPrev, uNext float64
		if bundleID == 2 {
			uPrev = math.Mod(u-epsU+1.0, 1.0)
			uNext = math.Mod(u+epsU, 1.0)
		} else {
			uPrev = math.Max(0.0, u-epsU)
			uNext = math.Min(1.0, u+epsU)
		}
		spPrev, _, _, _ := evalBundleSpine(bundleID, uPrev)
		spNext, _, _, _ := evalBundleSpine(bundleID, uNext)
		tang := spNext.sub(spPrev).normalize()

		// Construct orthonormal ribbon frame (tang, latAxis, normAxis) twisted by twist angle
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

		// Displace 3D coordinates smoothly using ne.FBM2DDeriv(u*curlScale, v*0.8, 3, 2.0, 0.45) and ne.Perlin3D
		bOffU := float64(bundleID)*3.7 + loopX
		bOffV := float64(bundleID)*2.3 + loopY

		var fbmVal, dFbmDu, dFbmDv float64
		if bundleID == 2 {
			// Seamless periodic blending across u in [0, 1] for the closed central orbital vortex ring
			v1, du1, dv1 := ne.FBM2DDeriv(u*cScale+bOffU, v*0.8+bOffV, 3, 2.0, 0.45)
			v2, du2, dv2 := ne.FBM2DDeriv((u-1.0)*cScale+bOffU, v*0.8+bOffV, 3, 2.0, 0.45)
			wBlend := common.Smoothstep(0.0, 1.0, u)
			fbmVal = (1.0-wBlend)*v1 + wBlend*v2
			dFbmDu = (1.0-wBlend)*du1 + wBlend*du2
			dFbmDv = (1.0-wBlend)*dv1 + wBlend*dv2
		} else {
			fbmVal, dFbmDu, dFbmDv = ne.FBM2DDeriv(u*cScale+bOffU, v*0.8+bOffV, 3, 2.0, 0.45)
		}

		p3 := ne.Perlin3D(
			spine.x*2.4+v*0.45+loopX,
			spine.y*2.4+v*0.45+loopY,
			spine.z*2.4+float64(bundleID)*1.9,
		)

		// Coherent fluted silk pleats across lateral strand coordinate v
		pleatPhase := v*math.Pi*3.5 + u*math.Pi*4.0 + 1.8*fbmVal - phi
		pleat := math.Sin(pleatPhase)

		latSpread := v * halfWidth * (1.0 + 0.16*fbmVal + 0.08*p3)
		sheetWave := (0.044*fbmVal + 0.034*p3 + 0.011*pleat) * env
		curlLat := -0.007 * dFbmDu * env
		curlNorm := 0.009 * dFbmDv * env

		pos := spine.
			add(latAxis.scale(latSpread + curlLat)).
			add(normAxis.scale(sheetWave + curlNorm))
		return pos
	}

	// Precompute 3D manifold grid for all 3 bundles: [bundleID][strandIdx][stepIdx]
	var grid3D [numBundles][numStrands][numPts]silkVec3
	var screen2D [numBundles][numStrands][numPts]render.Point
	var perspScale [numBundles][numStrands][numPts]float64

	const camDist = 2.65
	for b := 0; b < numBundles; b++ {
		for j := 0; j < numStrands; j++ {
			v := -1.0 + 2.0*float64(j)/float64(numStrands-1)
			for i := 0; i < numPts; i++ {
				u := float64(i) / float64(numSteps)
				if b == 2 && i == numSteps {
					// Exact closure for Bundle C orbital loop
					grid3D[b][j][i] = grid3D[b][j][0]
					screen2D[b][j][i] = screen2D[b][j][0]
					perspScale[b][j][i] = perspScale[b][j][0]
					continue
				}
				p := evalManifoldPoint(b, u, v)
				grid3D[b][j][i] = p

				denom := math.Max(0.85, camDist-p.z*0.58)
				persp := camDist / denom
				perspScale[b][j][i] = persp
				sx := cx + p.x*minDim*persp
				sy := cy + p.y*minDim*persp
				screen2D[b][j][i] = render.Pt(sx, sy)
			}
		}
	}

	// 3. Build all 96 * 220 = 21,120 3D micro-segments with analytical/manifold normals
	segments := make([]silkMicroSegment, 0, numBundles*numStrands*numSteps)

	for b := 0; b < numBundles; b++ {
		baseCordWidth := 1.85 * scale
		if b == 2 {
			baseCordWidth = 1.65 * scale
		}
		for j := 0; j < numStrands; j++ {
			v := -1.0 + 2.0*float64(j)/float64(numStrands-1)
			// Outer selvedge strands and every 4th fluted rib are slightly bolder
			ribBoost := 1.0 + 0.14*math.Cos(float64(j)*math.Pi*0.5)
			if j == 0 || j == numStrands-1 {
				ribBoost = 1.32
			}

			jPrev := j - 1
			if jPrev < 0 {
				jPrev = 0
			}
			jNext := j + 1
			if jNext >= numStrands {
				jNext = numStrands - 1
			}

			for i := 0; i < numSteps; i++ {
				p3d0 := grid3D[b][j][i]
				p3d1 := grid3D[b][j][i+1]

				// 3D tangent along cord
				dU := p3d1.sub(p3d0)
				tang3D := dU.normalize()

				// 3D lateral vector across adjacent strands on the ribbon surface
				lat0 := grid3D[b][jNext][i].sub(grid3D[b][jPrev][i])
				lat1 := grid3D[b][jNext][i+1].sub(grid3D[b][jPrev][i+1])
				dV := lat0.add(lat1).scale(0.5)
				lat3D := dV.normalize()

				// Sheet normal from cross product of tangent and lateral strand vector
				sheetNorm := tang3D.cross(lat3D).normalize()
				if sheetNorm.z < 0 {
					sheetNorm = sheetNorm.scale(-1.0)
				}

				// Add cylindrical fluting across each individual silk/metallic cord
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

				// Extend segment endpoints slightly so overlapping TaperBothEnds micro-segments
				// fuse into a continuous, silky-smooth tube without gaps or polygon chords.
				ext := 0.58
				pExt0 := render.Pt(pt0.X-dx*ext, pt0.Y-dy*ext)
				pExt1 := render.Pt(pt1.X+dx*ext, pt1.Y+dy*ext)
				if b != 2 {
					if i == 0 {
						pExt0 = pt0
					}
					if i == numSteps-1 {
						pExt1 = pt1
					}
				}

				uMid := (float64(i) + 0.5) / float64(numSteps)
				endTaper := 1.0
				if b != 2 {
					endTaper = 0.42 + 0.58*math.Sin(math.Pi*uMid)
				}
				persp := 0.5 * (perspScale[b][j][i] + perspScale[b][j][i+1])
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

	// Sort ALL micro-segments strictly back-to-front by zDepth (True 3D Painter's Algorithm)
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].zDepth < segments[j].zDepth
	})

	// 4. Palette & 3D Analytical Normal Lighting Setup
	// Bundle A: burnished 24k gold (#f8c246), fiery coral-vermilion (#e84838), champagne cream (#fff6df), bronze-violet shadows
	gold24k := color.RGBA{R: 248, G: 194, B: 70, A: 255}         // #f8c246
	coralVermilion := color.RGBA{R: 232, G: 72, B: 56, A: 255}   // #e84838
	champagneCream := color.RGBA{R: 255, G: 246, B: 223, A: 255} // #fff6df
	bronzeViolet := color.RGBA{R: 46, G: 22, B: 58, A: 255}      // deep bronze-violet fold shadow

	// Bundle B: electric peacock cyan (#20d8e0), royal azure (#2468d8), platinum-white specular, abyssal indigo shadows
	peacockCyan := color.RGBA{R: 32, G: 216, B: 224, A: 255}     // #20d8e0
	royalAzure := color.RGBA{R: 36, G: 104, B: 216, A: 255}      // #2468d8
	platinumWhite := color.RGBA{R: 244, G: 250, B: 255, A: 255}  // #f4faff
	abyssalIndigo := color.RGBA{R: 14, G: 26, B: 68, A: 255}     // deep abyssal indigo fold shadow

	// Bundle C: iridescent rose-gold and turquoise-pearl
	roseGold := color.RGBA{R: 242, G: 132, B: 148, A: 255}       // iridescent rose-gold
	warmAmberGold := color.RGBA{R: 246, G: 178, B: 112, A: 255}  // warm rose-gold highlight
	turquoisePearl := color.RGBA{R: 68, G: 222, B: 212, A: 255}  // turquoise-pearl
	plumShadow := color.RGBA{R: 32, G: 20, B: 62, A: 255}        // deep plum-indigo fold shadow

	shadowCol := color.RGBA{R: 4, G: 6, B: 15, A: 255}

	// Primary key light, secondary rim light, and view direction
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

		// 4a. Soft dark occlusion shadow offset by (+3*scale, +4*scale) so foreground ribbon folds
		// cast realistic 3D shadows onto background ribbons and deeper folds.
		shPts[0] = render.Pt(seg.pExt0.X+shOffX, seg.pExt0.Y+shOffY)
		shPts[1] = render.Pt(seg.pExt1.X+shOffX, seg.pExt1.Y+shOffY)
		depthNorm := common.Constrain((seg.zDepth+0.80)/1.60, 0.0, 1.0)
		shAlpha := 0.11 + 0.14*depthNorm
		rc.DrawTaperedSpline(
			shPts[:],
			seg.width*1.95,
			render.TaperBothEnds,
			func(t float64) color.RGBA {
				return shadowCol
			},
			shAlpha,
		)

		// 4b. Compute 3D Blinn-Phong diffuse + anisotropic metallic specular + SampleThinFilmLinearRGB iridescence
		n := seg.normal3D
		tg := seg.tangent3D

		ndotl1 := math.Max(0.0, n.dot(keyLight))
		ndotl2 := math.Max(0.0, n.dot(fillLight))
		diffuse := common.Constrain(0.18+0.68*ndotl1+0.26*ndotl2, 0.0, 1.0)

		// Anisotropic silk/metallic highlight (Kajiya-Kay sin(T, H) coupled with Blinn-Phong N.H)
		tdoth := tg.dot(halfVec)
		sinTH := math.Sqrt(math.Max(0.0, 1.0-tdoth*tdoth))
		ndoth := math.Max(0.0, n.dot(halfVec))
		anisoSpec := math.Pow(sinTH, 24.0) * math.Pow(ndoth, 5.0)
		sharpSpec := math.Pow(ndoth, 36.0)
		spec := common.Constrain(0.52*anisoSpec+0.68*sharpSpec, 0.0, 1.0)

		// Thin-film interference in linear RGB from viewing angle cosTheta = n.z
		cosTheta := common.Constrain(n.z, 0.08, 0.99)
		filmNm := 360.0 + 140.0*math.Sin(seg.uParam*math.Pi*4.0+vNorm*math.Pi*2.0+float64(seg.bundleID)*1.4-phi)
		iridR, iridG, iridB := render.SampleThinFilmLinearRGB(cosTheta, filmNm, 1.50)
		iridCol := color.RGBA{
			R: uint8(common.Constrain(math.Sqrt(common.Constrain(iridR, 0.0, 1.0))*255.0, 0.0, 255.0)),
			G: uint8(common.Constrain(math.Sqrt(common.Constrain(iridG, 0.0, 1.0))*255.0, 0.0, 255.0)),
			B: uint8(common.Constrain(math.Sqrt(common.Constrain(iridB, 0.0, 1.0))*255.0, 0.0, 255.0)),
			A: 255,
		}

		// Select bundle base palette & shade with fold depth, diffuse, iridescence, and specular crest
		var baseCol, crestCol color.RGBA
		waveMix := 0.5 + 0.5*math.Sin(seg.uParam*math.Pi*5.0+vNorm*math.Pi*2.5-phi)

		switch seg.bundleID {
		case 0:
			// Bundle A: burnished 24k gold (#f8c246), fiery coral-vermilion (#e84838), champagne cream (#fff6df), bronze-violet shadows
			midTone := render.LerpOklab(coralVermilion, gold24k, 0.28+0.64*waveMix+0.08*vSym)
			midTone = render.LerpOklab(midTone, iridCol, 0.18*(1.0-cosTheta))
			foldFactor := common.Smoothstep(0.12, 0.88, diffuse*(0.45+0.55*depthNorm))
			baseCol = render.LerpOklab(bronzeViolet, midTone, foldFactor)
			crestCol = champagneCream

		case 1:
			// Bundle B: electric peacock cyan (#20d8e0), royal azure (#2468d8), platinum-white highlights, abyssal indigo shadows
			midTone := render.LerpOklab(royalAzure, peacockCyan, 0.25+0.65*waveMix+0.10*(1.0-vSym))
			midTone = render.LerpOklab(midTone, iridCol, 0.24*(1.0-cosTheta))
			foldFactor := common.Smoothstep(0.12, 0.88, diffuse*(0.45+0.55*depthNorm))
			baseCol = render.LerpOklab(abyssalIndigo, midTone, foldFactor)
			crestCol = platinumWhite

		default:
			// Bundle C: iridescent rose-gold and turquoise-pearl
			midTone := render.LerpOklab(roseGold, turquoisePearl, 0.5+0.48*math.Sin(seg.uParam*math.Pi*6.0-vNorm*math.Pi*2.0-phi))
			midTone = render.LerpOklab(midTone, warmAmberGold, 0.28*vSym)
			midTone = render.LerpOklab(midTone, iridCol, 0.30)
			foldFactor := common.Smoothstep(0.10, 0.86, diffuse*(0.48+0.52*depthNorm))
			baseCol = render.LerpOklab(plumShadow, midTone, foldFactor)
			crestCol = champagneCream
		}

		// Subtle strand-to-strand fluted sheen variation so all 32 parallel cords read clearly
		ribLight := 0.5 + 0.5*math.Cos(float64(seg.strandIdx)*math.Pi)
		litCol := render.LerpOklab(baseCol, crestCol, common.Constrain(spec*0.78+0.08*ribLight, 0.0, 0.96))

		segPts[0] = seg.pExt0
		segPts[1] = seg.pExt1
		segAlpha := 0.84 + 0.14*depthNorm

		rc.DrawTaperedSpline(
			segPts[:],
			seg.width,
			render.TaperBothEnds,
			func(t float64) color.RGBA {
				return litCol
			},
			segAlpha,
		)

		// Crisp metallic specular filament along high-curvature sunlit crests and outer selvedge rims
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
				func(t float64) color.RGBA {
					return crestCol
				},
				specAlpha,
			)
		}
	}

	// 5. Post-processing: HDR Bloom and subtle Vignette, then merge into Canva
	rc.ApplyBloom(0.80, 4.0*scale, 0.16)
	rc.ApplyVignette(0.22)
	c.MergeRenderCanvas(rc)
}
