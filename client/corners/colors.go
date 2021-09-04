package corners

import (
	"image/color"
)

var (
	backgroundColor = color.RGBA{0xfa, 0xf8, 0xef, 0xff}
	frameColor      = color.RGBA{0xbb, 0xad, 0xa0, 0xff}
)

func tileColor(value int) color.Color {
	switch value {
	case 2, 4:
		return color.RGBA{0x77, 0x6e, 0x65, 0xff}
	case 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536:
		return color.RGBA{0xf9, 0xf6, 0xf2, 0xff}
	default:
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	}
}
