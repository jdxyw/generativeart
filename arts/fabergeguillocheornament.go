package arts

import (
	"image/color"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type fabergeGuillocheOrnament struct {
	petals         int
	guillocheWaves int
	enamelDepth    float64
}

type fabergeRibbonSeg struct {
	x0, y0, x1, y1 float64
	zDepth         float64
	width          float64
	colCore        color.RGBA
	colEdge        color.RGBA
	specBoost      float64
}

// NewFabergeGuillocheOrnament returns a fabergeGuillocheOrnament engine that synthesizes a
// 3D ray-traced Imperial Fabergé translucent guilloché enamel egg monument tilted on a
// sculpted 24K gold & cabochon chalice pedestal and encircled by 3D depth-sorted orbiting
// gilded silk filigree ribbons.
func NewFabergeGuillocheOrnament(petals, guillocheWaves int, enamelDepth float64) *fabergeGuillocheOrnament {
	if petals <= 0 {
		petals = 12
	}
	if guillocheWaves <= 0 {
		guillocheWaves = 28
	}
	if enamelDepth <= 0 {
		enamelDepth = 1.0
	}
	return &fabergeGuillocheOrnament{
		petals:         common.ConstrainInt(petals, 6, 20),
		guillocheWaves: common.ConstrainInt(guillocheWaves, 12, 48),
		enamelDepth:    common.Constrain(enamelDepth, 0.4, 2.2),
	}
}

// Generative renders a static frame of the 3D Imperial Fabergé Guilloché Egg & Filigree Ribbons.
func (fgo *fabergeGuillocheOrnament) Generative(c *generativeart.Canva) {
	fgo.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (fgo *fabergeGuillocheOrnament) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	phi := 2.0 * math.Pi * phase

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	numPetals := fgo.petals
	// Ensure numPetals is a multiple of 3 (e.g. 12) so a 120-degree (2*pi/3) rotation per loop is 100% seamless
	// while advancing by an unmistakable 5 degrees per frame (eliminating Nyquist wagon-wheel aliasing)!
	if numPetals%3 != 0 {
		numPetals = ((numPetals + 2) / 3) * 3
	}
	if numPetals < 6 {
		numPetals = 12
	}
	fPetals := float64(numPetals)
	gWaves := float64(fgo.guillocheWaves)
	eDepth := fgo.enamelDepth

	linRGB := func(r, g, b uint8) [3]float64 {
		lr, lg, lb, _ := render.ColorToLinearRGB(color.RGBA{R: r, G: g, B: b, A: 255})
		return [3]float64{lr, lg, lb}
	}
	lerp3 := func(a, b [3]float64, t float64) [3]float64 {
		t = common.Clamp01(t)
		return [3]float64{
			a[0] + (b[0]-a[0])*t,
			a[1] + (b[1]-a[1])*t,
			a[2] + (b[2]-a[2])*t,
		}
	}

	velvetCol := linRGB(0x05, 0x07, 0x14)
	nebulaBlue := linRGB(0x0e, 0x22, 0x52)
	nebulaCrimson := linRGB(0x3a, 0x0c, 0x28)

	silverBase := linRGB(0xf6, 0xf8, 0xff)
	gold24k := linRGB(0xf8, 0xc4, 0x42)
	goldShadow := linRGB(0x58, 0x36, 0x08)
	giltSpec := linRGB(0xff, 0xf6, 0xcc)
	diamondWhite := linRGB(0xff, 0xfe, 0xfb)
	pearlCream := linRGB(0xff, 0xf4, 0xe4)

	// Beer-Lambert optical absorption coefficients sigma_a for translucent vitreous enamels:
	//   T(d) = exp(-sigma_a * d)
	rubySigma := [3]float64{0.07, 2.55, 2.15}      // Imperial Coronation Crimson Ruby
	sapphireSigma := [3]float64{2.45, 0.72, 0.08}  // Romanov Royal Cobalt Sapphire
	emeraldSigma := [3]float64{2.10, 0.14, 0.62}   // Ural Peacock Emerald
	turquoiseSigma := [3]float64{1.85, 0.18, 0.20} // Winter Palace Turquoise
	amberSigma := [3]float64{0.08, 0.85, 2.80}     // Sunburst Imperial Gold

	// 3D Ovoid (Fabergé Egg) Geometry:
	// Whole egg precesses in a 3D conical sway over phi = 2*pi*phase while spinning smoothly around its 3D polar axis!
	eggCX := 0.50*fw + 0.016*fw*math.Sin(phi)
	eggCY := 0.43*fh + 0.008*fh*math.Cos(phi)
	eggRY := 0.305 * minDim
	eggRX := 0.222 * minDim
	tiltAng := -0.15*math.Cos(phi) + 0.06*math.Sin(phi) // 3D conical precession of the whole egg body
	pitchNod := 0.11 * math.Sin(phi)                    // 3D forward/backward axial nod
	cosTilt, sinTilt := math.Cos(tiltAng), math.Sin(tiltAng)
	plinthY := 0.86 * fh

	// Studio key light + warm rim light
	lx := -0.48 + 0.14*math.Sin(phi)
	ly := -0.46 - 0.10*math.Cos(phi)
	lz := 0.75
	lLen := math.Sqrt(lx*lx + ly*ly + lz*lz)
	lx, ly, lz = lx/lLen, ly/lLen, lz/lLen
	hx, hy, hz := lx, ly, lz+1.0
	hLen := math.Sqrt(hx*hx + hy*hy + hz*hz)
	hx, hy, hz = hx/hLen, hy/hLen, hz/hLen

	// Evaluate exact 3D Ovoid (Fabergé Egg) surface intersection and 3D normal at screen pixel (px, py)
	evalEggSurface := func(px, py float64) (hit bool, uLon, vLat, nxW, nyW, nzW, zHeight float64) {
		dx := px - eggCX
		dy := py - eggCY
		// Rotate into tilted egg local coordinates (xL, yL) where yL < 0 is top crown, yL > 0 is lower base
		xL := dx*cosTilt + dy*sinTilt
		yL := -dx*sinTilt + dy*cosTilt

		uy := yL / eggRY
		if math.Abs(uy) >= 1.0 {
			return false, 0, 0, 0, 0, 0, 0
		}
		// Authentic avian/Fabergé egg taper: narrower pointed upper crown (uy < 0), fuller lower belly (uy > 0)
		taper := 1.0 + 0.145*uy - 0.045*uy*uy
		locRX := eggRX * taper
		maxXSq := (1.0 - uy*uy) * locRX * locRX
		if xL*xL >= maxXSq {
			return false, 0, 0, 0, 0, 0, 0
		}

		zL := math.Sqrt(maxXSq - xL*xL)
		zHeight = zL / eggRX

		// Analytical 3D surface normal in local coordinates
		nxL := xL / (locRX * locRX)
		nyL := uy/eggRY - (xL*xL+zL*zL)*0.145/(locRX*locRX*eggRY)
		nzL := zL / (locRX * locRX)
		nLen := math.Sqrt(nxL*nxL + nyL*nyL + nzL*nzL)
		nxL, nyL, nzL = nxL/nLen, nyL/nLen, nzL/nLen

		// Rotate normal back to world coordinates
		nxW = nxL*cosTilt - nyL*sinTilt
		nyW = nxL*sinTilt + nyL*cosTilt
		nzW = nzL

		// Apply 3D forward/backward pitch nod to (yL, zL) before computing spherical surface coordinates
		uyNod := common.Constrain(uy*math.Cos(pitchNod)-(zL/eggRX)*math.Sin(pitchNod)*0.32, -1.0, 1.0)
		vLat = math.Acos(-uyNod) / math.Pi
		// Advance longitude by 2*pi/3 per loop (exact 3-fold symmetry period of the 3 medallions & 12 trellis columns,
		// yielding a silky-smooth 5 deg/frame rightward 3D rotation with zero Nyquist strobing!)
		uLon = math.Atan2(xL, zL) - phi/3.0
		return true, uLon, vLat, nxW, nyW, nzW, zHeight
	}

	// Shade the 3D Imperial Fabergé Egg surface at (uLon, vLat) with normal (nx, ny, nz)
	shadeFabergeEgg := func(uLon, vLat, nx, ny, nz, zHeight float64) (rOut, gOut, bOut float64) {
		// 1. 3D Surface Ornamentation Zones on the Imperial Egg:
		// - Top & Bottom Imperial Crown Rosettes (vLat < 0.15 or vLat > 0.86)
		// - Upper & Lower Scalloped Gilded Garland Borders (vLat ~ 0.24 and vLat ~ 0.76)
		// - Central Equatorial Band (vLat in [0.24, 0.76]): Interlocking 3D Spiral Diamond-Trellis & Sunburst Medallions!
		helicalA := (uLon/(2.0*math.Pi))*fPetals + vLat*fPetals*0.42
		helicalB := (uLon/(2.0*math.Pi))*fPetals - vLat*fPetals*0.42

		// Distance to the 3D spiral gold trellis ribbons
		dTrellisA := math.Abs(math.Sin(helicalA*math.Pi)) / fPetals * 3.8
		dTrellisB := math.Abs(math.Sin(helicalB*math.Pi)) / fPetals * 3.8
		dTrellis := math.Min(dTrellisA, dTrellisB)

		// Distance to trellis intersection nodes (where rose-cut diamonds & seed pearls sit!)
		nodeDist := math.Hypot(math.Sin(helicalA*math.Pi), math.Sin(helicalB*math.Pi)) / fPetals * 3.8

		// Scalloped horizontal gold swags at vLat = 0.23 and vLat = 0.77
		swagTop := 0.22 + 0.025*math.Cos(uLon*fPetals)
		swagBot := 0.78 - 0.025*math.Cos(uLon*fPetals)
		dSwagTop := math.Abs(vLat - swagTop)
		dSwagBot := math.Abs(vLat - swagBot)

		// Three Imperial Sunburst Medallions spaced every 120 degrees around the equator (uLon = 0, +/- 2pi/3)
		// so as the 3D egg rotates 360 degrees, medallions continuously revolve across the curved belly!
		lonFold3 := math.Atan2(math.Sin(uLon*3.0), math.Cos(uLon*3.0)) / 3.0
		dxMed := (lonFold3 / math.Pi) * 1.55
		dyMed := (vLat - 0.50) * 1.35
		dMedallion := math.Hypot(dxMed, dyMed)
		medAngle := math.Atan2(dyMed, dxMed)
		medScallop := 0.21 + 0.018*math.Cos(16.0*medAngle)
		dMedRim := math.Abs(dMedallion - medScallop)

		// Determine active enamel compartment, guilloché pattern & distance to nearest 24K gold cloison
		var sigma [3]float64
		var metalBase [3]float64
		var minCloison float64
		var grooveWave, tanX, tanY float64

		if dMedallion < medScallop {
			// FRONT IMPERIAL CORONATION SUNBURST MEDALLION!
			uMed := dMedallion / medScallop
			dInnerStar := math.Abs(uMed - (0.42 + 0.14*math.Cos(12.0*medAngle))) * 0.22
			dSpokeMed := math.Abs(math.Sin(12.0*medAngle)) * uMed * 0.06
			minCloison = math.Min(dMedRim, math.Min(dInnerStar, dSpokeMed))

			if uMed < 0.42+0.14*math.Cos(12.0*medAngle) {
				// Blazing Imperial Gold Sunburst Core with Ruby heart
				if uMed < 0.16 {
					sigma = rubySigma
					metalBase = gold24k
				} else {
					sigma = amberSigma
					metalBase = lerp3(silverBase, gold24k, 0.55)
				}
			} else {
				// Alternating Pigeon-Blood Ruby & Turquoise Sunburst Petals
				secIdx := int(math.Floor((medAngle + math.Pi) / (2.0 * math.Pi) * 24.0))
				if secIdx%2 == 0 {
					sigma = rubySigma
					metalBase = lerp3(silverBase, gold24k, 0.30)
				} else {
					sigma = turquoiseSigma
					metalBase = silverBase
				}
			}
			// Radial rose-engine sunburst waves emanating from the medallion center
			g1 := math.Sin(uMed*gWaves*math.Pi*1.8 + 16.0*medAngle)
			g2 := math.Cos(uMed*gWaves*math.Pi*1.8 - 16.0*medAngle)
			grooveWave = 0.55*g1 + 0.45*g2
			tanX = -math.Sin(medAngle + tiltAng)
			tanY = math.Cos(medAngle + tiltAng)

		} else if vLat < swagTop || vLat > swagBot {
			// POLAR CROWN & BASE ACANTHUS LOTUS PETALS
			dPolarSpoke := math.Abs(math.Sin(uLon*fPetals)) * 0.045
			dPolarRing := math.Min(math.Abs(vLat-0.11), math.Abs(vLat-0.89))
			minCloison = math.Min(math.Min(dSwagTop, dSwagBot), math.Min(dPolarSpoke, dPolarRing))

			petalID := int(math.Floor((uLon + math.Pi*4.0) / (2.0 * math.Pi) * fPetals * 2.0))
			if vLat < 0.11 || vLat > 0.89 {
				sigma = amberSigma
				metalBase = gold24k
			} else if petalID%2 == 0 {
				sigma = sapphireSigma
				metalBase = silverBase
			} else {
				sigma = rubySigma
				metalBase = lerp3(silverBase, gold24k, 0.30)
			}
			grooveWave = math.Sin(vLat*gWaves*math.Pi*3.2 + math.Cos(uLon*fPetals*2.0)*2.5)
			tanX, tanY = cosTilt, -sinTilt

		} else {
			// EQUATORIAL SPIRAL DIAMOND-TRELLIS ENAMEL COMPARTMENTS!
			minCloison = math.Min(dTrellis, math.Min(dMedRim, math.Min(dSwagTop, dSwagBot)))

			lozA := int(math.Floor(helicalA + 100.0))
			lozB := int(math.Floor(helicalB + 100.0))
			lozChoice := (lozA + lozB) % 4
			switch lozChoice {
			case 0:
				sigma = sapphireSigma
				metalBase = silverBase
			case 1:
				sigma = emeraldSigma
				metalBase = lerp3(silverBase, gold24k, 0.25)
			case 2:
				sigma = rubySigma
				metalBase = lerp3(silverBase, gold24k, 0.35)
			default:
				sigma = turquoiseSigma
				metalBase = silverBase
			}

			// Basket-weave (panier) & barleycorn engine-turned guilloché inside each diamond lozenge
			wA := math.Sin(helicalA * gWaves * math.Pi * 0.45)
			wB := math.Sin(helicalB * gWaves * math.Pi * 0.45)
			wLat := math.Cos(vLat * gWaves * math.Pi * 2.6)
			grooveWave = 0.45*(wA+wB)*0.5 + 0.55*wLat
			tanX = math.Cos(helicalA - helicalB)
			tanY = math.Sin(helicalA - helicalB)
		}

		// Check if pixel is on a Rose-Cut Diamond or Seed Pearl at a trellis intersection node!
		if vLat > swagTop+0.02 && vLat < swagBot-0.02 && dMedallion > medScallop+0.025 && nodeDist < 0.022 {
			uGem := nodeDist / 0.022
			if uGem > 0.72 {
				// Gold bezel collet around the diamond
				bevel := math.Sin((uGem - 0.72) / 0.28 * math.Pi)
				gCol := lerp3(goldShadow, giltSpec, 0.25+0.75*bevel)
				return gCol[0], gCol[1], gCol[2]
			}
			// Faceted Rose-Cut Diamond sparkle!
			uD := uGem / 0.72
			facetSpark := 0.5 + 0.5*math.Cos(uLon*48.0+vLat*48.0-phi*2.0)
			dIllum := (0.65 + 0.65*math.Sqrt(1.0-uD*uD)) * (0.75 + 0.55*facetSpark)
			return diamondWhite[0] * dIllum, diamondWhite[1] * dIllum, diamondWhite[2] * dIllum
		}

		// Check if pixel is on a 3D Beveled 24K Gold Cloisonné Wire or Trellis Strut
		const wireW = 0.0085
		if minCloison < wireW {
			uWire := minCloison / wireW
			wireZ := math.Sqrt(math.Max(0.0, 1.0-uWire*uWire))
			// Perturb normal along the 3D egg surface + wire bevel
			ndotlWire := math.Max(0.0, nx*lx+ny*ly+nz*lz)
			goldLight := common.Clamp01(0.20 + 0.55*ndotlWire + 0.45*wireZ)
			gCol := lerp3(goldShadow, giltSpec, goldLight)
			// Chased micro-pearl beading along the spiral trellis wires!
			bead := math.Pow(0.5+0.5*math.Cos((helicalA+helicalB)*math.Pi*14.0), 8.0)
			gCol = lerp3(gCol, diamondWhite, 0.28*bead*wireZ)
			return gCol[0], gCol[1], gCol[2]
		}

		// 2. Physical 3D Ovoid Kajiya-Kay Guilloché + Beer-Lambert Vitreous Enamel Shading!
		// Perturb the 3D ovoid normal with the micro-machined lathe groove wave
		nxGroove := nx + tanY*grooveWave*0.32
		nyGroove := ny - tanX*grooveWave*0.32
		nzGroove := nz
		gLen := math.Sqrt(nxGroove*nxGroove + nyGroove*nyGroove + nzGroove*nzGroove)
		nxGroove, nyGroove, nzGroove = nxGroove/gLen, nyGroove/gLen, nzGroove/gLen

		ndotl := math.Max(0.0, nxGroove*lx+nyGroove*ly+nzGroove*lz)
		ndoth := math.Max(0.0, nxGroove*hx+nyGroove*hy+nzGroove*hz)
		// Warm golden bounce light from the pedestal below (ny > 0) and cool rim light on the right
		pedestalBounce := math.Max(0.0, 0.35*nx+0.75*ny+0.30*nz) * 0.35

		// Kajiya-Kay anisotropic silk cat's-eye highlight across the curved 3D egg!
		tdoth := tanX*hx + tanY*hy
		sinTH := math.Sqrt(math.Max(0.0, 1.0-tdoth*tdoth))
		anisoSpec := math.Pow(sinTH, 24.0) * math.Pow(ndoth, 8.0)
		facetSpec := math.Pow(ndoth, 56.0)

		metalIllum := 0.22 + 0.88*ndotl + pedestalBounce + 1.35*anisoSpec + 0.95*facetSpec

		// Beer-Lambert optical path length through the translucent enamel coat:
		// Increases at grazing viewing angles (1 / max(0.32, nz)) just like real curved glass on a 3D sphere/egg!
		domeFactor := common.Smoothstep(0.0, 0.038, minCloison)
		grazingPath := 1.0 / math.Max(0.38, nz)
		optPath := eDepth * (0.40 + 0.48*domeFactor - 0.12*grooveWave) * (0.65 + 0.35*grazingPath)

		transR := math.Exp(-sigma[0] * optPath)
		transG := math.Exp(-sigma[1] * optPath)
		transB := math.Exp(-sigma[2] * optPath)

		// Clear-coat 3D ovoid glass Fresnel reflection & studio window highlight!
		fresnel := 0.03 + 0.30*math.Pow(common.Clamp01(1.0-nz), 3.5)
		clearSpec := math.Pow(math.Max(0.0, nx*hx+ny*hy+nz*hz), 150.0)*0.68 + fresnel*0.18

		wireAO := 0.65 + 0.35*common.Smoothstep(wireW, wireW*3.0, minCloison)
		rOut = (metalBase[0]*metalIllum*transR + pearlCream[0]*clearSpec) * wireAO
		gOut = (metalBase[1]*metalIllum*transG + pearlCream[1]*clearSpec) * wireAO
		bOut = (metalBase[2]*metalIllum*transB + diamondWhite[2]*clearSpec) * wireAO
		return rOut, gOut, bOut
	}

	// Evaluate the 3D Sculpted 24K Gold & Lapis Cabochon Chalice Pedestal beneath the Egg (py in [0.68*fh, plinthY])
	evalChaliceStand := func(px, py float64) (hit bool, rS, gS, bS float64) {
		topY := eggCY + math.Cos(tiltAng)*(0.86*eggRY)
		if py < topY || py > plinthY {
			return false, 0, 0, 0
		}
		vStand := (py - topY) / math.Max(1.0, plinthY-topY) // 0 at chalice neck -> 1 at plinth base
		// Smoothly transition stand axis from the tilted egg base to the vertical center
		standCX := eggCX - math.Sin(tiltAng)*(0.86*eggRY)*(1.0-0.65*vStand)

		// Classical Fabergé chalice profile: narrow upper neck tucked cleanly under the egg -> baluster stem -> wide gadrooned foot
		profileR := (0.048*(1.0-0.35*vStand) +
			0.036*math.Exp(-math.Pow((vStand-0.42)/0.15, 2.0)) +
			0.168*math.Pow(common.Smoothstep(0.48, 1.0, vStand), 1.45)) * minDim

		dx := px - standCX
		if math.Abs(dx) > profileR {
			return false, 0, 0, 0
		}

		ux := dx / profileR // [-1, 1] across the 3D turned gold cylinder
		uz := math.Sqrt(math.Max(0.0, 1.0-ux*ux))

		// Gadrooned vertical flutes & engine-turned horizontal moldings on the 24K gold pedestal (rotates with phi!)
		flutes := math.Sin(math.Asin(common.Constrain(ux, -1, 1))*18.0 + phi*2.0)
		molding := math.Sin(vStand * math.Pi * 10.0)
		nx := ux + 0.22*flutes*uz
		ny := -0.18 * molding
		nz := uz
		nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
		nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

		ndotl := math.Max(0.0, nx*lx+ny*ly+nz*lz)
		ndoth := math.Max(0.0, nx*hx+ny*hy+nz*hz)
		spec := math.Pow(ndoth, 48.0)

		// Royal Sapphire & Ruby enamel collar band on the pedestal foot (vStand in [0.74, 0.86])
		if vStand > 0.74 && vStand < 0.86 {
			uCollar := (vStand - 0.74) / 0.12
			bevel := math.Sin(uCollar * math.Pi)
			transR := math.Exp(-sapphireSigma[0] * (0.7 + 0.5*bevel))
			transG := math.Exp(-sapphireSigma[1] * (0.7 + 0.5*bevel))
			transB := math.Exp(-sapphireSigma[2] * (0.7 + 0.5*bevel))
			illum := 0.35 + 0.85*ndotl + 0.90*spec
			return true, silverBase[0]*illum*transR + giltSpec[0]*spec*0.5,
				silverBase[1]*illum*transG + giltSpec[1]*spec*0.5,
				silverBase[2]*illum*transB + giltSpec[2]*spec*0.5
		}

		goldShade := common.Clamp01(0.18 + 0.65*ndotl + 0.25*uz)
		gCol := lerp3(goldShadow, gold24k, goldShade)
		gCol = lerp3(gCol, giltSpec, common.Clamp01(0.65*spec+0.20*math.Pow(uz, 4.0)))
		return true, gCol[0], gCol[1], gCol[2]
	}

	// 1. Pre-compute 4 Gyroscopically Rotating & Precessing 3D Gold & Enamel Filigree Rings Around the Egg!
	// Each ring is a closed 3D orbital band whose tilt plane rocks/precesses over phi while a variable-width
	// sculpted ribbon wave & orbiting gemstone satellites revolve 360 degrees around the egg!
	var backRibbons, frontRibbons []fabergeRibbonSeg
	type orbitingGem struct {
		x, y, z, rad float64
		col          color.RGBA
	}
	var backGems, frontGems []orbitingGem

	ribbonCols := []struct {
		core, edge, gem color.RGBA
	}{
		{color.RGBA{255, 208, 72, 255}, color.RGBA{255, 248, 210, 255}, color.RGBA{236, 36, 82, 255}},
		{color.RGBA{24, 206, 192, 255}, color.RGBA{255, 228, 120, 255}, color.RGBA{255, 248, 232, 255}},
		{color.RGBA{232, 36, 84, 255}, color.RGBA{255, 228, 120, 255}, color.RGBA{28, 214, 168, 255}},
		{color.RGBA{42, 132, 248, 255}, color.RGBA{255, 236, 148, 255}, color.RGBA{255, 216, 84, 255}},
	}

	ribbonSpecs := []struct {
		yOff, radX, inclin, rollSlope, speedDir, psi0 float64
	}{
		{-0.07 * fh, 0.35 * minDim, 0.34, 0.24, 1.0, 0.15},
		{0.06 * fh, 0.37 * minDim, -0.32, -0.22, -1.0, 0.95},
		{-0.01 * fh, 0.40 * minDim, 0.42, 0.32, 1.0, 1.85},
		{0.03 * fh, 0.32 * minDim, -0.38, -0.28, -1.0, 2.65},
	}

	const segsPerRibbon = 128
	for rIdx, rs := range ribbonSpecs {
		colPair := ribbonCols[rIdx%len(ribbonCols)]
		// True 3D Gyroscopic Armillary Rotation:
		// Each tilted 3D ring rotates around the vertical Y-axis by psi = psi0 + speedDir * phi
		// while its inclination rocks smoothly in 3D space!
		psi := rs.psi0 + rs.speedDir*phi
		cosPsi, sinPsi := math.Cos(psi), math.Sin(psi)
		sinInc, cosInc := math.Sin(rs.inclin), math.Cos(rs.inclin)
		spinPhase := phi * rs.speedDir

		evalOrbitPt := func(ang float64) (x, y, z float64) {
			// Local tilted 3D ring coordinates
			x0 := math.Sin(ang) * rs.radX
			y0 := math.Cos(ang) * rs.radX * sinInc
			z0 := math.Cos(ang) * rs.radX * cosInc
			// Rotate 3D ring around vertical Y-axis by azimuth psi(phi)
			x3 := x0*cosPsi + z0*sinPsi
			z3 := -x0*sinPsi + z0*cosPsi
			y3 := rs.yOff + y0 + x3*rs.rollSlope*math.Cos(phi) + 8.0*scale*math.Sin(2.0*(ang-spinPhase)+rs.psi0)
			return eggCX + x3, eggCY + y3, z3 / rs.radX
		}

		for s := 0; s < segsPerRibbon; s++ {
			ang0 := (float64(s)/float64(segsPerRibbon))*2.0*math.Pi - math.Pi
			ang1 := (float64(s+1)/float64(segsPerRibbon))*2.0*math.Pi - math.Pi
			angMid := 0.5 * (ang0 + ang1)

			x0, y0, z0 := evalOrbitPt(ang0)
			x1, y1, z1 := evalOrbitPt(ang1)
			zMid := 0.5 * (z0 + z1)

			// Travelling wave thickness modulation that rotates around the ring with spinPhase!
			waveW := 0.45 + 0.55*(0.5+0.5*math.Cos(2.0*(angMid-spinPhase)+rs.psi0))
			persp := 0.76 + 0.38*(zMid+1.0)*0.5
			wSeg := (2.0 + 4.8*waveW) * persp * scale

			dxSeg, dySeg := x1-x0, y1-y0
			segLen := math.Hypot(dxSeg, dySeg) + 1e-6
			tdoth := (dxSeg/segLen)*hx + (dySeg/segLen)*hy
			specB := math.Pow(math.Sqrt(math.Max(0.0, 1.0-tdoth*tdoth)), 14.0)

			seg := fabergeRibbonSeg{
				x0: x0, y0: y0, x1: x1, y1: y1,
				zDepth:    zMid,
				width:     wSeg,
				colCore:   colPair.core,
				colEdge:   colPair.edge,
				specBoost: specB,
			}
			if zMid < 0 {
				backRibbons = append(backRibbons, seg)
			} else {
				frontRibbons = append(frontRibbons, seg)
			}
		}

		// Place 2 orbiting cabochon jewel / pearl satellites per ring that revolve 360 deg around the egg!
		for g := 0; g < 2; g++ {
			gAng := spinPhase + rs.psi0 + float64(g)*math.Pi + 0.65
			gx, gy, gz := evalOrbitPt(gAng)
			gRad := (3.6 + 1.4*(gz+1.0)*0.5) * scale
			og := orbitingGem{x: gx, y: gy, z: gz, rad: gRad, col: colPair.gem}
			if gz < 0 {
				backGems = append(backGems, og)
			} else {
				frontGems = append(frontGems, og)
			}
		}
	}

	sort.Slice(backRibbons, func(i, j int) bool { return backRibbons[i].zDepth < backRibbons[j].zDepth })
	sort.Slice(frontRibbons, func(i, j int) bool { return frontRibbons[i].zDepth < frontRibbons[j].zDepth })

	drawRibbonList := func(segs []fabergeRibbonSeg, gems []orbitingGem) {
		shCol := color.RGBA{2, 3, 9, 255}
		whiteCol := color.RGBA{255, 252, 245, 255}
		goldBezel := color.RGBA{248, 198, 68, 255}
		flatWidth := func(float64) float64 { return 1.0 }
		for _, sg := range segs {
			// Extend segment endpoints slightly along tangent so adjacent segments fuse seamlessly with zero gaps!
			dx, dy := sg.x1-sg.x0, sg.y1-sg.y0
			pts := []render.Point{{X: sg.x0 - dx*0.18, Y: sg.y0 - dy*0.18}, {X: sg.x1 + dx*0.18, Y: sg.y1 + dy*0.18}}
			shPts := []render.Point{{X: pts[0].X + 3.0*scale, Y: pts[0].Y + 3.5*scale}, {X: pts[1].X + 3.0*scale, Y: pts[1].Y + 3.5*scale}}
			rc.DrawTaperedSpline(shPts, sg.width*1.3, flatWidth, func(float64) color.RGBA { return shCol }, 0.35)
			rc.DrawTaperedSpline(pts, sg.width, flatWidth, func(float64) color.RGBA { return sg.colEdge }, 0.96)
			rc.DrawTaperedSpline(pts, sg.width*0.62, flatWidth, func(float64) color.RGBA {
				return render.LerpOklab(sg.colCore, whiteCol, sg.specBoost*0.58)
			}, 0.96)
		}
		for _, gm := range gems {
			rc.DrawGlowDot(gm.x+2.0*scale, gm.y+2.5*scale, gm.rad*1.3, shCol, 0.55, 0.5)
			rc.DrawGlowDot(gm.x, gm.y, gm.rad*1.35, goldBezel, 0.96, 0.85)
			rc.DrawGlowDot(gm.x, gm.y, gm.rad, gm.col, 0.96, 0.80)
			rc.DrawGlowDot(gm.x-gm.rad*0.28, gm.y-gm.rad*0.28, gm.rad*0.45, whiteCol, 0.95, 0.88)
		}
	}

	// 2. Render Background Midnight Velvet Cosmos + Celestial Astrolabe Halo
	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			dx := (fx - eggCX) / (0.45 * minDim)
			dy := (fy - eggCY) / (0.45 * minDim)
			r := math.Hypot(dx, dy)
			ang := math.Atan2(dy, dx)

			radGlow := math.Exp(-r * r * 1.15)
			swirl := 0.5 + 0.5*math.Cos(fPetals*ang+r*3.0-phi)
			rBg := velvetCol[0] + (nebulaBlue[0]*0.65+nebulaCrimson[0]*0.35*swirl)*radGlow
			gBg := velvetCol[1] + (nebulaBlue[1]*0.65+nebulaCrimson[1]*0.35*swirl)*radGlow
			bBg := velvetCol[2] + (nebulaBlue[2]*0.65+nebulaCrimson[2]*0.35*swirl)*radGlow

			// Counter-rotating engraved 24K gold celestial astrolabe rings behind the Fabergé Egg
			// Use 36 outer ticks advancing by 3 ticks/loop (36*(ang - phi/12)) & 8 major compass stars
			// advancing by a full turn (8*(ang + phi/4)) so rotation is silky-smooth with zero Nyquist aliasing!
			if fy < plinthY {
				r1 := math.Exp(-math.Pow((r-0.86)/0.004, 2.0))
				r2 := math.Exp(-math.Pow((r-0.98)/0.003, 2.0))
				ticksOuter := math.Pow(0.5+0.5*math.Cos(36.0*ang-3.0*phi), 16.0) * math.Exp(-math.Pow((r-0.92)/0.045, 2.0))
				compassInner := math.Pow(0.5+0.5*math.Cos(8.0*ang+2.0*phi), 24.0) * math.Exp(-math.Pow((r-0.86)/0.022, 2.0))
				halo := (0.28*r1 + 0.22*r2 + 0.20*ticksOuter + 0.32*compassInner) * common.Smoothstep(plinthY, plinthY-0.18*fh, fy)
				rBg += gold24k[0] * halo
				gBg += gold24k[1] * halo
				bBg += gold24k[2] * halo
			}
			rc.SetPixelLinear(x, y, rBg, gBg, bBg, 1.0)
		}
	}

	// 3. Draw Background 3D Orbiting Filigree Ribbons (zDepth < 0, behind the Fabergé Egg!)
	drawRibbonList(backRibbons, backGems)

	// 4. Parallel 2x2 SSAA Ray-Tracing of the 3D Imperial Fabergé Egg & 24K Gold Chalice Pedestal!
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	yMaxUpper := int(math.Ceil(plinthY))
	chunk := (yMaxUpper + workers - 1) / workers
	subOffsets := [2]float64{0.25, 0.75}

	for wk := 0; wk < workers; wk++ {
		y0 := wk * chunk
		y1 := y0 + chunk
		if y0 >= yMaxUpper {
			break
		}
		if y1 > yMaxUpper {
			y1 = yMaxUpper
		}
		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for y := startY; y < endY; y++ {
				fyBase := float64(y)
				for x := 0; x < w; x++ {
					fxBase := float64(x)
					var accR, accG, accB, hitCount float64

					for syIdx := 0; syIdx < 2; syIdx++ {
						py := fyBase + subOffsets[syIdx]
						for sxIdx := 0; sxIdx < 2; sxIdx++ {
							px := fxBase + subOffsets[sxIdx]

							hitEgg, uLon, vLat, nx, ny, nz, zH := evalEggSurface(px, py)
							if hitEgg {
								rE, gE, bE := shadeFabergeEgg(uLon, vLat, nx, ny, nz, zH)
								edgeAA := common.Smoothstep(0.0, 0.12, zH)
								accR += rE * edgeAA
								accG += gE * edgeAA
								accB += bE * edgeAA
								hitCount += edgeAA
								continue
							}

							hitStand, rS, gS, bS := evalChaliceStand(px, py)
							if hitStand {
								accR += rS
								accG += gS
								accB += bS
								hitCount += 1.0
							}
						}
					}

					if hitCount > 0 {
						alpha := hitCount * 0.25
						rc.BlendPixelLinear(x, y, accR/hitCount, accG/hitCount, accB/hitCount, alpha, render.BlendOver)
					}
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 5. Crown Imperial Finial (24K Gold Cross & Rose-Cut Diamond Orb aligned with the Egg's Precessing 3D Axis!)
	axisUX := math.Sin(tiltAng)
	axisUY := -math.Cos(tiltAng)
	apexX := eggCX + axisUX*eggRY
	apexY := eggCY + axisUY*eggRY
	goldColRGBA := color.RGBA{250, 202, 72, 255}
	whiteRGBA := color.RGBA{255, 252, 248, 255}
	rubyRGBA := color.RGBA{232, 32, 76, 255}

	rc.DrawGlowDot(apexX+axisUX*6.0*scale, apexY+axisUY*6.0*scale, 7.5*scale, goldColRGBA, 0.95, 0.82)
	rc.DrawGlowDot(apexX+axisUX*6.0*scale, apexY+axisUY*6.0*scale, 4.8*scale, rubyRGBA, 0.95, 0.82)
	rc.DrawGlowDot(apexX+axisUX*15.0*scale, apexY+axisUY*15.0*scale, 3.8*scale, goldColRGBA, 0.95, 0.85)
	rc.DrawGlowDot(apexX+axisUX*15.0*scale, apexY+axisUY*15.0*scale, 2.0*scale, whiteRGBA, 0.98, 0.90)

	// 6. Draw Foreground 3D Orbiting Filigree Ribbons (zDepth >= 0, sweeping in front of the Fabergé Egg!)
	drawRibbonList(frontRibbons, frontGems)

	// 7. Polished Midnight Obsidian Altar Reflection Below plinthY (y in [yMaxUpper, h))
	snapImg := c.Img()
	rc.ResolveToneMapped(snapImg, 1.0, 2.2)

	for y := yMaxUpper; y < h; y++ {
		fy := float64(y) + 0.5
		pDepth := (fy - plinthY) / math.Max(1.0, fh-plinthY)
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			wave := math.Sin((fy-plinthY)*0.55/scale+fx*0.025/scale-phi) * (1.2 + 5.0*pDepth) * scale
			sx := common.ConstrainInt(int(math.Round(fx+wave)), 0, w-1)
			sy := common.ConstrainInt(int(math.Round(plinthY-(fy-plinthY)*1.28)), 0, yMaxUpper-1)

			cRGBA := snapImg.RGBAAt(sx, sy)
			lr, lg, lb, _ := render.ColorToLinearRGB(cRGBA)

			fresnel := (0.28 + 0.36*math.Pow(1.0-pDepth, 1.5)) * common.Smoothstep(1.0, 0.04, pDepth)
			horizonRim := math.Exp(-math.Pow((fy-plinthY)/(1.4*scale), 2.0)) *
				math.Exp(-math.Pow((fx-eggCX)/(0.38*fw), 2.0))

			rW := velvetCol[0]*0.75 + lr*fresnel + gold24k[0]*0.62*horizonRim
			gW := velvetCol[1]*0.75 + lg*fresnel + gold24k[1]*0.56*horizonRim
			bW := velvetCol[2]*0.75 + lb*fresnel + gold24k[2]*0.42*horizonRim
			rc.SetPixelLinear(x, y, rW, gW, bW, 1.0)
		}
	}

	rc.ApplyBloom(0.76, 3.8*scale, 0.15)
	rc.ApplyVignette(0.20)
	c.MergeRenderCanvas(rc)
}
