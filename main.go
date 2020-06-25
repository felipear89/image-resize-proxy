package main

import (
	app "image-resize-proxy/app"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)


func main() {
	r := gin.New()
	
	if gin.IsDebugging() {
		r.Use(gin.Logger())
	} else {
		log.SetFormatter(&log.JSONFormatter{})
		r.Use(app.JSONLogMiddlewarego())
	}

	r.Use(gin.Recovery())
	r.POST("/google/bucket/download", app.DownloadAndResize)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})
	log.Info("Starting image-resize-proxy")
	r.Run()
}
