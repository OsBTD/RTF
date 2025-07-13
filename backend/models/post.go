package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Post struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Username   string     `json:"username"`
	UserImg    string     `json:"user_img"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	Categories []Category `json:"categories"` // category IDs
	CreatedAt  time.Time  `json:"created_at"`
}

type PostFilter struct {
	Target     string `json:"target"`
	StartID    int    `json:"start_id"`
	NPost      int    `json:"n_post"`
	CategoryID int    `json:"category_id"`
}

type PostModel struct {
	DB *sql.DB
}

// Insert Post
func (pm *PostModel) InsertPost(post Post) error {
	postQuery := `
		INSERT INTO posts (user_id, title, content)
		VALUES (?, ?, ?)
	`

	res, err := pm.DB.Exec(postQuery, post.UserID, post.Title, post.Content)
	if err != nil {
		return err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	catQuery := `
		INSERT OR IGNORE INTO post_categories (post_id, category_id)
		VALUES (?, ?)
	`
	for _, cat := range post.Categories {
		if _, err := pm.DB.Exec(catQuery, postID, cat.ID); err != nil {
			return err
		}
	}

	return nil
}

func ValidatePost(post *Post) error {
	if post == nil {
		return errors.New("post is nil")
	}

	post.Title = strings.TrimSpace(post.Title)
	if post.Title == "" {
		return errors.New("post.Title cannot be empty or whitespace")
	}
	if len(post.Title) > 70 {
		return errors.New("post.Title must be at most 70 characters long")
	}

	post.Content = strings.TrimSpace(post.Content)
	if post.Content == "" {
		return errors.New("post.Content cannot be empty or whitespace")
	}
	if len(post.Content) > 1000 {
		return errors.New("post.Content must be at most 1000 characters long")
	}

	categoryCount := len(post.Categories)
	if categoryCount < 1 || categoryCount > 3 {
		return errors.New("post must have between 1 and 3 categories")
	}

	return nil
}

// GetPostByID gets post and its author info
func (pm *PostModel) GetPostByID(id int) (Post, error) {
	postQuery := `
		SELECT 
			p.id, p.user_id, p.title, p.content, p.created_at,
			u.username, u.profile_img
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`

	var post Post
	err := pm.DB.QueryRow(postQuery, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.Username,
		&post.UserImg,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Post{}, fmt.Errorf("post with id %d not found", id)
		}
		return Post{}, fmt.Errorf("failed to get post [%d]: %w", id, err)
	}

	post.Categories, err = pm.GetPostCategoriesByPostID(post.ID)
	if err != nil {
		return Post{}, fmt.Errorf("failed to get categories for post %d: %w", post.ID, err)
	}

	return post, nil
}

func (pm *PostModel) GetLastPostID() (int, error) {
	var lastID int

	err := pm.DB.QueryRow(`SELECT id FROM posts ORDER BY id DESC LIMIT 1`).Scan(&lastID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No posts
			return -1, nil
		}
		return -1, fmt.Errorf("failed to get last post id: %w", err)
	}

	return lastID, nil
}

// FilterPosts supports filtering by feed (all), category (via join), or user posts with pagination
func (pm *PostModel) FilterPosts(filter *PostFilter, userID int) ([]Post, error, int) {
	var query string
	var args []any

	lastID, err := pm.GetLastPostID()
	if err != nil {
		return []Post{}, err, http.StatusInternalServerError
	}

	if filter.StartID == -1 {
		filter.StartID = lastID + 1
	}

	switch filter.Target {
	case "feed":
		query = `
			SELECT p.id, p.user_id, p.title, p.content, p.created_at,
			       u.username, u.profile_img
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.id < ?
			ORDER BY p.id DESC
			LIMIT ?
		`
		args = append(args, filter.StartID, filter.NPost)

	case "category":
		query = `
			SELECT p.id, p.user_id, p.title, p.content, p.created_at,
			       u.username, u.profile_img
			FROM posts p
			JOIN users u ON p.user_id = u.id
			JOIN post_categories pc ON p.id = pc.post_id
			WHERE pc.category_id = ? AND p.id < ?
			ORDER BY p.id DESC
			LIMIT ?
		`
		args = append(args, filter.CategoryID, filter.StartID, filter.NPost)

	case "user":
		query = `
			SELECT p.id, p.user_id, p.title, p.content, p.created_at,
			       u.username, u.profile_img
			FROM posts p
			JOIN users u ON p.user_id = u.id
			WHERE p.user_id = ? AND p.id < ?
			ORDER BY p.id DESC
			LIMIT ?
		`
		args = append(args, userID, filter.StartID, filter.NPost)

	default:
		return nil, errors.New("invalid target: must be 'feed', 'category' or 'user'"), http.StatusBadRequest
	}

	rows, err := pm.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error fetching posts: %w", err), http.StatusInternalServerError
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.Username,
			&post.UserImg,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning post: %w", err), http.StatusInternalServerError
		}

		// Load categories
		post.Categories, err = pm.GetPostCategoriesByPostID(post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load categories: %w", err), http.StatusInternalServerError
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over posts: %w", err), http.StatusInternalServerError
	}

	if len(posts) == 0 {
		return nil, nil, http.StatusNoContent
	}

	return posts, nil, http.StatusOK
}

// GetPostCategoriesByPostID fetches categories linked to a given post ID
func (pm *PostModel) GetPostCategoriesByPostID(postID int) ([]Category, error) {
	query := `
		SELECT c.id, c.icon, c.name, c.description, '' as target
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
	`

	rows, err := pm.DB.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories for post %d: %w", postID, err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Icon, &cat.Name, &cat.Description, &cat.Target); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}
	return categories, nil
}
