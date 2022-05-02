package grayscalor

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
)

func Convert(fromFile io.Reader, toFile io.Writer, quality int) error {
	fromImg, format, err := image.Decode(fromFile)
	if err != nil {
		return err
	}

	grayImg := image.NewGray(fromImg.Bounds())
	for x := fromImg.Bounds().Min.X; x < fromImg.Bounds().Max.X; x++ {
		for y := fromImg.Bounds().Min.Y; y < fromImg.Bounds().Max.Y; y++ {
			grayImg.Set(x, y, fromImg.At(x, y))
		}
	}

	switch format {
	case "jpeg":
		err = jpeg.Encode(toFile, grayImg, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(toFile, grayImg)
	}
	if err != nil {
		return err
	}

	return nil
}
