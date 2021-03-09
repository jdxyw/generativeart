package common

import (
	"math"
	"testing"
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
