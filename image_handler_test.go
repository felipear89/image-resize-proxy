package main

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/disintegration/imaging"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func TestResizeImage(t *testing.T) {
    f, err := os.Open("./images/img-to-test.jpg")
    check(err)
    r := bufio.NewReader(f)
    image, err := resize(r, 600)
    check(err)
    encodedImageJpg, err := encodeImageToJpg(&image)

    fmt.Println(encodedImageJpg.Len())

    imaging.Save(image, "./images/img-to-test-result.jpg")
}
