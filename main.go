// Copyright 2018 Miguel Angel Rivera Notararigo. All rights reserved.
// This source code was released under the MIT license.

package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/image/colornames"

	"github.com/ntrrg/tiler/pkg/tile"

	_ "golang.org/x/image/webp"
	_ "image/png"
)

// OutputSizes is a set of commonly used output sizes.
var OutputSizes = map[string]image.Rectangle{
	"letter72":  image.Rect(0, 0, 612, 792),
	"letter200": image.Rect(0, 0, 1700, 2200),
	"letter300": image.Rect(0, 0, 2550, 3300),
}

func main() {
	var (
		verbose bool
		tiles   int64
		bg      string
		reverse bool
		size    string
		output  string

		format = tile.DefaultFormat
	)

	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.Int64Var(&tiles, "n", 4, "[WIP] Number of tiles, at least 2.")

	flag.BoolVar(
		&reverse,
		"reverse",
		false,
		"Tile the given images in reverse order",
	)

	flag.StringVar(&size, "size", "letter300", "Output file size")
	flag.StringVar(&bg, "bg", "white", "Output file background color")
	flag.StringVar(&format.Resize, "resize", "auto", "Resizing mode")
	flag.StringVar(&format.Align, "align", "center", "Horizontal alignment")
	flag.StringVar(&format.VAlign, "valign", "middle", "Vertical alignment")

	flag.StringVar(
		&output,
		"o",
		"output%d.jpg",
		"Output file, %d in file name is replaced by file number",
	)

	flag.Parse()

	log.SetFlags(0)
	images := flag.Args()

	ni := len(images)

	if int64(ni) < tiles {
		log.Fatalf("At least %d images should be given\n", tiles)
	}

	if reverse {
		if verbose {
			fmt.Println("Using reverse mode")
		}

		for i, j := 0, ni-1; i < ni/2; i, j = i+1, j-1 {
			images[i], images[j] = images[j], images[i]
		}
	}

	var wt sync.WaitGroup
	nt := int64(ni) / tiles
	extra := int64(ni) % tiles

	if extra != 0 {
		nt++
	}

	if nt > 1 && verbose {
		fmt.Printf("%d files will be generated\n", nt)
	}

	for i := int64(0); i < nt; i++ {
		a := i * tiles
		b := a + tiles

		if i == nt-1 && extra != 0 {
			b += extra - tiles
		}

		wt.Add(1)

		go func(nt int64, images []string) {
			if verbose {
				fmt.Printf("Generating tiled image #%d using %v..\n", nt, images)
			}

			dst := tile.New(colornames.Map[bg], OutputSizes[size], tiles)

			for _, imgPath := range images {
				imgFile, err := os.Open(filepath.Clean(imgPath))

				if err != nil {
					log.Fatalf("Can't open image '%v' -> %v\n", imgPath, err)
				}

				if verbose {
					fmt.Printf("Writing image '%s'..\n", imgPath)
				}

				_, err = dst.Draw(imgFile, format)

				if err != nil && err != io.EOF {
					closeFile(imgPath, imgFile)
					log.Fatalf("Can't decode the image '%v' -> %v\n", imgPath, err)
				}

				closeFile(imgPath, imgFile)

				if verbose {
					fmt.Printf("Image '%s' written\n", imgPath)
				}
			}

			if verbose {
				fmt.Printf("Tiled image #%d generated\n", nt)
			}

			name := filepath.Clean(fmt.Sprintf(output, nt))
			imgFile, err := os.Create(name)

			if err != nil {
				log.Fatalf("Can't create the output file -> %v\n", err)
			}

			defer closeFile(name, imgFile)

			if verbose {
				fmt.Printf("Writing image #%d to '%s'..\n", nt, name)
			}

			err = jpeg.Encode(imgFile, dst, nil)

			if err != nil {
				log.Fatalf("Can't encode the output file -> %v\n", err)
			}

			if verbose {
				fmt.Printf("File '%s' written\n", name)
			}

			wt.Done()
		}(i, images[a:b])
	}

	wt.Wait()
}

func closeFile(name string, file *os.File) {
	err := file.Close()

	if err != nil {
		log.Printf("Can't close the file '%v' -> %v\n", name)
	}
}
