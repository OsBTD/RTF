package models

import (
	"database/sql"
	"fmt"
)

type Category struct {
	ID          int    `json:"id"`
	Icon        string `json:"icon"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Target string `json:"target"`
}

type CategoryModel struct {
	DB *sql.DB
}

// Insert Category 
func (cm *CategoryModel) InsertCategory(category Category) error {
	query := `INSERT OR IGNORE INTO categories (name, description, icon) VALUES (?, ?, ?)`

	_, err := cm.DB.Exec(query, category.Name, category.Description, category.Icon)
	if err != nil {
		return fmt.Errorf("failed to insert category: %w", err)
	}

	return nil
}

// Update Category
func (cm *CategoryModel) UpdateCategory(category Category) error {
	query := `UPDATE categories SET name = ?, description = ?, icon = ? WHERE id = ?`

	_, err := cm.DB.Exec(query, category.Name, category.Description, category.Icon, category.ID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// Delete Category
func (cm *CategoryModel) DeleteCategory(categoryID int) error {
	query := `DELETE FROM categories WHERE id = ?`

	_, err := cm.DB.Exec(query, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// GetCategoryByID returns a category by its ID
func (cm *CategoryModel) GetCategoryByID(categoryID int) (*Category, error) {
	query := `SELECT id, name, description, icon FROM categories WHERE id = ?`
	row := cm.DB.QueryRow(query, categoryID)

	var category Category
	err := row.Scan(&category.ID, &category.Name, &category.Description, &category.Icon)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category with id %d not found", categoryID)
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// GetAllCategories retrieves all categories by params
func (cm *CategoryModel) GetAllCategories() ([]Category, error) {
	query := `SELECT id, name, description, icon FROM categories`

	rows, err := cm.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.Icon); err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return categories, nil
}
