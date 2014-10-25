package main

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"flag"
	"image"
	"log"
	"os"

	"github.com/MatrixFrog/piet"
)

var verbose = flag.Bool("v", false, "verbose")

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	reader, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	i := piet.New(m)
	if *verbose {
		i.Logger = log.New(os.Stderr, "", 0)
	}
	i.Run()
	println() // In case the Piet program's output didn't end with a newline.
}
