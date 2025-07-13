package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Comment struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentsFilter struct {
	PostID   int `json:"post_id"`
	StartID  int `json:"start_id"`
	NComment int `json:"n_comment"`
}

type CommentModel struct {
	DB *sql.DB
}

func ValidateComment(comment *Comment) error {
	if comment == nil {
		return errors.New("comment is nil")
	}

	comment.Content = strings.TrimSpace(comment.Content)
	if comment.Content == "" {
		return errors.New("Comment cannot be empty or whitespace")
	}
	if len(comment.Content) > 1000 {
		return errors.New("Comment must be at most 1000 characters long")
	}

	return nil
}

// Insert Comment
func (cm *CommentModel) InsertComment(comment Comment) error {
	query := `
		INSERT OR IGNORE INTO comments (post_id, user_id, content)
		VALUES (?, ?, ?)`
	_, err := cm.DB.Exec(query, comment.PostID, comment.UserID, comment.Content)
	if err != nil {
		return err
	}
	return nil
}

// Update Comment
func (cm *CommentModel) UpdateComment(commentID int, newContent string) error {
	query := `
		UPDATE comments
		SET content = ?
		WHERE id = ?`
	res, err := cm.DB.Exec(query, newContent, commentID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no comment found with the given ID")
	}
	return nil
}

// Delete Comment
func (cm *CommentModel) DeleteComment(commentID int) error {
	query := `
		DELETE FROM comments
		WHERE id = ?`
	res, err := cm.DB.Exec(query, commentID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no comment found with the given ID")
	}
	return nil
}

// GetComment retrieves a comment by its ID
func (cm *CommentModel) GetComment(commentID int) (Comment, error) {
	query := `
		SELECT id, post_id, user_id, content, created_at
		FROM comments
		WHERE id = ?`

	row := cm.DB.QueryRow(query, commentID)
	var comment Comment
	err := row.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt)
	if err != nil {
		return Comment{}, err
	}

	return comment, nil
}

func (cm *CommentModel) GetLastCommentID(postID int) (int, error) {
	var lastID int

	query := `SELECT id FROM comments WHERE post_id = ? ORDER BY id DESC LIMIT 1`
	err := cm.DB.QueryRow(query, postID).Scan(&lastID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No comments for this post
			return -1, nil
		}
		return -1, fmt.Errorf("failed to get last comment id: %w", err)
	}

	return lastID, nil
}

// GetComments retrieves an array of comments by post ID
func (cm *CommentModel) GetComments(filter *CommentsFilter) ([]Comment, error) {
	lastID, err := cm.GetLastCommentID(filter.PostID)
	if err != nil {
		return nil, err
	}

	if filter.StartID == -1 {
		filter.StartID = lastID + 1
	}

	query := `
        SELECT 
    		comments.id,
    		comments.post_id,
    		comments.user_id,
    		comments.content,
    		comments.created_at,
    		users.username
		FROM comments
		JOIN users ON comments.user_id = users.id
		WHERE comments.post_id = ? AND comments.id < ?
		ORDER BY comments.id DESC
		LIMIT ?`

	rows, err := cm.DB.Query(query, filter.PostID, filter.StartID, filter.NComment)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.Username); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	
	return comments, nil
}
