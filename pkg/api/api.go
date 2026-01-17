package api

import "net/http"

func Init(mux *http.ServeMux) {
	if mux == nil {
		return
	}
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.HandleFunc("/api/task", taskHandler)
	mux.HandleFunc("/api/task/done", doneTaskHandler)
	mux.HandleFunc("/api/tasks", tasksHandler)
}
