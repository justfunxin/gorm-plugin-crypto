# [gorm-plugin-crypto](https://github.com/kangarooxin/gorm-plugin-crypto)

this is a gorm plugin, encrypt and decrypt struct field with tag.

## Usage:
1. import lib
```go
go get -u github.com/kangarooxin/gorm-plugin-crypto
```

2.  Register Plugin
```go
db, _ = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
db.Use(crypto.NewCryptoPlugin())
// Register Default AES Strategy
RegisterCryptoStrategy(strategy.NewAesCryptoStrategy("1234567890123456"))
db.AutoMigrate(&User{})

```
3. Mark the field with tag `crypto`
```go
type User struct {
	ID     uint   `gorm:"primarykey"`
	Name   string `gorm:"column:name"`
	Age    int    `gorm:"column:age"`
	Email  string `gorm:"column:email" crypto:"aes"`
	Mobile string `gorm:"column:mobile" crypto:"aes"`
}
```
4. Implement interface `CryptoStrategy` and register your own strategies.
```go
// register strategy
crypto.RegisterCryptoStrategy(MyAesCryptoStrategy("1234567890123456"))

// use in tag
Email  string `gorm:"column:email" crypto:"myAes"`
```

### Start use:

#### Insert
```go
// insert
user1 := &User{ID: 1, Name: "User1", Age: 18, Email: "user1@example.com", Mobile: "13812345671"}
user2 := &User{ID: 2, Name: "User2", Age: 12, Email: "user2@example.com", Mobile: "13812345672"}
user3 := &User{ID: 3, Name: "User3", Age: 16, Email: "user3@example.com", Mobile: "13812345673"}

db.Create(user1)
assert.Equal(t, "{AES}g+2fA4EbDGDPpZVCF0quCwjz4w22BRHHb0xqEG86zL0=", user1.Email)
assert.Equal(t, "{AES}Q/FDK7HDVHpArPRm3kCwEw==", user1.Mobile)

// insert batch
users := []*User{user2, user3}
db.Create(users)
assert.Equal(t, "{AES}VNrSbyrCwXfcBIoxYbO8hgjz4w22BRHHb0xqEG86zL0=", user2.Email)
assert.Equal(t, "{AES}yRPUirBKNe9UFlIStzft/gjz4w22BRHHb0xqEG86zL0=", user3.Email)
```

#### Query
```go
	// query by id
var queryUser User
db.First(&queryUser, 1)
assert.Equal(t, "user1@example.com", queryUser.Email)
assert.Equal(t, "13812345671", queryUser.Mobile)

// query batch by ids
var queryUsers []User
db.Find(&queryUsers, []int{1, 2, 3})
assert.Equal(t, "user1@example.com", queryUsers[0].Email)
assert.Equal(t, "13812345672", queryUsers[1].Mobile)

// query all
var queryAllUsers []User
db.Find(&queryAllUsers)
assert.Equal(t, "user1@example.com", queryUsers[0].Email)
assert.Equal(t, "13812345672", queryUsers[1].Mobile)

// query by string
var queryUser1 User
db.Where("name = ?", "User1").First(&queryUser1)
assert.Equal(t, "user1@example.com", queryUser1.Email)
assert.Equal(t, "13812345671", queryUser1.Mobile)

// query by encrypted string, must encrypt manually or wrapper with `NewCryptoValue`
var queryUser2 User
db.Where("email = ?", NewCryptoValue("email", "user1@example.com")).First(&queryUser2)
assert.Equal(t, "user1@example.com", queryUser2.Email)
assert.Equal(t, "13812345671", queryUser2.Mobile)

// query by Struct
var queryUser4 User
db.Where(&User{Email: "user1@example.com"}).First(&queryUser4)
assert.Equal(t, "13812345671", queryUser4.Mobile)

// query by Map
var queryUser5 User
db.Where(map[string]interface{}{"email": "user1@example.com"}).First(&queryUser5)
assert.Equal(t, "13812345671", queryUser5.Mobile)

// query by Map
var queryUser6 []User
db.Where(map[string]interface{}{
"email": []string{"user1@example.com", "user2@example.com"},
}).Find(&queryUser6)
assert.Equal(t, "13812345671", queryUser6[0].Mobile)
assert.Equal(t, "13812345672", queryUser6[1].Mobile)

// query by raw
var queryUser3 User
db.Raw("select * from test_user").Find(&queryUser3)
assert.Equal(t, "user1@example.com", queryUser3.Email)
```

#### Save and Update
```go
// save
var saveUser User
db.First(&saveUser)
saveUser.Email = "User11@example.com"
db.Save(&saveUser)
assert.Equal(t, "{AES}siKVK6qMulucOlmRoZWLiWcZIqVzlNkqP58lypIfHtg=", saveUser.Email)

// save without id
user4 := &User{Name: "User4", Age: 18, Email: "user4@example.com", Mobile: "13812345674"}
db.Save(user4)
assert.Equal(t, "{AES}g1WxCfYDcw/2k5g9kyFDpAjz4w22BRHHb0xqEG86zL0=", user4.Email)

// update one column with condition
db.Model(&User{}).Where("id = ?", 1).Update("email", "user111@example.com")
var queryUser7 User
err = db.First(&queryUser7, 1).Error
assert.NoError(t, err)
assert.Equal(t, "user111@example.com", queryUser7.Email)

// Update attributes with `map`
db.Model(&User{}).Where("id = ?", 2).Updates(map[string]interface{}{"email": "user222@example.com"})
var queryUser8 User
err = db.First(&queryUser8, 2).Error
assert.NoError(t, err)
assert.Equal(t, "user222@example.com", queryUser8.Email)

// Update attributes with `struct`
db.Model(User{}).Where("id = ?", "1").Updates(User{Email: "user1113@example.com"})
var queryUser9 User
err = db.First(&queryUser9, 1).Error
assert.NoError(t, err)
assert.Equal(t, "user1113@example.com", queryUser9.Email)
```
