package main

import (
	"bufio"
	"os"
    "testing"
     app "image-resize-proxy/app"

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
    image, err := app.Resize(r, 0, 600)
    check(err)
    encodedImageJpg, err := app.EncodeImageToJpg(&image)
    assert.Equal(28094, encodedImageJpg.Len())
}
