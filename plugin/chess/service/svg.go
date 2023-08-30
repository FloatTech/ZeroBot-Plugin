package service

import (
	"image"
	"image/png"
	"os"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func SVG2PNG(svgPath, pngPath string) error {
	w, h := 720, 720
	in, err := os.Open(svgPath)
	if err != nil {
		return err
	}
	defer in.Close()
	icon, _ := oksvg.ReadIconStream(in)
	icon.SetTarget(0, 0, float64(w), float64(h))
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 1)
	out, err := os.Create(pngPath)
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, rgba)
}
