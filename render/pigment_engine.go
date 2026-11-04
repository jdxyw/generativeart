package render

import (
	"math"

	"github.com/jdxyw/generativeart/common"
)

// PigmentKM represents a physical artist pigment in the two-flux Kubelka-Munk
// optical model with both 3-band RGB (K, S) and 8-band visible spectrum (K8, S8,
// spanning 405nm..685nm) absorption/scattering coefficients, plus mineral
// granulation weight (sedimentation into cold-pressed paper valleys).
type PigmentKM struct {
	Name        string
	K           [3]float64 // 3-band spectral absorption coefficients (R, G, B)
	S           [3]float64 // 3-band spectral scattering coefficients (R, G, B)
	K8          [8]float64 // 8-band visible spectrum absorption coefficients (405..685nm)
	S8          [8]float64 // 8-band visible spectrum scattering coefficients (405..685nm)
	Granulation float64    // Mineral sedimentation tendency in [0, 1]
}

// Spectral8BasisRGB defines a smooth partition-of-unity spectral uplift basis across
// 8 visible wavelengths {405, 445, 485, 525, 565, 605, 645, 685} nm for (R, G, B).
// Each column sums to 1.0 so neutral grays map to flat reflectance spectra.
var spectral8BasisRGB = [8][3]float64{
	{0.00, 0.02, 0.98}, // 405nm (Violet-Blue)
	{0.00, 0.12, 0.88}, // 445nm (Royal Blue)
	{0.02, 0.48, 0.50}, // 485nm (Cyan-Turquoise overlap)
	{0.06, 0.84, 0.10}, // 525nm (Emerald Green)
	{0.34, 0.64, 0.02}, // 565nm (Yellow-Green overlap)
	{0.78, 0.22, 0.00}, // 605nm (Amber-Orange)
	{0.96, 0.04, 0.00}, // 645nm (Vermilion Red)
	{0.99, 0.01, 0.00}, // 685nm (Deep Crimson)
}

// LinearRGBToSpectrum8 uplifts linear RGB reflectance (rLin, gLin, bLin) in [0, 1]^3
// to an 8-band visible reflectance spectrum R(lambda_0..7).
func LinearRGBToSpectrum8(rLin, gLin, bLin float64) [8]float64 {
	var spec [8]float64
	for i := 0; i < 8; i++ {
		b := spectral8BasisRGB[i]
		spec[i] = common.Constrain(b[0]*rLin+b[1]*gLin+b[2]*bLin, 0.004, 0.994)
	}
	return spec
}

// Spectrum8ToLinearRGB integrates an 8-band visible reflectance spectrum R(lambda_0..7)
// into linear sRGB under D65 illumination, calibrated as the exact left-inverse of
// LinearRGBToSpectrum8 so pure colors and neutral grays round-trip losslessly.
func Spectrum8ToLinearRGB(spec [8]float64) [3]float64 {
	// Precomputed Moore-Penrose pseudoinverse (B^T B)^{-1} B^T of spectral8BasisRGB
	const (
		r0, r1, r2, r3, r4, r5, r6, r7 = 0.0273933268, 0.0088295946, -0.0504372262, -0.1021414320, 0.0408626428, 0.2852078485, 0.3866861017, 0.4035991439
		g0, g1, g2, g3, g4, g5, g6, g7 = -0.1690867538, -0.0675977882, 0.2984595541, 0.6652139623, 0.4719949546, 0.0610767502, -0.1153297942, -0.1447308850
		b0, b1, b2, b3, b4, b5, b6, b7 = 0.5300188013, 0.4566102964, 0.1820677901, -0.1027466048, -0.0997360351, -0.0174018632, 0.0222864484, 0.0289011670
	)
	rLin := r0*spec[0] + r1*spec[1] + r2*spec[2] + r3*spec[3] + r4*spec[4] + r5*spec[5] + r6*spec[6] + r7*spec[7]
	gLin := g0*spec[0] + g1*spec[1] + g2*spec[2] + g3*spec[3] + g4*spec[4] + g5*spec[5] + g6*spec[6] + g7*spec[7]
	bLin := b0*spec[0] + b1*spec[1] + b2*spec[2] + b3*spec[3] + b4*spec[4] + b5*spec[5] + b6*spec[6] + b7*spec[7]
	return [3]float64{
		common.Clamp01(rLin),
		common.Clamp01(gLin),
		common.Clamp01(bLin),
	}
}

// PigmentWeight pairs a PigmentKM with its relative concentration weight.
type PigmentWeight struct {
	Pigment PigmentKM
	Weight  float64
}

// PW is a shorthand constructor for PigmentWeight.
func PW(p PigmentKM, weight float64) PigmentWeight {
	return PigmentWeight{Pigment: p, Weight: weight}
}

func srgbByteToLinearDirect(b uint8) float64 {
	v := float64(b) / 255.0
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// NewPigmentFromSRGB constructs a physical Kubelka-Munk pigment from an sRGB hex
// masstone color (r, g, b), scattering opacity in [0.15, 2.5], and granulation in [0, 1]
// using the exact inverse Kubelka-Munk remission equation: K/S = (1 - R_inf)^2 / (2 * R_inf).
func NewPigmentFromSRGB(name string, r, g, b uint8, scattering, granulation float64) PigmentKM {
	rLin := srgbByteToLinearDirect(r)
	gLin := srgbByteToLinearDirect(g)
	bLin := srgbByteToLinearDirect(b)
	return NewPigmentFromLinearRGB(name, rLin, gLin, bLin, scattering, granulation)
}

// NewPigmentFromLinearRGB constructs a physical Kubelka-Munk pigment from linear RGB
// masstone reflectance (rLin, gLin, bLin).
func NewPigmentFromLinearRGB(name string, rLin, gLin, bLin, scattering, granulation float64) PigmentKM {
	if scattering < 0.08 {
		scattering = 0.08
	}
	rgb := [3]float64{
		common.Constrain(rLin, 0.006, 0.992),
		common.Constrain(gLin, 0.006, 0.992),
		common.Constrain(bLin, 0.006, 0.992),
	}
	var kArr, sArr [3]float64
	for c := 0; c < 3; c++ {
		rem := rgb[c]
		kOverS := ((1.0 - rem) * (1.0 - rem)) / (2.0 * rem)
		sArr[c] = scattering
		kArr[c] = kOverS * scattering
	}
	spec8 := LinearRGBToSpectrum8(rgb[0], rgb[1], rgb[2])
	var k8Arr, s8Arr [8]float64
	for i := 0; i < 8; i++ {
		rem := spec8[i]
		kOverS := ((1.0 - rem) * (1.0 - rem)) / (2.0 * rem)
		s8Arr[i] = scattering
		k8Arr[i] = kOverS * scattering
	}
	return PigmentKM{
		Name:        name,
		K:           kArr,
		S:           sArr,
		K8:          k8Arr,
		S8:          s8Arr,
		Granulation: common.Clamp01(granulation),
	}
}

// Curated Museum Mineral, Gouache & Watercolor Pigment Catalog
var (
	PigmentUltramarine      = NewPigmentFromSRGB("French Ultramarine", 0x18, 0x52, 0xb8, 0.75, 0.88)
	PigmentAzurite          = NewPigmentFromSRGB("Song Azurite (Shiqing)", 0x1a, 0x4c, 0x98, 0.85, 0.85)
	PigmentCerulean         = NewPigmentFromSRGB("Moebius Cerulean", 0x2c, 0x86, 0xb8, 0.95, 0.76)
	PigmentCobaltTeal       = NewPigmentFromSRGB("Cobalt Turquoise", 0x28, 0xc4, 0xb8, 1.05, 0.72)
	PigmentMalachite        = NewPigmentFromSRGB("Song Malachite (Shilv)", 0x24, 0xa2, 0x7a, 0.90, 0.84)
	PigmentCadmiumYellow    = NewPigmentFromSRGB("Cadmium Lemon Gold", 0xfc, 0xe0, 0x42, 1.35, 0.25)
	PigmentYellowOchre      = NewPigmentFromSRGB("Roman Yellow Ochre", 0xe8, 0xae, 0x48, 1.15, 0.68)
	PigmentRawUmber         = NewPigmentFromSRGB("Piranesi Raw Umber", 0x58, 0x3c, 0x28, 0.78, 0.78)
	PigmentBurntSienna      = NewPigmentFromSRGB("Terracotta Burnt Sienna", 0xc8, 0x4e, 0x2c, 0.82, 0.74)
	PigmentVermilion        = NewPigmentFromSRGB("Cinnabar Vermilion (Zhusha)", 0xe2, 0x3c, 0x26, 1.25, 0.65)
	PigmentQuinacridoneRose = NewPigmentFromSRGB("Quinacridone Rose-Coral", 0xec, 0x4c, 0x6a, 0.55, 0.18)
	PigmentImperialPurple   = NewPigmentFromSRGB("Tyrian Imperial Violet", 0x52, 0x32, 0x7c, 0.60, 0.62)
	PigmentTitaniumWhite    = NewPigmentFromSRGB("Titanium Gouache White", 0xfc, 0xf8, 0xee, 2.40, 0.20)
	PigmentGoldLeaf         = NewPigmentFromSRGB("Burnished 24k Mud-Gold", 0xf8, 0xc8, 0x50, 1.85, 0.45)
	PigmentPineSootInk      = NewPigmentFromSRGB("Songyan Pine-Soot Ink", 0x12, 0x10, 0x16, 0.90, 0.55)
)

// MixKubelkaMunk computes the physical subtractive reflectance R_inf in linear RGB
// for an arbitrary mixture of pigments using the two-flux Kubelka-Munk equation:
//
//	R_inf = 1 + (K/S) - sqrt((K/S)^2 + 2*(K/S))
func MixKubelkaMunk(items ...PigmentWeight) [3]float64 {
	var kSum, sSum [3]float64
	totalW := 0.0
	for _, it := range items {
		if it.Weight <= 0 {
			continue
		}
		totalW += it.Weight
		for c := 0; c < 3; c++ {
			kSum[c] += it.Weight * it.Pigment.K[c]
			sSum[c] += it.Weight * it.Pigment.S[c]
		}
	}
	if totalW <= 1e-9 {
		return [3]float64{1, 1, 1}
	}
	var out [3]float64
	for c := 0; c < 3; c++ {
		s := math.Max(sSum[c], 1e-6)
		ks := kSum[c] / s
		rem := 1.0 + ks - math.Sqrt(ks*ks+2.0*ks)
		out[c] = common.Clamp01(rem)
	}
	return out
}

// LerpKubelkaMunk physically mixes pigment A and pigment B at fraction t in [0, 1]
// via subtractive Kubelka-Munk absorption/scattering rather than additive RGB lerp.
func LerpKubelkaMunk(a, b PigmentKM, t float64) [3]float64 {
	t = common.Clamp01(t)
	return MixKubelkaMunk(PW(a, 1.0-t), PW(b, t))
}

// GlazeKubelkaMunk evaluates a translucent watercolor/mineral glaze layer of optical
// thickness d >= 0 over an existing linear-RGB substrate (substrateRGB) using the
// exact finite-thickness hyperbolic Kubelka-Munk reflectance equation:
//
//	R = ((1 - Rg*a)*sinh(b*S*d) + Rg*b*cosh(b*S*d)) / ((a - Rg)*sinh(b*S*d) + b*cosh(b*S*d))
func GlazeKubelkaMunk(substrateRGB [3]float64, pigment PigmentKM, thickness float64) [3]float64 {
	if thickness <= 1e-6 {
		return substrateRGB
	}
	var out [3]float64
	for c := 0; c < 3; c++ {
		rg := common.Clamp01(substrateRGB[c])
		s := math.Max(pigment.S[c], 1e-5)
		k := math.Max(pigment.K[c], 1e-6)
		a := 1.0 + k/s
		b := math.Sqrt(math.Max(1e-9, a*a-1.0))
		bsd := common.Constrain(b*s*thickness, 0.0, 18.0)
		sh := math.Sinh(bsd)
		ch := math.Cosh(bsd)
		num := (1.0-rg*a)*sh + rg*b*ch
		den := (a-rg)*sh + b*ch
		if math.Abs(den) < 1e-9 {
			out[c] = rg
		} else {
			out[c] = common.Clamp01(num / den)
		}
	}
	return out
}

// WatercolorWashConfig controls capillary edge darkening (coffee-ring effect),
// cold-pressed paper granulation, and wet-on-wet feathering.
type WatercolorWashConfig struct {
	BaseThickness     float64 // Base optical glaze thickness (default 0.85)
	EdgeDarkening     float64 // Coffee-ring pigment rim concentration (default 0.75)
	EdgeWidth         float64 // Width of coffee-ring rim in normalized units (default 0.045)
	GranulationAmount float64 // Paper-valley mineral settling strength (default 0.65)
	PaperFreq         float64 // Spatial frequency of cold-pressed paper grain (default 48.0)
	WetFeather        float64 // Capillary boundary feathering width (default 0.015)
}

// DefaultWatercolorWashConfig returns balanced parameters for mineral & watercolor washes.
func DefaultWatercolorWashConfig() WatercolorWashConfig {
	return WatercolorWashConfig{
		BaseThickness:     0.85,
		EdgeDarkening:     0.78,
		EdgeWidth:         0.045,
		GranulationAmount: 0.68,
		PaperFreq:         48.0,
		WetFeather:        0.016,
	}
}

// EvaluateWatercolorWash composites a physical watercolor/mineral wash over substrateRGB
// inside a region defined by signed distance sdfDist (negative inside the wash),
// simulating capillary wet-edge backruns, coffee-ring edge darkening, and cold-pressed
// paper-valley mineral granulation.
func EvaluateWatercolorWash(
	ne *common.NoiseEngine,
	ux, uy, sdfDist float64,
	pigment PigmentKM,
	substrateRGB [3]float64,
	cfg WatercolorWashConfig,
) [3]float64 {
	if ne == nil {
		return substrateRGB
	}
	feather := math.Max(cfg.WetFeather, 1e-4)
	// Organic capillary flow perturbation along the wash boundary
	capillaryWarp := 0.018 * ne.FBM2D(ux*9.5+1.3, uy*9.5-2.1, 3, 2.0, 0.5)
	effDist := sdfDist + capillaryWarp
	if effDist >= feather {
		return substrateRGB
	}

	coverage := 1.0 - common.Smoothstep(-feather, feather, effDist)
	depthInside := math.Max(0.0, -effDist)

	// 1. Coffee-Ring Edge Darkening: evaporating water carries pigment particles
	// toward the wet perimeter (depthInside in [0, cfg.EdgeWidth]), creating a crisp
	// darkened mineral ring at the edge of the wash!
	edgeW := math.Max(cfg.EdgeWidth, 0.008)
	ringBell := math.Exp(-math.Pow((depthInside-0.35*edgeW)/(0.48*edgeW), 2.0))
	coffeeRing := 1.0 + cfg.EdgeDarkening*ringBell

	// 2. Cold-Pressed Paper-Valley Mineral Granulation:
	// Heavy mineral pigments (high pigment.Granulation) settle into paper valleys!
	pFreq := cfg.PaperFreq
	if pFreq <= 4.0 {
		pFreq = 48.0
	}
	paperGrain := 0.5 + 0.5*ne.FBM2D(ux*pFreq, uy*pFreq, 3, 2.05, 0.52)
	macroFlow := 0.85 + 0.30*ne.FBM2D(ux*3.8-0.7, uy*3.8+1.4, 3, 2.0, 0.5)
	granEffect := 1.0 + cfg.GranulationAmount*pigment.Granulation*(paperGrain-0.46)*1.35

	optThickness := cfg.BaseThickness * coverage * coffeeRing * granEffect * macroFlow
	return GlazeKubelkaMunk(substrateRGB, pigment, optThickness)
}

// ============================================================================
// Phase 5: 8-Band Visible Spectrum Kubelka-Munk & Wet-on-Wet Fluid Bleed
// ============================================================================

func ensurePigmentSpectral8(p *PigmentKM) {
	if p.S8[0] > 0 || p.S8[3] > 0 {
		return
	}
	// Derive 8-band K8/S8 from 3-band K/S if a custom struct literal was passed
	var remRGB [3]float64
	meanS := 0.0
	for c := 0; c < 3; c++ {
		s := math.Max(p.S[c], 0.1)
		meanS += s
		ks := math.Max(0, p.K[c]) / s
		remRGB[c] = common.Constrain(1.0+ks-math.Sqrt(ks*ks+2.0*ks), 0.004, 0.994)
	}
	meanS /= 3.0
	spec8 := LinearRGBToSpectrum8(remRGB[0], remRGB[1], remRGB[2])
	for i := 0; i < 8; i++ {
		rem := spec8[i]
		kOverS := ((1.0 - rem) * (1.0 - rem)) / (2.0 * rem)
		p.S8[i] = meanS
		p.K8[i] = kOverS * meanS
	}
}

// MixSpectralKubelkaMunk computes the 8-band visible spectrum Kubelka-Munk reflectance
// across {405..685}nm and integrates it to linear RGB, preserving intermediate spectral
// overlap bands (such as emerald/cyan at 485-565nm when mixing blue and yellow pigments).
func MixSpectralKubelkaMunk(items ...PigmentWeight) [3]float64 {
	var kSum, sSum [8]float64
	totalW := 0.0
	for _, it := range items {
		if it.Weight <= 0 {
			continue
		}
		pig := it.Pigment
		ensurePigmentSpectral8(&pig)
		totalW += it.Weight
		for b := 0; b < 8; b++ {
			kSum[b] += it.Weight * pig.K8[b]
			sSum[b] += it.Weight * pig.S8[b]
		}
	}
	if totalW <= 1e-9 {
		return [3]float64{1, 1, 1}
	}
	var specOut [8]float64
	for b := 0; b < 8; b++ {
		s := math.Max(sSum[b], 1e-6)
		ks := kSum[b] / s
		rem := 1.0 + ks - math.Sqrt(ks*ks+2.0*ks)
		specOut[b] = common.Clamp01(rem)
	}
	return Spectrum8ToLinearRGB(specOut)
}

// LerpSpectralKubelkaMunk mixes pigment A and pigment B at fraction t in [0, 1] using 8-band spectral Kubelka-Munk.
func LerpSpectralKubelkaMunk(a, b PigmentKM, t float64) [3]float64 {
	t = common.Clamp01(t)
	return MixSpectralKubelkaMunk(PW(a, 1.0-t), PW(b, t))
}

// GlazeSpectralKubelkaMunk evaluates a finite-thickness hyperbolic Kubelka-Munk glaze across
// all 8 visible wavelength bands (405..685nm) over substrateRGB and integrates back to linear RGB.
func GlazeSpectralKubelkaMunk(substrateRGB [3]float64, pigment PigmentKM, thickness float64) [3]float64 {
	if thickness <= 1e-6 {
		return substrateRGB
	}
	ensurePigmentSpectral8(&pigment)
	subSpec := LinearRGBToSpectrum8(substrateRGB[0], substrateRGB[1], substrateRGB[2])
	var outSpec [8]float64
	for i := 0; i < 8; i++ {
		rg := subSpec[i]
		s := math.Max(pigment.S8[i], 1e-5)
		k := math.Max(pigment.K8[i], 1e-6)
		a := 1.0 + k/s
		b := math.Sqrt(math.Max(1e-9, a*a-1.0))
		bsd := common.Constrain(b*s*thickness, 0.0, 18.0)
		sh := math.Sinh(bsd)
		ch := math.Cosh(bsd)
		num := (1.0-rg*a)*sh + rg*b*ch
		den := (a-rg)*sh + b*ch
		if math.Abs(den) < 1e-9 {
			outSpec[i] = rg
		} else {
			outSpec[i] = common.Clamp01(num / den)
		}
	}
	return Spectrum8ToLinearRGB(outSpec)
}

// WetWatercolorDeposit defines an initial wet pigment wash region for SimulateWetWatercolorBleed.
type WetWatercolorDeposit struct {
	Pigment   PigmentKM
	SDF       SDF2D   // Negative inside the wet pigment region (in pixel coordinates [0, W] x [0, H])
	Wetness   float64 // Initial water saturation in [0.2, 2.0] (default 1.0)
	Thickness float64 // Initial pigment optical thickness (default 0.95)
}

// WetWatercolorConfig configures the 2D capillary water-pigment advection-diffusion solver.
type WetWatercolorConfig struct {
	Width             int
	Height            int
	Steps             int        // Number of capillary fluid diffusion & drying steps (default 28)
	WaterDiffusivity  float64    // Capillary water spreading rate across paper fibers (default 0.22)
	PigmentMobility   float64    // Pigment diffusion & coffee-ring outward advection rate (default 0.18)
	CoffeeRingOutflow float64    // Strength of outward capillary flow toward drying edges (default 0.36)
	GranulationStrength float64  // Paper-valley sedimentation strength for heavy mineral pigments (default 0.72)
	PaperColor        [3]float64 // Linear RGB substrate color of the archival paper
	Seed              int64
}

// DefaultWetWatercolorConfig returns a balanced configuration for 2D wet-on-wet watercolor simulation.
func DefaultWetWatercolorConfig(w, h int, seed int64) WetWatercolorConfig {
	return WetWatercolorConfig{
		Width:               w,
		Height:              h,
		Steps:               28,
		WaterDiffusivity:    0.22,
		PigmentMobility:     0.18,
		CoffeeRingOutflow:   0.36,
		GranulationStrength: 0.72,
		PaperColor:          [3]float64{0.93, 0.89, 0.80},
		Seed:                seed,
	}
}

// SimulateWetWatercolorBleed simulates physical wet-on-wet capillary water diffusion,
// outward coffee-ring pigment advection along drying boundaries, multi-pigment cross-bleeding,
// and cold-pressed paper valley granulation across a w x h grid, returning the resolved
// linear-RGB buffer of length Width * Height * 3.
func SimulateWetWatercolorBleed(cfg WetWatercolorConfig, deposits []WetWatercolorDeposit) []float64 {
	w, h := cfg.Width, cfg.Height
	if w <= 0 || h <= 0 {
		return nil
	}
	nPix := w * h
	outRGB := make([]float64, nPix*3)
	if len(deposits) == 0 {
		for i := 0; i < nPix; i++ {
			outRGB[i*3+0] = cfg.PaperColor[0]
			outRGB[i*3+1] = cfg.PaperColor[1]
			outRGB[i*3+2] = cfg.PaperColor[2]
		}
		return outRGB
	}

	steps := cfg.Steps
	if steps <= 0 {
		steps = 24
	}
	waterDiff := common.Constrain(cfg.WaterDiffusivity, 0.02, 0.45)
	pigMob := common.Constrain(cfg.PigmentMobility, 0.02, 0.40)
	outflow := common.Constrain(cfg.CoffeeRingOutflow, 0.0, 1.2)
	granStr := common.Constrain(cfg.GranulationStrength, 0.0, 1.5)

	ne := common.NewNoiseEngine(cfg.Seed)
	paperHeight := make([]float64, nPix)
	permeability := make([]float64, nPix)
	fw, fh := float64(w), float64(h)

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			uy := float64(y) / math.Max(1.0, fh)
			rOff := y * w
			for x := 0; x < w; x++ {
				ux := float64(x) / math.Max(1.0, fw)
				grain := 0.5 + 0.5*ne.FBM2D(ux*42.0, uy*42.0, 3, 2.05, 0.52)
				fiber := 0.5 + 0.5*ne.FBM2D(ux*9.0+1.7, uy*9.0-2.3, 2, 2.0, 0.5)
				ph := common.Clamp01(0.68*grain + 0.32*fiber)
				paperHeight[rOff+x] = ph
				// Paper valleys (lower ph) have higher capillary water permeability
				permeability[rOff+x] = 0.55 + 0.90*(1.0-ph)
			}
		}
	})

	numP := len(deposits)
	water := make([]float64, nPix)
	nextWater := make([]float64, nPix)
	conc := make([][]float64, numP)
	nextConc := make([][]float64, numP)
	mobScale := make([]float64, numP)

	for k := 0; k < numP; k++ {
		conc[k] = make([]float64, nPix)
		nextConc[k] = make([]float64, nPix)
		// Finer organic pigments (low Granulation) diffuse faster than heavy mineral pigments
		mobScale[k] = 1.15 - 0.45*common.Clamp01(deposits[k].Pigment.Granulation)
	}

	// Seed initial water saturation and pigment concentrations from SDFs
	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y) + 0.5
			rOff := y * w
			for x := 0; x < w; x++ {
				fx := float64(x) + 0.5
				pt := Pt(fx, fy)
				idx := rOff + x
				for k, dep := range deposits {
					if dep.SDF == nil {
						continue
					}
					d := dep.SDF(pt)
					if d >= 2.5 {
						continue
					}
					wInit := dep.Wetness
					if wInit <= 0 {
						wInit = 1.0
					}
					tInit := dep.Thickness
					if tInit <= 0 {
						tInit = 0.95
					}
					cov := 1.0 - common.Smoothstep(-2.0, 2.0, d)
					if wInit*cov > water[idx] {
						water[idx] = wInit * cov
					}
					conc[k][idx] += tInit * cov
				}
			}
		}
	})

	// Integrate coupled capillary water diffusion + outward coffee-ring pigment advection
	evapRate := 0.35 / float64(steps)
	for step := 0; step < steps; step++ {
		parallelRows(w, h, func(y0, y1 int) {
			for y := y0; y < y1; y++ {
				yU := y - 1
				if yU < 0 {
					yU = 0
				}
				yD := y + 1
				if yD >= h {
					yD = h - 1
				}
				rC := y * w
				rU := yU * w
				rD := yD * w

				for x := 0; x < w; x++ {
					xL := x - 1
					if xL < 0 {
						xL = 0
					}
					xR := x + 1
					if xR >= w {
						xR = w - 1
					}
					idx := rC + x

					wC := water[idx]
					wL := water[rC+xL]
					wR := water[rC+xR]
					wU := water[rU+x]
					wD := water[rD+x]

					perm := permeability[idx]
					lapW := 0.25*(wL+wR+wU+wD) - wC
					// Evaporate faster near dry boundaries (driving outward capillary coffee-ring flow)
					edgeDry := math.Max(0.0, wC-0.25*(wL+wR+wU+wD))
					wNext := math.Max(0.0, wC+waterDiff*perm*lapW-evapRate*(0.4+1.8*edgeDry))
					nextWater[idx] = wNext

					wetFactor := common.Smoothstep(0.02, 0.35, math.Max(wC, 0.25*(wL+wR+wU+wD)))
					if wetFactor <= 1e-5 {
						for k := 0; k < numP; k++ {
							nextConc[k][idx] = conc[k][idx]
						}
						continue
					}

					// Outward capillary rim boost where water gradient is steep (drying perimeter)
					gradW := math.Hypot(wR-wL, wD-wU)
					rimAccum := 1.0 + outflow*gradW*0.35

					for k := 0; k < numP; k++ {
						ck := conc[k]
						cC := ck[idx]
						cAvg := 0.20*(ck[rC+xL]+ck[rC+xR]+ck[rU+x]+ck[rD+x]) +
							0.05*(ck[rU+xL]+ck[rU+xR]+ck[rD+xL]+ck[rD+xR])
						diff := pigMob * mobScale[k] * wetFactor * perm * (cAvg - cC)
						nextConc[k][idx] = math.Max(0.0, (cC+diff)*rimAccum)
					}
				}
			}
		})
		water, nextWater = nextWater, water
		for k := 0; k < numP; k++ {
			conc[k], nextConc[k] = nextConc[k], conc[k]
		}
	}

	// Resolve final 8-band spectral Kubelka-Munk pigment glazes + paper valley granulation
	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			rOff := y * w
			for x := 0; x < w; x++ {
				idx := rOff + x
				valley := 1.0 - paperHeight[idx]
				curRGB := cfg.PaperColor
				for k := 0; k < numP; k++ {
					cVal := conc[k][idx]
					if cVal <= 1e-4 {
						continue
					}
					gran := deposits[k].Pigment.Granulation
					granMod := 1.0 + granStr*gran*(valley-0.46)*1.45
					curRGB = GlazeSpectralKubelkaMunk(curRGB, deposits[k].Pigment, cVal* math.Max(0.15, granMod))
				}
				outRGB[idx*3+0] = curRGB[0]
				outRGB[idx*3+1] = curRGB[1]
				outRGB[idx*3+2] = curRGB[2]
			}
		}
	})

	return outRGB
}

// RenderWetWatercolorBleed runs SimulateWetWatercolorBleed and composites the resulting
// 64-bit linear HDR pigment wash directly onto c.
func (c *Canvas) RenderWetWatercolorBleed(cfg WetWatercolorConfig, deposits []WetWatercolorDeposit, alpha float64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || alpha <= 0 {
		return
	}
	cfg.Width = c.Width
	cfg.Height = c.Height
	buf := SimulateWetWatercolorBleed(cfg, deposits)
	if len(buf) < c.Width*c.Height*3 {
		return
	}
	effA := common.Clamp01(alpha)
	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			rOff := y * c.Width
			for x := 0; x < c.Width; x++ {
				pIdx := (rOff + x) * 3
				c.BlendPixelLinear(x, y, buf[pIdx+0], buf[pIdx+1], buf[pIdx+2], effA, c.blendMode)
			}
		}
	})
}
