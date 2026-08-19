package main

import (
	"fmt"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ExportReport prompts for a save location and writes an xlsx report of all tasks.
func (a *App) ExportReport() ExportResult {
	path, err := application.Get().Dialog.SaveFile().
		SetFilename(fmt.Sprintf("timespan-report-%s.xlsx", time.Now().Format("2006-01-02"))).
		AddFilter("Excel Workbook", "*.xlsx").
		CanCreateDirectories(true).
		PromptForSingleSelection()
	if err != nil {
		a.showError(err)
		return ExportResult{Status: ExportFailed}
	}
	if path == "" {
		return ExportResult{Status: ExportCancelled}
	}
	if err := buildReport(path); err != nil {
		a.showError(err)
		return ExportResult{Status: ExportFailed}
	}
	return ExportResult{Status: ExportSaved, Path: path}
}
