package common

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDistance(t *testing.T) {
	type args struct {
		x1 float64
		y1 float64
		x2 float64
		y2 float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "testcase1", args: args{x1: 0, y1: 0, x2: 0, y2: 0}, want: 0},
		{name: "testcase2", args: args{x1: 0, y1: 3, x2: 4, y2: 0}, want: 5},
		{name: "testcase3", args: args{x1: 1, y1: 1, x2: 0, y2: 0}, want: 1.414213562},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Distance(tt.args.x1, tt.args.y1, tt.args.x2, tt.args.y2); math.Abs(got-tt.want) > 0.00001 {
				t.Errorf("Distance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolarAndHSV(t *testing.T) {
	r, theta := ConvertCartesianToPolar(0, 5)
	assert.InDelta(t, 5.0, r, 1e-6)
	assert.InDelta(t, math.Pi/2, theta, 1e-6)

	c := HSV{H: -120, S: 100, V: 100}.ToRGB(360, 100, 100)
	assert.Equal(t, uint8(255), c.A)
	assert.Equal(t, uint8(255), c.B)

	cf := HSVFromFloat(120, 1.0, 1.0)
	assert.Equal(t, color.RGBA{R: 0, G: 255, B: 0, A: 255}, cf)
}

func TestVectorAndSplines(t *testing.T) {
	z := NewVector(0, 0)
	z.Normalize() // should not panic or produce NaN
	assert.False(t, math.IsNaN(z.X))
	assert.False(t, math.IsNaN(z.Y))

	v1 := Vector{X: 3, Y: 4}
	assert.InDelta(t, 5.0, v1.Len(), 1e-9)
	n := v1.Normalized()
	assert.InDelta(t, 1.0, n.Len(), 1e-9)
	p := v1.Perp()
	assert.InDelta(t, 0.0, v1.Dot(p), 1e-9)

	pts := []Vector{{0, 0}, {10, 20}, {20, 10}, {30, 30}}
	spline := CatmullRomSpline(pts, 8)
	assert.Equal(t, (len(pts)-1)*8+1, len(spline))
	assert.InDelta(t, pts[0].X, spline[0].X, 1e-9)
	assert.InDelta(t, pts[3].X, spline[len(spline)-1].X, 1e-9)

	chaikin := ChaikinSmooth(pts, 2, false)
	assert.Greater(t, len(chaikin), len(pts))
}

func TestBlendModes(t *testing.T) {
	src := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	dst := color.RGBA{R: 50, G: 150, B: 200, A: 255}
	modes := []BlendMode{
		Add, BlendOver, BlendMultiply, BlendScreen, BlendOverlay,
		BlendSoftLight, BlendHardLight, BlendDarken, BlendLighten,
		BlendDifference, BlendGlow,
	}
	for _, m := range modes {
		out := BlendColors(src, dst, 0.7, m)
		assert.Equal(t, uint8(255), out.A)
	}

	img1 := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img2 := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img1.SetRGBA(1, 1, src)
	img2.SetRGBA(1, 1, dst)
	blended := Blend(img1, img2, BlendScreen)
	assert.Equal(t, BlendColors(src, dst, 1.0, BlendScreen), blended.RGBAAt(1, 1))
}

func TestNoiseEngine(t *testing.T) {
	ne1 := NewNoiseEngine(42)
	ne2 := NewNoiseEngine(42)

	v1 := ne1.Simplex2D(1.23, 4.56)
	v2 := ne2.Simplex2D(1.23, 4.56)
	assert.Equal(t, v1, v2)
	assert.GreaterOrEqual(t, v1, -1.0)
	assert.LessOrEqual(t, v1, 1.0)

	v01 := ne1.Simplex2D01(1.23, 4.56)
	assert.GreaterOrEqual(t, v01, 0.0)
	assert.LessOrEqual(t, v01, 1.0)

	v3 := ne1.Simplex3D(1.1, 2.2, 3.3)
	assert.GreaterOrEqual(t, v3, -1.0)
	assert.LessOrEqual(t, v3, 1.0)

	fbm := ne1.FBM2D(2.5, 3.5, 4, 2.0, 0.5)
	assert.GreaterOrEqual(t, fbm, -1.0)
	assert.LessOrEqual(t, fbm, 1.0)

	ridged := ne1.Ridged2D(2.5, 3.5, 4, 2.0, 0.5)
	assert.GreaterOrEqual(t, ridged, 0.0)
	assert.LessOrEqual(t, ridged, 1.0)

	curl := ne1.Curl2D(1.5, 2.5, 1e-3)
	assert.False(t, math.IsNaN(curl.X))
	assert.False(t, math.IsNaN(curl.Y))

	f1, f2 := ne1.Worley2D(3.7, 2.1)
	assert.GreaterOrEqual(t, f1, 0.0)
	assert.LessOrEqual(t, f1, f2)

	p1 := NewPerlinNoiseWithSeed(99)
	p2 := NewPerlinNoiseWithSeed(99)
	assert.Equal(t, p1.Noise2D(1.2, 3.4), p2.Noise2D(1.2, 3.4))

	v4 := ne1.Simplex4D(0.7, 1.4, -0.8, 2.1)
	assert.GreaterOrEqual(t, v4, -1.0)
	assert.LessOrEqual(t, v4, 1.0)

	// Verify exact periodic loop closure at phase=0.0 and phase=1.0
	loop0 := ne1.FBM4DLoop(1.8, 2.4, 0.0, 1.2, 4, 2.0, 0.5)
	loop1 := ne1.FBM4DLoop(1.8, 2.4, 1.0, 1.2, 4, 2.0, 0.5)
	assert.InDelta(t, loop0, loop1, 1e-12)

	c0 := ne1.Curl4DLoop(1.8, 2.4, 0.0, 1.2, 1e-3)
	c1 := ne1.Curl4DLoop(1.8, 2.4, 1.0, 1.2, 1e-3)
	assert.InDelta(t, c0.X, c1.X, 1e-9)
	assert.InDelta(t, c0.Y, c1.Y, 1e-9)
}
