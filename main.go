package main

import (
	"fmt"
	"github.com/nfsarch33/console-image-loader/pkg/console_image"
)

func main() {
	if err := console_image.ShowImg("pkg/console_image/test_image/goku.gif"); err != nil {
		fmt.Errorf("ShowImg() error = %v", err)
	}
}
