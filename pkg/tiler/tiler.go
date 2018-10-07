// Copyright 2018 Miguel Angel Rivera Notararigo. All rights reserved.
// This source code was released under the MIT license.

// Package tiler provides tilling image composition.
package tiler

import (
	"image"
	"image/color"
	"io"

	"golang.org/x/image/draw"
)

// Default scaler
var scaler draw.Scaler = draw.ApproxBiLinear

// DefaultFormat is a set of commonly used format options and may be used as
// a starter point for custom format options.
var DefaultFormat = &Format{
	Align:  "left",
	VAlign: "top",
	Resize: "auto",
}

// Tiler is an image that supports tilling.
type Tiler interface {
	draw.Image
	io.Seeker

	// TileAt draws a tile using the r data in off position with f format,
	// returns the used decoder (see image.Decode) and an error, if any. When the
	// last position has been used, Tile returns io.EOF as error.
	TileAt(r io.Reader, off int64, f *Format) (string, error)

	// Tile is like TileAt, but it draws at the next position from the current
	// offset.
	Tile(r io.Reader, f *Format) (string, error)
}

type tiler struct {
	draw.Image

	bg   color.Color
	grid Grid
	off  int64
}

func (t *tiler) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (t *tiler) TileAt(r io.Reader, off int64, f *Format) (string, error) {
	img, df, err := image.Decode(r)

	if err != nil {
		return df, err
	}

	if f == nil {
		f = DefaultFormat
	}

	var tile image.Rectangle
	size := t.Bounds()

	switch off {
	case 0:
		tile = image.Rect(0, 0, size.Max.X/2, size.Max.Y/2)
	case 1:
		tile = image.Rect(size.Max.X/2, 0, size.Max.X, size.Max.Y/2)
	case 2:
		tile = image.Rect(0, size.Max.Y/2, size.Max.X/2, size.Max.Y)
	case 3:
		tile = image.Rect(size.Max.X/2, size.Max.Y/2, size.Max.X, size.Max.Y)
	}

	switch f.Resize {
	case "auto":
		r := img.Bounds()
		ls := r.Dx() > r.Dy()

		if (ls && r.Dx() > tile.Dx()) || (!ls && r.Dy() > tile.Dy()) {
			img = scaleImage(img, tile)
		}
	case "fill":
		img = scaleImage(img, tile)
	}

	draw.Draw(t, tile, img, image.ZP, draw.Src)
	return df, nil
}

func (t *tiler) Tile(r io.Reader, f *Format) (string, error) {
	return "", nil
}

// New returns a RGBA image that implements Tiler.
func New(bg color.Color, size image.Rectangle, tiles int) Tiler {
	if tiles%2 != 0 {
		tiles++
	}

	img := image.NewRGBA(size)
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	return &tiler{img, bg, Grid(tiles), 0}
}

// Grid provides grids composition utilities.
type Grid int

// Format is a set of format options used by Tiler for drawing a tile.
type Format struct {
	Align  string
	VAlign string
	Resize string
}

// SetScaler sets the scaler used to resize tiles.
func SetScaler(s draw.Scaler) {
	scaler = s
}

// scaleImage scales a to fill b.
func scaleImage(a, b image.Image) image.Image {
	ar := a.Bounds()
	br := b.Bounds()
	ax := float64(ar.Dx())
	ay := float64(ar.Dy())
	bx := float64(br.Dx())
	by := float64(br.Dy())

	var s float64
	ls := ax > ay

	if ls {
		if ax == bx {
			return a
		} else if ax > bx {
			s = 1 / (ax / bx)
		} else {
			s = bx / ax
		}
	} else {
		if ay == by {
			return a
		} else if ay > by {
			s = 1 / (ay / by)
		} else {
			s = by / ay
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, int(ax*s), int(ay*s)))
	scaler.Scale(dst, dst.Bounds(), a, a.Bounds(), draw.Over, nil)
	return dst
}
