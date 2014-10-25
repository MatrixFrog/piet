package piet

import (
	"bufio"
	"fmt"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"unicode"

	"image"
	"image/color"
	"sort"
	"strconv"
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

func (cb colorBlock) Bounds() (r image.Rectangle) {
	for p, _ := range cb {
		r = image.Rectangle{p, image.Point{p.X + 1, p.Y + 1}}
		break
	}
	for p, _ := range cb {
		r = r.Union(image.Rectangle{p, image.Point{p.X + 1, p.Y + 1}})
	}
	return
}

type interpreter struct {
	// The program being interpreted
	img image.Image
	stack

	// The Writer where program output is sent.
	io.Writer

	// The Reader where program input is received from.
	io.Reader

	// The direction pointer
	dp dpDir

	// The codel chooser
	cc ccDir

	// The interpeter's current position in the program.
	pos image.Point

	Logger *log.Logger
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

	r, g, b, _ := i.color().RGBA()
	color := fmt.Sprintf("%02X%02X%02X", r>>8, g>>8, b>>8)

	return fmt.Sprintf("dp:%s, cc:%s, pos:%s, color:%s", dp, cc, i.pos, color)
}

// Creates a new Piet interpreter for the given image.
// The interpreter will use os.Stdin and os.Stdout, but these
// can be changed (for example, for testing) by setting the
// interpreters Reader or Writer fields.
func New(img image.Image) interpreter {
	return interpreter{
		img:    img,
		Writer: os.Stdout,
		Reader: os.Stdin,
		dp:     east,
		cc:     left,
		pos:    img.Bounds().Min,
		Logger: log.New(ioutil.Discard, "", log.Lshortfile),
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

func (i *interpreter) pointer() {
	count := i.pop()
	if count < 0 {
		panic("negative count not implemented")
	}
	for j := 0; j < count; j++ {
		i.rotateDp()
	}
}

func (i *interpreter) switchCc() {
	count := i.pop()
	if count < 0 {
		count = -count
	}
	if count%2 == 1 {
		i.toggleCc()
	}
}

func (i *interpreter) toggleCc() {
	if i.cc == left {
		i.cc = right
	} else {
		i.cc = left
	}
}

// A SplitFunc which returns sequences of digits.
func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	var b byte
	for advance, b = range data {
		if unicode.IsDigit(rune(b)) {
			token = append(token, b)
		} else {
			// Reached a non-digit. Just return what we have.
			return
		}
	}
	if atEOF {
		return
	}

	// 'data' was entirely digits, and we are not at EOF, so signal
	// the Scanner to keep going.
	return 0, nil, nil
}

func (i *interpreter) inNum() {
	s := bufio.NewScanner(i)
	s.Split(splitFunc)
	if s.Scan() {
		n, err := strconv.Atoi(string(s.Bytes()))
		if err != nil {
			i.Logger.Fatal(err)
		}
		i.push(n)
		return
	}
	if err := s.Err(); err != nil {
		i.Logger.Fatal(err)
	}
}

func (i *interpreter) inChar() {
	buf := make([]byte, 1)
	_, err := i.Read(buf)
	if err != nil {
		i.Logger.Fatal(err)
	}
	i.push(int(buf[0]))
}

func (i *interpreter) outNum() {
	io.WriteString(i, strconv.Itoa(i.pop()))
}

func (i *interpreter) outChar() {
	i.Write([]byte{byte(i.pop())})
}

// getColorBlock returns the current color block.
func (i *interpreter) getColorBlock() (block colorBlock) {
	currentColor := i.color()
	block = map[image.Point]struct{}{
		i.pos: struct{}{},
	}
	// very naive implementation currently. At the very least we should be able
	// to cache the current block.
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

// Whether the interpreter is currently able to move, without changing its DP or CC.
func (i interpreter) canMove() bool {
	newPos := i.pos.Add(image.Point(i.dp))
	return newPos.In(i.img.Bounds()) &&
		!sameColors(color.Black, i.img.At(newPos.X, newPos.Y))
}

// move causes the interpreter to attempt execute a single move.
// Returns whether the move was successful.
func (i *interpreter) move() bool {
	// First, move to the edge of the current block.
	i.moveWithinBlock()
	// Then, try to move into the next block.
	if i.canMove() {
		newPos := i.pos.Add(image.Point(i.dp))
		oldColor := i.color()
		blockSize := len(i.getColorBlock())

		i.pos = newPos
		i.colorChange(oldColor, blockSize)

		return true
	}

	return i.recovery()
}

func (i *interpreter) recovery() bool {
	i.Logger.Println("entering recovery.")
	originalDp := i.dp
	originalCc := i.cc

	// When true, toggle the CC. When false, rotate the DP.
	cc := true

	for !i.canMove() {
		if cc {
			i.toggleCc()
		} else {
			i.rotateDp()
		}
		if i.dp == originalDp && i.cc == originalCc {
			i.Logger.Println("Failed recovery")
			return false
		}

		i.moveWithinBlock()
		i.Logger.Println(i)
		cc = !cc
	}

	i.Logger.Println("recovered.")
	return true
}

// hue: 0=red, 1=yellow, etc.
// lightness: 0=light, 1=normal, 2=dark
func colorInfo(c color.Color) (hue, lightness int) {
	r, g, b, _ := c.RGBA()
	switch {
	case r == 0xFFFF && g == 0xC0C0 && b == 0xC0C0:
		return 0, 0
	case r == 0xFFFF && g == 0xFFFF && b == 0xC0C0:
		return 1, 0
	case r == 0xC0C0 && g == 0xFFFF && b == 0xC0C0:
		return 2, 0
	case r == 0xC0C0 && g == 0xFFFF && b == 0xFFFF:
		return 3, 0
	case r == 0xC0C0 && g == 0xC0C0 && b == 0xFFFF:
		return 4, 0
	case r == 0xFFFF && g == 0xC0C0 && b == 0xFFFF:
		return 5, 0

	case r == 0xFFFF && g == 0x0000 && b == 0x0000:
		return 0, 1
	case r == 0xFFFF && g == 0xFFFF && b == 0x0000:
		return 1, 1
	case r == 0x0000 && g == 0xFFFF && b == 0x0000:
		return 2, 1
	case r == 0x0000 && g == 0xFFFF && b == 0xFFFF:
		return 3, 1
	case r == 0x0000 && g == 0x0000 && b == 0xFFFF:
		return 4, 1
	case r == 0xFFFF && g == 0x0000 && b == 0xFFFF:
		return 5, 1

	case r == 0xC0C0 && g == 0x0000 && b == 0x0000:
		return 0, 2
	case r == 0xC0C0 && g == 0xC0C0 && b == 0x0000:
		return 1, 2
	case r == 0x0000 && g == 0xC0C0 && b == 0x0000:
		return 2, 2
	case r == 0x0000 && g == 0xC0C0 && b == 0xC0C0:
		return 3, 2
	case r == 0x0000 && g == 0x0000 && b == 0xC0C0:
		return 4, 2
	case r == 0xC0C0 && g == 0x0000 && b == 0xC0C0:
		return 5, 2
	default:
		// This should only be called for the 18 colors with a well-defined
		// hue and lightness, not for white, black, or any other color.
		panic(c)
	}
}

// Called when the color changes to cause the interpreter to do things.
func (i *interpreter) colorChange(prevColor color.Color, blockSize int) {
	if sameColors(color.White, prevColor) || sameColors(color.White, i.color()) {
		return
		i.Logger.Println(i, "Moving to/from white: no command to execute")
	}

	oldHue, oldLightness := colorInfo(prevColor)
	newHue, newLightness := colorInfo(i.color())

	hueChange := (newHue - oldHue + 6) % 6
	lightnessChange := (newLightness - oldLightness + 3) % 3

	i.Logger.Println(i, "ΔH:", hueChange, "ΔL:", lightnessChange)
	switch lightnessChange {
	case 0:
		switch hueChange {
		case 1:
			i.add()
		case 2:
			i.divide()
		case 3:
			i.greater()
		case 4:
			i.duplicate()
		case 5:
			i.inChar()
		}
	case 1:
		switch hueChange {
		case 0:
			i.push(blockSize)
		case 1:
			i.subtract()
		case 2:
			i.mod()
		case 3:
			i.pointer()
		case 4:
			i.roll()
		case 5:
			i.outNum()
		}
	case 2:
		switch hueChange {
		case 0:
			i.pop()
		case 1:
			i.multiply()
		case 2:
			i.not()
		case 3:
			i.switchCc()
		case 4:
			i.inNum()
		case 5:
			i.outChar()
		}
	}
	i.Logger.Println("  stack:", i.stack.data)
}

func (i *interpreter) Run() {
	for i.move() {
	}
}

func (i *interpreter) moveWithinBlock() {
	if sameColors(color.White, i.color()) {
		newPos := i.pos.Add(image.Point(i.dp))
		for sameColors(color.White, i.img.At(newPos.X, newPos.Y)) {
			i.pos = newPos
			newPos = i.pos.Add(image.Point(i.dp))
		}
		return
	}

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
	i.pos = *newPos
}

func sameColors(c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return r1 == r2 &&
		g1 == g2 &&
		b1 == b2
}
