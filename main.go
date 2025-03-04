package main

import (
	"flag"

	"ascii/player/video"
)

func main() {
	// Define flags
	videoPath := flag.String("v", "video.mp4", "Path to the video file")
	frameRate := flag.Int("fr", 30, "Frame rate of the video")
	colored := flag.Bool("c", false, "Use colored output")

	// Parse flags
	flag.Parse()

	// Initialize settings with parsed flag values
	settings := video.Settings{
		VideoPath: *videoPath,
		FrameRate: *frameRate,
		Colored: *colored,
	}

	video.Play(settings)
}
