package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/datatypes" // ✅ เพิ่ม Import นี้สำหรับ JSON Type
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/modules/auth/controller"
	"p2p-back-end/modules/auth/service"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/modules/users/repository" // ✅ เรียกใช้ User Repo จาก Module Users
)

func TestAuthIntegration(t *testing.T) {
	// ปิด log
	logs.Logger = zap.NewNop()

	// --------------------
	// 1. Setup Database
	// --------------------
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	// AutoMigrate จะสร้างตารางตาม UserEntity ล่าสุด (ที่มี Roles เป็น JSON)
	err = db.AutoMigrate(&models.UserEntity{})
	require.NoError(t, err)

	// ✅ เปลี่ยนมาใช้ User Repository แทน Auth Repository เดิม
	userRepo := repository.NewUserRepositoryDB(db)

	// --------------------
	// 2. Setup Config
	// --------------------
	cfg := &configs.Config{
		KeyCloak: configs.KeyCloak{
			Host:          "http://localhost",
			Port:          "9080",
			RealmName:     "P2P-service",
			ClientID:      "p2p-client",
			ClientSecret:  "4J30ZBRYRUTJcw8YlfUrTeA4qxKitBuu",
			AdminUsername: "admin",
			AdminPassword: "secure+admin+password+p2p",
		},
		Redis: configs.Redis{
			Host:     "localhost",
			Port:     "6379",
			Password: "pr01ec1s413",
		},
	}

	// --------------------
	// 3. Setup Redis
	// --------------------
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
	})
	// Ping เช็คว่า Redis ติดจริงไหม (Optional)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		t.Logf("Warning: Redis connection failed: %v", err)
	}

	// --------------------
	// 4. Setup Service
	// --------------------
	keycloakClient := gocloak.NewClient(cfg.KeyCloak.Host + ":" + cfg.KeyCloak.Port)

	// ✅ Inject 'userRepo' เข้าไปแทนที่
	authSrv := service.NewAuthService(keycloakClient, cfg, userRepo, redisClient)

	// --------------------
	// 5. Setup Fiber App
	// --------------------
	app := fiber.New()
	controller.NewUserController(app.Group("/v1/auth"), authSrv)

	randID := uuid.NewString()[:5]
	testUsername := "test_user_" + randID
	testEmail := "test_" + randID + "@example.com"
	testPassword := "pass123456"

	// --------------------
	// Test 1: Register
	// --------------------
	registerReq := models.RegisterKCReq{
		FirstName: "Integration",
		LastName:  "Test",
		Email:     testEmail,
		Username:  testUsername,
		Password:  testPassword,
		Roles:     []string{"employee", "manager"}, // ✅ ส่ง Roles เป็น Array
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// --------------------
	// Test 2: Login
	// --------------------
	loginReq := models.LoginReq{
		Username: testUsername,
		Password: testPassword,
	}
	bodyLogin, _ := json.Marshal(loginReq)
	reqLogin := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(bodyLogin))
	reqLogin.Header.Set("Content-Type", "application/json")

	respLogin, err := app.Test(reqLogin, -1)
	require.NoError(t, err)
	require.Equal(t, 200, respLogin.StatusCode)

	// Check Response Body
	var result models.ResponseData
	respBody, _ := io.ReadAll(respLogin.Body)
	json.Unmarshal(respBody, &result)
	userDataMap, ok := result.Data.(map[string]interface{})
	if ok {
		require.Equal(t, testUsername, userDataMap["username"])
		require.Nil(t, userDataMap["access_token"])
	}

	// Check Cookie
	cookies := respLogin.Cookies()
	var accessTokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "access_token" {
			accessTokenCookie = c
			break
		}
	}
	require.NotNil(t, accessTokenCookie)

	// --------------------
	// Test 3: Repository Directly (แก้ไขการสร้าง Mock Data)
	// --------------------

	// แปลง Array String เป็น JSON Byte สำหรับเก็บลง DB จำลอง
	rolesJSON, _ := json.Marshal([]string{"employee"})

	userRepoCheck := models.UserEntity{
		ID:        uuid.NewString(),
		Username:  testUsername, // ซ้ำได้ใน test เพราะ mock db คนละตัวหรือเคลียร์ใหม่
		Email:     testEmail,
		FirstName: "RepoCheck",
		LastName:  "Test",

		// ✅ แก้ตรงนี้: ใช้ Roles (JSON) แทน Role (String)
		Roles: datatypes.JSON(rolesJSON),

		// Password:  "hashed-pass", // field นี้อาจจะไม่มีแล้วใน model ใหม่ถ้าเก็บแต่ใน KC
	}

	// Create ลง DB (SQLite memory)
	err = db.Create(&userRepoCheck).Error
	require.NoError(t, err)

	// เรียกใช้ UserRepo (ที่ย้ายไป Module Users)
	exist, err := userRepo.IsUserExistByID(userRepoCheck.ID)
	require.NoError(t, err)
	require.True(t, exist)
}
