package main

import (
	"bytes"
	"context"
	"image"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type bucketImage struct {
	BucketName	string `json:"bucketName" binding:"required"`
	Filename 	string `json:"filename" binding:"required"`
	MaxWidth 	int  	`json:"maxWidth"`
}

func getBucket(ctx context.Context, req bucketImage) (*storage.BucketHandle, error) {
	client, err := storage.NewClient(ctx)
	if err != nil { return nil, err }
	ctx, cancel := context.WithTimeout(ctx, time.Second * 10)
	defer cancel()
	bucket := client.Bucket(req.BucketName)
	return bucket, nil
}

func getImageFromGCP(ctx context.Context, bucket *storage.BucketHandle, filename string) (*bytes.Buffer, error) {
	obj := bucket.Object(filename)

	r, err := obj.NewReader(ctx)
	if err != nil { return nil, err }
	defer r.Close()
	
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, r) 
	if err != nil { return nil, err }
	return buf, nil
}

func downloadAndResize(c *gin.Context) {

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
	
	log.Info("Original image width: ", imageInfo.Width)
	log.Info("Original image size: ", buf.Len())
	
	if (imageInfo.Width > req.MaxWidth) {
		resizeImageAndWriteResponse(c, buf.Bytes(), req.MaxWidth)
		return
	}
	writeResponse(c, buf.Bytes())
}

func writeResponse(c *gin.Context, buf []byte) {
	img, err := imaging.Decode(bytes.NewBuffer(buf))
	// compress original image
	encodedImageJpg, err := encodeImageToJpg(&img)
	if err != nil { abort(c, "Failed to encode image", err); return }
	contentType := http.DetectContentType(buf)
	c.DataFromReader(200, int64(encodedImageJpg.Len()), contentType, encodedImageJpg, map[string]string{})
}

func resizeImageAndWriteResponse(c *gin.Context, buf []byte, maxWidth int) {
	img, err := resize(bytes.NewBuffer(buf), maxWidth)
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

func main() {
	r := gin.New()
	
	if gin.IsDebugging() {
		r.Use(gin.Logger())
	} else {
		log.SetFormatter(&log.JSONFormatter{})
		r.Use(jsonLogMiddleware())
	}

	r.Use(gin.Recovery())
	r.POST("/google/bucket/download", downloadAndResize)

	log.Info("Starting image-resize-proxy")
	r.Run() // listen and serve on 0.0.0.0:8080
}
