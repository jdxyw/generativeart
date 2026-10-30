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

// DrawTimelapseFrames renders totalFrames frames of e into c.Frames().
// If e implements AnimatedEngine, GenerativeFrame is invoked with phase = frameIdx / totalFrames.
// Otherwise, each frame advances the canvas deterministic seed by +1 and invokes e.Generative(c).
func (c *Canva) DrawTimelapseFrames(e Engine, totalFrames int) {
	if c == nil || e == nil || totalFrames <= 0 {
		return
	}
	c.lastRecipe = captureEngineRecipe(c, e)
	c.frames = make([]*image.RGBA, 0, totalFrames)
	baseSeed := c.Seed()
	anim, isAnim := e.(AnimatedEngine)

	for i := 0; i < totalFrames; i++ {
		phase := float64(i) / float64(totalFrames)
		c.FillBackground()
		if isAnim {
			c.SetSeed(baseSeed)
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
		c.frames = append(c.frames, frameCopy)
	}
	c.SetSeed(baseSeed)
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
	type bucket struct {
		count            int
		sumR, sumG, sumB int
		avgR, avgG, avgB int
		weight           float64
	}
	var buckets [32768]bucket

	// Sample up to 4 evenly-spaced frames across the animation to populate the color histogram
	stepFrame := 1
	if len(srcFrames) > 4 {
		stepFrame = len(srcFrames) / 4
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

	// Multi-pass perceptual diversity seeding so all 256 slots cover the full gamut without clumping
	centers := make([][3]int, 0, 256)
	used := make([]bool, len(active))
	thresholds := []int{420, 180, 72, 24, 0}
	for _, minDist := range thresholds {
		if len(centers) >= 256 {
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
				if len(centers) >= 256 {
					break
				}
			}
		}
	}

	// 3 iterations of weighted K-Means refinement over active histogram buckets
	if len(centers) > 1 && len(active) > len(centers) {
		sumW := make([]float64, len(centers))
		sumR := make([]float64, len(centers))
		sumG := make([]float64, len(centers))
		sumB := make([]float64, len(centers))
		for iter := 0; iter < 3; iter++ {
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

	for i := 0; len(centers) < 256 && i < len(palette.Plan9); i += 4 {
		r, g, b, _ := palette.Plan9[i].RGBA()
		centers = append(centers, [3]int{int(r >> 8), int(g >> 8), int(b >> 8)})
	}

	pal := make(color.Palette, len(centers))
	for i, c := range centers {
		pal[i] = color.RGBA{R: uint8(c[0]), G: uint8(c[1]), B: uint8(c[2]), A: 255}
	}

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

	sharedPal, lut, centers := buildSharedAnimationPalette(srcFrames)
	// Subtle 4x4 micro-dither matrix (-2..+2) applied only when the pixel lies between palette stops
	bayer4x4 := [4][4]int{
		{-2, 0, -1, 1},
		{1, -1, 2, 0},
		{-1, 1, -2, 0},
		{2, 0, 1, -1},
	}

	outGIF := &gif.GIF{
		Image:     make([]*image.Paletted, len(srcFrames)),
		Delay:     make([]int, len(srcFrames)),
		LoopCount: 0,
	}

	var wg sync.WaitGroup
	for i, rgba := range srcFrames {
		outGIF.Delay[i] = delayCentiseconds
		wg.Add(1)
		go func(frameIdx int, src *image.RGBA) {
			defer wg.Done()
			b := src.Bounds()
			fw, fh := b.Dx(), b.Dy()
			pm := image.NewPaletted(b, sharedPal)
			for y := 0; y < fh; y++ {
				bRow := bayer4x4[y&3]
				srcOff := y * src.Stride
				dstOff := y * pm.Stride
				for x := 0; x < fw; x++ {
					r0 := int(src.Pix[srcOff+x*4+0])
					g0 := int(src.Pix[srcOff+x*4+1])
					b0 := int(src.Pix[srcOff+x*4+2])
					baseKey := ((r0 >> 2) << 12) | ((g0 >> 2) << 6) | (b0 >> 2)
					idx0 := lut[baseKey]
					pc := centers[idx0]
					dr := r0 - pc[0]
					dg := g0 - pc[1]
					db := b0 - pc[2]
					// Skip dithering on exact/near-exact matches so solid fills and vector strokes stay razor-sharp
					if 3*dr*dr+4*dg*dg+2*db*db <= 28 {
						pm.Pix[dstOff+x] = idx0
						continue
					}
					dith := bRow[x&3]
					r := r0 + dith
					g := g0 + dith
					bl := b0 + dith
					if r < 0 {
						r = 0
					} else if r > 255 {
						r = 255
					}
					if g < 0 {
						g = 0
					} else if g > 255 {
						g = 255
					}
					if bl < 0 {
						bl = 0
					} else if bl > 255 {
						bl = 255
					}
					key := ((r >> 2) << 12) | ((g >> 2) << 6) | (bl >> 2)
					pm.Pix[dstOff+x] = lut[key]
				}
			}
			outGIF.Image[frameIdx] = pm
		}(i, rgba)
	}
	wg.Wait()
	return gif.EncodeAll(w, outGIF)
}

// ToGIF saves the recorded timelapse frames (or current image) as an animated GIF at fpath.
func (c *Canva) ToGIF(fpath string, delayCentiseconds int) error {
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	if err := c.WriteGIF(f, delayCentiseconds); err != nil {
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

// RecipeMetadata holds the reproducible seed, dimensions, engine name, and parameters
// embedded inside a generated PNG's tEXt chunk.
type RecipeMetadata struct {
	Engine      string                 `json:"engine"`
	Width       int                    `json:"width"`
	Height      int                    `json:"height"`
	Seed        int64                  `json:"seed"`
	Background  [4]uint8               `json:"background"`
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

func captureEngineRecipe(c *Canva, e Engine) *RecipeMetadata {
	bg := c.Opts().Background()
	meta := &RecipeMetadata{
		Engine:     "Custom",
		Width:      c.Width(),
		Height:     c.Height(),
		Seed:       c.Seed(),
		Background: [4]uint8{bg.R, bg.G, bg.B, bg.A},
		Params:     make(map[string]interface{}),
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
		v = v.Elem()
	}
	name := t.Name()
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}
	if name != "" {
		meta.Engine = name
	}

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			fType := t.Field(i)
			fVal := v.Field(i)
			switch fVal.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				meta.Params[fType.Name] = fVal.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				meta.Params[fType.Name] = fVal.Uint()
			case reflect.Float32, reflect.Float64:
				meta.Params[fType.Name] = fVal.Float()
			case reflect.Bool:
				meta.Params[fType.Name] = fVal.Bool()
			case reflect.String:
				meta.Params[fType.Name] = fVal.String()
			}
		}
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
