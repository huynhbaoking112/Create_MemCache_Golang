package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/global"
	"github.com/huynhbaoking112/Create_MemCache_Golang.git/models"
)

func RequireAuth(c *gin.Context) {

	// Get the cookie of req
	tokenString, err := c.Cookie("Authorization")
	if err != nil || tokenString == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Decode/validate it
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECERET")), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check the exp
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Find the user with token sub
		var user models.Employee
		userID := uint(claims["sub"].(float64)) // Chuyển đổi từ float64 sang uint
		if err := global.Mdb.First(&user, userID).Error; err != nil || user.ID == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if user.IsActive == "NO" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// Attach to req
		c.Set("user", user)

		// Continue
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
