package main

import (
	"bufio"
	"os"
	"testing"

    "github.com/stretchr/testify/assert"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

//imaging.Save(image, "./images/img-to-test-result.jpg")
func TestResizeImage(t *testing.T) {
    assert := assert.New(t)
    f, err := os.Open("./images/img-to-test.jpg")
    check(err)
    r := bufio.NewReader(f)
    image, err := resize(r, 600)
    check(err)
    encodedImageJpg, err := encodeImageToJpg(&image)
    assert.Equal(28094, encodedImageJpg.Len())
}
