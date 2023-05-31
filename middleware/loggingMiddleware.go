package middleware

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Open the log file
		f, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		// Set the log output to the file
		log.SetOutput(f)

		// Get the request details
		ip := c.ClientIP()
		startTime := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Process request
		c.Next()

		// Get the response details
		statusCode := c.Writer.Status()

		// Calculate response time
		latency := time.Since(startTime)

		// Log the request and response details
		log.Printf("Request details: IP: %s, Start Time: %s, Method: %s, Path: %s, Status Code: %d, Latency: %v",
			ip, startTime.Format(time.RFC1123), method, path, statusCode, latency)
	}
}
