package io

import (
	"fmt"
	"time"

	"github.com/ptonlix/spokenai/pkg/color"
)

type Ioer interface {
	i()
	PrintlnRed(string)
	PrintlnYellow(string)
	PrintlnBlue(string)
	Println(string)
	Print(string)

	SlowPrint(string)
	GetInput(...interface{})
}

func NewIoer() Ioer {
	return &Console{}
}

type Console struct{}

func (c *Console) i() {
}
func (c *Console) PrintlnRed(content string) {
	fmt.Println(color.Red(content))
}

func (c *Console) PrintlnYellow(content string) {
	fmt.Println(color.Yellow(content))
}

func (c *Console) PrintlnBlue(content string) {
	fmt.Println(color.Blue(content))
}

func (c *Console) Println(content string) {
	fmt.Println(content)
}

func (c *Console) Print(content string) {
	fmt.Print(content)
}

func (c *Console) SlowPrint(content string) {
	r := []rune(content)
	for i := 0; i < len(r); i++ {
		fmt.Printf("%c", r[i])
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	// fmt.Print(content)
}

func (c *Console) GetInput(v ...interface{}) {
	fmt.Scanln(v...)
}
