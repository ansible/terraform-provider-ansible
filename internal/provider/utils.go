package provider

import (
	"crypto/sha256"
	"encoding/hex"
)

func sha256Sum(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
