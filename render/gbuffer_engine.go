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
