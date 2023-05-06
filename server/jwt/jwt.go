package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

var MySigningKey = []byte("your-secret")

func GenerateToken(maxAge int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "your_name"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Second * time.Duration(maxAge)).Unix()

	// Generate encoded token and send it as response.
	return token.SignedString(MySigningKey)
}
