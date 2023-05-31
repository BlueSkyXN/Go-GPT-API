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
	file, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	logger = log.New(file, "", log.LstdFlags)

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
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Printf("Error reading body: %v", err)
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		ip := c.ClientIP()
		startTime := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		logger.Printf("Request details: IP: %s, Start Time: %s, Method: %s, Path: %s, Request Body: %s",
			ip, startTime.Format(time.RFC1123), method, path, string(bodyBytes))

		for name, values := range c.Request.Header {
			for _, value := range values {
				logger.Printf("Request header: %s: %s\n", name, value)
			}
		}

		writer := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}

		c.Writer = writer

		c.Next()

		statusCode := c.Writer.Status()

		latency := time.Since(startTime)

		filteredBody := &bytes.Buffer{}

		lines := strings.Split(writer.body.String(), "\n")
		for _, line := range lines {
			if strings.Contains(line, `"status": "finished_successfully"`) {
				filteredBody.WriteString(line + "\n")
			}
		}

		logger.Printf("Response details: Status Code: %d, Latency: %v, Response Body: %s",
			statusCode, latency, filteredBody.String())

		for name, values := range writer.Header() {
			for _, value := range values {
				logger.Printf("Response header: %s: %s\n", name, value)
			}
		}
	}
}
