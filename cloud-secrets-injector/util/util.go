package util

import (
	"encoding/base64"
	"os"
)

func GetEnv(key, fallback string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}
	return fallback
}

func DecodeBase64(input string) (string, error) {
	output, err := base64.RawStdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
