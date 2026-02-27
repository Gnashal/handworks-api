package middleware

import (
	"handworks-api/utils"
	"net/http"
	"os"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
)

type ContextString string

const ClerkClaimsKey ContextString = "clerk-claims"

func getSessionToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}

func ClerkAuthMiddleware(publicPaths []string, logger *utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// ADD DEBUG LOGGING
		logger.Info("=== AUTH MIDDLEWARE DEBUG ===")
		logger.Info("Request: %s %s", method, path)
		logger.Info("Full URL: %s", c.Request.URL.String())

		// Check each public path
		for i, p := range publicPaths {
			match := strings.HasPrefix(path, p)
			logger.Info("Check %d: path='%s' starts with '%s' = %v", i, path, p, match)
			if match {
				logger.Info("Path is public. Allowing access without auth")
				c.Next()
				return
			}
		}


		config := &clerk.ClientConfig{}
		config.Key = clerk.String(os.Getenv("CLERK_SECRET_KEY"))
		jwksClient := jwks.NewClient(config)

		// Get JWT from header
		token := getSessionToken(c.Request)
		if token == "" {
			logger.Info("No Authorization header found")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Decode without verifying to get key ID
		unsafeClaims, err := jwt.Decode(c.Request.Context(), &jwt.DecodeParams{
			Token: token,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Fetch JWKS for key ID
		jwk, err := jwt.GetJSONWebKey(c.Request.Context(), &jwt.GetJSONWebKeyParams{
			KeyID:      unsafeClaims.KeyID,
			JWKSClient: jwksClient,
		})
		if err != nil {
			logger.Info("JWK fetch error: %s", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Verify the token
		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token: token,
			JWK:   jwk,
		})
		if err != nil {
			logger.Info("Token verification error: %s", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Save claims into Gin context
		c.Set(string(ClerkClaimsKey), claims)
		logger.Info("Token verified successfully for user")

		c.Next()
	}
}
