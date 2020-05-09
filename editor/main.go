package main

import (
	"fmt"
	"strconv"

	"github.com/gonutz/mbw"
	"github.com/gonutz/prototype/draw"
)

func main() {
	type view struct {
		x, y, scale int
		grid        bool
	}

	var (
		fullscreen = false
		font       = mbw.NewFont(8, 14)
		curLetter  rune
		large      = view{x: 150, y: 30, scale: 30, grid: true}
		thrice     = view{x: 500, y: 30, scale: 3}
		twice      = view{x: 600, y: 30, scale: 2}
		unscaled   = view{x: 700, y: 30, scale: 1}
		views      = []view{large, thrice, twice, unscaled}
	)

	const fontFile = "font.mbw"
	if f, err := mbw.Load(fontFile); err == nil {
		font = f
	}
	defer func() {
		check(mbw.Save(font, fontFile))
	}()

	fmt.Println(len(font.Letters()), "letters")
	mbw.ToGlyphAtlas(font)

	check(draw.RunWindow("mbw Font Editor", 800, 600, func(window draw.Window) {
		if window.WasKeyPressed(draw.KeyEscape) {
			window.Close()
		}
		if window.WasKeyPressed(draw.KeyF11) {
			fullscreen = !fullscreen
		}
		window.SetFullscreen(fullscreen)

		for _, r := range window.Characters() {
			curLetter = r
		}

		if curLetter == 0 {
			window.DrawText("Type the letter that you want to edit.", 0, 0, draw.White)
		} else {
			letter := font.Letter(curLetter)
			window.DrawText(strconv.Itoa(font.Width())+"x"+strconv.Itoa(font.Height())+"   Letter: "+string(letter.Rune), 0, 0, draw.White)

			drawing := window.IsMouseDown(draw.LeftButton) || window.IsMouseDown(draw.RightButton)
			color := false
			if window.IsMouseDown(draw.LeftButton) {
				color = true
			}
			if drawing {
				mx, my := window.MousePosition()
				x := (mx - large.x) / large.scale
				y := (my - large.y) / large.scale
				letter.Set(x, y, color)
			}

			for _, view := range views {
				for y := 0; y < font.Height(); y++ {
					for x := 0; x < font.Width(); x++ {
						if view.grid {
							window.DrawRect(
								view.x+x*view.scale,
								view.y+y*view.scale,
								view.scale,
								view.scale,
								draw.Gray,
							)
						}
						if letter.At(x, y) {
							window.FillRect(
								view.x+x*view.scale,
								view.y+y*view.scale,
								view.scale,
								view.scale,
								draw.White,
							)
						}
					}
				}
			}
		}
	}))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
