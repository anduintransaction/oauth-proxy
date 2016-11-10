package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"gottb.io/goru/config"
	"gottb.io/goru/errors"
)

var key []byte

func Start(config *config.Config) error {
	secretConfig, err := config.Get("general.secret")
	if err != nil {
		return err
	}
	secret, err := secretConfig.Str()
	if err != nil {
		return err
	}
	if len(secret) == 0 {
		return errors.Errorf("secret string must be configured")
	}
	for len(secret) <= 32 {
		secret += "A"
	}
	secret = secret[0:32]
	key = []byte(secret)
	return nil
}

func Encrypt(text []byte) ([]byte, error) {
	return encrypt(key, text)
}

func Decrypt(text []byte) ([]byte, error) {
	return decrypt(key, text)
}

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, errors.Wrap(err)
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return ciphertext, nil
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if len(text) < aes.BlockSize {
		return nil, errors.Errorf("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	return text, nil
}
