package messaging

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	config  Config
	conn    *amqp.Connection
	channel *amqp.Channel
	service ServiceType
	mu      sync.Mutex // For thread safety

	// Connection management
	closed    bool
	reconnect chan bool
}

// NewRabbitMQ creates a new RabbitMQ instance
func NewRabbitMQ(config Config, serviceType ServiceType) (*RabbitMQ, error) {
	r := &RabbitMQ{
		config:    config,
		service:   serviceType,
		reconnect: make(chan bool),
	}

	if err := r.connect(); err != nil {
		return nil, err
	}

	// Start connection monitoring
	go r.monitorConnection()

	return r, nil
}

// connect establishes connection to RabbitMQ
func (r *RabbitMQ) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create connection
	conn, err := amqp.DialConfig(r.config.GetAMQPURL(), amqp.Config{
		Heartbeat: 10 * time.Second,
		Dial:      amqp.DefaultDial(r.config.ConnectionTimeout),
	})
	if err != nil {
		return &MessagingError{Code: "CONNECTION_FAILED", Message: "Failed to connect", Err: err}
	}

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return &MessagingError{Code: "CHANNEL_FAILED", Message: "Failed to create channel", Err: err}
	}

	// Set up exchanges
	if err := r.setupExchanges(ch); err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	r.conn = conn
	r.channel = ch
	r.closed = false

	return nil
}

// PublishMessage publishes a message to RabbitMQ
func (r *RabbitMQ) PublishMessage(ctx context.Context, msg Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	// Set message creation time
	if msg.Created.IsZero() {
		msg.Created = time.Now()
	}

	// Set source service
	msg.FromService = r.service

	// Prevent self-messaging
	if msg.ToService == r.service {
		return &MessagingError{Code: "INVALID_TARGET", Message: "Service cannot send message to itself"}
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return &MessagingError{Code: "MARSHAL_FAILED", Message: "Failed to marshal message", Err: err}
	}

	return r.channel.PublishWithContext(ctx,
		r.config.ExchangeName,
		"",    // routing key
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			MessageId:    msg.ID,
			Timestamp:    msg.Created,
			Priority:     uint8(msg.Priority),
			Headers:      amqp.Table(msg.Headers),
			DeliveryMode: 2, // persistent
		},
	)
}

// ConsumeMessages starts consuming messages
func (r *RabbitMQ) ConsumeMessages(handler MessageHandler) error {
	queueName := string(r.service) + ".queue"

	// Declare queue
	q, err := r.channel.QueueDeclare(
		queueName,
		r.config.QueueDurable,
		r.config.QueueAutoDelete,
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return &MessagingError{Code: "QUEUE_FAILED", Message: "Failed to declare queue", Err: err}
	}

	// Bind queue to exchange
	err = r.channel.QueueBind(
		q.Name,
		"", // routing key
		r.config.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return &MessagingError{Code: "BIND_FAILED", Message: "Failed to bind queue", Err: err}
	}

	msgs, err := r.channel.Consume(
		q.Name,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return &MessagingError{Code: "CONSUME_FAILED", Message: "Failed to start consuming", Err: err}
	}

	go func() {
		for msg := range msgs {
			var message Message
			if err := json.Unmarshal(msg.Body, &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				msg.Nack(false, false)
				continue
			}

			// Handle the message
			if err := handler(message); err != nil {
				// Check if retry is enabled and applicable
				if r.shouldRetry(message) {
					r.handleRetry(message)
					msg.Nack(false, false)
				} else {
					// Send to dead letter queue or log
					log.Printf("Message processing failed: %v", err)
					msg.Nack(false, false)
				}
			} else {
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// monitorConnection monitors and handles reconnection
func (r *RabbitMQ) monitorConnection() {
	for {
		if r.closed {
			return
		}

		if r.conn.IsClosed() {
			log.Println("Connection lost. Attempting to reconnect...")
			for {
				if err := r.connect(); err == nil {
					log.Println("Reconnected successfully")
					break
				}
				time.Sleep(5 * time.Second)
			}
		}

		time.Sleep(5 * time.Second)
	}
}

// Close closes the RabbitMQ connection
func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true

	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}

	return nil
}

// Private helper methods
func (r *RabbitMQ) setupExchanges(ch *amqp.Channel) error {
	// Declare main exchange
	err := ch.ExchangeDeclare(
		r.config.ExchangeName,
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	// Declare retry exchange if enabled
	if r.config.EnableRetry {
		err = ch.ExchangeDeclare(
			r.config.RetryExchangeName,
			"direct",
			true,
			false,
			false,
			false,
			nil,
		)
	}
	return err
}

func (r *RabbitMQ) shouldRetry(msg Message) bool {
	if !r.config.EnableRetry {
		return false
	}

	// Check if message type is in retry types
	for _, t := range r.config.RetryTypes {
		if t == msg.Type {
			return msg.RetryCount < r.config.MaxRetries
		}
	}
	return false
}

func (r *RabbitMQ) handleRetry(msg Message) {
	msg.RetryCount++
	// Implementation of retry logic
	// This could involve publishing to a delay queue
}
