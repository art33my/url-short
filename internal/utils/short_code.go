package utils

import (
	"crypto/rand"
	"encoding/base64"
	"url-short/internal/repositories"
)

func GenerateUniqueShortCode(repo *repositories.LinkRepository) (string, error) {
	for {
		code, err := GenerateRandomCode(6)
		if err != nil {
			return "", err
		}

		exists, err := repo.IsShortCodeExist(code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
}

func GenerateRandomCode(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}
