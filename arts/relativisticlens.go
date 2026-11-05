package arts

import (
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type relativisticLens struct {
	rings    int
	jets     int
	spinRate float64
}

// NewRelativisticLens returns a RelativisticLens generator that integrates curved
// general-relativistic null geodesics around a spinning Kerr/Schwarzschild black hole
// framed by a non-intersecting outer 3D engraved Renaissance brass armillary sphere
// and stone observatory oculus, rendering the primary equatorial accretion disk,
// gravitationally lensed upper/lower Einstein arches, critical photon sphere, and
// relativistic Doppler beaming with exact 3D depth ordering.
func NewRelativisticLens(rings, jets int, spinRate float64) *relativisticLens {
	if rings <= 0 {
		rings = 4
	}
	if jets <= 0 {
		jets = 2
	}
	if spinRate <= 0 {
		spinRate = 0.85
	}
	return &relativisticLens{
		rings:    rings,
		jets:     jets,
		spinRate: spinRate,
	}
}

// Generative renders the static RelativisticLens black-hole observatory plate.
func (rl *relativisticLens) Generative(c *generativeart.Canva) {
	rl.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping RelativisticLens animation.
func (rl *relativisticLens) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262403
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// 8-Band Spectral Kubelka-Munk Relativistic Accretion & Renaissance Observatory Palette
	photonWhiteGold := render.MixSpectralKubelkaMunk(render.PW(render.PigmentTitaniumWhite, 0.80), render.PW(render.PigmentGoldLeaf, 0.20))
	dopplerCyanBlue := render.MixSpectralKubelkaMunk(render.PW(render.PigmentCobaltTeal, 0.62), render.PW(render.PigmentTitaniumWhite, 0.38))
	accretionGold := render.MixSpectralKubelkaMunk(render.PW(render.PigmentCadmiumYellow, 0.58), render.PW(render.PigmentGoldLeaf, 0.42))
	vermilionPlasma := render.MixSpectralKubelkaMunk(render.PW(render.PigmentVermilion, 0.72), render.PW(render.PigmentCadmiumYellow, 0.28))
	redshiftUmber := render.MixSpectralKubelkaMunk(render.PW(render.PigmentBurntSienna, 0.65), render.PW(render.PigmentRawUmber, 0.35))
	cosmosIndigo := render.MixSpectralKubelkaMunk(render.PW(render.PigmentUltramarine, 0.66), render.PW(render.PigmentPineSootInk, 0.34))
	observatoryStone := render.MixSpectralKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.48), render.PW(render.PigmentRawUmber, 0.52))
	deepIntaglio := render.MixSpectralKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.90), render.PW(render.PigmentUltramarine, 0.10))

	lightDir3D := render.V3(-0.45, 0.64, 0.62).Normalize()

	// Camera views the accretion disk from 13 degrees above the equatorial plane (diskTiltX = 0.22)
	// with outer 3D Brass Armillary Rings at R = 0.55 and R = 0.60 completely outside the rDisk <= 0.35 disk wings!
	// Inner ring rotates a full +360° (-phi) and outer ring counter-rotates a full -360° (+phi) per cycle!
	diskTiltX := 0.22
	diskRollZ := -0.15

	armillaryScene := func(p render.Vec3) (float64, int, int, render.Vec3) {
		q1 := p.RotateZ(0.32 - phi).RotateX(0.44)
		ang1 := math.Atan2(q1.Y, q1.X)
		boss1 := 0.0020 * math.Cos(ang1*24.0)
		dRing1 := math.Hypot(math.Hypot(q1.X, q1.Y)-0.55, q1.Z) - (0.015 + boss1)

		q2 := p.RotateZ(-0.38 + phi).RotateY(-0.42)
		ang2 := math.Atan2(q2.Y, q2.X)
		boss2 := 0.0025 * math.Cos(ang2*24.0)
		dRing2 := math.Hypot(math.Hypot(q2.X, q2.Y)-0.60, q2.Z) - (0.014 + boss2)

		minD := dRing1
		objID := 20
		partID := 200 + (int(math.Floor((ang1+2.0*math.Pi)/(2.0*math.Pi)*24.0)) % 24)
		loc := q1
		if dRing2 < minD {
			minD = dRing2
			objID = 21
			partID = 250 + (int(math.Floor((ang2+2.0*math.Pi)/(2.0*math.Pi)*24.0)) % 24)
			loc = q2
		}
		return minD, objID, partID, loc
	}

	armillarySDF := func(p render.Vec3) float64 {
		d, _, _, _ := armillaryScene(p)
		return d
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.HatchSpacingPx = 3.5

	camPos := render.V3(0.0, 0.0, 1.56)
	const rs = 0.130 // Schwarzschild radius

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		vx := (ux - 0.5) * 2.0
		vy := (0.5 - uy) * 2.0
		rScreen := math.Hypot(vx, vy)
		angScreen := math.Atan2(vy, vx)

		// 1. Stationary Carved Stone Observatory Oculus Frame in the 4 outer corners
		if rScreen > 0.86 {
			stoneRel := render.EvaluateMoebiusStoneRelief(ne, ux, uy, 8.0, 8.0, 0.35)
			nStone := render.V3(-vx*0.55+stoneRel.NormalMod[0]*0.3, -vy*0.55+stoneRel.NormalMod[1]*0.3, 0.78).Normalize()
			illum := common.Clamp01((0.30 + 0.68*math.Max(0.0, nStone.Dot(lightDir3D))) * (0.85 + 0.15*stoneRel.ColorMod))
			sunCol := observatoryStone
			if rScreen < 0.895 {
				sunCol = accretionGold
				tick := 1.0 - common.Smoothstep(0.0, 0.14, math.Abs(math.Sin(angScreen*36.0)))
				illum = common.Clamp01(0.55 + 0.35*tick)
			}
			return render.MoebiusPixel{
				Depth:           0.82,
				ObjID:           4,
				PartID:          40 + (stoneRel.BlockID & 15),
				DetailID:        stoneRel.BlockID,
				Normal:          [3]float64{nStone.X, nStone.Y, nStone.Z},
				FlowTangent:     [2]float64{-vy / rScreen, vx / rScreen},
				Illum:           illum,
				Bounce:          0.25,
				SunColor:        sunCol,
				MidColor:        redshiftUmber,
				ShadeColor:      cosmosIndigo,
				DeepShadowColor: deepIntaglio,
				CrackInk:        stoneRel.CrackInk * 0.62,
				TickMask:        stoneRel.PockInk * 0.45,
				Stipple:         0.16 * (1.0 - illum),
			}
		}

		rd := render.V3(vx*0.56, vy*0.56, -1.0).Normalize()

		// 2. Raymarch outer 3D Brass Armillary Rings (we will compare armHit.T against diskDist for exact depth sorting!)
		armHit := render.RaymarchSDF3DBounded(armillarySDF, camPos, rd, lightDir3D, render.V3(0, 0, 0), 0.76, 3.0, 92)
		renderArmillaryHit := func() render.MoebiusPixel {
			_, objID, partID, loc := armillaryScene(armHit.Pos)
			ndotl := math.Max(0.0, armHit.Normal.Dot(lightDir3D))
			illum := common.Clamp01((0.24 + 0.76*ndotl*(0.42+0.58*armHit.Shadow)) * armHit.AO)
			inkW, chalkW, pc := render.Evaluate3DWeingartenEngraving(armillarySDF, loc, illum, 72.0)
			return render.MoebiusPixel{
				Depth:           armHit.T,
				ObjID:           objID,
				PartID:          partID,
				DetailID:        partID,
				Normal:          [3]float64{armHit.Normal.X, armHit.Normal.Y, armHit.Normal.Z},
				FlowTangent:     [2]float64{pc.E2.X, pc.E2.Y},
				Illum:           common.Clamp01(illum + 0.24*chalkW),
				Spec:            0.52 + 0.35*chalkW,
				Bounce:          0.36,
				SunColor:        photonWhiteGold,
				MidColor:        accretionGold,
				ShadeColor:      redshiftUmber,
				DeepShadowColor: deepIntaglio,
				CrackInk:        inkW * 0.56,
				Stipple:         0.12 * (1.0 - illum),
			}
		}

		// 3. Integrate Null Geodesic Ray Through Schwarzschild/Kerr Metric:
		//    d^2 r / d lambda^2 = -1.5 * rs * L^2 / r^5 * r
		pos := camPos.RotateZ(diskRollZ).RotateX(diskTiltX)
		vel := rd.RotateZ(diskRollZ).RotateX(diskTiltX)
		angMomVec := pos.Cross(vel)
		L2 := angMomVec.Dot(angMomVec)

		captured := false
		minR := 10.0
		diskCrossings := 0
		var diskRadiance [3]float64
		var diskFlowAngle float64
		var diskRingID int
		diskDist := 99.0
		var bestDoppler float64

		dt := 0.032
		for step := 0; step < 84; step++ {
			r := pos.Len()
			if r < minR {
				minR = r
			}
			if r <= rs*1.02 {
				captured = true
				break
			}
			if r > 2.4 && step > 20 {
				break
			}
			hStep := dt * common.Constrain(r*0.95, 0.16, 1.40)
			r2 := r * r
			r5 := r2 * r2 * r
			accel := pos.Scale(-1.5 * rs * L2 / math.Max(1e-6, r5))

			nextPos := pos.Add(vel.Scale(hStep)).Add(accel.Scale(0.5 * hStep * hStep))
			nextVel := vel.Add(accel.Scale(hStep)).Normalize()

			// Detect equatorial accretion disk plane crossing (Y = 0) between r_ISCO = 1.32*rs and r_out = 0.35
			// (r_out = 0.35 stays well inside the R = 0.55 brass armillary rings with a wide clearance margin!)
			if (pos.Y > 0 && nextPos.Y <= 0) || (pos.Y < 0 && nextPos.Y >= 0) {
				frac := math.Abs(pos.Y) / math.Max(1e-7, math.Abs(pos.Y)+math.Abs(nextPos.Y))
				crossPt := pos.Add(nextPos.Sub(pos).Scale(frac))
				rDisk := math.Hypot(crossPt.X, crossPt.Z)
				if rDisk >= rs*1.32 && rDisk <= 0.35 {
					diskCrossings++
					az := math.Atan2(crossPt.Z, crossPt.X)
					uRad := (rDisk - rs*1.32) / (0.35 - rs*1.32)

					// Smooth Keplerian concentric shear rings + spiral MHD waves
					spiralPhase := az*float64(rl.rings) - 14.0*math.Log(rDisk/rs) + phi
					keplerRings := 0.5 + 0.5*math.Cos(uRad*18.0*math.Pi)
					spiralWave := 0.5 + 0.5*math.Cos(spiralPhase)
					ringWave := 0.58*keplerRings + 0.42*spiralWave

					// Relativistic Doppler beaming: approaching wing is brilliant cyan-gold, receding wing is warm vermilion-amber
					orbVel := render.V3(-crossPt.Z/rDisk, 0, crossPt.X/rDisk)
					doppler := common.Constrain(1.0-0.54*rl.spinRate*orbVel.Dot(vel), 0.54, 1.68)
					radialEnv := common.Smoothstep(0.0, 0.08, uRad) * (1.0 - common.Smoothstep(0.75, 1.0, uRad))
					intensity := radialEnv * (0.54 + 0.46*ringWave) * doppler * doppler

					var bandCol, midBand [3]float64
					tDop := common.Clamp01((doppler - 0.58) / 1.05)
					for cIdx := 0; cIdx < 3; cIdx++ {
						midBand[cIdx] = vermilionPlasma[cIdx]*(1.0-tDop) + accretionGold[cIdx]*tDop
						bandCol[cIdx] = midBand[cIdx]*(1.0-0.55*tDop) + dopplerCyanBlue[cIdx]*(0.55*tDop)
					}

					weight := 1.0
					if diskCrossings > 1 {
						weight = 0.78 // Secondary lensed lower arch
					}
					for cIdx := 0; cIdx < 3; cIdx++ {
						diskRadiance[cIdx] += weight * intensity * bandCol[cIdx]
					}
					if diskCrossings == 1 {
						diskDist = camPos.Sub(crossPt).Len()
						diskRingID = int(math.Floor(uRad * 18.0))
						diskFlowAngle = math.Atan2(-crossPt.X, crossPt.Z)
						bestDoppler = doppler
					}
				}
			}

			pos = nextPos
			vel = nextVel
		}

		// Gravitationally Lensed Einstein Ring + Photon Sphere around minimum approach minR ~ 1.50 * rs:
		// Keep the radial boundary strictly circular (einsteinRadial > 0.20) while modulating the interior with spiral waves!
		einsteinRadial := math.Exp(-math.Pow((minR-rs*1.50)/(rs*0.22), 2.0))
		einsteinSpiral := 0.5 + 0.5*math.Cos(angScreen*float64(rl.rings*2)-16.0*(minR/rs)+phi)
		einsteinDoppler := common.Constrain(1.0+0.48*rl.spinRate*math.Cos(angScreen-diskRollZ), 0.55, 1.65)
		photonRing := einsteinRadial * (0.55 + 0.45*einsteinSpiral) * einsteinDoppler

		diskEnergy := diskRadiance[0] + diskRadiance[1] + diskRadiance[2]
		hasDiskHit := (diskCrossings > 0 && diskEnergy > 0.08) || einsteinRadial > 0.20

		if armHit.Hit && (!hasDiskHit || armHit.T < diskDist) && !captured {
			return renderArmillaryHit()
		}

		if hasDiskHit {
			if diskCrossings == 0 {
				bestDoppler = einsteinDoppler
				diskFlowAngle = angScreen + math.Pi*0.5
				diskRingID = int(math.Floor((minR - rs) * 42.0))
			}
			lum := common.Clamp01(0.26 + diskEnergy/2.2 + photonRing*0.65)
			var sunCol, midCol, shadeCol [3]float64
			tDop := common.Clamp01((bestDoppler - 0.58) / 1.05)
			for cIdx := 0; cIdx < 3; cIdx++ {
				sunCol[cIdx] = common.Clamp01(diskRadiance[cIdx]*0.78 + photonWhiteGold[cIdx]*(0.30+0.55*photonRing) + dopplerCyanBlue[cIdx]*0.22*tDop)
				midCol[cIdx] = vermilionPlasma[cIdx]*(1.0-0.55*tDop) + accretionGold[cIdx]*(0.55*tDop)
				shadeCol[cIdx] = redshiftUmber[cIdx]
			}
			return render.MoebiusPixel{
				Depth:           common.Constrain(diskDist, 0.9, 2.2),
				ObjID:           10,
				PartID:          100 + (diskRingID & 31),
				DetailID:        diskRingID,
				Normal:          [3]float64{math.Cos(diskFlowAngle) * 0.4, 0.85, math.Sin(diskFlowAngle) * 0.4},
				FlowTangent:     [2]float64{math.Cos(diskFlowAngle), math.Sin(diskFlowAngle)},
				Illum:           lum,
				Spec:            photonRing*0.85 + 0.35*tDop,
				Bounce:          0.42,
				SunColor:        sunCol,
				MidColor:        midCol,
				ShadeColor:      shadeCol,
				DeepShadowColor: deepIntaglio,
				CrackInk:        common.Smoothstep(0.60, 0.95, photonRing) * 0.42,
				Stipple:         0.12 * (1.0 - lum),
			}
		}

		if captured {
			// Event Horizon Black Hole Shadow Core framed by a razor-thin golden photon rim
			rimGlow := common.Smoothstep(rs*0.85, rs*1.02, minR) * 0.35
			return render.MoebiusPixel{
				Depth:           1.55,
				ObjID:           9,
				PartID:          90,
				DetailID:        90,
				Normal:          [3]float64{0, 0, 1},
				FlowTangent:     [2]float64{-math.Sin(angScreen), math.Cos(angScreen)},
				Illum:           0.04 + rimGlow,
				SunColor:        accretionGold,
				MidColor:        deepIntaglio,
				ShadeColor:      deepIntaglio,
				DeepShadowColor: deepIntaglio,
				Stipple:         0.24,
			}
		}

		if armHit.Hit {
			return renderArmillaryHit()
		}

		// 4. Background Relativistic Lensed Starfield, Polar Astrophysical Jets & Engraved Celestial Grid
		rotX := vx*math.Cos(diskRollZ) - vy*math.Sin(diskRollZ)
		rotY := vx*math.Sin(diskRollZ) + vy*math.Cos(diskRollZ)
		jetWidth := 0.042 + 0.13*math.Abs(rotY)
		jetEnv := math.Exp(-math.Pow(rotX/jetWidth, 2.0)) * common.Smoothstep(0.18, 0.45, math.Abs(rotY)) * (1.0 - common.Smoothstep(0.68, 0.85, rScreen))
		jetHelix := 0.5 + 0.5*math.Cos(rotX/jetWidth*6.0+rotY*18.0-phi*2.0)
		jetBeam := jetEnv * (0.45 + 0.55*jetHelix)

		skyGridR := 1.0 - common.Smoothstep(0.0, 0.008, math.Abs(math.Mod(rScreen, 0.18)-0.09))
		skyGridA := 1.0 - common.Smoothstep(0.0, 0.012, math.Abs(math.Sin(angScreen*12.0))*rScreen)
		skyChart := math.Max(skyGridR, skyGridA) * 0.35

		vStar := ne.Voronoi3D(vel.X*22.0, vel.Y*22.0, vel.Z*22.0, 0.85)
		starPin := 1.0 - common.Smoothstep(0.0, 0.08, vStar.F1)
		illum := common.Clamp01(0.15 + 0.68*jetBeam + 0.65*starPin + 0.24*skyChart)

		return render.MoebiusPixel{
			Depth:           95.0,
			ObjID:           1,
			PartID:          1 + int(jetBeam*6.0),
			DetailID:        vStar.CellID & 31,
			Normal:          [3]float64{0, 0, 1},
			FlowTangent:     [2]float64{-math.Sin(angScreen), math.Cos(angScreen)},
			Illum:           illum,
			Spec:            starPin*0.75 + jetBeam*0.45,
			SunColor:        dopplerCyanBlue,
			MidColor:        accretionGold,
			ShadeColor:      cosmosIndigo,
			DeepShadowColor: deepIntaglio,
			CrackInk:        math.Max(starPin*0.35, skyChart),
			Stipple:         0.16 * (1.0 - illum),
		}
	})
	c.MergeRenderCanvas(rc)
}
