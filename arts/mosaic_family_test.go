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
		{"KintsugiPorcelain", NewKintsugiPorcelain(48, 4)},
		{"WaterCaustics", NewWaterCaustics(14, 1.0)},
		{"AgateGeode", NewAgateGeode(16, 18)},
		{"CelestialArmillary", NewCelestialArmillary(4, 8)},
		{"SoapIridescence", NewSoapIridescence(24, 480.0)},
		{"MalachiteBanding", NewMalachiteBanding(8, 16)},
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
