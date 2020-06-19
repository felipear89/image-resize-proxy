package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type bucketImage struct {
	BucketName	string `json:"bucketName"`
	Filename 	string `json:"filename"`
}

func getClientIP(c *gin.Context) string {
	requester := c.Request.Header.Get("X-Forwarded-For")
	if len(requester) == 0 {
		requester = c.Request.Header.Get("X-Real-IP")
	}
	if len(requester) == 0 {
		requester = c.Request.RemoteAddr
	}
	if strings.Contains(requester, ",") {
		requester = strings.Split(requester, ",")[0]
	}
	return requester
}

func getDurationInMillseconds(start time.Time) float64 {
	end := time.Now()
	duration := end.Sub(start)
	milliseconds := float64(duration) / float64(time.Millisecond)
	rounded := float64(int(milliseconds*100+.5)) / 100
	return rounded
}

func jsonLogMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
     
        start := time.Now()
        c.Next()
        duration := getDurationInMillseconds(start)

        entry := log.WithFields(log.Fields{
            "client_ip":  getClientIP(c),
            "duration":   duration,
            "method":     c.Request.Method,
            "path":       c.Request.RequestURI,
            "status":     c.Writer.Status(),
            "referrer":   c.Request.Referer(),
            "request_id": c.Writer.Header().Get("Request-Id"),
        })

        if c.Writer.Status() >= 500 {
            entry.Error(c.Errors.String())
        } else {
            entry.Info("")
        }
    }
}

func bucketDownload(c *gin.Context) {

	var req bucketImage
	c.Bind(&req)
	
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Error("Failed to create storage client", err)
		c.Abort()
		return
	}

	bucket := client.Bucket(req.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	
	obj := bucket.Object(req.Filename)

	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Error("Failed to create GCP Reader", err)
	}
	defer r.Close()
	
	buf := &bytes.Buffer{}
	nRead, err := io.Copy(buf, r) 
	if err != nil {
		log.Error("Failed to load image", err)
	}

	// TODO Discover image type to set contentType
	c.DataFromReader(200, nRead, "image/png", buf, map[string]string{})
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
	r.POST("/bucket/download", bucketDownload)

	log.Info("Starting image-resize-proxy")
	r.Run() // listen and serve on 0.0.0.0:8080
}
