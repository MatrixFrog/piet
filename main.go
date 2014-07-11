package main

import (
	_ "image/gif"
	_ "image/png"

	"image"
	"log"
	"os"
)

func main() {
	reader, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	piet := New(m)
	piet.run()
}
