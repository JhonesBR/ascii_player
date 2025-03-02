package main

import (
	"ascii/player/processing"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/nsf/termbox-go"
	"golang.org/x/term"
)

func main() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)

	file, err := os.Open("./image.jpg")

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	terminalWidth, terminalHeight, err := term.GetSize(0)
	if err != nil {
		fmt.Println("Error: Could not get terminal size")
		os.Exit(1)
	}

	frame, err := processing.GetFrame(file, uint(terminalWidth), uint(terminalHeight))

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}

	err = displayFrame(frame)
	if err != nil {
		panic(err)
	}
}

func displayFrame(frame *processing.Frame) error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	defer termbox.Close()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for y, row := range frame.Chars {
		for x, char := range row {
			termbox.SetCell(x, y, rune(char[0]), termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	err = termbox.Flush()
	if err != nil {
		return err
	}

	// Wait for a key press to exit
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			return nil
		case termbox.EventError:
			return ev.Err
		}
	}
}
