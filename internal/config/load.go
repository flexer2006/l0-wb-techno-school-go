package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	if err := godotenv.Load("deploy/.env"); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	viperInstance := viper.New()

	setDefaults(viperInstance)

	viperInstance.AutomaticEnv()
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := bindEnvVariables(viperInstance); err != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %w", err)
	}

	viperInstance.SetConfigName("config")
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath("./configs")
	viperInstance.AddConfigPath(".")

	if err := viperInstance.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viperInstance.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

func bindEnvVariables(viperInstance *viper.Viper) error {
	envBindings := map[string]string{
		"database.host":           "POSTGRES_HOST",
		"database.port":           "POSTGRES_PORT",
		"database.user":           "POSTGRES_USER",
		"database.password":       "POSTGRES_PASSWORD",
		"database.database":       "POSTGRES_DB",
		"database.sslmode":        "POSTGRES_SSLMODE",
		"database.max_open_conns": "POSTGRES_MAX_CONN",
		"database.max_idle_conns": "POSTGRES_MIN_CONN",
		"server.host":             "SERVER_HOST",
		"server.port":             "SERVER_PORT",
		"kafka.brokers":           "KAFKA_BROKERS",
	}

	for configKey, envKey := range envBindings {
		if err := viperInstance.BindEnv(configKey, envKey); err != nil {
			return fmt.Errorf("failed to bind env variable %s: %w", envKey, err)
		}
	}

	return nil
}

func setDefaults(viperInstance *viper.Viper) {
	viperInstance.SetDefault("database.driver", "postgres")
	viperInstance.SetDefault("database.host", "localhost")
	viperInstance.SetDefault("database.port", 5432)
	viperInstance.SetDefault("database.user", "postgres")
	viperInstance.SetDefault("database.password", "postgres")
	viperInstance.SetDefault("database.database", "postgres")
	viperInstance.SetDefault("database.sslmode", "disable")
	viperInstance.SetDefault("database.migrations_path", "./migrations")
	viperInstance.SetDefault("database.max_open_conns", 25)
	viperInstance.SetDefault("database.max_idle_conns", 5)
	viperInstance.SetDefault("database.conn_max_lifetime", "1h")
	viperInstance.SetDefault("database.conn_max_idle_time", "25m")
	viperInstance.SetDefault("database.timeout", "5s")

	viperInstance.SetDefault("server.host", "0.0.0.0")
	viperInstance.SetDefault("server.port", 8080)
	viperInstance.SetDefault("server.timeout", "5s")
	viperInstance.SetDefault("server.idle_timeout", "60s")
	viperInstance.SetDefault("server.read_timeout", "5s")
	viperInstance.SetDefault("server.shutdown_timeout", "10s")

	viperInstance.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viperInstance.SetDefault("kafka.topic", "orders")
	viperInstance.SetDefault("kafka.group_id", "order-service-group")
	viperInstance.SetDefault("kafka.auto_offset_reset", "earliest")
	viperInstance.SetDefault("kafka.session_timeout", "30s")
	viperInstance.SetDefault("kafka.max_wait", "10s")
	viperInstance.SetDefault("kafka.min_bytes", 10240)
	viperInstance.SetDefault("kafka.max_bytes", 10485760)
	viperInstance.SetDefault("kafka.max_retries", 3)
	viperInstance.SetDefault("kafka.enable_auto_commit", false)
	viperInstance.SetDefault("kafka.commit_interval", "1s")

	viperInstance.SetDefault("logger.level", "info")
	viperInstance.SetDefault("logger.development", false)
	viperInstance.SetDefault("logger.encoding", "json")
	viperInstance.SetDefault("logger.output_paths", []string{"stdout"})
	viperInstance.SetDefault("logger.error_output_paths", []string{"stderr"})
	viperInstance.SetDefault("logger.encoder.time_key", "ts")
	viperInstance.SetDefault("logger.encoder.level_key", "level")
	viperInstance.SetDefault("logger.encoder.message_key", "msg")
	viperInstance.SetDefault("logger.encoder.caller_key", "caller")
	viperInstance.SetDefault("logger.encoder.time_encoder", "iso8601")
	viperInstance.SetDefault("logger.encoder.level_encoder", "lower")

	viperInstance.SetDefault("shutdown.timeout", "30s")
}
