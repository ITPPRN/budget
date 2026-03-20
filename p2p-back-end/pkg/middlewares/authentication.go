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

var (
	keySet           jwk.Set // Global variable to cache the public keys
	once             sync.Once
	keycloakIssuer   string
	keycloakClientID string

	secretGateWay  string // Secret from API Gateway (e.g. APISIX)
	internalSecret string // Secret for internal service-to-service calls
)

// InitKeycloakValidator fetches the Public Keys from Keycloak and caches them.
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
		keySet, err = jwk.Fetch(ctx, jwksURL)
		if err != nil {
			logs.Warnf("⚠️ Fallback Auth: Failed to fetch Keycloak JWKS from %s. Fallback might fail: %v", jwksURL, err)
			return
		}
		logs.Infof("✅ Keycloak Public Keys (JWKS) cached for Fallback Auth. Total keys: %d", keySet.Len())
	})
}

func InitInternalSecret(secret string) {
	internalSecret = secret
}

func InitGatewaySecret(secret string) {
	secretGateWay = secret
}

// JwtAuthentication is Hybrid (Gateway Headers + JWT Fallback)
func JwtAuthentication(authSrv models.AuthService, handler interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Try Gateway Auth First (Senior's Pattern)
		gatewaySecret := c.Get("X-Gateway-Secret")
		if gatewaySecret != "" && secretGateWay != "" && gatewaySecret == secretGateWay {
			user := GetUserFromHeaders(c)
			if user != nil {
				return processAuthenticatedUser(c, authSrv, user, handler)
			}
		}

		// 2. Fallback to Local JWT Auth (Local Dev)
		accessToken := extractAccessToken(c)
		if accessToken != "" {
			claims, err := parseAndValidateToken(accessToken)
			if err == nil {
				user := &models.UserInfo{
					ID:       claims.ID,
					Username: claims.Username,
					Email:    claims.Email,
					Roles:    claims.RealmAccess.Roles,
				}
				return processAuthenticatedUser(c, authSrv, user, handler)
			}
			logs.Debugf("Local JWT fallback validation failed: %v", err)
		}

		return unauthorizedResponse(c, "No valid Gateway session or Local JWT provided")
	}
}

// InternalAuth handles service-to-service communication via a shared secret
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

// GetUserFromHeaders extracts User info injected by API Gateway (APISIX/OIDC Plugin)
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
	}
}

func processAuthenticatedUser(c *fiber.Ctx, authSrv models.AuthService, user *models.UserInfo, handler interface{}) error {
	var profile *models.UserInfo
	if authSrv != nil {
		var err error
		profile, err = authSrv.GetUserProfile(c.UserContext(), user.ID)
		if err != nil {
			// Auto-provision if missing
			logs.Warnf("User profile not found for %s (%s). Attempting auto-provisioning...", user.Username, user.ID)
			profile, err = authSrv.ProvisionUser(c.UserContext(), user)
			if err != nil {
				logs.Errorf("Auto-provisioning failed for %s: %v", user.Username, err)
				profile = user // Fallback to basic info if provisioning fails
			}
		}
	} else {
		profile = user
	}

	c.Locals("user", profile)

	// Call the wrapped handler if provided
	if handler != nil {
		if h, ok := handler.(models.TokenHandler); ok {
			return h(c, profile)
		}
		if h, ok := handler.(func(*fiber.Ctx, *models.UserInfo) error); ok {
			return h(c, profile)
		}
		// If it's a standard fiber.Handler, call it but it won't have the userInfo param
		if h, ok := handler.(fiber.Handler); ok {
			return h(c)
		}
	}

	// Otherwise proceed to next middleware
	return c.Next()
}

func extractAccessToken(c *fiber.Ctx) string {
	token := c.Cookies("access_token")
	if token != "" {
		return token
	}
	authHeader := c.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

func parseAndValidateToken(accessToken string) (*models.JWTPayload, error) {
	if keySet == nil {
		return nil, errors.New("jwks cache is not initialized")
	}

	token, err := jwxtoken.Parse(
		[]byte(accessToken),
		jwxtoken.WithKeySet(keySet),
		jwxtoken.WithValidate(true),
		// jwxtoken.WithIssuer(keycloakIssuer), // Disabled for dev flexibility
	)

	if err != nil {
		return nil, fmt.Errorf("jwt validation failed: %w", err)
	}

	claimsMap := token.PrivateClaims()
	var realmRoles []string
	if rawAccess, ok := claimsMap["realm_access"].(map[string]interface{}); ok && rawAccess != nil {
		if rawRoles, ok := rawAccess["roles"].([]interface{}); ok && rawRoles != nil {
			realmRoles = utils.ConvertInterfaceSliceToStringSlice(rawRoles)
		}
	}

	return &models.JWTPayload{
		ID:       token.Subject(),
		Username: utils.GetSafeString(claimsMap, "preferred_username"),
		Email:    utils.GetSafeString(claimsMap, "email"),
		RealmAccess: models.RealmAccess{
			Roles: realmRoles,
		},
	}, nil
}

func unauthorizedResponse(c *fiber.Ctx, message string) error {
	logs.Debugf("Unauthorized: %v", message)
	return c.Status(fiber.ErrUnauthorized.Code).JSON(fiber.Map{
		"status":      fiber.ErrUnauthorized.Message,
		"status_code": fiber.ErrUnauthorized.Code,
		"message":     fmt.Sprintf("Error: Unauthorized - %s", message),
	})
}
