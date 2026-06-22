package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/yuusufyan/go-common/pkg/logger"
)

type Config struct {
	IsProd            bool
	AppHost           string
	AppPort           int
	DBHost            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBPort            int
	DBMaxIdleConns    int
	DBMaxOpenConns    int
	DBConnMaxLifetime int

	RedisHost string
	RedisPort int
	RedisPass string
	RedisDB   int

	RabbitMQURL string
}

func Load() (*Config, error) {
	configPaths := []string{
		"./",     // For app
		"../../", // For test folder
	}

	var configFound bool
	for _, path := range configPaths {
		viper.SetConfigFile(path + ".env")
		if err := viper.ReadInConfig(); err == nil {
			configFound = true
			break
		}
	}

	if !configFound {
		log := logger.New(false)
		log.Warn("failed to load any config file (.env), will rely on environment variables")
	}

	cfg := &Config{
		// server configuration
		IsProd:  viper.GetString("APP_ENV") == "prod",
		AppHost: viper.GetString("APP_HOST"),
		AppPort: viper.GetInt("APP_PORT"),

		// database configuration
		DBHost: viper.GetString("DB_HOST"),

		// redis configuration
		RedisHost: viper.GetString("REDIS_HOST"),
		RedisPort: viper.GetInt("REDIS_PORT"),
		RedisPass: viper.GetString("REDIS_PASS"),
		RedisDB:   viper.GetInt("REDIS_DB"),

		// DB Configuration
		DBUser:            viper.GetString("DB_USERNAME"),
		DBPassword:        viper.GetString("DB_PASSWORD"),
		DBName:            viper.GetString("DB_NAME"),
		DBPort:            viper.GetInt("DB_PORT"),
		DBMaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
		DBMaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
		DBConnMaxLifetime: viper.GetInt("DB_CONN_MAX_LIFETIME"),

		// rabbitmq configuration
		RabbitMQURL: viper.GetString("RABBITMQ_URL"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	type rule struct {
		env   string
		value interface{}
	}

	required := []rule{
		{"DB_HOST", c.DBHost},
		{"DB_USER", c.DBUser},
		{"DB_NAME", c.DBName},
	}

	var missing []string
	for _, r := range required {
		if s, ok := r.value.(string); ok && strings.TrimSpace(s) == "" {
			missing = append(missing, r.env)
		}
	}

	if len(missing) > 0 {
		err := fmt.Errorf("[CONFIG] validation failed. missing variables: %s", strings.Join(missing, ", "))
		log := logger.New(false)
		log.Error(err.Error())
		return err
	}

	return nil
}
