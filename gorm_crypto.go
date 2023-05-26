// Package crypto is a GORM plugin, encrypt and decrypt struct field with tag.
package crypto

import (
	"github.com/kangarooxin/gorm-plugin-crypto/strategy"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
)

const (
	// Name gorm plugin name
	Name = "crypto"
	// CryptoTag struct tag name
	CryptoTag = "crypto"
)

var cryptoStrategies map[string]strategy.CryptoStrategy

func RegisterCryptoStrategy(cryptoStrategy strategy.CryptoStrategy) {
	cryptoStrategies[strings.ToUpper(cryptoStrategy.Name())] = cryptoStrategy
}

func GetCryptoStrategy(cryptoType string) strategy.CryptoStrategy {
	return cryptoStrategies[strings.ToUpper(cryptoType)]
}

type CryptoPlugin struct {
}

func NewCryptoPlugin() *CryptoPlugin {
	return &CryptoPlugin{}
}

func (m *CryptoPlugin) Name() string {
	return Name
}

func (m *CryptoPlugin) Initialize(db *gorm.DB) error {
	cryptoStrategies = make(map[string]strategy.CryptoStrategy)
	db.Callback().Create().Before("gorm:create").Register("crypt_plugin:before_create", EncryptParamBeforeCreate)
	db.Callback().Update().Before("gorm:update").Register("crypt_plugin:before_update", EncryptParamBeforeUpdate)
	db.Callback().Query().Before("gorm:query").Register("crypt_plugin:before_query", EncryptParamBeforeQuery)
	db.Callback().Query().After("gorm:query").Register("crypt_plugin:after_query", DecryptResultAfterQuery)
	return nil
}

type CryptoValue struct {
	Column string
	Value  string
}

func NewCryptoValue(column string, val string) *CryptoValue {
	return &CryptoValue{
		Column: column,
		Value:  val,
	}
}

type CryptoField struct {
	Field      *schema.Field
	CryptoType string
	Strategy   strategy.CryptoStrategy
}

var cryptoFieldsMap = cmap.New[[]*CryptoField]()

func EncryptParamBeforeCreate(db *gorm.DB) {
	fields := getSchemaCryptoFields(db.Statement.Schema)
	if fields == nil || len(fields) == 0 {
		return
	}
	switch db.Statement.ReflectValue.Kind() {
	case reflect.Slice:
		for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
			sliceV := db.Statement.ReflectValue.Index(i)
			encryptFields(sliceV, fields, db)
		}
	case reflect.Struct:
		encryptFields(db.Statement.ReflectValue, fields, db)
	default:
		db.Config.Logger.Error(db.Statement.Context, "unsupported kind [%d] of value", db.Statement.ReflectValue.Kind())
	}
}

func EncryptParamBeforeQuery(db *gorm.DB) {
	fields := getSchemaCryptoFields(db.Statement.Schema)
	if len(fields) == 0 {
		return
	}
	if we := db.Statement.Clauses["WHERE"].Expression; we != nil {
		exprs := we.(clause.Where).Exprs
		for i, expr := range exprs {
			switch expr.(type) {
			case clause.Eq:
				eq := expr.(clause.Eq)
				if cf := findCryptFieldWithClauseColumn(eq.Column, fields); cf != nil {
					exprs[i] = clause.Eq{
						Column: eq.Column,
						Value:  cf.Strategy.Encrypt(eq.Value.(string), cf.Field, db),
					}
				}
			case clause.IN:
				in := expr.(clause.IN)
				if cf := findCryptFieldWithClauseColumn(in.Column, fields); cf != nil {
					exprs[i] = clause.IN{
						Column: in.Column,
						Values: encryptValues(in.Values, cf, db),
					}
				}
			case clause.Expr:
				expr := expr.(clause.Expr)
				if expr.Vars != nil {
					for i, v := range expr.Vars {
						if v, ok := v.(*CryptoValue); ok {
							if cf := findCryptFieldWithClauseColumn(v.Column, fields); cf != nil {
								expr.Vars[i] = cf.Strategy.Encrypt(v.Value, cf.Field, db)
							}
						}
					}
				}
			}

		}
	}
}

func EncryptParamBeforeUpdate(db *gorm.DB) {
	if updateInfo, ok := db.Statement.Dest.(map[string]interface{}); ok {
		for updateColumn := range updateInfo {
			updateV := updateInfo[updateColumn]
			updateField := db.Statement.Schema.LookUpField(updateColumn)
			if ct, ok := updateField.Tag.Lookup(CryptoTag); ok {
				fieldStrategy := GetCryptoStrategy(ct)
				encryptionValue := fieldStrategy.Encrypt(cast.ToString(updateV), updateField, db)
				updateInfo[updateColumn] = encryptionValue
			}
		}
		return
	}
	destType, destValue := getReflectElem(db.Statement.Dest)
	if destType != nil {
		for i := 0; i < destType.NumField(); i++ {
			field := destType.Field(i)
			if ct, ok := field.Tag.Lookup(CryptoTag); ok {
				val := destValue.Field(i).String()
				if len(val) == 0 {
					continue
				}
				fieldStrategy := GetCryptoStrategy(ct)
				dbField := db.Statement.Schema.LookUpField(field.Name)
				encryptionValue := fieldStrategy.Encrypt(cast.ToString(val), dbField, db)
				destValue.Field(i).SetString(encryptionValue)
			}
		}
	}
}

func getReflectElem(i any) (reflect.Type, reflect.Value) {
	destType := reflect.TypeOf(i)
	destValue := reflect.ValueOf(i)
	for {
		if destType.Elem().Kind() == reflect.Pointer {
			destType = destType.Elem()
			destValue = destValue.Elem()
		} else if destType.Elem().Kind() == reflect.Struct {
			break
		} else {
			return nil, reflect.ValueOf(nil)
		}
	}
	return destType.Elem(), destValue.Elem()
}

func DecryptResultAfterQuery(db *gorm.DB) {
	fields := getSchemaCryptoFields(db.Statement.Schema)
	if len(fields) == 0 {
		return
	}
	refVal := reflect.ValueOf(db.Statement.Dest).Elem()
	switch refVal.Kind() {
	case reflect.Struct:
		decryptFields(refVal, fields, db)
	case reflect.Slice:
		if refVal.Len() >= 1 {
			for i := 0; i < refVal.Len(); i++ {
				curVal := refVal.Index(i)
				if curVal.Kind() == reflect.Pointer {
					curVal = curVal.Elem()
				}
				if curVal.Kind() == reflect.Struct {
					decryptFields(curVal, fields, db)
				}
			}
		}
	}
}

func getSchemaCryptoFields(schema *schema.Schema) []*CryptoField {
	if schema == nil {
		return nil
	}
	schemaName := schema.String()
	if fields, t := cryptoFieldsMap.Get(schemaName); t {
		return fields
	}
	var fields []*CryptoField
	fieldNames := schema.FieldsByName
	for k := range fieldNames {
		v := fieldNames[k]
		if ct, ok := v.Tag.Lookup(CryptoTag); ok {
			if v.FieldType.Kind() != reflect.String {
				continue
			}
			fieldStrategy := GetCryptoStrategy(ct)
			fields = append(fields, &CryptoField{
				Field:      v,
				CryptoType: ct,
				Strategy:   fieldStrategy,
			})
		}
	}
	cryptoFieldsMap.Set(schemaName, fields)
	return fields
}

func encryptValues(values []interface{}, cf *CryptoField, db *gorm.DB) []interface{} {
	for i, v := range values {
		if v, ok := v.(string); ok {
			values[i] = cf.Strategy.Encrypt(v, cf.Field, db)
		}
	}
	return values
}

func findCryptFieldWithClauseColumn(column interface{}, fields []*CryptoField) *CryptoField {
	if c, ok := (column).(string); ok {
		return findCryptField(c, fields)
	}
	if c, ok := (column).(clause.Column); ok {
		return findCryptField(c.Name, fields)
	}
	return nil
}

func findCryptField(fieldName string, fields []*CryptoField) *CryptoField {
	if fieldName == "" {
		return nil
	}
	for _, f := range fields {
		if f.Field.DBName == fieldName {
			return f
		}
	}
	return nil
}

func decryptFields(src reflect.Value, fields []*CryptoField, db *gorm.DB) {
	for k := range fields {
		field := fields[k].Field
		fieldStrategy := fields[k].Strategy
		refValue := src.FieldByName(field.Name)
		encryptValue := refValue.String()
		if len(encryptValue) > 0 {
			actualValue := fieldStrategy.Decrypt(encryptValue, field, db)
			refValue.SetString(actualValue)
		}
	}
}

func encryptFields(reflectValue reflect.Value, fields []*CryptoField, db *gorm.DB) {
	dbCtx := db.Statement.Context
	for k := range fields {
		field := fields[k].Field
		fieldStrategy := fields[k].Strategy
		if actualValue, isZero := field.ValueOf(dbCtx, reflectValue); !isZero {
			encryptionValue := fieldStrategy.Encrypt(cast.ToString(actualValue), field, db)
			if err := field.Set(dbCtx, reflectValue, encryptionValue); err != nil {
				panic(err)
			}
		} else {
			continue
		}
	}
}
