package generativeart

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"image/png"
	"io"
	"math"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/jdxyw/generativeart/render"
)

// PNGRecipeKeyword is the standard PNG tEXt chunk keyword used to embed reproducible artwork metadata.
const PNGRecipeKeyword = "GenerativeArt-Recipe"

// AnimatedEngine is an optional interface implemented by engines that support
// phase-driven seamless loop or multi-frame timelapse rendering (phase in [0, 1)).
type AnimatedEngine interface {
	Engine
	GenerativeFrame(c *Canva, frameIdx, totalFrames int, phase float64)
}

type funcAnimatedEngine struct {
	fn func(c *Canva, frameIdx, totalFrames int, phase float64)
}

// NewAnimatedEngine wraps a frame generator callback into an AnimatedEngine.
func NewAnimatedEngine(fn func(c *Canva, frameIdx, totalFrames int, phase float64)) AnimatedEngine {
	return &funcAnimatedEngine{fn: fn}
}

func (fa *funcAnimatedEngine) Generative(c *Canva) {
	if fa.fn != nil {
		fa.fn(c, 0, 1, 0.0)
	}
}

func (fa *funcAnimatedEngine) GenerativeFrame(c *Canva, frameIdx, totalFrames int, phase float64) {
	if fa.fn != nil {
		fa.fn(c, frameIdx, totalFrames, phase)
	}
}

// DrawTimelapseFrames renders totalFrames frames of e into c.Frames() in parallel across
// runtime.NumCPU() worker goroutines using isolated per-worker Canva clones.
// If e implements AnimatedEngine, GenerativeFrame is invoked with phase = frameIdx / totalFrames.
// Otherwise, each frame advances the canvas deterministic seed by +1 and invokes e.Generative(c).
func (c *Canva) DrawTimelapseFrames(e Engine, totalFrames int) {
	if c == nil || e == nil || totalFrames <= 0 {
		return
	}
	c.lastRecipe = captureEngineRecipe(c, e)
	c.frames = make([]*image.RGBA, totalFrames)
	baseSeed := c.Seed()
	anim, isAnim := e.(AnimatedEngine)

	if totalFrames == 1 || !isAnim {
		// Sequential path for 1-frame renders or legacy non-AnimatedEngine generators that may touch package-level state
		for i := 0; i < totalFrames; i++ {
			phase := float64(i) / float64(totalFrames)
			c.FillBackground()
			if isAnim {
				c.SetSeedLocal(baseSeed)
				anim.GenerativeFrame(c, i, totalFrames, phase)
			} else {
				c.SetSeed(baseSeed + int64(i))
				e.Generative(c)
			}
			if len(c.postEffects) > 0 {
				c.ApplyPostEffects(c.postEffects...)
			}
			resolved := c.Img()
			frameCopy := image.NewRGBA(resolved.Bounds())
			copy(frameCopy.Pix, resolved.Pix)
			c.frames[i] = frameCopy
		}
		c.SetSeedLocal(baseSeed)
		return
	}

	// Lock-free parallel multi-core worker pool for AnimatedEngine
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	if workers > totalFrames {
		workers = totalFrames
	}

	var firstRC *render.Canvas
	var wg sync.WaitGroup
	for wIdx := 0; wIdx < workers; wIdx++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := workerID; i < totalFrames; i += workers {
				phase := float64(i) / float64(totalFrames)
				wCanva := c.CloneForWorker(baseSeed)
				anim.GenerativeFrame(wCanva, i, totalFrames, phase)
				if len(wCanva.postEffects) > 0 {
					wCanva.ApplyPostEffects(wCanva.postEffects...)
				}
				resolved := wCanva.Img()
				frameCopy := image.NewRGBA(resolved.Bounds())
				copy(frameCopy.Pix, resolved.Pix)
				c.frames[i] = frameCopy
				if i == 0 {
					firstRC = wCanva.LastRenderCanvas()
				}
			}
		}(wIdx)
	}
	wg.Wait()

	if len(c.frames) > 0 && c.frames[0] != nil {
		copy(c.img.Pix, c.frames[0].Pix)
		if firstRC != nil {
			c.lastRenderCanvas = firstRC
		}
		c.hdrDirty = false
		c.rgbaTouched = false
	}
	c.SetSeedLocal(baseSeed)
}

// DrawTimelapseFramesTemporalAA renders totalFrames frames of an AnimatedEngine with shutter-angle
// temporal supersampling (motion blur). Each frame i integrates subSamples sub-frames across
// phase in [i/totalFrames, (i + shutterFraction)/totalFrames) in 64-bit linear HDR space before
// executing post-effects and resolving to 8-bit sRGB.
func (c *Canva) DrawTimelapseFramesTemporalAA(e AnimatedEngine, totalFrames, subSamples int, shutterFraction float64) {
	if c == nil || e == nil || totalFrames <= 0 {
		return
	}
	if subSamples <= 1 || shutterFraction <= 0 {
		c.DrawTimelapseFrames(e, totalFrames)
		return
	}
	if shutterFraction > 1.0 {
		shutterFraction = 1.0
	}

	c.lastRecipe = captureEngineRecipe(c, e)
	c.frames = make([]*image.RGBA, totalFrames)
	baseSeed := c.Seed()

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	if workers > totalFrames {
		workers = totalFrames
	}

	var firstRC *render.Canvas
	var wg sync.WaitGroup
	invSub := 1.0 / float64(subSamples)

	for wIdx := 0; wIdx < workers; wIdx++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := workerID; i < totalFrames; i += workers {
				accumRC := render.NewCanvas(c.width, c.height)
				for s := 0; s < subSamples; s++ {
					subOffset := (float64(s) / float64(subSamples)) * shutterFraction
					phase := math.Mod((float64(i)+subOffset)/float64(totalFrames), 1.0)
					subCanva := c.CloneForWorker(baseSeed)
					e.GenerativeFrame(subCanva, i, totalFrames, phase)
					srcRC := subCanva.NewRenderCanvas()
					for y := 0; y < c.height; y++ {
						for x := 0; x < c.width; x++ {
							rL, gL, bL, aL := srcRC.PixelLinearAt(x, y)
							ar, ag, ab, aa := accumRC.PixelLinearAt(x, y)
							accumRC.SetPixelLinear(x, y, ar+rL*invSub, ag+gL*invSub, ab+bL*invSub, aa+aL*invSub)
						}
					}
				}
				wCanva := c.CloneForWorker(baseSeed)
				wCanva.MergeRenderCanvas(accumRC)
				if len(wCanva.postEffects) > 0 {
					wCanva.ApplyPostEffects(wCanva.postEffects...)
				}
				resolved := wCanva.Img()
				frameCopy := image.NewRGBA(resolved.Bounds())
				copy(frameCopy.Pix, resolved.Pix)
				c.frames[i] = frameCopy
				if i == 0 {
					firstRC = wCanva.LastRenderCanvas()
				}
			}
		}(wIdx)
	}
	wg.Wait()

	if len(c.frames) > 0 && c.frames[0] != nil {
		copy(c.img.Pix, c.frames[0].Pix)
		if firstRC != nil {
			c.lastRenderCanvas = firstRC
		}
		c.hdrDirty = false
		c.rgbaTouched = false
	}
	c.SetSeedLocal(baseSeed)
}

// Frames returns the sequence of frames recorded by DrawTimelapse / DrawTimelapseFrames.
func (c *Canva) Frames() []*image.RGBA {
	if c == nil {
		return nil
	}
	return c.frames
}

// buildSharedAnimationPalette constructs a high-precision 256-color palette tailored to
// the actual color distribution across srcFrames using importance-weighted diversity seeding
// and weighted K-Means centroid refinement, paired with an 18-bit (64x64x64 = 262,144) RGB
// lookup table for O(1) flicker-free frame quantization.
func buildSharedAnimationPalette(srcFrames []*image.RGBA) (color.Palette, *[262144]uint8, [][3]int) {
	return buildSharedAnimationPaletteWithLimit(srcFrames, 31)
}

func buildSharedAnimationPaletteWithLimit(srcFrames []*image.RGBA, maxOpaque int) (color.Palette, *[262144]uint8, [][3]int) {
	if maxOpaque <= 0 {
		maxOpaque = 31
	}
	if maxOpaque < 2 {
		maxOpaque = 2
	}
	if maxOpaque > 255 {
		maxOpaque = 255
	}

	type bucket struct {
		count            int
		sumR, sumG, sumB int
		avgR, avgG, avgB int
		weight           float64
	}
	var buckets [32768]bucket

	// Sample up to 8 evenly-spaced frames across the animation to populate the color histogram
	stepFrame := 1
	if len(srcFrames) > 8 {
		stepFrame = len(srcFrames) / 8
	}
	for fIdx := 0; fIdx < len(srcFrames); fIdx += stepFrame {
		pix := srcFrames[fIdx].Pix
		for i := 0; i+3 < len(pix); i += 4 {
			r, g, b := int(pix[i]), int(pix[i+1]), int(pix[i+2])
			key := ((r >> 3) << 10) | ((g >> 3) << 5) | (b >> 3)
			buckets[key].count++
			buckets[key].sumR += r
			buckets[key].sumG += g
			buckets[key].sumB += b
		}
	}

	active := make([]bucket, 0, 2048)
	for _, bk := range buckets {
		if bk.count > 0 {
			bk.avgR = bk.sumR / bk.count
			bk.avgG = bk.sumG / bk.count
			bk.avgB = bk.sumB / bk.count
			maxC := bk.avgR
			if bk.avgG > maxC {
				maxC = bk.avgG
			}
			if bk.avgB > maxC {
				maxC = bk.avgB
			}
			minC := bk.avgR
			if bk.avgG < minC {
				minC = bk.avgG
			}
			if bk.avgB < minC {
				minC = bk.avgB
			}
			chroma := float64(maxC-minC) / 255.0
			// Sub-linear frequency weighting + chroma boost preserves crisp highlights and filigree strokes
			bk.weight = math.Pow(float64(bk.count), 0.52) * (1.0 + 0.65*chroma)
			active = append(active, bk)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].weight > active[j].weight
	})

	// Multi-pass perceptual diversity seeding so all opaque slots cover the full gamut without clumping
	// (the final slot is reserved for GIF89a inter-frame delta transparency)
	centers := make([][3]int, 0, maxOpaque)
	used := make([]bool, len(active))
	thresholds := []int{1400, 650, 260, 90, 0}
	for _, minDist := range thresholds {
		if len(centers) >= maxOpaque {
			break
		}
		for i, bk := range active {
			if used[i] {
				continue
			}
			tooClose := false
			for _, c := range centers {
				dr := bk.avgR - c[0]
				dg := bk.avgG - c[1]
				db := bk.avgB - c[2]
				if 3*dr*dr+4*dg*dg+2*db*db < minDist {
					tooClose = true
					break
				}
			}
			if !tooClose {
				centers = append(centers, [3]int{bk.avgR, bk.avgG, bk.avgB})
				used[i] = true
				if len(centers) >= maxOpaque {
					break
				}
			}
		}
	}

	// 4 iterations of weighted K-Means refinement over active histogram buckets
	if len(centers) > 1 && len(active) > len(centers) {
		sumW := make([]float64, len(centers))
		sumR := make([]float64, len(centers))
		sumG := make([]float64, len(centers))
		sumB := make([]float64, len(centers))
		for iter := 0; iter < 4; iter++ {
			for k := range centers {
				sumW[k], sumR[k], sumG[k], sumB[k] = 0, 0, 0, 0
			}
			for _, bk := range active {
				bestIdx := 0
				bestDist := 1 << 30
				for k, c := range centers {
					dr := bk.avgR - c[0]
					dg := bk.avgG - c[1]
					db := bk.avgB - c[2]
					d := 3*dr*dr + 4*dg*dg + 2*db*db
					if d < bestDist {
						bestDist = d
						bestIdx = k
						if d == 0 {
							break
						}
					}
				}
				w := bk.weight
				sumW[bestIdx] += w
				sumR[bestIdx] += float64(bk.avgR) * w
				sumG[bestIdx] += float64(bk.avgG) * w
				sumB[bestIdx] += float64(bk.avgB) * w
			}
			for k := range centers {
				if sumW[k] > 0 {
					centers[k][0] = int(math.Round(sumR[k] / sumW[k]))
					centers[k][1] = int(math.Round(sumG[k] / sumW[k]))
					centers[k][2] = int(math.Round(sumB[k] / sumW[k]))
				}
			}
		}
	}

	for i := 0; len(centers) < maxOpaque && i < len(palette.Plan9); i += 4 {
		r, g, b, _ := palette.Plan9[i].RGBA()
		centers = append(centers, [3]int{int(r >> 8), int(g >> 8), int(b >> 8)})
	}

	pal := make(color.Palette, len(centers)+1)
	for i, c := range centers {
		pal[i] = color.RGBA{R: uint8(c[0]), G: uint8(c[1]), B: uint8(c[2]), A: 255}
	}
	pal[len(centers)] = color.RGBA{R: 0, G: 0, B: 0, A: 0}

	// Precompute 18-bit (64x64x64) RGB -> nearest palette index LUT in parallel
	var lut [262144]uint8
	var wg sync.WaitGroup
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	for wIdx := 0; wIdx < workers; wIdx++ {
		qrStart := (wIdx * 64) / workers
		qrEnd := ((wIdx + 1) * 64) / workers
		wg.Add(1)
		go func(q0, q1 int) {
			defer wg.Done()
			for qr := q0; qr < q1; qr++ {
				r8 := (qr << 2) | (qr >> 4)
				for qg := 0; qg < 64; qg++ {
					g8 := (qg << 2) | (qg >> 4)
					base := (qr << 12) | (qg << 6)
					for qb := 0; qb < 64; qb++ {
						b8 := (qb << 2) | (qb >> 4)
						bestIdx := 0
						bestDist := 1 << 30
						for pIdx, pc := range centers {
							dr := r8 - pc[0]
							dg := g8 - pc[1]
							db := b8 - pc[2]
							d := 3*dr*dr + 4*dg*dg + 2*db*db
							if d < bestDist {
								bestDist = d
								bestIdx = pIdx
								if d == 0 {
									break
								}
							}
						}
						lut[base|qb] = uint8(bestIdx)
					}
				}
			}
		}(qrStart, qrEnd)
	}
	wg.Wait()
	return pal, &lut, centers
}

// WriteGIF encodes the recorded timelapse frames (or the current canvas image if no
// timelapse frames were recorded) as a flicker-free looping animated GIF to w.
func (c *Canva) WriteGIF(w io.Writer, delayCentiseconds int) error {
	return c.WriteGIFWithColors(w, delayCentiseconds, 31)
}

// WriteGIFWithColors encodes the recorded timelapse frames as a looping animated GIF to w
// using up to maxOpaqueColors opaque palette entries (plus 1 transparent delta entry, in [2..255]).
func (c *Canva) WriteGIFWithColors(w io.Writer, delayCentiseconds, maxOpaqueColors int) error {
	if c == nil {
		return errors.New("nil Canva")
	}
	if delayCentiseconds <= 0 {
		delayCentiseconds = 8
	}
	srcFrames := c.frames
	if len(srcFrames) == 0 && c.img != nil {
		srcFrames = []*image.RGBA{c.Img()}
	}
	if len(srcFrames) == 0 {
		return errors.New("no frames available to encode")
	}

	sharedPal, lut, centers := buildSharedAnimationPaletteWithLimit(srcFrames, maxOpaqueColors)
	transIdx := uint8(len(sharedPal) - 1)

	outGIF := &gif.GIF{
		Image:     make([]*image.Paletted, len(srcFrames)),
		Delay:     make([]int, len(srcFrames)),
		Disposal:  make([]byte, len(srcFrames)),
		LoopCount: 0,
	}

	var wg sync.WaitGroup
	for i, rgba := range srcFrames {
		outGIF.Delay[i] = delayCentiseconds
		outGIF.Disposal[i] = gif.DisposalNone
		wg.Add(1)
		go func(frameIdx int, src *image.RGBA) {
			defer wg.Done()
			b := src.Bounds()
			fw, fh := b.Dx(), b.Dy()
			pm := image.NewPaletted(b, sharedPal)
			for y := 0; y < fh; y++ {
				srcOff := y * src.Stride
				dstOff := y * pm.Stride
				for x := 0; x < fw; x++ {
					r0 := int(src.Pix[srcOff+x*4+0])
					g0 := int(src.Pix[srcOff+x*4+1])
					b0 := int(src.Pix[srcOff+x*4+2])
					baseKey := ((r0 >> 2) << 12) | ((g0 >> 2) << 6) | (b0 >> 2)
					idx0 := lut[baseKey]
					if x > 0 {
						prevIdx := pm.Pix[dstOff+x-1]
						if prevIdx != idx0 && int(prevIdx) < len(centers) {
							pcPrev := centers[prevIdx]
							dr := r0 - pcPrev[0]
							dg := g0 - pcPrev[1]
							db := b0 - pcPrev[2]
							// Horizontal LZW run-coherence: preserve continuous scanline runs when within 1-2 LSBs
							if 3*dr*dr+4*dg*dg+2*db*db <= 220 {
								idx0 = prevIdx
							}
						}
					}
					pm.Pix[dstOff+x] = idx0
				}
			}
			outGIF.Image[frameIdx] = pm
		}(i, rgba)
	}
	wg.Wait()

	// GIF89a Inter-Frame Delta Transparency Coding (DisposalNone):
	// Any pixel in frame i > 0 whose quantized palette index or perceptual RGB matches
	// the currently displayed pixel from frames 0..i-1 is encoded as transIdx (255),
	// converting unchanged regions into long runs of 255 for 2x-5x LZW compression.
	if len(srcFrames) > 1 {
		b0 := srcFrames[0].Bounds()
		fw, fh := b0.Dx(), b0.Dy()
		canvasIdx := make([]uint8, fw*fh)
		canvasRGB := make([][3]uint8, fw*fh)
		pm0 := outGIF.Image[0]
		s0 := srcFrames[0]
		for y := 0; y < fh; y++ {
			for x := 0; x < fw; x++ {
				pos := y*fw + x
				canvasIdx[pos] = pm0.Pix[y*pm0.Stride+x]
				off := y*s0.Stride + x*4
				canvasRGB[pos] = [3]uint8{s0.Pix[off], s0.Pix[off+1], s0.Pix[off+2]}
			}
		}
		for fi := 1; fi < len(srcFrames); fi++ {
			pm := outGIF.Image[fi]
			sf := srcFrames[fi]
			if sf.Bounds().Dx() != fw || sf.Bounds().Dy() != fh {
				continue
			}
			minX, minY := fw, fh
			maxX, maxY := -1, -1
			for y := 0; y < fh; y++ {
				rowOff := y * pm.Stride
				sOff := y * sf.Stride
				posBase := y * fw
				for x := 0; x < fw; x++ {
					pos := posBase + x
					idx := pm.Pix[rowOff+x]
					if idx == canvasIdx[pos] {
						pm.Pix[rowOff+x] = transIdx
						continue
					}
					r := int(sf.Pix[sOff+x*4+0])
					g := int(sf.Pix[sOff+x*4+1])
					bl := int(sf.Pix[sOff+x*4+2])
					prev := canvasRGB[pos]
					dr := r - int(prev[0])
					dg := g - int(prev[1])
					db := bl - int(prev[2])
					if 3*dr*dr+4*dg*dg+2*db*db <= 240 {
						pm.Pix[rowOff+x] = transIdx
					} else {
						canvasIdx[pos] = idx
						canvasRGB[pos] = [3]uint8{uint8(r), uint8(g), uint8(bl)}
						if x < minX {
							minX = x
						}
						if x > maxX {
							maxX = x
						}
						if y < minY {
							minY = y
						}
						if y > maxY {
							maxY = y
						}
					}
				}
			}
			// Crop delta frame to its tight dirty bounding rectangle for additional LZW savings
			if maxX >= minX && maxY >= minY {
				outGIF.Image[fi] = pm.SubImage(image.Rect(minX, minY, maxX+1, maxY+1)).(*image.Paletted)
			} else {
				outGIF.Image[fi] = pm.SubImage(image.Rect(0, 0, 1, 1)).(*image.Paletted)
			}
		}
	}
	return gif.EncodeAll(w, outGIF)
}

// ToGIF saves the recorded timelapse frames (or current image) as an animated GIF at fpath.
func (c *Canva) ToGIF(fpath string, delayCentiseconds int) error {
	return c.ToGIFWithColors(fpath, delayCentiseconds, 31)
}

// ToGIFWithColors saves the recorded timelapse frames as an animated GIF at fpath with up to maxOpaqueColors.
func (c *Canva) ToGIFWithColors(fpath string, delayCentiseconds, maxOpaqueColors int) error {
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	if err := c.WriteGIFWithColors(f, delayCentiseconds, maxOpaqueColors); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// WriteSVG exports the canvas as an infinite-resolution vector SVG (or embedded raster
// fallback for pure pixel-splat engines) to w.
func (c *Canva) WriteSVG(w io.Writer) error {
	if c == nil {
		return errors.New("nil Canva")
	}
	rc := c.lastRenderCanvas
	if rc == nil {
		rc = render.NewCanvasFromRGBA(c.img)
	}
	return rc.WriteSVG(w)
}

// ToSVG saves the canvas as an SVG file at fpath.
func (c *Canva) ToSVG(fpath string) error {
	if c == nil {
		return errors.New("nil Canva")
	}
	rc := c.lastRenderCanvas
	if rc == nil {
		rc = render.NewCanvasFromRGBA(c.img)
	}
	return rc.SaveSVG(fpath)
}

// WritePlotterSVG exports the canvas as an AxiDraw / Inkscape pen-plotter SVG with
// color-grouped <g inkscape:groupmode="layer"> pen layers and Bidirectional 2-Opt TSP path ordering.
func (c *Canva) WritePlotterSVG(w io.Writer, cfg render.PlotterConfig) (render.PlotterStats, error) {
	if c == nil {
		return render.PlotterStats{}, errors.New("nil Canva")
	}
	rc := c.lastRenderCanvas
	if rc == nil {
		rc = render.NewCanvasFromRGBA(c.img)
	}
	return rc.WritePlotterSVG(w, cfg)
}

// ToPlotterSVG saves the canvas as a pen-plotter-optimized SVG file at fpath and returns PlotterStats.
func (c *Canva) ToPlotterSVG(fpath string, cfg render.PlotterConfig) (render.PlotterStats, error) {
	if c == nil {
		return render.PlotterStats{}, errors.New("nil Canva")
	}
	rc := c.lastRenderCanvas
	if rc == nil {
		rc = render.NewCanvasFromRGBA(c.img)
	}
	return rc.SavePlotterSVG(fpath, cfg)
}

// WritePDF exports the canvas as an archival vector PDF (or embedded raster fallback) to w.
func (c *Canva) WritePDF(w io.Writer) error {
	if c == nil {
		return errors.New("nil Canva")
	}
	rc := c.lastRenderCanvas
	if rc == nil {
		rc = render.NewCanvasFromRGBA(c.img)
	}
	return rc.WritePDF(w)
}

// ToPDF saves the canvas as a PDF file at fpath.
func (c *Canva) ToPDF(fpath string) error {
	if c == nil {
		return errors.New("nil Canva")
	}
	rc := c.lastRenderCanvas
	if rc == nil {
		rc = render.NewCanvasFromRGBA(c.img)
	}
	return rc.SavePDF(fpath)
}

// Img64 resolves and returns a 16-bit per channel (48/64-bit deep-color) *image.RGBA64
// directly from the 64-bit linear HDR buffer, preserving smooth gradients without 8-bit Mach banding.
func (c *Canva) Img64() *image.RGBA64 {
	if c == nil {
		return nil
	}
	rc := c.lastRenderCanvas
	if rc == nil || c.rgbaTouched {
		rc = render.NewCanvasFromRGBA(c.img)
	}
	return rc.ResolveToRGBA64(nil)
}

// WritePNG16 encodes the canvas as a 16-bit-per-channel deep-color PNG with embedded
// GenerativeArt-Recipe tEXt metadata to w.
func (c *Canva) WritePNG16(w io.Writer) error {
	if c == nil {
		return errors.New("nil Canva")
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, c.Img64()); err != nil {
		return err
	}
	rec := c.lastRecipe
	if rec == nil {
		rec = captureEngineRecipe(c, nil)
	}
	pngBytes := buf.Bytes()
	if withMeta, err := InjectPNGRecipe(pngBytes, rec); err == nil {
		pngBytes = withMeta
	}
	_, err := w.Write(pngBytes)
	return err
}

// ToPNG16 saves the canvas as a 16-bit-per-channel deep-color PNG with embedded recipe metadata at fpath.
func (c *Canva) ToPNG16(fpath string) error {
	var buf bytes.Buffer
	if err := c.WritePNG16(&buf); err != nil {
		return err
	}
	return os.WriteFile(fpath, buf.Bytes(), 0644)
}

// WriteHDR exports the un-clamped 64-bit linear HDR buffer in Radiance RGBE (.hdr) format to w.
func (c *Canva) WriteHDR(w io.Writer) error {
	if c == nil {
		return errors.New("nil Canva")
	}
	rc := c.lastRenderCanvas
	if rc == nil || c.rgbaTouched {
		rc = render.NewCanvasFromRGBA(c.img)
	}
	return rc.WriteHDR(w)
}

// ToHDR saves the un-clamped 64-bit linear HDR buffer as a Radiance RGBE (.hdr) file at fpath.
func (c *Canva) ToHDR(fpath string) error {
	if c == nil {
		return errors.New("nil Canva")
	}
	rc := c.lastRenderCanvas
	if rc == nil || c.rgbaTouched {
		rc = render.NewCanvasFromRGBA(c.img)
	}
	return rc.SaveHDR(fpath)
}

// RecipeMetadata holds the reproducible seed, dimensions, engine name, color palette, options,
// post-processing effects, and parameters embedded inside a generated PNG's tEXt chunk.
type RecipeMetadata struct {
	Engine      string                 `json:"engine"`
	Width       int                    `json:"width"`
	Height      int                    `json:"height"`
	Seed        int64                  `json:"seed"`
	Background  [4]uint8               `json:"background"`
	Foreground  [4]uint8               `json:"foreground,omitempty"`
	LineColor   [4]uint8               `json:"line_color,omitempty"`
	LineWidth   float64                `json:"line_width,omitempty"`
	NIters      int                    `json:"n_iters,omitempty"`
	Alpha       int                    `json:"alpha,omitempty"`
	Palette     [][4]uint8             `json:"palette,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
	PostEffects []string               `json:"post_effects,omitempty"`
}

// Recipe returns the RecipeMetadata captured from the most recent Draw call.
func (c *Canva) Recipe() *RecipeMetadata {
	if c == nil {
		return nil
	}
	return c.lastRecipe
}

// InspectEngineParams extracts all scalar struct fields from an Engine instance via reflection.
func InspectEngineParams(e Engine) map[string]interface{} {
	params := make(map[string]interface{})
	if e == nil {
		return params
	}
	t := reflect.TypeOf(e)
	v := reflect.ValueOf(e)
	if t.Kind() == reflect.Ptr && !v.IsNil() {
		t = t.Elem()
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			fType := t.Field(i)
			fVal := v.Field(i)
			switch fVal.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				params[fType.Name] = fVal.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				params[fType.Name] = fVal.Uint()
			case reflect.Float32, reflect.Float64:
				params[fType.Name] = fVal.Float()
			case reflect.Bool:
				params[fType.Name] = fVal.Bool()
			case reflect.String:
				params[fType.Name] = fVal.String()
			}
		}
	}
	return params
}

func captureEngineRecipe(c *Canva, e Engine) *RecipeMetadata {
	opts := c.Opts()
	bg := opts.Background()
	fg := opts.Foreground()
	lc := opts.LineColor()
	schema := opts.ColorSchema()
	var palStops [][4]uint8
	if len(schema) > 0 {
		palStops = make([][4]uint8, len(schema))
		for i, col := range schema {
			palStops[i] = [4]uint8{col.R, col.G, col.B, col.A}
		}
	}
	meta := &RecipeMetadata{
		Engine:     "Custom",
		Width:      c.Width(),
		Height:     c.Height(),
		Seed:       c.Seed(),
		Background: [4]uint8{bg.R, bg.G, bg.B, bg.A},
		Foreground: [4]uint8{fg.R, fg.G, fg.B, fg.A},
		LineColor:  [4]uint8{lc.R, lc.G, lc.B, lc.A},
		LineWidth:  opts.LineWidth(),
		NIters:     opts.NIters(),
		Alpha:      opts.Alpha(),
		Palette:    palStops,
		Params:     InspectEngineParams(e),
	}
	if len(c.postEffects) > 0 {
		for _, eff := range c.postEffects {
			if eff != nil {
				meta.PostEffects = append(meta.PostEffects, eff.Name())
			}
		}
	}
	if e == nil {
		return meta
	}

	t := reflect.TypeOf(e)
	v := reflect.ValueOf(e)
	if t.Kind() == reflect.Ptr && !v.IsNil() {
		t = t.Elem()
	}
	name := t.Name()
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}
	if name != "" {
		meta.Engine = name
	}
	return meta
}

var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

// InjectPNGRecipe inserts a standard PNG tEXt chunk with keyword "GenerativeArt-Recipe"
// right before the IEND chunk of pngBytes.
func InjectPNGRecipe(pngBytes []byte, recipe *RecipeMetadata) ([]byte, error) {
	if len(pngBytes) < 16 || !bytes.Equal(pngBytes[:8], pngSignature) {
		return nil, errors.New("invalid PNG byte stream")
	}
	if recipe == nil {
		return pngBytes, nil
	}
	jsonPayload, err := json.Marshal(recipe)
	if err != nil {
		return nil, err
	}

	// Build tEXt chunk data: Keyword + 0x00 + Latin1/JSON text
	var chunkData bytes.Buffer
	chunkData.WriteString(PNGRecipeKeyword)
	chunkData.WriteByte(0)
	chunkData.Write(jsonPayload)
	dataBytes := chunkData.Bytes()

	// Locate the IEND chunk by scanning PNG chunks from offset 8
	iendOffset := -1
	pos := 8
	for pos+8 <= len(pngBytes) {
		length := int(binary.BigEndian.Uint32(pngBytes[pos : pos+4]))
		chunkType := string(pngBytes[pos+4 : pos+8])
		if chunkType == "IEND" {
			iendOffset = pos
			break
		}
		next := pos + 12 + length
		if next > len(pngBytes) || length < 0 {
			break
		}
		pos = next
	}
	if iendOffset < 8 {
		return nil, errors.New("IEND chunk not found in PNG")
	}

	var out bytes.Buffer
	out.Grow(len(pngBytes) + len(dataBytes) + 12)
	out.Write(pngBytes[:iendOffset])

	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(dataBytes)))
	out.Write(lenBuf[:])
	out.WriteString("tEXt")
	out.Write(dataBytes)

	crc := crc32.NewIEEE()
	crc.Write([]byte("tEXt"))
	crc.Write(dataBytes)
	var crcBuf [4]byte
	binary.BigEndian.PutUint32(crcBuf[:], crc.Sum32())
	out.Write(crcBuf[:])

	out.Write(pngBytes[iendOffset:])
	return out.Bytes(), nil
}

// ReadPNGRecipeFromBytes parses PNG chunks in pngBytes and extracts the embedded
// GenerativeArt-Recipe JSON metadata if present.
func ReadPNGRecipeFromBytes(pngBytes []byte) (*RecipeMetadata, error) {
	if len(pngBytes) < 16 || !bytes.Equal(pngBytes[:8], pngSignature) {
		return nil, errors.New("invalid PNG signature")
	}
	prefix := append([]byte(PNGRecipeKeyword), 0)
	pos := 8
	for pos+8 <= len(pngBytes) {
		length := int(binary.BigEndian.Uint32(pngBytes[pos : pos+4]))
		chunkType := string(pngBytes[pos+4 : pos+8])
		dataStart := pos + 8
		dataEnd := dataStart + length
		if dataEnd > len(pngBytes) || length < 0 {
			break
		}
		if chunkType == "tEXt" {
			chunkData := pngBytes[dataStart:dataEnd]
			if bytes.HasPrefix(chunkData, prefix) {
				jsonBytes := chunkData[len(prefix):]
				var meta RecipeMetadata
				if err := json.Unmarshal(jsonBytes, &meta); err != nil {
					return nil, fmt.Errorf("failed to unmarshal PNG recipe: %w", err)
				}
				return &meta, nil
			}
		}
		if chunkType == "IEND" {
			break
		}
		pos = dataEnd + 4
	}
	return nil, errors.New("GenerativeArt-Recipe tEXt chunk not found")
}

// ReadPNGRecipe reads a PNG file at fpath and extracts its embedded RecipeMetadata.
func ReadPNGRecipe(fpath string) (*RecipeMetadata, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return ReadPNGRecipeFromBytes(data)
}
