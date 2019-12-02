package rot13

import (
	"bytes"
	"strings"
	"testing"
)

func TestRot13(t *testing.T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(Reader{strings.NewReader("Lbh penpxrq gur pbqr!"), 13})
	s := buf.String()

	if s != "You cracked the code!" {
		t.Error("TestRot: Failed to rotate. ")
	}
}

func TestReRotN(t *testing.T) {
	str := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()abcdefghijklmnopqrstuvwxyz"

	for i := 0; i < 30000; i++ {
		buf := new(bytes.Buffer)
		buf.ReadFrom(Reader{strings.NewReader(str), i})
		res := buf.String()

		buf2 := new(bytes.Buffer)
		buf2.ReadFrom(Reader{strings.NewReader(res), -i})

		if buf2.String() != str {
			t.Error("TestRotN: Decoding failed!")
		}
	}
}

func TestRotMod0(t *testing.T) {
	str := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()abcdefghijklmnopqrstuvwxyz"

	for i := 0; i < 30000; i += 26 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(Reader{strings.NewReader(str), i})
		res := buf.String()

		if res != str {
			t.Error("TestRotMod0: Rotation failed!")
		}
	}
}
