package processing

import (
	"image"
	"io"
	"math"
)

const CHARS = "                                    `.-':_,^=;><+!rc*/z?sLTv)J7(|Fi{C}fI31tlu[neoZ5Yxjya]2ESwqkP6h9d4VpOGbUAKXHm8RD#$Bg0MNWQ%&@"

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

func newPixel(r uint32, g uint32, b uint32) *Pixel {
	rVal := uint8(r >> 8)
	gVal := uint8(g >> 8)
	bVal := uint8(b >> 8)
	luminance := uint8(0.299*float32(r) + 0.587*float32(g) + 0.114*float32(b))
	return &Pixel{R: rVal, G: gVal, B: bVal, luminance: luminance}
}

func luminanceToChar(luminance uint8) string {
	index := int((float64(luminance) / 255.0) * float64(len(CHARS)-1))
	return string(CHARS[index])
}

func GetFrame(file io.Reader, width, height uint) (*Frame, error) {
	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	imageWidth, imageHeight := bounds.Max.X, bounds.Max.Y
	targetWidth, targetHeight := getOutputSize(width, height, uint(imageWidth), uint(imageHeight))
	chunkWidth := uint(math.Floor(float64(imageWidth) / float64(targetWidth)))
	chunkHeight := uint(math.Floor(float64(imageHeight) / float64(targetHeight)))

	var pixels [][]Pixel
	var chars [][]string

	for yChunk := 0; yChunk < imageHeight; yChunk += int(chunkHeight) {
		var row []Pixel
		var charRow []string

		for xChunk := 0; xChunk < imageWidth; xChunk += int(chunkWidth) {
			// Calculate pixel for the chunk
			var chunkPixels [][]Pixel
			for y := yChunk; y < yChunk+int(chunkHeight); y++ {
				if y >= imageHeight {
					break
				}
				for x := xChunk; x < xChunk+int(chunkWidth); x++ {
					if x >= imageWidth {
						break
					}
					r, g, b, _ := img.At(x, y).RGBA()
					pixel := newPixel(r, g, b)
					chunkPixels = append(chunkPixels, []Pixel{*pixel})
				}
			}

			chunkPixel := getMeanPixel(chunkPixels)
			row = append(row, *chunkPixel)
			charRow = append(charRow, luminanceToChar(chunkPixel.luminance))
		}

		pixels = append(pixels, row)
		chars = append(chars, charRow)
	}

	return &Frame{Pixels: pixels, Chars: chars}, nil
}

func getMeanPixel(pixels [][]Pixel) *Pixel {
	var rSum, gSum, bSum uint32
	for _, row := range pixels {
		for _, pixel := range row {
			rSum += uint32(pixel.R)
			gSum += uint32(pixel.G)
			bSum += uint32(pixel.B)
		}
	}

	pixelCount := len(pixels) * len(pixels[0])
	meanR := rSum / uint32(pixelCount)
	meanG := gSum / uint32(pixelCount)
	meanB := bSum / uint32(pixelCount)

	return newPixel(meanR, meanG, meanB)
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
