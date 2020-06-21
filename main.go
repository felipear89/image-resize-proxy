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
	BucketName	string `json:"bucketName"`
	Filename 	string `json:"filename"`
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
	c.Bind(&req)
	
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
		img, err := resize(bytes.NewBuffer(buf.Bytes()), req.MaxWidth)
		if err != nil { abort(c, "Failed to resize image", err); return }
		encodedImageJpg, err := encodeImageToJpg(&img)
		if err != nil { abort(c, "Failed to encode image", err); return }
		contentType := http.DetectContentType(buf.Bytes())
		log.Info("Resized image size: ", encodedImageJpg.Len())
		c.DataFromReader(200, int64(encodedImageJpg.Len()), contentType, encodedImageJpg, map[string]string{})
	}
	img, err := imaging.Decode(bytes.NewBuffer(buf.Bytes()))
	encodedImageJpg, err := encodeImageToJpg(&img)
	if err != nil { abort(c, "Failed to encode image", err); return }
	contentType := http.DetectContentType(buf.Bytes())
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
