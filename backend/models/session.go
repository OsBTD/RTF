package models

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// insert session
// update session
// delete session
// select session
// select all user session

type Session struct {
	Token     string
	UserID    int
	ExpiresAt time.Time
}

type SessionModel struct {
	DB *sql.DB
}

func (sm *SessionModel) GenerateNewSession(userID int) (Session, error) {
	exp := 24 * time.Hour
	newToken, err := uuid.NewRandom()
	if err != nil {
		return Session{}, fmt.Errorf("can't generate session")
	}

	return Session{
		UserID:    userID,
		Token:     newToken.String(),
		ExpiresAt: time.Now().Add(exp),
	}, nil
}

func (sm *SessionModel) GetUserBySession(Token string) (session Session, errCode int, err error) {
	query := `SELECT token, user_id, expires_at FROM sessions WHERE token = ?`
	err = sm.DB.QueryRow(query, Token).Scan(&session.Token, &session.UserID, &session.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Session{}, http.StatusUnauthorized, fmt.Errorf("session not found: %v", err)
		}
		return Session{}, http.StatusInternalServerError, err
	}

	if time.Now().After(session.ExpiresAt) {
		return Session{}, http.StatusUnauthorized, fmt.Errorf("session expired")
	}

	return session, http.StatusOK, nil
}

func (sm *SessionModel) UpSertSession(newSession Session) (http.Cookie, error) {
	query := `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id)
		DO UPDATE SET
			token = excluded.token,
			expires_at = excluded.expires_at;
	`

	_, err := sm.DB.Exec(query, newSession.UserID, newSession.Token, newSession.ExpiresAt)
	if err != nil {
		return http.Cookie{}, fmt.Errorf("session not found: %v", err)
	}

	cookie := http.Cookie{
		Name:     "session_id",
		Value:    newSession.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  newSession.ExpiresAt,
	}

	return cookie, nil
}

func (sm *SessionModel) DeleteSession(UserID int) error {
	_, err := sm.DB.Exec(`DELETE FROM sessions WHERE user_id = ?`, UserID)
	return err
}
