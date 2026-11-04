package render

import (
	"math"

	"github.com/jdxyw/generativeart/common"
)

// ReflectVec3 returns the exact mirror reflection vector for incident unit direction wi
// off surface normal n: wr = wi - 2*(wi . n)*n.
func ReflectVec3(wi, n Vec3) Vec3 {
	return wi.Sub(n.Scale(2.0 * wi.Dot(n))).Normalize()
}

// RefractVec3 computes the refracted unit ray direction wt using vector Snell's law
// (eta1 * sin(theta1) = eta2 * sin(theta2), where eta = eta1 / eta2) for incident unit
// vector wi and surface normal n oriented against wi (wi . n <= 0).
// If total internal reflection occurs (sin^2(theta_t) > 1), it returns the reflected vector and tir=true.
func RefractVec3(wi, n Vec3, eta float64) (wt Vec3, tir bool) {
	cosI := -wi.Dot(n)
	if cosI < 0 {
		n = n.Scale(-1.0)
		cosI = -cosI
	}
	cosI = common.Constrain(cosI, 0.0, 1.0)
	sin2T := eta * eta * (1.0 - cosI*cosI)
	if sin2T > 1.0 {
		return ReflectVec3(wi, n), true
	}
	cosT := math.Sqrt(math.Max(0.0, 1.0-sin2T))
	return wi.Scale(eta).Add(n.Scale(eta*cosI - cosT)).Normalize(), false
}

// SchlickFresnel evaluates unpolarized dielectric reflectance at incident cosine cosTheta in [0, 1]
// between media of refractive indices ior1 and ior2, accounting for Brewster/critical-angle TIR
// when transitioning from a denser medium (ior1 > ior2).
func SchlickFresnel(cosTheta, ior1, ior2 float64) float64 {
	ct := common.Constrain(math.Abs(cosTheta), 0.0, 1.0)
	if ior1 <= 0 {
		ior1 = 1.0
	}
	if ior2 <= 0 {
		ior2 = 1.5
	}
	r0 := (ior1 - ior2) / (ior1 + ior2)
	r0 = r0 * r0

	if ior1 > ior2 {
		eta := ior1 / ior2
		sin2T := eta * eta * (1.0 - ct*ct)
		if sin2T >= 1.0 {
			return 1.0
		}
		ct = math.Sqrt(math.Max(0.0, 1.0-sin2T))
	}
	oneMinus := 1.0 - ct
	oneMinus2 := oneMinus * oneMinus
	return common.Clamp01(r0 + (1.0-r0)*oneMinus2*oneMinus2*oneMinus)
}

// BeerLambertTransmittance computes spectral volume transmittance T(d) = exp(-sigma_a * d)
// across linear R, G, B bands for an interior ray path of length distance >= 0.
func BeerLambertTransmittance(absorbSigma [3]float64, distance float64) [3]float64 {
	if distance <= 0 {
		return [3]float64{1.0, 1.0, 1.0}
	}
	return [3]float64{
		math.Exp(-math.Max(0.0, absorbSigma[0]) * distance),
		math.Exp(-math.Max(0.0, absorbSigma[1]) * distance),
		math.Exp(-math.Max(0.0, absorbSigma[2]) * distance),
	}
}

// DielectricMaterial3D configures physical 3D SDF dielectric refraction, Beer-Lambert
// spectral volume absorption, Abbe chromatic dispersion, and optional thin-film iridescence.
type DielectricMaterial3D struct {
	IOR                float64    // Base index of refraction at green wavelength (e.g. 1.52 glass, 2.417 diamond)
	DispersionDelta    float64    // Spectral IOR difference (ior_Blue - ior_Red), e.g. 0.044 for flint prism
	AbsorbSigma        [3]float64 // Beer-Lambert linear RGB absorption coefficients per unit interior distance
	SpecularPower      float64    // Phong/Blinn surface highlight exponent (default 96)
	SpecularStrength   float64    // Direct sun specular highlight multiplier (default 0.85)
	ThinFilmNm         float64    // Optional surface Airy thin-film coating thickness in nm (0 = uncoated)
	ThinFilmIOR        float64    // Refractive index of thin-film coating (default 1.38)
	MaxInternalBounces int        // Maximum internal TIR/refraction bounces inside the negative SDF volume (default 3)
}

// DielectricCrownGlass returns an optical crown glass preset (clear, low dispersion).
func DielectricCrownGlass() DielectricMaterial3D {
	return DielectricMaterial3D{
		IOR:                1.52,
		DispersionDelta:    0.012,
		AbsorbSigma:        [3]float64{0.08, 0.03, 0.05},
		SpecularPower:      110.0,
		SpecularStrength:   0.90,
		MaxInternalBounces: 3,
	}
}

// DielectricFlintPrism returns a high-dispersion optical flint prism preset.
func DielectricFlintPrism() DielectricMaterial3D {
	return DielectricMaterial3D{
		IOR:                1.68,
		DispersionDelta:    0.055,
		AbsorbSigma:        [3]float64{0.04, 0.03, 0.02},
		SpecularPower:      140.0,
		SpecularStrength:   1.05,
		MaxInternalBounces: 4,
	}
}

// DielectricBalticAmber returns a warm honey-gold Baltic amber preset with strong blue/green Beer-Lambert absorption.
func DielectricBalticAmber() DielectricMaterial3D {
	return DielectricMaterial3D{
		IOR:                1.55,
		DispersionDelta:    0.016,
		AbsorbSigma:        [3]float64{0.18, 0.95, 3.60},
		SpecularPower:      84.0,
		SpecularStrength:   0.80,
		MaxInternalBounces: 3,
	}
}

// DielectricMuranoEmerald returns a Venetian Murano emerald-teal blown glass preset with iridescent Dichroic glaze.
func DielectricMuranoEmerald() DielectricMaterial3D {
	return DielectricMaterial3D{
		IOR:                1.56,
		DispersionDelta:    0.024,
		AbsorbSigma:        [3]float64{1.85, 0.16, 0.48},
		SpecularPower:      120.0,
		SpecularStrength:   0.95,
		ThinFilmNm:         340.0,
		ThinFilmIOR:        1.42,
		MaxInternalBounces: 3,
	}
}

// DielectricCobaltSapphire returns a deep royal-sapphire crystal preset.
func DielectricCobaltSapphire() DielectricMaterial3D {
	return DielectricMaterial3D{
		IOR:                1.77,
		DispersionDelta:    0.032,
		AbsorbSigma:        [3]float64{2.10, 0.72, 0.12},
		SpecularPower:      135.0,
		SpecularStrength:   1.0,
		MaxInternalBounces: 4,
	}
}

// DielectricDiamond returns a brilliant-cut diamond preset (IOR = 2.417, high fire dispersion).
func DielectricDiamond() DielectricMaterial3D {
	return DielectricMaterial3D{
		IOR:                2.417,
		DispersionDelta:    0.068,
		AbsorbSigma:        [3]float64{0.02, 0.02, 0.015},
		SpecularPower:      180.0,
		SpecularStrength:   1.25,
		MaxInternalBounces: 5,
	}
}

// DielectricEnvFn evaluates the background/environment linear RGB radiance along ray (ro, rd).
// channel is -1 for full-RGB queries or 0/1/2 when tracing dispersed R/G/B sub-rays.
type DielectricEnvFn func(ro, rd Vec3, channel int) [3]float64

// MarchSDF3DInterior marches a ray starting inside the negative interior of sdf (sdf(p) < 0)
// along unit direction rd until it exits back to the surface boundary (sdf(p) >= 0).
// Returns the boundary exit position, total interior path length, and exited=true if the boundary was reached.
func MarchSDF3DInterior(sdf SDF3D, ro, rd Vec3, maxDist float64, maxSteps int) (exitPos Vec3, pathLen float64, exited bool) {
	if maxSteps <= 0 {
		maxSteps = 56
	}
	if maxDist <= 0 {
		maxDist = 12.0
	}
	t := 0.004
	prevNeg := 0.0
	stepLen := 0.0

	for i := 0; i < maxSteps && t < maxDist; i++ {
		pos := ro.Add(rd.Scale(t))
		d := sdf(pos)
		if d >= -0.0008 {
			if d > 0 && prevNeg < 0 && stepLen > 0 {
				frac := (-prevNeg) / ((-prevNeg) + d)
				t = (t - stepLen) + stepLen*common.Clamp01(frac)
				pos = ro.Add(rd.Scale(t))
			}
			return pos, t, true
		}
		prevNeg = d
		// Inside the volume d is negative; step by |d| with a minimum step floor
		stepLen = math.Max(0.006, -d*0.92)
		t += stepLen
	}
	return ro.Add(rd.Scale(t)), t, false
}

func traceInteriorSingleChannel(
	sdf SDF3D,
	entryPos, entryNormal, wi Vec3,
	ior float64,
	maxBounces int,
	maxDist float64,
	envFn DielectricEnvFn,
	channel int,
) (transmitted [3]float64, totalInteriorDist float64) {
	etaIn := 1.0 / math.Max(1.01, ior)
	curDir, tir := RefractVec3(wi, entryNormal, etaIn)
	if tir {
		return envFn(entryPos, curDir, channel), 0
	}
	// Step slightly inside the negative SDF surface along -entryNormal
	curPos := entryPos.Sub(entryNormal.Scale(0.006))

	bounces := maxBounces
	if bounces <= 0 {
		bounces = 3
	}
	etaOut := math.Max(1.01, ior)
	accumRadiance := [3]float64{}
	throughput := 1.0

	for b := 0; b < bounces && throughput > 0.02; b++ {
		exitPos, segLen, exited := MarchSDF3DInterior(sdf, curPos, curDir, maxDist, 52)
		totalInteriorDist += segLen
		if !exited {
			envCol := envFn(exitPos, curDir, channel)
			accumRadiance[0] += throughput * envCol[0]
			accumRadiance[1] += throughput * envCol[1]
			accumRadiance[2] += throughput * envCol[2]
			break
		}

		// Outward surface normal at exitPos points toward exterior (positive SDF gradient);
		// interior normal facing the outgoing ray is -outNormal.
		outNormal, _ := CalcSDF3DNormalAndCurvature(sdf, exitPos, 0.0025)
		inNormal := outNormal.Scale(-1.0)
		cosIn := common.Clamp01(-curDir.Dot(inNormal))
		fresnelIn := SchlickFresnel(cosIn, etaOut, 1.0)

		refrOutDir, isTIR := RefractVec3(curDir, inNormal, etaOut)
		if !isTIR {
			transWeight := throughput * (1.0 - fresnelIn)
			outOrigin := exitPos.Add(outNormal.Scale(0.006))
			envCol := envFn(outOrigin, refrOutDir, channel)
			accumRadiance[0] += transWeight * envCol[0]
			accumRadiance[1] += transWeight * envCol[1]
			accumRadiance[2] += transWeight * envCol[2]
		}

		// Continue internal reflection bounce
		throughput *= fresnelIn
		curDir = ReflectVec3(curDir, inNormal)
		curPos = exitPos.Add(inNormal.Scale(0.006))
	}

	return accumRadiance, totalInteriorDist
}

// TraceSDF3DDielectric traces a primary ray (ro, rd) against a 3D dielectric SDF volume,
// evaluating exterior Schlick-Fresnel reflection, optional thin-film iridescence, interior
// multi-bounce Snell refraction + Total Internal Reflection (TIR), Beer-Lambert spectral
// volume absorption, and Abbe-number RGB chromatic dispersion.
func TraceSDF3DDielectric(
	sdf SDF3D,
	ro, rd, lightDir Vec3,
	maxDist float64,
	mat DielectricMaterial3D,
	envFn DielectricEnvFn,
) (rgb [3]float64, hit SDF3DHit) {
	if envFn == nil {
		envFn = func(_, d Vec3, _ int) [3]float64 {
			u := common.Clamp01(0.5 + 0.5*d.Y)
			return [3]float64{0.12 + 0.55*u, 0.16 + 0.62*u, 0.24 + 0.68*u}
		}
	}
	if mat.IOR <= 1.0 {
		mat.IOR = 1.52
	}
	if mat.SpecularPower <= 0 {
		mat.SpecularPower = 96.0
	}
	if mat.SpecularStrength <= 0 {
		mat.SpecularStrength = 0.85
	}
	if mat.MaxInternalBounces <= 0 {
		mat.MaxInternalBounces = 3
	}

	hit = RaymarchSDF3D(sdf, ro, rd, lightDir, maxDist, 72)
	if !hit.Hit {
		return envFn(ro, rd, -1), hit
	}

	n := hit.Normal
	cosTheta := common.Clamp01(-rd.Dot(n))
	fresnel := SchlickFresnel(cosTheta, 1.0, mat.IOR)

	// Exterior mirror reflection + Blinn-Phong sun specular highlight
	reflDir := ReflectVec3(rd, n)
	reflOrigin := hit.Pos.Add(n.Scale(0.006))
	reflCol := envFn(reflOrigin, reflDir, -1)

	if mat.ThinFilmNm > 0 {
		tfR, tfG, tfB := SampleThinFilmLinearRGB(cosTheta, mat.ThinFilmNm, mat.ThinFilmIOR)
		reflCol[0] = reflCol[0]*0.55 + tfR*0.65
		reflCol[1] = reflCol[1]*0.55 + tfG*0.65
		reflCol[2] = reflCol[2]*0.55 + tfB*0.65
	}

	halfVec := lightDir.Sub(rd).Normalize()
	spec := math.Pow(math.Max(0.0, n.Dot(halfVec)), mat.SpecularPower) * mat.SpecularStrength

	var refrCol [3]float64
	if mat.DispersionDelta > 1e-4 {
		// Trace 3 dispersed chromatic channels: Red (lowest IOR), Green (base IOR), Blue (highest IOR)
		iorR := math.Max(1.01, mat.IOR-0.5*mat.DispersionDelta)
		iorG := math.Max(1.01, mat.IOR)
		iorB := math.Max(1.01, mat.IOR+0.5*mat.DispersionDelta)

		colR, distR := traceInteriorSingleChannel(sdf, hit.Pos, n, rd, iorR, mat.MaxInternalBounces, maxDist, envFn, 0)
		colG, distG := traceInteriorSingleChannel(sdf, hit.Pos, n, rd, iorG, mat.MaxInternalBounces, maxDist, envFn, 1)
		colB, distB := traceInteriorSingleChannel(sdf, hit.Pos, n, rd, iorB, mat.MaxInternalBounces, maxDist, envFn, 2)

		attR := BeerLambertTransmittance(mat.AbsorbSigma, distR)
		attG := BeerLambertTransmittance(mat.AbsorbSigma, distG)
		attB := BeerLambertTransmittance(mat.AbsorbSigma, distB)

		refrCol[0] = colR[0] * attR[0]
		refrCol[1] = colG[1] * attG[1]
		refrCol[2] = colB[2] * attB[2]
	} else {
		colRGB, dist := traceInteriorSingleChannel(sdf, hit.Pos, n, rd, mat.IOR, mat.MaxInternalBounces, maxDist, envFn, -1)
		att := BeerLambertTransmittance(mat.AbsorbSigma, dist)
		refrCol[0] = colRGB[0] * att[0]
		refrCol[1] = colRGB[1] * att[1]
		refrCol[2] = colRGB[2] * att[2]
	}

	// Combine transmitted dielectric radiance and surface Fresnel reflection + specular highlight
	transFactor := 1.0 - fresnel
	rgb[0] = transFactor*refrCol[0] + fresnel*reflCol[0] + spec
	rgb[1] = transFactor*refrCol[1] + fresnel*reflCol[1] + spec*0.98
	rgb[2] = transFactor*refrCol[2] + fresnel*reflCol[2] + spec*0.94
	return rgb, hit
}

// EvalSDF3DCausticIntensity estimates refractive caustic focusing and Beer-Lambert transmission
// on a receiver surface at (surfacePos, surfaceNormal) illuminated along unit vector lightDir
// through dielectric volume sdf. Convex dielectric lenses (mean curvature kappa > 0) concentrate
// refracted rays into high-intensity caustic patterns.
func EvalSDF3DCausticIntensity(sdf SDF3D, surfacePos, surfaceNormal, lightDir Vec3, mat DielectricMaterial3D, focusGain float64) [3]float64 {
	ndotl := math.Max(0.0, surfaceNormal.Dot(lightDir))
	if ndotl <= 0.0 {
		return [3]float64{0, 0, 0}
	}
	if focusGain <= 0 {
		focusGain = 2.4
	}
	ro := surfacePos.Add(surfaceNormal.Scale(0.01))
	hit := raymarchSDF3DInterval(sdf, ro, lightDir, lightDir, 0.015, 10.0, 40)
	if !hit.Hit {
		// Unoccluded direct light
		return [3]float64{ndotl, ndotl, ndotl}
	}

	// Ray entered the dielectric lens from the receiver side; measure interior thickness along lightDir
	inPos := hit.Pos.Sub(hit.Normal.Scale(0.008))
	exitPos, chordLen, _ := MarchSDF3DInterior(sdf, inPos, lightDir, 8.0, 36)
	exitNorm, exitCurv := CalcSDF3DNormalAndCurvature(sdf, exitPos, 0.003)

	// Converging convex lens curvature (kappa > 0) and normal alignment with lightDir focus caustics
	align := math.Abs(hit.Normal.Dot(lightDir)) * math.Abs(exitNorm.Dot(lightDir))
	curvFocus := 1.0 + focusGain*math.Max(0.0, hit.Curvature+exitCurv)
	causticAmp := ndotl * align * curvFocus * curvFocus

	// Optional spectral caustic fringes when DispersionDelta > 0
	att := BeerLambertTransmittance(mat.AbsorbSigma, chordLen)
	if mat.DispersionDelta > 1e-4 {
		tangentShift := math.Sqrt(math.Max(0.0, 1.0-align*align)) * mat.DispersionDelta * 14.0
		rMod := 1.0 + 0.35*math.Sin(tangentShift*6.0+0.8)
		gMod := 1.0 + 0.35*math.Sin(tangentShift*6.0+2.1)
		bMod := 1.0 + 0.35*math.Sin(tangentShift*6.0+3.5)
		return [3]float64{
			causticAmp * att[0] * rMod,
			causticAmp * att[1] * gMod,
			causticAmp * att[2] * bMod,
		}
	}
	return [3]float64{
		causticAmp * att[0],
		causticAmp * att[1],
		causticAmp * att[2],
	}
}
