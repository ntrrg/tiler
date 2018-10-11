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
	Margin: 0,
	Align:  "center",
	VAlign: "middle",
	Resize: "contain",
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
	Margin int64
	Align  string
	VAlign string
	Resize string
}

// Format returns a tile and an image formatted with f format options.
func (f *Format) Format(tile image.Rectangle, img image.Image) (image.Rectangle, image.Image) {
	if f.Margin > 0 {
		tile.Min.X += int(f.Margin)
		tile.Min.Y += int(f.Margin)
		tile.Max.X -= int(f.Margin)
		tile.Max.Y -= int(f.Margin)
	}

	if f.Resize != "none" {
		img = scaleImage(img, tile, f.Resize)
	}

	// switch f.Algin {
	// case "center":
	// case "right":
	// }

	return tile, img
}

// scaleImage returns a scaled copy of a to b according to mode. There are
// three modes:
//
// * "auto": scales a if it is bigger.
//
// * "contain": scales a to fill b without overflow.
//
// * "cover": scales a to full fill b.
//
// If an invalid mode is given, a will be returned as is.
func scaleImage(a, b image.Image, mode string) image.Image {
	var x, y int
	ar := a.Bounds()
	br := b.Bounds()
	ax := ar.Dx()
	ay := ar.Dy()
	bx := br.Dx()
	by := br.Dy()
	xC := getScaleFactor(ax, bx)
	yC := getScaleFactor(ay, by)

	switch mode {
	case "auto":
		if xC >= 1 && yC >= 1 {
			return a
		}

		fallthrough
	case "contain":
		if xC == 1 || yC == 1 {
			return a
		} else if xC < yC {
			x, y = bx, int(float64(ay)*xC)
		} else {
			x, y = int(float64(ax)*yC), by
		}
	case "cover":
		if xC == 1 || yC == 1 {
			return a
		} else if xC < yC {
			x, y = int(float64(ax)*yC), by
		} else {
			x, y = bx, int(float64(ay)*xC)
		}
	default:
		return a
	}

	dst := image.NewRGBA(image.Rect(0, 0, x, y))
	scaler.Scale(dst, dst.Bounds(), a, a.Bounds(), draw.Over, nil)
	return dst
}

// getScaleFactor returns the scale factor for a to be similar to b.
func getScaleFactor(a, b int) (C float64) {
	x, y := float64(a), float64(b)

	if x == y {
		C = 1
	} else if x > y {
		C = 1 / (x / y)
	} else {
		C = y / x
	}

	return C
}
