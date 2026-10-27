package common

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// UseRandomJWTSecretForTests sets a random signing secret, so tests never depend on a fixed one
func UseRandomJWTSecretForTests() {
	jwtSecret = []byte(RandString(48))
}

// HeaderTokenMock adds authorization token to request header for testing
func HeaderTokenMock(req *http.Request, u uint) {
	req.Header.Set("Authorization", fmt.Sprintf("Token %v", GenToken(u)))
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
// Used for testing token extraction logic
func ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 6 && authHeader[:6] == "Token " {
		return authHeader[6:]
	}
	return ""
}

// VerifyTokenClaims verifies a JWT token and returns claims for testing
func VerifyTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}
