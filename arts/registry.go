package arts

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

// ParamSpec describes a tunable scalar parameter exposed by a registered ArtDescriptor.
type ParamSpec struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Default interface{} `json:"default"`
}

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

// DefaultParams returns the map of default scalar parameter values for this ArtDescriptor.
func (d ArtDescriptor) DefaultParams() map[string]interface{} {
	if d.Build == nil {
		return map[string]interface{}{}
	}
	return generativeart.InspectEngineParams(d.Build(nil))
}

// SupportsAnimation reports whether this ArtDescriptor's Engine implements AnimatedEngine.
func (d ArtDescriptor) SupportsAnimation() bool {
	if d.Build == nil {
		return false
	}
	_, ok := d.Build(nil).(generativeart.AnimatedEngine)
	return ok
}

// ParamSpecs returns the ordered list of tunable scalar parameters on this ArtDescriptor.
func (d ArtDescriptor) ParamSpecs() []ParamSpec {
	if d.Build == nil {
		return nil
	}
	eng := d.Build(nil)
	if eng == nil {
		return nil
	}
	t := reflect.TypeOf(eng)
	v := reflect.ValueOf(eng)
	if t.Kind() == reflect.Ptr && !v.IsNil() {
		t = t.Elem()
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	var specs []ParamSpec
	for i := 0; i < v.NumField(); i++ {
		fType := t.Field(i)
		fVal := v.Field(i)
		switch fVal.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			specs = append(specs, ParamSpec{Name: fType.Name, Type: "int", Default: fVal.Int()})
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			specs = append(specs, ParamSpec{Name: fType.Name, Type: "uint", Default: fVal.Uint()})
		case reflect.Float32, reflect.Float64:
			specs = append(specs, ParamSpec{Name: fType.Name, Type: "float64", Default: fVal.Float()})
		case reflect.Bool:
			specs = append(specs, ParamSpec{Name: fType.Name, Type: "bool", Default: fVal.Bool()})
		case reflect.String:
			specs = append(specs, ParamSpec{Name: fType.Name, Type: "string", Default: fVal.String()})
		}
	}
	return specs
}

// ResolveParams validates and merges user-supplied parameter overrides (matching keys case-insensitively)
// onto the descriptor's default parameters. Returns an error if an unknown parameter key is provided.
func (d ArtDescriptor) ResolveParams(overrides map[string]interface{}) (map[string]interface{}, error) {
	defaults := d.DefaultParams()
	out := make(map[string]interface{}, len(defaults))
	canonKey := make(map[string]string, len(defaults))
	for k, v := range defaults {
		out[k] = v
		canonKey[strings.ToLower(k)] = k
	}
	for k, v := range overrides {
		ck, exists := canonKey[strings.ToLower(strings.TrimSpace(k))]
		if !exists {
			var valid []string
			for _, sp := range d.ParamSpecs() {
				valid = append(valid, fmt.Sprintf("%s (%s=%v)", sp.Name, sp.Type, sp.Default))
			}
			return nil, fmt.Errorf("unknown parameter %q for engine %q (valid parameters: %s)", k, d.ID, strings.Join(valid, ", "))
		}
		out[ck] = v
	}
	return out, nil
}

var categoryHarmoniousPalettes = map[string][]string{
	"geom":  {"dunhuang", "sunset", "ukiyoe", "oceanic", "cyberpunk", "botanical"},
	"media": {"dunhuang", "sunset", "ukiyoe", "botanical", "oceanic", "pasteldream"},
	"chaos": {"inferno", "plasma", "magma", "cyberpunk", "dunhuang", "viridis"},
	"flow":  {"oceanic", "cyberpunk", "dunhuang", "sunset", "viridis", "plasma"},
	"bio":   {"botanical", "oceanic", "dunhuang", "viridis", "sunset", "ukiyoe"},
}

// RandomizePalette picks either the descriptor's tuned default palette or a category-harmonious palette.
func (d ArtDescriptor) RandomizePalette(rng *rand.Rand) string {
	if rng == nil {
		return d.DefaultPal
	}
	if rng.Float64() < 0.45 {
		return d.DefaultPal
	}
	cands, ok := categoryHarmoniousPalettes[d.Category]
	if !ok || len(cands) == 0 {
		cands = PaletteNames()
	}
	return cands[rng.Intn(len(cands))]
}

// RandomizeParams generates a mathematically valid, aesthetically balanced random parameter map
// for this ArtDescriptor around its golden default range.
func (d ArtDescriptor) RandomizeParams(rng *rand.Rand) map[string]interface{} {
	specs := d.ParamSpecs()
	if len(specs) == 0 {
		return nil
	}
	if rng == nil {
		return d.DefaultParams()
	}
	out := make(map[string]interface{}, len(specs))

	switch d.ID {
	case "eschercirclelimit":
		triples := [][3]int{{6, 4, 2}, {7, 3, 2}, {5, 4, 2}, {6, 3, 2}}
		pick := triples[rng.Intn(len(triples))]
		out["p"], out["q"], out["r"] = pick[0], pick[1], pick[2]
		return out
	case "escherdrostegallery":
		out["branches"] = 1 + rng.Intn(2)
		out["archesPerRing"] = 6 + rng.Intn(5)*2
		out["scaleRatio"] = 16.0
		return out
	case "weingartentorusknot":
		knots := [][2]int{{2, 3}, {2, 5}, {3, 4}}
		pick := knots[rng.Intn(len(knots))]
		out["pWind"] = pick[0]
		out["qWind"] = pick[1]
		out["tubeScale"] = math.Round((0.88+0.24*rng.Float64())*100) / 100
		return out
	case "lyapunovfractal":
		seqs := []string{"AABAB", "AB", "AABB", "ABBB", "AABBB"}
		out["seq"] = seqs[rng.Intn(len(seqs))]
		out["maxIter"] = 80
		return out
	}

	for _, sp := range specs {
		switch sp.Type {
		case "int", "uint":
			def := 1
			if v, ok := sp.Default.(int64); ok {
				def = int(v)
			} else if v, ok := sp.Default.(uint64); ok {
				def = int(v)
			}
			if def <= 1 {
				out[sp.Name] = def
				continue
			}
			if def >= 20000 {
				out[sp.Name] = int(float64(def) * (0.65 + 0.45*rng.Float64()))
			} else {
				minV := int(math.Max(2, math.Round(float64(def)*0.75)))
				maxV := int(math.Max(float64(minV+1), math.Round(float64(def)*1.32)))
				out[sp.Name] = minV + rng.Intn(maxV-minV+1)
			}
		case "float64":
			def, _ := sp.Default.(float64)
			if math.Abs(def) < 1e-6 {
				out[sp.Name] = 0.0
				continue
			}
			if strings.Contains(strings.ToLower(sp.Name), "angle") && def > 100 {
				out[sp.Name] = math.Round((137.507764+(rng.Float64()-0.5)*0.6)*1000) / 1000
			} else {
				factor := 0.78 + 0.44*rng.Float64()
				out[sp.Name] = math.Round(def*factor*1000) / 1000
			}
		default:
			out[sp.Name] = sp.Default
		}
	}
	return out
}

// NewParamKeyframeEngine wraps any ArtDescriptor into a seamless looping AnimatedEngine.
// If d already implements AnimatedEngine and len(keyframes) == 0, it delegates directly to d's native
// GenerativeFrame. Otherwise, it smoothly interpolates d's scalar parameters across keyframes
// (or along a closed cosine-eased harmonic parameter orbit when <=1 keyframe is supplied) over phase in [0, 1).
func NewParamKeyframeEngine(d ArtDescriptor, keyframes ...map[string]interface{}) generativeart.AnimatedEngine {
	baseParams := d.DefaultParams()
	if len(keyframes) > 0 && keyframes[0] != nil {
		for k, v := range keyframes[0] {
			baseParams[k] = v
		}
	}

	// If no explicit multi-keyframe trajectory was passed and the engine natively supports animation, use native GenerativeFrame
	if len(keyframes) <= 1 {
		if nativeAnim, ok := d.Build(baseParams).(generativeart.AnimatedEngine); ok {
			return nativeAnim
		}
	}

	// Build closed periodic ring of keyframes
	ring := keyframes
	if len(ring) <= 1 {
		// Automatically construct a smooth 2-keyframe harmonic excursion (+18% parameter orbit)
		kf0 := make(map[string]interface{}, len(baseParams))
		kf1 := make(map[string]interface{}, len(baseParams))
		for k, v := range baseParams {
			kf0[k] = v
			switch val := v.(type) {
			case float64:
				if math.Abs(val) > 1e-5 {
					kf1[k] = val * 1.18
				} else {
					kf1[k] = 0.20
				}
			case int64:
				kf1[k] = val
			case int:
				kf1[k] = val
			default:
				kf1[k] = v
			}
		}
		ring = []map[string]interface{}{kf0, kf1}
	}

	return generativeart.NewAnimatedEngine(func(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
		nK := len(ring)
		u := math.Mod(phase, 1.0)
		if u < 0 {
			u += 1.0
		}
		scaled := u * float64(nK)
		idx0 := int(math.Floor(scaled)) % nK
		idx1 := (idx0 + 1) % nK
		localT := scaled - math.Floor(scaled)
		// Smooth C1 cosine easing between keyframes
		smoothT := 0.5 - 0.5*math.Cos(localT*math.Pi)

		k0 := ring[idx0]
		k1 := ring[idx1]
		interp := make(map[string]interface{}, len(baseParams))
		for k, defVal := range baseParams {
			v0, ok0 := k0[k]
			if !ok0 {
				v0 = defVal
			}
			v1, ok1 := k1[k]
			if !ok1 {
				v1 = v0
			}
			f0, isF0 := toFloat64Param(v0)
			f1, isF1 := toFloat64Param(v1)
			if isF0 && isF1 {
				val := f0 + (f1-f0)*smoothT
				switch defVal.(type) {
				case int, int64:
					interp[k] = int(math.Round(val))
				default:
					interp[k] = val
				}
			} else {
				interp[k] = v0
			}
		}
		eng := d.Build(interp)
		if anim, ok := eng.(generativeart.AnimatedEngine); ok {
			anim.GenerativeFrame(c, frameIdx, totalFrames, phase)
		} else {
			eng.Generative(c)
		}
	})
}

func toFloat64Param(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

// ParseParamOverrides parses a CLI parameter override string such as "galleries=8,tiers=6,shaftScale=1.25"
// or a JSON object string '{"galleries":8,"shaftScale":1.25}' into a parameter map.
func ParseParamOverrides(expr string) (map[string]interface{}, error) {
	s := strings.TrimSpace(expr)
	if s == "" {
		return nil, nil
	}
	if strings.HasPrefix(s, "{") {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return nil, fmt.Errorf("invalid JSON parameter overrides: %w", err)
		}
		return m, nil
	}
	out := make(map[string]interface{})
	normalized := strings.ReplaceAll(s, ";", ",")
	for _, token := range strings.Split(normalized, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		eq := strings.IndexByte(token, '=')
		if eq <= 0 {
			eq = strings.IndexByte(token, ':')
		}
		if eq <= 0 {
			return nil, fmt.Errorf("expected key=value parameter pair, got %q", token)
		}
		k := strings.TrimSpace(token[:eq])
		valStr := strings.TrimSpace(token[eq+1:])
		if iVal, err := strconv.ParseInt(valStr, 10, 64); err == nil {
			out[k] = iVal
		} else if fVal, err := strconv.ParseFloat(valStr, 64); err == nil {
			out[k] = fVal
		} else if bVal, err := strconv.ParseBool(valStr); err == nil {
			out[k] = bVal
		} else {
			out[k] = strings.Trim(valStr, `"'`)
		}
	}
	return out, nil
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

func lookupParamValue(p map[string]interface{}, key string) (interface{}, bool) {
	if p == nil {
		return nil, false
	}
	if v, ok := p[key]; ok {
		return v, true
	}
	for k, v := range p {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}

func paramInt(p map[string]interface{}, key string, def int) int {
	v, ok := lookupParamValue(p, key)
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case uint64:
		return int(n)
	case float64:
		return int(math.Round(n))
	case float32:
		return int(math.Round(float64(n)))
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(n), 64); err == nil {
			return int(math.Round(f))
		}
		return def
	default:
		return def
	}
}

func paramFloat(p map[string]interface{}, key string, def float64) float64 {
	v, ok := lookupParamValue(p, key)
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
	case uint64:
		return float64(n)
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(n), 64); err == nil {
			return f
		}
		return def
	default:
		return def
	}
}

func paramString(p map[string]interface{}, key, def string) string {
	v, ok := lookupParamValue(p, key)
	if !ok {
		return def
	}
	if s, ok := v.(string); ok && len(s) > 0 {
		return s
	}
	return def
}

func paramBool(p map[string]interface{}, key string, def bool) bool {
	v, ok := lookupParamValue(p, key)
	if !ok {
		return def
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		if parsed, err := strconv.ParseBool(strings.TrimSpace(b)); err == nil {
			return parsed
		}
		return def
	default:
		return def
	}
}

var registry = []ArtDescriptor{
	// Wave 27: Multi-Paradigm Flagship Trio (AgateGeodeCrystal, FabergeGuillocheOrnament, AuroraCurtainFold)
	{
		ID: "agategeodecrystal", StructName: "agateGeodeCrystal", Category: "geom",
		DefaultBG: color.RGBA{7, 9, 18, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewAgateGeodeCrystal(paramInt(p, "bands", 20), paramInt(p, "druzyRings", 3), paramFloat(p, "crystalScale", 1.0))
		},
	},
	{
		ID: "fabergeguillocheornament", StructName: "fabergeGuillocheOrnament", Category: "geom",
		DefaultBG: color.RGBA{6, 8, 20, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewFabergeGuillocheOrnament(paramInt(p, "petals", 12), paramInt(p, "guillocheWaves", 28), paramFloat(p, "enamelDepth", 1.0))
		},
	},
	{
		ID: "auroracurtainfold", StructName: "auroraCurtainFold", Category: "flow",
		DefaultBG: color.RGBA{3, 6, 18, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewAuroraCurtainFold(paramInt(p, "curtains", 4), paramInt(p, "rays", 180), paramFloat(p, "foldCurl", 1.15))
		},
	},

	// Wave 26: Multi-Paradigm Flagship Quartet (FidenzaVelvetLoom, KuroshioBioluminescence, LorentzStrangeRibbon, CoronalMagneticArcade)
	{
		ID: "fidenzavelvetloom", StructName: "fidenzaVelvetLoom", Category: "geom",
		DefaultBG: color.RGBA{7, 11, 22, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewFidenzaVelvetLoom(paramInt(p, "beams", 18), paramInt(p, "octaves", 5), paramFloat(p, "turbulence", 2.2))
		},
	},
	{
		ID: "kuroshiobioluminescence", StructName: "kuroshioBioluminescence", Category: "bio",
		DefaultBG: color.RGBA{4, 10, 28, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewKuroshioBioluminescence(paramInt(p, "eddies", 14), paramInt(p, "streamlines", 18), paramFloat(p, "licLength", 24.0))
		},
	},
	{
		ID: "lorentzstrangeribbon", StructName: "lorentzStrangeRibbon", Category: "chaos",
		DefaultBG: color.RGBA{6, 9, 19, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewLorentzStrangeRibbon(paramInt(p, "strands", 26), paramFloat(p, "rho", 28.0), paramFloat(p, "beta", 2.667))
		},
	},
	{
		ID: "coronalmagneticarcade", StructName: "coronalMagneticArcade", Category: "media",
		DefaultBG: color.RGBA{5, 7, 16, 255}, DefaultPal: "inferno",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCoronalMagneticArcade(paramInt(p, "dipoles", 9), paramInt(p, "strokeCount", 19000), paramFloat(p, "flareTwist", 1.15))
		},
	},

	// Wave 25: Multi-Paradigm Non-Moebius Flagship Quartet (HopfTorusSilk, VolcanicAtollTerraces, PeacockOcelliPlumes, ByzantineRoseMosaic)
	{
		ID: "hopftorussilk", StructName: "hopfTorusSilk", Category: "flow",
		DefaultBG: color.RGBA{6, 9, 19, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewHopfTorusSilk(paramInt(p, "fibers", 6), paramInt(p, "strands", 22), paramFloat(p, "cliffordAngle", 0.72))
		},
	},
	{
		ID: "volcanicatollterraces", StructName: "volcanicAtollTerraces", Category: "geom",
		DefaultBG: color.RGBA{244, 236, 214, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewVolcanicAtollTerraces(paramInt(p, "tiers", 16), paramInt(p, "islands", 4), paramFloat(p, "erosionStrength", 1.75))
		},
	},
	{
		ID: "peacockocelliplumes", StructName: "peacockOcelliPlumes", Category: "bio",
		DefaultBG: color.RGBA{12, 20, 30, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPeacockOcelliPlumes(paramInt(p, "ocelli", 7), paramInt(p, "barbCount", 18000), paramFloat(p, "filmNm", 390.0))
		},
	},
	{
		ID: "byzantinerosemosaic", StructName: "byzantineRoseMosaic", Category: "geom",
		DefaultBG: color.RGBA{12, 14, 24, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewByzantineRoseMosaic(paramInt(p, "petals", 12), paramInt(p, "rings", 4), paramFloat(p, "ribbonSpiral", 1.15))
		},
	},

	// Wave 24: Phase 2–5 Flagship Quartet (MuranoGlassNautilus, HaeckelRadiolaria, RelativisticLens, WeingartenTorusKnot)
	{
		ID: "muranoglassnautilus", StructName: "muranoGlassNautilus", Category: "media",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMuranoGlassNautilus(paramInt(p, "chambers", 14), paramInt(p, "facets", 5), paramFloat(p, "dispersion", 0.048))
		},
	},
	{
		ID: "haeckelradiolaria", StructName: "haeckelRadiolaria", Category: "bio",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewHaeckelRadiolaria(paramInt(p, "shells", 3), paramInt(p, "spineOrder", 6), paramFloat(p, "poreScale", 1.0))
		},
	},
	{
		ID: "relativisticlens", StructName: "relativisticLens", Category: "chaos",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "inferno",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewRelativisticLens(paramInt(p, "rings", 4), paramInt(p, "jets", 2), paramFloat(p, "spinRate", 0.85))
		},
	},
	{
		ID: "weingartentorusknot", StructName: "weingartenTorusKnot", Category: "geom",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewWeingartenTorusKnot(paramInt(p, "pWind", 2), paramInt(p, "qWind", 3), paramFloat(p, "tubeScale", 1.0))
		},
	},

	// Wave 23: Master Architecture, Celestial Mechanics & Sci-Fi Megastructures Quartet (BabelLibrary, KeplerHarmonices, MoebiusHermeticGarage, CalabiYauManifold)
	{
		ID: "babellibrary", StructName: "babelLibrary", Category: "geom",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewBabelLibrary(paramInt(p, "galleries", 6), paramInt(p, "tiers", 5), paramFloat(p, "shaftScale", 1.0))
		},
	},
	{
		ID: "keplerharmonices", StructName: "keplerHarmonices", Category: "geom",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewKeplerHarmonices(paramInt(p, "shells", 5), paramInt(p, "rings", 3), paramFloat(p, "inclination", 0.26))
		},
	},
	{
		ID: "moebiushermeticgarage", StructName: "moebiusHermeticGarage", Category: "media",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMoebiusHermeticGarage(paramInt(p, "ribs", 24), paramInt(p, "spires", 6), paramFloat(p, "ringTilt", 0.36))
		},
	},
	{
		ID: "calabiyaumanifold", StructName: "calabiYauManifold", Category: "chaos",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCalabiYauManifold(paramInt(p, "foldOrder", 5), paramInt(p, "rings", 2), paramFloat(p, "phaseMorph", 0.35))
		},
	},

	// Wave 22: Part VIII Flagship Quartet (PiranesiCarceri, MoebiusLeviathan, DurerPolyhedron, MineralShanShui)
	{
		ID: "piranesicarceri", StructName: "piranesiCarceri", Category: "geom",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPiranesiCarceri(paramInt(p, "arches", 4), paramInt(p, "tiers", 4), paramFloat(p, "vaultDepth", 1.0))
		},
	},
	{
		ID: "moebiusleviathan", StructName: "moebiusLeviathan", Category: "media",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMoebiusLeviathan(paramInt(p, "ribs", 14), paramInt(p, "crystalCount", 5), paramFloat(p, "bankAngle", -0.18))
		},
	},
	{
		ID: "durerpolyhedron", StructName: "durerPolyhedron", Category: "geom",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewDurerPolyhedron(paramInt(p, "symmetry", 12), paramInt(p, "rings", 3), paramFloat(p, "twist", 0.32))
		},
	},
	{
		ID: "mineralshanshui", StructName: "mineralShanShui", Category: "media",
		DefaultBG: color.RGBA{244, 234, 210, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMineralShanShui(paramInt(p, "peaks", 5), paramInt(p, "mistLayers", 4), paramFloat(p, "mineralRichness", 1.0))
		},
	},

	// Wave 21: Jean "Mœbius" Giraud Ligne Claire Quartet (MoebiusCanyon, MoebiusIncal, MoebiusEdena, MoebiusStarwatcher)
	{
		ID: "moebiuscanyon", StructName: "moebiusCanyon", Category: "media",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMoebiusCanyon(paramInt(p, "spires", 7), paramInt(p, "strataLevels", 14), paramFloat(p, "monolithTilt", 0.22))
		},
	},
	{
		ID: "moebiusincal", StructName: "moebiusIncal", Category: "geom",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMoebiusIncal(paramInt(p, "facets", 8), paramInt(p, "ringLayers", 3), paramFloat(p, "crystalScale", 1.0))
		},
	},
	{
		ID: "moebiusedena", StructName: "moebiusEdena", Category: "bio",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "botanical",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMoebiusEdena(paramInt(p, "domes", 5), paramInt(p, "terraces", 10), paramFloat(p, "floraScale", 1.0))
		},
	},
	{
		ID: "moebiusstarwatcher", StructName: "moebiusStarwatcher", Category: "geom",
		DefaultBG: color.RGBA{248, 241, 224, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMoebiusStarwatcher(paramInt(p, "rings", 3), paramInt(p, "sails", 4), paramFloat(p, "orbitTilt", -0.26))
		},
	},

	// Wave 20: M.C. Escher Trio (Circle Limit Hyperbolic Tessellation, Print Gallery Droste Spiral & Isohedral Metamorphosis)
	{
		ID: "eschercirclelimit", StructName: "escherCircleLimit", Category: "geom",
		DefaultBG: color.RGBA{243, 236, 224, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewEscherCircleLimit(paramInt(p, "p", 6), paramInt(p, "q", 4), paramInt(p, "r", 2))
		},
	},
	{
		ID: "escherdrostegallery", StructName: "escherDrosteGallery", Category: "geom",
		DefaultBG: color.RGBA{18, 22, 34, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewEscherDrosteGallery(paramInt(p, "branches", 1), paramInt(p, "archesPerRing", 8), paramFloat(p, "scaleRatio", 16.0))
		},
	},
	{
		ID: "eschermetamorphosis", StructName: "escherMetamorphosis", Category: "geom",
		DefaultBG: color.RGBA{248, 242, 228, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewEscherMetamorphosis(paramInt(p, "cols", 10), paramInt(p, "rows", 10), paramFloat(p, "morphStrength", 1.0))
		},
	},

	// Wave 19: Analytical Perlin & Gabor Noise Quartet (Botryoidal Perlin Malachite, Gradient-Damped Erosion Canyon, Analytical Curl Silk & Anisotropic Gabor Wave Interference)
	{
		ID: "perlinmalachite", StructName: "perlinMalachite", Category: "geom",
		DefaultBG: color.RGBA{7, 24, 18, 255}, DefaultPal: "botanical",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPerlinMalachite(paramInt(p, "bands", 16), paramInt(p, "octaves", 5), paramFloat(p, "warpStrength", 2.4))
		},
	},
	{
		ID: "erosioncanyon", StructName: "erosionCanyon", Category: "geom",
		DefaultBG: color.RGBA{12, 16, 28, 255}, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewErosionCanyon(paramInt(p, "octaves", 6), paramFloat(p, "erosionStrength", 1.8), paramFloat(p, "strataFreq", 18.0))
		},
	},
	{
		ID: "analyticalcurlsilk", StructName: "analyticalCurlSilk", Category: "flow",
		DefaultBG: color.RGBA{10, 18, 40, 255}, DefaultPal: "dunhuang",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewAnalyticalCurlSilk(paramInt(p, "streamlines", 420), paramInt(p, "octaves", 4), paramFloat(p, "curlScale", 2.2))
		},
	},
	{
		ID: "gaborwaveinterference", StructName: "gaborWaveInterference", Category: "flow",
		DefaultBG: color.RGBA{8, 12, 24, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewGaborWaveInterference(paramInt(p, "vortexPoles", 6), paramFloat(p, "gaborFreq", 3.2), paramFloat(p, "bandwidth", 1.45))
		},
	},

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
	{
		ID: "chladni", StructName: "chladni", Category: "chaos",
		DefaultBG: color.RGBA{8, 10, 18, 255}, DefaultPal: "cyberpunk",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewChladni(paramInt(p, "particlesN", 35000), paramFloat(p, "m", 5.0), paramFloat(p, "n", 3.0), paramFloat(p, "a", 1.0), paramFloat(p, "b", -1.0))
		},
	},

	// Wave 1 Classic Engines (31 Foundational Generators)
	{
		ID: "blackhole", StructName: "blackHole", Category: "flow",
		DefaultBG: color.RGBA{20, 20, 20, 255}, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewBlackHole(paramInt(p, "circleN", 200), paramFloat(p, "density", 400.0), paramFloat(p, "circleGap", 0.01))
		},
	},
	{
		ID: "circlegrid", StructName: "circleGrid", Category: "geom",
		DefaultBG: color.RGBA{16, 22, 36, 255}, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCircleGrid(paramInt(p, "circleNumMin", 4), paramInt(p, "circleNumMax", 8))
		},
	},
	{
		ID: "circleline", StructName: "circleLine", Category: "geom",
		DefaultBG: common.Lavender, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCircleLine(paramFloat(p, "step", 0.02), paramInt(p, "lineNum", 600), paramFloat(p, "radius", 1.5), paramFloat(p, "xaixs", 2.0), paramFloat(p, "yaixs", 2.0))
		},
	},
	{
		ID: "circleloop", StructName: "circleLoop", Category: "geom",
		DefaultBG: common.Black, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCircleLoop(paramFloat(p, "radius", 100.0))
		},
	},
	{
		ID: "circleloop2", StructName: "circleLoop2", Category: "geom",
		DefaultBG: color.RGBA{230, 230, 230, 255}, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCircleLoop2(paramInt(p, "depth", 7))
		},
	},
	{
		ID: "circlemove", StructName: "circleMove", Category: "flow",
		DefaultBG: common.White, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCircleMove(paramInt(p, "circleNum", 1000))
		},
	},
	{
		ID: "circlenoise", StructName: "circleNoise", Category: "flow",
		DefaultBG: common.White, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewCircleNoise(paramInt(p, "dotsN", 2000), paramInt(p, "colorMin", 60), paramInt(p, "colorMax", 80))
		},
	},
	{
		ID: "colorcanva", StructName: "colorCanva", Category: "geom",
		DefaultBG: common.Black, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewColorCanve(paramFloat(p, "seg", 5.0))
		},
	},
	{
		ID: "colorcircle", StructName: "colorCircle", Category: "geom",
		DefaultBG: common.White, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewColorCircle(paramInt(p, "circleNum", 30))
		},
	},
	{
		ID: "colorcircle2", StructName: "colorCircle2", Category: "geom",
		DefaultBG: common.White, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewColorCircle2(paramInt(p, "circleNum", 30))
		},
	},
	{
		ID: "contourline", StructName: "contourLine", Category: "flow",
		DefaultBG: color.RGBA{26, 6, 51, 255}, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewContourLine(paramInt(p, "lineNum", 500))
		},
	},
	{
		ID: "domainwrap", StructName: "domainWrap", Category: "flow",
		DefaultBG: common.Black, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewDomainWrap(
				paramFloat(p, "scale", 0.01),
				paramFloat(p, "scale2", 4.0),
				paramFloat(p, "xOffset", 4.0),
				paramFloat(p, "yOffset", 20.0),
				func(r, m1, m2 float64) color.RGBA {
					return color.RGBA{
						R: uint8(common.Constrain(m1*200*r, 0, 255)),
						G: uint8(common.Constrain(r*200, 0, 255)),
						B: uint8(common.Constrain(m2*255*r, 70, 255)),
						A: 255,
					}
				},
			)
		},
	},
	{
		ID: "dotline", StructName: "dotLine", Category: "geom",
		DefaultBG: color.RGBA{230, 230, 230, 255}, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewDotLine(paramInt(p, "n", 100), paramFloat(p, "ras", 20.0), paramFloat(p, "canv", 50.0), paramBool(p, "randColor", false))
		},
	},
	{
		ID: "dotswave", StructName: "dotsWave", Category: "flow",
		DefaultBG: common.Black, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewDotsWave(paramInt(p, "dotsN", 300))
		},
	},
	{
		ID: "janus", StructName: "janus", Category: "geom",
		DefaultBG: common.Black, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewJanus(paramInt(p, "n", 10), paramFloat(p, "decay", 0.2))
		},
	},
	{
		ID: "julia", StructName: "julia", Category: "chaos",
		DefaultBG: common.Black, DefaultPal: "inferno",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewJulia(func(z complex128) complex128 {
				return z*z + complex(-0.1, 0.651)
			}, paramFloat(p, "maxz", 40.0), paramFloat(p, "xaixs", 1.5), paramFloat(p, "yaixs", 1.5))
		},
	},
	{
		ID: "maze", StructName: "maze", Category: "geom",
		DefaultBG: common.Azure, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewMaze(paramInt(p, "step", 20))
		},
	},
	{
		ID: "noiseline", StructName: "noiseLine", Category: "flow",
		DefaultBG: color.RGBA{230, 230, 230, 255}, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewNoiseLine(paramInt(p, "n", 1000))
		},
	},
	{
		ID: "oceanfish", StructName: "oceanFish", Category: "bio",
		DefaultBG: color.RGBA{18, 24, 42, 255}, DefaultPal: "oceanic",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewOceanFish(paramInt(p, "lineNum", 100), paramInt(p, "fishNum", 8))
		},
	},
	{
		ID: "perlinpearls", StructName: "perlinPearls", Category: "flow",
		DefaultBG: color.RGBA{230, 230, 230, 255}, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPerlinPerls(paramInt(p, "circleN", 10), paramInt(p, "dotsN", 200), paramInt(p, "colorMin", 40), paramInt(p, "colorMax", 80))
		},
	},
	{
		ID: "pixelhole", StructName: "pixelHole", Category: "flow",
		DefaultBG: common.Black, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPixelHole(paramInt(p, "dotN", 60))
		},
	},
	{
		ID: "pointribbon", StructName: "pointRibbon", Category: "geom",
		DefaultBG: common.Lavender, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewPointRibbon(paramFloat(p, "r", 50.0))
		},
	},
	{
		ID: "randcircle", StructName: "randCircle", Category: "geom",
		DefaultBG: common.MistyRose, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewRandCicle(paramInt(p, "maxCircle", 30), paramInt(p, "maxStepsPerCircle", 80), paramFloat(p, "minSteps", 0.2), paramFloat(p, "maxSteps", 2.0), paramFloat(p, "minRadius", 10.0), paramFloat(p, "maxRadius", 30.0), paramBool(p, "isRandColor", true))
		},
	},
	{
		ID: "randomshape", StructName: "randomShape", Category: "geom",
		DefaultBG: common.White, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewRandomShape(paramInt(p, "shapeNum", 150))
		},
	},
	{
		ID: "silksky", StructName: "silkSky", Category: "media",
		DefaultBG: common.Black, DefaultPal: "sunset",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSilkSky(paramInt(p, "circleNum", 15), paramFloat(p, "sunRadius", 5.0))
		},
	},
	{
		ID: "silksmoke", StructName: "sileSmoke", Category: "media",
		DefaultBG: common.Black, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSilkSmoke(paramInt(p, "maxCircle", 400), paramInt(p, "maxStepsPerCircle", 20), paramFloat(p, "minSteps", 0.2), paramFloat(p, "maxSteps", 2.0), paramFloat(p, "minRadius", 10.0), paramFloat(p, "maxRadius", 30.0), paramBool(p, "isRandColor", true))
		},
	},
	{
		ID: "solarflare", StructName: "solarFlare", Category: "flow",
		DefaultBG: common.Black, DefaultPal: "inferno",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSolarFlare()
		},
	},
	{
		ID: "spiralsquare", StructName: "spiralSquare", Category: "geom",
		DefaultBG: common.MistyRose, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSpiralSquare(paramInt(p, "squareNum", 40), paramFloat(p, "rectSide", 400.0), paramFloat(p, "decay", 0.05), paramBool(p, "randColor", true))
		},
	},
	{
		ID: "gridsquare", StructName: "girdSquares", Category: "geom",
		DefaultBG: common.Black, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewGirdSquares(paramInt(p, "step", 24), paramInt(p, "rectSize", 10), paramFloat(p, "decay", 0.2))
		},
	},
	{
		ID: "swirl", StructName: "swirl", Category: "chaos",
		DefaultBG: color.RGBA{113, 3, 0, 255}, DefaultPal: "inferno",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewSwirl(paramFloat(p, "a", 0.970), paramFloat(p, "b", -1.899), paramFloat(p, "c", 1.381), paramFloat(p, "d", -1.506), paramFloat(p, "xaixs", 2.4), paramFloat(p, "yaixs", 2.4))
		},
	},
	{
		ID: "yarn", StructName: "yarn", Category: "flow",
		DefaultBG: common.Orange, DefaultPal: "youthful",
		Build: func(p map[string]interface{}) generativeart.Engine {
			return NewYarn(paramInt(p, "n", 2000))
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
	if q == "escherdrostewave" {
		return ArtDescriptor{
			ID: "escherdrostewave", StructName: "escherDrosteGallery", Category: "geom",
			DefaultBG: color.RGBA{18, 22, 34, 255}, DefaultPal: "dunhuang",
			Build: func(p map[string]interface{}) generativeart.Engine {
				return NewEscherDrosteGallery(paramInt(p, "branches", 2), paramInt(p, "archesPerRing", 8), paramFloat(p, "scaleRatio", 16.0))
			},
		}, true
	}
	for _, d := range registry {
		if d.ID == q || strings.ToLower(d.StructName) == q {
			return d, true
		}
	}
	return ArtDescriptor{}, false
}

// PrepareReplayCanva constructs a configured *generativeart.Canva and Engine from meta,
// restoring exact dimensions (or overrideW x overrideH), seed, background/foreground/line colors,
// exact Palette stops, Options (LineWidth, NIters, Alpha), PostEffects, and optional parameter overrides
// without immediately executing Draw.
func PrepareReplayCanva(meta *generativeart.RecipeMetadata, overrideW, overrideH int, paramOverrides map[string]interface{}) (*generativeart.Canva, generativeart.Engine, error) {
	if meta == nil {
		return nil, nil, fmt.Errorf("nil RecipeMetadata")
	}
	desc, ok := Lookup(meta.Engine)
	if !ok {
		return nil, nil, fmt.Errorf("engine %q is not registered for automatic replay", meta.Engine)
	}
	w, h := meta.Width, meta.Height
	if overrideW > 0 && overrideH > 0 {
		w, h = overrideW, overrideH
	} else if overrideW > 0 {
		w = overrideW
	} else if overrideH > 0 {
		h = overrideH
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
	if meta.Foreground != ([4]uint8{}) {
		c.SetForeground(color.RGBA{
			R: meta.Foreground[0],
			G: meta.Foreground[1],
			B: meta.Foreground[2],
			A: meta.Foreground[3],
		})
	}
	if meta.LineColor != ([4]uint8{}) {
		c.SetLineColor(color.RGBA{
			R: meta.LineColor[0],
			G: meta.LineColor[1],
			B: meta.LineColor[2],
			A: meta.LineColor[3],
		})
	}
	if meta.LineWidth > 0 {
		c.SetLineWidth(meta.LineWidth)
	}
	if meta.NIters > 0 {
		c.SetIterations(meta.NIters)
	}
	if meta.Alpha > 0 {
		c.SetAlpha(meta.Alpha)
	}
	if len(meta.Palette) > 0 {
		pal := make([]color.RGBA, len(meta.Palette))
		for i, stop := range meta.Palette {
			pal[i] = color.RGBA{R: stop[0], G: stop[1], B: stop[2], A: stop[3]}
		}
		c.SetColorSchema(pal)
	} else if pal, exists := PaletteByName(desc.DefaultPal); exists {
		c.SetColorSchema(pal)
	}
	for _, effSpec := range meta.PostEffects {
		if eff, parsed := render.ParsePostEffect(effSpec); parsed && eff != nil {
			c.AddPostEffect(eff)
		}
	}
	c.FillBackground()

	mergedParams := make(map[string]interface{}, len(meta.Params)+len(paramOverrides))
	for k, v := range meta.Params {
		mergedParams[k] = v
	}
	if len(paramOverrides) > 0 {
		resolved, err := desc.ResolveParams(paramOverrides)
		if err != nil {
			return nil, nil, err
		}
		for k := range paramOverrides {
			for rk, rv := range resolved {
				if strings.EqualFold(k, rk) {
					mergedParams[rk] = rv
				}
			}
		}
	}
	eng := desc.Build(mergedParams)
	return c, eng, nil
}

// ReplayRecipe reconstructs and renders the artwork described by meta at the original
// resolution (or overrideW x overrideH if > 0).
func ReplayRecipe(meta *generativeart.RecipeMetadata, overrideW, overrideH int) (*generativeart.Canva, error) {
	c, eng, err := PrepareReplayCanva(meta, overrideW, overrideH, nil)
	if err != nil {
		return nil, err
	}
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
