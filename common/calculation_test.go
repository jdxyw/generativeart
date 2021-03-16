package common

import "testing"

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
