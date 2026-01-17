package server

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

const defaultPort = 7540

// Run starts the HTTP file server serving static assets from webDir.
func Run(webDir string) error {
	port := defaultPort
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if parsed, err := strconv.Atoi(envPort); err == nil {
			port = parsed
		} else {
			log.Printf("invalid TODO_PORT %q, using default %d", envPort, defaultPort)
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	addr := ":" + strconv.Itoa(port)
	log.Printf("starting server on %s", addr)
	return http.ListenAndServe(addr, mux)
}
