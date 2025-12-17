// package auth

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http/httptest"
// 	"testing"
//
//
//
//
//
//
//

// 	"github.com/Nerzal/gocloak/v13"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/require"
// 	"go.uber.org/zap"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"

// 	"p2p-back-end/configs"
// 	"p2p-back-end/logs"
// 	"p2p-back-end/modules/auth/controller"
// 	"p2p-back-end/modules/auth/repository"
// 	"p2p-back-end/modules/auth/service"
// 	"p2p-back-end/modules/entities/models"
// )

// func TestAuthIntegration(t *testing.T) {
// 	// ปิด log
// 	logs.Logger = zap.NewNop()

// 	// --------------------
// 	// setup database (SQLite in-memory)
// 	// --------------------
// 	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
// 	require.NoError(t, err)
// 	err = db.AutoMigrate(&models.UserEntity{})
// 	require.NoError(t, err)

// 	authRepo := repository.NewAuthRepositoryDB(db)

// 	// --------------------
// 	// setup Keycloak config จริง
// 	// --------------------
// 	cfg := &configs.Config{
// 		KeyCloak: configs.KeyCloak{
// 			Host:          "http://localhost", // เปลี่ยนเป็น host Keycloak จริง
// 			Port:          "9080",      // port Keycloak
// 			RealmName:     "P2P-service",
// 			ClientID:      "p2p-client",
// 			ClientSecret:  "4J30ZBRYRUTJcw8YlfUrTeA4qxKitBuu",
// 			AdminUsername: "admin",
// 			AdminPassword: "secure+admin+password+p2p",
// 		},
// 	}

// 	// --------------------
// 	// service ใช้ Keycloak จริง
// 	// --------------------
// 	keycloakClient := gocloak.NewClient(cfg.KeyCloak.Host + ":" + cfg.KeyCloak.Port)
// 	authSrv := service.NewAuthService(keycloakClient, cfg, authRepo)

// 	// --------------------
// 	// controller + fiber app
// 	// --------------------
// 	app := fiber.New()
// 	controller.NewUserController(app.Group("/v1/auth"), authSrv)

// 	// --------------------
// 	// Test Register
// 	// --------------------
// 	registerReq := models.RegisterReq{
// 		FirstName: "Integrat1ion",
// 		LastName:  "T1est",
// 		Email:     "int_te1st@example.com",
// 		Username:  "int_test11",
// 		Password:  "pass121134",
// 		Role:      "employee",
// 	}
// 	body, _ := json.Marshal(registerReq)
// 	req := httptest.NewRequest("POST", "/v1/auth/register", bytes.NewReader(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	resp, _ := app.Test(req, -1)
// 	require.Equal(t, 200, resp.StatusCode)

// 	// --------------------
// 	// Test Login
// 	// --------------------
// 	loginReq := models.LoginReq{
// 		Username: "int_test11",
// 		Password: "pass121134",
// 	}
// 	bodyLogin, _ := json.Marshal(loginReq)
// 	reqLogin := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewReader(bodyLogin))
// 	reqLogin.Header.Set("Content-Type", "application/json")
// 	respLogin, _ := app.Test(reqLogin, -1)
// 	require.Equal(t, 200, respLogin.StatusCode)

// 	// --------------------
// 	// Test repository directly
// 	// --------------------
// 	user := models.UserEntity{
// 		ID:        uuid.NewString(),
// 		Username:  registerReq.Username,
// 		Email:     registerReq.Email,
// 		FirstName: registerReq.FirstName,
// 		LastName:  registerReq.LastName,
// 		Role:      registerReq.Role,
// 		Password:  "hashed-pass",
// 	}
// 	db.Create(&user)

// 	exist, err := authRepo.IsUserExistByID(user.ID)
// 	require.NoError(t, err)
// 	require.True(t, exist)
// }

package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"p2p-back-end/configs"
	"p2p-back-end/logs"
	"p2p-back-end/modules/auth/controller"
	"p2p-back-end/modules/auth/repository"
	"p2p-back-end/modules/auth/service"
	"p2p-back-end/modules/entities/models"
)

func TestAuthIntegration(t *testing.T) {
	// ปิด log เพื่อความสะอาดของ Output
	logs.Logger = zap.NewNop()

	// --------------------
	// 1. Setup Database (SQLite in-memory)
	// --------------------
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&models.UserEntity{})
	require.NoError(t, err)

	authRepo := repository.NewAuthRepositoryDB(db)

	// --------------------
	// 2. Setup Keycloak Config (ต้องมี Keycloak รันอยู่จริงตาม config นี้)
	// --------------------
	cfg := &configs.Config{
		KeyCloak: configs.KeyCloak{
			Host:          "http://localhost",
			Port:          "9080", // เช็คว่า port ตรงกับ docker-compose
			RealmName:     "P2P-service",
			ClientID:      "p2p-client",
			ClientSecret:  "4J30ZBRYRUTJcw8YlfUrTeA4qxKitBuu", // เช็ค secret ให้ตรง
			AdminUsername: "admin",
			AdminPassword: "secure+admin+password+p2p",
		},
	}

	keycloakClient := gocloak.NewClient(cfg.KeyCloak.Host + ":" + cfg.KeyCloak.Port)
	authSrv := service.NewAuthService(keycloakClient, cfg, authRepo)

	// --------------------
	// 3. Setup Fiber App
	// --------------------
	app := fiber.New()
	controller.NewUserController(app.Group("/v1/auth"), authSrv)

	// สร้างตัวแปร Random เพื่อไม่ให้ User ซ้ำเวลารันเทสหลายรอบ
	randID := uuid.NewString()[:5]
	testUsername := "test_user_" + randID
	testEmail := "test_" + randID + "@example.com"
	testPassword := "pass123456"

	// --------------------
	// Test 1: Register
	// --------------------
	registerReq := models.RegisterReq{
		FirstName: "Integration",
		LastName:  "Test",
		Email:     testEmail,
		Username:  testUsername,
		Password:  testPassword,
		Role:      "employee",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode, "Register should return 200 OK")

	// --------------------
	// Test 2: Login (สำคัญ: ต้องเช็ค Cookie)
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
	require.Equal(t, 200, respLogin.StatusCode, "Login should return 200 OK")

	// --- [Check 1] ตรวจสอบ Response Body (ต้องเป็น UserInfo ไม่ใช่ Token) ---
	var result models.ResponseData
	respBody, _ := io.ReadAll(respLogin.Body)
	err = json.Unmarshal(respBody, &result)
	require.NoError(t, err)

	// แปลง data interface{} ให้เป็น map เพื่อเช็คค่าข้างใน
	userDataMap, ok := result.Data.(map[string]interface{})
	if ok {
		require.Equal(t, testUsername, userDataMap["username"], "Response Body should contain correct username")
		// ต้องไม่มี Token ใน Body
		require.Nil(t, userDataMap["access_token"], "Body should NOT contain access_token")
	}

	// --- [Check 2] ตรวจสอบ Cookie (สำคัญที่สุด) ---
	cookies := respLogin.Cookies()
	var accessTokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "access_token" {
			accessTokenCookie = c
			break
		}
	}

	require.NotNil(t, accessTokenCookie, "Response must contain 'access_token' cookie")
	require.True(t, accessTokenCookie.HttpOnly, "Cookie must be HttpOnly")
	require.NotEmpty(t, accessTokenCookie.Value, "Token in cookie must not be empty")

	// --------------------
	// Test 3: Repository Directly (Mock Check)
	// --------------------
	userRepoCheck := models.UserEntity{
		ID:        uuid.NewString(),
		Username:  testUsername, // ใช้ชื่อเดียวกับข้างบน
		Email:     testEmail,
		FirstName: "RepoCheck",
		LastName:  "Test",
		Role:      "employee",
		Password:  "hashed-pass",
	}
	db.Create(&userRepoCheck)

	exist, err := authRepo.IsUserExistByID(userRepoCheck.ID)
	require.NoError(t, err)
	require.True(t, exist)
}
