package main

import "testing"

func TestCountTasksInCategory(t *testing.T) {
	app := newTestApp(t)

	cat := app.CreateCategory("Work")
	if cat == nil {
		t.Fatal("CreateCategory returned nil")
	}

	if count, err := CountTasksInCategory(cat.ID); err != nil || count != 0 {
		t.Fatalf("empty category: count=%d err=%v, want 0 and no error", count, err)
	}

	app.CreateTask("first", "", cat.ID)
	app.CreateTask("second", "", cat.ID)
	app.CreateTask("uncategorised", "", 0)

	count, err := CountTasksInCategory(cat.ID)
	if err != nil {
		t.Fatalf("CountTasksInCategory: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestDeleteCategoryRefusesWhileTasksReferenceIt(t *testing.T) {
	app := newTestApp(t)

	cat := app.CreateCategory("Work")
	app.CreateTask("still here", "", cat.ID)

	if app.DeleteCategory(cat.ID) {
		t.Error("DeleteCategory succeeded while a task still uses the category")
	}
	if _, err := GetCategoryByID(cat.ID); err != nil {
		t.Errorf("category was deleted anyway: %v", err)
	}
}

func TestDeleteCategoryWhenUnused(t *testing.T) {
	app := newTestApp(t)

	cat := app.CreateCategory("Spare")

	if !app.DeleteCategory(cat.ID) {
		t.Fatal("DeleteCategory returned false for an unused category")
	}
	if _, err := GetCategoryByID(cat.ID); err == nil {
		t.Error("category still readable after delete")
	}
}
