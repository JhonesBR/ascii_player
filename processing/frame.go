package processing

import (
	"image"
	"math"
	"sync"
)

const CHARS = "                          `.-':_,^=;><+!rc*/z?sLTv)J7(|Fi{C}fI31tlu[neoZ5Yxjya]2ESwqkP6h9d4VpOGbUAKXHm8RD#$Bg0MNWQ%&@"

type Frame struct {
	Pixels [][]Pixel
	Chars  [][]string
}

type Pixel struct {
	R         uint8
	G         uint8
	B         uint8
	luminance uint8
}

func newPixel(r uint32, g uint32, b uint32) Pixel {
	rVal := uint8(r >> 8)
	gVal := uint8(g >> 8)
	bVal := uint8(b >> 8)
	luminance := uint8(0.299*float32(rVal) + 0.587*float32(gVal) + 0.114*float32(bVal))
	return Pixel{R: rVal, G: gVal, B: bVal, luminance: luminance}
}

func luminanceToChar(luminance uint8, fullfilled bool) string {
	if fullfilled {
		return string('\u2588')
	}

	index := int((float64(luminance) / 255.0) * float64(len(CHARS)-1))
	return string(CHARS[index])
}

func GetFrame(img image.Image, width, height uint, fullfilled bool) (*Frame, error) {
	bounds := img.Bounds()
	imageWidth, imageHeight := bounds.Max.X, bounds.Max.Y
	targetWidth, targetHeight := getOutputSize(width, height, uint(imageWidth), uint(imageHeight))
	chunkWidth := uint(math.Floor(float64(imageWidth) / float64(targetWidth)))
	chunkHeight := uint(math.Floor(float64(imageHeight) / float64(targetHeight)))

	pixels := make([][]Pixel, targetHeight)
	chars := make([][]string, targetHeight)

	var wg sync.WaitGroup
	for yChunk := 0; yChunk < int(targetHeight); yChunk++ {
		wg.Add(1)
		go func(yChunk int) {
			defer wg.Done()
			row := make([]Pixel, targetWidth)
			charRow := make([]string, targetWidth)

			for xChunk := 0; xChunk < int(targetWidth); xChunk++ {
				var rSum, gSum, bSum uint32
				var pixelCount uint32

				for y := yChunk * int(chunkHeight); y < (yChunk+1)*int(chunkHeight) && y < imageHeight; y++ {
					for x := xChunk * int(chunkWidth); x < (xChunk+1)*int(chunkWidth) && x < imageWidth; x++ {
						r, g, b, _ := img.At(x, y).RGBA()
						rSum += r
						gSum += g
						bSum += b
						pixelCount++
					}
				}

				meanR := rSum / pixelCount
				meanG := gSum / pixelCount
				meanB := bSum / pixelCount

				chunkPixel := newPixel(meanR, meanG, meanB)
				row[xChunk] = chunkPixel
				charRow[xChunk] = luminanceToChar(chunkPixel.luminance, fullfilled)
			}

			pixels[yChunk] = row
			chars[yChunk] = charRow
		}(yChunk)
	}

	wg.Wait()
	return &Frame{Pixels: pixels, Chars: chars}, nil
}

func getOutputSize(desiredWidth, desiredHeight, currentWidth, currentHeight uint) (uint, uint) {
	var newWidth, newHeight uint

	// Approximate aspect ratio correction factor for terminal characters
	const aspectCorrection = 0.5

	heightFactor := float64(currentHeight) / (float64(desiredHeight) / aspectCorrection)
	widthFactor := float64(currentWidth) / float64(desiredWidth)

	if heightFactor > widthFactor {
		newHeight = desiredHeight
		newWidth = uint(math.Floor(float64(currentWidth) / heightFactor))
	} else {
		newWidth = desiredWidth
		newHeight = uint(math.Floor((float64(currentHeight) / widthFactor) * aspectCorrection))
	}

	return newWidth, newHeight
}
