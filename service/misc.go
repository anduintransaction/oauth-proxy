package service

import (
	"crypto/rand"
	"fmt"
)

func Version() string {
	return "0.3.0"
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
