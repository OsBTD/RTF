package models

import (
	"database/sql"
	"errors"
	"time"
)

type Message struct {
	ID             int            `json:"id"`
	AuthorID       int            `json:"author_id"`
	ConversationID sql.NullInt64  `json:"conversation_id"`
	RecieverID     int            `json:"reciever_id"` // you had a typo here ("Reciever" instead of "Receiver") - you may want to fix that too
	Content        string         `json:"content"`
	SentAt         time.Time      `json:"sent_at"` // use string for datetime
	SeenAt         sql.NullString `json:"seen_at"` // use string for datetime
	IsOutgoing     bool           `json:"is_outgoing"`
	Type           string         `json:"type"`
	TempID         int64          `json:"temp_id,omitempty"`
}

type MessageModel struct {
	DB *sql.DB
}

// Insert Message
func (m *MessageModel) InsertMessage(msg Message) error {
	query := `
        INSERT OR IGNORE INTO messages (author_id, conversation_id, content, seen_at)
        VALUES (?, ?, ?, ?)`
	_, err := m.DB.Exec(query, msg.AuthorID, msg.ConversationID, msg.Content, msg.SeenAt)
	return err
}

// Update Message
func (m *MessageModel) UpdateMessage(messageID int, newContent string, seenAt string) error {
	query := `
        UPDATE messages
        SET content = ?, seen_at = ?
        WHERE id = ?`
	res, err := m.DB.Exec(query, newContent, seenAt, messageID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no message found with the given ID")
	}
	return nil
}

// Delete Message
func (m *MessageModel) DeleteMessage(messageID int) error {
	query := `
		DELETE FROM messages
		WHERE id = ?`
	res, err := m.DB.Exec(query, messageID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no message found with the given ID")
	}
	return nil
}

// GetMessageByID retrieves a message by its ID
func (m *MessageModel) GetMessageByID(messageID int) (Message, error) {
	query := `
        SELECT id, author_id, conversation_id, content, sent_at, seen_at
        FROM messages
        WHERE id = ?`

	row := m.DB.QueryRow(query, messageID)
	var msg Message
	err := row.Scan(
		&msg.ID, &msg.AuthorID, &msg.ConversationID,
		&msg.Content, &msg.SentAt, &msg.SeenAt,
	)
	if err != nil {
		return Message{}, err
	}
	return msg, nil
}

func (m *MessageModel) GetLastMsgID(conversationID int) (int, error) {
	var lastID int
	query := `
		SELECT id FROM messages
		WHERE conversation_id = ?
		ORDER BY id DESC
		LIMIT 1;
	`
	err := m.DB.QueryRow(query, conversationID).Scan(&lastID)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, nil // No message yet
		}
		return -1, err
	}
	return lastID, nil
}

type MessagesFilter struct {
	ConversationID sql.NullInt64 `json:"conversation_id"`
	StartID        int           `json:"start_id"`
	NMsg           int           `json:"n_message"`
}

// GetMessagesByConversation
func (m *MessageModel) GetMessages(UserID int, filter *MessagesFilter) ([]Message, error) {
	lastID, err := m.GetLastMsgID(int(filter.ConversationID.Int64))
	if err != nil {
		return []Message{}, err
	}

	if filter.StartID == -1 {
		filter.StartID = lastID + 1
	}
	query := `
        SELECT id, author_id, conversation_id, content, sent_at, seen_at
        FROM messages
        WHERE conversation_id = ? AND id < ?
        ORDER BY id DESC
        LIMIT ?`

	rows, err := m.DB.Query(query, filter.ConversationID, filter.StartID, filter.NMsg)
	if err != nil {
		return []Message{}, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.AuthorID,
			&msg.ConversationID,
			&msg.Content,
			&msg.SentAt,
			&msg.SeenAt,
		)
		if err != nil {
			return []Message{}, err
		}
		msg.IsOutgoing = (msg.AuthorID == UserID)
		messages = append(messages, msg)
	}

	return messages, nil
}
