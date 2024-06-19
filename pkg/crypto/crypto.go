package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"os"
	"errors"
	"io"
	"crypto/rand"
)

func DecryptData(data []byte) (string, error) {
	key := []byte(os.Getenv("CRYPTO_KEY"))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(data) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	if len(data) % aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	return string(data), nil
}

func EncryptData(data []byte) (string, error) {
	key := []byte(os.Getenv("CRYPTO_KEY"))

	if len(data) % aes.BlockSize != 0 {
		data = addPaddingBytes(data)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize + len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], data)

	return string(data), nil
}

func addPaddingBytes(data []byte) []byte {
    num := aes.BlockSize - len(data) % aes.BlockSize
    for i := 0; i < num; i++ {
        data = append(data, 0)
    }
	
    return data
}