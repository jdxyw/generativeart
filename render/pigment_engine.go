package render

import (
	"math"

	"github.com/jdxyw/generativeart/common"
)

// PigmentKM represents a physical artist pigment in the two-flux Kubelka-Munk
// optical model with 3-band spectral absorption K, scattering S, and mineral
// granulation weight (sedimentation into cold-pressed paper valleys).
type PigmentKM struct {
	Name        string
	K           [3]float64 // Spectral absorption coefficients (R, G, B)
	S           [3]float64 // Spectral scattering coefficients (R, G, B)
	Granulation float64    // Mineral sedimentation tendency in [0, 1]
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
	return PigmentKM{
		Name:        name,
		K:           kArr,
		S:           sArr,
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
