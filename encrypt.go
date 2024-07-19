package hubur

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"github.com/spf13/cast"
	"github.com/twmb/murmur3"
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

// Mmh3Base64Encode 计算 base64 的值,mmh3 base64 编码，编码后的数据要求每 76 个字符加上换行符。具体原因 RFC 822 文档上有说明。然后 32 位 mmh3 hash
func Mmh3Base64Encode(braw string) string {
	bckd := base64.StdEncoding.EncodeToString([]byte(braw))
	var buffer bytes.Buffer
	for i := 0; i < len(bckd); i++ {
		ch := bckd[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')
	return buffer.String()
}

func Mmh3Hash32(raw string) string {
	h32 := murmur3.New32()
	_, _ = h32.Write([]byte(raw))
	return fmt.Sprintf("%d", int32(h32.Sum32()))
}

func FaviconHash(favicon []byte) int64 {
	return cast.ToInt64(Mmh3Hash32(Mmh3Base64Encode(string(favicon))))
}
