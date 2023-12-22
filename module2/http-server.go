package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Log status after execute request (middleware)
func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use custom ResponseWriter to catch status code, by default is 200
		lw := &loggingResponseWriter{w, http.StatusOK}
		next.ServeHTTP(lw, r)

		// Print ip, method, path, and status code
		clientIP := r.RemoteAddr
		method := r.Method
		path := r.URL.Path
		statusCode := lw.statusCode

		log.Printf("%s %s %s %d", clientIP, method, path, statusCode)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// Overwrite WriteHeader(code int) to save status code
func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func hello(w http.ResponseWriter, req *http.Request) {
	// Print response headers before giving request headers
	fmt.Println(w.Header())

	// Loop request headers
	for key, values := range req.Header {
		// Add value to header key one by one
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Get os env variable
	ENV_VERSION := os.Getenv("VERSION")

	// Add env variable to response header
	w.Header().Add("Os-Version", ENV_VERSION)

	// Print request headers
	fmt.Println(req.Header)
	// Print response headers after giving request headers
	fmt.Println(w.Header())
	io.WriteString(w, "Hello\n")

}

// Healthz route return status 200
func healthz(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	io.WriteString(w, "Ok\n")
}

func main() {
	// Set env variable
	os.Setenv("VERSION", "1.0")

	http.Handle("/hello", loggingHandler(http.HandlerFunc(hello)))
	http.Handle("/healthz", loggingHandler(http.HandlerFunc(healthz)))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
