package middleware

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

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

		// Read the request body
		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			return
		}

		// Restore the body to its original state for the next middleware
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// Get the request details
		ip := c.ClientIP()
		startTime := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Log the request details
		log.Printf("Request details: IP: %s, Start Time: %s, Method: %s, Path: %s, Request Body: %s",
			ip, startTime.Format(time.RFC1123), method, path, string(bodyBytes))

		// Create our custom response writer
		writer := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}

		// Replace the original response writer with our custom one
		c.Writer = writer

		// Process the request
		c.Next()

		// Get the response details
		statusCode := c.Writer.Status()

		// Calculate response time
		latency := time.Since(startTime)

		// Log the response details
		log.Printf("Response details: Status Code: %d, Latency: %v, Response Body: %s",
			statusCode, latency, writer.body.String())
	}
}
