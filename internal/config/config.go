package config

import "time"

type Config struct {
	Database DatabaseConfig `yaml:"database" mapstructure:"database"`
	Server   ServerConfig   `yaml:"server" mapstructure:"server"`
	Kafka    KafkaConfig    `yaml:"kafka" mapstructure:"kafka"`
	Logger   LoggerConfig   `yaml:"logger" mapstructure:"logger"`
	Shutdown ShutdownConfig `yaml:"shutdown" mapstructure:"shutdown"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host" mapstructure:"host"`
	User            string        `yaml:"user" mapstructure:"user"`
	Password        string        `yaml:"password" mapstructure:"password"`
	Database        string        `yaml:"database" mapstructure:"database"`
	Driver          string        `yaml:"driver" mapstructure:"driver"`
	SSLMode         string        `yaml:"sslmode" mapstructure:"sslmode"`
	MigrationsPath  string        `yaml:"migrations_path" mapstructure:"migrations_path"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
	Timeout         time.Duration `yaml:"timeout" mapstructure:"timeout"`
	Port            int           `yaml:"port" mapstructure:"port"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
}

type ServerConfig struct {
	Host            string        `yaml:"host" mapstructure:"host"`
	Timeout         time.Duration `yaml:"timeout" mapstructure:"timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" mapstructure:"shutdown_timeout"`
	Port            int           `yaml:"port" mapstructure:"port"`
}

type KafkaConfig struct {
	Brokers          []string      `yaml:"brokers" mapstructure:"brokers"`
	Topic            string        `yaml:"topic" mapstructure:"topic"`
	GroupID          string        `yaml:"group_id" mapstructure:"group_id"`
	AutoOffsetReset  string        `yaml:"auto_offset_reset" mapstructure:"auto_offset_reset"`
	SessionTimeout   time.Duration `yaml:"session_timeout" mapstructure:"session_timeout"`
	MaxWait          time.Duration `yaml:"max_wait" mapstructure:"max_wait"`
	CommitInterval   time.Duration `yaml:"commit_interval" mapstructure:"commit_interval"`
	MinBytes         int           `yaml:"min_bytes" mapstructure:"min_bytes"`
	MaxBytes         int           `yaml:"max_bytes" mapstructure:"max_bytes"`
	MaxRetries       int           `yaml:"max_retries" mapstructure:"max_retries"`
	EnableAutoCommit bool          `yaml:"enable_auto_commit" mapstructure:"enable_auto_commit"`
}

type LoggerConfig struct {
	OutputPaths []string `yaml:"output_paths" mapstructure:"output_paths"`
	ErrorPaths  []string `yaml:"error_output_paths" mapstructure:"error_output_paths"`
	Level       string   `yaml:"level" mapstructure:"level"`
	Encoding    string   `yaml:"encoding" mapstructure:"encoding"`
	Encoder     Encoder  `yaml:"encoder" mapstructure:"encoder"`
	Development bool     `yaml:"development" mapstructure:"development"`
}

type Encoder struct {
	TimeKey      string `yaml:"time_key" mapstructure:"time_key"`
	LevelKey     string `yaml:"level_key" mapstructure:"level_key"`
	MessageKey   string `yaml:"message_key" mapstructure:"message_key"`
	CallerKey    string `yaml:"caller_key" mapstructure:"caller_key"`
	TimeEncoder  string `yaml:"time_encoder" mapstructure:"time_encoder"`
	LevelEncoder string `yaml:"level_encoder" mapstructure:"level_encoder"`
}

type ShutdownConfig struct {
	Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"`
}
