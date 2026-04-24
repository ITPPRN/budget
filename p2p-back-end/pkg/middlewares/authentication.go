package middlewares

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"p2p-back-end/logs"
	"p2p-back-end/modules/entities/models"
)

// =========================================================================
// OLD JWT-BASED AUTH (kept for reference)
// Replaced by Gateway-based authentication (APISIX injects X-User-* headers).
// =========================================================================
//
// import (
// 	"context"
// 	"errors"
// 	"strings"
// 	"sync"
// 	"time"
//
// 	"github.com/lestrrat-go/jwx/v2/jwk"
// 	jwxtoken "github.com/lestrrat-go/jwx/v2/jwt"
//
// 	"p2p-back-end/pkg/utils"
// )
//
// var (
// 	keySet           jwk.Set // Global variable to cache the public keys
// 	once             sync.Once
// 	keycloakIssuer   string
// 	// keycloakClientID string

// 	secretGateWay  string // Secret from API Gateway (e.g. APISIX)
// 	internalSecret string // Secret for internal service-to-service calls
// )

// // InitKeycloakValidator fetches the Public Keys from Keycloak and caches them.
// func InitKeycloakValidator(host string, port string, realm string, clientID string) {
// 	once.Do(func() {
// 		// Construct the necessary OIDC URLs
// 		keycloakIssuer = fmt.Sprintf("http://%s:%s/realms/%s", host, port, realm)
// 		jwksURL := fmt.Sprintf("%s/protocol/openid-connect/certs", keycloakIssuer)
// 		// keycloakClientID = clientID

// 		logs.Infof("Initializing Keycloak Validator, fetching keys from: %s", jwksURL)

// 		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
// 		defer cancel()

// 		var err error
// 		keySet, err = jwk.Fetch(ctx, jwksURL)
// 		if err != nil {
// 			logs.Warnf("⚠️ Fallback Auth: Failed to fetch Keycloak JWKS from %s. Fallback might fail: %v", jwksURL, err)
// 			return
// 		}
// 		logs.Infof("✅ Keycloak Public Keys (JWKS) cached for Fallback Auth. Total keys: %d", keySet.Len())
// 	})
// }

// func InitInternalSecret(secret string) {
// 	internalSecret = secret
// }

// func InitGatewaySecret(secret string) {
// 	secretGateWay = secret
// }

// // JwtAuthentication is Hybrid (Gateway Headers + JWT Fallback)
// func JwtAuthentication(authSrv models.AuthService, handler interface{}) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		// 1. Try Gateway Auth First (Senior's Pattern)
// 		gatewaySecret := c.Get("X-Gateway-Secret")
// 		if gatewaySecret != "" && secretGateWay != "" && gatewaySecret == secretGateWay {
// 			user := GetUserFromHeaders(c)
// 			if user != nil {
// 				return processAuthenticatedUser(c, authSrv, user, handler)
// 			}
// 		}

// 		// 2. Fallback to Local JWT Auth (Local Dev)
// 		accessToken := extractAccessToken(c)
// 		if accessToken != "" {
// 			claims, err := parseAndValidateToken(accessToken)
// 			if err == nil {
// 				user := &models.UserInfo{
// 					ID:       claims.ID,
// 					Username: claims.Username,
// 					Email:    claims.Email,
// 					Roles:    claims.RealmAccess.Roles,
// 				}
// 				return processAuthenticatedUser(c, authSrv, user, handler)
// 			}
// 			logs.Debugf("Local JWT fallback validation failed: %v", err)
// 		}

// 		return unauthorizedResponse(c, "No valid Gateway session or Local JWT provided")
// 	}
// }

// // InternalAuth handles service-to-service communication via a shared secret
// func InternalAuth(handler fiber.Handler) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		providedSecret := c.Get("X-Internal-Secret")
// 		if providedSecret == "" || providedSecret != internalSecret {
// 			logs.Warnf("InternalAuth: Unauthorized access attempt from %s", c.IP())
// 			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 				"status":      "Unauthorized",
// 				"status_code": fiber.StatusUnauthorized,
// 				"message":     "Invalid or missing internal secret",
// 			})
// 		}
// 		return handler(c)
// 	}
// }

// // GetUserFromHeaders extracts User info injected by API Gateway (APISIX/OIDC Plugin)
// func GetUserFromHeaders(c *fiber.Ctx) *models.UserInfo {
// 	userID := c.Get("X-User-ID")
// 	userName := c.Get("X-User-Name")
// 	userEmail := c.Get("X-User-Email")

// 	if userID == "" || userName == "" {
// 		return nil
// 	}

// 	return &models.UserInfo{
// 		ID:       userID,
// 		Username: userName,
// 		Email:    userEmail,
// 	}
// }

// func processAuthenticatedUser(c *fiber.Ctx, authSrv models.AuthService, user *models.UserInfo, handler interface{}) error {
// 	var profile *models.UserInfo
// 	if authSrv != nil {
// 		var err error
// 		profile, err = authSrv.GetUserProfile(c.UserContext(), user.Username)
// 		if err != nil {
// 			// Auto-provision if missing
// 			logs.Warnf("User profile not found for %s (%s). Attempting auto-provisioning...", user.Username, user.ID)
// 			// profile, err = authSrv.ProvisionUser(c.UserContext(), user)
// 			// if err != nil {
// 			// 	logs.Errorf("Auto-provisioning failed for %s: %v", user.Username, err)
// 			// 	profile = user // Fallback to basic info if provisioning fails
// 			// }
// 		}
// 	} else {
// 		profile = user
// 	}

// 	c.Locals("user", profile)
// 	c.Locals("userID", profile.ID)

// 	// Call the wrapped handler if provided
// 	if handler != nil {
// 		if h, ok := handler.(models.TokenHandler); ok {
// 			return h(c, profile)
// 		}
// 		if h, ok := handler.(func(*fiber.Ctx, *models.UserInfo) error); ok {
// 			return h(c, profile)
// 		}
// 		// If it's a standard fiber.Handler, call it but it won't have the userInfo param
// 		if h, ok := handler.(fiber.Handler); ok {
// 			return h(c)
// 		}
// 	}

// 	// Otherwise proceed to next middleware
// 	return c.Next()
// }

// func extractAccessToken(c *fiber.Ctx) string {
// 	token := c.Cookies("access_token")
// 	if token != "" {
// 		return token
// 	}
// 	authHeader := c.Get("Authorization")
// 	if strings.HasPrefix(authHeader, "Bearer ") {
// 		return strings.TrimPrefix(authHeader, "Bearer ")
// 	}
// 	return ""
// }

// func parseAndValidateToken(accessToken string) (*models.JWTPayload, error) {
// 	if keySet == nil {
// 		return nil, errors.New("jwks cache is not initialized")
// 	}

// 	token, err := jwxtoken.Parse(
// 		[]byte(accessToken),
// 		jwxtoken.WithKeySet(keySet),
// 		jwxtoken.WithValidate(true),
// 		// jwxtoken.WithIssuer(keycloakIssuer), // Disabled for dev flexibility
// 	)

// 	if err != nil {
// 		return nil, fmt.Errorf("jwt validation failed: %w", err)
// 	}

// 	claimsMap := token.PrivateClaims()
// 	var realmRoles []string
// 	if rawAccess, ok := claimsMap["realm_access"].(map[string]interface{}); ok && rawAccess != nil {
// 		if rawRoles, ok := rawAccess["roles"].([]interface{}); ok && rawRoles != nil {
// 			realmRoles = utils.ConvertInterfaceSliceToStringSlice(rawRoles)
// 		}
// 	}

// 	return &models.JWTPayload{
// 		ID:       token.Subject(),
// 		Username: utils.GetSafeString(claimsMap, "preferred_username"),
// 		Email:    utils.GetSafeString(claimsMap, "email"),
// 		RealmAccess: models.RealmAccess{
// 			Roles: realmRoles,
// 		},
// 	}, nil
// }

// func unauthorizedResponse(c *fiber.Ctx, message string) error {
// 	logs.Debugf("Unauthorized: %v", message)
// 	return c.Status(fiber.ErrUnauthorized.Code).JSON(fiber.Map{
// 		"status":      fiber.ErrUnauthorized.Message,
// 		"status_code": fiber.ErrUnauthorized.Code,
// 		"message":     fmt.Sprintf("Error: Unauthorized - %s", message),
// 	})
// }

// =========================================================================

var (
	secretGateWay  string // Secret from API Gateway (e.g. APISIX)
	internalSecret string // Secret for internal service-to-service calls
)

// InitKeycloakValidator is kept as a no-op stub for compatibility with main.go.
// Gateway-based auth no longer needs JWKS fetching.
func InitKeycloakValidator(host string, port string, realm string, clientID string) {
	logs.Info("Keycloak JWKS validator disabled: using Gateway-based authentication.")
}

func InitInternalSecret(secret string) {
	internalSecret = secret
}

func InitGatewaySecret(secret string) {
	secretGateWay = secret
	logs.Infof("InitGatewaySecret: loaded=%v len=%d fingerprint=%s", secret != "", len(secret), fingerprint(secret))
}

func fingerprint(s string) string {
	if len(s) == 0 {
		return "<empty>"
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "…" + s[len(s)-2:]
}

// JwtAuthentication validates requests coming through the API Gateway (APISIX).
// Flow:
//  1. Verify X-Gateway-Secret header matches the configured secret.
//  2. Extract user info from X-User-* headers injected by the gateway.
//  3. (Optional) Enrich user profile via AuthService.
//  4. Dispatch to the wrapped handler (TokenHandler / fiber.Handler / nil for .Use).
//
// Function name kept for backward compatibility with existing callers.
func JwtAuthentication(authSrv models.AuthService, handler interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		gatewaySecret := GatewaySecretMiddleware(c)
		if gatewaySecret == "" {
			return unauthorizedResponse(c, "Gateway-Secret header is empty")
		}
		if gatewaySecret != secretGateWay {
			logs.Warnf("Gateway secret mismatch: got=%s expected=%s", fingerprint(gatewaySecret), fingerprint(secretGateWay))
			return unauthorizedResponse(c, "Invalid Gateway-Secret header")
		}

		user := GetUserFromHeaders(c)
		if user == nil {
			return unauthorizedResponse(c, "Access via Gateway only")
		}

		return processAuthenticatedUser(c, authSrv, user, handler)
	}
}

// RequireGatewaySecret blocks direct requests when the gateway secret is configured.
// In local dev (secretGateWay empty) it is a no-op. In production it rejects any
// request missing or mismatching X-Gateway-Secret — ensuring the endpoint is only
// reachable through the APISIX gateway.
func RequireGatewaySecret() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if secretGateWay == "" {
			return c.Next()
		}
		if c.Get("X-Gateway-Secret") != secretGateWay {
			return unauthorizedResponse(c, "Direct access blocked; use gateway")
		}
		return c.Next()
	}
}

// InternalAuth handles service-to-service communication via a shared secret.
func InternalAuth(handler fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		providedSecret := c.Get("X-Internal-Secret")
		if providedSecret == "" || providedSecret != internalSecret {
			logs.Warnf("InternalAuth: Unauthorized access attempt from %s", c.IP())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":      "Unauthorized",
				"status_code": fiber.StatusUnauthorized,
				"message":     "Invalid or missing internal secret",
			})
		}
		return handler(c)
	}
}

// GetUserFromHeaders extracts user info injected by the API Gateway (APISIX/OIDC Plugin).
func GetUserFromHeaders(c *fiber.Ctx) *models.UserInfo {
	userID := c.Get("X-User-ID")
	userName := c.Get("X-User-Name")
	userEmail := c.Get("X-User-Email")

	if userID == "" || userName == "" {
		return nil
	}

	return &models.UserInfo{
		ID:       userID,
		Username: userName,
		Email:    userEmail,
		// Roles: populated by processAuthenticatedUser via AuthService (if available).
	}
}

// GatewaySecretMiddleware returns the X-Gateway-Secret value, or "" if missing.
func GatewaySecretMiddleware(c *fiber.Ctx) string {
	token := c.Get("X-Gateway-Secret")
	if token != "" {
		return token
	}
	return ""
}

func processAuthenticatedUser(c *fiber.Ctx, authSrv models.AuthService, user *models.UserInfo, handler interface{}) error {
	profile := user
	if authSrv != nil {
		loaded, err := authSrv.GetUserProfile(c.UserContext(), user.Username)
		if err != nil {
			logs.Warnf("User profile not found for %s (%s). Using gateway header info only.", user.Username, user.ID)
		} else if loaded != nil {
			profile = loaded
		}
	}

	c.Locals("user", profile)
	c.Locals("userID", profile.ID)

	if handler != nil {
		if h, ok := handler.(models.TokenHandler); ok {
			return h(c, profile)
		}
		if h, ok := handler.(func(*fiber.Ctx, *models.UserInfo) error); ok {
			return h(c, profile)
		}
		if h, ok := handler.(fiber.Handler); ok {
			return h(c)
		}
	}

	return c.Next()
}

func unauthorizedResponse(c *fiber.Ctx, message string) error {
	logs.Debugf("Unauthorized: %v", message)
	return c.Status(fiber.ErrUnauthorized.Code).JSON(fiber.Map{
		"status":      fiber.ErrUnauthorized.Message,
		"status_code": fiber.ErrUnauthorized.Code,
		"message":     fmt.Sprintf("Error: Unauthorized - %s", message),
	})
}
