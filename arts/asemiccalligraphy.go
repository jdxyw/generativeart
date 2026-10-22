package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type asemicCalligraphy struct {
	cols         int
	glyphsPerCol int
}

// NewAsemicCalligraphy returns an asemicCalligraphy engine that renders a
// classical Chinese Running-Cursive Script scroll (行草书法《兰亭集序》名篇长卷)
// on Zhisi-ruled Xuan paper with anatomical brush-stroke physics (顿笔/提按/折肩/出锋),
// running-script ligatures (牵丝游丝), flying-white (Feibai) dry bristles, and cinnabar seals.
func NewAsemicCalligraphy(cols, glyphsPerCol int) *asemicCalligraphy {
	if cols <= 0 {
		cols = 6
	}
	if glyphsPerCol <= 0 {
		glyphsPerCol = 5
	}
	return &asemicCalligraphy{
		cols:         cols,
		glyphsPerCol: glyphsPerCol,
	}
}

type strokeKind int

const (
	skHeng        strokeKind = iota // 横 (angled attack, slight upward arch, rounded press end)
	skShuChuilu                     // 垂露竖 (heavy top attack, firm spine, rounded bottom stop)
	skShuXuanzhen                   // 悬针竖 (heavy top attack, needle-sharp bottom taper)
	skPie                           // 撇 (heavy top attack, sweeping curve, razor tip)
	skNa                            // 捺 (light entry, swelling wave, heavy foot + sharp flick)
	skDian                          // 点 (pointed entry, plump belly, brisk release)
	skZhe                           // 折 (horizontal/vertical turn with heavy corner shoulder press)
	skGou                           // 钩 (vertical/curved spine with bottom swell and sharp hook)
	skTi                            // 提 (heavy lower-left press, sharp upper-right flick)
)

type calStroke struct {
	kind strokeKind
	pts  [][2]float64 // normalized [0, 1] x [0, 1] character-cell coordinates
}

type hanziGlyph struct {
	char    string
	strokes []calStroke
}

// strokeWidthProfile returns the normalized brush width multiplier [0, 1.35] along arc-length t in [0, 1].
func strokeWidthProfile(kind strokeKind) func(t float64) float64 {
	return func(t float64) float64 {
		t = common.Constrain(t, 0, 1)
		switch kind {
		case skHeng:
			if t < 0.14 {
				return 0.45 + 0.72*math.Sin((t/0.14)*0.5*math.Pi)
			}
			if t > 0.84 {
				u := (t - 0.84) / 0.16
				return 0.85 + 0.30*math.Sin(u*math.Pi) - 0.55*u*u
			}
			return 0.92 - 0.18*math.Sin(((t-0.14)/0.70)*math.Pi)

		case skShuChuilu:
			if t < 0.14 {
				return 0.50 + 0.75*math.Sin((t/0.14)*0.5*math.Pi)
			}
			if t > 0.84 {
				u := (t - 0.84) / 0.16
				return 0.88 + 0.26*math.Sin(u*math.Pi) - 0.48*u*u
			}
			return 0.96 - 0.14*math.Sin(((t-0.14)/0.70)*math.Pi)

		case skShuXuanzhen:
			if t < 0.14 {
				return 0.52 + 0.72*math.Sin((t/0.14)*0.5*math.Pi)
			}
			u := (t - 0.14) / 0.86
			return 1.24 * math.Pow(1.0-u, 1.35)

		case skPie:
			if t < 0.12 {
				return 0.55 + 0.74*math.Sin((t/0.12)*0.5*math.Pi)
			}
			u := (t - 0.12) / 0.88
			return 1.29 * math.Pow(1.0-u, 1.55)

		case skNa:
			// 一波三折: light entry -> swelling wave -> heavy foot at t=0.76 -> sharp flick
			if t < 0.76 {
				u := t / 0.76
				return 0.28 + 1.08*math.Pow(u, 1.25)
			}
			u := (t - 0.76) / 0.24
			return 1.36 * math.Pow(1.0-u, 1.65)

		case skDian:
			if t < 0.65 {
				u := t / 0.65
				return 0.22 + 1.12*math.Sin(u*0.5*math.Pi)
			}
			u := (t - 0.65) / 0.35
			return 1.34 * math.Cos(u*0.5*math.Pi)

		case skZhe:
			// Swells at start (t~0.1), corner pivot (t~0.48), and end (t~0.88)
			cornerBump := 0.42 * math.Exp(-math.Pow((t-0.46)/0.11, 2.0))
			endBump := 0.24 * math.Exp(-math.Pow((t-0.88)/0.10, 2.0))
			taper := 1.0
			if t > 0.90 {
				taper = (1.0 - t) / 0.10
			}
			return (0.76 + cornerBump + endBump) * taper

		case skGou:
			// Spine with heavy swell at hook bend (t~0.76) and sharp flick to 0
			if t < 0.12 {
				return 0.50 + 0.65*(t/0.12)
			}
			if t < 0.78 {
				u := (t - 0.12) / 0.66
				return 0.92 + 0.34*math.Sin(u*math.Pi)
			}
			u := (t - 0.78) / 0.22
			return 1.22 * math.Pow(1.0-u, 1.6)

		case skTi:
			if t < 0.18 {
				return 0.48 + 0.82*(t/0.18)
			}
			u := (t - 0.18) / 0.82
			return 1.30 * math.Pow(1.0-u, 1.6)

		default:
			return render.TaperCalligraphic(t)
		}
	}
}

// buildLantingScrollGlyphs returns 30 classical Chinese characters from Wang Xizhi's
// Lantingji Xu (《兰亭集序》) & classical shan-shui poetry, arranged in 6 vertical columns of 5:
// Col 1 (Right): 永 和 九 年 春
// Col 2:         会 于 山 阴 亭
// Col 3:         崇 山 茂 林 竹
// Col 4:         清 流 映 左 右
// Col 5:         天 朗 气 和 畅
// Col 6 (Left):  明 月 松 风 云
func buildLantingScrollGlyphs() []hanziGlyph {
	return []hanziGlyph{
		// ==================== Column 1 (Rightmost): 永 和 九 年 春 ====================
		{char: "永", strokes: []calStroke{
			{skDian, [][2]float64{{0.46, 0.10}, {0.54, 0.18}}},
			{skGou, [][2]float64{{0.30, 0.28}, {0.52, 0.25}, {0.51, 0.86}, {0.38, 0.78}}},
			{skZhe, [][2]float64{{0.18, 0.48}, {0.36, 0.44}, {0.26, 0.68}, {0.14, 0.80}}},
			{skPie, [][2]float64{{0.78, 0.34}, {0.64, 0.46}, {0.54, 0.54}}},
			{skNa, [][2]float64{{0.52, 0.50}, {0.68, 0.68}, {0.88, 0.82}}},
		}},
		{char: "和", strokes: []calStroke{
			{skPie, [][2]float64{{0.44, 0.16}, {0.30, 0.22}, {0.16, 0.26}}},
			{skHeng, [][2]float64{{0.12, 0.44}, {0.32, 0.41}, {0.48, 0.39}}},
			{skShuChuilu, [][2]float64{{0.33, 0.22}, {0.33, 0.56}, {0.32, 0.88}}},
			{skPie, [][2]float64{{0.32, 0.44}, {0.22, 0.62}, {0.12, 0.78}}},
			{skDian, [][2]float64{{0.36, 0.52}, {0.46, 0.62}}},
			{skShuChuilu, [][2]float64{{0.56, 0.36}, {0.57, 0.60}, {0.58, 0.78}}},
			{skZhe, [][2]float64{{0.56, 0.36}, {0.84, 0.33}, {0.82, 0.58}, {0.80, 0.76}}},
			{skHeng, [][2]float64{{0.58, 0.76}, {0.70, 0.75}, {0.82, 0.74}}},
		}},
		{char: "九", strokes: []calStroke{
			{skPie, [][2]float64{{0.48, 0.14}, {0.42, 0.48}, {0.28, 0.74}, {0.12, 0.88}}},
			{skGou, [][2]float64{{0.18, 0.38}, {0.62, 0.32}, {0.58, 0.76}, {0.72, 0.84}, {0.88, 0.82}, {0.86, 0.64}}},
		}},
		{char: "年", strokes: []calStroke{
			{skPie, [][2]float64{{0.38, 0.10}, {0.28, 0.22}, {0.18, 0.34}}},
			{skHeng, [][2]float64{{0.32, 0.22}, {0.54, 0.19}, {0.76, 0.17}}},
			{skHeng, [][2]float64{{0.26, 0.42}, {0.52, 0.39}, {0.74, 0.37}}},
			{skShuChuilu, [][2]float64{{0.30, 0.42}, {0.30, 0.54}, {0.29, 0.62}}},
			{skHeng, [][2]float64{{0.12, 0.64}, {0.50, 0.60}, {0.88, 0.57}}},
			{skShuXuanzhen, [][2]float64{{0.54, 0.20}, {0.54, 0.58}, {0.53, 0.92}}},
		}},
		{char: "春", strokes: []calStroke{
			{skHeng, [][2]float64{{0.24, 0.20}, {0.50, 0.17}, {0.76, 0.15}}},
			{skHeng, [][2]float64{{0.28, 0.32}, {0.50, 0.30}, {0.72, 0.28}}},
			{skHeng, [][2]float64{{0.14, 0.45}, {0.50, 0.42}, {0.86, 0.39}}},
			{skPie, [][2]float64{{0.50, 0.10}, {0.42, 0.42}, {0.26, 0.66}, {0.10, 0.82}}},
			{skNa, [][2]float64{{0.54, 0.40}, {0.72, 0.62}, {0.90, 0.78}}},
			{skShuChuilu, [][2]float64{{0.36, 0.58}, {0.36, 0.74}, {0.37, 0.88}}},
			{skZhe, [][2]float64{{0.36, 0.58}, {0.66, 0.55}, {0.65, 0.74}, {0.64, 0.88}}},
			{skHeng, [][2]float64{{0.38, 0.72}, {0.50, 0.71}, {0.63, 0.70}}},
			{skHeng, [][2]float64{{0.38, 0.86}, {0.51, 0.85}, {0.64, 0.84}}},
		}},

		// ==================== Column 2: 会 于 山 阴 亭 ====================
		{char: "会", strokes: []calStroke{
			{skPie, [][2]float64{{0.50, 0.12}, {0.32, 0.34}, {0.12, 0.54}}},
			{skNa, [][2]float64{{0.50, 0.16}, {0.70, 0.36}, {0.90, 0.52}}},
			{skHeng, [][2]float64{{0.34, 0.44}, {0.50, 0.42}, {0.66, 0.40}}},
			{skHeng, [][2]float64{{0.22, 0.58}, {0.50, 0.55}, {0.78, 0.53}}},
			{skZhe, [][2]float64{{0.46, 0.58}, {0.30, 0.78}, {0.52, 0.82}, {0.70, 0.80}}},
			{skDian, [][2]float64{{0.64, 0.68}, {0.76, 0.86}}},
		}},
		{char: "于", strokes: []calStroke{
			{skHeng, [][2]float64{{0.24, 0.22}, {0.50, 0.19}, {0.76, 0.17}}},
			{skHeng, [][2]float64{{0.12, 0.48}, {0.50, 0.44}, {0.88, 0.41}}},
			{skGou, [][2]float64{{0.52, 0.20}, {0.53, 0.58}, {0.52, 0.86}, {0.38, 0.78}}},
		}},
		{char: "山", strokes: []calStroke{
			{skShuChuilu, [][2]float64{{0.50, 0.16}, {0.50, 0.52}, {0.49, 0.78}}},
			{skZhe, [][2]float64{{0.20, 0.42}, {0.18, 0.78}, {0.50, 0.76}, {0.80, 0.73}}},
			{skShuChuilu, [][2]float64{{0.82, 0.36}, {0.81, 0.62}, {0.80, 0.84}}},
		}},
		{char: "阴", strokes: []calStroke{
			{skZhe, [][2]float64{{0.14, 0.20}, {0.36, 0.17}, {0.26, 0.40}, {0.38, 0.54}, {0.26, 0.66}}},
			{skShuChuilu, [][2]float64{{0.16, 0.20}, {0.16, 0.56}, {0.15, 0.88}}},
			{skPie, [][2]float64{{0.52, 0.20}, {0.51, 0.56}, {0.42, 0.86}}},
			{skGou, [][2]float64{{0.52, 0.20}, {0.82, 0.17}, {0.81, 0.56}, {0.80, 0.86}, {0.68, 0.78}}},
			{skHeng, [][2]float64{{0.54, 0.42}, {0.66, 0.40}, {0.78, 0.39}}},
			{skHeng, [][2]float64{{0.52, 0.60}, {0.66, 0.58}, {0.78, 0.57}}},
		}},
		{char: "亭", strokes: []calStroke{
			{skDian, [][2]float64{{0.48, 0.08}, {0.54, 0.16}}},
			{skHeng, [][2]float64{{0.16, 0.22}, {0.50, 0.19}, {0.84, 0.17}}},
			{skShuChuilu, [][2]float64{{0.34, 0.30}, {0.35, 0.42}}},
			{skZhe, [][2]float64{{0.34, 0.30}, {0.66, 0.28}, {0.64, 0.41}}},
			{skHeng, [][2]float64{{0.36, 0.42}, {0.50, 0.41}, {0.64, 0.40}}},
			{skZhe, [][2]float64{{0.18, 0.54}, {0.82, 0.50}, {0.76, 0.62}}},
			{skHeng, [][2]float64{{0.28, 0.66}, {0.50, 0.63}, {0.72, 0.61}}},
			{skGou, [][2]float64{{0.52, 0.64}, {0.52, 0.88}, {0.38, 0.80}}},
		}},

		// ==================== Column 3: 崇 山 茂 林 竹 ====================
		{char: "崇", strokes: []calStroke{
			{skShuChuilu, [][2]float64{{0.50, 0.08}, {0.50, 0.26}}},
			{skZhe, [][2]float64{{0.24, 0.16}, {0.22, 0.28}, {0.50, 0.26}, {0.76, 0.24}}},
			{skShuChuilu, [][2]float64{{0.78, 0.14}, {0.77, 0.28}}},
			{skDian, [][2]float64{{0.48, 0.32}, {0.53, 0.39}}},
			{skZhe, [][2]float64{{0.20, 0.42}, {0.80, 0.38}, {0.74, 0.48}}},
			{skHeng, [][2]float64{{0.32, 0.52}, {0.50, 0.50}, {0.68, 0.49}}},
			{skHeng, [][2]float64{{0.18, 0.64}, {0.50, 0.61}, {0.82, 0.59}}},
			{skGou, [][2]float64{{0.52, 0.62}, {0.52, 0.90}, {0.40, 0.82}}},
			{skPie, [][2]float64{{0.34, 0.72}, {0.20, 0.86}}},
			{skDian, [][2]float64{{0.68, 0.72}, {0.82, 0.86}}},
		}},
		{char: "岭", strokes: []calStroke{
			{skShuChuilu, [][2]float64{{0.24, 0.22}, {0.24, 0.68}}},
			{skZhe, [][2]float64{{0.12, 0.42}, {0.11, 0.70}, {0.36, 0.66}}},
			{skShuChuilu, [][2]float64{{0.36, 0.38}, {0.36, 0.72}}},
			{skPie, [][2]float64{{0.64, 0.14}, {0.50, 0.38}, {0.38, 0.56}}},
			{skNa, [][2]float64{{0.64, 0.18}, {0.78, 0.38}, {0.92, 0.52}}},
			{skDian, [][2]float64{{0.60, 0.44}, {0.68, 0.52}}},
			{skZhe, [][2]float64{{0.48, 0.60}, {0.80, 0.56}, {0.64, 0.78}}},
			{skDian, [][2]float64{{0.54, 0.72}, {0.72, 0.88}}},
		}},
		{char: "茂", strokes: []calStroke{
			{skHeng, [][2]float64{{0.16, 0.22}, {0.50, 0.19}, {0.84, 0.17}}},
			{skShuChuilu, [][2]float64{{0.36, 0.10}, {0.37, 0.30}}},
			{skShuChuilu, [][2]float64{{0.64, 0.09}, {0.63, 0.29}}},
			{skHeng, [][2]float64{{0.22, 0.42}, {0.50, 0.39}, {0.78, 0.37}}},
			{skPie, [][2]float64{{0.26, 0.42}, {0.22, 0.66}, {0.12, 0.88}}},
			{skGou, [][2]float64{{0.52, 0.32}, {0.58, 0.64}, {0.74, 0.86}, {0.86, 0.90}, {0.88, 0.74}}},
			{skPie, [][2]float64{{0.78, 0.52}, {0.62, 0.72}, {0.42, 0.86}}},
			{skDian, [][2]float64{{0.72, 0.28}, {0.82, 0.36}}},
		}},
		{char: "林", strokes: []calStroke{
			{skHeng, [][2]float64{{0.12, 0.42}, {0.28, 0.39}, {0.46, 0.37}}},
			{skShuChuilu, [][2]float64{{0.32, 0.16}, {0.32, 0.54}, {0.31, 0.88}}},
			{skPie, [][2]float64{{0.31, 0.42}, {0.20, 0.64}, {0.10, 0.80}}},
			{skDian, [][2]float64{{0.35, 0.48}, {0.45, 0.58}}},
			{skHeng, [][2]float64{{0.52, 0.38}, {0.68, 0.36}, {0.86, 0.34}}},
			{skShuChuilu, [][2]float64{{0.66, 0.14}, {0.66, 0.54}, {0.66, 0.90}}},
			{skPie, [][2]float64{{0.65, 0.40}, {0.54, 0.64}, {0.42, 0.80}}},
			{skNa, [][2]float64{{0.67, 0.44}, {0.78, 0.66}, {0.92, 0.82}}},
		}},
		{char: "竹", strokes: []calStroke{
			{skPie, [][2]float64{{0.28, 0.14}, {0.20, 0.32}, {0.12, 0.46}}},
			{skHeng, [][2]float64{{0.22, 0.32}, {0.34, 0.30}, {0.46, 0.28}}},
			{skShuXuanzhen, [][2]float64{{0.34, 0.32}, {0.34, 0.62}, {0.33, 0.88}}},
			{skPie, [][2]float64{{0.64, 0.12}, {0.56, 0.30}, {0.48, 0.44}}},
			{skHeng, [][2]float64{{0.58, 0.30}, {0.72, 0.28}, {0.88, 0.26}}},
			{skGou, [][2]float64{{0.70, 0.30}, {0.70, 0.64}, {0.69, 0.88}, {0.56, 0.80}}},
		}},

		// ==================== Column 4: 清 流 映 左 右 ====================
		{char: "清", strokes: []calStroke{
			{skDian, [][2]float64{{0.16, 0.20}, {0.26, 0.28}}},
			{skDian, [][2]float64{{0.12, 0.44}, {0.22, 0.52}}},
			{skTi, [][2]float64{{0.14, 0.84}, {0.28, 0.62}}},
			{skHeng, [][2]float64{{0.42, 0.22}, {0.62, 0.19}, {0.80, 0.18}}},
			{skShuChuilu, [][2]float64{{0.60, 0.10}, {0.60, 0.44}}},
			{skHeng, [][2]float64{{0.44, 0.33}, {0.60, 0.31}, {0.78, 0.30}}},
			{skHeng, [][2]float64{{0.36, 0.45}, {0.62, 0.42}, {0.88, 0.40}}},
			{skShuChuilu, [][2]float64{{0.46, 0.54}, {0.46, 0.72}, {0.45, 0.90}}},
			{skGou, [][2]float64{{0.46, 0.54}, {0.76, 0.51}, {0.76, 0.72}, {0.75, 0.90}, {0.64, 0.82}}},
			{skHeng, [][2]float64{{0.48, 0.66}, {0.60, 0.65}, {0.74, 0.64}}},
			{skHeng, [][2]float64{{0.48, 0.77}, {0.60, 0.76}, {0.74, 0.75}}},
		}},
		{char: "流", strokes: []calStroke{
			{skDian, [][2]float64{{0.16, 0.20}, {0.26, 0.28}}},
			{skDian, [][2]float64{{0.12, 0.44}, {0.22, 0.52}}},
			{skTi, [][2]float64{{0.14, 0.84}, {0.28, 0.62}}},
			{skDian, [][2]float64{{0.58, 0.12}, {0.64, 0.20}}},
			{skHeng, [][2]float64{{0.38, 0.26}, {0.62, 0.23}, {0.86, 0.21}}},
			{skZhe, [][2]float64{{0.56, 0.26}, {0.42, 0.46}, {0.76, 0.44}}},
			{skDian, [][2]float64{{0.70, 0.34}, {0.82, 0.48}}},
			{skPie, [][2]float64{{0.44, 0.56}, {0.42, 0.74}, {0.32, 0.88}}},
			{skShuChuilu, [][2]float64{{0.58, 0.54}, {0.58, 0.72}, {0.58, 0.86}}},
			{skGou, [][2]float64{{0.72, 0.52}, {0.72, 0.82}, {0.80, 0.88}, {0.90, 0.86}, {0.88, 0.72}}},
		}},
		{char: "映", strokes: []calStroke{
			{skShuChuilu, [][2]float64{{0.14, 0.26}, {0.14, 0.54}, {0.15, 0.76}}},
			{skZhe, [][2]float64{{0.14, 0.26}, {0.34, 0.24}, {0.33, 0.52}, {0.32, 0.74}}},
			{skHeng, [][2]float64{{0.15, 0.49}, {0.24, 0.48}, {0.32, 0.47}}},
			{skHeng, [][2]float64{{0.15, 0.74}, {0.24, 0.73}, {0.32, 0.72}}},
			{skShuChuilu, [][2]float64{{0.44, 0.30}, {0.45, 0.52}}},
			{skZhe, [][2]float64{{0.44, 0.30}, {0.80, 0.27}, {0.78, 0.50}}},
			{skHeng, [][2]float64{{0.38, 0.52}, {0.62, 0.49}, {0.88, 0.47}}},
			{skPie, [][2]float64{{0.62, 0.12}, {0.60, 0.52}, {0.48, 0.76}, {0.34, 0.90}}},
			{skNa, [][2]float64{{0.62, 0.54}, {0.76, 0.72}, {0.92, 0.88}}},
		}},
		{char: "左", strokes: []calStroke{
			{skHeng, [][2]float64{{0.18, 0.32}, {0.50, 0.28}, {0.84, 0.25}}},
			{skPie, [][2]float64{{0.48, 0.12}, {0.38, 0.46}, {0.24, 0.72}, {0.10, 0.88}}},
			{skHeng, [][2]float64{{0.42, 0.56}, {0.60, 0.53}, {0.78, 0.51}}},
			{skShuChuilu, [][2]float64{{0.58, 0.55}, {0.58, 0.70}, {0.58, 0.84}}},
			{skHeng, [][2]float64{{0.32, 0.86}, {0.58, 0.83}, {0.86, 0.81}}},
		}},
		{char: "右", strokes: []calStroke{
			{skHeng, [][2]float64{{0.16, 0.32}, {0.50, 0.28}, {0.86, 0.25}}},
			{skPie, [][2]float64{{0.50, 0.12}, {0.38, 0.46}, {0.24, 0.72}, {0.10, 0.88}}},
			{skShuChuilu, [][2]float64{{0.38, 0.56}, {0.39, 0.72}, {0.40, 0.86}}},
			{skZhe, [][2]float64{{0.38, 0.56}, {0.78, 0.53}, {0.76, 0.70}, {0.74, 0.84}}},
			{skHeng, [][2]float64{{0.40, 0.84}, {0.58, 0.83}, {0.75, 0.82}}},
		}},

		// ==================== Column 5: 天 朗 气 和 畅 ====================
		{char: "天", strokes: []calStroke{
			{skHeng, [][2]float64{{0.24, 0.24}, {0.50, 0.21}, {0.76, 0.19}}},
			{skHeng, [][2]float64{{0.16, 0.46}, {0.50, 0.43}, {0.84, 0.40}}},
			{skPie, [][2]float64{{0.50, 0.22}, {0.46, 0.52}, {0.32, 0.76}, {0.12, 0.90}}},
			{skNa, [][2]float64{{0.50, 0.48}, {0.70, 0.70}, {0.90, 0.88}}},
		}},
		{char: "朗", strokes: []calStroke{
			{skDian, [][2]float64{{0.26, 0.12}, {0.32, 0.20}}},
			{skZhe, [][2]float64{{0.16, 0.24}, {0.42, 0.21}, {0.40, 0.52}}},
			{skHeng, [][2]float64{{0.18, 0.38}, {0.30, 0.37}, {0.40, 0.36}}},
			{skHeng, [][2]float64{{0.18, 0.52}, {0.30, 0.51}, {0.40, 0.50}}},
			{skTi, [][2]float64{{0.16, 0.24}, {0.16, 0.82}, {0.38, 0.70}}},
			{skDian, [][2]float64{{0.36, 0.62}, {0.44, 0.76}}},
			{skPie, [][2]float64{{0.56, 0.18}, {0.55, 0.56}, {0.46, 0.88}}},
			{skGou, [][2]float64{{0.56, 0.18}, {0.84, 0.15}, {0.83, 0.56}, {0.82, 0.88}, {0.70, 0.80}}},
			{skHeng, [][2]float64{{0.58, 0.40}, {0.70, 0.39}, {0.80, 0.38}}},
			{skHeng, [][2]float64{{0.56, 0.58}, {0.70, 0.57}, {0.80, 0.56}}},
		}},
		{char: "气", strokes: []calStroke{
			{skPie, [][2]float64{{0.36, 0.12}, {0.26, 0.28}, {0.14, 0.42}}},
			{skHeng, [][2]float64{{0.32, 0.24}, {0.56, 0.21}, {0.78, 0.19}}},
			{skHeng, [][2]float64{{0.28, 0.38}, {0.52, 0.35}, {0.72, 0.33}}},
			{skGou, [][2]float64{{0.18, 0.52}, {0.68, 0.46}, {0.64, 0.72}, {0.76, 0.86}, {0.88, 0.88}, {0.86, 0.72}}},
		}},
		{char: "清", strokes: []calStroke{
			{skDian, [][2]float64{{0.16, 0.20}, {0.26, 0.28}}},
			{skDian, [][2]float64{{0.12, 0.44}, {0.22, 0.52}}},
			{skTi, [][2]float64{{0.14, 0.84}, {0.28, 0.62}}},
			{skHeng, [][2]float64{{0.42, 0.22}, {0.62, 0.19}, {0.80, 0.18}}},
			{skShuChuilu, [][2]float64{{0.60, 0.10}, {0.60, 0.44}}},
			{skHeng, [][2]float64{{0.44, 0.33}, {0.60, 0.31}, {0.78, 0.30}}},
			{skHeng, [][2]float64{{0.36, 0.45}, {0.62, 0.42}, {0.88, 0.40}}},
			{skShuChuilu, [][2]float64{{0.46, 0.54}, {0.46, 0.72}, {0.45, 0.90}}},
			{skGou, [][2]float64{{0.46, 0.54}, {0.76, 0.51}, {0.76, 0.72}, {0.75, 0.90}, {0.64, 0.82}}},
			{skHeng, [][2]float64{{0.48, 0.66}, {0.60, 0.65}, {0.74, 0.64}}},
			{skHeng, [][2]float64{{0.48, 0.77}, {0.60, 0.76}, {0.74, 0.75}}},
		}},
		{char: "风", strokes: []calStroke{
			{skPie, [][2]float64{{0.24, 0.20}, {0.22, 0.56}, {0.12, 0.88}}},
			{skGou, [][2]float64{{0.24, 0.20}, {0.74, 0.16}, {0.72, 0.62}, {0.82, 0.86}, {0.90, 0.88}, {0.88, 0.70}}},
			{skPie, [][2]float64{{0.60, 0.34}, {0.46, 0.58}, {0.28, 0.78}}},
			{skDian, [][2]float64{{0.34, 0.42}, {0.52, 0.58}, {0.66, 0.76}}},
		}},

		// ==================== Column 6 (Leftmost): 明 月 松 间 云 ====================
		{char: "明", strokes: []calStroke{
			{skShuChuilu, [][2]float64{{0.16, 0.26}, {0.16, 0.54}, {0.16, 0.78}}},
			{skZhe, [][2]float64{{0.16, 0.26}, {0.38, 0.24}, {0.38, 0.52}, {0.37, 0.76}}},
			{skHeng, [][2]float64{{0.17, 0.50}, {0.27, 0.49}, {0.37, 0.48}}},
			{skHeng, [][2]float64{{0.17, 0.75}, {0.27, 0.74}, {0.37, 0.73}}},
			{skPie, [][2]float64{{0.54, 0.18}, {0.53, 0.56}, {0.42, 0.88}}},
			{skGou, [][2]float64{{0.54, 0.18}, {0.82, 0.15}, {0.82, 0.56}, {0.81, 0.88}, {0.68, 0.80}}},
			{skHeng, [][2]float64{{0.55, 0.40}, {0.68, 0.39}, {0.80, 0.38}}},
			{skHeng, [][2]float64{{0.54, 0.58}, {0.68, 0.57}, {0.80, 0.56}}},
		}},
		{char: "月", strokes: []calStroke{
			{skPie, [][2]float64{{0.34, 0.14}, {0.33, 0.54}, {0.20, 0.88}}},
			{skGou, [][2]float64{{0.34, 0.14}, {0.72, 0.11}, {0.72, 0.54}, {0.71, 0.88}, {0.56, 0.80}}},
			{skHeng, [][2]float64{{0.35, 0.38}, {0.52, 0.36}, {0.70, 0.35}}},
			{skHeng, [][2]float64{{0.34, 0.58}, {0.52, 0.56}, {0.70, 0.55}}},
		}},
		{char: "松", strokes: []calStroke{
			{skHeng, [][2]float64{{0.12, 0.38}, {0.28, 0.36}, {0.44, 0.34}}},
			{skShuChuilu, [][2]float64{{0.30, 0.14}, {0.30, 0.54}, {0.29, 0.88}}},
			{skPie, [][2]float64{{0.29, 0.40}, {0.18, 0.62}, {0.10, 0.78}}},
			{skDian, [][2]float64{{0.33, 0.46}, {0.43, 0.56}}},
			{skPie, [][2]float64{{0.58, 0.22}, {0.50, 0.38}, {0.42, 0.50}}},
			{skNa, [][2]float64{{0.68, 0.20}, {0.78, 0.36}, {0.90, 0.48}}},
			{skZhe, [][2]float64{{0.64, 0.46}, {0.46, 0.78}, {0.66, 0.76}, {0.82, 0.74}}},
			{skDian, [][2]float64{{0.74, 0.62}, {0.86, 0.84}}},
		}},
		{char: "间", strokes: []calStroke{
			{skDian, [][2]float64{{0.22, 0.12}, {0.30, 0.22}}},
			{skShuChuilu, [][2]float64{{0.18, 0.26}, {0.18, 0.58}, {0.17, 0.88}}},
			{skGou, [][2]float64{{0.42, 0.16}, {0.82, 0.13}, {0.82, 0.56}, {0.81, 0.88}, {0.68, 0.80}}},
			{skShuChuilu, [][2]float64{{0.38, 0.38}, {0.38, 0.56}, {0.39, 0.72}}},
			{skZhe, [][2]float64{{0.38, 0.38}, {0.62, 0.36}, {0.61, 0.54}, {0.60, 0.70}}},
			{skHeng, [][2]float64{{0.39, 0.54}, {0.50, 0.53}, {0.60, 0.52}}},
			{skHeng, [][2]float64{{0.39, 0.70}, {0.50, 0.69}, {0.60, 0.68}}},
		}},
		{char: "云", strokes: []calStroke{
			{skHeng, [][2]float64{{0.26, 0.24}, {0.50, 0.21}, {0.74, 0.19}}},
			{skHeng, [][2]float64{{0.14, 0.46}, {0.50, 0.42}, {0.86, 0.39}}},
			{skZhe, [][2]float64{{0.48, 0.45}, {0.26, 0.78}, {0.52, 0.76}, {0.76, 0.73}}},
			{skDian, [][2]float64{{0.66, 0.60}, {0.82, 0.84}}},
		}},
	}
}

// Generative renders the classical Chinese Running-Cursive Script scroll on ruled Xuan paper.
func (ac *asemicCalligraphy) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Aged Xuan Paper, Pine-Soot Sumi Ink, Silk Brocade & Cinnabar Seal Palette
	xuanPaper := color.RGBA{R: 244, G: 236, B: 214, A: 255}
	brocadeGold := color.RGBA{R: 218, G: 194, B: 144, A: 255}
	zhisiRed := color.RGBA{R: 198, G: 96, B: 78, A: 255}

	sumiDeep := color.RGBA{R: 18, G: 16, B: 20, A: 255}
	sumiMid := color.RGBA{R: 44, G: 40, B: 46, A: 255}
	sumiWash := color.RGBA{R: 112, G: 106, B: 108, A: 255}
	cinnabarSeal := color.RGBA{R: 196, G: 42, B: 36, A: 255}

	rc.FillBackground(xuanPaper)
	rc.SetBlendMode(render.BlendOver)

	// 1. Top & Bottom Silk Brocade Scroll Borders + Ruled Zhisi (朱丝栏) Columns
	borderH := 26.0 * scale
	rc.FillPolygon([]render.Point{
		render.Pt(0, 0), render.Pt(w, 0), render.Pt(w, borderH), render.Pt(0, borderH),
	}, brocadeGold, 0.55)
	rc.FillPolygon([]render.Point{
		render.Pt(0, h-borderH), render.Pt(w, h-borderH), render.Pt(w, h), render.Pt(0, h),
	}, brocadeGold, 0.55)

	numCols := ac.cols
	if numCols < 3 {
		numCols = 6
	}
	marginX := 42.0 * scale
	marginY := 42.0 * scale
	innerW := w - 2.0*marginX
	innerH := h - 2.0*marginY
	colW := innerW / float64(numCols)

	rulePath := &canvas.Path{}
	rulePath.MoveTo(marginX, marginY)
	rulePath.LineTo(w-marginX, marginY)
	rulePath.LineTo(w-marginX, h-marginY)
	rulePath.LineTo(marginX, h-marginY)
	rulePath.Close()

	for i := 1; i < numCols; i++ {
		x := marginX + float64(i)*colW
		rulePath.MoveTo(x, marginY)
		rulePath.LineTo(x, h-marginY)
	}
	rc.StrokePath(rulePath, 1.05*scale, canvas.RoundCap, canvas.RoundJoin, zhisiRed, 0.42)

	// 2. Render Classical Chinese Calligraphy Characters (Right-to-Left Columns, Top-to-Bottom)
	glyphBank := buildLantingScrollGlyphs()
	rng := rand.New(rand.NewSource(202620))
	glyphs := ac.glyphsPerCol
	if glyphs < 3 {
		glyphs = 5
	}
	cellH := innerH / float64(glyphs)
	charBoxW := colW * 0.76
	charBoxH := cellH * 0.78
	baseBrushW := math.Min(charBoxW, charBoxH) * 0.145

	seqIdx := 0
	for colIdx := numCols - 1; colIdx >= 0; colIdx-- {
		colLeftX := marginX + float64(colIdx)*colW + (colW-charBoxW)*0.5
		inkLoad := 1.0 // Freshly dipped sumi brush at top of column

		for gIdx := 0; gIdx < glyphs; gIdx++ {
			// Re-dip brush when ink gets dry after 2–3 characters
			if inkLoad < 0.42 {
				inkLoad = 0.96
			}

			cellTopY := marginY + float64(gIdx)*cellH + (cellH-charBoxH)*0.5
			// Slight organic hand-written posture shift per character
			jx := (rng.Float64() - 0.5) * 3.5 * scale
			jy := (rng.Float64() - 0.5) * 3.5 * scale

			glyph := glyphBank[seqIdx%len(glyphBank)]
			seqIdx++

			var prevStrokeEnd render.Point
			hasPrevStroke := false

			for sIdx, st := range glyph.strokes {
				pts := make([]render.Point, len(st.pts))
				for k, uv := range st.pts {
					pts[k] = render.Pt(
						colLeftX+jx+uv[0]*charBoxW,
						cellTopY+jy+uv[1]*charBoxH,
					)
				}

				// Running-script intra-character ligature (笔断意连 / 牵丝游丝) between consecutive strokes
				if hasPrevStroke {
					d := prevStrokeEnd.Dist(pts[0])
					if d > 4.0*scale && d < 0.62*charBoxW && (sIdx%2 == 1 || rng.Float64() < 0.55) {
						ctrl := prevStrokeEnd.Lerp(pts[0], 0.5).Add(render.Pt(-2.2*scale, 1.5*scale))
						rc.DrawTaperedSpline(
							[]render.Point{prevStrokeEnd, ctrl, pts[0]},
							1.45*scale,
							render.TaperBothEnds,
							func(t float64) color.RGBA { return sumiMid },
							0.52*inkLoad,
						)
					}
				}

				// Subtle Xuan paper ink diffusion halo at heavy attack when brush is wet
				if inkLoad > 0.72 && (st.kind == skDian || st.kind == skShuChuilu || st.kind == skPie) {
					rc.DrawGlowDot(pts[0].X, pts[0].Y, baseBrushW*0.72, sumiWash, 0.16, 0.25)
				}

				// Chisel-nib angled attack press-down (藏锋/露锋起笔顿笔)
				if st.kind == skHeng || st.kind == skShuChuilu || st.kind == skShuXuanzhen || st.kind == skPie {
					nibStart := pts[0].Add(render.Pt(-baseBrushW*0.24, -baseBrushW*0.18))
					nibEnd := pts[0].Add(render.Pt(baseBrushW*0.14, baseBrushW*0.12))
					rc.DrawTaperedSpline(
						[]render.Point{nibStart, nibEnd},
						baseBrushW*0.92,
						render.TaperBothEnds,
						func(t float64) color.RGBA { return sumiDeep },
						0.90,
					)
				}

				// Main anatomical stroke body
				wProf := strokeWidthProfile(st.kind)
				strokeAlpha := common.Constrain(0.76+0.22*inkLoad, 0.72, 0.98)
				rc.DrawTaperedSpline(
					pts,
					baseBrushW,
					wProf,
					func(t float64) color.RGBA {
						if inkLoad > 0.65 || t < 0.55 {
							return sumiDeep
						}
						return render.LerpOklab(sumiDeep, sumiMid, (t-0.55)*1.4)
					},
					strokeAlpha,
				)

				// Flying-white (飞白) multi-hair bristle texture along the stroke
				rc.DrawBristleStroke(pts, baseBrushW*0.86, 9, sumiDeep, 0.52, rng)

				prevStrokeEnd = pts[len(pts)-1]
				hasPrevStroke = true
				inkLoad -= 0.035
			}
		}
	}

	// 3. Carved Vermilion Cinnabar Seals (落款朱文印 + 白文印 + 引首闲章)
	drawCinnabarSeal := func(sx, sy, sw, sh float64, inverted bool) {
		rect := []render.Point{
			render.Pt(sx, sy),
			render.Pt(sx+sw, sy),
			render.Pt(sx+sw, sy+sh),
			render.Pt(sx, sy+sh),
		}
		if inverted {
			rc.FillPolygon(rect, cinnabarSeal, 0.92)
			glyph := &canvas.Path{}
			p := 3.0 * scale
			glyph.MoveTo(sx+p, sy+p)
			glyph.LineTo(sx+sw-p, sy+p)
			glyph.LineTo(sx+sw-p, sy+sh-p)
			glyph.LineTo(sx+p, sy+sh-p)
			glyph.Close()
			for r := 0; r < 2; r++ {
				ry := sy + (0.28+float64(r)*0.44)*sh
				glyph.MoveTo(sx+5*scale, ry)
				glyph.LineTo(sx+sw-5*scale, ry)
				glyph.MoveTo(sx+0.5*sw, ry-5*scale)
				glyph.LineTo(sx+0.5*sw, ry+5*scale)
			}
			rc.StrokePath(glyph, 1.5*scale, canvas.RoundCap, canvas.RoundJoin, xuanPaper, 0.92)
		} else {
			glyph := &canvas.Path{}
			glyph.MoveTo(sx, sy)
			glyph.LineTo(sx+sw, sy)
			glyph.LineTo(sx+sw, sy+sh)
			glyph.LineTo(sx, sy+sh)
			glyph.Close()
			glyph.MoveTo(sx+0.5*sw, sy+3*scale)
			glyph.LineTo(sx+0.5*sw, sy+sh-3*scale)
			glyph.MoveTo(sx+4*scale, sy+0.5*sh)
			glyph.LineTo(sx+sw-4*scale, sy+0.5*sh)
			rc.StrokePath(glyph, 2.2*scale, canvas.RoundCap, canvas.RoundJoin, cinnabarSeal, 0.92)
		}
	}

	drawCinnabarSeal(9*scale, h-marginY-72*scale, 25*scale, 30*scale, true)
	drawCinnabarSeal(9*scale, h-marginY-34*scale, 25*scale, 25*scale, false)
	drawCinnabarSeal(w-marginX+8*scale, marginY+8*scale, 22*scale, 42*scale, true)

	rc.ApplyPaperTexture(0.11, 202620)

	c.MergeRenderCanvas(rc)
}
