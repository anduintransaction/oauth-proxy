package crypto

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("abcdefghijklmnopqrstuvwxyz123456")
	text := []byte("plain text")
	secretText, err := encrypt(key, text)
	if err != nil {
		t.Fatal(err)
	}
	plainText, err := decrypt(key, secretText)
	if err != nil {
		t.Fatal(err)
	}
	if string(plainText) != string(text) {
		t.Fatal(text, plainText)
	}
}
