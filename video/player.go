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

const FRAME_BUFFER_SIZE = 50

// FPS calculation
var (
	lastFrameTime time.Time
	mu            sync.Mutex
)

func Play(settings Settings) {
	terminalWidth, terminalHeight, err := term.GetSize(0)
	if err != nil {
		fmt.Println("Error: Could not get terminal size")
		os.Exit(1)
	}

	err = termbox.Init()
	if err != nil {
		log.Fatal("Error initializing termbox:", err)
	}
	if settings.Colored {
		termbox.SetOutputMode(termbox.OutputRGB)
	}
	defer termbox.Close()

	// Channel for stopping playback on Enter key press or video end
	stop := make(chan struct{})
	done := make(chan struct{})

	// Processing buffer
	buffer := make(chan image.Image, FRAME_BUFFER_SIZE)
	var wg sync.WaitGroup
	wg.Add(1)

	// Start processing frames
	go streamFrames(settings, buffer, &wg, stop)

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
	go consumeValues(settings, buffer, terminalWidth, terminalHeight, stop)

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

	// Wait for either Enter key or video completion
	select {
	case <-stop:
	case <-done:
	}
}

func streamFrames(settings Settings, buffer chan image.Image, wg *sync.WaitGroup, stop chan struct{}) {
	defer wg.Done()

	cmd := ffmpeg_go.Input(settings.VideoPath).
		Output("pipe:", ffmpeg_go.KwArgs{
			"format":    "image2pipe",
			"vcodec":    "png",
			"loglevel":  "quiet",
			"framerate": settings.FrameRate,
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
				return
			}
			if err != nil {
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

func consumeValues(settings Settings, ch chan image.Image, terminalWidth, terminalHeight int, stop chan struct{}) {
	ticker := time.NewTicker(time.Second / time.Duration(settings.FrameRate))
	defer ticker.Stop()
	for {
		select {
		case img, ok := <-ch:
			if !ok {
				close(stop)
				return
			}
			loadAndDisplayFrame(settings, img, terminalWidth, terminalHeight)
			<-ticker.C
		case <-stop:
			return
		}
	}
}

func loadAndDisplayFrame(settings Settings, img image.Image, terminalWidth, terminalHeight int) {
	frame, err := processing.GetFrame(img, uint(terminalWidth), uint(terminalHeight), settings.FullFilled && settings.Colored)
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

	var color termbox.Attribute
	var pixel processing.Pixel
	for y, row := range frame.Chars {
		for x, char := range row {
			pixel = frame.Pixels[y][x]
			color = termbox.RGBToAttribute(pixel.R, pixel.G, pixel.B)
			termbox.SetCell(x, y+1, rune(char[0]), color, termbox.ColorDefault)
		}
	}

	err := termbox.Flush()
	if err != nil {
		return err
	}

	return nil
}
