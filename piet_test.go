package piet

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"bytes"
	"image"
	"os"
	"testing"
)

// Runs the first Hello World example from http://www.dangermouse.net/esoteric/piet/samples.html
func TestHelloWorld(t *testing.T) {
	reader, _ := os.Open("testdata/Piet_hello.png")
	m, _, _ := image.Decode(reader)
	outputBuffer := &bytes.Buffer{}
	i := New(m)
	i.Writer = outputBuffer
	i.Run()

	output := outputBuffer.String()
	if output != "Hello world!" {
		t.Error("Incorrect output", output)
	}
}
