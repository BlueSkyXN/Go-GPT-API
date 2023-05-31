package middleware

import (
	"log"
	"net/http"
	"os"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Open the log file
		f, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		// Set the log output to the file
		log.SetOutput(f)

		// Get the request details
		ip := r.RemoteAddr
		time := time.Now().Format(time.RFC1123)
		host := r.Host
		headers := r.Header
		auth := r.Header.Get("Authorization")
		path := r.URL.Path
		body := r.Body

		// Log the request details
		log.Printf("Request details: IP: %s, Time: %s, Host: %s, Headers: %v, Auth: %s, Path: %s, Body: %v",
			ip, time, host, headers, auth, path, body)

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)

		// Get the response details and log them
		// Note: You'll need to replace this with code to get the actual response details,
		// as this example only gets the status code. Getting the response body in a middleware
		// is a bit more complex, as you need to use a ResponseWriter wrapper to capture it.
		// status := w.WriteHeader()
		// log.Printf("Response details: Status: %s, Headers: %v, Body: %v", status, headers, body)
	})
}
