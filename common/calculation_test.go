package common

import (
	"image/color"
	"math"
	"testing"
)

func TestConstrain(t *testing.T) {
	type args struct {
		val  float64
		low  float64
		high float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "testcase1", args: args{val: 1.0, low: 0.5, high: 1.5}, want: 1.0},
		{name: "testcase2", args: args{val: 0.4, low: 0.5, high: 1.5}, want: 0.5},
		{name: "testcase3", args: args{val: -1.0, low: -3.5, high: 1.5}, want: -1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Constrain(tt.args.val, tt.args.low, tt.args.high); got != tt.want {
				t.Errorf("Constrain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstrainInt(t *testing.T) {
	type args struct {
		val  int
		low  int
		high int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "testcase1", args: args{val: 256, low: 0, high: 255}, want: 255},
		{name: "testcase2", args: args{val: -1, low: 0, high: 255}, want: 0},
		{name: "testcase3", args: args{val: 100, low: 0, high: 255}, want: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConstrainInt(tt.args.val, tt.args.low, tt.args.high); got != tt.want {
				t.Errorf("ConstrainInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterpolationFunctions(t *testing.T) {
	if got := Smoothstep(0, 10, 5); math.Abs(got-0.5) > 1e-9 {
		t.Errorf("Smoothstep(0, 10, 5) = %v, want 0.5", got)
	}
	if got := Smootherstep(0, 10, 5); math.Abs(got-0.5) > 1e-9 {
		t.Errorf("Smootherstep(0, 10, 5) = %v, want 0.5", got)
	}
	if got := Gaussian(2.0, 2.0, 1.5); math.Abs(got-1.0) > 1e-9 {
		t.Errorf("Gaussian at mean = %v, want 1.0", got)
	}
	if got := Clamp01(1.8); got != 1.0 {
		t.Errorf("Clamp01(1.8) = %v, want 1.0", got)
	}
	if got := Lerp(10, 20, 0.25); math.Abs(got-12.5) > 1e-9 {
		t.Errorf("Lerp(10, 20, 0.25) = %v, want 12.5", got)
	}

	c1 := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	c2 := color.RGBA{R: 10, G: 20, B: 30, A: 255}
	if got := LerpColor(c1, c2, 2.5); got != c1 {
		t.Errorf("LerpColor clamped at >1 = %v, want %v", got, c1)
	}
	if got := LerpColor(c1, c2, -1.5); got != c2 {
		t.Errorf("LerpColor clamped at <0 = %v, want %v", got, c2)
	}
}
