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

type kintsugiPorcelain struct {
	shards    int
	goldVeins int
}

// NewKintsugiPorcelain returns a kintsugiPorcelain engine that renders a Song-Dynasty
// Ru-ware celadon mallow-rimmed porcelain plate (汝窑天青八瓣葵口盘) with Voronoi
// ice-crackle glaze (冰裂纹开片), incised lotus relief, and 3D 24K gold Kintsugi lacquer seams (金缮描金).
func NewKintsugiPorcelain(shards, goldVeins int) *kintsugiPorcelain {
	if shards <= 0 {
		shards = 96
	}
	if goldVeins <= 0 {
		goldVeins = 5
	}
	return &kintsugiPorcelain{
		shards:    shards,
		goldVeins: goldVeins,
	}
}

// Generative renders the Song Ru-ware celadon plate and 24K gold Kintsugi seams.
func (kp *kintsugiPorcelain) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	cx, cy := 0.5*fw, 0.5*fh

	// Authentic Song Ru-Ware Celadon (雨过天青) & 24K Urushi Gold Kintsugi Palette
	ebonyTable := color.RGBA{R: 12, G: 14, B: 20, A: 255}
	sesameRim := color.RGBA{R: 92, G: 74, B: 56, A: 255}
	celadonDeep := color.RGBA{R: 68, G: 118, B: 124, A: 255}
	celadonSky := color.RGBA{R: 128, G: 182, B: 180, A: 255}
	celadonMist := color.RGBA{R: 176, G: 214, B: 208, A: 255}
	agateBlush := color.RGBA{R: 212, G: 188, B: 176, A: 255}
	ironCrackle := color.RGBA{R: 54, G: 66, B: 68, A: 255}
	goldCrackle := color.RGBA{R: 168, G: 142, B: 96, A: 255}

	urushiLacquer := color.RGBA{R: 42, G: 22, B: 14, A: 255}
	goldBase := color.RGBA{R: 238, G: 178, B: 48, A: 255}
	goldBright := color.RGBA{R: 255, G: 226, B: 112, A: 255}
	ivoryGlint := color.RGBA{R: 255, G: 250, B: 228, A: 255}

	rc.FillBackground(ebonyTable)
	rc.SetBlendMode(render.BlendOver)

	// 1. Subtle woven bamboo / dark silk mat rings on the ebony table backdrop
	for r := 0.48 * fw; r > 0.42*fw; r -= 4.5 * scale {
		rc.DrawSoftRing(cx, cy, r, 1.2*scale, color.RGBA{R: 28, G: 34, B: 46, A: 255}, 0.45)
	}

	// 8-lobed Song Mallow-Rim (八瓣葵口) radius function
	rBase := 0.435 * math.Min(fw, fh)
	plateRadius := func(theta float64) float64 {
		lobe := math.Cos(8.0 * theta)
		// Slightly notched mallow indentation at 8 cusp angles
		cusp := -0.022 * math.Pow(0.5+0.5*lobe, 3.0) + 0.018*lobe
		return rBase * (1.0 + cusp)
	}

	// Build high-res polygon for the mallow-rimmed porcelain plate
	const rimSteps = 160
	outerRimPts := make([]render.Point, rimSteps)
	innerWellPts := make([]render.Point, rimSteps)
	for i := 0; i < rimSteps; i++ {
		th := (float64(i) / float64(rimSteps)) * 2.0 * math.Pi
		rOut := plateRadius(th)
		outerRimPts[i] = render.Pt(cx+rOut*math.Cos(th), cy+rOut*math.Sin(th))
		rWell := rOut * 0.78
		innerWellPts[i] = render.Pt(cx+rWell*math.Cos(th), cy+rWell*math.Sin(th))
	}

	plateSDF := render.SDFPolygon(outerRimPts)
	rc.RenderSDFDropShadow(plateSDF, render.Pt(7*scale, 10*scale), 20*scale, 3*scale, color.RGBA{R: 2, G: 3, B: 6, A: 255}, 0.88)

	// Fill outer sesame-iron foot rim, then clip all porcelain glaze & Kintsugi inside the plate
	rc.FillPolygon(outerRimPts, sesameRim, 1.0)
	innerGlazePts := render.InsetPolygon(outerRimPts, 3.2*scale)
	rc.FillPolygon(innerGlazePts, celadonSky, 1.0)

	rc.PushClipPolygon(innerGlazePts)

	// 2. Continuous 3D Bowl Concavity & Celadon Glaze Pooling Shading
	lightDir := render.Pt(-0.62, -0.64)
	ne := common.NewNoiseEngine(c.Seed() + 301)

	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			dx := fx - cx
			dy := fy - cy
			r := math.Hypot(dx, dy)
			th := math.Atan2(dy, dx)
			rMax := plateRadius(th)
			if r > rMax+2.0 {
				continue
			}
			u := r / rMax // normalized radius [0, 1]

			// Concave bowl normal + raised rim lip at u in [0.76, 0.98]
			slope := 0.42 * u
			if u > 0.76 && u < 0.82 {
				// Pooled celadon groove at the cavetto ring (圈足内凹积釉处)
				slope -= 0.25 * math.Sin(((u-0.76)/0.06)*math.Pi)
			}
			nx := (dx / math.Max(1.0, r)) * slope
			ny := (dy / math.Max(1.0, r)) * slope
			ndotl := common.Constrain(0.58-(nx*lightDir.X+ny*lightDir.Y), 0, 1)

			// Subtle agate pinkish-ivory underglaze warmth where glaze is thin, deep celadon where pooled
			pool := 0.5 + 0.5*math.Cos(u*math.Pi*2.5)
			if u > 0.74 && u < 0.83 {
				pool = 0.92 // thick glaze pooling in the well ring
			}
			nVal := ne.FBM2D(fx/(85.0*scale), fy/(85.0*scale), 3, 2.0, 0.5)
			var col color.RGBA
			if ndotl < 0.48 {
				col = render.LerpOklab(celadonDeep, celadonSky, ndotl/0.48)
			} else {
				col = render.LerpOklab(celadonSky, celadonMist, (ndotl-0.48)/0.52)
			}
			if pool > 0.65 {
				col = render.LerpOklab(col, celadonDeep, (pool-0.65)*0.55)
			} else if u > 0.88 {
				// Thin glaze near the mallow rim reveals warm pinkish agate body (玛瑙入釉微泛粉)
				col = render.LerpOklab(col, agateBlush, (u-0.88)*2.8)
			}
			if nVal > 0.15 {
				col = render.LerpOklab(col, celadonMist, nVal*0.22)
			}

			// Soft broad ceramic specular reflection in upper-left quadrant
			specDist := math.Hypot(dx+0.22*rBase, dy+0.24*rBase) / rBase
			if specDist < 0.42 {
				spec := math.Pow(1.0-specDist/0.42, 2.4) * 0.36
				col = render.LerpOklab(col, ivoryGlint, spec)
			}
			rc.BlendPixel(x, y, col, 1.0)
		}
	}

	// 3. Underglaze Carved Song Lotus Medallion (暗刻宝相莲瓣纹) in the Central Well
	const lotusPetals = 12
	dLotus := (2.0 * math.Pi) / float64(lotusPetals)
	rLotusOut := rBase * 0.54
	rLotusIn := rBase * 0.14
	for p := 0; p < lotusPetals; p++ {
		thC := float64(p) * dLotus
		span := dLotus * 0.44
		petalPts := []render.Point{
			render.Pt(cx+rLotusIn*math.Cos(thC-span*0.6), cy+rLotusIn*math.Sin(thC-span*0.6)),
			render.Pt(cx+(rLotusIn+0.48*(rLotusOut-rLotusIn))*math.Cos(thC-span), cy+(rLotusIn+0.48*(rLotusOut-rLotusIn))*math.Sin(thC-span)),
			render.Pt(cx+rLotusOut*math.Cos(thC), cy+rLotusOut*math.Sin(thC)),
			render.Pt(cx+(rLotusIn+0.48*(rLotusOut-rLotusIn))*math.Cos(thC+span), cy+(rLotusIn+0.48*(rLotusOut-rLotusIn))*math.Sin(thC+span)),
			render.Pt(cx+rLotusIn*math.Cos(thC+span*0.6), cy+rLotusIn*math.Sin(thC+span*0.6)),
		}
		// Carved pooling shadow on one side, translucent highlight on the other
		path := &canvas.Path{}
		path.MoveTo(petalPts[0].X, petalPts[0].Y)
		for i := 1; i < len(petalPts); i++ {
			path.LineTo(petalPts[i].X, petalPts[i].Y)
		}
		path.Close()
		rc.StrokePath(path, 2.4*scale, canvas.RoundCap, canvas.RoundJoin, celadonDeep, 0.42)
		inPetal := render.InsetPolygon(petalPts, 2.2*scale)
		inPath := &canvas.Path{}
		inPath.MoveTo(inPetal[0].X, inPetal[0].Y)
		for i := 1; i < len(inPetal); i++ {
			inPath.LineTo(inPetal[i].X, inPetal[i].Y)
		}
		inPath.Close()
		rc.StrokePath(inPath, 1.1*scale, canvas.RoundCap, canvas.RoundJoin, celadonMist, 0.52)
	}
	rc.DrawSoftRing(cx, cy, rBase*0.77, 2.8*scale, celadonDeep, 0.48)
	rc.DrawSoftRing(cx, cy, rBase*0.76, 1.2*scale, celadonMist, 0.55)

	// 4. Multi-Scale Ru-Ware Ice-Crackle Network (鱼鳞开片与金丝铁线)
	rng := rand.New(rand.NewSource(c.Seed() + 707))
	numCells := kp.shards
	if numCells < 40 {
		numCells = 96
	}
	seeds := make([]render.Point, 0, numCells)
	for i := 0; i < numCells; i++ {
		// Golden-angle phyllotaxis + jitter inside plate for balanced glaze crackle shards
		u := math.Sqrt((float64(i) + 0.5) / float64(numCells))
		th := float64(i)*2.39996323 + (rng.Float64()-0.5)*0.22
		r := u * rBase * 0.98
		seeds = append(seeds, render.Pt(cx+r*math.Cos(th), cy+r*math.Sin(th)))
	}

	// Compute Voronoi cells by half-plane clipping and stroke each glaze shard
	ironPath := &canvas.Path{}
	goldHairPath := &canvas.Path{}
	highlightPath := &canvas.Path{}

	for i, sA := range seeds {
		poly := []render.Point{
			render.Pt(0, 0),
			render.Pt(fw, 0),
			render.Pt(fw, fh),
			render.Pt(0, fh),
		}
		for j, sB := range seeds {
			if i == j {
				continue
			}
			if sA.Dist(sB) > rBase*0.62 {
				continue
			}
			mid := sA.Lerp(sB, 0.5)
			dir := sB.Sub(sA)
			poly = clipHalfPlane(poly, mid, dir)
			if len(poly) < 3 {
				break
			}
		}
		if len(poly) < 3 {
			continue
		}

		// Subtle per-shard glaze tilt tint
		shardTint := celadonDeep
		if i%3 == 0 {
			shardTint = celadonMist
		}
		rc.FillPolygon(render.InsetPolygon(poly, 1.2*scale), shardTint, 0.08)

		for k := 0; k < len(poly); k++ {
			p0 := poly[k]
			p1 := poly[(k+1)%len(poly)]
			if (i+k)%3 == 0 {
				goldHairPath.MoveTo(p0.X, p0.Y)
				goldHairPath.LineTo(p1.X, p1.Y)
			} else {
				ironPath.MoveTo(p0.X, p0.Y)
				ironPath.LineTo(p1.X, p1.Y)
			}
			// Refractive light catch along one side of the glaze crackle
			highlightPath.MoveTo(p0.X+0.8*scale, p0.Y+0.8*scale)
			highlightPath.LineTo(p1.X+0.8*scale, p1.Y+0.8*scale)
		}
	}

	rc.StrokePath(highlightPath, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, celadonMist, 0.50)
	rc.StrokePath(ironPath, 0.95*scale, canvas.RoundCap, canvas.RoundJoin, ironCrackle, 0.62)
	rc.StrokePath(goldHairPath, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, goldCrackle, 0.58)

	// 5. 3D 24K Gold Kintsugi Repair Seams (金缮描金主裂谷与分叉枝脉)
	numVeins := kp.goldVeins
	if numVeins < 3 {
		numVeins = 5
	}

	// Primary fracture hub slightly off-center for wabi-sabi asymmetry
	hubA := render.Pt(cx-0.08*rBase, cy-0.05*rBase)
	hubB := render.Pt(cx+0.14*rBase, cy+0.11*rBase)

	buildFracturePolyline := func(start, end render.Point, seedOff float64, amp float64) []render.Point {
		const steps = 44
		pts := make([]render.Point, 0, steps+1)
		chord := end.Sub(start)
		perp := render.Pt(-chord.Y, chord.X).Norm()
		for s := 0; s <= steps; s++ {
			t := float64(s) / float64(steps)
			base := start.Lerp(end, t)
			env := math.Sin(t * math.Pi)
			n := ne.FBM2D(t*3.8+seedOff, seedOff*1.7, 4, 2.1, 0.55)
			pts = append(pts, base.Add(perp.Mul(n*amp*env)))
		}
		return pts
	}

	renderGoldSeam := func(pts []render.Point, baseWidth float64) {
		if len(pts) < 2 {
			return
		}
		// 1. Dark urushi lacquer bedding shadow beneath the raised gold seam
		shadowPath := &canvas.Path{}
		shadowPath.MoveTo(pts[0].X+1.4*scale, pts[0].Y+1.8*scale)
		for i := 1; i < len(pts); i++ {
			shadowPath.LineTo(pts[i].X+1.4*scale, pts[i].Y+1.8*scale)
		}
		rc.StrokePath(shadowPath, baseWidth*1.45, canvas.RoundCap, canvas.RoundJoin, urushiLacquer, 0.72)

		// 2. 3D Beveled Impasto Gold Lacquer Tube
		rc.DrawImpastoStroke(pts, baseWidth, goldBase, lightDir, 0.98)

		// 3. Burnished 24K Gold Specular Spine
		rc.DrawTaperedSpline(pts, baseWidth*0.42, render.TaperBothEnds, func(t float64) color.RGBA {
			shimmer := 0.5 + 0.5*math.Sin(t*math.Pi*6.0)
			return render.LerpOklab(goldBright, ivoryGlint, shimmer*0.75)
		}, 0.95)
	}

	// Bridge fracture between hubA and hubB
	bridgePts := buildFracturePolyline(hubA, hubB, 1.1, 18.0*scale)
	renderGoldSeam(bridgePts, 7.2*scale)

	// Radiate main Kintsugi fractures to the mallow rim and fork secondary branches
	for v := 0; v < numVeins; v++ {
		ang := (float64(v)/float64(numVeins))*2.0*math.Pi + 0.28 + 0.16*math.Sin(float64(v)*2.1)
		origin := hubA
		if v%2 == 1 {
			origin = hubB
		}
		rRim := plateRadius(ang) * 1.02
		target := render.Pt(cx+rRim*math.Cos(ang), cy+rRim*math.Sin(ang))
		mainPts := buildFracturePolyline(origin, target, float64(v)*3.7+2.5, 24.0*scale)
		renderGoldSeam(mainPts, 6.4*scale)

		// Fork a secondary branch from the middle of each main fracture toward an adjacent rim lobe
		forkIdx := len(mainPts) * 4 / 9
		forkStart := mainPts[forkIdx]
		forkAng := ang + 0.42
		if v%2 == 1 {
			forkAng = ang - 0.38
		}
		rFork := plateRadius(forkAng) * 1.02
		forkEnd := render.Pt(cx+rFork*math.Cos(forkAng), cy+rFork*math.Sin(forkAng))
		forkPts := buildFracturePolyline(forkStart, forkEnd, float64(v)*5.3+9.1, 15.0*scale)
		renderGoldSeam(forkPts, 4.2*scale)

		// Raised golden lacquer bead at the fork junction and rim staple
		rc.DrawGlowDot(forkStart.X, forkStart.Y, 3.6*scale, ivoryGlint, 0.95, 0.65)
	}

	// Glowing golden lacquer pools at the two central fracture hubs
	for _, hub := range []render.Point{hubA, hubB} {
		orb := render.SDFCircle(hub, 5.2*scale)
		rc.RenderSDFBevel(orb, goldBase, urushiLacquer, ivoryGlint, 2.5*scale, lightDir, 0.98)
	}

	rc.PopClip()

	// 6. Outer Mallow-Rim Sesame-Iron & Gold Trim (葵口铁足与描金口沿)
	rimPath := &canvas.Path{}
	rimPath.MoveTo(outerRimPts[0].X, outerRimPts[0].Y)
	for i := 1; i < len(outerRimPts); i++ {
		rimPath.LineTo(outerRimPts[i].X, outerRimPts[i].Y)
	}
	rimPath.Close()
	rc.StrokePath(rimPath, 3.2*scale, canvas.RoundCap, canvas.RoundJoin, sesameRim, 0.96)
	rc.StrokePath(rimPath, 1.2*scale, canvas.RoundCap, canvas.RoundJoin, goldCrackle, 0.78)

	rc.ApplyPaperTexture(0.035, c.Seed()+19)
	c.MergeRenderCanvas(rc)
}

// clipHalfPlane clips polygon poly against the half-plane (P - mid) . dir <= 0.
func clipHalfPlane(poly []render.Point, mid, dir render.Point) []render.Point {
	if len(poly) == 0 {
		return nil
	}
	out := make([]render.Point, 0, len(poly)+2)
	for i := 0; i < len(poly); i++ {
		cur := poly[i]
		prev := poly[(i+len(poly)-1)%len(poly)]
		dCur := (cur.X-mid.X)*dir.X + (cur.Y-mid.Y)*dir.Y
		dPrev := (prev.X-mid.X)*dir.X + (prev.Y-mid.Y)*dir.Y
		if dCur <= 0 {
			if dPrev > 0 {
				t := dPrev / (dPrev - dCur)
				out = append(out, prev.Lerp(cur, t))
			}
			out = append(out, cur)
		} else if dPrev <= 0 {
			t := dPrev / (dPrev - dCur)
			out = append(out, prev.Lerp(cur, t))
		}
	}
	return out
}
