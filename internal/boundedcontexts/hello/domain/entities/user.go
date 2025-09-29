package entities

import (
	"time"
	"errors"
)

// User represents a user entity
type User struct {
	ID        string
	Email     string
	Phone     string
	NickName  string
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBlocked  UserStatus = "blocked"
)

// NewUser creates a new User
func NewUser(email string, nickName string, phone string) (*User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	if nickName == "" {
		return nil, errors.New("nickname cannot be empty")
	}

	user := &User{
		ID:        generateID(), // Simple ID generation
		Email:     email,
		Phone:     phone,
		NickName:  nickName,
		Status:    UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return user, nil
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(nickName string, phone string) error {
	if nickName == "" {
		return errors.New("nickname cannot be empty")
	}

	u.NickName = nickName
	u.Phone = phone
	u.UpdatedAt = time.Now()

	return nil
}

// ChangeEmail changes the user's email
func (u *User) ChangeEmail(newEmail string) error {
	if u.Status == UserStatusBlocked {
		return errors.New("cannot change email for blocked user")
	}

	if newEmail == "" {
		return errors.New("email cannot be empty")
	}

	u.Email = newEmail
	u.UpdatedAt = time.Now()

	return nil
}

// Activate activates the user
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
}

// Deactivate deactivates the user
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now()
}

// Block blocks the user
func (u *User) Block() {
	u.Status = UserStatusBlocked
	u.UpdatedAt = time.Now()
}

// generateID generates a simple unique ID
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(result)
}