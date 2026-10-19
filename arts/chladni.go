package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type chladni struct {
	particlesN int
	m, n, a, b float64
}

// NewChladni returns a chladni engine that simulates acoustic cymatic standing-wave
// resonance plates with sub-pixel HDR grain accumulation and Oklab harmonic fields.
func NewChladni(particlesN int, m, n, a, b float64) *chladni {
	if particlesN <= 0 {
		particlesN = 90000
	}
	if m <= 0 {
		m = 7
	}
	if n <= 0 {
		n = 3
	}
	if a <= 0 {
		a = 1.0
	}
	if b <= 0 {
		b = -1.0
	}
	return &chladni{
		particlesN: particlesN,
		m:          m,
		n:          n,
		a:          a,
		b:          b,
	}
}

// waveAndGrad evaluates the coupled Chladni standing-wave amplitude W(u, v)
// and its analytical partial derivatives (dW/du, dW/dv) on u, v in [-1, 1].
func (ch *chladni) waveAndGrad(u, v float64) (w, dwdu, dwdv float64) {
	pi := math.Pi
	m1, n1 := ch.m, ch.n
	m2, n2 := ch.m+2.0, ch.n+1.0

	// Primary mode L1(u, v) = a*cos(n*pi*u)*cos(m*pi*v) + b*cos(m*pi*u)*cos(n*pi*v)
	cn1u, sn1u := math.Cos(n1*pi*u), math.Sin(n1*pi*u)
	cm1v, sm1v := math.Cos(m1*pi*v), math.Sin(m1*pi*v)
	cm1u, sm1u := math.Cos(m1*pi*u), math.Sin(m1*pi*u)
	cn1v, sn1v := math.Cos(n1*pi*v), math.Sin(n1*pi*v)

	l1 := ch.a*cn1u*cm1v + ch.b*cm1u*cn1v
	dl1du := -ch.a*n1*pi*sn1u*cm1v - ch.b*m1*pi*sm1u*cn1v
	dl1dv := -ch.a*m1*pi*cn1u*sm1v - ch.b*n1*pi*cm1u*sn1v

	// Secondary harmonic mode L2 with (m+2, n+1) weighted by 0.25
	cn2u, sn2u := math.Cos(n2*pi*u), math.Sin(n2*pi*u)
	cm2v, sm2v := math.Cos(m2*pi*v), math.Sin(m2*pi*v)
	cm2u, sm2u := math.Cos(m2*pi*u), math.Sin(m2*pi*u)
	cn2v, sn2v := math.Cos(n2*pi*v), math.Sin(n2*pi*v)

	l2 := ch.a*cn2u*cm2v + ch.b*cm2u*cn2v
	dl2du := -ch.a*n2*pi*sn2u*cm2v - ch.b*m2*pi*sm2u*cn2v
	dl2dv := -ch.a*m2*pi*cn2u*sm2v - ch.b*n2*pi*cm2u*sn2v

	w = l1 + 0.25*l2
	dwdu = dl1du + 0.25*dl2du
	dwdv = dl1dv + 0.25*dl2dv
	return w, dwdu, dwdv
}

// Generative renders the Chladni cymatic resonance plate into c.
func (ch *chladni) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	width, height := c.Width(), c.Height()
	if width <= 0 || height <= 0 {
		return
	}
	fw, fh := float64(width), float64(height)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Oceanic
	}
	grad := render.NewMultiStopGradient(schema...)

	deepSapphire := color.RGBA{R: 5, G: 10, B: 28, A: 255}
	indigoShimmer := render.LerpOklab(color.RGBA{R: 16, G: 28, B: 78, A: 255}, grad.SampleOklab(0.25), 0.55)
	cyanGlow := render.LerpOklab(color.RGBA{R: 48, G: 210, B: 235, A: 255}, grad.SampleOklab(0.72), 0.5)
	goldGlow := color.RGBA{R: 255, G: 215, B: 125, A: 255}

	maxAmp := math.Max(1e-6, (math.Abs(ch.a)+math.Abs(ch.b))*1.25)

	// 1. Render subtle harmonic energy field across the canvas in linear RGB.
	for y := 0; y < height; y++ {
		v := (float64(y)+0.5)/fh*2.0 - 1.0
		for x := 0; x < width; x++ {
			u := (float64(x)+0.5)/fw*2.0 - 1.0
			wVal, _, _ := ch.waveAndGrad(u, v)
			normW := common.Constrain(math.Abs(wVal)/maxAmp, 0, 1)

			// Antinodes shimmer in deep sapphire/indigo; nodal trenches glow with cyan/gold undertone.
			nodalProximity := math.Exp(-14.0 * normW * normW)
			antinodeEnergy := math.Pow(normW, 0.75)

			baseCol := render.LerpOklab(deepSapphire, indigoShimmer, antinodeEnergy*0.68)
			nodalTone := render.LerpOklab(cyanGlow, goldGlow, 0.35+0.35*math.Cos(math.Pi*(u*u+v*v)))
			pixelCol := render.LerpOklab(baseCol, nodalTone, nodalProximity*0.32)

			rLin, gLin, bLin, _ := render.ColorToLinearRGB(pixelCol)
			rc.SetPixelLinear(x, y, rLin*0.42, gLin*0.42, bLin*0.42, 1.0)
		}
	}

	// 2. Simulate stochastic cymatic resonance grains migrating toward nodal lines |W(u, v)| == 0.
	rng := rand.New(rand.NewSource(rand.Int63()))
	rc.SetBlendMode(render.BlendAdd)

	const descendSteps = 18
	for i := 0; i < ch.particlesN; i++ {
		u := rng.Float64()*2.0 - 1.0
		v := rng.Float64()*2.0 - 1.0

		var wVal, dwdu, dwdv float64
		for step := 0; step < descendSteps; step++ {
			wVal, dwdu, dwdv = ch.waveAndGrad(u, v)
			absW := math.Abs(wVal)

			// Gradient of E(u, v) = 0.5 * W(u, v)^2 is W * (dW/du, dW/dv).
			// Normalize direction and step proportionally to |W(u, v)| plus acoustic vibration jitter.
			gradMag := math.Hypot(dwdu, dwdv)
			stepScale := (absW / maxAmp) * 0.085
			if gradMag > 1e-9 {
				dirU := (wVal * dwdu) / (absW*gradMag + 1e-12)
				dirV := (wVal * dwdv) / (absW*gradMag + 1e-12)
				u -= dirU * stepScale
				v -= dirV * stepScale
			}

			// Stochastic Chladni plate excitation proportional to local wave amplitude.
			vib := (0.0012 + 0.024*(absW/maxAmp)) * (1.0 - 0.35*float64(step)/float64(descendSteps))
			u += rng.NormFloat64() * vib
			v += rng.NormFloat64() * vib

			// Reflective acoustic boundary on [-1, 1]^2
			if u < -1.0 {
				u = -2.0 - u
			} else if u > 1.0 {
				u = 2.0 - u
			}
			if v < -1.0 {
				v = -2.0 - v
			} else if v > 1.0 {
				v = 2.0 - v
			}
		}

		wVal, _, _ = ch.waveAndGrad(u, v)
		residual := common.Constrain(math.Abs(wVal)/maxAmp, 0, 1)

		px := (u + 1.0) * 0.5 * (fw - 1.0)
		py := (v + 1.0) * 0.5 * (fh - 1.0)

		radial := common.Constrain(math.Hypot(u, v)/math.Sqrt2, 0, 1)
		harmonicT := common.Constrain(0.5+0.5*math.Sin(ch.n*math.Pi*u)*math.Cos(ch.m*math.Pi*v), 0, 1)
		palCol := grad.SampleOklab(common.Constrain(0.15+0.65*radial+0.2*harmonicT, 0, 1))

		// Grains tightly settled on nodal lines shine warm golden-white; outer halo grains shimmer cyan/palette.
		nodalSharpness := math.Exp(-18.0 * residual * residual)
		grainCol := render.LerpOklab(palCol, goldGlow, nodalSharpness*0.72)
		alpha := (0.014 + 0.036*nodalSharpness) * (0.75 + 0.25*rng.Float64())

		rc.SplatPoint(px, py, grainCol, alpha)
	}

	rc.ApplyBloom(0.5, 8.0, 0.55)
	rc.ApplyVignette(0.35)
	c.MergeRenderCanvasToneMapped(rc, 1.1, 0)
}
