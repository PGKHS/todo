package db

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	dateLayout       = "20060102"
	searchDateLayout = "02.01.2006"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

var ErrTaskNotFound = errors.New("task not found")

func AddTask(task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

func Tasks(limit int, search string) ([]*Task, error) {
	if limit <= 0 {
		limit = 50
	}

	var (
		rows *sql.Rows
		err  error
	)

	search = strings.TrimSpace(search)
	if search == "" {
		rows, err = DB.Query(
			`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`,
			limit,
		)
	} else if date, ok := parseSearchDate(search); ok {
		rows, err = DB.Query(
			`SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?`,
			date,
			limit,
		)
	} else {
		like := "%" + search + "%"
		rows, err = DB.Query(
			`SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`,
			like,
			like,
			limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*Task, 0)
	for rows.Next() {
		var (
			task Task
			id   int64
		)
		if err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		task.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTask(id string) (*Task, error) {
	taskID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}

	var (
		task Task
		dbID int64
	)
	err = DB.QueryRow(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`,
		taskID,
	).Scan(&dbID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	task.ID = strconv.FormatInt(dbID, 10)

	return &task, nil
}

func UpdateTask(task *Task) error {
	taskID, err := strconv.ParseInt(task.ID, 10, 64)
	if err != nil {
		return err
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, taskID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func UpdateDate(next string, id string) error {
	taskID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(query, next, taskID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func DeleteTask(id string) error {
	taskID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	res, err := DB.Exec(`DELETE FROM scheduler WHERE id = ?`, taskID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func parseSearchDate(value string) (string, bool) {
	parsed, err := time.ParseInLocation(searchDateLayout, value, time.Local)
	if err != nil {
		return "", false
	}
	return parsed.Format(dateLayout), true
}
