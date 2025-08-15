package config

import "time"

type Config struct {
	HTTPPort    string        `envconfig:"HTTP_PORT" default:"8080"`
	HTTPTimeout time.Duration `envconfig:"HTTP_TIMEOUT" default:"30s"`
	LogLevel    string        `envconfig:"LOG_LEVEL" default:"info"`

	Postgres PostgresConfig `envconfig:"POSTGRES"`
	Kafka    KafkaConfig    `envconfig:"KAFKA"`
}

type PostgresConfig struct {
	Host              string        `envconfig:"HOST" default:"localhost"`
	Port              string        `envconfig:"PORT" default:"5432"`
	User              string        `envconfig:"USER" default:"postgres"`
	Password          string        `envconfig:"PASSWORD" default:"postgres"`
	DBName            string        `envconfig:"DB_NAME" default:"postgres"`
	SSLMode           string        `envconfig:"SSL_MODE" default:"disable"`
	// MinConns          int           `envconfig:"MIN_CONNS" default:"1"`
	// MaxConns          int           `envconfig:"MAX_CONNS" default:"10"`
	// MaxConnTTL        time.Duration `envconfig:"MAX_CONN_TTL" default:"1h"`
	// HealthCheckPeriod time.Duration `envconfig:"HEALTH_CHECK_PERIOD" default:"30s"`
	PingTimeout       time.Duration `envconfig:"PING_TIMEOUT" default:"5s"`
}

type KafkaConfig struct {
	Brokers []string `envconfig:"BROKERS" default:"localhost:9092"`
	Topic   string   `envconfig:"TOPIC" default:"orders"`
	GroupID string `envconfig:"GROUP_ID" default:"order-processor"`
	MaxRetries int `envconfig:"MAX_RETRIES" default:"3"`
}
