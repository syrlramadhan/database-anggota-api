package helper

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

type Claims struct {
	NRA    string `json:"nra"`
	Nama   string `json:"nama"`
	Status string `json:"status"`
	jwt.StandardClaims
}

func GenerateJWT(nra, nama, status string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	jwtKey := os.Getenv("JWT_SECRET")
	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &Claims{
		NRA:    nra,
		Nama:   nama,
		Status: status,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtKey))
}
