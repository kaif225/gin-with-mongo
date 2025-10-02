package utils

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerifyPassword(password, encodedHash string) error {

	parts := strings.Split(encodedHash, ".")

	if len(parts) != 2 {
		return errors.New("invalid format")
	}
	saltBase64 := parts[0]
	hashBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)

	if err != nil {
		return err
	}

	hash, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return err
	}

	hashPassword := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	if len(hash) != len(hashPassword) {
		return errors.New("incorrect password")
	}

	if subtle.ConstantTimeCompare(hash, hashPassword) != 1 {
		return errors.New("incorrect password")
	}

	return nil
}
