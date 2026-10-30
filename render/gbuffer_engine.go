package render

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart/common"
)

// FiberGBuffer is a true 3D Z-buffered Geometry Buffer (Height Z, 3D Surface Normal,
// Linear Albedo, Tangent, and Roughness) for rendering overlapping 3D fur & impasto tubes
// with Screen-Space Ambient Occlusion (SSAO) and physically based specular shading.
type FiberGBuffer struct {
	Width     int
	Height    int
	Z         []float64
	Nx        []float64
	Ny        []float64
	Nz        []float64
	AlbR      []float64
	AlbG      []float64
	AlbB      []float64
	Roughness []float64
}

// NewFiberGBuffer allocates a 3D FiberGBuffer initialized with background color bg at Z = 0.
func NewFiberGBuffer(w, h int, bg color.RGBA) *FiberGBuffer {
	n := w * h
	g := &FiberGBuffer{
		Width:     w,
		Height:    h,
		Z:         make([]float64, n),
		Nx:        make([]float64, n),
		Ny:        make([]float64, n),
		Nz:        make([]float64, n),
		AlbR:      make([]float64, n),
		AlbG:      make([]float64, n),
		AlbB:      make([]float64, n),
		Roughness: make([]float64, n),
	}
	rLin, gLin, bLin, _ := ColorToLinearRGB(bg)
	for i := 0; i < n; i++ {
		g.Z[i] = 0.0
		g.Nz[i] = 1.0
		g.AlbR[i] = rLin
		g.AlbG[i] = gLin
		g.AlbB[i] = bLin
		g.Roughness[i] = 0.85
	}
	return g
}

// RasterizeFiber3D rasterizes a 3D cylindrical bristle/hair tube along pts with
// elevation rising from zRoot (at t=0) to zTip (at t=1), radius maxRadius,
// longitudinal color progression (rootCol -> midCol -> tipCol), and true 3D cylindrical + hemispherical-tip normals.
func (g *FiberGBuffer) RasterizeFiber3D(
	pts []Point,
	zRoot, zTip float64,
	maxRadius float64,
	rootCol, midCol, tipCol color.RGBA,
	tipStart float64,
	roughness float64,
) {
	if g == nil || len(pts) < 2 || maxRadius <= 0 {
		return
	}
	curve := CatmullRomSpline(pts, 8)
	m := len(curve)
	if m < 2 {
		return
	}

	arcLen := make([]float64, m)
	totalLen := 0.0
	for i := 1; i < m; i++ {
		totalLen += curve[i].Dist(curve[i-1])
		arcLen[i] = totalLen
	}
	if totalLen < 1.0 {
		return
	}

	for i := 0; i < m-1; i++ {
		p0 := curve[i]
		p1 := curve[i+1]
		seg := p1.Sub(p0)
		segLen := seg.Len()
		if segLen < 1e-5 {
			continue
		}
		tan := seg.Mul(1.0 / segLen)
		norm := tan.Perp()

		t0 := arcLen[i] / totalLen
		t1 := arcLen[i+1] / totalLen

		steps := int(math.Max(1, math.Ceil(segLen*2.2)))
		for s := 0; s < steps; s++ {
			u := (float64(s) + 0.5) / float64(steps)
			t := t0 + (t1-t0)*u
			center := p0.Lerp(p1, u)

			// Radius profile: smooth root emergence, full mid-body, rounded paintbrush tip
			var rProf float64
			var tipTilt float64
			if t < 0.16 {
				rProf = math.Sin((t / 0.16) * math.Pi * 0.5)
			} else if t > 0.85 {
				uTip := (t - 0.85) / 0.15
				rProf = math.Sqrt(math.Max(0, 1.0-uTip*uTip))
				tipTilt = uTip * 0.85
			} else {
				rProf = 1.0
			}
			rad := maxRadius * rProf
			if rad < 0.35 {
				continue
			}

			// 3D centerline elevation zCenter(t) rises smoothly from zRoot to zTip
			zCenter := zRoot + (zTip-zRoot)*math.Pow(t, 1.15)

			// Longitudinal pigment color along the filament
			var col color.RGBA
			if t < tipStart {
				lt := common.Smoothstep(0.0, tipStart, t)
				col = LerpOklab(rootCol, midCol, lt)
			} else {
				lt := common.Smoothstep(tipStart, 0.90, t)
				col = LerpOklab(midCol, tipCol, lt)
			}
			rLin, gLin, bLin, _ := ColorToLinearRGB(col)

			minX := common.MaxInt(0, int(math.Floor(center.X-rad-1.0)))
			maxX := common.MinInt(g.Width-1, int(math.Ceil(center.X+rad+1.0)))
			minY := common.MaxInt(0, int(math.Floor(center.Y-rad-1.0)))
			maxY := common.MinInt(g.Height-1, int(math.Ceil(center.Y+rad+1.0)))

			invRad := 1.0 / rad
			for py := minY; py <= maxY; py++ {
				dy := (float64(py) + 0.5) - center.Y
				for px := minX; px <= maxX; px++ {
					dx := (float64(px) + 0.5) - center.X
					// Signed lateral distance across the fiber cross-section
					latDist := dx*norm.X + dy*norm.Y
					lonDist := dx*tan.X + dy*tan.Y
					if math.Abs(lonDist) > 1.15 {
						continue
					}
					v := latDist * invRad
					if math.Abs(v) >= 1.12 {
						continue
					}

					// Anti-aliased edge coverage for soft sub-pixel blending
					cov := 1.0
					if math.Abs(v) > 0.78 {
						cov = common.Constrain((1.12-math.Abs(v))/0.34, 0, 1)
						cov = cov * cov * (3.0 - 2.0*cov)
					}
					// Also fade coverage at the very root (t < 0.12) so roots merge seamlessly
					if t < 0.12 {
						cov *= common.Smoothstep(0.0, 0.12, t)
					}
					if cov <= 1e-4 {
						continue
					}

					vClamped := common.Constrain(v, -0.99, 0.99)
					nzCyl := math.Sqrt(1.0 - vClamped*vClamped)
					zSurf := zCenter + rad*nzCyl

					idx := py*g.Width + px
					// Soft Z-test: overwrite/blend if zSurf is above or within soft Z tolerance of existing surface
					zDiff := zSurf - g.Z[idx]
					if zDiff < -0.65 {
						continue
					}
					zWeight := cov
					if zDiff < 0.25 {
						zWeight *= common.Smoothstep(-0.65, 0.25, zDiff)
					}
					if zWeight <= 1e-4 {
						continue
					}

					// 3D surface normal of the cylindrical bristle + hemispherical tip tilt
					nx := vClamped*norm.X + tipTilt*tan.X
					ny := vClamped*norm.Y + tipTilt*tan.Y
					nz := nzCyl
					nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
					if nLen > 1e-9 {
						nx, ny, nz = nx/nLen, ny/nLen, nz/nLen
					}

					invW := 1.0 - zWeight
					if zSurf > g.Z[idx] {
						g.Z[idx] = g.Z[idx]*invW + zSurf*zWeight
					}
					g.Nx[idx] = g.Nx[idx]*invW + nx*zWeight
					g.Ny[idx] = g.Ny[idx]*invW + ny*zWeight
					g.Nz[idx] = g.Nz[idx]*invW + nz*zWeight
					g.AlbR[idx] = g.AlbR[idx]*invW + rLin*zWeight
					g.AlbG[idx] = g.AlbG[idx]*invW + gLin*zWeight
					g.AlbB[idx] = g.AlbB[idx]*invW + bLin*zWeight
					g.Roughness[idx] = g.Roughness[idx]*invW + roughness*zWeight
				}
			}
		}
	}
}

// ResolveToCanvas performs deferred 3D shading over the FiberGBuffer with multi-scale
// Screen-Space Ambient Occlusion (SSAO), directional key + rim lighting, and glossy impasto specular highlights.
func (g *FiberGBuffer) ResolveToCanvas(rc *Canvas, lightDir [3]float64) {
	if g == nil || rc == nil {
		return
	}
	w, h := g.Width, g.Height
	lx, ly, lz := lightDir[0], lightDir[1], lightDir[2]
	lLen := math.Sqrt(lx*lx + ly*ly + lz*lz)
	if lLen < 1e-9 {
		lx, ly, lz = -0.45, -0.60, 0.66
	} else {
		lx, ly, lz = lx/lLen, ly/lLen, lz/lLen
	}

	// Half-vector H between light L and viewer V = (0, 0, 1)
	hx, hy, hz := lx, ly, lz+1.0
	hLen := math.Sqrt(hx*hx + hy*hy + hz*hz)
	hx, hy, hz = hx/hLen, hy/hLen, hz/hLen

	// 12-tap multi-scale SSAO kernel offsets
	aoOffsets := [12][2]int{
		{-2, 0}, {2, 0}, {0, -2}, {0, 2},
		{-4, -3}, {4, -3}, {-4, 3}, {4, 3},
		{-8, -5}, {8, -5}, {-5, 8}, {5, 8},
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			z0 := g.Z[idx]
			if z0 <= 1e-4 {
				rc.SetPixelLinear(x, y, g.AlbR[idx], g.AlbG[idx], g.AlbB[idx], 1.0)
				continue
			}

			nx, ny, nz := g.Nx[idx], g.Ny[idx], g.Nz[idx]
			nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
			if nLen > 1e-9 {
				nx, ny, nz = nx/nLen, ny/nLen, nz/nLen
			} else {
				nx, ny, nz = 0, 0, 1
			}

			// Compute Screen-Space Ambient Occlusion (SSAO) from surrounding Z buffer heights
			occ := 0.0
			for _, off := range aoOffsets {
				sx := common.ConstrainInt(x+off[0], 0, w-1)
				sy := common.ConstrainInt(y+off[1], 0, h-1)
				dz := g.Z[sy*w+sx] - z0
				if dz > 0.15 {
					occ += common.Constrain(dz/5.5, 0.0, 1.0)
				}
			}
			ao := common.Constrain(1.0-(occ/12.0)*0.78, 0.18, 1.0)
			// Also darken deep roots near Z = 0
			heightLift := common.Smoothstep(0.0, 4.5, z0)
			ao *= (0.45 + 0.55*heightLift)

			// Lambertian + wrap diffuse lighting
			ndotl := nx*lx + ny*ly + nz*lz
			wrapDiff := common.Constrain((ndotl+0.35)/1.35, 0.0, 1.0)

			// Blinn-Phong glossy oil-paint / silk specular highlight
			ndoth := math.Max(0.0, nx*hx+ny*hy+nz*hz)
			rough := common.Constrain(g.Roughness[idx], 0.15, 0.95)
			shininess := 8.0 + (1.0-rough)*42.0
			spec := math.Pow(ndoth, shininess) * (1.0 - rough*0.65) * 0.42 * ao

			// Cylindrical rim / velvet sheen near grazing angles
			rim := math.Pow(1.0-math.Max(0, nz), 2.2) * 0.16 * ao

			shade := (0.28 + 0.82*wrapDiff) * ao
			rOut := g.AlbR[idx]*shade + spec*1.05 + rim*g.AlbR[idx]
			gOut := g.AlbG[idx]*shade + spec*0.96 + rim*g.AlbG[idx]
			bOut := g.AlbB[idx]*shade + spec*0.85 + rim*g.AlbB[idx]

			rc.SetPixelLinear(x, y, rOut, gOut, bOut, 1.0)
		}
	}
}

// ============================================================================
// Part VII — Section 7.2: Thin-Film Optical Iridescence & Spectral Interference
// ============================================================================

// SampleThinFilmLinearRGB models physical Airy thin-film interference for an optical film
// of refractive index iorFilm (defaulting to 1.38 if <= 1.0) and thickness thicknessNm (in nm)
// viewed at cosine angle cosTheta in [0, 1].
// Evaluates 3-tap spectral sub-band integration across Red (630, 650, 670 nm),
// Green (520, 535, 550 nm), and Blue (450, 465, 480 nm) with an exponential coherence
// envelope so high-order fringes naturally desaturate toward pearlescent white.
func SampleThinFilmLinearRGB(cosTheta, thicknessNm, iorFilm float64) (rLin, gLin, bLin float64) {
	if iorFilm <= 1.0 {
		iorFilm = 1.38
	}
	if thicknessNm < 0 {
		thicknessNm = 0
	}
	ct := common.Constrain(cosTheta, 0.0, 1.0)
	sin2Theta := 1.0 - ct*ct
	cosThetaT := math.Sqrt(math.Max(0.0, 1.0-sin2Theta/(iorFilm*iorFilm)))

	// Optical path difference (in nm) between top and bottom film reflections
	opd := 2.0 * iorFilm * thicknessNm * cosThetaT

	evalBand := func(l0, l1, l2 float64) float64 {
		lambdas := [3]float64{l0, l1, l2}
		sum := 0.0
		for _, lam := range lambdas {
			order := opd / lam
			// Half-wavelength phase shift (+ pi) at air-film interface:
			// R(lambda) = 0.5 * (1 - cos(2*pi*OPD/lambda)) = sin^2(pi*OPD/lambda)
			fringe := 0.5 * (1.0 - math.Cos(2.0*math.Pi*order))
			coherence := math.Exp(-0.0008 * order * order)
			val := 0.5 + (fringe-0.5)*coherence
			sum += val
		}
		return common.Constrain(sum/3.0, 0.0, 1.0)
	}

	rLin = evalBand(630.0, 650.0, 670.0)
	gLin = evalBand(520.0, 535.0, 550.0)
	bLin = evalBand(450.0, 465.0, 480.0)
	return rLin, gLin, bLin
}

// SampleThinFilmIridescence evaluates physical thin-film interference and returns an sRGB color.RGBA.
func SampleThinFilmIridescence(cosTheta, thicknessNm, iorFilm float64) color.RGBA {
	rLin, gLin, bLin := SampleThinFilmLinearRGB(cosTheta, thicknessNm, iorFilm)
	return color.RGBA{
		R: uint8(math.Round(LinearToSRGB(rLin) * 255.0)),
		G: uint8(math.Round(LinearToSRGB(gLin) * 255.0)),
		B: uint8(math.Round(LinearToSRGB(bLin) * 255.0)),
		A: 255,
	}
}

// ============================================================================
// Part VII — Section 7.3: Heightfield Horizon-Mapped Soft Shadows & SSS
// ============================================================================

// GBufferMaterialConfig configures advanced deferred shading over a FiberGBuffer,
// including screen-space horizon soft shadows, translucent subsurface scattering (SSS),
// and thin-film structural iridescence.
type GBufferMaterialConfig struct {
	LightDir              [3]float64
	HorizonShadowStrength float64    // [0, 1], e.g. 0.65
	HorizonRadius         int        // raymarch distance in pixels, e.g. 18
	SSSStrength           float64    // translucent subsurface scattering strength, e.g. 0.35
	SSSTint               color.RGBA // subsurface scatter tint
	IridescenceStrength   float64    // [0, 1] thin-film structural color blend, e.g. 0.40
	FilmThicknessNm       float64    // base optical film thickness in nm, e.g. 380.0
	FilmIOR               float64    // thin-film index of refraction, e.g. 1.42
}

// DefaultGBufferMaterialConfig returns a balanced default material configuration for G-Buffer shading.
func DefaultGBufferMaterialConfig(lightDir [3]float64) GBufferMaterialConfig {
	return GBufferMaterialConfig{
		LightDir:              lightDir,
		HorizonShadowStrength: 0.65,
		HorizonRadius:         18,
		SSSStrength:           0.35,
		SSSTint:               color.RGBA{R: 255, G: 145, B: 95, A: 255},
		IridescenceStrength:   0.40,
		FilmThicknessNm:       380.0,
		FilmIOR:               1.42,
	}
}

// ComputeHorizonShadows computes a w*h slice of shadow visibility factors in [0, 1]
// (1.0 = fully lit, <1.0 = in horizon shadow) by marching 3 penumbra rays in the 2D direction
// of lightDir (center ray + +/-15 deg spread) up to maxRadius steps.
func (g *FiberGBuffer) ComputeHorizonShadows(lightDir [3]float64, maxRadius int, softness float64) []float64 {
	if g == nil || g.Width <= 0 || g.Height <= 0 {
		return nil
	}
	w, h := g.Width, g.Height
	vis := make([]float64, w*h)
	if maxRadius <= 0 {
		for i := range vis {
			vis[i] = 1.0
		}
		return vis
	}

	lx, ly, lz := lightDir[0], lightDir[1], lightDir[2]
	lLen := math.Sqrt(lx*lx + ly*ly + lz*lz)
	if lLen < 1e-9 {
		lx, ly, lz = -0.45, -0.60, 0.66
	} else {
		lx, ly, lz = lx/lLen, ly/lLen, lz/lLen
	}

	horizLen := math.Hypot(lx, ly)
	if horizLen < 1e-5 {
		for i := range vis {
			vis[i] = 1.0
		}
		return vis
	}
	dirX, dirY := lx/horizLen, ly/horizLen
	tanElev := math.Max(0.08, lz/horizLen)
	if softness <= 0 {
		softness = 1.5
	}

	// 3 penumbra ray directions: center, +15 deg, -15 deg
	const spreadRad = 15.0 * math.Pi / 180.0
	cosS, sinS := math.Cos(spreadRad), math.Sin(spreadRad)
	rayDirs := [3][2]float64{
		{dirX, dirY},
		{dirX*cosS - dirY*sinS, dirX*sinS + dirY*cosS},
		{dirX*cosS + dirY*sinS, -dirX*sinS + dirY*cosS},
	}
	rayWeights := [3]float64{0.50, 0.25, 0.25}

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				idx := y*w + x
				z0 := g.Z[idx]
				if z0 <= 1e-4 {
					vis[idx] = 1.0
					continue
				}

				totalVis := 0.0
				for rIdx, rd := range rayDirs {
					rx, ry := rd[0], rd[1]
					maxPen := 0.0
					for step := 1; step <= maxRadius; step++ {
						dist := float64(step)
						sx := int(math.Round(float64(x) + rx*dist))
						sy := int(math.Round(float64(y) + ry*dist))
						if sx < 0 || sx >= w || sy < 0 || sy >= h {
							break
						}
						zSample := g.Z[sy*w+sx]
						rayAlt := z0 + dist*tanElev
						if zSample > rayAlt {
							over := zSample - rayAlt
							pen := common.Constrain(over/(softness+0.12*dist), 0.0, 1.0)
							if pen > maxPen {
								maxPen = pen
							}
						}
					}
					totalVis += rayWeights[rIdx] * (1.0 - maxPen)
				}
				vis[idx] = common.Constrain(totalVis, 0.0, 1.0)
			}
		}
	})

	return vis
}

// ResolveToCanvasAdvanced performs deferred 3D shading over the FiberGBuffer combining:
// 1. 12-tap multi-scale Screen-Space Ambient Occlusion (SSAO)
// 2. Screen-space horizon-mapped soft shadows
// 3. Translucent Subsurface Scattering (SSS) across shadow terminators and elevated ridges
// 4. Physical Airy thin-film iridescence
// 5. Glossy Blinn-Phong specular highlights and grazing velvet rim light.
func (g *FiberGBuffer) ResolveToCanvasAdvanced(rc *Canvas, cfg GBufferMaterialConfig) {
	if g == nil || rc == nil {
		return
	}
	w, h := g.Width, g.Height
	lx, ly, lz := cfg.LightDir[0], cfg.LightDir[1], cfg.LightDir[2]
	lLen := math.Sqrt(lx*lx + ly*ly + lz*lz)
	if lLen < 1e-9 {
		lx, ly, lz = -0.45, -0.60, 0.66
	} else {
		lx, ly, lz = lx/lLen, ly/lLen, lz/lLen
	}

	hx, hy, hz := lx, ly, lz+1.0
	hLen := math.Sqrt(hx*hx + hy*hy + hz*hz)
	hx, hy, hz = hx/hLen, hy/hLen, hz/hLen

	var shadowVis []float64
	shadowStrength := common.Constrain(cfg.HorizonShadowStrength, 0.0, 1.0)
	if shadowStrength > 0 && cfg.HorizonRadius > 0 {
		shadowVis = g.ComputeHorizonShadows([3]float64{lx, ly, lz}, cfg.HorizonRadius, 1.6)
	}

	sssStrength := common.Constrain(cfg.SSSStrength, 0.0, 2.0)
	hasSSSTint := cfg.SSSTint.R > 0 || cfg.SSSTint.G > 0 || cfg.SSSTint.B > 0
	sssTintR, sssTintG, sssTintB, _ := ColorToLinearRGB(cfg.SSSTint)

	iridStrength := common.Constrain(cfg.IridescenceStrength, 0.0, 1.0)
	filmThickness := cfg.FilmThicknessNm
	if filmThickness <= 0 {
		filmThickness = 380.0
	}
	filmIOR := cfg.FilmIOR
	if filmIOR <= 1.0 {
		filmIOR = 1.42
	}

	aoOffsets := [12][2]int{
		{-2, 0}, {2, 0}, {0, -2}, {0, 2},
		{-4, -3}, {4, -3}, {-4, 3}, {4, 3},
		{-8, -5}, {8, -5}, {-5, 8}, {5, 8},
	}

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				idx := y*w + x
				z0 := g.Z[idx]
				if z0 <= 1e-4 {
					rc.SetPixelLinear(x, y, g.AlbR[idx], g.AlbG[idx], g.AlbB[idx], 1.0)
					continue
				}

				nx, ny, nz := g.Nx[idx], g.Ny[idx], g.Nz[idx]
				nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
				if nLen > 1e-9 {
					nx, ny, nz = nx/nLen, ny/nLen, nz/nLen
				} else {
					nx, ny, nz = 0, 0, 1
				}

				// 1. 12-tap multi-scale SSAO and local thickness estimation
				occ := 0.0
				ridgeExposure := 0.0
				for _, off := range aoOffsets {
					sx := common.ConstrainInt(x+off[0], 0, w-1)
					sy := common.ConstrainInt(y+off[1], 0, h-1)
					dz := g.Z[sy*w+sx] - z0
					if dz > 0.15 {
						occ += common.Constrain(dz/5.5, 0.0, 1.0)
					} else if dz < -0.15 {
						ridgeExposure += common.Constrain(-dz/4.5, 0.0, 1.0)
					}
				}
				ao := common.Constrain(1.0-(occ/12.0)*0.78, 0.18, 1.0)
				heightLift := common.Smoothstep(0.0, 4.5, z0)
				ao *= (0.45 + 0.55*heightLift)

				// 2. Horizon-mapped soft shadows
				shadow := 1.0
				if shadowVis != nil {
					shadow = 1.0 - shadowStrength*(1.0-shadowVis[idx])
				}

				albR, albG, albB := g.AlbR[idx], g.AlbG[idx], g.AlbB[idx]

				// 4. Thin-Film Iridescence structural color modulation
				cosTheta := common.Constrain(nz, 0.0, 1.0)
				var iridR, iridG, iridB float64
				if iridStrength > 0 {
					localThickness := filmThickness * (0.85 + 0.30*cosTheta + 0.04*z0)
					iridR, iridG, iridB = SampleThinFilmLinearRGB(cosTheta, localThickness, filmIOR)
					fresnelMix := iridStrength * (0.35 + 0.65*math.Pow(1.0-cosTheta, 1.4))
					albR = albR*(1.0-fresnelMix) + albR*iridR*1.35*fresnelMix
					albG = albG*(1.0-fresnelMix) + albG*iridG*1.35*fresnelMix
					albB = albB*(1.0-fresnelMix) + albB*iridB*1.35*fresnelMix
				}

				// Diffuse lighting with shadow attenuation
				ndotl := nx*lx + ny*ly + nz*lz
				wrapDiff := common.Constrain((ndotl+0.35)/1.35, 0.0, 1.0)
				litDiff := wrapDiff * (0.25 + 0.75*shadow)

				// 3. Subsurface Scattering (SSS) across shadow terminators and thin/elevated ridges
				var sssR, sssG, sssB float64
				if sssStrength > 0 {
					terminator := math.Exp(-math.Pow((ndotl+0.10)/0.42, 2.0))
					thinness := 0.35 + 0.65*common.Constrain(ridgeExposure/12.0, 0.0, 1.0)
					backScatter := common.Constrain((0.55-ndotl)/1.2, 0.0, 1.0)
					sssFactor := sssStrength * (0.65*terminator + 0.35*backScatter) * thinness * (0.4 + 0.6*heightLift)
					tR, tG, tB := albR*1.25+0.08, albG*0.85+0.02, albB*0.65
					if hasSSSTint {
						tR = 0.5*tR + 0.5*sssTintR
						tG = 0.5*tG + 0.5*sssTintG
						tB = 0.5*tB + 0.5*sssTintB
					}
					sssR = tR * sssFactor
					sssG = tG * sssFactor
					sssB = tB * sssFactor
				}

				// 5. Glossy Blinn-Phong specular highlights & grazing velvet rim light
				ndoth := math.Max(0.0, nx*hx+ny*hy+nz*hz)
				rough := common.Constrain(g.Roughness[idx], 0.15, 0.95)
				shininess := 8.0 + (1.0-rough)*42.0
				spec := math.Pow(ndoth, shininess) * (1.0 - rough*0.65) * 0.45 * ao * (0.20 + 0.80*shadow)

				rim := math.Pow(1.0-cosTheta, 2.2) * 0.18 * ao

				specR, specG, specB := 1.05, 0.96, 0.85
				if iridStrength > 0 {
					specR = specR*(1.0-iridStrength) + iridR*1.25*iridStrength
					specG = specG*(1.0-iridStrength) + iridG*1.25*iridStrength
					specB = specB*(1.0-iridStrength) + iridB*1.25*iridStrength
				}

				shade := (0.24 + 0.84*litDiff) * ao
				rOut := albR*shade + sssR + spec*specR + rim*albR
				gOut := albG*shade + sssG + spec*specG + rim*albG
				bOut := albB*shade + sssB + spec*specB + rim*albB

				rc.SetPixelLinear(x, y, rOut, gOut, bOut, 1.0)
			}
		}
	})
}
