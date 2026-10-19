package common

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHexColor(t *testing.T) {
	tests := []struct{
		name string
		s string
		wanterr error
		want color.RGBA
	}{
		{name: "testcase1", s: "#112233", wanterr: nil, want: color.RGBA{R: 17, G: 34, B: 51, A: 255}},
		{name: "testcase2", s: "#123", wanterr: nil, want: color.RGBA{R: 17, G: 34, B: 51, A: 255}},
		{name: "testcase3", s: "#000233", wanterr: nil, want: color.RGBA{R: 0, G: 2, B: 51, A: 255}},
		{name: "testcase4", s: "#FFFFFFFF", wanterr: nil, want: color.RGBA{R: 255, G: 255, B: 255, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHexColor(tt.s)
			assert.Equal(t, tt.wanterr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithAlphaAndPremultiply(t *testing.T) {
	c := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	nrgba := WithAlpha(c, 128)
	assert.Equal(t, color.NRGBA{R: 200, G: 100, B: 50, A: 128}, nrgba)

	pm := PremultiplyRGBA(c, 128)
	assert.Equal(t, uint8(128), pm.A)
	assert.LessOrEqual(t, pm.R, pm.A)
	assert.LessOrEqual(t, pm.G, pm.A)
	assert.LessOrEqual(t, pm.B, pm.A)
}

func TestPerceptualColorBlendingAndGradient(t *testing.T) {
	c1 := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	c2 := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	assert.Equal(t, c1, LerpColorOklab(c1, c2, 0.0))
	assert.Equal(t, c2, LerpColorOklab(c1, c2, 1.0))
	assert.Equal(t, c1, LerpColorHCL(c1, c2, 0.0))
	assert.Equal(t, c2, LerpColorHCL(c1, c2, 1.0))
	assert.Equal(t, c1, LerpColorLab(c1, c2, 0.0))
	assert.Equal(t, c2, LerpColorLab(c1, c2, 1.0))
	assert.Equal(t, c1, LerpColorLinear(c1, c2, 0.0))
	assert.Equal(t, c2, LerpColorLinear(c1, c2, 1.0))

	mid := LerpColorOklab(c1, c2, 0.5)
	assert.Equal(t, uint8(255), mid.A)
	assert.Greater(t, mid.R, uint8(50))
	assert.Greater(t, mid.B, uint8(50))

	g := NewGradient(Plasma...)
	assert.Equal(t, Plasma[0], g.SampleOklab(0.0))
	assert.Equal(t, Plasma[len(Plasma)-1], g.SampleOklab(1.0))
	assert.Equal(t, Plasma[0], g.SampleHCL(-0.5))
	assert.Equal(t, Plasma[len(Plasma)-1], g.SampleLinear(1.5))

	gs := NewGradientWithStops(
		GradientStop{Pos: 1.0, Color: White},
		GradientStop{Pos: 0.0, Color: Black},
	)
	assert.Equal(t, Black, gs.SampleOklab(0.0))
	assert.Equal(t, White, gs.SampleOklab(1.0))
}

func TestCosinePaletteAndGenerators(t *testing.T) {
	c := CosineRainbow.Sample(0.25)
	assert.Equal(t, uint8(255), c.A)

	soft := GenerateSoftPalette(5)
	assert.Len(t, soft, 5)
	warm := GenerateWarmPalette(5)
	assert.Len(t, warm, 5)
}