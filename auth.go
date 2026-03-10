package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

func initAuth() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "puundoho-dashboard-secret-key-2026"
		log.Println("⚠️  JWT_SECRET not set, using default (not safe for production)")
	}
	jwtSecret = []byte(secret)

	// Seed default admin user if the table is empty
	if DB != nil {
		seedAdminUser()
	}
}

// ---------------------------------------------------------------------------
// Password helpers
// ---------------------------------------------------------------------------

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ---------------------------------------------------------------------------
// JWT helpers
// ---------------------------------------------------------------------------

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func generateToken(userID int, username, role string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func parseToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// ---------------------------------------------------------------------------
// Auth middleware
// ---------------------------------------------------------------------------

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token diperlukan"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := parseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau kedaluwarsa"})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan password diperlukan"})
		return
	}

	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Database tidak tersedia"})
		return
	}

	var id int
	var passwordHash, role string
	err := DB.QueryRow(
		"SELECT id, password_hash, role FROM admin_users WHERE username = $1", req.Username,
	).Scan(&id, &passwordHash, &role)

	if err == sql.ErrNoRows || !checkPassword(passwordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses login"})
		return
	}

	token, err := generateToken(id, req.Username, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": req.Username,
		"role":     role,
	})
}

func meHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id":  c.GetInt("user_id"),
		"username": c.GetString("username"),
		"role":     c.GetString("role"),
	})
}

// ---------------------------------------------------------------------------
// ImageKit auth endpoint   (server-side signature for secure client uploads)
// ---------------------------------------------------------------------------

func imagekitAuthHandler(c *gin.Context) {
	privateKey := os.Getenv("IMAGEKIT_PRIVATE_KEY")
	if privateKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ImageKit not configured"})
		return
	}

	token := fmt.Sprintf("%d", time.Now().UnixNano())
	expire := fmt.Sprintf("%d", time.Now().Add(30*time.Minute).Unix())

	// ImageKit requires HMAC-SHA1(privateKey, token + expire)
	mac := hmac.New(sha1.New, []byte(privateKey))
	mac.Write([]byte(token + expire))
	signature := hex.EncodeToString(mac.Sum(nil))

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"expire":    expire,
		"signature": signature,
	})
}

// ---------------------------------------------------------------------------
// Seed admin
// ---------------------------------------------------------------------------

func seedAdminUser() {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&count)
	if err != nil {
		log.Printf("⚠️  Could not check admin_users table: %v", err)
		return
	}
	if count > 0 {
		return
	}

	hash, _ := hashPassword("admin123")
	_, err = DB.Exec(
		"INSERT INTO admin_users (username, password_hash, role) VALUES ($1, $2, $3)",
		"admin", hash, "admin",
	)
	if err != nil {
		log.Printf("⚠️  Failed to seed admin user: %v", err)
		return
	}
	log.Println("✅ Default admin user created (admin / admin123)")
}
