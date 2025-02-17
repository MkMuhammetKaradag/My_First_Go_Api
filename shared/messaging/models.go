package messaging

import (
	"fmt"
	"time"
)

// ServiceType defines the type of microservice
type ServiceType string

const (
	AuthService  ServiceType = "auth"
	UserService  ServiceType = "user"
	EmailService ServiceType = "email"
	ChatService  ServiceType = "chat"
)

// Message represents a message in the system
type Message struct {
	ID          string      `json:"id"`           // Unique message ID
	Type        string      `json:"type"`         // Message type (e.g., "user_created")
	Data        interface{} `json:"data"`         // Actual message payload
	Created     time.Time   `json:"created"`      // Message creation time
	FromService ServiceType `json:"from_service"` // Source service
	ToService   ServiceType `json:"to_service"`   // Target service (empty for broadcast)
	RetryCount  int         `json:"retry_count"`  // Number of retry attempts
	Priority    int         `json:"priority"`     // Message priority (0-9)
	Headers     Headers     `json:"headers"`      // Custom message headers
}

// Headers contains custom message metadata
type Headers map[string]interface{}

// MessageHandler defines the function signature for message handlers
type MessageHandler func(Message) error

// Error types for messaging
type MessagingError struct {
	Code    string
	Message string
	Err     error
}

func (e *MessagingError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
}

// Common messaging errors
var (
	ErrConnectionFailed = &MessagingError{Code: "CONNECTION_FAILED", Message: "Failed to connect to RabbitMQ"}
	ErrPublishFailed    = &MessagingError{Code: "PUBLISH_FAILED", Message: "Failed to publish message"}
	ErrConsumeFailed    = &MessagingError{Code: "CONSUME_FAILED", Message: "Failed to consume messages"}
)
