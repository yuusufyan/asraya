package main

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	bootsrap "github.com/yuusufyan/asraya/internal/boostrap"
	"github.com/yuusufyan/asraya/internal/config"
	"github.com/yuusufyan/go-common/pkg/database"
	"github.com/yuusufyan/go-common/pkg/logger"
	commonMiddleware "github.com/yuusufyan/go-common/pkg/middleware/fiber"
	"github.com/yuusufyan/go-common/pkg/rabbitmq"
	"github.com/yuusufyan/go-common/pkg/utils"
)

type customLogger struct {
	*logrus.Logger
}

func (l *customLogger) WithCtx(ctx context.Context) *logrus.Entry {
	return logger.WithCtx(ctx, l.Logger)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fallbackLog := logger.New(false)
		fallbackLog.Fatalf("Failed to load config: %v", err)
	}

	log := logrus.New()
	if cfg.IsProd {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
		})
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "15:04:05.000",
			FullTimestamp:   true,
			ForceColors:     true,
		})
		log.SetLevel(logrus.DebugLevel)
	}

	appLogger := &customLogger{Logger: log}

	dbConfig := &database.DBConfig{
		Host:     cfg.DBHost,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		Port:     cfg.DBPort,
	}
	db, err := database.Connect(dbConfig, appLogger, cfg.IsProd)
	if err != nil {
		log.Warnf("Database connection skipped or failed: %v", err)
	} else {
		log.Info("Database connection success")
		db.Use(database.NewAuditPlugin())
	}

	var mq rabbitmq.RabbitMQClient
	if cfg.RabbitMQURL != "" {
		mq, err = rabbitmq.NewRabbitMQClient(cfg.RabbitMQURL)
		if err != nil {
			log.Errorf("Failed to connect to RabbitMQ: %v", err)
		} else {
			log.Info("Connected to RabbitMQ")
		}
	} else {
		log.Info("RabbitMQ URL not provided, skipping connection")
	}

	fiberCfg := bootsrap.FiberConfig(cfg, appLogger.Logger)
	app := fiber.New(fiberCfg)

	app.Get("/health", utils.NewHealthHandler(db, nil))
	commonMiddleware.InstallCommonMiddleware(app, appLogger)

	go func() {
		port := fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort)
		log.Infof("Starting server on port %s", port)
		if err := app.Listen(port); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	sh := utils.NewShutdownHelper(log)
	sh.Wait()

	shutdownTasks := map[string]func(ctx context.Context) error{
		"Fiber": func(ctx context.Context) error {
			return app.ShutdownWithContext(ctx)
		},
	}

	if db != nil {
		shutdownTasks["Database"] = func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		}
	}

	if mq != nil {
		shutdownTasks["RabbitMQ"] = func(ctx context.Context) error {
			return mq.Close()
		}
	}

	sh.Graceful(shutdownTasks)
}
