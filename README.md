# Console Image Loader

Load static images and gifs and display in console output


### Features

- Static image support
- Animated GIF support
- Transparency

### Installation

#### Download package to your project

```bash
go get github.com/nfsarch33/console-image-loader
```

### Usage

```go
package main

import (
	"fmt"
	"github.com/nfsarch33/console-image-loader/pkg/console_image"
)

func main() {
        // console_image.ShowImg() takes the image path param, change this to your image path.
	if err := console_image.ShowImg("pkg/console_image/test_image/goku.gif"); err != nil {
		fmt.Errorf("ShowImg() error = %v", err)
	}
}
```
