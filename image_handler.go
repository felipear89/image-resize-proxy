package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
)

func resize(reader io.Reader, width int) (image.Image, error) {
	
	image, err := imaging.Decode(reader)
	if err != nil {
		return nil, err
	}
	resizedImage := imaging.Resize(image, width, 0, imaging.Lanczos)
	
	return resizedImage, nil
}


func encodeImageToJpg(img *image.Image) (*bytes.Buffer, error) {
	encoded := &bytes.Buffer{}
	err := jpeg.Encode(encoded, *img, nil)
	return encoded, err
}