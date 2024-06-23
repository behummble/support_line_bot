package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"os"
	"errors"
	"io"
	"crypto/rand"
)

func DecryptData(data string) (string, error) {
	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}
	key := []byte(os.Getenv("CRYPTO_KEY"))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(dataBytes) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := dataBytes[:aes.BlockSize]
	dataBytes = dataBytes[aes.BlockSize:]

	if len(dataBytes) % aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(dataBytes, dataBytes)

	dataBytes = removePaddingBytes(dataBytes)
	
	return string(dataBytes), nil
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

	return hex.EncodeToString(ciphertext), nil
}

func addPaddingBytes(data []byte) []byte {
	l := 16 - len(data) % 16
	padding := make([]byte, l)
	padding[l-1] = byte(l)
	return append(data, padding...)
}

func removePaddingBytes(data []byte) []byte {
	l := int(data[len(data)-1])
	if l > 16 {
		return data
	}

	return data[:len(data)-l]
}