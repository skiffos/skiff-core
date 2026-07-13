package setup

import (
	"strings"
	"testing"
)

func TestRandomPasswordUsesConfiguredAlphabetAndLength(t *testing.T) {
	password, err := randomPassword()
	if err != nil {
		t.Fatal(err)
	}
	if len(password) != generatedPasswordLength {
		t.Fatalf("expected %d bytes, got %d", generatedPasswordLength, len(password))
	}
	for _, character := range password {
		if !strings.ContainsRune(passwordCharset, character) {
			t.Fatalf("unexpected password character %q", character)
		}
	}
}
