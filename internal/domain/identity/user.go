package identity

import (
	"errors"
	"time"
)

// User represents a user in the identity domain
type User struct {
	id        UserID
	email     Email
	password  PasswordHash
	createdAt time.Time
}

// UserID is a value object representing a user identifier
type UserID struct {
	value int
}

func NewUserID(id int) UserID {
	return UserID{value: id}
}

func (u UserID) Value() int {
	return u.value
}

// Email is a value object representing an email address
type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	if email == "" {
		return Email{}, errors.New("email cannot be empty")
	}
	// Add email validation logic here
	return Email{value: email}, nil
}

func (e Email) Value() string {
	return e.value
}

// PasswordHash is a value object representing a hashed password
type PasswordHash struct {
	value string
}

func NewPasswordHash(hash string) PasswordHash {
	return PasswordHash{value: hash}
}

func (p PasswordHash) Value() string {
	return p.value
}

// NewUser creates a new user entity
func NewUser(id UserID, email Email, password PasswordHash) *User {
	return &User{
		id:        id,
		email:     email,
		password:  password,
		createdAt: time.Now(),
	}
}

// Getters
func (u *User) ID() UserID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) PasswordHash() PasswordHash {
	return u.password
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// ChangeEmail changes the user's email
func (u *User) ChangeEmail(newEmail Email) error {
	u.email = newEmail
	return nil
}

// ChangePassword changes the user's password
func (u *User) ChangePassword(newPassword PasswordHash) error {
	u.password = newPassword
	return nil
}
