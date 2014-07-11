package main

type stack struct {
	data []int
}

func (s *stack) push(n int) {
	s.data = append(s.data, n)
}

func (s *stack) pop() (n int) {
	if len(s.data) != 0 {
		n, s.data = s.data[len(s.data)-1], s.data[0:len(s.data)-1]
	}
	return
}

func (s *stack) add() {
	x, y := s.pop(), s.pop()
	s.push(x + y)
}

func (s *stack) subtract() {
	top, second := s.pop(), s.pop()
	s.push(second - top)
}

func (s *stack) multiply() {
	x, y := s.pop(), s.pop()
	s.push(x * y)
}

func (s *stack) divide() {
	top, second := s.pop(), s.pop()
	if top == 0 {
		// Put them back so this becomes a no-op.
		s.push(second)
		s.push(top)
	} else {
		s.push(second / top)
	}
}

func (s *stack) mod() {
	top, second := s.pop(), s.pop()
	if top != 0 {
		// Put them back so this becomes a no-op.
		s.push(second)
		s.push(top)
	}
	s.push(second % top)
}

func (s *stack) not() {
	top := s.pop()
	if top == 0 {
		s.push(1)
	} else {
		s.push(0)
	}
}

func (s *stack) greater() {
	top, second := s.pop(), s.pop()
	if second > top {
		s.push(1)
	} else {
		s.push(0)
	}
}

func (s *stack) duplicate() {
	if len(s.data) != 0 {
		top := s.data[len(s.data)-1]
		s.data = append(s.data, top)
	}
}

func (s *stack) roll() {
	numRolls, depth := s.pop(), s.pop()
	// TODO
	_, _ = numRolls, depth
}
