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

func ClerkAuthMiddleware(publicPaths []string, logger * utils.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, p := range publicPaths {
			if strings.HasPrefix(path, p) {
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
			logger.Debug("Token is empty")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Decode without verifying to get key ID
		unsafeClaims, err := jwt.Decode(c.Request.Context(), &jwt.DecodeParams{
			Token: token,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			logger.Debug("Claims error: %s", err)
			return
		}

		// Fetch JWKS for key ID
		jwk, err := jwt.GetJSONWebKey(c.Request.Context(), &jwt.GetJSONWebKeyParams{
			KeyID:      unsafeClaims.KeyID,
			JWKSClient: jwksClient,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			logger.Debug("JWK error: %s", err)
			return
		}

		// Verify the token
		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token: token,
			JWK:   jwk,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			logger.Debug("Token verification error: %s", err)
			return
		}

		// Save claims into Gin context
		c.Set(string(ClerkClaimsKey), claims)

		c.Next()
	}
}


