package models

import (
	"database/sql"
	"errors"
	"time"
)

type Conversation struct {
	ID            int            `json:"id"`
	User1ID       int            `json:"user1_id"`
	User2ID       int            `json:"user2_id"`
	LastMessageAt sql.NullString `json:"last_message_at"` // ISO8601 datetime string or empty
	CreatedAt     time.Time      `json:"created_at"`      // ISO8601 datetime string
}

type ConversationModel struct {
	DB *sql.DB
}

// InsertConversation inserts a new conversation or ignores if exists (based on UNIQUE constraint)
func (cm *ConversationModel) InsertConversation(user1ID, user2ID int) (int, error) {
	query := `
		INSERT OR IGNORE INTO conversations (user1_id, user2_id, last_message_at, created_at)
		VALUES (?, ?, datetime('now'), datetime('now'))
	`
	res, err := cm.DB.Exec(query, user1ID, user2ID)
	if err != nil {
		return 0, err
	}

	convID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// If no row inserted (convID == 0), try to fetch the existing conversation ID
	if convID == 0 {
		conv, err := cm.GetConversation(user1ID, user2ID)
		if err != nil {
			return 0, err
		}
		if conv == nil {
			return 0, errors.New("failed to insert or find conversation")
		}
		return conv.ID, nil
	}

	return int(convID), nil
}

// UpdateConversation updates last_message_at by string datetime (ISO8601)
func (cm *ConversationModel) UpdateConversation(conversationID int, lastMessageAt string) error {
	query := `
		UPDATE conversations
		SET last_message_at = ?
		WHERE id = ?
	`
	res, err := cm.DB.Exec(query, lastMessageAt, conversationID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no conversation found with the given ID")
	}
	return nil
}

// DeleteConversation deletes a conversation by ID
func (cm *ConversationModel) DeleteConversation(conversationID int) error {
	query := `
		DELETE FROM conversations
		WHERE id = ?
	`
	res, err := cm.DB.Exec(query, conversationID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no conversation found with the given ID")
	}
	return nil
}

// GetConversationByID fetches conversation by ID
func (cm *ConversationModel) GetConversationByID(conversationID int64) (*Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, last_message_at, created_at
		FROM conversations
		WHERE id = ?
	`
	row := cm.DB.QueryRow(query, conversationID)
	var conv Conversation
	err := row.Scan(&conv.ID, &conv.User1ID, &conv.User2ID, &conv.LastMessageAt, &conv.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

// GetConversation gets conversation between two users (in any order)
func (cm *ConversationModel) GetConversation(user1ID, user2ID int) (*Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, last_message_at, created_at
		FROM conversations
		WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)
	`
	row := cm.DB.QueryRow(query, user1ID, user2ID, user2ID, user1ID)
	var conv Conversation
	err := row.Scan(&conv.ID, &conv.User1ID, &conv.User2ID, &conv.LastMessageAt, &conv.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

// UpdateLastMessageAt sets last_message_at to current datetime
func (cm *ConversationModel) UpdateLastMessageAt(conversationID int64) error {
	query := `
		UPDATE conversations
		SET last_message_at = datetime('now')
		WHERE id = ?
	`
	_, err := cm.DB.Exec(query, conversationID)
	return err
}

// GetUserConversationsByRecentOrder fetches conversations for user ordered by last_message_at desc
func (ccm *ConversationModel) GetUserConversationsByRecentOrder(userID int) ([]Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, last_message_at, created_at
		FROM conversations
		WHERE user1_id = ? OR user2_id = ?
		ORDER BY last_message_at DESC NULLS LAST, created_at DESC
	`
	rows, err := ccm.DB.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.User1ID, &conv.User2ID, &conv.LastMessageAt, &conv.CreatedAt)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}
	return conversations, nil
}

