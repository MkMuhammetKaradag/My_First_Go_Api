package messaging

import (
	"fmt"
	"time"
)

type Config struct {
	// RabbitMQ bağlantı ayarları
	Host     string
	Port     string
	User     string
	Password string
	VHost    string `default:"/"`

	// Exchange ayarları
	ExchangeName      string `default:"microservices.broadcast"`
	RetryExchangeName string `default:"microservices.retry"`

	// Retry ayarları
	MaxRetries  int           `default:"3"`
	RetryDelay  time.Duration `default:"5s"`
	RetryTypes  []string
	EnableRetry bool `default:"true"`

	// Connection ayarları
	ConnectionTimeout time.Duration `default:"30s"`
	ReadTimeout       time.Duration `default:"30s"`
	WriteTimeout      time.Duration `default:"30s"`

	// Queue ayarları
	QueueDurable    bool `default:"true"`
	QueueAutoDelete bool `default:"false"`
}

// NewDefaultConfig creates a new Config with default values
func NewDefaultConfig() Config {
	return Config{
		Host:              "localhost",
		Port:              "5672",
		User:              "user",
		Password:          "password",
		VHost:             "/",
		ExchangeName:      "microservices.broadcast",
		RetryExchangeName: "microservices.retry",
		MaxRetries:        3,
		RetryDelay:        5 * time.Second,
		EnableRetry:       true,
		ConnectionTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		QueueDurable:      true,
		QueueAutoDelete:   false,
	}
}

func (c Config) GetAMQPURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.User, c.Password, c.Host, c.Port)
}
