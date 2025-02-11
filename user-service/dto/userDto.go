package dto

import "time"

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName,omitempty"`
	Age       int       `json:"age,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
