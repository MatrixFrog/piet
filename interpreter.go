package main

import (
	"fmt"
	_ "image/png"
	"log"

	"image"
	"image/color"
	"sort"
)

type dpDir image.Point

var (
	east  dpDir
	south dpDir
	west  dpDir
	north dpDir
)

func init() {
	east = dpDir{1, 0}
	south = dpDir{0, 1}
	west = dpDir{-1, 0}
	north = dpDir{0, -1}

}

type ccDir int

const (
	left ccDir = iota
	right
)

// adj returns the four points adjacent to a point.
func adj(p image.Point) [4]image.Point {
	return [4]image.Point{
		{p.X, p.Y + 1},
		{p.X, p.Y - 1},
		{p.X + 1, p.Y},
		{p.X - 1, p.Y},
	}
}

type colorBlock map[image.Point]struct{}

func (cb colorBlock) String() string {
	s := make(sort.StringSlice, len(cb))
	for p, _ := range cb {
		s = append(s, p.String())
	}
	sort.Sort(s)

	return fmt.Sprint(s)
}

func (cb colorBlock) Bounds() (r *image.Rectangle) {
	for p, _ := range cb {
		if r == nil {
			r = &image.Rectangle{p, image.Point{p.X + 1, p.Y + 1}}
		} else {
			newR := image.Rectangle{p, image.Point{p.X + 1, p.Y + 1}}.Union(*r)
			r = &newR
		}
	}
	return
}

type interpreter struct {
	img image.Image
	dp  dpDir
	cc  ccDir
	pos image.Point
}

func (i interpreter) String() string {
	var dp string
	switch i.dp {
	case east:
		dp = "\u261E"
	case south:
		dp = "\u261F"
	case west:
		dp = "\u261C"
	case north:
		dp = "\u261D"
	}

	var cc string
	switch i.cc {
	case left:
		cc = "<"
	case right:
		cc = ">"
	}

	return fmt.Sprintf("dp:%s, cc:%s, pos:%s", i.pos, dp, cc)
}

func New(img image.Image) interpreter {
	return interpreter{
		img: img,
		dp:  east,
		cc:  left,
		pos: img.Bounds().Min,
	}
}

func (i interpreter) color() color.Color {
	return i.img.At(i.pos.X, i.pos.Y)
}

func (i *interpreter) rotateDp() {
	switch i.dp {
	case east:
		i.dp = south
	case south:
		i.dp = west
	case west:
		i.dp = north
	case north:
		i.dp = east
	}
}

func (i *interpreter) toggleCc() {
	if i.cc == left {
		i.cc = right
	} else {
		i.cc = left
	}
}

// getColorBlock returns the current color block.
func (i *interpreter) getColorBlock() (block colorBlock) {
	currentColor := i.color()
	block = map[image.Point]struct{}{
		i.pos: struct{}{},
	}
	// very naive implementation currently
	done := false
	for !done {
		done = true
		for pos, _ := range block {
			for _, newPos := range adj(pos) {
				if newPos.In(i.img.Bounds()) {
					_, inBlock := block[newPos]
					if !inBlock && sameColors(i.img.At(newPos.X, newPos.Y), currentColor) {
            block[newPos] = struct{}{}
            done = false
					}
				}
			}
		}
	}
	return
}

func (i interpreter) canMoveTo(p image.Point) bool {
	return p.In(i.img.Bounds()) && !sameColors(color.Black, i.img.At(p.X, p.Y))
}

// move causes the interpreter to attempt execute a single move.
// Returns whether the move was successful.
func (i *interpreter) move() bool {
	originalDp := i.dp
	originalCc := i.cc

	// When true, toggle the CC. When false, rotate the DP.
	cc := true

	for {
		// First, move to the edge of the current block.
		i.pos = *i.newPosInBlock()
		// Then, try to move into the next block.
		newPos := i.pos.Add(image.Point(i.dp))
		if i.canMoveTo(newPos) {
			i.pos = newPos
      if i.pos.X == 4 && i.pos.Y  == 8 {
        log.Println("Should be halting!")
      }
			return true
		}

		if cc {
			i.toggleCc()
		} else {
			i.rotateDp()
		}
		if i.dp == originalDp && i.cc == originalCc {
      log.Println("Halting!")
			return false
		}

		cc = !cc
	}
}

func (i *interpreter) run() {
	for i.move() {
	}
}
func (i *interpreter) newPosInBlock() *image.Point {
	var newPos *image.Point
	block := i.getColorBlock()
	bounds := block.Bounds()

	switch i.dp {
	case east:
		for p, _ := range block {
			if p.X == bounds.Max.X-1 {
				if newPos == nil ||
					i.cc == left && p.Y < newPos.Y ||
					i.cc == right && p.Y > newPos.Y {
					newPos = &image.Point{p.X, p.Y}
				}
			}
		}
	case south:
		for p, _ := range block {
			if p.Y == bounds.Max.Y-1 {
				if newPos == nil ||
					i.cc == left && p.X > newPos.X ||
					i.cc == right && p.X < newPos.X {
					newPos = &image.Point{p.X, p.Y}
				}
			}
		}
	case west:
		for p, _ := range block {
			if p.X == bounds.Min.X {
				if newPos == nil ||
					i.cc == left && p.Y > newPos.Y ||
					i.cc == right && p.Y < newPos.Y {
					newPos = &image.Point{p.X, p.Y}
				}
			}
		}
	case north:
		for p, _ := range block {
			if p.Y == bounds.Min.Y {
				if newPos == nil ||
					i.cc == left && p.X < newPos.X ||
					i.cc == right && p.X > newPos.X {
					newPos = &image.Point{p.X, p.Y}
				}
			}
		}
	}
	return newPos
}

func sameColors(c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return r1 == r2 &&
		g1 == g2 &&
		b1 == b2
}
