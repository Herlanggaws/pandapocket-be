package identity

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the identity domain
type User struct {
	id        UserID
	email     Email
	password  PasswordHash
	role      Role
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

func NewPasswordHash(value string) PasswordHash {
	return PasswordHash{value: value}
}

func NewPasswordHashFromPlain(plain string) (PasswordHash, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return PasswordHash{}, err
	}
	return PasswordHash{value: string(hashedBytes)}, nil
}

func (p PasswordHash) Value() string {
	return p.value
}

// Role is a value object representing a user role
type Role struct {
	value string
}

func NewRole(role string) (Role, error) {
	validRoles := map[string]bool{
		"user":        true,
		"admin":       true,
		"super_admin": true,
	}

	if !validRoles[role] {
		return Role{}, errors.New("invalid role")
	}

	return Role{value: role}, nil
}

func (r Role) Value() string {
	return r.value
}

func (r Role) IsAdmin() bool {
	return r.value == "admin" || r.value == "super_admin"
}

func (r Role) IsSuperAdmin() bool {
	return r.value == "super_admin"
}

// NewUser creates a new user entity
func NewUser(id UserID, email Email, password PasswordHash, role Role) *User {
	return &User{
		id:        id,
		email:     email,
		password:  password,
		role:      role,
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

func (u *User) Role() Role {
	return u.role
}

func (u *User) UpdatePassword(passwordHash PasswordHash) {
	u.password = passwordHash
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

// ChangeRole changes the user's role
func (u *User) ChangeRole(newRole Role) error {
	u.role = newRole
	return nil
}
