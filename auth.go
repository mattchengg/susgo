package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

const (
	KEY_1 = "vicopx7dqu06emacgpnpy8j8zwhduwlh"
	KEY_2 = "9u7qab84rpc16gvk"
)

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}
	padding := int(data[length-1])
	if padding > length {
		return data
	}
	return data[:length-padding]
}

func aesEncrypt(input, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := key[:16]
	padded := pkcs7Pad(input, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)
	return ciphertext, nil
}

func aesDecrypt(input, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := key[:16]
	plaintext := make([]byte, len(input))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, input)
	return pkcs7Unpad(plaintext), nil
}

func deriveKey(nonce string) []byte {
	key := make([]byte, 32)
	for i := 0; i < 16; i++ {
		key[i] = KEY_1[int(nonce[i])%16]
	}
	copy(key[16:], KEY_2)
	return key
}

func getAuth(nonce string) (string, error) {
	nkey := deriveKey(nonce)
	authData, err := aesEncrypt([]byte(nonce), nkey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(authData), nil
}

func decryptNonce(encNonce string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encNonce)
	if err != nil {
		return "", err
	}
	decrypted, err := aesDecrypt(data, []byte(KEY_1))
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}
