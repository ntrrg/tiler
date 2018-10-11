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
	"time"

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

	"hletter72":  image.Rect(0, 0, 792, 612),
	"hletter200": image.Rect(0, 0, 2200, 1700),
	"hletter300": image.Rect(0, 0, 3300, 2550),
}

func main() {
	start := time.Now()

	var (
		verbose bool
		debug   bool
		dryrun  bool
		tiles   int64
		bg      string
		reverse bool
		size    string
		output  string

		format = tile.DefaultFormat
	)

	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.BoolVar(&debug, "debug", false, "Enable debugging")

	flag.BoolVar(
		&dryrun,
		"dry-run",
		false,
		"Process the images but don't write to disk",
	)

	flag.BoolVar(
		&reverse,
		"reverse",
		false,
		"Tile the given images in reverse order",
	)

	flag.Int64Var(&tiles, "tiles", 4, "[WIP] Number of tiles, at least 2")
	flag.StringVar(&size, "size", "letter300", "Output file size")
	flag.StringVar(&bg, "bg", "white", "Output file background color")
	flag.StringVar(&format.Resize, "resize", "contain", "Resizing mode")
	flag.StringVar(&format.Align, "align", "center", "Horizontal alignment")
	flag.StringVar(&format.VAlign, "valign", "middle", "Vertical alignment")

	flag.StringVar(
		&output,
		"o",
		"output%d.jpg",
		"Output file, %d in file name is replaced by file number",
	)

	flag.Parse()

	if debug {
		verbose = true
	}

	log.SetFlags(0)
	images := flag.Args()

	ni := len(images)

	if ni < 2 {
		log.Fatalln("At least 1 image should be given")
	}

	if reverse {
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

	for i := int64(0); i < nt; i++ {
		a := i * tiles
		b := a + tiles

		if i == nt-1 && extra != 0 {
			b += extra - tiles
		}

		wt.Add(1)

		go func(nt int64, images []string) {
			if debug {
				fmt.Printf("Generating tiled image #%d using %v..\n", nt, images)
			}

			dst := tile.New(colornames.Map[bg], OutputSizes[size], tiles)

			for _, imgPath := range images {
				imgFile, err := os.Open(filepath.Clean(imgPath))

				if err != nil {
					log.Fatalf("Can't open image '%v' -> %v\n", imgPath, err)
				}

				if debug {
					fmt.Printf("Writing image '%s at tiled image #%d'..\n", imgPath, nt)
				}

				_, err = dst.Draw(imgFile, format)

				if err != nil && err != io.EOF {
					closeFile(imgPath, imgFile)
					log.Fatalf("Can't decode the image '%v' -> %v\n", imgPath, err)
				}

				closeFile(imgPath, imgFile)

				if debug {
					fmt.Printf("Image '%s' written at tiled image #%d\n", imgPath, nt)
				}
			}

			name := filepath.Clean(fmt.Sprintf(output, nt))

			if debug {
				fmt.Printf("Tiled image #%d generated, writing to '%s'\n", nt, name)
			}

			if !dryrun {
				imgFile, err := os.Create(name)

				if err != nil {
					log.Fatalf("Can't create the output file -> %v\n", err)
				}

				defer closeFile(name, imgFile)

				err = jpeg.Encode(imgFile, dst, nil)

				if err != nil {
					log.Fatalf("Can't encode the output file -> %v\n", err)
				}
			}

			if debug {
				fmt.Printf("Tiled image #%d has been written ('%s')\n", nt, name)
			}

			wt.Done()
		}(i, images[a:b])
	}

	wt.Wait()

	if verbose {
		fmt.Println("Used options:")

		if reverse {
			fmt.Println("  Reverse mode: true")
		}

		fmt.Printf("  Name: %s\n", output)
		fmt.Printf("  Size: %s\n", size)
		fmt.Printf("  Background color: %s\n", bg)
		fmt.Printf("  Tiles: %d\n", tiles)
		fmt.Printf("    Resize mode: %s\n", format.Resize)
		fmt.Printf("    Alignment: %s\n", format.Align)
		fmt.Printf("    Vertical alignment: %s\n", format.VAlign)

		format = tile.DefaultFormat

		if nt > 1 {
			fmt.Printf(
				"\n%d files generated from %d images in %s\n",
				nt,
				ni,
				time.Since(start),
			)
		}
	}
}

func closeFile(name string, file *os.File) {
	err := file.Close()

	if err != nil {
		log.Printf("Can't close the file '%v' -> %v\n", name)
	}
}
