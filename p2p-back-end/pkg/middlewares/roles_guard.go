package middlewares

import (
	"github.com/gofiber/fiber/v2"

	"p2p-back-end/modules/entities/models" // import models ให้ถูกต้อง
)

// RolesGuard จะรับ "Controller ปลายทาง (next)" เข้ามาด้วย
// และรับ "allowedRoles" เป็นตัวสุดท้าย
func RolesGuard(next models.TokenHandler, allowedRoles ...string) models.TokenHandler {

	return func(c *fiber.Ctx, user *models.UserInfo) error {
		// 1. เช็ค Role จาก user struct ที่ JwtAuthentication ส่งมาให้โดยตรง
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User info missing"})
		}

		userRoles := user.Role // สมมติว่าใน UserInfo field ชื่อ Role เป็น []string

		// 2. แปลง User Role ให้เป็น Map เพื่อเช็ค
		userRolesMap := make(map[string]bool)
		for _, r := range userRoles {
			userRolesMap[r] = true
		}

		// 3. ตรวจสอบสิทธิ
		isAllowed := false
		for _, allowed := range allowedRoles {
			if userRolesMap[allowed] {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access Denied: คุณไม่มีสิทธิ์ใช้งานส่วนนี้",
			})
		}

		// 4. ถ้าผ่าน ให้เรียก Controller ตัวจริง (next) ทำงานต่อ
		return next(c, user)
	}
}
