package app

import (
	"bytes"
	"context"
	"image"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/disintegration/imaging"
	log "github.com/sirupsen/logrus"
)

type bucketImage struct {
	BucketName		string 	`json:"bucketName" binding:"required"`
	Filename 		string 	`json:"filename" binding:"required"`
	MaxWidth 		int  	`json:"maxWidth"`
	MaxHeight 		int  	`json:"maxHeight"`
	PortraitOption 	bool 	`json:"portraitOption"`
}

// DownloadAndResize get image from gco and resize to client
func DownloadAndResize(c *gin.Context) {

	var req bucketImage
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "json decoding : " + err.Error(),
			"status": http.StatusBadRequest,
		})
		return
	}
	
	ctx := context.Background()
	bucket, err := getBucket(ctx, req)
	if err != nil { abort(c, "Failed load bucket", err); return }

	buf, err := getImageFromGCP(ctx, bucket, req.Filename)
	if err != nil { abort(c, "Failed get image from bucket", err); return }

	imageInfo, _, err := image.DecodeConfig(bytes.NewBuffer(buf.Bytes()))
	if err != nil { abort(c, "Failed to get image dimentions", err); return }
	
	log.Info("Original image width: ", imageInfo.Width, ", height: ", imageInfo.Height)
	log.Info("Original image size: ", buf.Len())
	
	if (shouldResize(&imageInfo, req.MaxWidth, req.MaxHeight)) {
		w, h := getDimentions(req.MaxWidth, req.MaxHeight, req.PortraitOption)
		resizeImageAndWriteResponse(c, buf.Bytes(), w, h)
		return
	}
	writeResponse(c, buf.Bytes())
}

func getDimentions(width, height int, PortraitOption bool) (int, int) {
	if width > height && PortraitOption {
		return height, height
	}
	return width, height
}

func shouldResize(imageInfo *image.Config, maxWidth, maxHeight int) bool {
	if (maxWidth != 0 && imageInfo.Width > maxWidth) { return true} 
	if (maxHeight != 0 && imageInfo.Height > maxHeight) { return true} 
	return false;
}

func writeResponse(c *gin.Context, buf []byte) {
	img, err := imaging.Decode(bytes.NewBuffer(buf))
	// compress original image
	encodedImageJpg, err := encodeImageToJpg(&img)
	if err != nil { abort(c, "Failed to encode image", err); return }
	contentType := http.DetectContentType(buf)
	c.DataFromReader(200, int64(encodedImageJpg.Len()), contentType, encodedImageJpg, map[string]string{})
}

func resizeImageAndWriteResponse(c *gin.Context, buf []byte, maxWidth, maxHeight int) {
	img, err := resize(bytes.NewBuffer(buf), maxWidth, maxHeight)
	if err != nil { abort(c, "Failed to resize image", err); return }
	encodedImageJpg, err := encodeImageToJpg(&img)
	if err != nil { abort(c, "Failed to encode image", err); return }
	contentType := http.DetectContentType(buf)
	log.Info("Resized image size: ", encodedImageJpg.Len())
	c.DataFromReader(200, int64(encodedImageJpg.Len()), contentType, encodedImageJpg, map[string]string{})
}

func abort(c *gin.Context, msg string, err error) {
	log.Error(msg," ", err)
	c.AbortWithStatusJSON(500, gin.H{
		"error": err.Error(),
	})
}
