package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-redis/redis/v8"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/errs"
)

type authService struct {
	keycloak *gocloak.GoCloak
	cfg      *configs.Config
	authRepo models.UserRepository
	Redis    *redis.Client
}

func NewAuthService(
	keycloak *gocloak.GoCloak,
	cfg *configs.Config,
	authRepo models.UserRepository,
	Redis *redis.Client,
) models.AuthService {
	return &authService{keycloak, cfg, authRepo, Redis}
}

func (s authService) Register(req *models.RegisterKCReq) (string, error) {
	ctx := context.Background()

	token, err := s.keycloak.LoginAdmin(ctx, s.cfg.KeyCloak.AdminUsername, s.cfg.KeyCloak.AdminPassword, "master")
	if err != nil {
		logs.Error(err)
		return "", errs.NewUnexpectedError()
	}

	user := gocloak.User{
		FirstName: gocloak.StringP(req.FirstName),
		LastName:  gocloak.StringP(req.LastName),
		Email:     gocloak.StringP(req.Email),
		Enabled:   gocloak.BoolP(true),
		Username:  gocloak.StringP(req.Username),
	}

	var rolesToAdd []gocloak.Role

	if len(req.Roles) > 0 {
		for _, r := range req.Roles {
			roleName := strings.ToLower(r)

			// เช็คว่า Role มีอยู่จริงไหม
			role, err := s.keycloak.GetRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, roleName)

			if err != nil {
				if strings.Contains(err.Error(), "404 Not Found") || strings.Contains(err.Error(), "Could not find role") {
					logs.Warnf("Role '%s' not found. Creating automatically...", roleName)

					newRole := gocloak.Role{
						Name:        gocloak.StringP(roleName),
						Description: gocloak.StringP(fmt.Sprintf("Auto-generated role: %s", roleName)),
					}

					_, err = s.keycloak.CreateRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, newRole)
					if err != nil {
						logs.Errorf("Failed to auto-create role '%s': %v", roleName, err)
						return "", errs.NewUnexpectedError() // ถ้าสร้างไม่ผ่าน ก็จบเลย
					}

					role, err = s.keycloak.GetRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, roleName)
					if err != nil {
						logs.Error(err)
						return "", errs.NewUnexpectedError()
					}
				} else {
					logs.Error(err)
					return "", errs.NewUnexpectedError()
				}
			}
			if role != nil {
				rolesToAdd = append(rolesToAdd, *role)
			}
		}
	}

	// role, err := s.keycloak.GetRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, roleName)
	// // if err != nil {
	// // 	logs.Error(err)
	// // 	return "", errs.NewUnexpectedError()
	// // }
	// if err != nil {
	// 	// ถ้าเกิด Error: 404 Not Found (Could not find role)
	// 	// เราจะทำการสร้าง Role นั้นขึ้นมา
	// 	if strings.Contains(err.Error(), "404 Not Found") || strings.Contains(err.Error(), "Could not find role") {
	// 		logs.Warnf("Role '%s' not found in Keycloak. Attempting to create it automatically.", roleName)

	// 		// สร้าง Role object ใหม่
	// 		newRole := gocloak.Role{
	// 			Name:        gocloak.StringP(roleName),
	// 			Description: gocloak.StringP(fmt.Sprintf("Auto-generated role for user type: %s", roleName)),
	// 		}

	// 		// เรียก API สร้าง Role
	// 		_, err = s.keycloak.CreateRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, newRole)
	// 		if err != nil {
	// 			logs.Errorf("FATAL: Failed to auto-create role '%s': %v", roleName, err)
	// 			return "", errs.NewUnexpectedError()
	// 		}
	// 		logs.Infof("✅ Successfully auto-created role: %s", roleName)

	// 		// ดึง Role ที่สร้างใหม่มาอีกครั้งเพื่อใช้ในขั้นตอนต่อไป
	// 		role, err = s.keycloak.GetRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, roleName)
	// 		if err != nil {
	// 			logs.Errorf("Failed to retrieve newly created role '%s': %v", roleName, err)
	// 			return "", errs.NewUnexpectedError()
	// 		}

	// 	} else {
	// 		// ถ้าเป็น Error อื่นๆ ที่ไม่ใช่ 404 (เช่น Connection Error) ให้ Fail
	// 		logs.Error(err)
	// 		return "", errs.NewUnexpectedError()
	// 	}
	// }

	userID, err := s.keycloak.CreateUser(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, user)
	if err != nil {
		logs.Error(err)
		return "", errs.NewUnexpectedError()
	}

	err = s.keycloak.SetPassword(ctx, token.AccessToken, userID, s.cfg.KeyCloak.RealmName, req.Password, false)
	if err != nil {
		logs.Error(err)
		return "", errs.NewUnexpectedError()
	}

	if len(rolesToAdd) > 0 {
		// ฟังก์ชันนี้รับ []gocloak.Role อยู่แล้ว ใส่ slice เข้าไปได้เลย
		err = s.keycloak.AddRealmRoleToUser(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, userID, rolesToAdd)
		if err != nil {
			logs.Error(err)
			return "", errs.NewUnexpectedError()
		}
	}

	// // กำหนดเฉพาะบทบาทที่ระบุใน req.Role
	// err = s.keycloak.AddRealmRoleToUser(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, userID, []gocloak.Role{*role})
	// if err != nil {
	// 	logs.Error(err)
	// 	return "", errs.NewUnexpectedError()
	// }

	// ลบบทบาท default-roles-master ถ้ามันถูกเพิ่มโดยอัตโนมัติ
	defaultRole, err := s.keycloak.GetRealmRole(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, "default-roles-"+s.cfg.KeyCloak.RealmName)
	if err == nil {
		err = s.keycloak.DeleteRealmRoleFromUser(ctx, token.AccessToken, s.cfg.KeyCloak.RealmName, userID, []gocloak.Role{*defaultRole})
		if err != nil {
			logs.Error(err)
			return "", errs.NewUnexpectedError()
		}
	}

	_, err = s.keycloak.Login(ctx, s.cfg.KeyCloak.ClientID, s.cfg.KeyCloak.ClientSecret, s.cfg.KeyCloak.RealmName, req.Username, req.Password)
	if err != nil {
		logs.Error(err)
		return "", errs.NewLoginFailedError()
	}
	return userID, nil
}

func (s authService) Login(req *models.LoginReq) (string, error) {

	ctx := context.Background()

	redisKey := fmt.Sprintf("login_attempts:%s", req.Username)

	token, err := s.keycloak.Login(ctx, s.cfg.KeyCloak.ClientID, s.cfg.KeyCloak.ClientSecret, s.cfg.KeyCloak.RealmName, req.Username, req.Password)
	if err != nil {

		logs.Error(err)
		// --- ตรวจสอบ Error ว่าโดนล็อคจาก Keycloak หรือยัง? ---
		errStr := err.Error()
		// เช็คข้อความ Error ที่ Keycloak ส่งมา
		if strings.Contains(errStr, "Account disabled") || strings.Contains(errStr, "Account temporarily disabled") {
			return "", errors.New("Account Locked: บัญชีของคุณถูกระงับถาวร กรุณาติดต่อผู้ดูแลระบบ")
		}

		// --- 2. นับจำนวนผิดใน Redis (เฉพาะกรณีที่ยังไม่โดนล็อค) ---
		failCount, _ := s.Redis.Incr(ctx, redisKey).Result()

		// ตั้งเวลาหมดอายุของ Key นี้ (10 นาที) เพื่อไม่ให้ค้างตลอดไป
		if failCount == 1 {
			s.Redis.Expire(ctx, redisKey, 10*time.Minute)
		}

		// --- 3. เช็คเงื่อนไขการแจ้งเตือน ---
		if failCount == 3 {
			// ผิดครั้งที่ 3: ส่ง Error Message เตือน
			return "", errors.New("Warning: คุณใส่รหัสผิด 3 ครั้งแล้ว โปรดตรวจสอบรหัสให้ดี")
		}

		if failCount >= 5 {
			// ผิดครั้งที่ 5 หรือมากกว่า: แจ้งว่าโดนล็อค (User จะเห็นข้อความนี้ก่อนที่ Keycloak จะส่ง error disabled ในครั้งถัดไป)
			return "", errors.New("Account Locked: คุณใส่รหัสผิดเกิน 5 ครั้ง บัญชีถูกระงับถาวร กรุณาติดต่อผู้ดูแลระบบ")
		}
		return "", errs.NewLoginFailedError()
	}

	// --- กรณี Login สำเร็จ ---
	// ล้างตัวนับใน Redis ทิ้งทันที (Reset Counter)
	s.Redis.Del(ctx, redisKey)

	return token.AccessToken, nil
}
