package output

import (
	"fmt"
	"strconv"
)

type Output interface {
	Write([]fmt.Stringer) error
}

type String struct {
	str string
}

func (s *String) String() string {
	return s.str
}

func NewString(str string) *String {
	return &String{str}
}

type Int struct {
	intval int
}

func (i *Int) String() string {
	return strconv.Itoa(i.intval)
}

func NewInt(i int) *Int {
	return &Int{i}
}
