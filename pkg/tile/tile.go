// Copyright 2018 Miguel Angel Rivera Notararigo. All rights reserved.
// This source code was released under the MIT license.

// Package tile provides image tilling composition.
package tile

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
	Align:  "center",
	VAlign: "middle",
	Resize: "auto",
}

// SetScaler sets the scaler used to scale images written into tiles.
func SetScaler(s draw.Scaler) {
	scaler = s
}

// Tiler is an image that supports tilling.
type Tiler struct {
	draw.Image

	bg color.Color

	grid, off int64
}

// New returns a Tiler that produces blocks with bg background, s size and t
// tiles.
func New(bg color.Color, s image.Rectangle, t int64) *Tiler {
	if t%2 != 0 {
		t++
	}

	img := image.NewRGBA(s)
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	return &Tiler{
		Image: img,
		bg:    bg,
		grid:  t,
	}
}

// Seek implements io.Seeker.
func (t *Tiler) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

// DrawAt draws a tile using the r data in off position with f format, returns
// the used decoder (see image.Decode) and an error, if any.
func (t *Tiler) DrawAt(r io.Reader, off int64, f *Format) (string, error) {
	img, df, err := image.Decode(r)

	if err != nil {
		return df, err
	}

	if f == nil {
		f = DefaultFormat
	}

	var tile image.Rectangle
	tr := t.Bounds()

	switch off {
	case 0:
		tile = image.Rect(0, 0, tr.Max.X/2, tr.Max.Y/2)
	case 1:
		tile = image.Rect(0, tr.Max.Y/2, tr.Max.X/2, tr.Max.Y)
	case 2:
		tile = image.Rect(tr.Max.X/2, 0, tr.Max.X, tr.Max.Y/2)
	case 3:
		tile = image.Rect(tr.Max.X/2, tr.Max.Y/2, tr.Max.X, tr.Max.Y)
	}

	tile, img = f.Format(tile, img)

	draw.Draw(t, tile, img, image.ZP, draw.Src)
	return df, nil
}

// Draw is like DrawAt, but it draws at the next position from the current
// offset. When the last position has been used, DrawAt returns io.EOF as
// error.
func (t *Tiler) Draw(r io.Reader, f *Format) (string, error) {
	df, err := t.DrawAt(r, t.off, f)

	if err != nil {
		return df, err
	}

	if t.off == t.grid-1 {
		err = io.EOF
	}

	t.off++
	return df, err
}

// Format is a set of format options used by Tiler for drawing a tile.
type Format struct {
	Align  string
	VAlign string
	Resize string
}

// Format returns a tile and an image formatted with f format options.
func (f *Format) Format(tile image.Rectangle, img image.Image) (image.Rectangle, image.Image) {
	switch f.Resize {
	case "auto":
		r := img.Bounds()
		ls := r.Dx() > r.Dy()

		if (ls && r.Dx() > tile.Dx()) || (!ls && r.Dy() > tile.Dy()) {
			img = scaleAToB(img, tile)
		}
	case "fill":
		img = scaleAToB(img, tile)
	}

	return tile, img
}

// scaleAToB scales a to fill b.
func scaleAToB(a, b image.Image) image.Image {
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
