package config

import "os"

// LocalMode is true when the process authenticates with a local JWT
// instead of Firebase. Unset AUTH_MODE means local, so the API boots
// without a service account.
func LocalMode() bool {
	switch os.Getenv("AUTH_MODE") {
	case "", "local":
		return true
	default:
		return false
	}
}

func JWTSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return "dev-secret-change-me"
}
