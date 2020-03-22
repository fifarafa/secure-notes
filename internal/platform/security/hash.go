package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func GenerateHashWithSalt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", errors.New("bcrypt generate from password")
	}
	return string(hash), nil
}
