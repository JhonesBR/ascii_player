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
	"sync"
	"time"

	"github.com/nsf/termbox-go"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"golang.org/x/term"
)

const FRAME_RATE = 30
const FRAME_BUFFER_SIZE = 50

// FPS calculation
var (
	lastFrameTime time.Time
	mu            sync.Mutex
)

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

	// Channel for stopping playback on Enter key press
	stop := make(chan struct{})

	// Processing buffer
	buffer := make(chan image.Image, FRAME_BUFFER_SIZE)
	var wg sync.WaitGroup
	wg.Add(1)

	// Start processing frames
	go streamFrames(videoPath, buffer, &wg, stop)

	// Wait for buffer to be filled
	valuesReady := 0
	for valuesReady < FRAME_BUFFER_SIZE {
		select {
		case msg, ok := <-buffer:
			if !ok {
				// Wait
				if valuesReady >= FRAME_BUFFER_SIZE {
					return
				} else {
					continue
				}
			}

			// Put the value back into the channel
			valuesReady++
			buffer <- msg
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	// Start consuming values
	go consumeValues(buffer, FRAME_RATE, terminalWidth, terminalHeight, stop)

	go func() {
		wg.Wait()
		close(buffer)
	}()

	// Listen for Enter key press to stop playback
	go func() {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				if ev.Key == termbox.KeyEnter {
					close(stop)
					return
				}
			case termbox.EventError:
				log.Fatal(ev.Err)
			}
		}
	}()

	<-stop
}

func streamFrames(videoPath string, buffer chan image.Image, wg *sync.WaitGroup, stop chan struct{}) {
	defer wg.Done()

	cmd := ffmpeg_go.Input(videoPath).
		Output("pipe:", ffmpeg_go.KwArgs{
			"format":    "image2pipe",
			"vcodec":    "png",
			"loglevel":  "quiet",
			"framerate": FRAME_RATE,
		})

	var videoBuffer bytes.Buffer

	err := cmd.WithOutput(&videoBuffer).Run()
	if err != nil {
		log.Fatal("Error running ffmpeg:", err)
		close(buffer)
		return
	}

	reader := bytes.NewReader(videoBuffer.Bytes())
	for {
		select {
		case <-stop:
			return
		default:
			img, err := png.Decode(reader)
			if err == io.EOF {
				close(stop) // Signal stop when video finishes
				return
			}
			if err != nil {
				close(stop) // Signal stop on error
				return
			}

			// Wait for buffer to have space
			select {
			case buffer <- img:
			case <-stop:
				return
			}
		}
	}
}

func consumeValues(ch chan image.Image, fps, terminalWidth, terminalHeight int, stop chan struct{}) {
	ticker := time.NewTicker(time.Second / time.Duration(fps))
	defer ticker.Stop()
	for {
		select {
		case img, ok := <-ch:
			if !ok {
				return
			}
			loadAndDisplayFrame(img, terminalWidth, terminalHeight)
			<-ticker.C
		case <-stop:
			return
		}
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
	mu.Lock()
	defer mu.Unlock()

	// FPS Calculation
	now := time.Now()
	if !lastFrameTime.IsZero() {
		elapsed := now.Sub(lastFrameTime).Seconds()
		fps := 1 / elapsed

		fmt.Fprintf(os.Stdout, "\033[H\033[2K[Press ENTER to stop the video] FPS: %.2f\n", fps)
	}
	lastFrameTime = now

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for y, row := range frame.Chars {
		for x, char := range row {
			termbox.SetCell(x, y+1, rune(char[0]), termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	err := termbox.Flush()
	if err != nil {
		return err
	}

	return nil
}
