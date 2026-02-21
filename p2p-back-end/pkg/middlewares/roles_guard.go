package middlewares

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models" // import models ให้ถูกต้อง
)

// RolesGuard จะรับ "Controller ปลายทาง (next)" เข้ามาด้วย
// และรับ "allowedRoles" เป็นตัวสุดท้าย
func RolesGuard(next models.TokenHandler, allowedRoles ...string) models.TokenHandler {

	return func(c *fiber.Ctx, user *models.UserInfo) error {
		// 1. เช็ค Role จาก user struct ที่ JwtAuthentication ส่งมาให้โดยตรง
		// 1. เช็ค Role จาก user struct ที่JwtAuthentication ส่งมาให้โดยตรง
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User info missing"})
		}

		userRoles := user.Roles // สมมติว่าใน UserInfo field ชื่อ Roles เป็น []string

		// 2. Convert User Role to Map for checking (Case-Insensitive)
		userRolesMap := make(map[string]bool)
		for _, r := range userRoles {
			userRolesMap[strings.ToUpper(r)] = true
		}

		logs.Info(fmt.Sprintf("RolesGuard Check: UserRoles=%v Allowed=%v", userRoles, allowedRoles))

		// 3. Check permission
		isAllowed := false
		for _, allowed := range allowedRoles {
			// Allowed roles from constants (e.g. models.RoleAdmin) are typically Uppercase.
			// regardless, we check against the Uppercase map key.
			if userRolesMap[strings.ToUpper(allowed)] {
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
