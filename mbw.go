package mbw

import (
	"encoding/binary"
	"errors"
	"image"
	"io"
	"os"
	"sort"
)

// NewFont creates a new font in which every character is width by height pixels
// in size. No letters are contained in a new font. Add them with the Letter
// function.
func NewFont(width, height int) *Font {
	return &Font{
		width:  width,
		height: height,
	}
}

// Font represents a monospaced black and white font. Each letter is Width x
// Height pixels in size.
type Font struct {
	width   int
	height  int
	letters []Letter
}

// Width returns the width of a letter in pixels.
func (f *Font) Width() int { return f.width }

// Height returns the height of a letter in pixels.
func (f *Font) Height() int { return f.height }

// Letters returns the slice of all existing letters in the font. These are all
// letters that were added through the Letter function.
func (f *Font) Letters() []Letter { return f.letters }

// Letter returns the Letter with the given rune. If the rune is not present in
// the Font, a new Letter is created and returned.
func (f *Font) Letter(r rune) *Letter {
	for i := range f.letters {
		if f.letters[i].Rune == r {
			return &f.letters[i]
		}
	}
	f.letters = append(f.letters, newLetter(r, f.width, f.height))
	return &f.letters[len(f.letters)-1]
}

// Sort sorts the letters by their characters, from lowest to highest.
func (f *Font) Sort() {
	sort.Sort(sortable{f})
}

type sortable struct {
	*Font
}

func (x sortable) Len() int {
	return len(x.letters)
}

func (x sortable) Less(i, j int) bool {
	return x.letters[i].Rune < x.letters[j].Rune
}

func (x sortable) Swap(i, j int) {
	x.letters[i], x.letters[j] = x.letters[j], x.letters[i]
}

func newLetter(r rune, width, height int) Letter {
	return Letter{
		Rune:   r,
		width:  width,
		height: height,
		set:    make([]bool, width*height),
	}
}

// Letter is a single letter in a Font. It represents the character Rune as a
// binary Bitmap.
type Letter struct {
	Rune   rune
	width  int
	height int
	set    []bool
}

// At returns true if the pixel at x,y is set. X goes from left to right. Y goes
// from top to bottom.
func (b *Letter) At(x, y int) bool {
	return x >= 0 && x < b.width &&
		y >= 0 && y < b.height &&
		b.set[x+y*b.width]
}

// Set sets the pixel value at x,y. X goes from left to right. Y goes from top
// to bottom.
func (b *Letter) Set(x, y int, to bool) {
	if x >= 0 && x < b.width &&
		y >= 0 && y < b.height {
		b.set[x+y*b.width] = to
	}
}

// Load reads a Font from a file.
func Load(path string) (*Font, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("mbw.Load: " + err.Error())
	}
	defer f.Close()
	return Read(f)
}

// Save writes a Font to a file.
func Save(font *Font, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.New("mbw.Save: " + err.Error())
	}
	defer f.Close()
	return Write(f, font)
}

// Read reads a Font from an io.Reader.
func Read(r io.Reader) (*Font, error) {
	// Read the file header and create an empty font to be filled with the
	// letters.
	var header fileHeader
	err := binary.Read(r, encoding, &header)
	if err != nil {
		return nil, errors.New("mbw.Read failed to read header: " + err.Error())
	}
	if string(header.Version[:]) != "MBW1" ||
		header.FontWidth == 0 ||
		header.FontHeight == 0 {
		return nil, errors.New("mbw.Read: unknown file header")
	}

	font := NewFont(int(header.FontWidth), int(header.FontHeight))

	// Read the characters.
	runes := make([]uint32, header.LetterCount)
	err = binary.Read(r, encoding, runes)
	if err != nil {
		return nil, errors.New("mbw.Read failed to read ´characters: " + err.Error())
	}

	// Read the bitmap data.
	bits := make(
		[]byte,
		int(header.LetterCount)*int(header.FontWidth)*int(header.FontHeight)/8,
	)
	_, err = r.Read(bits)
	if err != nil {
		return nil, errors.New("mbw.Read failed to read bitmap data: " + err.Error())
	}
	for i := range runes {
		letter := font.Letter(rune(runes[i]))
		for y := 0; y < font.height; y++ {
			for x := 0; x < font.width; x++ {
				n := i*font.width*font.height + y*font.width + x
				if bits[n/8]&(0x80>>uint(n%8)) != 0 {
					letter.Set(x, y, true)
				}
			}
		}
	}

	return font, nil
}

// Write writes a Font to a io.Writer.
func Write(w io.Writer, font *Font) error {
	// Write the file header, font size and letter count.
	header := fileHeader{
		Version:     [4]byte{'M', 'B', 'W', '1'},
		FontWidth:   uint16(font.width),
		FontHeight:  uint16(font.height),
		LetterCount: uint64(len(font.letters)),
	}
	err := binary.Write(w, encoding, &header)
	if err != nil {
		return errors.New("mbw.Write failed to write header: " + err.Error())
	}

	// Write the runes for all characters in the same order as the bitmap images
	// that follow.
	runes := make([]uint32, len(font.letters))
	for i := range runes {
		runes[i] = uint32(font.letters[i].Rune)
	}
	err = binary.Write(w, encoding, runes)
	if err != nil {
		return errors.New("mbw.Write failed to write characters: " + err.Error())
	}

	// Write the binary bitmap data of all characters as a byte sequence.
	bits := make([]byte, len(font.letters)*font.width*font.height/8)
	for i, letter := range font.letters {
		for y := 0; y < font.height; y++ {
			for x := 0; x < font.width; x++ {
				if letter.At(x, y) {
					n := i*font.width*font.height + y*font.width + x
					bits[n/8] |= 0x80 >> uint(n%8)
				}
			}
		}
	}
	_, err = w.Write(bits)
	if err != nil {
		return errors.New("mbw.Write failed to write bitmap data: " + err.Error())
	}

	return nil
}

var encoding = binary.LittleEndian

type fileHeader struct {
	Version     [4]byte
	FontWidth   uint16
	FontHeight  uint16
	LetterCount uint64
}

func ToGlyphAtlas(font *Font) *GlyphAtlas {
	// TODO
	return &GlyphAtlas{}
}

type GlyphAtlas struct {
	Image *image.Gray
}
