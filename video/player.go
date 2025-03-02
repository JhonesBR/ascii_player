package video

import (
	"ascii/player/processing"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"time"

	"github.com/nsf/termbox-go"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"golang.org/x/term"
)

func StreamFrames(videoPath string, processFrame func(image.Image)) error {
	cmd := ffmpeg_go.Input(videoPath).
		Output("pipe:", ffmpeg_go.KwArgs{"format": "image2pipe", "vcodec": "png"})

	var buffer bytes.Buffer

	err := cmd.WithOutput(&buffer).Run()
	if err != nil {
		return fmt.Errorf("error running ffmpeg: %w", err)
	}

	reader := bytes.NewReader(buffer.Bytes())
	frameIndex := 0
	for {
		img, err := png.Decode(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println("Error decoding frame:", err)
			break
		}

		processFrame(img) // Process each frame
		frameIndex++
	}

	return nil
}

func Play(videoPath string) {
	terminalWidth, terminalHeight, err := term.GetSize(0)
	if err != nil {
		fmt.Println("Error: Could not get terminal size")
		os.Exit(1)
	}

	err = termbox.Init()
	if err != nil {
		log.Fatal("Error initializing termbox:", err)
	}
	defer termbox.Close()

	err = StreamFrames(videoPath, func(img image.Image) {
		loadAndDisplayFrame(img, terminalWidth, terminalHeight)
	})

	if err != nil {
		log.Fatal("Error processing frames:", err)
	}
}

func loadAndDisplayFrame(img image.Image, terminalWidth, terminalHeight int) {
	frame, err := processing.GetFrame(img, uint(terminalWidth), uint(terminalHeight))
	if err != nil {
		log.Fatal("Error getting frame:", err)
	}

	err = displayFrame(frame)
	if err != nil {
		log.Fatal("Error displaying frame:", err)
	}
}

func displayFrame(frame *processing.Frame) error {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for y, row := range frame.Chars {
		for x, char := range row {
			termbox.SetCell(x, y, rune(char[0]), termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	err := termbox.Flush()
	if err != nil {
		return err
	}

	// ~30 FPS
	time.Sleep(33 * time.Millisecond)

	return nil
}
