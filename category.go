package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
)

var categoryTable = goqu.T("categories")

type Category struct {
	ID        int64     `db:"id"         json:"id"         goqu:"skipinsert,skipupdate"`
	Name      string    `db:"name"       json:"name"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"  goqu:"skipupdate"`
}

// GetCategories returns all categories sorted alphabetically.
func GetCategories() ([]Category, error) {
	var cats []Category
	err := goquDB.From(categoryTable).
		Order(goqu.I("name").Asc()).
		ScanStructs(&cats)
	if cats == nil {
		cats = []Category{}
	}
	return cats, err
}

// GetCategoryByID returns the category with the given id, or an error if not found.
func GetCategoryByID(id int64) (*Category, error) {
	var cat Category
	found, err := goquDB.From(categoryTable).
		Where(goqu.C("id").Eq(id)).
		ScanStruct(&cat)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("category %d not found", id)
	}
	return &cat, nil
}

// Save inserts or updates the category. ID == 0 means a new record.
func (c *Category) Save() error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return errors.New("category name cannot be empty")
	}

	if c.ID == 0 {
		c.CreatedAt = time.Now()
		result, err := goquDB.Insert(categoryTable).Rows(c).Executor().Exec()
		if err != nil {
			return friendlyCategoryError(err, c.Name)
		}
		c.ID, err = result.LastInsertId()
		return err
	}
	_, err := goquDB.Update(categoryTable).
		Set(c).
		Where(goqu.C("id").Eq(c.ID)).
		Executor().Exec()
	return friendlyCategoryError(err, c.Name)
}

// friendlyCategoryError turns the raw UNIQUE constraint failure into something readable.
func friendlyCategoryError(err error, name string) error {
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return fmt.Errorf("category %q already exists", name)
	}
	return err
}

// Delete removes the category from the database.
func (c *Category) Delete() error {
	_, err := goquDB.Delete(categoryTable).
		Where(goqu.C("id").Eq(c.ID)).
		Executor().Exec()
	return err
}
