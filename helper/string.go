package helper

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"math/big"
)

func ChooseString(newVal, oldVal string) string {
	if newVal != "" {
		return newVal
	}
	return oldVal
}

func ChooseNullString(newVal string, oldVal sql.NullString) sql.NullString {
	if newVal != "" {
		return sql.NullString{String: newVal, Valid: true}
	}
	return oldVal
}

// GenerateRandomToken generates a random token for login
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateSimpleToken generates a simple alphanumeric token
func GenerateSimpleToken(length int) (string, error) {
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ123456789"
	token := make([]byte, length)
	for i := range token {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		token[i] = charset[randomIndex.Int64()]
	}
	return string(token), nil
}
