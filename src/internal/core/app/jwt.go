package app

import (
	"errors"
	"fmt"
	"net/http"
	"shmoopicks/src/internal/core/timex"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtCookieName = "shmoopicks_token"
	jwtTTL        = 1 * timex.Day
)

// Claims struct to encode in JWT
type Claims struct {
	jwt.RegisteredClaims
	UserID *string `json:"user_id"`
}

func NewClaims() *Claims {
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "shmoopicks",
		},
	}
}

func (c *Claims) JWT(secret string) (string, error) {
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	// Sign token with secret (convert to []byte for HMAC)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (c *Claims) Save(cfg Config, w http.ResponseWriter) error {
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(jwtTTL))
	token, err := c.JWT(cfg.JwtSecret)
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	cookie := &http.Cookie{
		Name:     jwtCookieName,
		Value:    token,
		Path:     "/", // Cookie available for entire domain
		MaxAge:   int(jwtTTL.Seconds()),
		HttpOnly: true,
		Secure:   cfg.Env == EnvProd,
		SameSite: http.SameSiteLaxMode, // CSRF protection, allows same-site redirects
	}
	http.SetCookie(w, cookie)

	return nil
}

// clearTokenCookie removes the JWT cookie (for logout)
func (c *Claims) Delete(cfg Config, w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     jwtCookieName,
		Value:    "",
		Path:     "/", // Must match the path used when setting the cookie
		MaxAge:   -1,  // Delete cookie
		HttpOnly: true,
		Secure:   cfg.Env == EnvProd,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// validateJWT validates and parses a JWT token
func ValidateClaims(tokenString string, secret string) (*Claims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// getTokenFromCookie extracts JWT from cookie
func ValidateClaimsFromCookie(r *http.Request, secret string) (*Claims, error) {
	cookie, err := r.Cookie(jwtCookieName)
	if err != nil {
		return nil, err
	}
	return ValidateClaims(cookie.Value, secret)
}

func ValidateClaimsFromHeader(r *http.Request, secret string) (*Claims, error) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("Authorization header required")
	}

	// Extract token (format: "Bearer <token>")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("Invalid authorization header format")
	}

	tokenString := parts[1]

	// Validate token
	claims, err := ValidateClaims(tokenString, secret)
	if err != nil {
		return nil, fmt.Errorf("Invalid or expired token: %v", err)
	}

	return claims, nil
}

func ValidateClaimsFromRequest(r *http.Request, secret string) (*Claims, error) {
	cookie, err := r.Cookie(jwtCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		cookie = nil
	} else if err != nil {
		return nil, fmt.Errorf("error reading cookie: %w", err)
	}

	if cookie != nil {
		return ValidateClaimsFromCookie(r, secret)
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		return ValidateClaimsFromHeader(r, secret)
	}

	return nil, fmt.Errorf("no valid auth source found")
}
