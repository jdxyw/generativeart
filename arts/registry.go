package arts

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
)

// ArtDescriptor describes a registered generative art engine, its default visual
// configuration, and a parameter-aware factory capable of reconstructing an Engine
// from either default settings or a decoded PNG RecipeMetadata parameter map.
type ArtDescriptor struct {
	ID         string
	StructName string
	Category   string
	DefaultBG  color.RGBA
	DefaultPal string
	Build      func(params map[string]interface{}) generativeart.Engine
}

var namedPalettes = map[string][]color.RGBA{
	"cyberpunk":   common.Cyberpunk,
	"dunhuang":    common.Dunhuang,
	"ukiyoe":      common.UkiyoE,
	"inferno":     common.Inferno,
	"magma":       common.Magma,
	"viridis":     common.Viridis,
	"plasma":      common.Plasma,
	"oceanic":     common.Oceanic,
	"sunset":      common.Sunset,
	"botanical":   common.Botanical,
	"pasteldream": common.PastelDream,
	"youthful":    common.Youthful,
}

// PaletteByName returns a named color palette and true if found.
func PaletteByName(name string) ([]color.RGBA, bool) {
	p, ok := namedPalettes[strings.ToLower(strings.TrimSpace(name))]
	return p, ok
}

// PaletteNames returns the sorted list of registered palette names.
func PaletteNames() []string {
	out := make([]string, 0, len(namedPalettes))
	for k := range namedPalettes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func paramInt(p map[string]interface{}, key string, def int) int {
	if p == nil {
		return def
	}
	v, ok := p[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return def
	}
}

func paramFloat(p map[string]interface{}, key string, def float64) float64 {
	if p == nil {
		return def
	}
	v, ok := p[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return def
	}
}

func paramString(p map[string]interface{}, key, def string) string {
	if p == nil {
		return def
	}
	if s, ok := p[key].(string); ok && len(s) > 0 {
		return s
	}
	return def
}

var registry = []ArtDescriptor{
	// Wave 18: LIC Jovian Magnetosphere, Thin-Film Iridescent Nacre, Kissing-Schottky Kleinian Pearls & 3D Bismuth Hopper Crystals
	{
		ID: "jupitermagnetosphere", StructName: "jupiterMagnetosphere", Category: "flow",
		DefaultBG: color.RGBA{9, 14, 28, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewJupiterMagnetosphere(paramInt(p, "bands", 9), paramInt(p, "vortices", 16), paramFloat(p, "licLength", 22.0))
		},
	},
	{
		ID: "iridescentnacre", StructName: "iridescentNacre", Category: "bio",
		DefaultBG: color.RGBA{14, 22, 30, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewIridescentNacre(paramInt(p, "whorls", 7), paramInt(p, "fiberCount", 18000), paramFloat(p, "filmThicknessNm", 410.0))
		},
	},
	{
		ID: "schottkykleinian", StructName: "schottkyKleinian", Category: "chaos",
		DefaultBG: color.RGBA{7, 9, 20, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSchottkyKleinian(paramInt(p, "maxDepth", 7), paramFloat(p, "loxodromicTwist", 1.0))
		},
	},
	{
		ID: "bismuthhopper", StructName: "bismuthHopper", Category: "geom",
		DefaultBG: color.RGBA{10, 12, 20, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewBismuthHopper(paramInt(p, "terraces", 14), paramInt(p, "crystals", 9), paramFloat(p, "oxideNm", 420.0))
		},
	},

	// Wave 17: Tactile Bristle & Vortex Impasto Series (Binary Nebula Plumes, Abyssal Kuroshio Currents, Solar Prominence Tondo & Botanical Phyllotaxis Whorls)
	{
		ID: "nebulaplumes", StructName: "nebulaPlumes", Category: "media",
		DefaultBG: color.RGBA{243, 236, 225, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewNebulaPlumes(paramInt(p, "spiralArms", 3), paramInt(p, "strokeCount", 23000), paramFloat(p, "swirlTightness", 1.0))
		},
	},
	{
		ID: "abyssalcurrents", StructName: "abyssalCurrents", Category: "media",
		DefaultBG: color.RGBA{13, 16, 24, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewAbyssalCurrents(paramInt(p, "jetBands", 18), paramInt(p, "strokeCount", 23000), paramFloat(p, "meanderAmp", 1.0))
		},
	},
	{
		ID: "solarprominence", StructName: "solarProminence", Category: "media",
		DefaultBG: color.RGBA{242, 235, 224, 255}, DefaultPal: "inferno",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSolarProminence(paramInt(p, "archCount", 14), paramInt(p, "strokeCount", 24000), paramFloat(p, "magneticTwist", 1.0))
		},
	},
	{
		ID: "botanicalwhorls", StructName: "botanicalWhorls", Category: "media",
		DefaultBG: color.RGBA{238, 235, 226, 255}, DefaultPal: "botanical",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewBotanicalWhorls(paramInt(p, "whorlCount", 64), paramInt(p, "strokeCount", 22000), paramFloat(p, "phylloAngle", 137.507764))
		},
	},

	// Wave 16: Molnár Order-to-Disorder Concentric Morph, Swirling Vortex Impasto, Mid-Century Sliced Pebbles, Hexagonal Mandala Honeycomb, Richter Squeegee Scrape & Retro Sci-Fi Halftone Panels
	{
		ID: "concentricmorph", StructName: "concentricMorph", Category: "geom",
		DefaultBG: color.RGBA{22, 27, 34, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewConcentricMorph(paramInt(p, "gridCols", 18), paramInt(p, "gridRows", 18), paramFloat(p, "disorder", 1.0))
		},
	},
	{
		ID: "vorteximpasto", StructName: "vortexImpasto", Category: "media",
		DefaultBG: color.RGBA{239, 232, 222, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewVortexImpasto(paramInt(p, "vortexCount", 46), paramInt(p, "strokeCount", 22000), paramFloat(p, "curlStrength", 1.0))
		},
	},
	{
		ID: "slicedpebbles", StructName: "slicedPebbles", Category: "geom",
		DefaultBG: color.RGBA{15, 18, 24, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSlicedPebbles(paramInt(p, "gridSize", 6), paramInt(p, "maxCuts", 3), paramFloat(p, "cornerRadius", 18.0))
		},
	},
	{
		ID: "hexmandala", StructName: "hexMandala", Category: "geom",
		DefaultBG: color.RGBA{7, 6, 5, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewHexMandala(paramInt(p, "cols", 13), paramInt(p, "rows", 16), paramFloat(p, "colorScale", 1.0))
		},
	},
	{
		ID: "squeegeescrape", StructName: "squeegeeScrape", Category: "media",
		DefaultBG: color.RGBA{24, 28, 32, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSqueegeeScrape(paramInt(p, "passes", 8), paramInt(p, "dropletCount", 460), paramFloat(p, "scrapePressure", 1.0))
		},
	},
	{
		ID: "halftonepanels", StructName: "halftonePanels", Category: "media",
		DefaultBG: color.RGBA{244, 247, 244, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewHalftonePanels(paramInt(p, "panelCount", 22), paramFloat(p, "dotSpacing", 5.5))
		},
	},

	// Wave 15: Hydraulic Erosion Alpine Peaks, Isometric Basalt Columns, Aeolian Sand Dunes, Bathymetric Trench Chart & 3D Wireframe Glacier Mesh
	{
		ID: "hydraulicerosion", StructName: "hydraulicErosion", Category: "geom",
		DefaultBG: color.RGBA{10, 16, 34, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewHydraulicErosion(paramInt(p, "dropletsN", 32000), paramFloat(p, "erosionStrength", 1.0))
		},
	},
	{
		ID: "basaltcolumns", StructName: "basaltColumns", Category: "geom",
		DefaultBG: color.RGBA{8, 11, 20, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewBasaltColumns(paramInt(p, "hexRings", 16), paramFloat(p, "pillarRelief", 1.0))
		},
	},
	{
		ID: "sanddunes", StructName: "sandDunes", Category: "flow",
		DefaultBG: color.RGBA{12, 14, 32, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSandDunes(paramInt(p, "dunesN", 7), paramFloat(p, "windShear", 1.0))
		},
	},
	{
		ID: "bathymetricchart", StructName: "bathymetricChart", Category: "geom",
		DefaultBG: color.RGBA{5, 10, 28, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewBathymetricChart(paramInt(p, "isobaths", 22), paramInt(p, "seamounts", 7))
		},
	},
	{
		ID: "wireframeterrain", StructName: "wireframeTerrain", Category: "geom",
		DefaultBG: color.RGBA{6, 9, 22, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewWireframeTerrain(paramInt(p, "gridRes", 76), paramFloat(p, "reliefScale", 1.0))
		},
	},

	// Wave 14: Riemann Zeta Critical Strip, 3D Hopf Fibration, Thomas Labyrinth Attractor, 12-Fold Faraday Quasicrystal Waves, Conformal Möbius Loxodrome & FTLE Lagrangian Coherent Structures
	{
		ID: "riemannzeta", StructName: "riemannZeta", Category: "chaos",
		DefaultBG: color.RGBA{6, 9, 18, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewRiemannZeta(paramInt(p, "termsN", 26), paramInt(p, "phaseRays", 16))
		},
	},
	{
		ID: "hopffibration", StructName: "hopfFibration", Category: "geom",
		DefaultBG: color.RGBA{6, 8, 18, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewHopfFibration(paramInt(p, "toriN", 4), paramInt(p, "fibersPerTorus", 24))
		},
	},
	{
		ID: "thomasattractor", StructName: "thomasAttractor", Category: "chaos",
		DefaultBG: color.RGBA{4, 7, 16, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewThomasAttractor(paramInt(p, "pointsN", 900000), paramFloat(p, "dampingB", 0.19))
		},
	},
	{
		ID: "faradaywaves", StructName: "faradayWaves", Category: "flow",
		DefaultBG: color.RGBA{7, 11, 24, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewFaradayWaves(paramInt(p, "symmetry", 12), paramFloat(p, "harmonicScale", 1.0))
		},
	},
	{
		ID: "mobiusloxodrome", StructName: "mobiusLoxodrome", Category: "geom",
		DefaultBG: color.RGBA{7, 9, 18, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMobiusLoxodrome(paramInt(p, "arms", 14), paramInt(p, "spheresPerArm", 44))
		},
	},
	{
		ID: "lagrangiancoherent", StructName: "lagrangianCoherent", Category: "flow",
		DefaultBG: color.RGBA{5, 8, 18, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewLagrangianCoherent(paramInt(p, "advSteps", 18), paramFloat(p, "ftlScale", 1.0))
		},
	},

	// Wave 13: 3D Halvorsen Attractor, Kleinian Fuchsian Pearls, Abelian Sandpile, Kuramoto Phase Defects, Farey Ford Circles & Quantum Chaos Scars (100-Art Milestone)
	{
		ID: "halvorsenattractor", StructName: "halvorsenAttractor", Category: "chaos",
		DefaultBG: color.RGBA{4, 6, 14, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewHalvorsenAttractor(paramInt(p, "pointsN", 1200000), paramFloat(p, "alpha", 1.4))
		},
	},
	{
		ID: "hyperbolicfuchsian", StructName: "hyperbolicFuchsian", Category: "chaos",
		DefaultBG: color.RGBA{7, 9, 18, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewHyperbolicFuchsian(paramInt(p, "maxIter", 36), paramFloat(p, "zoom", 1.0))
		},
	},
	{
		ID: "abeliansandpile", StructName: "abelianSandpile", Category: "geom",
		DefaultBG: color.RGBA{8, 14, 32, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewAbelianSandpile(paramInt(p, "grains", 260000))
		},
	},
	{
		ID: "kuramotowave", StructName: "kuramotoWave", Category: "flow",
		DefaultBG: color.RGBA{8, 11, 22, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewKuramotoWave(paramInt(p, "defectsN", 8), paramInt(p, "streamlines", 180))
		},
	},
	{
		ID: "fordcircles", StructName: "fordCircles", Category: "geom",
		DefaultBG: color.RGBA{8, 10, 20, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewFordCircles(paramInt(p, "maxDenom", 14))
		},
	},
	{
		ID: "quantumbilliard", StructName: "quantumBilliard", Category: "flow",
		DefaultBG: color.RGBA{6, 9, 20, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewQuantumBilliard(paramInt(p, "modesN", 48), paramFloat(p, "waveNumber", 1.0))
		},
	},

	// Wave 12: Dunhuang Caisson, Celtic Knotwork, 3D Gyroid TPMS, Lichtenberg Plasma, Persian Muqarnas & Warring States Bronze
	{
		ID: "dunhuangapsara", StructName: "dunhuangApsara", Category: "media",
		DefaultBG: color.RGBA{20, 14, 18, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewDunhuangApsara(paramInt(p, "ribbons", 8), paramInt(p, "haloRings", 5))
		},
	},
	{
		ID: "celticknotwork", StructName: "celticKnotwork", Category: "geom",
		DefaultBG: color.RGBA{14, 18, 26, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCelticKnotwork(paramInt(p, "gridSize", 8), paramFloat(p, "ribbonWidth", 1.0))
		},
	},
	{
		ID: "gyroidsurface", StructName: "gyroidSurface", Category: "geom",
		DefaultBG: color.RGBA{10, 14, 26, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewGyroidSurface(paramFloat(p, "scaleFreq", 4.2), paramFloat(p, "thickness", 0.085))
		},
	},
	{
		ID: "lichtenbergtree", StructName: "lichtenbergTree", Category: "bio",
		DefaultBG: color.RGBA{6, 8, 18, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewLichtenbergTree(paramInt(p, "branches", 9), paramFloat(p, "roughness", 1.0))
		},
	},
	{
		ID: "persiantilework", StructName: "persianTilework", Category: "geom",
		DefaultBG: color.RGBA{16, 12, 24, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPersianTilework(paramInt(p, "symmetry", 16), paramInt(p, "rings", 6))
		},
	},
	{
		ID: "champleveenamel", StructName: "champleveEnamel", Category: "media",
		DefaultBG: color.RGBA{12, 14, 18, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewChampleveEnamel(paramInt(p, "symmetry", 8), paramInt(p, "ringCount", 4))
		},
	},

	// Wave 11: Jovian Vortex Fluid, BZ Spiral Waves, Gilded PCB WFC, Cursive Calligraphy, Buddhabrot & Lyapunov
	{
		ID: "vortexfluid", StructName: "vortexFluid", Category: "flow",
		DefaultBG: color.RGBA{12, 22, 46, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewVortexFluid(paramInt(p, "vorticesN", 9), paramInt(p, "steps", 18))
		},
	},
	{
		ID: "cyclicautomata", StructName: "cyclicAutomata", Category: "bio",
		DefaultBG: color.RGBA{10, 16, 34, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCyclicAutomata(paramInt(p, "states", 14), paramInt(p, "steps", 120))
		},
	},
	{
		ID: "circuitwfc", StructName: "circuitWFC", Category: "geom",
		DefaultBG: color.RGBA{8, 20, 26, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCircuitWFC(paramInt(p, "gridSize", 28), paramInt(p, "busCount", 64))
		},
	},
	{
		ID: "asemiccalligraphy", StructName: "asemicCalligraphy", Category: "media",
		DefaultBG: color.RGBA{244, 236, 214, 255}, DefaultPal: "ukiyoe",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewAsemicCalligraphy(paramInt(p, "cols", 6), paramInt(p, "glyphsPerCol", 5))
		},
	},
	{
		ID: "buddhabrot", StructName: "buddhabrot", Category: "chaos",
		DefaultBG: color.RGBA{4, 6, 14, 255}, DefaultPal: "inferno",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewBuddhabrot(paramInt(p, "samplesN", 450000), paramInt(p, "maxIter", 360))
		},
	},
	{
		ID: "lyapunovfractal", StructName: "lyapunovFractal", Category: "chaos",
		DefaultBG: color.RGBA{8, 12, 26, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewLyapunovFractal(paramString(p, "seq", "AABAB"), paramInt(p, "maxIter", 84))
		},
	},

	// Wave 10: 8-Fold Quasicrystals, Sashiko Embroidery, Electromagnetic Dipoles & Complex Newton Fractals
	{
		ID: "ammannbeenker", StructName: "ammannBeenker", Category: "geom",
		DefaultBG: color.RGBA{14, 14, 22, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewAmmannBeenker(paramInt(p, "depth", 5))
		},
	},
	{
		ID: "sashikoweave", StructName: "sashikoWeave", Category: "media",
		DefaultBG: color.RGBA{16, 28, 58, 255}, DefaultPal: "ukiyoe",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSashikoWeave(paramInt(p, "gridSize", 3))
		},
	},
	{
		ID: "magneticdipole", StructName: "magneticDipole", Category: "flow",
		DefaultBG: color.RGBA{8, 11, 22, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMagneticDipole(paramInt(p, "polesN", 6), paramInt(p, "linesPerPole", 28))
		},
	},
	{
		ID: "newtonbasin", StructName: "newtonBasin", Category: "chaos",
		DefaultBG: color.RGBA{10, 12, 22, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewNewtonBasin(paramInt(p, "degree", 6), paramInt(p, "maxIter", 48))
		},
	},

	// Wave 9: Topological Contours, Streamline Engraving, 3D Origami & Bauhaus SDF Relief
	{
		ID: "evenstreamlines", StructName: "evenStreamlines", Category: "flow",
		DefaultBG: color.RGBA{12, 16, 28, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewEvenStreamlines(paramFloat(p, "dSep", 7.5), paramFloat(p, "dTest", 3.9))
		},
	},
	{
		ID: "marchingisolines", StructName: "marchingIsolines", Category: "geom",
		DefaultBG: color.RGBA{12, 22, 42, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMarchingIsolines(paramInt(p, "levels", 16), paramFloat(p, "noiseScale", 1.0))
		},
	},
	{
		ID: "origamitessellation", StructName: "origamiTessellation", Category: "geom",
		DefaultBG: color.RGBA{16, 22, 44, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewOrigamiTessellation(paramInt(p, "cols", 14), paramInt(p, "rows", 14), paramFloat(p, "foldDepth", 1.0))
		},
	},
	{
		ID: "mondriansubdivision", StructName: "mondrianSubdivision", Category: "geom",
		DefaultBG: color.RGBA{18, 18, 24, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMondrianSubdivision(paramInt(p, "depth", 4), paramFloat(p, "pad", 6.0))
		},
	},

	// Wave 8: Decorative Arts, Cloisonné Enamel & Printmaking
	{
		ID: "cloisonneenamel", StructName: "cloisonneEnamel", Category: "media",
		DefaultBG: color.RGBA{18, 14, 22, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCloisonneEnamel(paramInt(p, "rings", 6), paramInt(p, "petals", 12))
		},
	},
	{
		ID: "woodblockwave", StructName: "woodblockWave", Category: "media",
		DefaultBG: color.RGBA{242, 232, 206, 255}, DefaultPal: "ukiyoe",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewWoodblockWave(paramInt(p, "ridgeLayers", 5), paramInt(p, "waveLines", 42))
		},
	},
	{
		ID: "artdecosunburst", StructName: "artDecoSunburst", Category: "geom",
		DefaultBG: color.RGBA{14, 16, 24, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewArtDecoSunburst(paramInt(p, "rayCount", 36), paramInt(p, "archTiers", 7))
		},
	},
	{
		ID: "stainedrosewindow", StructName: "stainedRoseWindow", Category: "geom",
		DefaultBG: color.RGBA{12, 11, 18, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewStainedRoseWindow(paramInt(p, "petals", 12), paramInt(p, "rings", 5))
		},
	},

	// Wave 7: Anisotropic Flow Mosaic & Low-Poly Crystal Series
	{
		ID: "mosaicvortex", StructName: "mosaicVortex", Category: "media",
		DefaultBG: color.RGBA{16, 20, 38, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMosaicVortex(paramInt(p, "ringSteps", 34), paramFloat(p, "brickScale", 1.0))
		},
	},
	{
		ID: "mosaicwave", StructName: "mosaicWave", Category: "media",
		DefaultBG: color.RGBA{18, 24, 44, 255}, DefaultPal: "ukiyoe",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMosaicWave(paramInt(p, "waveCourses", 32), paramFloat(p, "brickScale", 1.0))
		},
	},
	{
		ID: "mosaiccanyon", StructName: "mosaicCanyon", Category: "media",
		DefaultBG: color.RGBA{22, 18, 32, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMosaicCanyon(paramInt(p, "strataSteps", 32), paramFloat(p, "brickScale", 1.0))
		},
	},
	{
		ID: "mosaiclotus", StructName: "mosaicLotus", Category: "media",
		DefaultBG: color.RGBA{16, 24, 42, 255}, DefaultPal: "botanical",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMosaicLotus(paramInt(p, "petals", 12), paramFloat(p, "brickScale", 1.0))
		},
	},
	{
		ID: "mosaicsun", StructName: "mosaicSun", Category: "media",
		DefaultBG: color.RGBA{22, 26, 48, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMosaicSun(paramInt(p, "ringSteps", 32), paramFloat(p, "brickScale", 1.0))
		},
	},

	// Wave 6: G-Buffer Impasto, Voxel Terrain & Geometric Replications
	{
		ID: "impastoflow", StructName: "impastoFlow", Category: "media",
		DefaultBG: color.RGBA{8, 7, 16, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewImpastoFlow(paramInt(p, "strokeCount", 3200), paramInt(p, "steps", 520), paramFloat(p, "fieldScale", 0.0032))
		},
	},
	{
		ID: "voxelterrain", StructName: "voxelTerrain", Category: "geom",
		DefaultBG: color.RGBA{10, 12, 26, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewVoxelTerrain(paramInt(p, "gridRes", 340), paramFloat(p, "heightScale", 2.6), paramFloat(p, "contrast", 1.0))
		},
	},
	{
		ID: "packedcapsules", StructName: "packedCapsules", Category: "flow",
		DefaultBG: color.RGBA{238, 236, 229, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPackedCapsules(paramInt(p, "capsuleCount", 2200), paramFloat(p, "maxScale", 7.2), paramFloat(p, "minScale", 3.6))
		},
	},
	{
		ID: "orphicdiscs", StructName: "orphicDiscs", Category: "geom",
		DefaultBG: color.RGBA{72, 178, 102, 255}, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewOrphicDiscs(paramInt(p, "discCount", 34))
		},
	},
	{
		ID: "nautilusspiral", StructName: "nautilusSpiral", Category: "bio",
		DefaultBG: color.RGBA{18, 24, 34, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewNautilusSpiral(paramInt(p, "turns", 5), paramInt(p, "chambersPerTurn", 68))
		},
	},
	{
		ID: "tangentcones", StructName: "tangentCones", Category: "geom",
		DefaultBG: color.RGBA{234, 88, 56, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewTangentCones(paramInt(p, "coneCount", 26))
		},
	},

	// Wave 1–5
	{
		ID: "poincaredisk", StructName: "poincareDisk", Category: "geom",
		DefaultBG: color.RGBA{8, 10, 20, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPoincareDisk(paramInt(p, "p", 7), paramInt(p, "q", 3), paramInt(p, "depth", 5))
		},
	},
	{
		ID: "suminagashi", StructName: "suminagashi", Category: "media",
		DefaultBG: color.RGBA{244, 239, 229, 255}, DefaultPal: "ukiyoe",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSuminagashi(paramInt(p, "dropCount", 48), paramInt(p, "combPasses", 7), paramFloat(p, "swirlStrength", 1.0))
		},
	},
	{
		ID: "aizawaattractor", StructName: "aizawaAttractor", Category: "chaos",
		DefaultBG: color.RGBA{5, 7, 15, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewAizawaAttractor(paramInt(p, "iterations", 900000), paramFloat(p, "dofStrength", 1.0))
		},
	},
	{
		ID: "penrosetiling", StructName: "penroseTiling", Category: "geom",
		DefaultBG: color.RGBA{10, 14, 24, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPenroseTiling(paramInt(p, "subdivisions", 6))
		},
	},
	{
		ID: "isometriccity", StructName: "isometricCity", Category: "geom",
		DefaultBG: color.RGBA{18, 20, 34, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewIsometricCity(paramInt(p, "gridSize", 11), paramFloat(p, "density", 1.0))
		},
	},
	{
		ID: "dlafrost", StructName: "dlaFrost", Category: "bio",
		DefaultBG: color.RGBA{4, 8, 18, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewDLAFrost(paramInt(p, "particleCount", 18000), paramFloat(p, "stickiness", 0.55))
		},
	},
	{
		ID: "curlflow", StructName: "curlFlow", Category: "flow",
		DefaultBG: color.RGBA{14, 18, 28, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCurlFlow(paramInt(p, "lineNum", 110), paramInt(p, "stepNum", 36), paramFloat(p, "noiseScale", 0.0045))
		},
	},
	{
		ID: "sandring", StructName: "sandRing", Category: "media",
		DefaultBG: color.RGBA{10, 10, 18, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSandRing(paramInt(p, "ringNum", 22), paramInt(p, "grainsPerStep", 22), paramFloat(p, "noiseScale", 1.3))
		},
	},
	{
		ID: "watercolor", StructName: "watercolor", Category: "media",
		DefaultBG: color.RGBA{246, 242, 234, 255}, DefaultPal: "botanical",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewWatercolor(paramInt(p, "shapeNum", 9), paramInt(p, "layers", 22), paramInt(p, "subdivisions", 5), paramFloat(p, "roughness", 0.32))
		},
	},
	{
		ID: "guilloche", StructName: "guilloche", Category: "geom",
		DefaultBG: color.RGBA{8, 12, 22, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewGuilloche(paramInt(p, "curveNum", 28), paramFloat(p, "majorR", 175), paramFloat(p, "minorR", 65), paramFloat(p, "penOffset", 115))
		},
	},
	{
		ID: "cliffordnebula", StructName: "cliffordNebula", Category: "chaos",
		DefaultBG: color.RGBA{6, 6, 14, 255}, DefaultPal: "inferno",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCliffordNebula(paramInt(p, "iterations", 1500000), paramFloat(p, "a", -1.7), paramFloat(p, "b", 1.8), paramFloat(p, "c", -1.9), paramFloat(p, "d", -0.4))
		},
	},
	{
		ID: "physarum", StructName: "physarum", Category: "bio",
		DefaultBG: color.RGBA{6, 10, 18, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPhysarum(paramInt(p, "agentNum", 45000), paramInt(p, "steps", 110), paramFloat(p, "sensorAngle", 0), paramFloat(p, "turnAngle", 0))
		},
	},
	{
		ID: "differentialgrowth", StructName: "differentialGrowth", Category: "bio",
		DefaultBG: color.RGBA{10, 12, 22, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewDifferentialGrowth(paramInt(p, "steps", 220), paramInt(p, "maxNodes", 1600), paramFloat(p, "minDist", 3.5), paramFloat(p, "maxDist", 7.5), paramFloat(p, "repulsionRadius", 18.0))
		},
	},
	{
		ID: "inklandscape", StructName: "inkLandscape", Category: "media",
		DefaultBG: color.RGBA{244, 239, 228, 255}, DefaultPal: "ukiyoe",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewInkLandscape(paramInt(p, "layerNum", 8), paramFloat(p, "roughness", 0.45))
		},
	},
	{
		ID: "voronoicrystal", StructName: "voronoiCrystal", Category: "geom",
		DefaultBG: color.RGBA{24, 22, 30, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewVoronoiCrystal(paramInt(p, "cellNum", 85), paramInt(p, "relaxSteps", 4))
		},
	},
	{
		ID: "grayscott", StructName: "grayScott", Category: "bio",
		DefaultBG: color.RGBA{8, 12, 20, 255}, DefaultPal: "viridis",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewGrayScott(paramInt(p, "steps", 3200), paramFloat(p, "feed", 0.037), paramFloat(p, "kill", 0.060))
		},
	},
	{
		ID: "lissajousribbon", StructName: "lissajousRibbon", Category: "chaos",
		DefaultBG: color.RGBA{6, 8, 16, 255}, DefaultPal: "plasma",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewLissajousRibbon(paramInt(p, "curves", 140), paramFloat(p, "freqX", 3.0), paramFloat(p, "freqY", 4.0), paramFloat(p, "freqZ", 5.0))
		},
	},
	{
		ID: "apolloniangasket", StructName: "apollonianGasket", Category: "geom",
		DefaultBG: color.RGBA{10, 14, 24, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewApollonianGasket(paramInt(p, "maxDepth", 7), paramFloat(p, "minRadius", 2.5))
		},
	},
	{
		ID: "spacecolonization", StructName: "spaceColonization", Category: "bio",
		DefaultBG: color.RGBA{8, 14, 18, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSpaceColonization(paramInt(p, "attractorNum", 1800), paramInt(p, "steps", 160), paramFloat(p, "stepDist", 6.0), paramFloat(p, "killDist", 55.0))
		},
	},
	{
		ID: "truchetflow", StructName: "truchetFlow", Category: "geom",
		DefaultBG: color.RGBA{245, 240, 230, 255}, DefaultPal: "ukiyoe",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewTruchetFlow(paramInt(p, "baseGrid", 6), paramInt(p, "maxDepth", 3))
		},
	},
	{
		ID: "aurorawave", StructName: "auroraWave", Category: "flow",
		DefaultBG: color.RGBA{6, 10, 22, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewAuroraWave(paramInt(p, "curtainNum", 7), paramInt(p, "ribbonSteps", 160), paramFloat(p, "turbulence", 1.0))
		},
	},
	{
		ID: "fractalflame", StructName: "fractalFlame", Category: "chaos",
		DefaultBG: color.RGBA{4, 4, 10, 255}, DefaultPal: "magma",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewFractalFlame(paramInt(p, "iterations", 1200000), paramInt(p, "symmetry", 3))
		},
	},
	{
		ID: "joydivision", StructName: "joyDivision", Category: "flow",
		DefaultBG: color.RGBA{10, 12, 20, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewJoyDivision(paramInt(p, "lineNum", 58), paramFloat(p, "amplitude", 46.0), paramFloat(p, "frequency", 0.022))
		},
	},
	{
		ID: "boidsflock", StructName: "boidsFlock", Category: "bio",
		DefaultBG: color.RGBA{10, 14, 24, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewBoidsFlock(paramInt(p, "boidNum", 260), paramInt(p, "steps", 140), paramFloat(p, "perceptionRadius", 48.0))
		},
	},
	{
		ID: "islamicstar", StructName: "islamicStar", Category: "geom",
		DefaultBG: color.RGBA{12, 18, 32, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewIslamicStar(paramInt(p, "gridSize", 6), paramFloat(p, "contactAngle", 60.0), paramFloat(p, "strapWidth", 4.5))
		},
	},
	{
		ID: "delaunaystained", StructName: "delaunayStained", Category: "geom",
		DefaultBG: color.RGBA{16, 14, 24, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewDelaunayStained(paramInt(p, "siteNum", 130), paramFloat(p, "jitter", 0.65))
		},
	},
	{
		ID: "phyllotaxis", StructName: "phyllotaxis", Category: "bio",
		DefaultBG: color.RGBA{10, 14, 20, 255}, DefaultPal: "botanical",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPhyllotaxis(paramInt(p, "seedCount", 950), paramFloat(p, "divergenceAngle", 137.507764), paramFloat(p, "scaleFactor", 8.2))
		},
	},
}

// ListDescriptors returns a copy of all registered ArtDescriptors.
func ListDescriptors() []ArtDescriptor {
	out := make([]ArtDescriptor, len(registry))
	copy(out, registry)
	return out
}

// Lookup finds an ArtDescriptor by its CLI ID (e.g. "cloisonneenamel") or its Go struct
// name recorded in PNG RecipeMetadata (e.g. "cloisonneEnamel").
func Lookup(nameOrStruct string) (ArtDescriptor, bool) {
	q := strings.ToLower(strings.TrimSpace(nameOrStruct))
	for _, d := range registry {
		if d.ID == q || strings.ToLower(d.StructName) == q {
			return d, true
		}
	}
	return ArtDescriptor{}, false
}

// ReplayRecipe reconstructs and renders the artwork described by meta at the original
// resolution (or overrideW x overrideH if both > 0).
func ReplayRecipe(meta *generativeart.RecipeMetadata, overrideW, overrideH int) (*generativeart.Canva, error) {
	if meta == nil {
		return nil, fmt.Errorf("nil RecipeMetadata")
	}
	desc, ok := Lookup(meta.Engine)
	if !ok {
		return nil, fmt.Errorf("engine %q is not registered for automatic replay", meta.Engine)
	}
	w, h := meta.Width, meta.Height
	if overrideW > 0 && overrideH > 0 {
		w, h = overrideW, overrideH
	}
	if w <= 0 || h <= 0 {
		w, h = 600, 600
	}

	c := generativeart.NewCanva(w, h)
	c.SetSeed(meta.Seed)
	c.SetBackground(color.RGBA{
		R: meta.Background[0],
		G: meta.Background[1],
		B: meta.Background[2],
		A: meta.Background[3],
	})
	if pal, exists := PaletteByName(desc.DefaultPal); exists {
		c.SetColorSchema(pal)
	}
	c.FillBackground()
	eng := desc.Build(meta.Params)
	c.Draw(eng)
	return c, nil
}

// ReplayPNG reads the embedded GenerativeArt-Recipe tEXt metadata from pngPath and
// re-renders the exact artwork at overrideW x overrideH (or original size if 0, 0).
func ReplayPNG(pngPath string, overrideW, overrideH int) (*generativeart.Canva, *generativeart.RecipeMetadata, error) {
	meta, err := generativeart.ReadPNGRecipe(pngPath)
	if err != nil {
		return nil, nil, err
	}
	c, err := ReplayRecipe(meta, overrideW, overrideH)
	if err != nil {
		return nil, meta, err
	}
	return c, meta, nil
}
