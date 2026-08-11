package main

import (
	"strings"
	"testing"
)

func TestTaskTitleMustNotBeBlank(t *testing.T) {
	newTestApp(t)

	for _, title := range []string{"", "   ", "\t\n"} {
		task := Task{Title: title, Status: StatusReadyToStart}
		if err := task.Save(); err == nil {
			t.Errorf("saving title %q succeeded, want an error", title)
		}
	}
}

func TestTaskTitleIsTrimmed(t *testing.T) {
	app := newTestApp(t)

	task := app.CreateTask("  spaced out  ", "", 0)
	if task == nil {
		t.Fatal("CreateTask returned nil")
	}
	if task.Title != "spaced out" {
		t.Errorf("title = %q, want %q", task.Title, "spaced out")
	}
}

func TestCreateTaskRejectsBlankTitle(t *testing.T) {
	app := newTestApp(t)

	if task := app.CreateTask("   ", "", 0); task != nil {
		t.Errorf("CreateTask returned %+v, want nil", task)
	}

	tasks, err := GetTasks()
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("%d tasks stored, want none", len(tasks))
	}
}

func TestCategoryNameMustNotBeBlank(t *testing.T) {
	newTestApp(t)

	cat := Category{Name: "  "}
	if err := cat.Save(); err == nil {
		t.Error("saving a blank category name succeeded, want an error")
	}
}

func TestDuplicateCategoryReportsFriendlyError(t *testing.T) {
	newTestApp(t)

	first := Category{Name: "Work"}
	if err := first.Save(); err != nil {
		t.Fatalf("first save: %v", err)
	}

	duplicate := Category{Name: "Work"}
	err := duplicate.Save()
	if err == nil {
		t.Fatal("duplicate name saved, want an error")
	}
	if strings.Contains(err.Error(), "UNIQUE constraint") {
		t.Errorf("raw sqlite error reached the user: %v", err)
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v, want it to say the category already exists", err)
	}
}

func TestRenameToExistingCategoryReportsFriendlyError(t *testing.T) {
	newTestApp(t)

	work := Category{Name: "Work"}
	if err := work.Save(); err != nil {
		t.Fatalf("save Work: %v", err)
	}
	home := Category{Name: "Home"}
	if err := home.Save(); err != nil {
		t.Fatalf("save Home: %v", err)
	}

	home.Name = "Work"
	err := home.Save()
	if err == nil {
		t.Fatal("renaming onto an existing name succeeded, want an error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v, want it to say the category already exists", err)
	}
}
