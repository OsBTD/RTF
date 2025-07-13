package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int            `json:"id"`
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	UserName       string         `json:"username"`
	Email          string         `json:"email"`
	Birthday       string         `json:"birth_date"` // ISO8601 date string
	Gender         string         `json:"gender"`
	Password       string         `json:"password"` // plain text (only for input)
	RepeatedPass   string         `json:"repeated_password"`
	HashedPassword []byte         // stored hashed password
	Token          string         `json:"token"`
	ProfileImg     string         `json:"profile_img"`
	ConversationID sql.NullInt64  `json:"conversation_id"`
	CreatedAt      time.Time      `json:"created_at"`      // ISO8601 datetime string
	LastMessageAt  sql.NullString `json:"last_message_at"` // ISO8601 datetime string or empty
}

type UserModel struct {
	DB *sql.DB
}

// InsertUser inserts a new user, setting created_at via SQLite default
func (um *UserModel) InsertUser(user User) error {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return err
	}
	user.HashedPassword = hashedPwd

	// Assign avatar before insertion
	user.ProfileImg, err = AssignAvatar(user.UserName, user.Gender)
	if err != nil {
		return err
	}

	query := `
		INSERT OR IGNORE INTO users 
			(username, first_name, last_name, email, birth_date, gender, hashed_password, profile_img)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = um.DB.Exec(query,
		user.UserName,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Birthday,
		user.Gender,
		user.HashedPassword,
		user.ProfileImg,
	)
	if err != nil {
		return err
	}

	return nil
}

func (um *UserModel) GetUserByID(userID int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, first_name, last_name, email, birth_date, gender, profile_img, created_at FROM users WHERE id = ?`
	err := um.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.UserName,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Birthday,
		&user.Gender,
		&user.ProfileImg,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (um *UserModel) ValidateUser(user *User, state string) error {
	user.UserName = strings.ToLower(strings.TrimSpace(user.UserName))
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	switch state {
	case "newUser":
		if err := firstNameCheck(user.FirstName); err != nil {
			return err
		}
		if err := um.lastNameCheck(user.LastName); err != nil {
			return err
		}
		if err := um.usernameCheck(user.UserName); err != nil {
			return err
		}
		if err := um.emailCheck(user.Email); err != nil {
			return err
		}
		if err := ageCheck(user.Birthday); err != nil {
			return err
		}
		if err := genderCheck(user.Gender); err != nil {
			return err
		}
		if err := passwordCheck(user.Password, user.RepeatedPass); err != nil {
			return err
		}
		return nil

	case "User":
		if user.UserName == "" && user.Email == "" {
			return fmt.Errorf("nickname or Email is required")
		}

		var identifier string
		var query string
		if user.Email != "" {
			query = `SELECT id, hashed_password FROM users WHERE email = ?`
			identifier = user.Email
		} else {
			query = `SELECT id, hashed_password FROM users WHERE username = ?`
			identifier = user.UserName
		}

		err := um.DB.QueryRow(query, identifier).Scan(&user.ID, &user.HashedPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("user not found")
			}
			return fmt.Errorf("database error: %v", err)
		}

		user.Password = strings.TrimSpace(user.Password)
		if user.Password == "" {
			return fmt.Errorf("password is required")
		}

		if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(user.Password)); err != nil {
			return fmt.Errorf("invalid password")
		}
		return nil

	default:
		return fmt.Errorf("invalid validation state")
	}
}

func (um *UserModel) GetSortedUsersByConversation(userID int) ([]User, error) {
	query := `
		SELECT
			users.id,
			users.first_name,
			users.last_name,
			users.username,
			users.email,
			users.birth_date,
			users.gender,
			users.profile_img,
			users.created_at,
			conversations.id as convid,
			CASE
				WHEN conversations.user2_id = ? THEN conversations.last_message_at
				ELSE NULL
			END AS last_message_at
		FROM users
		LEFT JOIN conversations ON (
			(conversations.user1_id = ? AND conversations.user2_id = users.id)
			OR
			(conversations.user2_id = ? AND conversations.user1_id = users.id)
		)
		WHERE users.id != ?
		ORDER BY
			CASE WHEN conversations.id IS NOT NULL THEN 0 ELSE 1 END,
			last_message_at DESC,
			users.first_name ASC;
	`

	rows, err := um.DB.Query(query, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.UserName,
			&user.Email,
			&user.Birthday,
			&user.Gender,
			&user.ProfileImg,
			&user.CreatedAt,
			&user.ConversationID,
			&user.LastMessageAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// usernameCheck ensures username validity and uniqueness
func (um *UserModel) usernameCheck(username string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" {
		return errors.New("username is required")
	}
	if len(username) < 3 || len(username) > 20 {
		return errors.New("username must be between 3 and 20 characters")
	}
	if username[0] == '_' || username[len(username)-1] == '_' {
		return errors.New("username cannot start or end with '_'")
	}
	for _, c := range username {
		if !(c >= 'a' && c <= 'z') && !(c >= '0' && c <= '9') && c != '_' {
			return errors.New("username can only contain letters, numbers and underscores")
		}
	}
	var existing string
	err := um.DB.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existing)
	if err == nil {
		return errors.New("'" + username + "' is already taken")
	}
	if err != sql.ErrNoRows {
		return err
	}
	return nil
}

// emailCheck ensures email validity and uniqueness
func (um *UserModel) emailCheck(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return errors.New("email is required")
	}
	at := strings.Index(email, "@")
	if at == -1 {
		return errors.New("email must contain '@'")
	}
	if strings.Count(email, "@") > 1 {
		return errors.New("email must contain only one '@'")
	}
	localPart := email[:at]
	domainPart := email[at+1:]
	if len(localPart) == 0 {
		return errors.New("email must have a local part before '@'")
	}
	if len(domainPart) == 0 {
		return errors.New("email must have a domain part after '@'")
	}
	if !strings.Contains(domainPart, ".") {
		return errors.New("email domain must contain '.'")
	}
	if domainPart[0] == '.' || domainPart[len(domainPart)-1] == '.' {
		return errors.New("email domain cannot start or end with '.'")
	}
	var existing string
	err := um.DB.QueryRow("SELECT email FROM users WHERE email = ?", email).Scan(&existing)
	if err == nil {
		return errors.New("this email '" + email + "' is already registered")
	}
	if err != sql.ErrNoRows {
		return err
	}
	return nil
}

func ageCheck(birthDate string) error {
	// Optional: implement actual age check parsing birthDate string here
	return nil
}

func genderCheck(gender string) error {
	if gender != "male" && gender != "female" {
		return errors.New("gender must be 'male' or 'female'")
	}
	return nil
}

func passwordCheck(password, repeatedPassword string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if password != repeatedPassword {
		return errors.New("password and repeated password must match")
	}
	if len(password) < 8 || len(password) > 64 {
		return errors.New("password length must be between 8 and 64 characters")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	specialChars := "!@#$%^&*()-_=+[]{}|;:',.<>?/"
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case strings.ContainsRune(specialChars, c):
			hasSpecial = true
		case c == ' ':
			return errors.New("password cannot contain spaces")
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	return nil
}

func firstNameCheck(firstName string) error {
	firstName = strings.TrimSpace(firstName)
	if len(firstName) < 2 || len(firstName) > 50 {
		return errors.New("first name must be between 2 and 50 characters")
	}
	for _, c := range firstName {
		if !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && c != ' ' && c != '-' {
			return errors.New("first name can only contain letters, spaces, and hyphens")
		}
	}
	return nil
}

func (um *UserModel) lastNameCheck(lastName string) error {
	lastName = strings.TrimSpace(lastName)
	if len(lastName) < 2 || len(lastName) > 50 {
		return errors.New("last name must be between 2 and 50 characters")
	}
	for _, c := range lastName {
		if !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && c != ' ' && c != '-' {
			return errors.New("last name can only contain letters, spaces, and hyphens")
		}
	}
	return nil
}
