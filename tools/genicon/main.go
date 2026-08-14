package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
)

var sizes = []int{16, 20, 24, 32, 40, 48, 64, 128, 256}

func main() {
	src := "assets/logo.png"
	outIco := "assets/app.ico"
	outTray := "internal/ui/assets/tray.png"

	f, err := os.Open(src)
	if err != nil {
		fatal(err)
	}
	img, err := png.Decode(f)
	f.Close()
	if err != nil {
		fatal(fmt.Errorf("decode %s: %w", src, err))
	}
	base := image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(base, base.Bounds(), img, img.Bounds().Min, draw.Src)

	removeWhiteBackground(base)

	if err := os.MkdirAll(filepath.Dir(outIco), 0o755); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(outTray), 0o755); err != nil {
		fatal(err)
	}

	var icoBuf bytes.Buffer
	writeIco(&icoBuf, base)

	if err := os.WriteFile(outIco, icoBuf.Bytes(), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s (%d bytes)\n", outIco, icoBuf.Len())

	tray := resize(base, 32)
	var trayBuf bytes.Buffer
	if err := png.Encode(&trayBuf, tray); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(outTray, trayBuf.Bytes(), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s (%d bytes)\n", outTray, trayBuf.Len())
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func removeWhiteBackground(img *image.NRGBA) {
	b := img.Bounds()
	visited := make([]bool, b.Dx()*b.Dy())
	type pt struct{ x, y int }
	queue := []pt{}
	mark := func(x, y int) {
		if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
			return
		}
		idx := (y-b.Min.Y)*b.Dx() + (x - b.Min.X)
		if visited[idx] {
			return
		}
		visited[idx] = true
		c := img.NRGBAAt(x, y)
		if c.A > 0 && c.R > 245 && c.G > 245 && c.B > 245 {
			img.SetNRGBA(x, y, color.NRGBA{0, 0, 0, 0})
			queue = append(queue, pt{x, y})
		}
	}
	for x := b.Min.X; x < b.Max.X; x++ {
		mark(x, b.Min.Y)
		mark(x, b.Max.Y-1)
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		mark(b.Min.X, y)
		mark(b.Max.X-1, y)
	}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		mark(p.x+1, p.y)
		mark(p.x-1, p.y)
		mark(p.x, p.y+1)
		mark(p.x, p.y-1)
	}
}

func resize(src *image.NRGBA, s int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, s, s))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func writeIco(buf *bytes.Buffer, src *image.NRGBA) {
	type entry struct {
		size int
		data []byte
	}
	var entries []entry
	for _, s := range sizes {
		r := resize(src, s)
		var p bytes.Buffer
		if err := png.Encode(&p, r); err != nil {
			fatal(err)
		}
		entries = append(entries, entry{s, p.Bytes()})
	}

	hdr := make([]byte, 6)
	hdr[2] = 1
	binary.LittleEndian.PutUint16(hdr[4:], uint16(len(entries)))
	buf.Write(hdr)

	offset := 6 + 16*len(entries)
	for _, e := range entries {
		d := make([]byte, 16)
		w, h := byte(e.size), byte(e.size)
		if e.size >= 256 {
			w, h = 0, 0
		}
		d[0], d[1] = w, h
		d[4], d[5] = 1, 0
		binary.LittleEndian.PutUint16(d[6:], 32)
		binary.LittleEndian.PutUint32(d[8:], uint32(len(e.data)))
		binary.LittleEndian.PutUint32(d[12:], uint32(offset))
		buf.Write(d)
		offset += len(e.data)
	}
	for _, e := range entries {
		buf.Write(e.data)
	}
}
