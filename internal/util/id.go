package util

import "github.com/google/uuid"

// MustUUID returns a new UUID string or panics (used during controlled creation).
func MustUUID() string {
	return uuid.NewString()
}
