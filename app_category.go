package main

import "fmt"

// GetCategories returns all categories sorted alphabetically.
func (a *App) GetCategories() []Category {
	cats, err := ListCategories()
	if err != nil {
		a.showError(err)
		return []Category{}
	}
	return cats
}

// CreateCategory creates a new category with the given name.
func (a *App) CreateCategory(name string) *Category {
	cat := &Category{Name: name}
	if err := cat.Save(); err != nil {
		a.showError(err)
		return nil
	}
	return cat
}

// RenameCategory updates the name of an existing category.
func (a *App) RenameCategory(id int64, name string) *Category {
	cat, err := GetCategoryByID(id)
	if err != nil {
		a.showError(err)
		return nil
	}
	cat.Name = name
	if err := cat.Save(); err != nil {
		a.showError(err)
		return nil
	}
	return cat
}

// DeleteCategory removes a category. Returns false if any tasks still reference it.
func (a *App) DeleteCategory(id int64) bool {
	count, err := CountTasksInCategory(id)
	if err != nil {
		a.showError(err)
		return false
	}
	if count > 0 {
		a.showError(fmt.Errorf("cannot delete category: %d task(s) still use it", count))
		return false
	}
	cat, err := GetCategoryByID(id)
	if err != nil {
		a.showError(err)
		return false
	}
	if err := cat.Delete(); err != nil {
		a.showError(err)
		return false
	}
	return true
}
