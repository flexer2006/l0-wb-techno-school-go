package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	// if err := loadEnvironment(); err != nil {
	// 	return nil, fmt.Errorf("failed to load environment: %w", err)
	// }

	viperInstance := viper.New()

	configureViper(viperInstance)

	if err := loadConfigFile(viperInstance); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
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

// func loadEnvironment() error {
//
// 	const envFile = "deploy/.env"
// 	if err := godotenv.Load(envFile); err != nil {
// 		switch {
// 		case os.IsNotExist(err):
// 			if _, fprintfErr := fmt.Fprintf(os.Stderr, "info: %s not found, continuing: %v\n", envFile, err); fprintfErr != nil {
// 				return fmt.Errorf("failed to write to stderr: %w", fprintfErr)
// 			}
// 		case os.IsPermission(err):
// 			if _, fprintfErr := fmt.Fprintf(os.Stderr, "warning: cannot read %s (permission denied), continuing: %v\n", envFile, err); fprintfErr != nil {
// 				return fmt.Errorf("failed to write to stderr: %w", fprintfErr)
// 			}
// 		default:
// 			return fmt.Errorf("failed to load .env file: %w", err)
// 		}
// 	}

// 	return nil
// }

func configureViper(vpr *viper.Viper) {
	setDefaults(vpr)

	vpr.AutomaticEnv()
	vpr.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	bindEnvVariables(vpr)

	vpr.SetConfigName("config")
	vpr.SetConfigType("yaml")
	vpr.AddConfigPath("./configs")
	vpr.AddConfigPath(".")
}

func loadConfigFile(vpr *viper.Viper) error {
	if err := vpr.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}
	return nil
}

func bindEnvVariables(vpr *viper.Viper) {
	envBindings := make(map[string]string, 11)
	envBindings["database.host"] = "POSTGRES_HOST"
	envBindings["database.port"] = "POSTGRES_PORT"
	envBindings["database.user"] = "POSTGRES_USER"
	envBindings["database.password"] = "POSTGRES_PASSWORD"
	envBindings["database.database"] = "POSTGRES_DB"
	envBindings["database.sslmode"] = "POSTGRES_SSLMODE"
	envBindings["database.max_open_conns"] = "POSTGRES_MAX_CONN"
	envBindings["database.max_idle_conns"] = "POSTGRES_MIN_CONN"
	envBindings["server.host"] = "SERVER_HOST"
	envBindings["server.port"] = "SERVER_PORT"
	envBindings["kafka.brokers"] = "KAFKA_BROKERS"

	for configKey, envKey := range envBindings {
		_ = vpr.BindEnv(configKey, envKey)
	}
}

func setDefaults(vpr *viper.Viper) {
	setDatabaseDefaults(vpr)
	setServerDefaults(vpr)
	setKafkaDefaults(vpr)
	setLoggerDefaults(vpr)
	setShutdownDefaults(vpr)
}

func setDatabaseDefaults(vpr *viper.Viper) {
	defaults := map[string]interface{}{
		"database.driver":             "postgres",
		"database.host":               "localhost",
		"database.port":               5432,
		"database.user":               "postgres",
		"database.password":           "postgres",
		"database.database":           "postgres",
		"database.sslmode":            "disable",
		"database.migrations_path":    "./migrations",
		"database.max_open_conns":     25,
		"database.max_idle_conns":     5,
		"database.conn_max_lifetime":  "1h",
		"database.conn_max_idle_time": "25m",
		"database.timeout":            "5s",
	}

	for key, value := range defaults {
		vpr.SetDefault(key, value)
	}
}

func setServerDefaults(vpr *viper.Viper) {
	defaults := map[string]interface{}{
		"server.host":             "0.0.0.0",
		"server.port":             8080,
		"server.timeout":          "5s",
		"server.idle_timeout":     "60s",
		"server.read_timeout":     "5s",
		"server.shutdown_timeout": "10s",
	}

	for key, value := range defaults {
		vpr.SetDefault(key, value)
	}
}

func setKafkaDefaults(vpr *viper.Viper) {
	defaults := map[string]interface{}{
		"kafka.brokers":            []string{"localhost:9092"},
		"kafka.topic":              "orders",
		"kafka.group_id":           "order-service-group",
		"kafka.auto_offset_reset":  "earliest",
		"kafka.session_timeout":    "30s",
		"kafka.max_wait":           "10s",
		"kafka.min_bytes":          10240,
		"kafka.max_bytes":          10485760,
		"kafka.max_retries":        3,
		"kafka.enable_auto_commit": false,
		"kafka.commit_interval":    "1s",
	}

	for key, value := range defaults {
		vpr.SetDefault(key, value)
	}
}

func setLoggerDefaults(vpr *viper.Viper) {
	defaults := map[string]interface{}{
		"logger.level":                 "info",
		"logger.development":           false,
		"logger.encoding":              "json",
		"logger.output_paths":          []string{"stdout"},
		"logger.error_output_paths":    []string{"stderr"},
		"logger.encoder.time_key":      "ts",
		"logger.encoder.level_key":     "level",
		"logger.encoder.message_key":   "msg",
		"logger.encoder.caller_key":    "caller",
		"logger.encoder.time_encoder":  "iso8601",
		"logger.encoder.level_encoder": "lower",
	}

	for key, value := range defaults {
		vpr.SetDefault(key, value)
	}
}

func setShutdownDefaults(vpr *viper.Viper) {
	vpr.SetDefault("shutdown.timeout", "30s")
}
