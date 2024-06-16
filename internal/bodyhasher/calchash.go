package bodyhasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

type HashKey struct {
	Key []byte
}

func (h *HashKey) UnmarshalText(text []byte) error {
	h.Key = text
	return nil
}
func (h *HashKey) MarshalText() ([]byte, error) {
	return h.Key, nil
}

func CalculateHash(data []byte, hashKey []byte) (string, error) {
	h := hmac.New(sha256.New, hashKey)
	_, err := h.Write(data)
	if err != nil {
		return "", fmt.Errorf("didn't come up with %w", err)
	}
	hs := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return hs, nil
}
