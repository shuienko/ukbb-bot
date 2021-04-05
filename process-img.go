package main

import (
	"image"
	"image/png"
	"io"
	"log"
	"math"
	"os"
)

// Pixel represents single pixel ad (R, G, B) data set
type Pixel struct {
	R int
	G int
	B int
}

// Point is just Pixel RGB data + coordinates
type Point struct {
	px Pixel
	x  int
	y  int
}

// RadarImage is go representation of jpeg image plus additional info: Width and Height
type RadarImage struct {
	Bitmap [][]Pixel
	Width  int
	Height int
}

// ComparePixels returns true if pixels are similar and false if not
func ComparePixels(p1, p2 Pixel) bool {
	if math.Abs(float64(p1.R-p2.R)) < float64(RGBDeviation) &&
		math.Abs(float64(p1.G-p2.G)) < float64(RGBDeviation) &&
		math.Abs(float64(p1.B-p2.B)) < float64(RGBDeviation) {
		return true
	} else {
		return false
	}
}

// openPNG return *os.File object. Please close file outside of the function
func openPNG(path string) *os.File {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	file, err := os.Open(path)

	if err != nil {
		log.Println("File could not be opened")
	}

	return file
}

// rgbaToPixel return Pixel based on four uint32 output from img.At(x, y).RGBA()
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257)}
}

// AnalizeArea returns count of pixels for each of the weather condition
func AnalizeArea(area []Point) map[string]int {

	weatherData := make(map[string]int)

	for _, pt := range area {
		switch {
		// "Normal" Precipitations
		case ComparePixels(pt.px, PrecipLow):
			weatherData["PrecipLow"]++
		case ComparePixels(pt.px, PrecipMed):
			weatherData["PrecipMed"]++
		case ComparePixels(pt.px, PrecipHigh):
			weatherData["PrecipHigh"]++

		// Convective Precipitations
		case ComparePixels(pt.px, ConvPrecipLow):
			weatherData["ConvPrecipLow"]++
		case ComparePixels(pt.px, ConvPrecipMed):
			weatherData["ConvPrecipMed"]++
		case ComparePixels(pt.px, ConvPrecipHigh):
			weatherData["ConvPrecipHigh"]++

		// Storm probability
		case ComparePixels(pt.px, Storm70):
			weatherData["Storm70"]++
		case ComparePixels(pt.px, Storm90):
			weatherData["Storm90"]++
		case ComparePixels(pt.px, Storm100):
			weatherData["Storm100"]++

		// Hail
		case ComparePixels(pt.px, HailLow):
			weatherData["HailLow"]++
		case ComparePixels(pt.px, HailMed):
			weatherData["HailMed"]++
		case ComparePixels(pt.px, HailHigh):
			weatherData["HailHigh"]++

		// Squall
		case ComparePixels(pt.px, SquallLow):
			weatherData["SquallLow"]++
		case ComparePixels(pt.px, SquallMed):
			weatherData["SquallMed"]++
		case ComparePixels(pt.px, SquallHigh):
			weatherData["SquallHigh"]++
		}
	}

	return weatherData
}

// isItGettingWorse returns true if something is withing InfoDist range
func isItGettingWorse() bool {
	var picNow, picPrev RadarImage

	fileNow := openPNG(NowImageName)
	defer fileNow.Close()

	filePrev := openPNG(PrevImageName)
	defer filePrev.Close()

	picNow.New(fileNow)
	picPrev.New(filePrev)

	areaNow := picNow.GetArea(HomeX, HomeY, InfoDist)
	areaPrev := picPrev.GetArea(HomeX, HomeY, InfoDist)

	weatherDataNow := AnalizeArea(areaNow)
	weatherDataPrev := AnalizeArea(areaPrev)

	if len(weatherDataPrev) == 0 && len(weatherDataNow) != 0 {
		return true
	} else {
		return false
	}
}

// New initialize RadarImage struct from file
func (r *RadarImage) New(file io.Reader) error {
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	r.Width, r.Height = bounds.Max.X, bounds.Max.Y

	for x := 0; x < r.Width; x++ {
		var row []Pixel
		for y := 0; y < r.Height; y++ {
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		r.Bitmap = append(r.Bitmap, row)
	}
	return nil
}

// GetArea return []Point with all pixels within 'radius' pixels
func (r *RadarImage) GetArea(xCenter, yCenter, radius int) []Point {
	var result []Point

	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			if math.Pow(float64(x-xCenter), 2)+math.Pow(float64(y-yCenter), 2) <= math.Pow(float64(radius), 2) {
				result = append(result, Point{r.Bitmap[x][y], x, y})
			}
		}
	}

	return result
}
