package arts

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type fractalFlame struct {
	pointsN  int
	symmetry int
}

// NewFractalFlame returns a fractalFlame engine that renders Scott Draves' Non-Linear IFS
// Fractal Flame with rotational & dihedral symmetry, sub-pixel bilinear histogram splatting,
// 99.8th-percentile log-density tone mapping, and Oklab HDR color accumulation.
func NewFractalFlame(pointsN, symmetry int) *fractalFlame {
	if pointsN <= 0 {
		pointsN = 1800000
	}
	if symmetry <= 0 {
		symmetry = 5
	}
	return &fractalFlame{
		pointsN:  pointsN,
		symmetry: symmetry,
	}
}

// Generative renders the Non-Linear IFS Fractal Flame onto the canvas.
func (ff *fractalFlame) Generative(c *generativeart.Canva) {
	pointsN := ff.pointsN
	if pointsN <= 0 {
		pointsN = 1800000
	}
	symmetry := ff.symmetry
	if symmetry <= 0 {
		symmetry = 5
	}

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

	nPix := w * h
	density := make([]float64, nPix)
	rAcc := make([]float64, nPix)
	gAcc := make([]float64, nPix)
	bAcc := make([]float64, nPix)

	rng := rand.New(rand.NewSource(rand.Int63()))

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Magma
	}
	grad := render.NewMultiStopGradient(schema...)
	whiteHot := color.RGBA{R: 255, G: 250, B: 236, A: 255}

	// Precompute a 512-entry linear RGB lookup table across the palette with a vibrant
	// complementary cyan/electric-blue harmonic stop so different flame branches have strong color contrast.
	const lutSize = 512
	var lutR, lutG, lutB [lutSize]float64
	accentCyan := color.RGBA{R: 48, G: 215, B: 255, A: 255}
	for i := 0; i < lutSize; i++ {
		t := float64(i) / float64(lutSize-1)
		// Map [0, 0.80] across main gradient, and [0.80, 1.0] into a luminous cyan-violet filament accent
		var cCol color.RGBA
		if t < 0.80 {
			cCol = grad.SampleOklab(common.Clamp01(0.12 + 0.86*(t/0.80)))
		} else {
			u := (t - 0.80) / 0.20
			cCol = render.LerpOklab(grad.SampleOklab(0.65), accentCyan, math.Sin(u*math.Pi)*0.82)
		}
		lutR[i], lutG[i], lutB[i], _ = render.ColorToLinearRGB(cCol)
	}

	// Precompute rotational symmetry angles for fast lookup in the inner chaos-game loop.
	cosSym := make([]float64, symmetry)
	sinSym := make([]float64, symmetry)
	for s := 0; s < symmetry; s++ {
		theta := 2.0 * math.Pi * float64(s) / float64(symmetry)
		cosSym[s] = math.Cos(theta)
		sinSym[s] = math.Sin(theta)
	}

	// Cumulative probability weights for 4 anisotropic flame branches (producing sharp ribbon caustics and lacy voids):
	// F0 (weight 0.48): Contractive logarithmic spiral similarity + gentle swirl (sweeping outer spiral arms)
	// F1 (weight 0.26): Off-center Spherical inversion (recursive lacy circular voids and arches)
	// F2 (weight 0.15): Thin sheared ribbon + Horseshoe fold (razor-sharp golden/crimson caustic ridges)
	// F3 (weight 0.11): Eccentric contractive curl (electric cyan/indigo counter-tendrils)
	const (
		cdf0 = 0.48
		cdf1 = 0.74
		cdf2 = 0.89
	)
	colIdx := [4]float64{0.22, 0.56, 0.76, 0.92}

	swirlTwist := 0.38 + 0.05*rng.Float64()
	ringShiftX := 0.68 + 0.04*(rng.Float64()-0.5)
	ringShiftY := -0.42 + 0.04*(rng.Float64()-0.5)

	evalFlameStep := func(x, y float64, u float64) (nx, ny, cTarget float64, cSpeed float64) {
		switch {
		case u < cdf0:
			// F0: Contractive spiral similarity + subtle Swirl (sweeping logarithmic arms with dark voids between)
			ax := 0.76*x - 0.36*y + 0.34
			ay := 0.36*x + 0.72*y - 0.16
			r2 := ax*ax + ay*ay
			sinR2, cosR2 := math.Sincos(r2 * swirlTwist)
			swirlX := ax*cosR2 - ay*sinR2
			swirlY := ax*sinR2 + ay*cosR2
			nx = 0.84*swirlX + 0.16*math.Sin(ax*1.8)
			ny = 0.84*swirlY + 0.16*math.Sin(ay*1.8)
			return nx, ny, colIdx[0], 0.42

		case u < cdf1:
			// F1: Off-center Spherical inversion creating recursive hollow circular voids and lacy filaments
			ax := 0.46*x + 0.16*y + ringShiftX
			ay := -0.16*x + 0.46*y + ringShiftY
			r2 := ax*ax + ay*ay + 0.14
			invR2 := 0.62 / r2
			nx = ax*invR2 - 0.24
			ny = ay*invR2 + 0.18
			return nx, ny, colIdx[1], 0.68

		case u < cdf2:
			// F2: Thin sheared ribbon + Horseshoe variation (sharp golden/amber caustic folds)
			ax := 0.58*x - 0.20*y - 0.62
			ay := 0.12*x + 0.18*y + 0.44
			r := math.Hypot(ax, ay) + 1e-7
			invR := 0.82 / r
			hsX := (ax - ay) * (ax + ay) * invR
			hsY := 2.0 * ax * ay * invR
			nx = 0.68*hsX + 0.32*ax
			ny = 0.68*hsY + 0.32*ay
			return nx, ny, colIdx[2], 0.74

		default:
			// F3: Eccentric sheared ribbon + sinusoidal curl (delicate cyan/violet outer plumes)
			ax := 0.52*x + 0.28*y - 0.28
			ay := -0.14*x + 0.22*y - 0.64
			r := math.Hypot(ax, ay) + 1e-6
			theta := math.Atan2(ay, ax)
			nx = 0.78*ax + 0.36*math.Sin(theta+r*1.7)
			ny = 0.78*ay + 0.36*math.Cos(theta-r*1.7)
			return nx, ny, colIdx[3], 0.82
		}
	}

	x := 0.15
	y := -0.12
	col := 0.5

	// 40 burn-in steps to settle onto the fractal flame attractor.
	for i := 0; i < 40; i++ {
		var cTarget, cSpeed float64
		x, y, cTarget, cSpeed = evalFlameStep(x, y, rng.Float64())
		col = (1.0-cSpeed)*col + cSpeed*cTarget
	}

	cx := fw * 0.5
	cy := fh * 0.5
	// Map [-1.85, 1.85] to centered canvas pixel coordinates so tendrils fill the frame generously.
	coordScale := (minDim * 0.5) / 1.85

	for i := 0; i < pointsN; i++ {
		nx, ny, cTarget, cSpeed := evalFlameStep(x, y, rng.Float64())
		if math.IsNaN(nx) || math.IsNaN(ny) || math.IsInf(nx, 0) || math.IsInf(ny, 0) || nx*nx+ny*ny > 64.0 {
			x = (rng.Float64()*2.0 - 1.0) * 0.5
			y = (rng.Float64()*2.0 - 1.0) * 0.5
			col = 0.5
			continue
		}
		x, y = nx, ny
		col = (1.0-cSpeed)*col + cSpeed*cTarget

		// Apply clean N-fold rotational symmetry when splatting so the asymmetric spiral attractor
		// forms an intricate multi-armed galaxy of interlocking colored filaments and dark voids.
		s := i % symmetry
		cs := cosSym[s]
		sn := sinSym[s]
		sx := x*cs - y*sn
		sy := x*sn + y*cs

		px := cx + sx*coordScale
		py := cy + sy*coordScale

		x0 := int(math.Floor(px))
		y0 := int(math.Floor(py))
		if x0 < 0 || x0 >= w-1 || y0 < 0 || y0 >= h-1 {
			continue
		}

		cIdx := int(common.Clamp01(col) * float64(lutSize-1))
		cr, cg, cb := lutR[cIdx], lutG[cIdx], lutB[cIdx]

		fx := px - float64(x0)
		fy := py - float64(y0)
		w00 := (1.0 - fx) * (1.0 - fy)
		w10 := fx * (1.0 - fy)
		w01 := (1.0 - fx) * fy
		w11 := fx * fy

		idx00 := y0*w + x0
		idx10 := idx00 + 1
		idx01 := idx00 + w
		idx11 := idx01 + 1

		density[idx00] += w00
		density[idx10] += w10
		density[idx01] += w01
		density[idx11] += w11

		rAcc[idx00] += w00 * cr
		gAcc[idx00] += w00 * cg
		bAcc[idx00] += w00 * cb

		rAcc[idx10] += w10 * cr
		gAcc[idx10] += w10 * cg
		bAcc[idx10] += w10 * cb

		rAcc[idx01] += w01 * cr
		gAcc[idx01] += w01 * cg
		bAcc[idx01] += w01 * cb

		rAcc[idx11] += w11 * cr
		gAcc[idx11] += w11 * cg
		bAcc[idx11] += w11 * cb
	}

	// Compute the 99.7th percentile density (p997) across non-zero pixels.
	nonZero := make([]float64, 0, nPix/4)
	p997 := 0.0
	for _, d := range density {
		if d > 0 {
			nonZero = append(nonZero, d)
			if d > p997 {
				p997 = d
			}
		}
	}
	if len(nonZero) > 64 {
		sort.Float64s(nonZero)
		pIdx := int(float64(len(nonZero)-1) * 0.997)
		if nonZero[pIdx] > 0 {
			p997 = nonZero[pIdx]
		}
	}
	if p997 <= 0 {
		p997 = 1.0
	}

	rc := c.NewRenderCanvas()
	// Deep cosmic void base so sparse fractal tendrils glow cleanly against dark space.
	rc.FillBackground(color.RGBA{R: 4, G: 3, B: 10, A: 255})
	rc.SetBlendMode(render.BlendAdd)

	const densityGain = 6.5
	logDenom := math.Log1p(p997 * densityGain)
	if logDenom <= 1e-9 {
		logDenom = 1.0
	}

	whR, whG, whB, _ := render.ColorToLinearRGB(whiteHot)

	type sparkCandidate struct {
		x, y int
		val  float64
		col  color.RGBA
	}
	var sparks []sparkCandidate

	for py := 0; py < h; py++ {
		rowOff := py * w
		for px := 0; px < w; px++ {
			idx := rowOff + px
			d := density[idx]
			if d <= 0 {
				continue
			}

			rawRatio := common.Clamp01(math.Log1p(d*densityGain) / logDenom)
			// Gamma > 1 preserves rich dark voids and crisp high-contrast caustic filaments
			dNorm := math.Pow(rawRatio, 1.18)

			invD := 1.0 / d
			rLin := rAcc[idx] * invD
			gLin := gAcc[idx] * invD
			bLin := bAcc[idx] * invD

			// Only the sharpest caustic ridges (> 0.82) warm toward white-hot incandescent gold
			if rawRatio > 0.82 {
				coreT := common.Smoothstep(0.82, 1.0, rawRatio) * 0.58
				rLin = rLin*(1.0-coreT) + whR*coreT
				gLin = gLin*(1.0-coreT) + whG*coreT
				bLin = bLin*(1.0-coreT) + whB*coreT
			}

			hdrBoost := 0.85 + 1.15*math.Pow(rawRatio, 2.2)
			rc.BlendPixelLinear(px, py, rLin*hdrBoost, gLin*hdrBoost, bLin*hdrBoost, dNorm, render.BlendAdd)

			if rawRatio > 0.92 && px > 4 && px < w-4 && py > 4 && py < h-4 {
				if d >= density[idx-1] && d >= density[idx+1] && d >= density[idx-w] && d >= density[idx+w] {
					spCol := render.OklabToRGBA(render.LinearRGBAToOklab(rLin, gLin, bLin), 255)
					sparks = append(sparks, sparkCandidate{x: px, y: py, val: d, col: spCol})
				}
			}
		}
	}

	// Add subtle luminous jewel nodes on top-ranked structural caustic peaks.
	sort.Slice(sparks, func(i, j int) bool {
		return sparks[i].val > sparks[j].val
	})
	chosen := make([]sparkCandidate, 0, 22)
	minDist := 22.0 * scale
	for _, sp := range sparks {
		if len(chosen) >= 22 {
			break
		}
		tooClose := false
		for _, ch := range chosen {
			if math.Hypot(float64(sp.x-ch.x), float64(sp.y-ch.y)) < minDist {
				tooClose = true
				break
			}
		}
		if !tooClose {
			chosen = append(chosen, sp)
		}
	}
	for i, sp := range chosen {
		u := 1.0 - float64(i)/float64(common.MaxInt(len(chosen), 1))
		rad := (1.6 + 2.8*u) * scale
		sparkCol := render.LerpOklab(sp.col, whiteHot, 0.45)
		rc.DrawGlowDot(float64(sp.x)+0.5, float64(sp.y)+0.5, rad, sparkCol, 0.18+0.16*u, 0.20)
	}

	rc.ApplyBloom(0.68, 7.0, 0.28)
	rc.ApplyVignette(0.28)
	c.MergeRenderCanvasToneMapped(rc, 1.08, 0)
}
