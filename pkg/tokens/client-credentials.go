package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateClientCredentials generates a new client ID, client secret, and the SHA-256 hash of the client secret.
func GenerateClientCredentials() (string, string, string) {
	clientID := rand.Text()
	clientSecret := rand.Text() + rand.Text()

	secretHash := sha256.Sum256([]byte(clientSecret))
	secretHashHex := hex.EncodeToString(secretHash[:])

	return clientID, clientSecret, secretHashHex
}

// CompareClientSecretHash compares a given client secret with its SHA-256 hash.
func CompareClientSecretHash(clientSecret, clientSecretHash string) bool {
	hash := sha256.Sum256([]byte(clientSecret))
	hashHex := hex.EncodeToString(hash[:])

	return hashHex == clientSecretHash
}
