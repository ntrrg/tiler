// Copyright 2018 Miguel Angel Rivera Notararigo. All rights reserved.
// This source code was released under the MIT license.

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "image/png"
)

// Output sizes
var (
	// Letter72 = image.Rect(0, 0, 612, 792)
	// Letter200 = image.Rect(0, 0, 1700, 2200)
	Letter300 = image.Rect(0, 0, 2550, 3300)
)

// Background colors
var (
	White = color.RGBA{255, 255, 255, 255}
	// Red   = color.RGBA{255, 0, 0, 255}
	// Green = color.RGBA{0, 255, 0, 255}
	// Blue  = color.RGBA{0, 0, 255, 255}
)

func main() {
	images := []string{}

	for _, p := range []string{"*.jpg", "*.jpeg", "*.png"} {
		imgs, err := filepath.Glob(p)

		if err != nil {
			log.Fatal("Bad GLOB pattern")
		}

		images = append(images, imgs...)
	}

	var dst *image.RGBA
	size := Letter300

	var i, n int8
	var r image.Rectangle
	var partial bool

	for _, path := range images {
		file, err := os.Open(strings.Replace(path, " ", "\\ ", 0))

		if err != nil {
			log.Fatalf("Can't open image %v -> %v", path, err)
		}

		img, _, err := image.Decode(file)

		if err != nil {
			log.Fatalf("Can't decode the image %v -> %v", path, err)
		}

		switch i {
		case 0:
			dst = image.NewRGBA(size)
			draw.Draw(dst, dst.Bounds(), &image.Uniform{White}, image.ZP, draw.Src)

			r = image.Rect(0, 0, size.Max.X/2, size.Max.Y/2)
		case 1, 3:
			r = r.Add(image.Pt(r.Max.X, 0))
		case 2:
			r = r.Sub(image.Pt(r.Min.X, -r.Max.Y))
		}

		draw.Draw(dst, r, img, image.ZP, draw.Src)

		if i == 3 {
			i = 0
			partial = false
			imgFile, err := os.Create(fmt.Sprintf("output-%d.jpg", n))

			if err != nil {
				log.Fatalf("Can't create the output file -> %v", err)
			}

			if err := jpeg.Encode(imgFile, dst, nil); err != nil {
				imgFile.Close()
				log.Fatalf("Can't create the image -> %v", err)
			}

			imgFile.Close()
			n++
		} else {
			partial = true
			i++
		}
	}

	if partial {
		imgFile, err := os.Create(fmt.Sprintf("output-%d.jpg", n))

		if err != nil {
			log.Fatalf("Can't create the output file -> %v", err)
		}

		if err := jpeg.Encode(imgFile, dst, nil); err != nil {
			imgFile.Close()
			log.Fatalf("Can't create the image -> %v", err)
		}

		imgFile.Close()
	}
}
