package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type wireframeTerrain struct {
	gridRes     int
	reliefScale float64
}

// NewWireframeTerrain returns a wireframeTerrain engine that renders a 3D
// perspective-projected low-poly alpine glacier & fjord canyon with true
// per-row hidden-line removal, seamless forward canyon flight animation,
// undulating 3D fjord waves, and Aurora Borealis sky reflections.
func NewWireframeTerrain(gridRes int, reliefScale float64) *wireframeTerrain {
	if gridRes <= 0 {
		gridRes = 52
	}
	if reliefScale <= 0 {
		reliefScale = 1.0
	}
	return &wireframeTerrain{
		gridRes:     gridRes,
		reliefScale: reliefScale,
	}
}

// Generative renders the 3D perspective triangulated alpine terrain mesh.
func (wt *wireframeTerrain) Generative(c *generativeart.Canva) {
	wt.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree forward canyon flight loop (phase in [0, 1)).
func (wt *wireframeTerrain) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, loopPhase float64) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	animPhase := loopPhase * 2.0 * math.Pi

	// High-Contrast Glacier, Fjord, Sunset & Aurora Palette
	nightSky := color.RGBA{R: 5, G: 7, B: 18, A: 255}
	horizonIndigo := color.RGBA{R: 18, G: 32, B: 74, A: 255}
	auroraTeal := color.RGBA{R: 36, G: 206, B: 196, A: 255}
	auroraMagenta := color.RGBA{R: 204, G: 48, B: 142, A: 255}
	fjordCobalt := color.RGBA{R: 12, G: 44, B: 108, A: 255}
	glacierCyan := color.RGBA{R: 78, G: 232, B: 252, A: 255}
	sunsetCoral := color.RGBA{R: 236, G: 76, B: 56, A: 255}
	solarGold := color.RGBA{R: 255, G: 204, B: 68, A: 255}
	snowIvory := color.RGBA{R: 255, G: 252, B: 238, A: 255}
	shadowSlate := color.RGBA{R: 9, G: 14, B: 32, A: 255}

	rc.FillBackground(nightSky)
	rc.SetBlendMode(render.BlendOver)

	// 1. Dynamic Aurora Borealis Curtains & Blazing Horizon Sun at the Canyon Vanishing Point
	sunX := 0.50*w + 10.0*scale*math.Sin(animPhase)
	sunY := 0.185 * h
	skyH := int(0.40 * h)
	for y := 0; y < skyH; y++ {
		fy := float64(y) / float64(skyH)
		baseSky := render.LerpOklab(nightSky, horizonIndigo, math.Pow(fy, 1.35))
		for x := 0; x < int(w); x++ {
			fx := float64(x) / w
			curtain1 := 0.5 + 0.5*math.Sin(fx*math.Pi*3.5-animPhase+0.65*math.Sin(fy*math.Pi*2.0))
			curtain2 := 0.5 + 0.5*math.Cos(fx*math.Pi*5.0+animPhase-0.8)
			auroraEnv := math.Sin(fy*math.Pi) * (1.0 - fy*0.40)
			auroraW := math.Pow(curtain1*0.65+curtain2*0.35, 2.0) * auroraEnv

			col := baseSky
			if auroraW > 0.04 {
				aCol := render.LerpOklab(auroraTeal, auroraMagenta, 0.5+0.5*math.Sin(fx*math.Pi*2.2+animPhase))
				col = render.LerpOklab(col, aCol, common.Clamp01(auroraW*0.82))
			}

			dSun := math.Hypot(float64(x)-sunX, float64(y)-sunY) / (92.0 * scale)
			if dSun < 1.85 {
				glow := math.Exp(-dSun * dSun * 2.1)
				sCol := render.LerpOklab(sunsetCoral, solarGold, common.Clamp01(1.25-dSun))
				if dSun < 0.34 {
					sCol = render.LerpOklab(solarGold, snowIvory, (0.34-dSun)/0.34)
				}
				col = render.LerpOklab(col, sCol, common.Clamp01(glow*0.96))
			}
			rc.BlendPixel(x, y, col, 1.0)
		}
	}

	starRng := rand.New(rand.NewSource(202615))
	for i := 0; i < 95; i++ {
		sx := starRng.Float64() * w
		sy := starRng.Float64() * 0.23 * h
		twinkle := 0.45 + 0.45*math.Sin(animPhase+float64(i)*1.7)
		rc.SplatPoint(sx, sy, snowIvory, twinkle*(1.0-sy/(0.24*h)))
	}

	ne := common.NewNoiseEngine(c.Seed() + 707)
	nx := wt.gridRes
	if nx < 36 || nx > 68 {
		nx = 50
	}
	ny := nx

	// 2. Seamless Forward-Flying Canyon Heightfield:
	// World longitudinal phase phiV = v * pi - animPhase wraps with exact 2*pi periodicity
	// as the viewer flies continuously forward down the glacial canyon!
	evalTerrain := func(u, vScreen float64) (elev float64, isWater bool) {
		phiV := vScreen*math.Pi*1.35 - animPhase
		cv := math.Cos(phiV)
		sv := math.Sin(phiV)

		// Meandering glacial fjord centerline
		riverCenter := 0.15 * math.Sin(phiV)
		fjordDist := math.Abs(u - riverCenter)

		// Towering left & right alpine canyon walls with periodic mountain massifs advancing toward viewer
		wallU := common.Smoothstep(0.11, 0.46, fjordDist)
		massifWave := 0.62 + 0.38*(0.5+0.5*math.Cos(phiV-0.4*math.Abs(u)))

		// Sample 3D/2D ridged noise along the closed cylinder (u, cos(phiV), sin(phiV)) for 100% seamless flight!
		r1 := ne.Ridged2D(u*2.4+0.65*cv+1.3, 0.65*sv-1.7, 4, 2.1, 0.52)
		r2 := ne.Ridged2D(u*3.6-0.55*sv-0.8, 0.55*cv+2.1, 3, 2.05, 0.5)
		fbm := 0.5 + 0.5*ne.FBM2D(u*2.0+0.5*cv, 0.5*sv, 3, 2.0, 0.5)

		// Dramatic twin ridge crests at |u - riverCenter| ~ 0.56
		ridgeProfile := math.Exp(-math.Pow((fjordDist-0.56)/0.38, 2.0))*0.78 + 0.28*wallU
		mtnElev := (ridgeProfile*massifWave*(0.52+0.48*r1) + 0.26*wallU*r2 + 0.10*wallU*fbm) * wt.reliefScale

		// 3D undulating fjord waves propagating along the canyon channel
		waterWave := 0.13 + 0.020*math.Sin(phiV*4.0)*math.Cos((u-riverCenter)*math.Pi*5.0)
		if fjordDist < 0.13 || mtnElev <= waterWave+0.008 {
			return waterWave, true
		}
		return mtnElev, false
	}

	type meshVert struct {
		p3      vec3
		pt2D    render.Point
		elev    float64
		isWater bool
	}

	verts := make([][]meshVert, ny+1)
	for iy := 0; iy <= ny; iy++ {
		verts[iy] = make([]meshVert, nx+1)
		tz := float64(iy) / float64(ny) // 0 = far horizon, 1 = foreground
		vScreen := tz*2.0 - 1.0

		persp := 0.94 + 0.28*tz
		syBase := (0.22 + tz*0.98) * h
		zScale := 268.0 * scale * (0.72 + 0.36*tz)

		rowShift := 0.0
		if iy%2 == 1 {
			rowShift = 0.5 / float64(nx)
		}

		for ix := 0; ix <= nx; ix++ {
			tx := (float64(ix)/float64(nx) - 0.5) + rowShift
			u := tx * 2.18
			elev, isWater := evalTerrain(u, vScreen)

			sx := 0.5*w + tx*w*1.16*persp
			sy := syBase - (elev-0.12)*zScale

			verts[iy][ix] = meshVert{
				p3:      vec3{u, vScreen, elev * 0.68},
				pt2D:    render.Pt(sx, sy),
				elev:    elev,
				isWater: isWater,
			}
		}
	}

	lightAz := 0.35 * math.Sin(animPhase)
	keyLight := (vec3{-0.58*math.Cos(lightAz) + 0.32*math.Sin(lightAz), -0.46, 0.65}).norm()
	fillLight := (vec3{0.58, 0.42, 0.70}).norm()
	halfVec := (vec3{keyLight.x, keyLight.y, keyLight.z + 1.0}).norm()

	// Elevation scan pulse climbing from valley floor to summit
	scanCenter := 0.20 + loopPhase*0.68

	// 3. Render Row by Row from Back (iy = 0) to Front (iy = ny-1) with Immediate Per-Row Wireframe Stroking
	// so foreground mountain peaks cleanly occlude rear wireframe lines (true hidden-line removal!)
	for iy := 0; iy < ny; iy++ {
		rowCyan := &canvas.Path{}
		rowGold := &canvas.Path{}
		rowPulse := &canvas.Path{}
		var rowSummits []render.Point

		for ix := 0; ix < nx; ix++ {
			v00 := verts[iy][ix]
			v10 := verts[iy][ix+1]
			v11 := verts[iy+1][ix+1]
			v01 := verts[iy+1][ix]

			var tris [2][3]meshVert
			if iy%2 == 0 {
				tris = [2][3]meshVert{{v00, v10, v01}, {v10, v11, v01}}
			} else {
				tris = [2][3]meshVert{{v00, v10, v11}, {v00, v11, v01}}
			}

			for _, tri := range tris {
				a, b, cPt := tri[0], tri[1], tri[2]
				avgElev := (a.elev + b.elev + cPt.elev) / 3.0

				n := b.p3.sub(a.p3).cross(cPt.p3.sub(a.p3)).norm()
				if n.z < 0 {
					n = vec3{-n.x, -n.y, -n.z}
				}

				ndotl := common.Constrain(n.dot(keyLight), 0, 1)
				ndotFill := common.Constrain(n.dot(fillLight), 0, 1)
				spec := math.Pow(common.Constrain(n.dot(halfVec), 0, 1), 16.0)

				dScan := math.Abs(avgElev - scanCenter)
				if dScan > 0.34 {
					dScan = math.Abs(0.68 - dScan)
				}
				scanGlow := math.Exp(-math.Pow(dScan/0.048, 2.0))

				allWater := a.isWater && b.isWater && cPt.isWater
				anyWater := a.isWater || b.isWater || cPt.isWater

				var col color.RGBA
				if allWater {
					waveCrest := common.Clamp01((avgElev - 0.11) / 0.040)
					col = render.LerpOklab(fjordCobalt, auroraTeal, waveCrest*0.82)
					uMid := math.Abs((a.p3.x + b.p3.x + cPt.p3.x) / 3.0)
					if uMid < 0.18 {
						sunPath := (1.0 - uMid/0.18) * (0.42 + 0.58*waveCrest)
						col = render.LerpOklab(col, solarGold, sunPath*0.72)
					}
					if spec > 0.05 {
						col = render.LerpOklab(col, snowIvory, common.Constrain(spec*0.85, 0, 1))
					}
				} else {
					uE := common.Clamp01((avgElev - 0.14) / 0.64)
					var baseCol color.RGBA
					switch {
					case uE < 0.34:
						baseCol = render.LerpOklab(fjordCobalt, auroraTeal, uE/0.34)
					case uE < 0.66:
						baseCol = render.LerpOklab(auroraTeal, sunsetCoral, (uE-0.34)/0.32)
					default:
						baseCol = render.LerpOklab(solarGold, snowIvory, (uE-0.66)/0.34)
					}

					illum := common.Constrain(0.18+0.74*ndotl+0.24*ndotFill, 0.08, 1.0)
					col = render.LerpOklab(shadowSlate, baseCol, illum)
					if !anyWater && scanGlow > 0.15 {
						pulseCol := render.LerpOklab(auroraMagenta, solarGold, uE)
						col = render.LerpOklab(col, pulseCol, scanGlow*0.52)
					}
					if spec > 0.04 {
						col = render.LerpOklab(col, snowIvory, common.Constrain(spec*0.82, 0, 1))
					}
				}

				rc.FillPolygon([]render.Point{a.pt2D, b.pt2D, cPt.pt2D}, col, 1.0)

				// Classify wireframe edge color (never put white pulse on shoreline/water)
				var edgePath *canvas.Path
				if !anyWater && avgElev > 0.24 && scanGlow > 0.52 {
					edgePath = rowPulse
				} else if !anyWater && avgElev > 0.44 {
					edgePath = rowGold
				} else {
					edgePath = rowCyan
				}
				edgePath.MoveTo(a.pt2D.X, a.pt2D.Y)
				edgePath.LineTo(b.pt2D.X, b.pt2D.Y)
				edgePath.LineTo(cPt.pt2D.X, cPt.pt2D.Y)
				edgePath.Close()
			}

			if !v00.isWater && v00.elev > 0.58 && (ix*3+iy*5)%5 == 0 {
				rowSummits = append(rowSummits, v00.pt2D)
			}
		}

		// Stroke this row's wireframe immediately so closer rows (iy+1..ny) occlude rear wireframes!
		rc.StrokePath(rowCyan, 0.65*scale, canvas.RoundCap, canvas.RoundJoin, glacierCyan, 0.38)
		rc.StrokePath(rowGold, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, solarGold, 0.68)
		rc.StrokePath(rowPulse, 1.15*scale, canvas.RoundCap, canvas.RoundJoin, snowIvory, 0.82)

		for i, pt := range rowSummits {
			dotCol := solarGold
			if i%2 == 1 {
				dotCol = glacierCyan
			}
			rc.DrawGlowDot(pt.X, pt.Y, 2.2*scale, dotCol, 0.88, 0.75)
		}
	}

	rc.ApplyBloom(0.80, 6, 0.24)
	c.MergeRenderCanvas(rc)
}
