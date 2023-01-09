package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func Base64(input []byte) string {
	return base64.RawStdEncoding.EncodeToString(input)
}

func GenRandomBytes(n uint32) ([]byte, error) {
	data := make([]byte, n)
	if _, err := rand.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}
