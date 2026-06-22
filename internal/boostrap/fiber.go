package bootsrap

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/yuusufyan/asraya/internal/config"
	"github.com/yuusufyan/go-common/pkg/utils"
)

func FiberConfig(cfg *config.Config, log *logrus.Logger) fiber.Config {
	return fiber.Config{
		Prefork:       cfg.IsProd,
		CaseSensitive: true,
		ServerHeader:  "Fiber",
		AppName:       "Fiber API",
		ErrorHandler:  utils.NewErrorHandler(log),
		JSONEncoder:   sonic.Marshal,
		JSONDecoder:   sonic.Unmarshal,
	}
}
