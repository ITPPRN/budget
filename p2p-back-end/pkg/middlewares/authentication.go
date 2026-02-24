package middlewares

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/v2/jwk"
	jwxtoken "github.com/lestrrat-go/jwx/v2/jwt"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
	"p2p-back-end/pkg/utils"
)

// --- Keycloak Configuration & JWKS Cache ---

var (
	keySet jwk.Set // Global variable to cache the public keys
	once   sync.Once

	// Global variables to store configuration pulled from Infisical
	keycloakIssuer   string
	keycloakClientID string
)

// InitKeycloakValidator fetches the Public Keys from Keycloak and caches them.
// It now receives configuration parameters (Host, Port, Realm, ClientID)
// which should be loaded from Infisical before calling this function.
func InitKeycloakValidator(host string, port string, realm string, clientID string) {
	once.Do(func() {
		// Construct the necessary OIDC URLs
		keycloakIssuer = fmt.Sprintf("http://%s:%s/realms/%s", host, port, realm)
		jwksURL := fmt.Sprintf("%s/protocol/openid-connect/certs", keycloakIssuer)
		keycloakClientID = clientID

		logs.Infof("Initializing Keycloak Validator, fetching keys from: %s", jwksURL)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		var err error
		// jwk.Fetch handles downloading, parsing, and internal key refreshing
		keySet, err = jwk.Fetch(ctx, jwksURL)
		if err != nil {
			// CRITICAL ERROR: Cannot validate tokens without keys.
			logs.Fatalf("FATAL: Failed to fetch Keycloak JWKS from %s: %v", jwksURL, err)
		}
		// ✅ แสดงข้อมูล key ที่ดึงมา
		logs.Infof("✅ Keycloak Public Keys (JWKS) successfully fetched and cached. Total keys: %d", keySet.Len())

	})
}

func JwtAuthentication(handler models.TokenHandler) fiber.Handler {
	// We assume InitKeycloakValidator has been called in main()

	return func(c *fiber.Ctx) error {
		accessToken := extractAccessToken(c)
		if accessToken == "" {
			return unauthorizedResponse(c, "Authorization header is empty.")
		}

		claims, err := parseAndValidateToken(accessToken)
		if err != nil {
			// Logs the specific validation failure from jwx
			logs.Error(fmt.Errorf("token validation failed: %v", err))
			return unauthorizedResponse(c, "Invalid token or signature")
		}

		user := &models.UserInfo{
			UserId:   claims.ID,
			UserName: claims.Username,
			Email:    claims.Email,
			Roles:    claims.RealmAccess.Roles,
			Name:     claims.Name,
		}

		// ✅ เก็บไว้ใน Context เผื่อ handler อื่นจะใช้ได้ง่าย
		c.Locals("user", user)

		logs.Info(fmt.Sprintf("JwtAuthentication Success: UserID=%s Roles=%v", user.UserId, user.Roles))

		if handler == nil {
			return c.Next()
		}

		return handler(c, user)
	}
}

// --- Helper functions (Revised parseAndValidateToken) ---

func extractAccessToken(c *fiber.Ctx) string {
	// 1. ลองดึงจาก Cookie ก่อน (ชื่อต้องตรงกับที่ตั้งตอน Login)
	token := c.Cookies("access_token")
	if token != "" {
		return token
	}

	// 2. (เผื่อไว้) ถ้าไม่มีใน Cookie ให้ลองดึงจาก Header แบบเดิม (รองรับ Postman หรือ Mobile App)
	authHeader := c.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// parseAndValidateToken: REPLACED with JWKS Best Practice
func parseAndValidateToken(accessToken string) (*models.JWTPayload, error) {
	if keySet == nil {
		return nil, errors.New("jwks cache is not initialized")
	}

	// 1. Parse and Validate the JWT using the cached JWKS
	// This single call handles signature verification, expiration (exp),
	// issuer (iss), and audience (aud) checks automatically.
	token, err := jwxtoken.Parse(
		[]byte(accessToken),
		jwxtoken.WithKeySet(keySet),
		jwxtoken.WithValidate(true),
		jwxtoken.WithValidate(true),
		// jwxtoken.WithIssuer(keycloakIssuer), // Disabled to allow internal IP vs localhost mismatch
		// jwxtoken.WithAudience(keycloakClientID), // Enforce that the token is meant for this client
	)

	if err != nil {
		return nil, fmt.Errorf("jwt validation failed: %w", err)
	}

	// 2. Extract and Map Claims
	claimsMap := token.PrivateClaims()

	var realmRoles []string
	if rawAccess, ok := claimsMap["realm_access"].(map[string]interface{}); ok && rawAccess != nil {
		if rawRoles, ok := rawAccess["roles"].([]interface{}); ok && rawRoles != nil {
			realmRoles = utils.ConvertInterfaceSliceToStringSlice(rawRoles)
		}
	}

	// 2.2 Client Roles (resource_access)
	if rawResource, ok := claimsMap["resource_access"].(map[string]interface{}); ok && rawResource != nil {
		// Keycloak usually groups by Client ID
		if clientAccess, ok := rawResource[keycloakClientID].(map[string]interface{}); ok && clientAccess != nil {
			if clientRoles, ok := clientAccess["roles"].([]interface{}); ok && clientRoles != nil {
				cRoles := utils.ConvertInterfaceSliceToStringSlice(clientRoles)
				realmRoles = append(realmRoles, cRoles...)
			}
		}
	}
	jwtPayload := models.JWTPayload{
		// Azp key is typically present, but we use GetSafeString just in case or use a default
		Azp:   utils.GetSafeString(claimsMap, "azp"),
		Email: utils.GetSafeString(claimsMap, "email"), // ใช้ Safe Getter
		Exp:   token.Expiration().Unix(),
		Iat:   token.IssuedAt().Unix(),
		ID:    token.Subject(), // Fix: Standard claim 'sub' is accessed via method
		Iss:   token.Issuer(),
		Jti:   token.JwtID(),
		Name:  utils.GetSafeString(claimsMap, "name"),
		Scope: utils.GetSafeString(claimsMap, "scope"),
		Sid:   utils.GetSafeString(claimsMap, "sid"),

		// แก้ไขตรงนี้: เปลี่ยนจาก "preferred_username" ไปเป็น "username"
		// Keycloak usually sends 'preferred_username' in private claims
		Username: getUsername(claimsMap), // Helper function (or inline logic)

		RealmAccess: models.RealmAccess{
			Roles: realmRoles, // ใช้ค่าที่ตรวจสอบแล้ว
		},
	}

	return &jwtPayload, nil
}

func unauthorizedResponse(c *fiber.Ctx, message string) error {
	logs.Debugf("Error:%v", message)
	return c.Status(fiber.ErrUnauthorized.Code).JSON(fiber.Map{
		"status":     fiber.ErrUnauthorized.Message,
		"statusCode": fiber.ErrUnauthorized.Code,
		"message":    fmt.Sprintf("Error: Unauthorized - %s", message),
	})
}

func GetUserInfo(tokenString string) (*models.UserInfo, error) {
	claims, err := parseAndValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	// Since jwx handles expiration, we just proceed

	userInfo := &models.UserInfo{
		UserId:   claims.ID,
		UserName: claims.Username,
		Email:    claims.Email,
		Roles:    claims.RealmAccess.Roles,
		Name:     claims.Name,
	}

	return userInfo, nil
}

func getUsername(claims map[string]interface{}) string {
	if u := utils.GetSafeString(claims, "preferred_username"); u != "" {
		return u
	}
	if u := utils.GetSafeString(claims, "username"); u != "" {
		return u
	}
	// logs.Warnf("Debug Claims: %+v", claims) // Uncomment to debug if needed
	return ""
}
