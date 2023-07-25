package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"os"
)

func Encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

func EncryptFile(key []byte, inputFile, outputFile string) error {
	plaintext, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer plaintext.Close()

	ciphertext, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer ciphertext.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	ciphertext.Write(iv)

	buffer := make([]byte, 4096)
	for {
		n, err := plaintext.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		stream.XORKeyStream(buffer[:n], buffer[:n])
		ciphertext.Write(buffer[:n])
	}

	return nil
}

func DecryptFile(key []byte, inputFile, outputFile string) error {
	ciphertext, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer ciphertext.Close()

	plaintext, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer plaintext.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	iv := make([]byte, aes.BlockSize)
	_, err = ciphertext.Read(iv)
	if err != nil {
		return err
	}

	stream := cipher.NewCFBDecrypter(block, iv)

	buffer := make([]byte, 4096)
	for {
		n, err := ciphertext.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		stream.XORKeyStream(buffer[:n], buffer[:n])
		plaintext.Write(buffer[:n])
	}

	return nil
}