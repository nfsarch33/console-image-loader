package console_image

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"image"
	"image/draw"
	"image/gif"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/mat/besticon/ico"
	"github.com/mattn/go-isatty"
)

// Create a new instance of the logger. You can have any number of instances.
var log = logrus.New()

func loadImage(filePath string) []byte {
	var err error
	var buf []byte

	if filePath == "" {
		if buf, err = io.ReadAll(os.Stdin); err != nil {
			log.Panicf("failed to loadImage the stdin: %v", err)
		}
	} else {
		if buf, err = os.ReadFile(filePath); err != nil {
			log.Panicf("failed to loadImage the filePath file: %v", err)
		}
	}

	return buf
}

func decode(buf []byte) []image.Image {
	mime, err := mimetype.DetectReader(bytes.NewReader(buf))
	if err != nil {
		log.Panicf("failed to detect the mime type: %v", err)
	}

	allowed := []string{"image/gif", "image/png", "image/jpeg", "image/bmp", "image/x-icon"}
	if !mimetype.EqualsAny(mime.String(), allowed...) {
		log.Fatal("invalid MIME type")
	}

	frames := make([]image.Image, 0)

	if mime.Is("image/gif") {
		gifImage, err := gif.DecodeAll(bytes.NewReader(buf))

		if err != nil {
			log.Panicf("failed to decode the gif: %v", err)
		}

		var lowestX int
		var lowestY int
		var highestX int
		var highestY int

		for _, img := range gifImage.Image {
			if img.Rect.Min.X < lowestX {
				lowestX = img.Rect.Min.X
			}
			if img.Rect.Min.Y < lowestY {
				lowestY = img.Rect.Min.Y
			}
			if img.Rect.Max.X > highestX {
				highestX = img.Rect.Max.X
			}
			if img.Rect.Max.Y > highestY {
				highestY = img.Rect.Max.Y
			}
		}

		imgWidth := highestX - lowestX
		imgHeight := highestY - lowestY

		overPaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
		draw.Draw(overPaintImage, overPaintImage.Bounds(), gifImage.Image[0], image.Point{}, draw.Src)

		for _, srcImg := range gifImage.Image {
			draw.Draw(overPaintImage, overPaintImage.Bounds(), srcImg, image.Point{}, draw.Over)
			frame := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
			draw.Draw(frame, frame.Bounds(), overPaintImage, image.Point{}, draw.Over)
			frames = append(frames, frame)
		}

		return frames
	}

	var frame image.Image

	if mime.Is("image/x-icon") {
		frame, err = ico.Decode(bytes.NewReader(buf))
	} else {
		frame, _, err = image.Decode(bytes.NewReader(buf))
	}

	if err != nil {
		log.Panicf("failed to decode the image: %v", err)
	}

	imb := frame.Bounds()
	if imb.Max.X < 2 || imb.Max.Y < 2 {
		log.Fatal("the input image is to small")
	}

	return append(frames, frame)
}

func scale(frames []image.Image) []image.Image {
	type data struct {
		i  int
		im image.Image
	}

	var err error

	cols := DefaultTermCols
	rows := DefaultTermRows

	if isatty.IsTerminal(os.Stdout.Fd()) {
		if rows, cols, err = pty.Getsize(os.Stdout); err != nil {
			log.Panicf("failed to get the terminal size: %v", err)
		}
	}

	w := cols * ResizeFactorX
	h := (rows - ResizeOffsetY) * ResizeFactorY

	l := len(frames)
	r := make([]image.Image, l)
	c := make(chan *data, l)

	for i, f := range frames {
		go func(i int, f image.Image) {
			c <- &data{i, imaging.Fit(f, w, h, imaging.Lanczos)}
		}(i, f)
	}

	for range r {
		d := <-c
		r[d.i] = d.im
	}

	return r
}

func escape(frames []image.Image) ([][]string, error) {
	type data struct {
		i   int
		str string
	}

	escaped := make([][]string, 0)

	for _, f := range frames {
		imb := f.Bounds()
		maxY := imb.Max.Y - imb.Max.Y%2
		maxX := imb.Max.X

		c := make(chan *data, maxY/2)
		lines := make([]string, maxY/2)

		for y := 0; y < maxY; y += 2 {
			go func(y int) {
				var sb strings.Builder

				for x := 0; x < maxX; x++ {
					r, g, b, a := f.At(x, y).RGBA()
					if a>>8 < 128 {
						_, err := sb.WriteString(AnsiBgTransparentColor)
						if err != nil {
							log.Error(errors.New(err.Error()))
						}
					} else {
						_, err := sb.WriteString(fmt.Sprintf(AnsiBgRgbColor, r>>8, g>>8, b>>8))
						if err != nil {
							log.Error(errors.New(err.Error()))
						}
					}

					r, g, b, a = f.At(x, y+1).RGBA()
					if a>>8 < 128 {
						_, err := sb.WriteString(AnsiFgTransparentColor)
						if err != nil {
							log.Error(errors.New(err.Error()))
						}
					} else {
						_, err := sb.WriteString(fmt.Sprintf(AnsiFgRgbColor, r>>8, g>>8, b>>8))
						if err != nil {
							log.Error(errors.New(err.Error()))
						}
					}
				}

				_, err := sb.WriteString(AnsiReset)
				if err != nil {
					log.Error(errors.New(err.Error()))
				}
				_, err = sb.WriteString("\n")
				if err != nil {
					log.Error(errors.New(err.Error()))
				}

				c <- &data{y / 2, sb.String()}
			}(y)
		}

		for range lines {
			line := <-c
			lines[line.i] = line.str
		}

		escaped = append(escaped, lines)
	}

	return escaped, nil
}

func display(frames [][]string) error {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		defer enableEcho(disableEcho())
	}

	_, err := os.Stdout.WriteString(AnsiCursorHide)
	if err != nil {
		return fmt.Errorf("failed to write the string: %w", err)
	}
	_, err = os.Stdout.WriteString("\n")
	if err != nil {
		return fmt.Errorf("failed to write the string: %w", err)
	}

	frameCount := len(frames)

	if frameCount == 1 {
		os.Stdout.WriteString(strings.Join(frames[0], ""))
	} else {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		tick := time.Tick(time.Second / time.Duration(FPS))
		h := len(frames[0]) + 2 // two extra lines for the exit msg
		playing := true

		go func() {
			<-c
			playing = false
		}()

		for i := 0; playing; i++ {
			if i != 0 {
				_, err := os.Stdout.WriteString(fmt.Sprintf(AnsiCursorUp, h))
				if err != nil {
					return fmt.Errorf("failed to write the string: %w", err)
				}
			}

			_, err := os.Stdout.WriteString(strings.Join(frames[i%frameCount], ""))
			if err != nil {
				return fmt.Errorf("failed to write the frame rate string: %w", err)
			}
			_, err = os.Stdout.WriteString("\npress `ctrl c` to exit\n")
			if err != nil {
				return fmt.Errorf("failed to write the press contrl c string: %w", err)
			}

			<-tick
		}
	}

	_, err = os.Stdout.WriteString(AnsiCursorShow)
	if err != nil {
		return err
	}

	return nil
}

func ShowImg(imagePath string) error {
	esc, err := escape(scale(decode(loadImage(imagePath))))
	if err != nil {
		return err
	}
	err = display(esc)
	if err != nil {
		return err
	}

	return nil
}
