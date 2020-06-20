package main

import (
	"bytes"
	"context"
	"io"
	"time"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type bucketImage struct {
	BucketName	string `json:"bucketName"`
	Filename 	string `json:"filename"`
}

func getBucket(ctx context.Context, req bucketImage) (*storage.BucketHandle, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second * 10)
	defer cancel()
	bucket := client.Bucket(req.BucketName)
	return bucket, nil
}

func getImageFromGCP(ctx context.Context, bucket *storage.BucketHandle, filename string) (int64, *bytes.Buffer, error) {
	obj := bucket.Object(filename)

	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Error("Failed to create GCP Reader", err)
	}
	defer r.Close()
	
	buf := &bytes.Buffer{}
	size, err := io.Copy(buf, r) 
	if err != nil {
		return 0, nil, err
	}
	return size, buf, nil
}

func downloadAndResize(c *gin.Context) {

	var req bucketImage
	c.Bind(&req)
	
	ctx := context.Background()
	bucket, err := getBucket(ctx, req)
	if err != nil {
		log.Error("Failed load bucket", err)
		c.Abort()
		return
	}

	size, buf, err := getImageFromGCP(ctx, bucket, req.Filename)
	if err != nil {
		log.Error("Failed get image from bucket", err)
		c.Abort()
		return
	}
	contentType := http.DetectContentType(buf.Bytes())
	
	img, err := resize(buf, 400)
	if err != nil {
		log.Error("Failed resize image", err)
		c.Abort()
		return
	}
	encodedImageJpg, err := encodeImageToJpg(&img)
	
	c.DataFromReader(200, size, contentType, encodedImageJpg, map[string]string{})
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
	r.POST("/bucket/download", downloadAndResize)

	log.Info("Starting image-resize-proxy")
	r.Run() // listen and serve on 0.0.0.0:8080
}
