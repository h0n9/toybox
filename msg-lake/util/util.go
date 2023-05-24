package util

import (
	"crypto/rand"
	"encoding/base64"
)

func CheckStrLen(target string, min, max int) bool {
	l := len(target)
	return l > min && l < max
}

func GenerateRandomBase64String(size int) string {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return base64.RawStdEncoding.EncodeToString(bytes)
}
