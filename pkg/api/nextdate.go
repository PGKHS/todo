package api

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "20060102"

// NextDate returns the next date in YYYYMMDD format based on the repeat rule.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if strings.TrimSpace(repeat) == "" {
		return "", errors.New("repeat rule is empty")
	}

	startDate, err := time.ParseInLocation(dateLayout, dstart, now.Location())
	if err != nil {
		return "", err
	}

	fields := strings.Fields(repeat)
	if len(fields) == 0 {
		return "", errors.New("repeat rule is empty")
	}

	switch fields[0] {
	case "d":
		return nextDateByDays(now, startDate, fields)
	case "y":
		return nextDateByYears(now, startDate, fields)
	case "w":
		return nextDateByWeekdays(now, startDate, fields)
	case "m":
		return nextDateByMonths(now, startDate, fields)
	default:
		return "", fmt.Errorf("unsupported repeat rule: %s", fields[0])
	}
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")
	if dateParam == "" || repeatParam == "" {
		http.Error(w, "missing date or repeat", http.StatusBadRequest)
		return
	}

	now := time.Now()
	if nowParam != "" {
		parsedNow, err := time.ParseInLocation(dateLayout, nowParam, time.Local)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		now = parsedNow
	}

	next, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, _ = w.Write([]byte(next))
}

func nextDateByDays(now time.Time, startDate time.Time, fields []string) (string, error) {
	if len(fields) != 2 {
		return "", errors.New("invalid day repeat format")
	}
	interval, err := strconv.Atoi(fields[1])
	if err != nil {
		return "", errors.New("invalid day interval")
	}
	if interval < 1 || interval > 400 {
		return "", errors.New("day interval out of range")
	}

	date := normalizeDate(startDate).AddDate(0, 0, interval)
	for !afterNow(date, now) {
		date = date.AddDate(0, 0, interval)
	}
	return date.Format(dateLayout), nil
}

func nextDateByYears(now time.Time, startDate time.Time, fields []string) (string, error) {
	if len(fields) != 1 {
		return "", errors.New("invalid yearly repeat format")
	}

	date := normalizeDate(startDate).AddDate(1, 0, 0)
	for !afterNow(date, now) {
		date = date.AddDate(1, 0, 0)
	}
	return date.Format(dateLayout), nil
}

func nextDateByWeekdays(now time.Time, startDate time.Time, fields []string) (string, error) {
	if len(fields) != 2 {
		return "", errors.New("invalid weekly repeat format")
	}

	weekdays, err := parseWeekdays(fields[1])
	if err != nil {
		return "", err
	}

	start := maxDate(normalizeDate(startDate).AddDate(0, 0, 1), normalizeDate(now).AddDate(0, 0, 1))
	for date := start; ; date = date.AddDate(0, 0, 1) {
		if weekdays[weekdayNumber(date)] {
			return date.Format(dateLayout), nil
		}
	}
}

func nextDateByMonths(now time.Time, startDate time.Time, fields []string) (string, error) {
	if len(fields) < 2 || len(fields) > 3 {
		return "", errors.New("invalid monthly repeat format")
	}

	days, err := parseMonthDays(fields[1])
	if err != nil {
		return "", err
	}

	months := map[int]bool{}
	if len(fields) == 3 {
		months, err = parseMonths(fields[2])
		if err != nil {
			return "", err
		}
	}

	start := maxDate(normalizeDate(startDate).AddDate(0, 0, 1), normalizeDate(now).AddDate(0, 0, 1))
	loc := start.Location()
	monthStart := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, loc)

	for current := monthStart; ; current = current.AddDate(0, 1, 0) {
		if len(months) > 0 && !months[int(current.Month())] {
			continue
		}

		minDate := current
		if current.Year() == start.Year() && current.Month() == start.Month() {
			minDate = start
		}

		daysInMonth := daysInMonth(current.Year(), current.Month(), loc)
		candidates := monthCandidates(days, daysInMonth)
		for _, day := range candidates {
			candidate := time.Date(current.Year(), current.Month(), day, 0, 0, 0, 0, loc)
			if !candidate.Before(minDate) {
				return candidate.Format(dateLayout), nil
			}
		}
	}
}

func parseWeekdays(value string) (map[int]bool, error) {
	items := strings.Split(value, ",")
	weekdays := map[int]bool{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, errors.New("empty weekday value")
		}
		day, err := strconv.Atoi(item)
		if err != nil {
			return nil, errors.New("invalid weekday")
		}
		if day < 1 || day > 7 {
			return nil, errors.New("weekday out of range")
		}
		weekdays[day] = true
	}
	if len(weekdays) == 0 {
		return nil, errors.New("weekday list is empty")
	}
	return weekdays, nil
}

func parseMonthDays(value string) ([]int, error) {
	items := strings.Split(value, ",")
	days := make([]int, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, errors.New("empty month day value")
		}
		day, err := strconv.Atoi(item)
		if err != nil {
			return nil, errors.New("invalid month day")
		}
		if day == -1 || day == -2 {
			days = append(days, day)
			continue
		}
		if day < 1 || day > 31 {
			return nil, errors.New("month day out of range")
		}
		days = append(days, day)
	}
	if len(days) == 0 {
		return nil, errors.New("month day list is empty")
	}
	return days, nil
}

func parseMonths(value string) (map[int]bool, error) {
	items := strings.Split(value, ",")
	months := map[int]bool{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			return nil, errors.New("empty month value")
		}
		month, err := strconv.Atoi(item)
		if err != nil {
			return nil, errors.New("invalid month")
		}
		if month < 1 || month > 12 {
			return nil, errors.New("month out of range")
		}
		months[month] = true
	}
	if len(months) == 0 {
		return nil, errors.New("month list is empty")
	}
	return months, nil
}

func monthCandidates(days []int, daysInMonth int) []int {
	candidateSet := map[int]bool{}
	for _, day := range days {
		switch day {
		case -1:
			candidateSet[daysInMonth] = true
		case -2:
			if daysInMonth > 1 {
				candidateSet[daysInMonth-1] = true
			}
		default:
			if day <= daysInMonth {
				candidateSet[day] = true
			}
		}
	}
	if len(candidateSet) == 0 {
		return nil
	}

	candidates := make([]int, 0, len(candidateSet))
	for day := range candidateSet {
		candidates = append(candidates, day)
	}
	sort.Ints(candidates)
	return candidates
}

func weekdayNumber(t time.Time) int {
	if t.Weekday() == time.Sunday {
		return 7
	}
	return int(t.Weekday())
}

func normalizeDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func afterNow(date time.Time, now time.Time) bool {
	return normalizeDate(date).After(normalizeDate(now))
}

func maxDate(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func daysInMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}
