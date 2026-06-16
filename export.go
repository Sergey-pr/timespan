package main

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

var statusLabels = map[TaskStatus]string{
	StatusReadyToStart: "Ready to start",
	StatusActive:       "Active",
	StatusPaused:       "Paused",
	StatusFinished:     "Finished",
}

// formatDuration renders elapsed milliseconds as HH:MM:SS.
func formatDuration(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	totalSec := ms / 1000
	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	s := totalSec % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// formatTime renders a time pointer as a local datetime, or empty string if nil.
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04")
}

// liveElapsed returns the task's elapsed time including the active running segment.
func liveElapsed(t Task) int64 {
	if t.Status == StatusActive && t.StartedAt != nil {
		return t.ElapsedMs + time.Since(*t.StartedAt).Milliseconds()
	}
	return t.ElapsedMs
}

// buildReport creates an xlsx report of all tasks and writes it to path.
func buildReport(path string) error {
	tasks, err := GetTasks()
	if err != nil {
		return err
	}
	cats, err := GetCategories()
	if err != nil {
		return err
	}

	catNames := make(map[int64]string, len(cats))
	for _, c := range cats {
		catNames[c.ID] = c.Name
	}

	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Отчёт"
	idx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	f.SetActiveSheet(idx)
	f.DeleteSheet("Sheet1")

	headers := []string{
		"Название", "Категория", "Статус",
		"Дата создания", "Дата завершения", "Затраченное время",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E6E6F0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err == nil {
		f.SetCellStyle(sheet, "A1", "F1", headerStyle)
	}

	for r, t := range tasks {
		row := r + 2
		category := ""
		if t.CategoryID != nil {
			category = catNames[*t.CategoryID]
		}
		values := []interface{}{
			t.Title,
			category,
			statusLabels[t.Status],
			formatTime(&t.CreatedAt),
			formatTime(t.FinishedAt),
			formatDuration(liveElapsed(t)),
		}
		for c, v := range values {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	widths := []float64{32, 18, 14, 18, 18, 18}
	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	return f.SaveAs(path)
}
