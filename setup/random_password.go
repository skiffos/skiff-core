package setup

import "crypto/rand"

const passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const generatedPasswordLength = 64

func randomPassword() (string, error) {
	password := make([]byte, generatedPasswordLength)
	const unbiasedLimit = byte(248)
	for offset := 0; offset < len(password); {
		var randomBytes [generatedPasswordLength]byte
		if _, err := rand.Read(randomBytes[:]); err != nil {
			return "", err
		}
		for _, value := range randomBytes {
			if value >= unbiasedLimit {
				continue
			}
			password[offset] = passwordCharset[int(value)%len(passwordCharset)]
			offset++
			if offset == len(password) {
				break
			}
		}
	}
	return string(password), nil
}
