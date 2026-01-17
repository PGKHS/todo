package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo/pkg/db"
)

type errorResponse struct {
	Error string `json:"error"`
}

type idResponse struct {
	ID string `json:"id"`
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, errorResponse{Error: err.Error()})
		return
	}

	if strings.TrimSpace(task.Title) == "" {
		writeJSON(w, errorResponse{Error: "missing title"})
		return
	}

	if err := checkDate(&task); err != nil {
		writeJSON(w, errorResponse{Error: err.Error()})
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, idResponse{ID: strconv.FormatInt(id, 10)})
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.FormValue("id"))
	if id == "" {
		writeJSON(w, errorResponse{Error: "Не указан идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			writeJSON(w, errorResponse{Error: "Задача не найдена"})
			return
		}
		writeJSON(w, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, errorResponse{Error: err.Error()})
		return
	}

	if strings.TrimSpace(task.ID) == "" {
		writeJSON(w, errorResponse{Error: "Не указан идентификатор"})
		return
	}

	if strings.TrimSpace(task.Title) == "" {
		writeJSON(w, errorResponse{Error: "missing title"})
		return
	}

	if err := checkDate(&task); err != nil {
		writeJSON(w, errorResponse{Error: err.Error()})
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		if errors.Is(err, db.ErrTaskNotFound) {
			writeJSON(w, errorResponse{Error: "Задача не найдена"})
			return
		}
		writeJSON(w, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, struct{}{})
}

func checkDate(task *db.Task) error {
	now := time.Now()
	if strings.TrimSpace(task.Date) == "" {
		task.Date = now.Format(dateLayout)
	}

	t, err := time.ParseInLocation(dateLayout, task.Date, time.Local)
	if err != nil {
		return err
	}

	next := ""
	if strings.TrimSpace(task.Repeat) != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	if afterNow(now, t) {
		if strings.TrimSpace(task.Repeat) == "" {
			task.Date = now.Format(dateLayout)
		} else {
			task.Date = next
		}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_ = json.NewEncoder(w).Encode(data)
}
