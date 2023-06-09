package middleware

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
)

var (
	logger         *log.Logger
	filterResponse bool
	headerFields   []string
	responseFields []string
	loggingEnabled bool
)

func init() {
	// Initialize the logger
	file, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	logger = log.New(file, "", log.LstdFlags)

	// Load configuration
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	filterResponse = cfg.Section("loggingMiddleware").Key("filterResponse").MustBool(true)
	headerFields = strings.Split(cfg.Section("loggingMiddleware").Key("headerFields").String(), ",")
	responseFields = strings.Split(cfg.Section("loggingMiddleware").Key("responseFields").String(), ",")
	loggingEnabled = cfg.Section("loggingMiddleware").Key("loggingEnabled").MustBool(true)

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
		// 如果日志记录被关闭，直接返回
		if !loggingEnabled {
			c.Next()
			return
		}
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
		//for name, values := range c.Request.Header {
		// Loop over all values for the name.
		//	for _, value := range values {
		//		logger.Printf("Request header: %s: %s\n", name, value)
		//	}
		//}
		for _, name := range headerFields {
			values, ok := c.Request.Header[name]
			if ok {
				for _, value := range values {
					logger.Printf("Request header: %s: %s\n", name, value)
				}
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

		filteredBody := &bytes.Buffer{}

		lines := strings.Split(writer.body.String(), "\n")
		for _, line := range lines {
			if filterResponse && strings.Contains(line, `"status": "finished_successfully"`) {
				filteredBody.WriteString(line + "\n")
			}
		}
		// Log the response details
		if filterResponse {
			logger.Printf("Response details: Status Code: %d, Latency: %v, Response Body: %s",
				statusCode, latency, filteredBody.String())
		} else {
			logger.Printf("Response details: Status Code: %d, Latency: %v, Response Body: %s",
				statusCode, latency, writer.body.String())
		}

		// Log the response headers
		//for name, values := range writer.Header() {
		// Loop over all values for the name.
		//	for _, value := range values {
		//		logger.Printf("Response header: %s: %s\n", name, value)
		//	}
		//}
		// Log the response headers
		for _, name := range responseFields {
			values, ok := writer.Header()[name]
			if ok {
				for _, value := range values {
					logger.Printf("Response header: %s: %s\n", name, value)
				}
			}
		}
	}
}
