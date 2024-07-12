package hubur

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

func AESEncrypt(plainText string, keyByte []byte) (string, error) {
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	origData := PKCS5Padding([]byte(plainText), blockSize)
	blockMode := cipher.NewCBCEncrypter(block, keyByte[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	cipherText := base64.StdEncoding.EncodeToString(crypted)
	return cipherText, nil
}

// AESDecrypt AES解密函数实现
func AESDecrypt(cipherText string, keyByte []byte) ([]byte, error) {
	crypted, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(keyByte)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, keyByte[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

// PKCS5Padding PKCS5加密
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS5UnPadding PKCS5解密
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// //////////////////////AES CBC模式////////////////////////////////
// Zero填充方式
func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	//填充
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

// Zero反填充
func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// 加密
func AesEncrypt(encodeStr string, key []byte, iv string) (string, error) {
	encodeBytes := []byte(encodeStr)
	//根据key 生成密文
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	encodeBytes = ZeroPadding(encodeBytes, blockSize) //PKCS5Padding(encodeBytes, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, []byte(iv))
	crypted := make([]byte, len(encodeBytes))
	blockMode.CryptBlocks(crypted, encodeBytes)

	hexstr := fmt.Sprintf("%x", crypted)
	return hexstr, nil
	//return base64.StdEncoding.EncodeToString(crypted), nil
}

func Base64Decode(data string) (string, error) {
	decode, err := base64.StdEncoding.DecodeString(data)
	return string(decode), err
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
