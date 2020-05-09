package mbw_test

import (
	"bytes"
	"testing"

	"github.com/gonutz/mbw"
)

func TestReadWrittenFont(t *testing.T) {
	font := mbw.NewFont(14, 8)
	a := font.Letter('A')
	b := font.Letter('B')
	a.Set(1, 2, true)
	b.Set(2, 3, true)
	b.Set(3, 4, true)

	var buf bytes.Buffer
	err := mbw.Write(&buf, font)
	if err != nil {
		t.Error(err)
	}

	have, err := mbw.Read(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Error(err)
	}
	if have.Width() != font.Width() ||
		have.Height() != font.Height() ||
		len(have.Letters()) != len(font.Letters()) {
		t.Errorf("have size %dx%d and %d letters", have.Width(), have.Height(), len(have.Letters()))
	}
	haveA := have.Letter('A')
	if !haveA.At(1, 2) {
		t.Error("letter A has different bitmap")
	}
	haveB := have.Letter('B')
	if !haveB.At(2, 3) || !haveB.At(3, 4) {
		t.Error("letter B has different bitmap")
	}
}
