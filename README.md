# Ascii Player

A Go project to display videos using ascii characters in the terminal.

## Demo

![Showcase video](showcase/showcase.gif)

## Features

- Display colored pixels
- Supports 256 colors
- Easy to use

## Options

|   **Option**   |  **Flag** | **Default** |                          **Description**                         |
|:--------------:|:---------:|:-----------:|:----------------------------------------------------------------:|
| Video path     | -v string | video.mp4   | Path to the video                                                |
| Colored output | -c        | False       | Enable colored display                                           |
| Fullfilled     | -f        | False       | Display characters at max alpha (Only works with colored output) |
| Frame rate     | -fr int   | 30          | Set framerate                                                    |

## Usage

Run the project:

```sh
./player
```
or
```sh
go run main.go
```