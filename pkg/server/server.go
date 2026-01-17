package server

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"todo/pkg/api"
)

const defaultPort = 7540

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
	api.Init(mux)
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	addr := ":" + strconv.Itoa(port)
	log.Printf("starting server on http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
	return http.ListenAndServe(addr, nil)
}
