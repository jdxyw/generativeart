package arts

import (
	"testing"

	"github.com/jdxyw/generativeart"
)

func TestMosaicFamilyGenerators(t *testing.T) {
	cases := []struct {
		name string
		eng  generativeart.Engine
	}{
		{"MosaicSun", NewMosaicSun(24, 1.0)},
		{"MosaicVortex", NewMosaicVortex(24, 1.0)},
		{"MosaicWave", NewMosaicWave(24, 1.0)},
		{"MosaicCanyon", NewMosaicCanyon(24, 1.0)},
		{"MosaicLotus", NewMosaicLotus(10, 1.0)},
		{"CloisonneEnamel", NewCloisonneEnamel(5, 10)},
		{"WoodblockWave", NewWoodblockWave(4, 32)},
		{"ArtDecoSunburst", NewArtDecoSunburst(24, 5)},
		{"StainedRoseWindow", NewStainedRoseWindow(10, 4)},
		{"EvenStreamlines", NewEvenStreamlines(10.0, 5.2)},
		{"MarchingIsolines", NewMarchingIsolines(10, 1.0)},
		{"OrigamiTessellation", NewOrigamiTessellation(8, 8, 1.0)},
		{"MondrianSubdivision", NewMondrianSubdivision(3, 5.0)},
		{"AmmannBeenker", NewAmmannBeenker(4)},
		{"SashikoWeave", NewSashikoWeave(2)},
		{"MagneticDipole", NewMagneticDipole(4, 16)},
		{"NewtonBasin", NewNewtonBasin(5, 24)},
		{"VortexFluid", NewVortexFluid(6, 10)},
		{"CyclicAutomata", NewCyclicAutomata(10, 60)},
		{"CircuitWFC", NewCircuitWFC(18, 24)},
		{"AsemicCalligraphy", NewAsemicCalligraphy(4, 3)},
		{"Buddhabrot", NewBuddhabrot(50000, 120)},
		{"LyapunovFractal", NewLyapunovFractal("AABAB", 40)},
		{"DunhuangApsara", NewDunhuangApsara(6, 4)},
		{"CelticKnotwork", NewCelticKnotwork(6, 1.0)},
		{"GyroidSurface", NewGyroidSurface(4.0, 0.085)},
		{"LichtenbergTree", NewLichtenbergTree(7, 1.0)},
		{"PersianTilework", NewPersianTilework(12, 4)},
		{"ChampleveEnamel", NewChampleveEnamel(6, 4)},
		{"HalvorsenAttractor", NewHalvorsenAttractor(80000, 1.4)},
		{"HyperbolicFuchsian", NewHyperbolicFuchsian(18, 1.0)},
		{"AbelianSandpile", NewAbelianSandpile(45000)},
		{"KuramotoWave", NewKuramotoWave(6, 40)},
		{"FordCircles", NewFordCircles(10)},
		{"QuantumBilliard", NewQuantumBilliard(28, 1.0)},
		{"RiemannZeta", NewRiemannZeta(20, 12)},
		{"HopfFibration", NewHopfFibration(3, 14)},
		{"ThomasAttractor", NewThomasAttractor(60000, 0.19)},
		{"FaradayWaves", NewFaradayWaves(12, 1.0)},
		{"MobiusLoxodrome", NewMobiusLoxodrome(10, 24)},
		{"LagrangianCoherent", NewLagrangianCoherent(10, 1.0)},
		{"HydraulicErosion", NewHydraulicErosion(4000, 1.0)},
		{"BasaltColumns", NewBasaltColumns(10, 1.0)},
		{"SandDunes", NewSandDunes(5, 1.0)},
		{"BathymetricChart", NewBathymetricChart(14, 5)},
		{"WireframeTerrain", NewWireframeTerrain(42, 1.0)},
		{"ConcentricMorph", NewConcentricMorph(12, 12, 1.0)},
		{"VortexImpasto", NewVortexImpasto(16, 2000, 1.0)},
		{"SlicedPebbles", NewSlicedPebbles(4, 2, 12.0)},
		{"HexMandala", NewHexMandala(8, 8, 1.0)},
		{"SqueegeeScrape", NewSqueegeeScrape(4, 80, 1.0)},
		{"HalftonePanels", NewHalftonePanels(22, 6.0)},
		{"NebulaPlumes", NewNebulaPlumes(3, 2000, 1.0)},
		{"AbyssalCurrents", NewAbyssalCurrents(12, 2000, 1.0)},
		{"SolarProminence", NewSolarProminence(10, 2000, 1.0)},
		{"BotanicalWhorls", NewBotanicalWhorls(24, 2000, 137.507764)},
		{"JupiterMagnetosphere", NewJupiterMagnetosphere(6, 8, 12.0)},
		{"IridescentNacre", NewIridescentNacre(5, 1600, 410.0)},
		{"SchottkyKleinian", NewSchottkyKleinian(5, 1.0)},
		{"BismuthHopper", NewBismuthHopper(10, 6, 420.0)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := generativeart.NewCanva(160, 160)
			c.FillBackground()
			c.Draw(tc.eng)
			img := c.Img()
			if img == nil || img.Bounds().Dx() != 160 || img.Bounds().Dy() != 160 {
				t.Fatalf("%s produced invalid canvas image", tc.name)
			}
		})
	}
}
