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
	"os"
	"reflect"
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

// buildSharedAnimationPalette constructs a high-fidelity 256-color palette tailored to
// the actual colors present in srcFrames (top 192 histogram colors + 64 Plan9 fallback colors),
// along with a 15-bit (32x32x32) RGB lookup table for O(1) flicker-free frame quantization.
func buildSharedAnimationPalette(srcFrames []*image.RGBA) (color.Palette, *[32768]uint8) {
	type bucket struct {
		idx            int
		count          int
		sumR, sumG, sumB int
	}
	var buckets [32768]bucket
	for i := range buckets {
		buckets[i].idx = i
	}

	// Sample frames to populate 15-bit RGB color histogram
	for fIdx, img := range srcFrames {
		if fIdx > 0 && fIdx != len(srcFrames)/2 && len(srcFrames) > 4 {
			continue
		}
		pix := img.Pix
		for i := 0; i+3 < len(pix); i += 8 {
			r, g, b := int(pix[i]), int(pix[i+1]), int(pix[i+2])
			key := ((r >> 3) << 10) | ((g >> 3) << 5) | (b >> 3)
			buckets[key].count++
			buckets[key].sumR += r
			buckets[key].sumG += g
			buckets[key].sumB += b
		}
	}

	active := make([]bucket, 0, 1024)
	for _, bk := range buckets {
		if bk.count > 0 {
			active = append(active, bk)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].count > active[j].count
	})

	pal := make(color.Palette, 0, 256)
	for i := 0; i < len(active) && len(pal) < 200; i++ {
		bk := active[i]
		pal = append(pal, color.RGBA{
			R: uint8(bk.sumR / bk.count),
			G: uint8(bk.sumG / bk.count),
			B: uint8(bk.sumB / bk.count),
			A: 255,
		})
	}
	for i := 0; len(pal) < 256 && i < len(palette.Plan9); i += 4 {
		pal = append(pal, palette.Plan9[i])
	}

	// Precompute 15-bit RGB -> nearest palette index LUT
	palRGBA := make([][3]int, len(pal))
	for i, c := range pal {
		r, g, b, _ := c.RGBA()
		palRGBA[i] = [3]int{int(r >> 8), int(g >> 8), int(b >> 8)}
	}

	var lut [32768]uint8
	for qr := 0; qr < 32; qr++ {
		r8 := (qr << 3) | (qr >> 2)
		for qg := 0; qg < 32; qg++ {
			g8 := (qg << 3) | (qg >> 2)
			for qb := 0; qb < 32; qb++ {
				b8 := (qb << 3) | (qb >> 2)
				bestIdx := 0
				bestDist := 1 << 30
				for pIdx, pc := range palRGBA {
					dr := r8 - pc[0]
					dg := g8 - pc[1]
					db := b8 - pc[2]
					// Perceptually weighted RGB distance
					d := 3*dr*dr + 4*dg*dg + 2*db*db
					if d < bestDist {
						bestDist = d
						bestIdx = pIdx
						if d == 0 {
							break
						}
					}
				}
				lut[(qr<<10)|(qg<<5)|qb] = uint8(bestIdx)
			}
		}
	}
	return pal, &lut
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

	sharedPal, lut := buildSharedAnimationPalette(srcFrames)
	bayer4x4 := [4][4]int{
		{-4, 0, -3, 1},
		{2, -2, 3, -1},
		{-3, 1, -4, 0},
		{3, -1, 2, -2},
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
					dith := bRow[x&3]
					r := int(src.Pix[srcOff+x*4+0]) + dith
					g := int(src.Pix[srcOff+x*4+1]) + dith
					bl := int(src.Pix[srcOff+x*4+2]) + dith
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
					key := ((r >> 3) << 10) | ((g >> 3) << 5) | (bl >> 3)
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
	Engine     string                 `json:"engine"`
	Width      int                    `json:"width"`
	Height     int                    `json:"height"`
	Seed       int64                  `json:"seed"`
	Background [4]uint8               `json:"background"`
	Params     map[string]interface{} `json:"params,omitempty"`
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
