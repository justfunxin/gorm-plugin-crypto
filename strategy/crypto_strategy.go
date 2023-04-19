package strategy

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type CryptoStrategy interface {
	//
	// Name
	//  @Description:  CryptoType Name
	//  @return string crypto type name
	//
	Name() string
	//
	// Encrypt
	//  @Description: encrypt data
	//  @param src
	//  @param field
	//  @param db
	//  @return string encryption value
	//
	Encrypt(src string, field *schema.Field, db *gorm.DB) string
	//
	// Decrypt
	//  @Description: decrypt data
	//  @param src
	//  @param field
	//  @param db
	//  @return string decrypted value
	//
	Decrypt(src string, field *schema.Field, db *gorm.DB) string
}
