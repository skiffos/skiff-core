package setup

import (
	"math/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomPassword() string {
	b := make([]byte, 100+rand.Intn(100))
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
