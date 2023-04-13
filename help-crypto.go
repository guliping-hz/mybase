package mybase

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

func Xor(buf []byte, key string) []byte {
	keyBuf := []byte(key)
	size := len(keyBuf)
	result := make([]byte, len(buf))
	for i := range buf {
		result[i] = buf[i] ^ keyBuf[i%size]
	}
	return result
}

// MD5 生成32位MD5,小写
func MD5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

func HMACSHA1Buf(text, key []byte) []byte {
	ctx := hmac.New(sha1.New, key)
	ctx.Write(text)
	return ctx.Sum(nil)
}

func HMACSHA1Base64(text, key []byte) string {
	return base64.StdEncoding.EncodeToString(HMACSHA1Buf(text, key))
}

func HMACSHA1(text, key []byte) string {
	return hex.EncodeToString(HMACSHA1Buf(text, key))
}

func HMACSHA256Buf(text, key []byte) []byte {
	ctx := hmac.New(sha256.New, key)
	ctx.Write(text)
	return ctx.Sum(nil)
}

func HMACSHA256(text, key []byte) string {
	return hex.EncodeToString(HMACSHA256Buf(text, key))
}

func AESEncrypt(plantText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key) //选择加密算法
	if err != nil {
		return nil, err
	}
	plantText = PKCS7Padding(plantText, block.BlockSize())

	iv := make([]byte, block.BlockSize())
	for i := 0; i < len(key); i++ {
		iv[i] = 0
	}
	blockModel := cipher.NewCBCEncrypter(block, iv)

	ciphertext := make([]byte, len(plantText))

	blockModel.CryptBlocks(ciphertext, plantText)
	return ciphertext, nil
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func AesDecrypt(crypted, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) != block.BlockSize() { //提前检查，防止NewCBCDecrypter中崩溃
		return nil, fmt.Errorf("cipher.NewCBCDecrypter: IV length must equal block size")
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	dataLen := len(crypted)
	origData := make([]byte, dataLen)
	blockMode.CryptBlocks(origData, crypted)
	/*
		@注意：有填充需要移除 末尾的数字表示填充的数量
	*/
	return origData[0 : dataLen-int(origData[dataLen-1])], nil
}

//func AESDecrypt(ciphertext, keyBytes []byte) ([]byte, error) {
//	block, err := aes.NewCipher(keyBytes) //选择加密算法
//	if err != nil {
//		return nil, err
//	}
//	blockModel := cipher.NewCBCDecrypter(block, keyBytes)
//	plantText := make([]byte, len(ciphertext))
//	blockModel.CryptBlocks(plantText, ciphertext)
//	plantText = PKCS7UnPadding(plantText, block.BlockSize())
//	return plantText, nil
//}
//
//func PKCS7UnPadding(plantText []byte, blockSize int) []byte {
//	length := len(plantText)
//	unpadding := int(plantText[length-1])
//	return plantText[:(length - unpadding)]
//}

/*
   AES  GCM 加密
   key:加密key
   plaintext：加密明文
   ciphertext:解密返回字节字符串[ 整型以十六进制方式显示]

*/

func AESGCMEncrypt(key, plaintext string) (cipherbyte []byte, err error) {
	keybyte, _ := hex.DecodeString(key)
	plainbyte := []byte(plaintext)
	block, err := aes.NewCipher(keybyte)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err1 := io.ReadFull(rand.Reader, nonce); err1 != nil {
		err = err1
		return
	}

	//@注意：这里dst传入nonce表示在密文头部加入nonce
	cipherbyte = aesgcm.Seal(nonce, nonce, plainbyte, nil)
	return
}

/*
   AES  CBC 解码
   key:解密key
   ciphertext:加密返回的串
   plaintext：解密后的字符串
*/
func AESGCMDecrypter(key string, cipherbyte []byte) (plaintext string, err error) {
	keybyte, _ := hex.DecodeString(key) //转换成16字节。
	block, err := aes.NewCipher(keybyte)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonce := cipherbyte[:aesgcm.NonceSize()] //nonce在密文的前面，我们把它取出来。
	cipherByte1 := cipherbyte[aesgcm.NonceSize():]
	plainbyte, err := aesgcm.Open(nil, nonce, cipherByte1, nil)
	if err != nil {
		return
	}

	//fmt.Printf("%s\n", ciphertext)
	plaintext = string(plainbyte)
	return
}
