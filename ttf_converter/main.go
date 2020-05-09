package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strconv"

	"golang.org/x/image/font/gofont/gomono"

	"github.com/gonutz/mbw"
	"github.com/gonutz/truetype"
)

func main() {
	printFontAtSize(10)
	printFontAtSize(15)
	printFontAtSize(20)
	printFontAtSize(25)
	printFontAtSize(30)
}

func printFontAtSize(pixelHeight int) {
	var letters = []rune{'ä', 'ö', 'ü', 'Ä', 'Ö', 'Ü', 'ß', '$', 'µ', '€', '°'}
	for i := 33; i < 127; i++ {
		letters = append(letters, rune(i))
	}

	ttf, err := truetype.InitFont(gomono.TTF, 0)
	check(err)
	scale := ttf.ScaleForPixelHeight(float64(pixelHeight))

	advance, _ := ttf.GetGlyphHMetrics('M')
	advance = int(float64(advance)*scale + 0.5)
	fontW := advance
	fmt.Println("letter width", fontW)

	ascend, descend, lineGap := ttf.GetFontVMetrics()
	ascend = int(float64(ascend)*scale + 0.5)
	descend = int(float64(descend)*scale + 0.5)
	lineGap = int(float64(lineGap)*scale + 0.5)
	fontH := ascend - descend
	fmt.Println("letter height", fontH)

	all := image.NewGray(image.Rect(0, -1, fontW*len(letters), fontH+2))
	for x := range letters {
		if x%2 == 1 {
			draw.Draw(
				all,
				image.Rect(x*fontW, -99, x*fontW+fontW, 99),
				image.NewUniform(color.Gray{50}),
				image.ZP,
				draw.Src,
			)
		}
	}

	font := mbw.NewFont(fontW, fontH)
	defer func() {
		font.Sort()
		check(mbw.Save(font, "gomono"+strconv.Itoa(pixelHeight)+".mbw"))
	}()

	for i, letter := range letters {
		x, y, _, _ := ttf.GetCodepointBitmapBox(int(letter), scale, scale)
		gray, w, h := ttf.GetCodepointBitmap(scale, scale, int(letter), 0, 0)

		if true {
			for i := range gray {
				if gray[i] < 100 {
					gray[i] = 0
				} else {
					gray[i] = 255
				}
			}
		}

		if w > fontW || h > fontH {
			fmt.Println("letter", string(letter), "has size", w, h)
		}
		img := &image.Gray{
			Pix:    gray,
			Stride: w,
			Rect:   image.Rect(0, 0, w, h),
		}
		whole := image.NewGray(image.Rect(0, 0, fontW, fontH))
		left := x
		top := ascend + y
		draw.Draw(whole, image.Rect(left, top, left+w, top+h), img, image.ZP, draw.Src)
		left += i * fontW
		draw.Draw(all, image.Rect(left, top, left+w, top+h), img, image.ZP, draw.Src)
		f, err := os.Create(strconv.Itoa(int(letter)) + ".png")
		check(err)
		defer f.Close()
		check(png.Encode(f, whole))

		l := font.Letter(letter)
		for y := 0; y < fontH; y++ {
			for x := 0; x < fontW; x++ {
				l.Set(x, y, whole.GrayAt(x, y).Y >= 100)
			}
		}
	}
	f, err := os.Create("all_" + strconv.Itoa(pixelHeight) + ".png")
	check(err)
	defer f.Close()
	check(png.Encode(f, all))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
