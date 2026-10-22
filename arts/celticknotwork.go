package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
)

type celticKnotwork struct {
	gridSize    int
	ribbonWidth float64
}

// NewCelticKnotwork returns a celticKnotwork engine that renders an illuminated
// Book-of-Kells-inspired interlaced knotwork tapestry with exact alternating
// over-under strand crossings, umbral occlusion shadows, and gilded borders.
func NewCelticKnotwork(gridSize int, ribbonWidth float64) *celticKnotwork {
	if gridSize <= 0 {
		gridSize = 8
	}
	if ribbonWidth <= 0 {
		ribbonWidth = 1.0
	}
	return &celticKnotwork{
		gridSize:    gridSize,
		ribbonWidth: ribbonWidth,
	}
}

// Generative renders the interlaced Celtic knotwork illuminated manuscript panel.
func (ck *celticKnotwork) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Book of Kells Illuminated Vellum, Verdigris, Lapis, Kermes & Gold Leaf Palette
	obsidianVellum := color.RGBA{R: 14, G: 18, B: 26, A: 255}
	ironGallDark := color.RGBA{R: 8, G: 10, B: 16, A: 255}
	kellsGold := color.RGBA{R: 238, G: 184, B: 64, A: 255}
	ivoryCord := color.RGBA{R: 255, G: 244, B: 206, A: 255}
	verdigrisTeal := color.RGBA{R: 42, G: 158, B: 142, A: 255}
	kermesCrimson := color.RGBA{R: 204, G: 58, B: 48, A: 255}
	insularLapis := color.RGBA{R: 38, G: 96, B: 172, A: 255}

	rc.FillBackground(obsidianVellum)
	rc.SetBlendMode(render.BlendOver)

	n := ck.gridSize
	if n < 5 {
		n = 8
	}
	if n%2 != 0 {
		n++
	}

	margin := 38.0 * scale
	boardW := w - 2.0*margin
	boardH := h - 2.0*margin
	cellW := boardW / float64(n)
	cellH := boardH / float64(n)
	ribW := math.Min(cellW, cellH) * 0.34 * ck.ribbonWidth

	// 1. Illuminated Manuscript Double Border Frame
	frameRect := []render.Point{
		render.Pt(margin*0.52, margin*0.52),
		render.Pt(w-margin*0.52, margin*0.52),
		render.Pt(w-margin*0.52, h-margin*0.52),
		render.Pt(margin*0.52, h-margin*0.52),
	}
	rc.DrawCloisonneCell(
		frameRect,
		obsidianVellum, ironGallDark, obsidianVellum,
		ironGallDark, kellsGold, ivoryCord, 2.4*scale,
	)

	// Break-wall definitions on the primary grid:
	// horizWall[y][x] is true if there is a horizontal mirror wall along row y from x to x+1
	// vertWall[y][x] is true if there is a vertical mirror wall along col x from y to y+1
	horizWall := make([][]bool, n+1)
	vertWall := make([][]bool, n)
	for y := 0; y <= n; y++ {
		horizWall[y] = make([]bool, n)
	}
	for y := 0; y < n; y++ {
		vertWall[y] = make([]bool, n+1)
	}

	// Outer boundary walls (all true so strands bounce back cleanly at the border)
	for x := 0; x < n; x++ {
		horizWall[0][x] = true
		horizWall[n][x] = true
	}
	for y := 0; y < n; y++ {
		vertWall[y][0] = true
		vertWall[y][n] = true
	}

	// Symmetric internal break walls creating a outer border band + central Celtic cross + 4 corner knots
	for i := 1; i < n-1; i++ {
		if i != n/2 && i != n/2-1 {
			horizWall[1][i] = true
			horizWall[n-1][i] = true
			vertWall[i][1] = true
			vertWall[i][n-1] = true
		}
	}
	// Central medallion break walls
	mid := n / 2
	horizWall[mid-1][mid-1] = true
	horizWall[mid-1][mid] = true
	horizWall[mid+1][mid-1] = true
	horizWall[mid+1][mid] = true
	vertWall[mid-1][mid-1] = true
	vertWall[mid][mid-1] = true
	vertWall[mid-1][mid+1] = true
	vertWall[mid][mid+1] = true

	// Helper to draw a single 4-layer illuminated manuscript ribbon segment from pStart -> pCtrl -> pEnd
	drawRibbonSegment := func(pts []render.Point, bodyCol color.RGBA, drawShadow bool) {
		smooth := render.CatmullRomSpline(pts, 10)
		if drawShadow {
			// Cast dark umbral shadow across the under-strand at the crossing
			rc.DrawCapsule(smooth, ribW*1.58, ironGallDark, 0.94)
		} else {
			rc.DrawCapsule(smooth, ribW*1.32, ironGallDark, 0.96)
		}
		// Saturated mineral pigment / gold ribbon body
		rc.DrawCapsule(smooth, ribW, bodyCol, 1.0)
		// Inner gilded trim edges
		rc.DrawCapsule(smooth, ribW*0.64, ironGallDark, 0.85)
		rc.DrawCapsule(smooth, ribW*0.44, bodyCol, 1.0)
		// Central white-gold cord highlight
		rc.DrawCapsule(smooth, ribW*0.16, ivoryCord, 0.90)
	}

	strandPalette := []color.RGBA{kellsGold, verdigrisTeal, kermesCrimson, insularLapis}

	// In Mercat's grid construction, each cell (gx, gy) in [0..n-1] x [0..n-1] has 4 edge midpoints:
	// Top (gx+0.5, gy), Right (gx+1, gy+0.5), Bottom (gx+0.5, gy+1), Left (gx, gy+0.5).
	// Two diagonal strands pass through the center of cell (gx, gy):
	// - Diagonal A connects Top <-> Right and Bottom <-> Left (or crosses Top-Bottom / Left-Right if no walls)
	// More specifically: consider each interior crossing at edge midpoints!
	// Even simpler and 100% exact: consider the n x n cells. Inside each cell (gx, gy), two strands cross
	// at the cell center (cx, cy):
	//   Strand 1 (posSlope): from Top-Left corner (gx, gy) toward Bottom-Right corner (gx+1, gy+1)
	//   Strand 2 (negSlope): from Bottom-Left corner (gx, gy+1) toward Top-Right corner (gx+1, gy)
	//Wait: if strands pass through cell centers and cross at grid edge midpoints, let's trace the exact segments inside each cell (gx, gy)!
	// Inside cell (gx, gy), the 4 edge midpoints are:
	//   M_top   = (x0 + 0.5*cellW, y0)
	//   M_right = (x0 + cellW,     y0 + 0.5*cellH)
	//   M_bot   = (x0 + 0.5*cellW, y0 + cellH)
	//   M_left  = (x0,             y0 + 0.5*cellH)
	// If there are NO walls on the 4 edges of cell (gx, gy), two straight diagonal strands cross at the cell center C = (x0 + 0.5*cellW, y0 + 0.5*cellH):
	//    Wait: M_left to M_right is horizontal, M_top to M_bot is vertical!
	//   If we rotate the grid by 45° (or connect corner-to-corner vs edge-midpoints):
	//   Wait: if we connect M_left <-> M_right (horizontal) and M_top <-> M_bot (vertical) when unblocked, and when a wall blocks an edge (say Top is a wall), M_left curves smoothly to M_bot and M_right curves to... wait: if Top is blocked, there is no strand crossing Top!
	//   Let's check the classic **Truchet-Knotwork with Over-Under Crossings**!
	//   In a Truchet-Knotwork grid of n x n cells, EVERY cell has all 4 edge midpoints (M_top, M_right, M_bot, M_left) active!
	//   Each cell is assigned one of 3 tile types:
	//     - Type 0 (Crossing `+` rotated 45° or straight `+` with curved S-inflection): M_top crosses over/under M_bot & M_left <-> M_right!
	//     - Type 1 (Turn `/` Arc): M_top curves to M_left, and M_bot curves to M_right!
	//     - Type 2 (Turn `\` Arc): M_top curves to M_right, and M_bot curves to M_left!
	//   And on the outer boundary of the n x n grid, semicircular U-turn caps connect adjacent boundary ports (e.g. M_top of cell `2k` curves smoothly into M_top of cell `2k+1`) so **every single strand is a 100% closed, continuous, unbroken loop with zero dead ends**!!

	// Let's verify that!
	// If inside the n x n grid (where n is even, e.g. n = 8):
	//   1. Every interior cell (gx, gy) connects its 4 edge midpoints (M_top, M_right, M_bot, M_left) in pairs (either Crossing, Turn `/`, or Turn `\`).
	//   2. Along the 4 outer boundaries of the n x n grid (`gx = 0..n-1` in pairs `(0,1), (2,3), (4,5), (6,7)`):
	//      - Top boundary: M_top(2k, 0) connects via a semicircular arch to M_top(2k+1, 0)!
	//      - Bottom boundary: M_bot(2k, n-1) connects via a semicircular arch to M_bot(2k+1, n-1)!
	//      - Left boundary: M_left(0, 2k) connects via a semicircular arch to M_left(0, 2k+1)!
	//      - Right boundary: M_right(n-1, 2k) connects via a semicircular arch to M_right(n-1, 2k+1)!
	//   Because every interior edge is shared by two cells and every boundary edge is paired by a U-turn arch, **100% of strands form closed, seamless, continuous interlaced loops with ZERO loose ends anywhere**!

	for gy := 0; gy < n; gy++ {
		y0 := margin + float64(gy)*cellH
		for gx := 0; gx < n; gx++ {
			x0 := margin + float64(gx)*cellW

			mTop := render.Pt(x0+0.5*cellW, y0)
			mRight := render.Pt(x0+cellW, y0+0.5*cellH)
			mBot := render.Pt(x0+0.5*cellW, y0+cellH)
			mLeft := render.Pt(x0, y0+0.5*cellH)
			center := render.Pt(x0+0.5*cellW, y0+0.5*cellH)

			// Symmetric knotwork pattern: crossings in the main weave, curved turns at symmetric focal rings
			dxCenter := math.Abs(float64(gx) - float64(n-1)*0.5)
			dyCenter := math.Abs(float64(gy) - float64(n-1)*0.5)

			tileMode := 0 // 0 = Over/Under Crossing, 1 = Turn (/), 2 = Turn (\)
			if (gx == 1 || gx == n-2) && (gy == 1 || gy == n-2) {
				if (gx == 1 && gy == 1) || (gx == n-2 && gy == n-2) {
					tileMode = 1
				} else {
					tileMode = 2
				}
			} else if dxCenter < 1.1 && dyCenter < 1.1 {
				// Central quatrefoil rosette
				if (gx+gy)%2 == 0 {
					tileMode = 1
				} else {
					tileMode = 2
				}
			}

			colA := strandPalette[(gx+gy)%len(strandPalette)]
			colB := strandPalette[(gx+gy+2)%len(strandPalette)]

			switch tileMode {
			case 1:
				// Turn (/): Top <-> Left and Bottom <-> Right
				cTL := render.Pt(x0+0.22*cellW, y0+0.22*cellH)
				cBR := render.Pt(x0+0.78*cellW, y0+0.78*cellH)
				drawRibbonSegment([]render.Point{mTop, cTL, mLeft}, colA, false)
				drawRibbonSegment([]render.Point{mBot, cBR, mRight}, colB, false)

			case 2:
				// Turn (\): Top <-> Right and Bottom <-> Left
				cTR := render.Pt(x0+0.78*cellW, y0+0.22*cellH)
				cBL := render.Pt(x0+0.22*cellW, y0+0.78*cellH)
				drawRibbonSegment([]render.Point{mTop, cTR, mRight}, colA, false)
				drawRibbonSegment([]render.Point{mBot, cBL, mLeft}, colB, false)

			default:
				// Interlaced Over/Under Crossing with sinuous Celtic S-curve inflection
				sVert := []render.Point{
					mTop,
					render.Pt(center.X+0.12*cellW, y0+0.25*cellH),
					center,
					render.Pt(center.X-0.12*cellW, y0+0.75*cellH),
					mBot,
				}
				sHoriz := []render.Point{
					mLeft,
					render.Pt(x0+0.25*cellW, center.Y-0.12*cellH),
					center,
					render.Pt(x0+0.75*cellW, center.Y+0.12*cellH),
					mRight,
				}
				if (gx+gy)%2 == 0 {
					// Horizontal under, Vertical over
					drawRibbonSegment(sHoriz, colB, false)
					drawRibbonSegment(sVert, colA, true)
				} else {
					// Vertical under, Horizontal over
					drawRibbonSegment(sVert, colA, false)
					drawRibbonSegment(sHoriz, colB, true)
				}
			}
		}
	}

	// 2. Seamless Semicircular U-Turn Caps Along All 4 Outer Borders
	for k := 0; k < n; k += 2 {
		col := strandPalette[k%len(strandPalette)]

		// Top border pair (k, k+1)
		xA := margin + (float64(k)+0.5)*cellW
		xB := margin + (float64(k+1)+0.5)*cellW
		xMid := 0.5 * (xA + xB)
		drawRibbonSegment([]render.Point{
			render.Pt(xA, margin),
			render.Pt(xA+0.15*cellW, margin-0.38*cellH),
			render.Pt(xMid, margin-0.48*cellH),
			render.Pt(xB-0.15*cellW, margin-0.38*cellH),
			render.Pt(xB, margin),
		}, col, false)

		// Bottom border pair (k, k+1)
		yBot := margin + boardH
		drawRibbonSegment([]render.Point{
			render.Pt(xA, yBot),
			render.Pt(xA+0.15*cellW, yBot+0.38*cellH),
			render.Pt(xMid, yBot+0.48*cellH),
			render.Pt(xB-0.15*cellW, yBot+0.38*cellH),
			render.Pt(xB, yBot),
		}, col, false)

		// Left border pair (k, k+1)
		yA := margin + (float64(k)+0.5)*cellH
		yB := margin + (float64(k+1)+0.5)*cellH
		yMid := 0.5 * (yA + yB)
		drawRibbonSegment([]render.Point{
			render.Pt(margin, yA),
			render.Pt(margin-0.38*cellW, yA+0.15*cellH),
			render.Pt(margin-0.48*cellW, yMid),
			render.Pt(margin-0.38*cellW, yB-0.15*cellH),
			render.Pt(margin, yB),
		}, col, false)

		// Right border pair (k, k+1)
		xRight := margin + boardW
		drawRibbonSegment([]render.Point{
			render.Pt(xRight, yA),
			render.Pt(xRight+0.38*cellW, yA+0.15*cellH),
			render.Pt(xRight+0.48*cellW, yMid),
			render.Pt(xRight+0.38*cellW, yB-0.15*cellH),
			render.Pt(xRight, yB),
		}, col, false)
	}

	c.MergeRenderCanvas(rc)
}
