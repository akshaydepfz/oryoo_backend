package helper

import (
	"bytes"
	"fmt"
	"image"
	stdraw "image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const productImageMaxDim = 1600
const productJPEGQuality = 82

// CompressProductImage decodes common image formats, optionally downscales large images,
// and re-encodes as JPEG for smaller uploads.
func CompressProductImage(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	rgba := toRGBASolidBackground(img)
	b := rgba.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > productImageMaxDim || h > productImageMaxDim {
		var nw, nh int
		if w >= h {
			nw = productImageMaxDim
			nh = h * productImageMaxDim / w
			if nh < 1 {
				nh = 1
			}
		} else {
			nh = productImageMaxDim
			nw = w * productImageMaxDim / h
			if nw < 1 {
				nw = 1
			}
		}
		scaled := image.NewRGBA(image.Rect(0, 0, nw, nh))
		xdraw.ApproxBiLinear.Scale(scaled, scaled.Bounds(), rgba, b, xdraw.Src, nil)
		rgba = scaled
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: productJPEGQuality}); err != nil {
		return nil, fmt.Errorf("encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

func toRGBASolidBackground(img image.Image) *image.RGBA {
	b := img.Bounds()
	dst := image.NewRGBA(b)
	stdraw.Draw(dst, b, image.White, image.Point{}, stdraw.Src)
	stdraw.Draw(dst, b, img, b.Min, stdraw.Over)
	return dst
}
