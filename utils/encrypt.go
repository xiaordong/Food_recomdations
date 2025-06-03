package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var key = []byte("something for no") //16，24，32

// 填充函数
func padding(text []byte, size int) []byte {
	last := len(text) % size
	paddingLen := size - last
	padding := bytes.Repeat([]byte{0}, paddingLen)
	res := append(text, padding...)
	return res
}

// Crypto 加密函数
func Crypto(text string) (string, error) {
	txt := []byte(text)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	move := key[:blockSize]
	padded := padding(txt, blockSize)
	mode := cipher.NewCBCEncrypter(block, move)
	encrypt := make([]byte, len(padded))
	mode.CryptBlocks(encrypt, padded)
	return base64.StdEncoding.EncodeToString(encrypt), nil
}
