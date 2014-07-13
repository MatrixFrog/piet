package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"image"
	"log"
	"os"

	"github.com/MatrixFrog/piet"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	reader, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	i := piet.New(m)
	i.Run()
	println() // In case the Piet program's output didn't end with a newline.
}

func usage() {
	log.Println("Usage: " + os.Args[0] + " program")
	log.Println("program may be a .gif, .jpg, or .png image")
	os.Exit(1)
}
