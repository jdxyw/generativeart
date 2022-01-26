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