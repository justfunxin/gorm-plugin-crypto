package crypto

import (
	"encoding/base64"
	"github.com/duke-git/lancet/v2/cryptor"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

type AesCryptoStrategy struct {
	secretKey string
}

func NewAesCryptoStrategy(secretKey string) *AesCryptoStrategy {
	return &AesCryptoStrategy{secretKey: secretKey}
}

func (d *AesCryptoStrategy) Name() string {
	return "AES"
}

func (d *AesCryptoStrategy) Encrypt(src string, field *schema.Field, db *gorm.DB) string {
	if len(src) == 0 {
		return ""
	}
	return EncryptValue(src, d.secretKey)
}

func (d *AesCryptoStrategy) Decrypt(src string, field *schema.Field, db *gorm.DB) string {
	if len(src) == 0 {
		return ""
	}
	return DecryptValue(src, d.secretKey)
}

func EncryptValue(str, key string) string {
	return "{AES}" + base64.StdEncoding.EncodeToString(cryptor.AesEcbEncrypt([]byte(str), []byte(key)))
}

func DecryptValue(str, key string) string {
	if strings.HasPrefix(str, "{AES}") {
		rs, _ := base64.StdEncoding.DecodeString(str[5:])
		v := cryptor.AesEcbDecrypt(rs, []byte(key))
		return string(v)
	}
	return str
}
