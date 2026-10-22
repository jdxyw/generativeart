package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type circuitWFC struct {
	gridSize int
	busCount int
}

// NewCircuitWFC returns a circuitWFC engine that synthesizes a gilded silicon
// wafer and cyberpunk PCB architecture using constraint-propagated octilinear
// (0/45/90-degree) trace routing, 3D beveled QFP microprocessor dies, and plated vias.
func NewCircuitWFC(gridSize, busCount int) *circuitWFC {
	if gridSize <= 0 {
		gridSize = 28
	}
	if busCount <= 0 {
		busCount = 64
	}
	return &circuitWFC{
		gridSize: gridSize,
		busCount: busCount,
	}
}

// Generative renders the gilded silicon wafer and octilinear PCB labyrinth.
func (cw *circuitWFC) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Gilded Silicon Wafer, Emerald Solder-Mask & Copper/Gold Trace Palette
	solderAbyss := color.RGBA{R: 8, G: 20, B: 26, A: 255}
	substrateTeal := color.RGBA{R: 14, G: 44, B: 52, A: 255}
	gridTrace := color.RGBA{R: 22, G: 68, B: 76, A: 255}

	copperTrace := color.RGBA{R: 212, G: 112, B: 48, A: 255}
	gildedBrass := color.RGBA{R: 244, G: 188, B: 64, A: 255}
	champagneSpec := color.RGBA{R: 255, G: 242, B: 172, A: 255}
	electricCyan := color.RGBA{R: 54, G: 214, B: 226, A: 255}
	coralSignal := color.RGBA{R: 232, G: 76, B: 62, A: 255}
	siliconDie := color.RGBA{R: 18, G: 24, B: 38, A: 255}

	rc.FillBackground(solderAbyss)
	rc.SetBlendMode(render.BlendOver)

	n := cw.gridSize
	if n < 16 {
		n = 28
	}
	margin := 24.0 * scale
	cellS := (math.Min(w, h) - 2.0*margin) / float64(n)

	// 1. Background PCB Etched Micro-Grid
	bgPath := &canvas.Path{}
	for i := 0; i <= n; i++ {
		pos := margin + float64(i)*cellS
		bgPath.MoveTo(margin, pos)
		bgPath.LineTo(w-margin, pos)
		bgPath.MoveTo(pos, margin)
		bgPath.LineTo(pos, h-margin)
	}
	rc.StrokePath(bgPath, 0.75*scale, canvas.RoundCap, canvas.RoundJoin, gridTrace, 0.55)

	// Occupied lattice nodes so WFC octilinear traces never short-circuit or cross IC packages
	occupied := make([][]bool, n+1)
	for y := 0; y <= n; y++ {
		occupied[y] = make([]bool, n+1)
	}

	type icChip struct {
		gx0, gy0, gx1, gy1 int
		title              string
		accentCol          color.RGBA
	}

	chips := []icChip{
		{n/2 - 4, n/2 - 4, n/2 + 4, n/2 + 4, "CORE", gildedBrass},
		{3, 3, 8, 8, "DSP1", electricCyan},
		{n - 8, 3, n - 3, 8, "DSP2", coralSignal},
		{3, n - 8, 8, n - 3, "MEM1", coralSignal},
		{n - 8, n - 8, n - 3, n - 3, "MEM2", electricCyan},
	}

	for _, ch := range chips {
		for gy := ch.gy0; gy <= ch.gy1; gy++ {
			for gx := ch.gx0; gx <= ch.gx1; gx++ {
				if gy >= 0 && gy <= n && gx >= 0 && gx <= n {
					occupied[gy][gx] = true
				}
			}
		}
	}

	gridPt := func(gx, gy int) render.Point {
		return render.Pt(margin+float64(gx)*cellS, margin+float64(gy)*cellS)
	}

	// Collect all IC pin start nodes around the perimeter of every chip
	type pinSeed struct {
		gx, gy int
		dx, dy int
		col    color.RGBA
	}
	var pins []pinSeed
	traceCols := []color.RGBA{gildedBrass, copperTrace, electricCyan, gildedBrass, coralSignal}

	for cIdx, ch := range chips {
		for gx := ch.gx0 + 1; gx < ch.gx1; gx++ {
			pins = append(pins,
				pinSeed{gx, ch.gy0, 0, -1, traceCols[(gx+cIdx)%len(traceCols)]},
				pinSeed{gx, ch.gy1, 0, +1, traceCols[(gx+cIdx+1)%len(traceCols)]},
			)
		}
		for gy := ch.gy0 + 1; gy < ch.gy1; gy++ {
			pins = append(pins,
				pinSeed{ch.gx0, gy, -1, 0, traceCols[(gy+cIdx)%len(traceCols)]},
				pinSeed{ch.gx1, gy, +1, 0, traceCols[(gy+cIdx+2)%len(traceCols)]},
			)
		}
	}

	rng := rand.New(rand.NewSource(202619))

	// Helper to draw a 3D metallic PCB trace with plated through-hole vias at both ends
	drawPCBTrace := func(pts []render.Point, col color.RGBA, width float64) {
		if len(pts) < 2 {
			return
		}
		rc.DrawCapsule(pts, width*1.45, solderAbyss, 0.92)
		rc.DrawCapsule(pts, width, col, 1.0)
		rc.DrawCapsule(pts, width*0.35, champagneSpec, 0.72)

		// Plated through-hole via pads at start and end
		for _, ep := range []render.Point{pts[0], pts[len(pts)-1]} {
			rc.DrawGlowDot(ep.X, ep.Y, width*1.65, col, 1.0, 0.92)
			rc.DrawGlowDot(ep.X, ep.Y, width*0.72, solderAbyss, 1.0, 0.95)
		}
	}

	// 2. Route Constraint-Propagated Octilinear (0°/45°/90°) PCB Traces from IC Pins
	for _, pin := range pins {
		cx, cy := pin.gx+pin.dx, pin.gy+pin.dy
		if cx < 0 || cx > n || cy < 0 || cy > n || occupied[cy][cx] {
			continue
		}
		occupied[cy][cx] = true
		pts := []render.Point{gridPt(pin.gx, pin.gy), gridPt(cx, cy)}

		dx, dy := pin.dx, pin.dy
		maxLen := 6 + rng.Intn(14)
		for step := 0; step < maxLen; step++ {
			// Candidate octilinear turns: straight, +45 deg, -45 deg
			cands := [][2]int{
				{dx, dy},
				{dx - dy, dy + dx},
				{dx + dy, dy - dx},
			}
			// Clamp components to [-1, 1]
			for i := range cands {
				if cands[i][0] > 1 {
					cands[i][0] = 1
				} else if cands[i][0] < -1 {
					cands[i][0] = -1
				}
				if cands[i][1] > 1 {
					cands[i][1] = 1
				} else if cands[i][1] < -1 {
					cands[i][1] = -1
				}
			}
			// Occasionally prefer a 45-degree miter turn
			if step%3 == 1 {
				cands[0], cands[1] = cands[1], cands[0]
			}

			moved := false
			for _, cd := range cands {
				nx, ny := cx+cd[0], cy+cd[1]
				if nx < 0 || nx > n || ny < 0 || ny > n || occupied[ny][nx] {
					continue
				}
				// Prevent diagonal X-crossings between adjacent cells
				if cd[0] != 0 && cd[1] != 0 {
					if occupied[cy][nx] && occupied[ny][cx] {
						continue
					}
				}
				occupied[ny][nx] = true
				cx, cy = nx, ny
				dx, dy = cd[0], cd[1]
				pts = append(pts, gridPt(cx, cy))
				moved = true
				break
			}
			if !moved {
				break
			}
		}
		drawPCBTrace(pts, pin.col, 2.6*scale)
	}

	// 3. Fill Remaining Free Lattice Corridors with Secondary WFC Bus Traces
	for extra := 0; extra < cw.busCount*2; extra++ {
		gx := rng.Intn(n + 1)
		gy := rng.Intn(n + 1)
		if occupied[gy][gx] {
			continue
		}
		occupied[gy][gx] = true
		pts := []render.Point{gridPt(gx, gy)}
		dirs := [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}, {1, 1}, {-1, 1}, {1, -1}, {-1, -1}}
		d := dirs[rng.Intn(len(dirs))]
		cx, cy := gx, gy

		for step := 0; step < 10; step++ {
			nx, ny := cx+d[0], cy+d[1]
			if nx < 0 || nx > n || ny < 0 || ny > n || occupied[ny][nx] {
				d = dirs[rng.Intn(len(dirs))]
				nx, ny = cx+d[0], cy+d[1]
				if nx < 0 || nx > n || ny < 0 || ny > n || occupied[ny][nx] {
					break
				}
			}
			if d[0] != 0 && d[1] != 0 && occupied[cy][nx] && occupied[ny][cx] {
				break
			}
			occupied[ny][nx] = true
			cx, cy = nx, ny
			pts = append(pts, gridPt(cx, cy))
		}
		if len(pts) >= 2 {
			drawPCBTrace(pts, traceCols[extra%len(traceCols)], 2.1*scale)
		} else {
			// Standalone plated test pad
			p := gridPt(gx, gy)
			rc.DrawSoftRing(p.X, p.Y, 3.2*scale, 1.4*scale, substrateTeal, 0.9)
		}
	}

	// 4. Render 3D Beveled QFP Microprocessor Packages & Internal Silicon Die Sub-Arrays
	for _, ch := range chips {
		p0 := gridPt(ch.gx0, ch.gy0)
		p1 := gridPt(ch.gx1, ch.gy1)
		pad := 4.5 * scale
		x0, y0 := p0.X+pad, p0.Y+pad
		x1, y1 := p1.X-pad, p1.Y-pad
		cx, cy := 0.5*(x0+x1), 0.5*(y0+y1)
		hw, hh := 0.5*(x1-x0), 0.5*(y1-y0)

		chipSDF := render.SDFRoundedBox(render.Pt(cx, cy), render.Pt(hw, hh), 6.0*scale)
		rc.RenderSDFDropShadow(chipSDF, render.Pt(5*scale, 6*scale), 10*scale, 2*scale, solderAbyss, 0.88)
		rc.RenderSDFBevel(chipSDF, siliconDie, solderAbyss, ch.accentCol, 9.0*scale, render.Pt(-0.7, -0.7), 1.0)

		// Clip inside the silicon die to draw internal SRAM cache sub-blocks and gold wire-bonds
		dieMargin := 10.0 * scale
		dieRect := []render.Point{
			render.Pt(x0+dieMargin, y0+dieMargin),
			render.Pt(x1-dieMargin, y0+dieMargin),
			render.Pt(x1-dieMargin, y1-dieMargin),
			render.Pt(x0+dieMargin, y1-dieMargin),
		}
		rc.PushClipPolygon(dieRect)

		// Internal 2x2 silicon core sub-array
		dw := (x1 - x0 - 2*dieMargin) * 0.44
		dh := (y1 - y0 - 2*dieMargin) * 0.44
		for sy := 0; sy < 2; sy++ {
			for sx := 0; sx < 2; sx++ {
				bx := x0 + dieMargin + float64(sx)*(dw*1.12) + 2*scale
				by := y0 + dieMargin + float64(sy)*(dh*1.12) + 2*scale
				subRect := []render.Point{
					render.Pt(bx, by),
					render.Pt(bx+dw, by),
					render.Pt(bx+dw, by+dh),
					render.Pt(bx, by+dh),
				}
				rc.FillPolygon(subRect, substrateTeal, 0.9)
				rc.DrawHatchPatch(
					[]render.Point{subRect[0], subRect[1]},
					[]render.Point{subRect[3], subRect[2]},
					6, 0.9*scale, 0.0,
					func(s, t float64) color.RGBA { return ch.accentCol },
					0.65,
				)
			}
		}
		rc.PopClip()

		// Outer gilded QFP bevel border
		border := &canvas.Path{}
		border.MoveTo(x0+dieMargin, y0+dieMargin)
		border.LineTo(x1-dieMargin, y0+dieMargin)
		border.LineTo(x1-dieMargin, y1-dieMargin)
		border.LineTo(x0+dieMargin, y1-dieMargin)
		border.Close()
		rc.StrokePath(border, 1.5*scale, canvas.RoundCap, canvas.RoundJoin, ch.accentCol, 0.95)
	}

	c.MergeRenderCanvas(rc)
}
