// Copyright 2018 Miguel Angel Rivera Notararigo. All rights reserved.
// This source code was released under the MIT license.

package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/colornames"

	"github.com/ntrrg/tiler/pkg/tile"

	_ "golang.org/x/image/webp"
	_ "image/png"
)

// Output sizes
var (
	// Letter72 = image.Rect(0, 0, 612, 792)
	// Letter200 = image.Rect(0, 0, 1700, 2200)
	Letter300 = image.Rect(0, 0, 2550, 3300)
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

	var dst *tile.Tiler

	var n int8
	var partial bool

	for j, path := range images {
		i := int64(j % 4)

		if i == 0 {
			dst = tile.New(colornames.Map["white"], Letter300, 4)
		}

		file, err := os.Open(path)

		if err != nil {
			log.Fatalf("Can't open image %v -> %v", path, err)
		}

		defer file.Close()

		_, err = dst.DrawAt(file, i, nil)

		if err != nil {
			log.Fatalf("Can't decode the image %v -> %v", path, err)
		}

		if i == 3 {
			partial = false

			err := writeBlock(dst, n)

			if err != nil {
				log.Fatalf("Can't create the output file -> %v", err)
			}

			n++
		} else {
			partial = true
		}
	}

	if partial {
		err := writeBlock(dst, n)

		if err != nil {
			log.Fatalf("Can't create the output file -> %v", err)
		}
	}
}

func writeBlock(dst image.Image, n int8) error {
	imgFile, err := os.Create(fmt.Sprintf("output-%d.jpg", n))

	if err != nil {
		return err
	}

	defer imgFile.Close()

	return jpeg.Encode(imgFile, dst, nil)
}
