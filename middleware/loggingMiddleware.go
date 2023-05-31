package middleware

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	logger *log.Logger
)

func init() {
	// Initialize the logger
	file, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	logger = log.New(file, "", log.LstdFlags)

	// Initialize Gin
	gin.ForceConsoleColor()
	gin.SetMode(gin.ReleaseMode)
}

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

		// Read the request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Printf("Error reading body: %v", err)
			return
		}

		// Restore the body to its original state for the next middleware
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Get the request details
		ip := c.ClientIP()
		startTime := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Log the request details
		logger.Printf("Request details: IP: %s, Start Time: %s, Method: %s, Path: %s, Request Body: %s",
			ip, startTime.Format(time.RFC1123), method, path, string(bodyBytes))

		// Log the request headers
		for name, values := range c.Request.Header {
			// Loop over all values for the name.
			for _, value := range values {
				logger.Printf("Request header: %s: %s\n", name, value)
			}
		}

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

		// Only log the response details if the response body does not contain "status": "in_progress"
		if !strings.Contains(writer.body.String(), "\"status\": \"in_progress\"") {
			// Log the response details
			logger.Printf("Response details: Status Code: %d, Latency: %v, Response Body: %s",
				statusCode, latency, writer.body.String())

			// Log the response headers
			for name, values := range writer.Header() {
				// Loop over all values for the name.
				for _, value := range values {
					logger.Printf("Response header: %s: %s\n", name, value)
				}
			}
		}
	}
}
