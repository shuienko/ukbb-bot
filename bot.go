package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math"
	"os"
)

const (
	RGBDeviation = 17
)

/*
TODO:
- Get location from user. Save it with user unique ID.
- Validate location. Should be within 100 km from image center.
- Parsre website http page and download jpg image.
- For each saved user(location) process nearby pixels (10-20 km range?). If something is there - alert.
- Probably, should split users and start processing within goroutines. Just to speed up.
- Need to have image cache. 10m min expiration. Just a local disk cache

Need to have:
1. Clear sky default
2. Current map
3. Previous map

Within 10 km:
if previous map == defalt && current map != default then alert
if previous map != default && current map != defalt then silece
if previous map == default && current map == default then silence

URLs:
view-source:https://meteo.gov.ua/ua/33345/radar
https://meteo.gov.ua/radars/Ukr_J%202021-03-22%2010-39-00.jpg
https://stackoverflow.com/questions/33186783/get-a-pixel-array-from-from-golang-image-image

Home: x: 323 y: 281
KBP: x: 395 y: 300
Distance: 40 km
Pixel dist: 75
1.875 px/km

*/

// Pixel represents single pixel ad (R, G, B) data set
type Pixel struct {
	R int
	G int
	B int
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

type Point struct {
	px Pixel
	x  int
	y  int
}

type RadarImage struct {
	Bitmap [][]Pixel
	Width  int
	Height int
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

// Center gives x and y coordinates of the center of RadarImage object
func (r *RadarImage) Center() (int, int) {
	return r.Height / 2, r.Width / 2
}

// GetArea return []Point with all pixels within radius
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

// rgbaToPixel return Pixel based on four uint32 output from img.At(x, y).RGBA()
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257)}
}

func main() {
	fmt.Println("Start Bot")

	image.RegisterFormat("jpg", "jpg", jpeg.Decode, jpeg.DecodeConfig)
	file, err := os.Open("TestImage2.jpg")
	defer file.Close()
	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	var pic RadarImage
	pic.New(file)

	yc, xc := pic.Center()
	fmt.Printf("Center: x = %d, y = %d\n", xc, yc)
	area := pic.GetArea(xc, yc, 100)

	target := Pixel{132, 77, 229}

	for _, pt := range area {
		if ComparePixels(pt.px, target) {
			fmt.Println("Found!!!", "x:", pt.x, "y:", pt.y)
		}
	}
}
