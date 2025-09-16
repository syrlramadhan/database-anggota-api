package helper

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

type Claims struct {
	NRA  string `json:"nra"`
	Nama string `json:"nama"`
	Role string `json:"role"`
	jwt.StandardClaims
}

func GenerateJWT(nra, nama, role string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	jwtKey := os.Getenv("JWT_SECRET")
	expirationTime := time.Now().Add(5 * time.Hour)

	claims := &Claims{
		NRA:  nra,
		Nama: nama,
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtKey))
}

func ValidateJWT(tokenString string) (*Claims, error) {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	jwtKey := os.Getenv("JWT_SECRET")

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.NewValidationError("token is not valid", jwt.ValidationErrorMalformed)
	}

	return claims, nil
}
