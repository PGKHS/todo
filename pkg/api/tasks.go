package api

import (
	"net/http"

	"todo/pkg/db"
)

type tasksResponse struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tasks, err := db.Tasks(50, r.FormValue("search"))
	if err != nil {
		writeJSON(w, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, tasksResponse{Tasks: tasks})
}
