package bilibili

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"rsc.io/qr"

	"github.com/ac1982/haul/internal/console"
)

// showQR prints a login QR code in the terminal and saves it as a PNG at path, for terminals where the printed one
// does not scan. The returned func removes the PNG.
func showQR(text, path string) (func(), error) {
	code, err := qr.Encode(text, qr.Q)
	if err != nil {
		return nil, err
	}
	if err := writeQRPNG(code, path, 7); err != nil {
		return nil, err
	}
	console.Info("Scan the QR code with the bilibili app " + console.CurrentStyle().Dim("(also saved as "+path+")"))
	console.Lines(qrLines(code))
	return func() { os.Remove(path) }, nil
}

// qrLines draws the code with two terminal cells per module and a quiet zone of two, in background colours so it
// scans on dark and light themes alike.
func qrLines(code *qr.Code) []string {
	const quiet = 2
	const dark, light, reset = "\x1b[40m  ", "\x1b[47m  ", "\x1b[0m"
	var lines []string
	for y := -quiet; y < code.Size+quiet; y++ {
		var b strings.Builder
		for x := -quiet; x < code.Size+quiet; x++ {
			if code.Black(x, y) {
				b.WriteString(dark)
			} else {
				b.WriteString(light)
			}
		}
		lines = append(lines, b.String()+reset)
	}
	return lines
}

// writeQRPNG saves the code as a greyscale PNG, scale pixels per module, with a quiet zone of four modules.
func writeQRPNG(code *qr.Code, path string, scale int) error {
	const quiet = 4
	side := (code.Size + 2*quiet) * scale
	img := image.NewGray(image.Rect(0, 0, side, side))
	for y := range side {
		for x := range side {
			c := color.Gray{Y: 0xFF}
			if code.Black(x/scale-quiet, y/scale-quiet) {
				c.Y = 0
			}
			img.SetGray(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
