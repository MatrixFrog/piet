package piet

import "testing"

func TestRoll1(t *testing.T) {
	s := stack{}
	s.push(10)
	s.push(20)
	s.push(30)
	s.push(2) // depth
	s.push(1) // numRolls

	s.roll()

	if s.pop() != 20 {
		t.Error(s.data)
	}
	if s.pop() != 30 {
		t.Error(s.data)
	}
	if s.pop() != 10 {
		t.Error(s.data)
	}
}

func TestRoll2(t *testing.T) {
	s := stack{}
	s.push(10)
	s.push(20)
	s.push(30)
	s.push(2) // depth
	s.push(4) // numRolls

	s.roll()

	if s.pop() != 30 {
		t.Error(s.data)
	}
	if s.pop() != 20 {
		t.Error(s.data)
	}
	if s.pop() != 10 {
		t.Error(s.data)
	}
}

func TestRoll3(t *testing.T) {
	s := stack{}
	s.push(10)
	s.push(100)
	s.push(108)
	s.push(108)
	s.push(3)
	s.push(3)
	s.push(3) // depth
	s.push(2) // numRolls

	s.roll()

	if s.pop() != 108 {
		t.Error(s.data)
	}
	if s.pop() != 3 {
		t.Error(s.data)
	}
	if s.pop() != 3 {
		t.Error(s.data)
	}
	if s.pop() != 108 {
		t.Error(s.data)
	}
	if s.pop() != 100 {
		t.Error(s.data)
	}
	if s.pop() != 10 {
		t.Error(s.data)
	}
}
