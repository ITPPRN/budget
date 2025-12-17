package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewCorsOriginMiddleWare() func(*fiber.Ctx) error {
	return cors.New(cors.Config{
		Next:             nil,
		AllowOriginsFunc: nil,
		// AllowOrigins:     "*",
		// 1. ระบุ URL ของ Frontend (ห้ามใช้ "*")
        AllowOrigins: "http://localhost:3000, http://localhost:8121", 
        
        // 2. ต้องเปิดเป็น true เพื่อให้ browser ยอมรับ cookie
        AllowCredentials: true,

		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
		}, ","),
		AllowHeaders:     "",
		ExposeHeaders:    "",
		MaxAge:           0,
	})
}
