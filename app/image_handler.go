package app

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
)

func resize(reader io.Reader, maxWidth, maxHeight int) (image.Image, error) {
	
	image, err := imaging.Decode(reader)
	if err != nil { return nil, err }
	resizedImage := imaging.Resize(image, maxWidth, maxHeight, imaging.Lanczos)
	
	return resizedImage, nil
}


func encodeImageToJpg(img *image.Image) (*bytes.Buffer, error) {
	encoded := &bytes.Buffer{}
	err := jpeg.Encode(encoded, *img, &jpeg.Options{
		Quality: 75,
	})
	return encoded, err
}