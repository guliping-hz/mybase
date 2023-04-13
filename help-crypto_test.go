package mybase

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"testing"
)

func TestXor(t *testing.T) {
	xorKey := "xorKey"
	miwen := Xor([]byte("511322200007155913"), xorKey)
	t.Logf("mi=%v,orgin=%s", miwen, string(Xor(miwen, xorKey)))
}

func TestAESGCMDecrypter(t *testing.T) {
	key := "2836e95fcd10e04b0069bb1ee659955b"
	plainText := `{"ai":"test-accountId","name":"用户姓名","idNum":"222222222222222222"}`
	cipherBytes2, _ := AESGCMEncrypt(key, plainText)
	cipherBase64Text1 := base64.StdEncoding.EncodeToString(cipherBytes2)
	cipherBytes3, _ := base64.StdEncoding.DecodeString(cipherBase64Text1)
	plainText3, err := AESGCMDecrypter(key, cipherBytes3)
	if plainText3 == plainText {
		t.Logf("原文加密再解密成功\n")
	} else {
		t.Errorf("解密失败,err=%s\n", err)
	}
}

func TestHMACSHA1Base64(t *testing.T) {
	key := []byte("123456")
	text := []byte("aaaaaa")

	ctx := hmac.New(sha1.New, key)
	t.Log(base64.StdEncoding.EncodeToString(ctx.Sum(text)))

	//key = []byte("123456")
	//text = []byte("aaaaaa")

	ctx2 := hmac.New(sha1.New, key)
	ctx2.Write(text)
	t.Log(base64.StdEncoding.EncodeToString(ctx2.Sum(nil)))

	ctx3 := hmac.New(sha1.New, key)
	ctx3.Write(text)
	t.Log(base64.StdEncoding.EncodeToString(ctx3.Sum(nil)))

}
