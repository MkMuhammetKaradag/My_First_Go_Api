package dto

import (
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/shared/models"
)

// type UserRole string

// const (
// 	ADMIN UserRole = "admin"
// 	TEST  UserRole = "test"
// 	USER  UserRole = "user"
// )

type UserResponse struct {
	ID        string            `json:"id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	Roles     []models.UserRole `json:"roles"  `
	FirstName string            `json:"firstName,omitempty"`
	Age       int               `json:"age,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
}
