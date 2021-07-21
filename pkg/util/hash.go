package util

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
)

func SHA256Hash(data io.Reader) (string, error) {
	hasher := sha256.New()
	_, err := io.Copy(hasher, data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}
