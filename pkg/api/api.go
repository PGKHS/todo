package api

import "net/http"

func Init(mux *http.ServeMux) {
	if mux == nil {
		return
	}
	mux.HandleFunc("/api/nextdate", nextDateHandler)
}
